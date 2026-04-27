package service

import (
	"context"
	"fmt"
	"time"

	"driver/taketaxi/common/constants"
	"driver/taketaxi/common/errors"
	"driver/taketaxi/srvDriver/internal/model"
	"driver/taketaxi/srvDriver/internal/repository"

	"go.mongodb.org/mongo-driver/v2/mongo"
)

type PoolService struct {
	repo    *repository.DriverRepo
	mongoDb *mongo.Database
}

func NewPoolService(repo *repository.DriverRepo, mongoDb *mongo.Database) *PoolService {
	return &PoolService{repo: repo, mongoDb: mongoDb}
}

// IsSpecialOrder 判断订单是否为特殊订单（直接进抢单池，不走智能派单）
func IsSpecialOrder(originLat, originLng, destLat, destLng float64, estimateDistance int) bool {
	// 条件1：目的地距离超过 50km（长途/跨城）
	destDistance := haversine(originLat, originLng, destLat, destLng)
	if destDistance > 50000.0 {
		return true
	}

	// 条件2：预估里程超过 50km
	if estimateDistance > 50000 {
		return true
	}

	// 条件3：接驾距离 > 5km（偏僻郊区，附近没司机）
	// 改由 Dispatch 发现无候选司机后自动进池

	return false
}

// PoolOrderResult ListPoolOrders 返回结果（包含计算字段）
type PoolOrderResult struct {
	Order       model.Order
	SecondsLeft int64
}

// ListPoolOrders 查看抢单池中的订单（status=0, driver_id=0）
func (s *PoolService) ListPoolOrders(ctx context.Context, driverID int64, page, pageSize int32) ([]PoolOrderResult, int32, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > constants.PoolPageSizeMax {
		pageSize = constants.PoolPageSizeDefault
	}

	db := s.repo.GetDB()
	now := time.Now()
	timeoutThreshold := now.Add(-time.Duration(constants.PoolOrderTimeoutSec) * time.Second)

	// 先将已超时订单批量取消（lazy cleanup）
	db.Model(&model.Order{}).
		Where("status = ? AND driver_id = ? AND created_at < ?",
			constants.OrderStatusPending, 0, timeoutThreshold).
		Updates(map[string]interface{}{
			"status":       constants.OrderStatusCancelled,
			"cancel_by":    3, // 系统取消
			"cancel_time":  now,
			"cancel_reason": "抢单超时",
		})

	// 查询池中订单：status=0, driver_id=0, 且未超时
	var orders []model.Order
	offset := (page - 1) * pageSize
	if err := db.Where("status = ? AND driver_id = ? AND created_at > ?",
		constants.OrderStatusPending, 0, timeoutThreshold).
		Order("created_at DESC").
		Limit(int(pageSize)).
		Offset(int(offset)).
		Find(&orders).Error; err != nil {
		return nil, 0, fmt.Errorf("list pool orders: %w", err)
	}

	// 查询池中订单总数（超时取消后）
	var total int64
	if err := db.Model(&model.Order{}).Where("status = ? AND driver_id = ? AND created_at > ?",
		constants.OrderStatusPending, 0, timeoutThreshold).
		Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count pool orders: %w", err)
	}

	// 组装结果
	results := make([]PoolOrderResult, 0, len(orders))
	for _, o := range orders {
		elapsed := now.Sub(o.CreatedAt).Seconds()
		secondsLeft := int64(constants.PoolOrderTimeoutSec) - int64(elapsed)
		if secondsLeft < 0 {
			secondsLeft = 0
		}
		results = append(results, PoolOrderResult{
			Order:       o,
			SecondsLeft: secondsLeft,
		})
	}

	return results, int32(total), nil
}

// GrabOrder 司机抢单
func (s *PoolService) GrabOrder(ctx context.Context, orderID, driverID int64) error {
	// 1. 查司机信息
	driver, err := s.repo.GetDriverSByDriverId(ctx, driverID)
	if err != nil {
		return fmt.Errorf("get driver: %w", err)
	}

	// 2. 校验账号状态
	if driver.Status != constants.AccountStatusNormal {
		return errors.NewBusinessError(GrabErrAccountAbnormal, "您的账号状态异常，无法抢单")
	}

	// 3. 校验实名已通过
	realname, err := s.repo.GetRealnameByDriverId(ctx, driverID)
	if err != nil || realname.Status != constants.AuthStatusApproved {
		return errors.NewBusinessError(GrabErrRealname, "您的资质未审核通过，无法抢单")
	}

	// 4. 校验车辆认证已通过
	vehicle, err := s.repo.GetVehicleByDriverId(ctx, driverID)
	if err != nil || vehicle.Status != constants.AuthStatusApproved {
		return errors.NewBusinessError(GrabErrVehicle, "您的车辆认证未通过，无法抢单")
	}

	// 5. 校验工作状态 = 听单中
	if int8(driver.WorkStatus) != constants.WorkStatusListening {
		return errors.NewBusinessError(GrabErrNotListening, "请先开始听单后再抢单")
	}

	// 6. 校验无进行中订单
	hasOrder, err := s.repo.HasOngoingOrder(ctx, driverID)
	if err != nil {
		return fmt.Errorf("check ongoing order: %w", err)
	}
	if hasOrder {
		return errors.NewBusinessError(GrabErrHasOrder, "您当前有进行中的订单，无法抢单")
	}

	// 7. 校验当日接单上限
	if driver.DailyOrderLimit > 0 {
		today := time.Now().Format("2006-01-02")
		todayCount, err := s.repo.GetTodayOrderCount(ctx, driverID, today)
		if err != nil {
			return fmt.Errorf("get today order count: %w", err)
		}
		if todayCount >= driver.DailyOrderLimit {
			return errors.NewBusinessError(constants.GrabErrDailyLimit, "您已达到当日接单上限，无法继续抢单")
		}
	}

	// 8. 查订单信息
	var order model.Order
	if err := s.repo.GetDB().Where("order_id = ?", orderID).First(&order).Error; err != nil {
		return errors.NewBusinessError(constants.GrabErrOrderNotFound, "订单不存在")
	}

	// 9. 校验订单是否还在池中（status=0 AND driver_id=0）
	if order.Status != constants.OrderStatusPending || order.DriverId != 0 {
		return errors.NewBusinessError(constants.GrabErrNotInPool, "订单不在抢单池中")
	}

	// 10. 校验超时
	elapsed := time.Since(order.CreatedAt).Seconds()
	if elapsed > float64(constants.PoolOrderTimeoutSec) {
		return errors.NewBusinessError(constants.GrabErrTimeout, "抢单超时，订单已取消")
	}

	// 11. 城市匹配校验
	if order.CityId > 0 && driver.CityId > 0 && order.CityId != driver.CityId {
		return errors.NewBusinessError(constants.GrabErrCityMismatch, "该订单不在您的服务城市，无法抢单")
	}

	// 13. CAS 抢占：status=0 AND driver_id=0 → status=1(已派单), driver_id=司机ID
	result := s.repo.GetDB().Model(&model.Order{}).
		Where("order_id = ? AND status = ? AND driver_id = ?", orderID, constants.OrderStatusPending, 0).
		Updates(map[string]interface{}{
			"driver_id": driverID,
			"status":    constants.OrderStatusDispatched,
		})
	if result.Error != nil {
		return fmt.Errorf("grab order: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return errors.NewBusinessError(constants.GrabErrGrabbed, "订单已被抢走")
	}

	// 14. 写派单日志（dispatch_type=2 抢单）
	now := time.Now()
	s.repo.CreateDispatchLog(ctx, &model.DispatchLog{
		OrderId:      orderID,
		DriverId:     driverID,
		DispatchType: 2, // 抢单
		DispatchTime: now,
		ExpireTime:   now.Add(time.Duration(constants.PoolOrderTimeoutSec) * time.Second),
		Result:       1, // 接受
		ResponseTime: now,
	})

	// 15. 写状态日志
	s.repo.CreateStatusLog(ctx, &model.DriverStatusLog{
		DriverId:   driverID,
		FromStatus: constants.WorkStatusListening,
		ToStatus:   constants.WorkStatusListening,
		Reason:     fmt.Sprintf("抢单成功 order_id=%d", orderID),
	})

	return nil
}

// Grab 错误码（司机端抢单校验专用）
const (
	GrabErrAccountAbnormal = 5010 // 账号异常
	GrabErrRealname        = 5011 // 实名未通过
	GrabErrVehicle         = 5012 // 车辆未通过
	GrabErrNotListening    = 5013 // 未听单
	GrabErrHasOrder        = 5014 // 有进行中订单
)

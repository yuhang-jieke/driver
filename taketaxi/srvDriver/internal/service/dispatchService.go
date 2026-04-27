package service

import (
	"context"
	"fmt"
	"math"
	"time"

	"driver/taketaxi/common/constants"
	"driver/taketaxi/common/errors"
	"driver/taketaxi/pkg/config"
	"driver/taketaxi/srvDriver/internal/model"
	"driver/taketaxi/srvDriver/internal/repository"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type DispatchService struct {
	repo    *repository.DriverRepo
	mongoDb *mongo.Database
	cfg     *config.DispatchConfig
}

func NewDispatchService(repo *repository.DriverRepo, mongoDb *mongo.Database, cfg *config.DispatchConfig) *DispatchService {
	return &DispatchService{repo: repo, mongoDb: mongoDb, cfg: cfg}
}

// Dispatch 派单核心逻辑：根据 7 条条件过滤候选司机并推送订单
func (s *DispatchService) Dispatch(ctx context.Context, orderID int64, serviceType int32, originLat, originLng float64, passengerID int64) error {
	radiusKm := s.cfg.RadiusKm
	if radiusKm <= 0 {
		radiusKm = 3.0
	}
	minScore := s.cfg.MinServiceScore
	if minScore <= 0 {
		minScore = 60.0
	}
	radiusMeters := radiusKm * 1000.0

	// 1. 从 MongoDB 按距离查出 radius 内的候选司机（最近优先）
	candidates, err := s.findNearbyDrivers(ctx, originLat, originLng, radiusMeters)
	if err != nil {
		return fmt.Errorf("query nearby drivers: %w", err)
	}
	if len(candidates) == 0 {
		return errors.NewDispatchRejectError(constants.DispatchTooFar, "附近没有可用司机")
	}

	// 2. 逐司机过滤 7 条条件
	for _, c := range candidates {
		driverID := int64(c["driver_id"].(int64))

		// 查司机详情
		driver, err := s.repo.GetDriverSByDriverId(ctx, driverID)
		if err != nil {
			continue
		}

		// 条件 1: 工作状态 = 听单中(2)
		if int8(driver.WorkStatus) != constants.WorkStatusListening {
			s.logDispatchResult(ctx, orderID, driverID, constants.DispatchNotListening, originLat, originLng, c)
			continue
		}

		// 条件 6: 服务分 ≥ 最低限制分
		if driver.ServiceScore < minScore {
			s.logDispatchResult(ctx, orderID, driverID, constants.DispatchLowScore, originLat, originLng, c)
			continue
		}

		// 条件 4: 无进行中订单
		hasOrder, err := s.repo.HasOngoingOrder(ctx, driverID)
		if err != nil || hasOrder {
			if hasOrder {
				s.logDispatchResult(ctx, orderID, driverID, constants.DispatchHasOngoingOrder, originLat, originLng, c)
			}
			continue
		}

		// 条件 5: 未达当日接单上限
		if driver.DailyOrderLimit > 0 {
			today := time.Now().Format("2006-01-02")
			todayCount, err := s.repo.GetTodayOrderCount(ctx, driverID, today)
			if err != nil || todayCount >= driver.DailyOrderLimit {
				if todayCount >= driver.DailyOrderLimit {
					s.logDispatchResult(ctx, orderID, driverID, constants.DispatchDailyLimit, originLat, originLng, c)
				}
				continue
			}
		}

		// 条件 2: 距离校验（MongoDB 已按半径过滤，双重确认）
		distance := haversine(originLat, originLng, c["lat"].(float64), c["lng"].(float64))
		if distance > radiusMeters {
			s.logDispatchResult(ctx, orderID, driverID, constants.DispatchTooFar, originLat, originLng, c)
			continue
		}

		// 条件 3: 车型一致（service_type: 1-快车 2-特惠快车）
		vehicle, err := s.repo.GetVehicleByDriverIdAndStatus(ctx, driverID)
		if err != nil || int8(vehicle.ServiceType) != int8(serviceType) {
			if vehicle != nil {
				s.logDispatchResult(ctx, orderID, driverID, constants.DispatchVehicleMismatch, originLat, originLng, c)
			}
			continue
		}

		// 条件 7: CAS 抢占订单（status=0 AND driver_id=0 → status=1）
		affected, err := s.claimOrder(ctx, orderID, driverID)
		if err != nil || affected == 0 {
			// 订单已被抢走或状态不对，跳过
			continue
		}

		// 派单成功，写日志
		s.logDispatchResult(ctx, orderID, driverID, 0, originLat, originLng, c)
		return nil
	}

	return errors.NewDispatchRejectError(constants.DispatchTooFar, "附近没有符合条件的司机")
}

// RejectOrder 司机拒绝接单（CAS 校验：已派单 → 待派单，重置 driver_id）
func (s *DispatchService) RejectOrder(ctx context.Context, orderID, driverID int64) error {
	// CAS: 只有 status=1(已派单) AND driver_id 匹配的订单才能拒绝
	result := s.repo.GetDB().Model(&model.Order{}).
		Where("order_id = ? AND status = ? AND driver_id = ?", orderID, constants.OrderStatusDispatched, driverID).
		Updates(map[string]interface{}{
			"status":    constants.OrderStatusPending,
			"driver_id": 0,
		})
	if result.Error != nil {
		return fmt.Errorf("reject order: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return errors.NewDispatchRejectError(1, "订单状态已变更，无法拒绝")
	}

	// 写派单日志（rejectCode=0 表示成功，这里用于记录拒绝行为）
	s.logDispatchResult(ctx, orderID, driverID, 0, 0, 0, bson.M{"note": "司机拒绝接单"})

	// 重新派单给下一个候选司机（排除拒绝的司机）
	s.retryDispatch(ctx, orderID, driverID)

	return nil
}

// CancelOrder 司机主动取消订单（status=2/3/4 → status=6 已取消）
func (s *DispatchService) CancelOrder(ctx context.Context, orderID, driverID int64, reason string) error {
	// 查订单校验
	var order model.Order
	if err := s.repo.GetDB().Where("order_id = ? AND driver_id = ?", orderID, driverID).First(&order).Error; err != nil {
		return errors.NewDispatchRejectError(1, "订单不存在")
	}

	// 校验状态：只能取消已接单、已到达、行程中的订单
	if order.Status < constants.OrderStatusAccepted || order.Status > constants.OrderStatusInTrip {
		return errors.NewDispatchRejectError(1, "当前订单状态不允许取消")
	}

	now := time.Now()

	// CAS 取消：status IN (2,3,4) AND driver_id 匹配 → status=6(已取消)
	result := s.repo.GetDB().Model(&model.Order{}).
		Where("order_id = ? AND driver_id = ? AND status IN ?", orderID, driverID,
			[]int8{constants.OrderStatusAccepted, constants.OrderStatusArrived, constants.OrderStatusInTrip}).
		Updates(map[string]interface{}{
			"status":        constants.OrderStatusCancelled,
			"cancel_by":     2, // 司机取消
			"cancel_reason": reason,
			"cancel_time":   now,
		})
	if result.Error != nil {
		return fmt.Errorf("cancel order: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return errors.NewDispatchRejectError(1, "订单状态已变更，无法取消")
	}

	// 写派单日志
	s.repo.CreateDispatchLog(ctx, &model.DispatchLog{
		OrderId:      orderID,
		DriverId:     driverID,
		DispatchType: 2, // 抢单
		DispatchTime: now,
		Result:       3, // 取消
		ResponseTime: now,
		RejectReason: reason,
	})

	// 写状态日志
	s.repo.CreateStatusLog(ctx, &model.DriverStatusLog{
		DriverId:   driverID,
		FromStatus: order.Status,
		ToStatus:   constants.OrderStatusCancelled,
		Reason:     fmt.Sprintf("司机主动取消 order_id=%d reason=%s", orderID, reason),
	})

	return nil
}

// retryDispatch 重新派单（跳过指定 driverID 的司机）
func (s *DispatchService) retryDispatch(ctx context.Context, orderID, excludeDriverID int64) {
	// 1. 查询订单，确认状态为 Pending
	var order model.Order
	if err := s.repo.GetDB().Where("order_id = ? AND status = ?", orderID, constants.OrderStatusPending).First(&order).Error; err != nil {
		fmt.Printf("[retryDispatch] order %d not found or status changed, err=%v\n", orderID, err)
		return
	}

	// 2. 执行 Dispatch，排除刚拒绝的司机
	err := s.dispatchWithExclude(ctx, orderID, int32(order.ServiceType), order.OriginLat, order.OriginLng, order.PassengerId, excludeDriverID)
	if err != nil {
		fmt.Printf("[retryDispatch] order %d re-dispatch failed: %v\n", orderID, err)
	} else {
		fmt.Printf("[retryDispatch] order %d re-dispatch success\n", orderID)
	}
}

// dispatchWithExclude 派单，但排除指定司机
func (s *DispatchService) dispatchWithExclude(ctx context.Context, orderID int64, serviceType int32, originLat, originLng float64, passengerID, excludeDriverID int64) error {
	radiusKm := s.cfg.RadiusKm
	if radiusKm <= 0 {
		radiusKm = 3.0
	}
	minScore := s.cfg.MinServiceScore
	if minScore <= 0 {
		minScore = 60.0
	}
	radiusMeters := radiusKm * 1000.0

	// 1. 从 MongoDB 按距离查出 radius 内的候选司机
	candidates, err := s.findNearbyDrivers(ctx, originLat, originLng, radiusMeters)
	if err != nil {
		return fmt.Errorf("query nearby drivers: %w", err)
	}
	if len(candidates) == 0 {
		return errors.NewDispatchRejectError(constants.DispatchTooFar, "附近没有可用司机")
	}

	// 2. 逐司机过滤 7 条条件
	for _, c := range candidates {
		driverID := int64(c["driver_id"].(int64))

		// 排除刚拒绝的司机
		if driverID == excludeDriverID {
			continue
		}

		// 查司机详情
		driver, err := s.repo.GetDriverSByDriverId(ctx, driverID)
		if err != nil {
			continue
		}

		// 条件 1: 工作状态 = 听单中(2)
		if int8(driver.WorkStatus) != constants.WorkStatusListening {
			s.logDispatchResult(ctx, orderID, driverID, constants.DispatchNotListening, originLat, originLng, c)
			continue
		}

		// 条件 6: 服务分 ≥ 最低限制分
		if driver.ServiceScore < minScore {
			s.logDispatchResult(ctx, orderID, driverID, constants.DispatchLowScore, originLat, originLng, c)
			continue
		}

		// 条件 4: 无进行中订单
		hasOrder, err := s.repo.HasOngoingOrder(ctx, driverID)
		if err != nil || hasOrder {
			if hasOrder {
				s.logDispatchResult(ctx, orderID, driverID, constants.DispatchHasOngoingOrder, originLat, originLng, c)
			}
			continue
		}

		// 条件 5: 未达当日接单上限
		if driver.DailyOrderLimit > 0 {
			today := time.Now().Format("2006-01-02")
			todayCount, err := s.repo.GetTodayOrderCount(ctx, driverID, today)
			if err != nil || todayCount >= driver.DailyOrderLimit {
				if todayCount >= driver.DailyOrderLimit {
					s.logDispatchResult(ctx, orderID, driverID, constants.DispatchDailyLimit, originLat, originLng, c)
				}
				continue
			}
		}

		// 条件 2: 距离校验
		distance := haversine(originLat, originLng, c["lat"].(float64), c["lng"].(float64))
		if distance > radiusMeters {
			s.logDispatchResult(ctx, orderID, driverID, constants.DispatchTooFar, originLat, originLng, c)
			continue
		}

		// 条件 3: 车型一致
		vehicle, err := s.repo.GetVehicleByDriverIdAndStatus(ctx, driverID)
		if err != nil || int8(vehicle.ServiceType) != int8(serviceType) {
			if vehicle != nil {
				s.logDispatchResult(ctx, orderID, driverID, constants.DispatchVehicleMismatch, originLat, originLng, c)
			}
			continue
		}

		// 条件 7: CAS 抢占订单
		affected, err := s.claimOrder(ctx, orderID, driverID)
		if err != nil || affected == 0 {
			continue
		}

		// 派单成功，写日志
		s.logDispatchResult(ctx, orderID, driverID, 0, originLat, originLng, c)
		return nil
	}

	return errors.NewDispatchRejectError(constants.DispatchTooFar, "附近没有符合条件的司机")
}

// AcceptOrder 司机接单（CAS 校验：已派单 → 已接单）
func (s *DispatchService) AcceptOrder(ctx context.Context, orderID, driverID int64) error {
	fmt.Printf("[AcceptOrder] called: orderID=%d driverID=%d\n", orderID, driverID)
	result := s.repo.GetDB().Model(&model.Order{}).
		Where("order_id = ? AND status = ? AND driver_id = ?", orderID, constants.OrderStatusDispatched, driverID).
		Update("status", constants.OrderStatusAccepted)
	fmt.Printf("[AcceptOrder] RowsAffected=%d Error=%v\n", result.RowsAffected, result.Error)
	if result.Error != nil {
		return fmt.Errorf("accept order: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return errors.NewDispatchRejectError(1, "订单状态已变更，无法接单")
	}

	var order model.Order
	if err := s.repo.GetDB().Where("order_id = ?", orderID).First(&order).Error; err != nil {
		return fmt.Errorf("query order: %w", err)
	}

	now := time.Now()
	if err := s.repo.GetDB().Exec(
		"INSERT INTO trip_service (order_id, driver_id, passenger_id, accept_time) VALUES (?, ?, ?, ?)",
		orderID, driverID, order.PassengerId, now,
	).Error; err != nil {
		fmt.Printf("[AcceptOrder] create trip_service failed: order_id=%d driver_id=%d err=%v\n", orderID, driverID, err)
		return fmt.Errorf("create trip service: %w", err)
	}
	return nil
}

// DriverArrive 司机到达上车点（CAS 校验：已接单 → 已到达，距离≤30米）
func (s *DispatchService) DriverArrive(ctx context.Context, orderID, driverID int64) error {
	arriveRadius := s.cfg.ArriveCheckRadius
	if arriveRadius <= 0 {
		arriveRadius = 30.0
	}

	// 查订单
	var order model.Order
	if err := s.repo.GetDB().Where("order_id = ? AND status = ? AND driver_id = ?", orderID, constants.OrderStatusAccepted, driverID).First(&order).Error; err != nil {
		return errors.NewDispatchRejectError(1, "订单状态已变更，无法确认到达")
	}

	// 查司机实时位置
	if s.mongoDb == nil {
		return fmt.Errorf("mongodb not available")
	}
	mongoCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var loc struct {
		Lat float64 `bson:"lat"`
		Lng float64 `bson:"lng"`
	}
	err := s.mongoDb.Collection("driver_local").FindOne(mongoCtx, bson.M{"driver_id": driverID}).Decode(&loc)
	if err != nil {
		return errors.NewDispatchRejectError(1, "无法获取司机位置")
	}

	// 校验距离
	distance := haversine(order.OriginLat, order.OriginLng, loc.Lat, loc.Lng)
	if distance > arriveRadius {
		return errors.NewDispatchRejectError(1, fmt.Sprintf("您距离上车点还有%d米，请靠近后再确认", int(distance)))
	}

	// CAS: 2(已接单) → 3(已到达)
	result := s.repo.GetDB().Model(&model.Order{}).
		Where("order_id = ? AND status = ? AND driver_id = ?", orderID, constants.OrderStatusAccepted, driverID).
		Update("status", constants.OrderStatusArrived)
	if result.Error != nil {
		return fmt.Errorf("update order status: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return errors.NewDispatchRejectError(1, "订单状态已变更，无法确认到达")
	}

	s.repo.GetDB().Model(&model.TripService{}).
		Where("order_id = ? AND driver_id = ?", orderID, driverID).
		Update("arrive_time", time.Now())
	return nil
}

// VerifyPassengerPhone 验证乘客手机号后四位（仅校验，不变更状态）
func (s *DispatchService) VerifyPassengerPhone(ctx context.Context, orderID, driverID int64, phoneLast4 string) error {
	var order model.Order
	if err := s.repo.GetDB().Where("order_id = ? AND status = ? AND driver_id = ?", orderID, constants.OrderStatusArrived, driverID).First(&order).Error; err != nil {
		return errors.NewDispatchRejectError(1, "订单状态已变更，无法验证")
	}
	if len(order.PassengerMobile) < 4 {
		return errors.NewDispatchRejectError(constants.ErrCodePhoneMismatch, "订单手机号格式异常")
	}
	mobileLast4 := order.PassengerMobile[len(order.PassengerMobile)-4:]
	if mobileLast4 != phoneLast4 {
		return errors.NewDispatchRejectError(constants.ErrCodePhoneMismatch, "乘客手机号后四位不正确")
	}
	return nil
}

// StartTrip 开始行程（CAS 校验：已到达 → 行程中）
func (s *DispatchService) StartTrip(ctx context.Context, orderID, driverID int64) error {
	result := s.repo.GetDB().Model(&model.Order{}).
		Where("order_id = ? AND status = ? AND driver_id = ?", orderID, constants.OrderStatusArrived, driverID).
		Update("status", constants.OrderStatusInTrip)
	if result.Error != nil {
		return fmt.Errorf("start trip: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return errors.NewDispatchRejectError(1, "订单状态已变更，无法开始行程")
	}

	s.repo.GetDB().Model(&model.TripService{}).
		Where("order_id = ? AND driver_id = ?", orderID, driverID).
		Update("start_time", time.Now())
	return nil
}

// EndTrip 到达目的地（CAS 校验：行程中 → 已完成，距离≤30米）
func (s *DispatchService) EndTrip(ctx context.Context, orderID, driverID int64) error {
	endTripRadius := s.cfg.EndTripCheckRadius
	if endTripRadius <= 0 {
		endTripRadius = 30.0
	}

	// 查订单
	var order model.Order
	if err := s.repo.GetDB().Where("order_id = ? AND status = ? AND driver_id = ?", orderID, constants.OrderStatusInTrip, driverID).First(&order).Error; err != nil {
		return errors.NewDispatchRejectError(1, "订单状态已变更，无法结束行程")
	}

	// 查行程记录
	var trip model.TripService
	if err := s.repo.GetDB().Where("order_id = ? AND driver_id = ?", orderID, driverID).First(&trip).Error; err != nil {
		return errors.NewDispatchRejectError(1, "行程记录不存在")
	}

	// 查司机实时位置
	if s.mongoDb == nil {
		return fmt.Errorf("mongodb not available")
	}
	mongoCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var loc struct {
		Lat float64 `bson:"lat"`
		Lng float64 `bson:"lng"`
	}
	err := s.mongoDb.Collection("driver_local").FindOne(mongoCtx, bson.M{"driver_id": driverID}).Decode(&loc)
	if err != nil {
		return errors.NewDispatchRejectError(1, "无法获取司机位置")
	}

	// 校验距离（司机位置 vs 订单目的地）
	distance := haversine(loc.Lat, loc.Lng, order.DestLat, order.DestLng)
	if distance > endTripRadius {
		return errors.NewDispatchRejectError(1, fmt.Sprintf("您距离目的地还有%d米，请到达后再点击结束", int(distance)))
	}

	now := time.Now()

	// 计算里程和时长
	actualDistance := int(haversine(order.OriginLat, order.OriginLng, order.DestLat, order.DestLng))
	tripDuration := 0
	if !trip.StartTime.IsZero() {
		tripDuration = int(now.Sub(trip.StartTime).Seconds())
	}

	// CAS: 4(行程中) → 5(已完成)
	result := s.repo.GetDB().Model(&model.Order{}).
		Where("order_id = ? AND status = ? AND driver_id = ?", orderID, constants.OrderStatusInTrip, driverID).
		Updates(map[string]interface{}{
			"status":           constants.OrderStatusCompleted,
			"actual_distance":  actualDistance,
			"actual_duration":  tripDuration,
		})
	if result.Error != nil {
		return fmt.Errorf("end trip: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return errors.NewDispatchRejectError(1, "订单状态已变更，无法结束行程")
	}

	// 更新行程记录
	s.repo.GetDB().Model(&model.TripService{}).
		Where("order_id = ? AND driver_id = ?", orderID, driverID).
		Updates(map[string]interface{}{
			"end_time":      now,
			"trip_duration": tripDuration,
			"trip_distance": actualDistance,
		})

	// 写状态日志
	s.repo.CreateStatusLog(ctx, &model.DriverStatusLog{
		DriverId:   driverID,
		FromStatus: constants.OrderStatusInTrip,
		ToStatus:   constants.OrderStatusCompleted,
		Reason:     fmt.Sprintf("到达目的地 order_id=%d distance=%dm", orderID, int(distance)),
	})

	return nil
}

// findNearbyDrivers 从 MongoDB 查找半径内的司机，按距离升序
func (s *DispatchService) findNearbyDrivers(ctx context.Context, lat, lng, maxDistance float64) ([]bson.M, error) {
	if s.mongoDb == nil {
		return nil, fmt.Errorf("mongodb not available")
	}

	collection := s.mongoDb.Collection("driver_local")

	// $nearSphere 不需要显式创建 2dsphere 索引即可工作
	filter := bson.M{
		"loc": bson.M{
			"$nearSphere": bson.M{
				"$geometry": bson.M{
					"type":        "Point",
					"coordinates": []float64{lng, lat},
				},
				"$maxDistance": maxDistance,
			},
		},
		// 只查最近 5 分钟内上报过位置的司机
		"updated_at": bson.M{
			"$gte": time.Now().Unix() - 300,
		},
	}

	opts := options.Find().SetLimit(50)

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []bson.M
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	return results, nil
}

// claimOrder CAS 抢占订单：status=0(待派单) AND driver_id=0 → status=1(已派单)
func (s *DispatchService) claimOrder(ctx context.Context, orderID, driverID int64) (int64, error) {
	result := s.repo.GetDB().WithContext(ctx).
		Model(&model.Order{}).
		Where("order_id = ? AND status = ? AND driver_id = 0", orderID, constants.OrderStatusPending).
		Updates(map[string]interface{}{
			"driver_id": driverID,
			"status":    constants.OrderStatusDispatched,
		})
	if result.Error != nil {
		return 0, result.Error
	}
	return result.RowsAffected, nil
}

// logDispatchResult 记录派单日志
func (s *DispatchService) logDispatchResult(ctx context.Context, orderID, driverID int64, rejectCode int, originLat, originLng float64, candidate bson.M) {
	driverLat, _ := candidate["lat"].(float64)
	driverLng, _ := candidate["lng"].(float64)
	distance := int(haversine(originLat, originLng, driverLat, driverLng))

	if rejectCode != 0 {
		// 派单失败，写状态日志
		s.repo.CreateStatusLog(ctx, &model.DriverStatusLog{
			DriverId:   driverID,
			FromStatus: constants.WorkStatusListening,
			ToStatus:   constants.WorkStatusListening,
			Reason:     fmt.Sprintf("派单失败 order_id=%d code=%d distance=%dm", orderID, rejectCode, distance),
		})
		return
	}

	// 派单成功，写状态日志
	s.repo.CreateStatusLog(ctx, &model.DriverStatusLog{
		DriverId:   driverID,
		FromStatus: constants.WorkStatusListening,
		ToStatus:   constants.WorkStatusListening,
		Reason:     fmt.Sprintf("派单成功 order_id=%d distance=%dm", orderID, distance),
	})
}

// haversine 计算两点间距离(米)
func haversine(lat1, lng1, lat2, lng2 float64) float64 {
	const R = 6371000 // 地球半径(米)
	dLat := (lat2 - lat1) * math.Pi / 180
	dLng := (lng2 - lng1) * math.Pi / 180
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180)*math.Cos(lat2*math.Pi/180)*
			math.Sin(dLng/2)*math.Sin(dLng/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return R * c
}

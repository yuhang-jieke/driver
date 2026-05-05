package service

import (
	"context"

	driver "driver/taketaxi/common/kitexGen"
	"driver/taketaxi/srvDriver/internal/model"
	"driver/taketaxi/srvDriver/internal/repository"
)

type DriverService struct{ repo *repository.DriverRepo }

func NewDriverService(repo *repository.DriverRepo) *DriverService {
	return &DriverService{repo: repo}
}

func (s *DriverService) Create(ctx context.Context, req *driver.CreateDriverReq) (*driver.CreateDriverResp, error) {
	m := &model.Driver{Name: req.Name}
	return &driver.CreateDriverResp{Id: int64(m.ID)}, s.repo.Create(ctx, m)
}
func (s *DriverService) Get(ctx context.Context, req *driver.GetDriverReq) (*driver.GetDriverResp, error) {
	m, err := s.repo.GetByID(ctx, uint(req.Id))
	if err != nil {
		return nil, err
	}
	return &driver.GetDriverResp{Id: int64(m.ID), Name: m.Name, Status: int32(m.Status)}, nil
}
func (s *DriverService) List(ctx context.Context, req *driver.ListDriverReq) (*driver.ListDriverResp, error) {
	list, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}
	var items []*driver.DriverItem
	for _, m := range list {
		items = append(items, &driver.DriverItem{Id: int64(m.ID), Name: m.Name, Status: int32(m.Status)})
	}
	return &driver.ListDriverResp{Items: items}, nil
}
func (s *DriverService) Update(ctx context.Context, req *driver.UpdateDriverReq) (*driver.UpdateDriverResp, error) {
	m, err := s.repo.GetByID(ctx, uint(req.Id))
	if err != nil {
		return nil, err
	}
	if req.Name != "" {
		m.Name = req.Name
	}
	return &driver.UpdateDriverResp{Success: true}, s.repo.Update(ctx, m)
}
func (s *DriverService) Delete(ctx context.Context, req *driver.DeleteDriverReq) (*driver.DeleteDriverResp, error) {
	return &driver.DeleteDriverResp{Success: true}, s.repo.Delete(ctx, uint(req.Id))
}

func (s *DriverService) DriverDetails(ctx context.Context, req *driver.DriverDetailsReq) (*driver.DriverDetailsResp, error) {
	order, err := s.repo.GetOrderByID(ctx, req.OrderId)
	if err != nil {
		return nil, err
	}

	trajectories, err := s.repo.GetTrajectoriesByOrderID(ctx, req.OrderId)
	if err != nil {
		return nil, err
	}

	var tripList []*driver.TripTrajectory
	for _, t := range trajectories {
		tripList = append(tripList, &driver.TripTrajectory{
			Id:             t.Id,
			TripId:         t.TripId,
			TrajectoryData: t.TrajectoryData,
			PointCount:     int64(t.PointCount),
			StartTime:      t.StartTime.Format("2006-01-02 15:04:05"),
			EndTime:        t.EndTime.Format("2006-01-02 15:04:05"),
			Distance:       int64(t.Distance),
			FileUrl:        t.FileUrl,
		})
	}

	return &driver.DriverDetailsResp{
		OrderId:          order.OrderId,
		OrderNo:          order.OrderNo,
		OrderType:        int64(order.OrderType),
		ServiceType:      int64(order.ServiceType),
		Status:           int64(order.Status),
		DriverId:         order.DriverId,
		PassengerId:      order.PassengerId,
		PassengerMobile:  order.PassengerMobile,
		PassengerName:    order.PassengerName,
		OriginAddress:    order.OriginAddress,
		OriginLat:        float32(order.OriginLat),
		OriginLng:        float32(order.OriginLng),
		OriginPoi:        order.OriginPoi,
		DestAddress:      order.DestAddress,
		DestLat:          float32(order.DestLat),
		DestLng:          float32(order.DestLng),
		DestPoi:          order.DestPoi,
		EstimateDistance: int64(order.EstimateDistance),
		EstimateDuration: int64(order.EstimateDuration),
		EstimateFee:      float32(order.EstimateFee),
		ActualDistance:   int64(order.ActualDistance),
		ActualDuration:   int64(order.ActualDuration),
		ActualFee:        float32(order.ActualFee),
		BaseFee:          float32(order.BaseFee),
		DistanceFee:      float32(order.DistanceFee),
		DurationFee:      float32(order.DurationFee),
		WaitFee:          float32(order.WaitFee),
		CouponId:         order.CouponId,
		CouponAmount:     float32(order.CouponAmount),
		PayStatus:        int64(order.PayStatus),
		PayType:          int64(order.PayType),
		PayTime:          order.PayTime.Format("2006-01-02 15:04:05"),
		CancelReason:     order.CancelReason,
		CancelBy:         int64(order.CancelBy),
		CancelTime:       order.CancelTime.Format("2006-01-02 15:04:05"),
		CityId:           order.CityId,
		List:             tripList,
	}, nil
}

var orderStatusText = map[int8]string{
	0: "待派单",
	1: "已派单",
	2: "司机已到达",
	3: "行程中",
	4: "已完成",
	5: "已取消",
}

func (s *DriverService) DriverOrderList(ctx context.Context, req *driver.DriverOrderListReq) (*driver.DriverOrderListResp, error) {
	page := int(req.Page)
	pageSize := int(req.PageSize)
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	orders, total, err := s.repo.ListOrdersByDriverID(ctx, req.DriverId, page, pageSize)
	if err != nil {
		return nil, err
	}

	var items []*driver.OrderListItem
	for _, o := range orders {
		statusText := orderStatusText[o.Status]
		if statusText == "" {
			statusText = "未知"
		}
		items = append(items, &driver.OrderListItem{
			OrderNo:     o.OrderNo,
			OriginPoi:   o.OriginPoi,
			DestPoi:     o.DestPoi,
			StatusText:  statusText,
			EstimateFee: int64(o.EstimateFee * 100), // 元转分
			CreatedAt:   o.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	return &driver.DriverOrderListResp{
		Total:    int32(total),
		Page:     int32(page),
		PageSize: int32(pageSize),
		Items:    items,
	}, nil
}


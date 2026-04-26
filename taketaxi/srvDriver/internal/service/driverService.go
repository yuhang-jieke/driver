package service

import (
	"context"
	"errors"
	"time"

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

// GetProfile 查询司机个人信息与接单统计
func (s *DriverService) GetProfile(ctx context.Context, req *driver.GetDriverProfileReq) (*driver.GetDriverProfileResp, error) {
	// 参数校验
	if req.DriverId <= 0 {
		return nil, errors.New("invalid driver_id")
	}

	// 解析日期参数，默认当天
	var queryDate time.Time
	if req.Date == "" {
		queryDate = time.Now()
	} else {
		parsed, err := time.Parse("2006-01-02", req.Date)
		if err != nil {
			return nil, errors.New("invalid date format")
		}
		queryDate = parsed
	}

	// 校验日期不能为未来
	if queryDate.After(time.Now()) {
		return nil, errors.New("date cannot be in the future")
	}

	// 校验天数参数
	days := req.Days
	if days <= 0 {
		days = 1
	}
	if days > 30 {
		return nil, errors.New("days cannot exceed 30")
	}

	// 查询个人信息
	profile, err := s.repo.GetDriverProfile(ctx, req.DriverId)
	if err != nil {
		return nil, err
	}

	// 计算日期范围
	endDate := time.Date(queryDate.Year(), queryDate.Month(), queryDate.Day(), 0, 0, 0, 0, time.Local)
	startDate := endDate.AddDate(0, 0, -int(days)+1)

	// 查询接单统计
	stats, err := s.repo.GetOrderStats(ctx, req.DriverId, startDate, endDate)
	if err != nil {
		return nil, err
	}

	return &driver.GetDriverProfileResp{
		PersonalInfo: &driver.PersonalInfo{
			Nickname:     profile.Nickname,
			Avatar:       profile.Avatar,
			ServiceScore: profile.ServiceScore,
			OrderCount:   int32(profile.OrderCount),
			Phone:        profile.Phone,
			Plate:        profile.Plate,
			Car:          profile.Car,
			Online:       profile.Online,
			Status:       profile.Status,
		},
		OrderStats: &driver.OrderStats{
			OrderCount:     int32(stats.OrderCount),
			Income:         stats.TotalIncome,
			OnlineDuration: int32(stats.OnlineDuration),
		},
		VerifyStatus: int32(profile.VerifyStatus),
	}, nil
}

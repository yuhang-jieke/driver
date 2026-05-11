package service

import (
	"context"
	"time"

	driver "driver/taketaxi/common/kitexGen"
	"driver/taketaxi/pkg/errcode"
)

// ========== 个人信息 & 收入 ==========

// GetProfile 查询司机个人信息与接单统计
func (s *DriverService) GetProfile(ctx context.Context, req *driver.GetDriverProfileReq) (*driver.GetDriverProfileResp, error) {
	if req.DriverId <= 0 {
		return nil, errcode.New(errcode.ErrInvalidDriverID)
	}

	// 解析日期参数，默认当天
	var queryDate time.Time
	if req.Date == "" {
		queryDate = time.Now()
	} else {
		parsed, err := time.Parse("2006-01-02", req.Date)
		if err != nil {
			return nil, errcode.New(errcode.ErrInvalidDate)
		}
		queryDate = parsed
	}

	// 校验日期不能为未来
	if queryDate.After(time.Now()) {
		return nil, errcode.New(errcode.ErrInvalidDate)
	}

	// 校验天数参数
	days := req.Days
	if days <= 0 {
		days = 1
	}
	if days > 30 {
		return nil, errcode.New(errcode.ErrInvalidDays)
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

	// 查询认证信息（供前端回填已提交的图片和表单数据）
	verifyInfo, err := s.repo.GetVerifyInfo(ctx, req.DriverId)
	if err != nil {
		return nil, err
	}

	resp := &driver.GetDriverProfileResp{
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
	}

	// 填充认证信息（仅在有数据时填充）
	if verifyInfo.Realname.Status > 0 {
		resp.RealnameInfo = &driver.RealnameInfo{
			RealName:       verifyInfo.Realname.RealName,
			IdCardFrontUrl: verifyInfo.Realname.IdCardFrontUrl,
			IdCardBackUrl:  verifyInfo.Realname.IdCardBackUrl,
			Status:         int32(verifyInfo.Realname.Status),
		}
	}
	if verifyInfo.License.Status > 0 {
		resp.LicenseInfo = &driver.LicenseInfo{
			LicenseNo:   verifyInfo.License.LicenseNo,
			LicenseType: verifyInfo.License.LicenseType,
			LicenseUrl:  verifyInfo.License.LicenseUrl,
			Status:      int32(verifyInfo.License.Status),
		}
	}
	if verifyInfo.Vehicle.Status > 0 {
		resp.VehicleInfo = &driver.VehicleInfo{
			PlateNo:           verifyInfo.Vehicle.PlateNo,
			VehicleBrand:      verifyInfo.Vehicle.VehicleBrand,
			VehicleModel:      verifyInfo.Vehicle.VehicleModel,
			VehicleColor:      verifyInfo.Vehicle.VehicleColor,
			SeatCount:         int32(verifyInfo.Vehicle.SeatCount),
			DrivingLicenseUrl: verifyInfo.Vehicle.DrivingLicenseUrl,
			VehiclePhotoUrl:   verifyInfo.Vehicle.VehiclePhotoUrl,
			Status:            int32(verifyInfo.Vehicle.Status),
		}
	}

	return resp, nil
}

// GetIncome 查询收入明细
func (s *DriverService) GetIncome(ctx context.Context, req *driver.GetDriverIncomeReq) (*driver.GetDriverIncomeResp, error) {
	if req.DriverId <= 0 {
		return nil, errcode.New(errcode.ErrInvalidDriverID)
	}

	// 根据 period 计算日期范围（日=今天，周=本周一，月=本月1号）
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)

	var startDate, endDate time.Time
	switch req.Period {
	case "today":
		startDate = todayStart
		endDate = todayStart
	case "week":
		weekday := int(now.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		startDate = todayStart.AddDate(0, 0, -weekday+1) // 本周一
		endDate = todayStart
	case "month":
		startDate = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.Local) // 本月1号
		endDate = todayStart
	default:
		startDate = todayStart
		endDate = todayStart
	}

	// 查询汇总统计
	stats, err := s.repo.GetOrderStats(ctx, req.DriverId, startDate, endDate)
	if err != nil {
		return nil, err
	}

	// 趋势图与汇总统计共用同一日期范围
	dailyData, err := s.repo.GetDailyIncome(ctx, req.DriverId, startDate, endDate)
	if err != nil {
		return nil, err
	}

	// 将查询结果转为 map：date → income
	incomeMap := make(map[string]float64)
	for _, d := range dailyData {
		incomeMap[d.Date] = d.Income
	}

	// 逐日遍历，补全缺失日期
	var trend []*driver.DailyIncome
	for d := startDate; !d.After(endDate); d = d.AddDate(0, 0, 1) {
		dateStr := d.Format("2006-01-02")
		income := incomeMap[dateStr]
		trend = append(trend, &driver.DailyIncome{
			Date:   dateStr,
			Income: int64(income),
		})
	}

	return &driver.GetDriverIncomeResp{
		Summary: &driver.IncomeSummary{
			OrderCount:     int32(stats.OrderCount),
			Income:         int64(stats.TotalIncome),
			OnlineDuration: int32(stats.OnlineDuration),
		},
		Trend: trend,
	}, nil
}

// UpdateProfile 更新司机个人资料
func (s *DriverService) UpdateProfile(ctx context.Context, req *driver.UpdateProfileReq) (*driver.UpdateProfileResp, error) {
	if req.DriverId <= 0 {
		return &driver.UpdateProfileResp{Success: false, Message: errcode.ErrInvalidDriverID.Message()}, errcode.New(errcode.ErrInvalidDriverID)
	}

	updates := make(map[string]interface{})
	if req.Nickname != "" {
		updates["nickname"] = req.Nickname
	}
	if req.Avatar != "" {
		updates["avatar"] = req.Avatar
	}
	if req.Gender > 0 {
		updates["gender"] = req.Gender
	}
	if len(updates) == 0 {
		return &driver.UpdateProfileResp{Success: false, Message: errcode.ErrNoUpdateFields.Message()}, errcode.New(errcode.ErrNoUpdateFields)
	}

	if err := s.repo.UpdateProfile(ctx, req.DriverId, updates); err != nil {
		return &driver.UpdateProfileResp{Success: false, Message: errcode.ErrInternal.Message()}, err
	}
	return &driver.UpdateProfileResp{Success: true, Message: errcode.Success.Message()}, nil
}

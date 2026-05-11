package repository

import (
	"context"
	"driver/taketaxi/srvDriver/internal/model"
	"time"
)

// ==================== 个人信息 & 统计 & 认证查询 ====================

// GetDriverProfile 查询司机个人信息（多表 JOIN + 计算）
func (r *DriverRepo) GetDriverProfile(ctx context.Context, driverID int64) (*model.DriverProfileResult, error) {
	var driver model.DriverS
	if err := r.db.WithContext(ctx).
		Select("nickname, avatar, service_score, order_count, verify_status, mobile, status, last_online_at").
		Where("driver_id = ?", driverID).
		First(&driver).Error; err != nil {
		return nil, err
	}

	// 查询已审核通过的车辆信息
	var vehicle model.DriverVehicle
	r.db.WithContext(ctx).
		Select("plate_no, vehicle_model, vehicle_color").
		Where("driver_id = ? AND status = 2", driverID).
		First(&vehicle)

	// 在线判定：最后活动时间在 5 分钟内视为在线
	online := false
	if driver.LastOnlineAt != "" {
		if t, err := time.Parse("2006-01-02 15:04:05", driver.LastOnlineAt); err == nil {
			online = time.Since(t) < 5*time.Minute
		}
	}

	// 状态码映射
	var status string
	switch driver.Status {
	case 1:
		status = "idle"
	case 2:
		status = "busy"
	default:
		status = "offline"
	}

	// 组装车辆描述
	carDesc := ""
	if vehicle.VehicleModel != "" {
		carDesc = vehicle.VehicleModel
		if vehicle.VehicleColor != "" {
			carDesc += " · " + vehicle.VehicleColor
		}
	}

	return &model.DriverProfileResult{
		Nickname:     driver.Nickname,
		Avatar:       driver.Avatar,
		ServiceScore: driver.ServiceScore,
		OrderCount:   driver.OrderCount,
		VerifyStatus: driver.VerifyStatus,
		Phone:        driver.Mobile,
		Plate:        vehicle.PlateNo,
		Car:          carDesc,
		Online:       online,
		Status:       status,
	}, nil
}

// GetOrderStats 查询指定日期范围内的接单统计数据
func (r *DriverRepo) GetOrderStats(ctx context.Context, driverID int64, startDate, endDate time.Time) (*model.OrderStatsResult, error) {
	var result model.OrderStatsResult
	err := r.db.WithContext(ctx).
		Model(&model.DriverStatisticsSummary{}).
		Select("COALESCE(SUM(order_count), 0) as order_count, COALESCE(SUM(total_income), 0) as total_income, COALESCE(SUM(online_duration), 0) as online_duration").
		Where("driver_id = ? AND stat_date BETWEEN ? AND ?", driverID, startDate, endDate).
		Scan(&result).Error
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// GetDailyIncome 查询每日收入（趋势图数据）
func (r *DriverRepo) GetDailyIncome(ctx context.Context, driverID int64, startDate, endDate time.Time) ([]model.DailyIncomeResult, error) {
	var results []model.DailyIncomeResult
	err := r.db.WithContext(ctx).
		Model(&model.DriverStatisticsSummary{}).
		Select("DATE(stat_date) as date, COALESCE(total_income, 0) as income").
		Where("driver_id = ? AND stat_date BETWEEN ? AND ?", driverID, startDate, endDate).
		Order("stat_date ASC").
		Scan(&results).Error
	if err != nil {
		return nil, err
	}
	return results, nil
}

// GetVerifyInfo 查询所有认证信息（3次独立查询，互不影响）
func (r *DriverRepo) GetVerifyInfo(ctx context.Context, driverID int64) (*model.VerifyInfoResult, error) {
	result := &model.VerifyInfoResult{}

	// 实名认证
	var rn model.DriverRealname
	if err := r.db.WithContext(ctx).Where("driver_id = ?", driverID).
		Select("real_name, id_card_front_url, id_card_back_url, status").
		First(&rn).Error; err == nil {
		result.Realname = model.RealnameInfoResult{
			RealName:       rn.RealName,
			IdCardFrontUrl: rn.IdCardFrontUrl,
			IdCardBackUrl:  rn.IdCardBackUrl,
			Status:         rn.Status,
		}
	}

	// 驾驶证认证
	var lic model.DriverLicense
	if err := r.db.WithContext(ctx).Where("driver_id = ?", driverID).
		Select("license_no, license_type, license_url, status").
		First(&lic).Error; err == nil {
		result.License = model.LicenseInfoResult{
			LicenseNo:   lic.LicenseNo,
			LicenseType: lic.LicenseType,
			LicenseUrl:  lic.LicenseUrl,
			Status:      lic.Status,
		}
	}

	// 车辆信息
	var v model.DriverVehicle
	if err := r.db.WithContext(ctx).Where("driver_id = ?", driverID).
		Select("plate_no, vehicle_brand, vehicle_model, vehicle_color, seat_count, driving_license_url, vehicle_photo_url, status").
		First(&v).Error; err == nil {
		result.Vehicle = model.VehicleInfoResult{
			PlateNo:           v.PlateNo,
			VehicleBrand:      v.VehicleBrand,
			VehicleModel:      v.VehicleModel,
			VehicleColor:      v.VehicleColor,
			SeatCount:         v.SeatCount,
			DrivingLicenseUrl: v.DrivingLicenseUrl,
			VehiclePhotoUrl:   v.VehiclePhotoUrl,
			Status:            v.Status,
		}
	}

	return result, nil
}

// GetDriverRealname 查询司机实名认证信息（用于提现资格校验）
func (r *DriverRepo) GetDriverRealname(ctx context.Context, driverID int64) (*model.DriverRealname, error) {
	var rn model.DriverRealname
	err := r.db.WithContext(ctx).Where("driver_id = ?", driverID).First(&rn).Error
	if err != nil {
		return nil, err
	}
	return &rn, nil
}

package repository

import (
	"context"
	"driver/taketaxi/srvDriver/internal/model"
	"errors"

	"gorm.io/gorm"
)

// ==================== 认证信息 Upsert（先查后建/更新） ====================

// UpdateRealname 更新/创建实名认证信息（Upsert）
func (r *DriverRepo) UpdateRealname(ctx context.Context, rn *model.DriverRealname) error {
	var existing model.DriverRealname
	err := r.db.WithContext(ctx).Where("driver_id = ?", rn.DriverId).First(&existing).Error
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		return r.db.WithContext(ctx).
			Omit("birthday", "expire_date", "verify_time").
			Create(rn).Error
	}
	updates := make(map[string]interface{})
	if rn.RealName != "" {
		updates["real_name"] = rn.RealName
	}
	if rn.IdCardNo != "" {
		updates["id_card_no"] = rn.IdCardNo
	}
	if rn.IdCardFrontUrl != "" {
		updates["id_card_front_url"] = rn.IdCardFrontUrl
	}
	if rn.IdCardBackUrl != "" {
		updates["id_card_back_url"] = rn.IdCardBackUrl
	}
	if rn.Gender > 0 {
		updates["gender"] = rn.Gender
	}
	if !rn.Birthday.IsZero() {
		updates["birthday"] = rn.Birthday
	}
	if rn.Address != "" {
		updates["address"] = rn.Address
	}
	if rn.Nation != "" {
		updates["nation"] = rn.Nation
	}
	if !rn.ExpireDate.IsZero() {
		updates["expire_date"] = rn.ExpireDate
	}
	updates["status"] = int8(1)
	if len(updates) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Model(&model.DriverRealname{}).
		Where("driver_id = ?", rn.DriverId).
		Updates(updates).Error
}

// UpdateLicense 更新/创建驾驶证认证信息（Upsert）
func (r *DriverRepo) UpdateLicense(ctx context.Context, lic *model.DriverLicense) error {
	var existing model.DriverLicense
	err := r.db.WithContext(ctx).Where("driver_id = ?", lic.DriverId).First(&existing).Error
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		return r.db.WithContext(ctx).
			Omit("first_issue_date", "issue_date", "expire_date", "verify_time").
			Create(lic).Error
	}
	updates := make(map[string]interface{})
	if lic.LicenseNo != "" {
		updates["license_no"] = lic.LicenseNo
	}
	if lic.LicenseType != "" {
		updates["license_type"] = lic.LicenseType
	}
	if lic.LicenseUrl != "" {
		updates["license_url"] = lic.LicenseUrl
	}
	if !lic.FirstIssueDate.IsZero() {
		updates["first_issue_date"] = lic.FirstIssueDate
	}
	if !lic.IssueDate.IsZero() {
		updates["issue_date"] = lic.IssueDate
	}
	if !lic.ExpireDate.IsZero() {
		updates["expire_date"] = lic.ExpireDate
	}
	updates["status"] = int8(1)
	if len(updates) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Model(&model.DriverLicense{}).
		Where("driver_id = ?", lic.DriverId).
		Updates(updates).Error
}

// UpdateVehicle 更新/创建车辆信息（Upsert）
func (r *DriverRepo) UpdateVehicle(ctx context.Context, v *model.DriverVehicle) error {
	var existing model.DriverVehicle
	err := r.db.WithContext(ctx).Where("driver_id = ?", v.DriverId).First(&existing).Error
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		return r.db.WithContext(ctx).
			Omit("register_date", "verify_time").
			Create(v).Error
	}
	updates := make(map[string]interface{})
	if v.PlateNo != "" {
		updates["plate_no"] = v.PlateNo
	}
	if v.VehicleModel != "" {
		updates["vehicle_model"] = v.VehicleModel
	}
	if v.VehicleBrand != "" {
		updates["vehicle_brand"] = v.VehicleBrand
	}
	if v.VehicleColor != "" {
		updates["vehicle_color"] = v.VehicleColor
	}
	if v.VehicleColorCode != "" {
		updates["vehicle_color_code"] = v.VehicleColorCode
	}
	if v.SeatCount > 0 {
		updates["seat_count"] = v.SeatCount
	}
	if v.DrivingLicenseUrl != "" {
		updates["driving_license_url"] = v.DrivingLicenseUrl
	}
	if v.VehiclePhotoUrl != "" {
		updates["vehicle_photo_url"] = v.VehiclePhotoUrl
	}
	updates["status"] = int8(1)
	if len(updates) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Model(&model.DriverVehicle{}).
		Where("driver_id = ?", v.DriverId).
		Updates(updates).Error
}

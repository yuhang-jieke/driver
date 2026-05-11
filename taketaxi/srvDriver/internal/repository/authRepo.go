package repository

import (
	"context"
	"driver/taketaxi/srvDriver/internal/model"
	"errors"
	"time"
)

// ==================== 账号安全（手机号/密码/资料更新） ====================

// GetDriverByID 通过 driver_id 查询完整司机记录
func (r *DriverRepo) GetDriverByID(ctx context.Context, driverID int64) (*model.DriverS, error) {
	var driver model.DriverS
	if err := r.db.WithContext(ctx).Where("driver_id = ?", driverID).First(&driver).Error; err != nil {
		return nil, err
	}
	return &driver, nil
}

// GetDriverByMobile 通过手机号反查司机（重置密码场景）
func (r *DriverRepo) GetDriverByMobile(ctx context.Context, mobile string) (*model.DriverS, error) {
	var driver model.DriverS
	if err := r.db.WithContext(ctx).Where("mobile = ?", mobile).First(&driver).Error; err != nil {
		return nil, err
	}
	return &driver, nil
}

// UpdateMobile 更新手机号（同时记录更新时间）
func (r *DriverRepo) UpdateMobile(ctx context.Context, driverID int64, mobile string) error {
	now := time.Now().Format("2006-01-02 15:04:05")
	return r.db.WithContext(ctx).Model(&model.DriverS{}).
		Where("driver_id = ?", driverID).
		Updates(map[string]interface{}{
			"mobile":            mobile,
			"mobile_updated_at": now,
		}).Error
}

// UpdatePassword 更新密码（同时记录更新时间）
func (r *DriverRepo) UpdatePassword(ctx context.Context, driverID int64, password string) error {
	now := time.Now().Format("2006-01-02 15:04:05")
	return r.db.WithContext(ctx).Model(&model.DriverS{}).
		Where("driver_id = ?", driverID).
		Updates(map[string]interface{}{
			"password":            password,
			"password_updated_at": now,
		}).Error
}

// UpdateProfile 增量更新个人资料（只更新传入的非空字段）
func (r *DriverRepo) UpdateProfile(ctx context.Context, driverID int64, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return nil
	}
	result := r.db.WithContext(ctx).Model(&model.DriverS{}).
		Where("driver_id = ?", driverID).
		Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("driver not found")
	}
	return nil
}

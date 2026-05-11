package repository

import (
	"context"
	"driver/taketaxi/srvDriver/internal/model"
	"errors"

	"gorm.io/gorm"
)

// ==================== 银行卡操作 ====================

// GetBankCard 查询司机当前有效的银行卡绑定
func (r *DriverRepo) GetBankCard(ctx context.Context, driverID int64) (*model.DriverBankCard, error) {
	var card model.DriverBankCard
	err := r.db.WithContext(ctx).Where("driver_id = ? AND status = 1", driverID).First(&card).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 未绑定 ≠ 错误
		}
		return nil, err
	}
	return &card, nil
}

// CreateBankCard 创建新的银行卡绑定记录
func (r *DriverRepo) CreateBankCard(ctx context.Context, card *model.DriverBankCard) error {
	return r.db.WithContext(ctx).Create(card).Error
}

// UpdateBankCard 更新已有银行卡信息
func (r *DriverRepo) UpdateBankCard(ctx context.Context, driverID int64, updates map[string]interface{}) error {
	result := r.db.WithContext(ctx).Model(&model.DriverBankCard{}).
		Where("driver_id = ?", driverID).
		Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("bank card not found")
	}
	return nil
}

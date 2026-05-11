package repository

import (
	"context"
	"driver/taketaxi/srvDriver/internal/model"
	"errors"

	"gorm.io/gorm"
)

// ==================== 钱包操作 ====================

// GetWallet 查询钱包信息（含懒创建）
func (r *DriverRepo) GetWallet(ctx context.Context, driverID int64) (*model.DriverWallet, error) {
	var wallet model.DriverWallet
	err := r.db.WithContext(ctx).Where("driver_id = ?", driverID).First(&wallet).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			wallet = model.DriverWallet{DriverId: driverID}
			if createErr := r.db.WithContext(ctx).Create(&wallet).Error; createErr != nil {
				return nil, createErr
			}
			return &wallet, nil
		}
		return nil, err
	}
	return &wallet, nil
}

// UpdateWallet 更新钱包余额（乐观锁）
func (r *DriverRepo) UpdateWallet(ctx context.Context, wallet *model.DriverWallet) error {
	result := r.db.WithContext(ctx).Model(&model.DriverWallet{}).
		Where("id = ? AND version = ?", wallet.Id, wallet.Version).
		Updates(map[string]interface{}{
			"balance":        wallet.Balance,
			"frozen_amount":  wallet.FrozenAmount,
			"total_income":   wallet.TotalIncome,
			"total_withdraw": wallet.TotalWithdraw,
			"version":        gorm.Expr("version + 1"),
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("wallet version conflict, please retry")
	}
	return nil
}

// CreateWalletTransactionLog 创建钱包流水记录
func (r *DriverRepo) CreateWalletTransactionLog(ctx context.Context, log *model.WalletTransactionLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

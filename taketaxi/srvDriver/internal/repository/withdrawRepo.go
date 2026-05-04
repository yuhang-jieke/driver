package repository

import (
	"context"
	"driver/taketaxi/srvDriver/internal/model"
	"errors"
	"time"

	"gorm.io/gorm"
)

// ==================== 提现 & 收入明细 ====================

// CreateWithdrawRecord 创建提现记录
func (r *DriverRepo) CreateWithdrawRecord(ctx context.Context, record *model.DriverWithdrawRecord) error {
	return r.db.WithContext(ctx).Create(record).Error
}

// GetTodayWithdrawCount 查询当日有效提现次数（只统计处理中+成功，失败不计入每日限额）
func (r *DriverRepo) GetTodayWithdrawCount(ctx context.Context, driverID int64) (int64, error) {
	var count int64
	today := time.Now().Format("2006-01-02")
	err := r.db.WithContext(ctx).Model(&model.DriverWithdrawRecord{}).
		Where("driver_id = ? AND DATE(apply_time) = ? AND status IN ?", driverID, today,
			[]int8{model.WithdrawStatusPending, model.WithdrawStatusSuccess}).
		Count(&count).Error
	return count, err
}

// FindRecentPendingWithdraw 幂等校验：查找指定时间窗口内的相同金额待处理提现
// 未找到时返回 (nil, nil)，不将 ErrRecordNotFound 当作错误
func (r *DriverRepo) FindRecentPendingWithdraw(ctx context.Context, driverID int64, amount int64, window time.Duration) (*model.DriverWithdrawRecord, error) {
	var record model.DriverWithdrawRecord
	cutoff := time.Now().Add(-window)
	err := r.db.WithContext(ctx).
		Where("driver_id = ? AND amount = ? AND status = ? AND apply_time >= ?",
			driverID, amount, model.WithdrawStatusPending, cutoff).
		First(&record).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 无重复，不是错误
		}
		return nil, err
	}
	return &record, nil
}

// GetWithdrawRecords 分页查询提现记录
func (r *DriverRepo) GetWithdrawRecords(ctx context.Context, driverID int64, page, pageSize int32) ([]*model.DriverWithdrawRecord, int64, error) {
	var records []*model.DriverWithdrawRecord
	var total int64

	db := r.db.WithContext(ctx).Model(&model.DriverWithdrawRecord{}).Where("driver_id = ?", driverID)
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := db.Order("apply_time DESC").Offset(int(offset)).Limit(int(pageSize)).Find(&records).Error
	return records, total, err
}

// GetWithdrawRecordByNo 根据提现单号查询提现记录（回调场景）
func (r *DriverRepo) GetWithdrawRecordByNo(ctx context.Context, withdrawNo string) (*model.DriverWithdrawRecord, error) {
	var record model.DriverWithdrawRecord
	err := r.db.WithContext(ctx).Where("withdraw_no = ?", withdrawNo).First(&record).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}

// UpdateWithdrawStatusByNo 根据提现单号更新状态（仅当当前状态为 pending 时更新，保证幂等）
func (r *DriverRepo) UpdateWithdrawStatusByNo(ctx context.Context, withdrawNo string, updates map[string]interface{}) error {
	result := r.db.WithContext(ctx).Model(&model.DriverWithdrawRecord{}).
		Where("withdraw_no = ? AND status = ?", withdrawNo, model.WithdrawStatusPending).
		Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("withdraw record not found or already processed")
	}
	return nil
}

// GetIncomeDetail 查询收入分类明细（GROUP BY type）
// period: today/week/month
func (r *DriverRepo) GetIncomeDetail(ctx context.Context, driverID int64, period string) ([]model.IncomeDetailResult, error) {
	var results []model.IncomeDetailResult
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)

	var startDate time.Time
	switch period {
	case "today":
		startDate = today
	case "week":
		// 中国习惯：周一为一周开始
		weekday := int(now.Weekday())
		if weekday == 0 {
			weekday = 7 // 周日转为7
		}
		startDate = today.AddDate(0, 0, -weekday+1)
	case "month":
		startDate = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.Local)
	default:
		startDate = today // 默认今天
	}
	endDate := today.Add(24*time.Hour - time.Second)

	typeNames := map[int8]string{
		1: "基础车费", 2: "奖励", 3: "空驶补偿", 4: "高速费", 5: "其他",
	}

	rows, err := r.db.WithContext(ctx).Model(&model.DriverIncomeLog{}).
		Select("type, COALESCE(SUM(amount), 0) as amount, COUNT(*) as count").
		Where("driver_id = ? AND created_at BETWEEN ? AND ? AND amount > 0", driverID, startDate, endDate).
		Group("type").
		Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var typeCode int8
		var amount float64
		var count int
		if err := rows.Scan(&typeCode, &amount, &count); err != nil {
			continue
		}
		name, ok := typeNames[typeCode]
		if !ok {
			name = "其他"
		}
		results = append(results, model.IncomeDetailResult{
			TypeCode: typeCode,
			TypeName: name,
			Amount:   amount,
			Count:    count,
		})
	}
	return results, nil
}

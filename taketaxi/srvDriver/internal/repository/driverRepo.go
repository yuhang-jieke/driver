package repository

import (
	"context"
	"driver/taketaxi/pkg/database"
	"driver/taketaxi/srvDriver/internal/model"
	"time"

	"gorm.io/gorm"
)

type DriverRepo struct{ db *gorm.DB }

func NewDriverRepo(db *gorm.DB) *DriverRepo {
	if db == nil {
		db, _ = database.NewDB(nil)
	}
	return &DriverRepo{db: db}
}

func (r *DriverRepo) Create(ctx context.Context, m *model.Driver) error {
	return r.db.WithContext(ctx).Create(m).Error
}
func (r *DriverRepo) GetByID(ctx context.Context, id uint) (*model.Driver, error) {
	var m model.Driver
	if err := r.db.WithContext(ctx).First(&m, id).Error; err != nil {
		return nil, err
	}
	return &m, nil
}
func (r *DriverRepo) List(ctx context.Context) ([]*model.Driver, error) {
	var list []*model.Driver
	return list, r.db.WithContext(ctx).Find(&list).Error
}
func (r *DriverRepo) Update(ctx context.Context, m *model.Driver) error {
	return r.db.WithContext(ctx).Save(m).Error
}
func (r *DriverRepo) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.Driver{}, id).Error
}

// DriverProfileResult 个人信息查询结果
type DriverProfileResult struct {
	Nickname     string
	Avatar       string
	ServiceScore float64
	OrderCount   int
	VerifyStatus int8
}

// OrderStatsResult 接单统计结果
type OrderStatsResult struct {
	OrderCount     int
	TotalIncome    float64
	OnlineDuration int
}

// GetDriverProfile 查询司机个人信息
func (r *DriverRepo) GetDriverProfile(ctx context.Context, driverID int64) (*DriverProfileResult, error) {
	var driver model.DriverS
	if err := r.db.WithContext(ctx).
		Select("nickname, avatar, service_score, order_count, verify_status").
		Where("driver_id = ?", driverID).
		First(&driver).Error; err != nil {
		return nil, err
	}
	return &DriverProfileResult{
		Nickname:     driver.Nickname,
		Avatar:       driver.Avatar,
		ServiceScore: driver.ServiceScore,
		OrderCount:   driver.OrderCount,
		VerifyStatus: driver.VerifyStatus,
	}, nil
}

// GetOrderStats 查询接单统计（指定日期范围内汇总）
func (r *DriverRepo) GetOrderStats(ctx context.Context, driverID int64, startDate, endDate time.Time) (*OrderStatsResult, error) {
	var result OrderStatsResult
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

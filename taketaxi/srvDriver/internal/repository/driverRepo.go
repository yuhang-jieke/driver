package repository

import (
	"context"
	"driver/taketaxi/pkg/database"
	"driver/taketaxi/srvDriver/internal/model"

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

func (r *DriverRepo) GetOrderByID(ctx context.Context, orderId int64) (*model.Order, error) {
	var order model.Order
	if err := r.db.WithContext(ctx).Where("order_id = ?", orderId).First(&order).Error; err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *DriverRepo) GetTrajectoriesByOrderID(ctx context.Context, orderId int64) ([]*model.TripTrajectory, error) {
	var list []*model.TripTrajectory
	if err := r.db.WithContext(ctx).Where("order_id = ?", orderId).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *DriverRepo) ListOrdersByDriverID(ctx context.Context, driverID int64, page, pageSize int) ([]*model.Order, int64, error) {
	var total int64
	if err := r.db.WithContext(ctx).Model(&model.Order{}).Where("driver_id = ?", driverID).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []*model.Order
	offset := (page - 1) * pageSize
	if err := r.db.WithContext(ctx).Where("driver_id = ?", driverID).Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}



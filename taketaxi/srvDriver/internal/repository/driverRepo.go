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

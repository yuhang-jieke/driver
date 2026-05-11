package repository

import (
	"context"
	"driver/taketaxi/srvDriver/internal/model"

	"gorm.io/gorm"
)

// DriverRepo 司机数据仓库，持有 GORM 数据库连接实例
// 方法按业务域拆分到不同文件：
//   - repo.go          : 结构体定义 + 基础 CRUD
//   - profileRepo.go   : 个人信息 & 统计 & 认证查询
//   - authRepo.go      : 账号安全（手机号/密码/资料更新）
//   - verifyRepo.go    : 认证信息 Upsert（实名/驾驶证/车辆）
//   - bankCardRepo.go  : 银行卡操作
//   - walletRepo.go    : 钱包操作
//   - withdrawRepo.go  : 提现 & 收入明细
type DriverRepo struct{ db *gorm.DB }

// NewDriverRepo 创建数据仓库实例
func NewDriverRepo(db *gorm.DB) *DriverRepo {
	return &DriverRepo{db: db}
}

// ==================== 基础 CRUD（drivers 表通用操作） ====================

// Create 创建司机记录（INSERT）
func (r *DriverRepo) Create(ctx context.Context, m *model.Driver) error {
	return r.db.WithContext(ctx).Create(m).Error
}

// GetByID 根据主键 ID 查询单条司机记录
func (r *DriverRepo) GetByID(ctx context.Context, id uint) (*model.Driver, error) {
	var m model.Driver
	if err := r.db.WithContext(ctx).First(&m, id).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

// List 查询全部司机记录
func (r *DriverRepo) List(ctx context.Context) ([]*model.Driver, error) {
	var list []*model.Driver
	return list, r.db.WithContext(ctx).Find(&list).Error
}

// Update 全量更新司机记录
func (r *DriverRepo) Update(ctx context.Context, m *model.Driver) error {
	return r.db.WithContext(ctx).Save(m).Error
}

// Delete 根据 ID 物理删除司机记录
func (r *DriverRepo) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.Driver{}, id).Error
}

// RunInTx 在数据库事务中执行 fn，支持乐观锁重试场景
// fn 接收一个基于 tx 的新 DriverRepo，所有操作在同一个事务内
func (r *DriverRepo) RunInTx(ctx context.Context, fn func(txRepo *DriverRepo) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txRepo := &DriverRepo{db: tx}
		return fn(txRepo)
	})
}

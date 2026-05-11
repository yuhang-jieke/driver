package service

import (
	"context"

	driver "driver/taketaxi/common/kitexGen"
	"driver/taketaxi/srvDriver/internal/model"
	"driver/taketaxi/pkg/errcode"
)

// ========== 基础 CRUD（管理后台使用） ==========
// 注意：crudHandler.go / crudService.go 仅用于管理后台基础 CRUD，
// 个人中心新能力禁止进入此文件，应按子域拆分到对应 service 文件。

// Create 创建司机
func (s *DriverService) Create(ctx context.Context, req *driver.CreateDriverReq) (*driver.CreateDriverResp, error) {
	m := &model.Driver{Name: req.Name}
	if err := s.repo.Create(ctx, m); err != nil {
		return nil, errcode.NewWithDetail(errcode.ErrCreateRecord, err.Error())
	}
	return &driver.CreateDriverResp{Id: int64(m.ID)}, nil
}

// Get 查询单个司机
func (s *DriverService) Get(ctx context.Context, req *driver.GetDriverReq) (*driver.GetDriverResp, error) {
	if req.Id <= 0 {
		return nil, errcode.New(errcode.ErrInvalidDriverID)
	}
	m, err := s.repo.GetByID(ctx, uint(req.Id))
	if err != nil {
		return nil, errcode.New(errcode.ErrDriverNotFound)
	}
	return &driver.GetDriverResp{Id: int64(m.ID), Name: m.Name, Status: int32(m.Status)}, nil
}

// List 查询司机列表
func (s *DriverService) List(ctx context.Context, req *driver.ListDriverReq) (*driver.ListDriverResp, error) {
	list, err := s.repo.List(ctx)
	if err != nil {
		return nil, errcode.NewWithDetail(errcode.ErrInternal, err.Error())
	}
	var items []*driver.DriverItem
	for _, m := range list {
		items = append(items, &driver.DriverItem{Id: int64(m.ID), Name: m.Name, Status: int32(m.Status)})
	}
	return &driver.ListDriverResp{Items: items}, nil
}

// Update 更新司机（仅覆盖 Name 字段）
func (s *DriverService) Update(ctx context.Context, req *driver.UpdateDriverReq) (*driver.UpdateDriverResp, error) {
	if req.Id <= 0 {
		return nil, errcode.New(errcode.ErrInvalidDriverID)
	}
	m, err := s.repo.GetByID(ctx, uint(req.Id))
	if err != nil {
		return nil, errcode.New(errcode.ErrDriverNotFound)
	}
	if req.Name != "" {
		m.Name = req.Name
	}
	if err := s.repo.Update(ctx, m); err != nil {
		return nil, errcode.NewWithDetail(errcode.ErrInternal, err.Error())
	}
	return &driver.UpdateDriverResp{Success: true}, nil
}

// Delete 删除司机
func (s *DriverService) Delete(ctx context.Context, req *driver.DeleteDriverReq) (*driver.DeleteDriverResp, error) {
	if req.Id <= 0 {
		return nil, errcode.New(errcode.ErrInvalidDriverID)
	}
	if err := s.repo.Delete(ctx, uint(req.Id)); err != nil {
		return nil, errcode.NewWithDetail(errcode.ErrInternal, err.Error())
	}
	return &driver.DeleteDriverResp{Success: true}, nil
}

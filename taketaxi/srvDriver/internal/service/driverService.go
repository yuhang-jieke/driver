package service

import (
	"context"
	driver "driver/taketaxi/common/kitexGen"
	"driver/taketaxi/srvDriver/internal/model"
	"driver/taketaxi/srvDriver/internal/repository"
)

type DriverService struct{ repo *repository.DriverRepo }

func NewDriverService(repo *repository.DriverRepo) *DriverService {
	return &DriverService{repo: repo}
}

func (s *DriverService) Create(ctx context.Context, req *driver.CreateDriverReq) (*driver.CreateDriverResp, error) {
	m := &model.Driver{Name: req.Name}
	return &driver.CreateDriverResp{Id: int64(m.ID)}, s.repo.Create(ctx, m)
}
func (s *DriverService) Get(ctx context.Context, req *driver.GetDriverReq) (*driver.GetDriverResp, error) {
	m, err := s.repo.GetByID(ctx, uint(req.Id))
	if err != nil {
		return nil, err
	}
	return &driver.GetDriverResp{Id: int64(m.ID), Name: m.Name, Status: int32(m.Status)}, nil
}
func (s *DriverService) List(ctx context.Context, req *driver.ListDriverReq) (*driver.ListDriverResp, error) {
	list, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}
	var items []*driver.DriverItem
	for _, m := range list {
		items = append(items, &driver.DriverItem{Id: int64(m.ID), Name: m.Name, Status: int32(m.Status)})
	}
	return &driver.ListDriverResp{Items: items}, nil
}
func (s *DriverService) Update(ctx context.Context, req *driver.UpdateDriverReq) (*driver.UpdateDriverResp, error) {
	m, err := s.repo.GetByID(ctx, uint(req.Id))
	if err != nil {
		return nil, err
	}
	if req.Name != "" {
		m.Name = req.Name
	}
	return &driver.UpdateDriverResp{Success: true}, s.repo.Update(ctx, m)
}
func (s *DriverService) Delete(ctx context.Context, req *driver.DeleteDriverReq) (*driver.DeleteDriverResp, error) {
	return &driver.DeleteDriverResp{Success: true}, s.repo.Delete(ctx, uint(req.Id))
}

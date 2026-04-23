package handler

import (
	"context"
	driver "driver/taketaxi/common/kitexGen"
	"driver/taketaxi/srvDriver/internal/repository"
	"driver/taketaxi/srvDriver/internal/service"
)

type DriverHandler struct {
	driver.UnimplementedDriverServiceServer
	svc *service.DriverService
}

func NewDriverHandler(repo *repository.DriverRepo) *DriverHandler {
	return &DriverHandler{svc: service.NewDriverService(repo)}
}

func (h *DriverHandler) Create(ctx context.Context, req *driver.CreateDriverReq) (*driver.CreateDriverResp, error) {
	return h.svc.Create(ctx, req)
}
func (h *DriverHandler) Get(ctx context.Context, req *driver.GetDriverReq) (*driver.GetDriverResp, error) {
	return h.svc.Get(ctx, req)
}
func (h *DriverHandler) List(ctx context.Context, req *driver.ListDriverReq) (*driver.ListDriverResp, error) {
	return h.svc.List(ctx, req)
}
func (h *DriverHandler) Update(ctx context.Context, req *driver.UpdateDriverReq) (*driver.UpdateDriverResp, error) {
	return h.svc.Update(ctx, req)
}
func (h *DriverHandler) Delete(ctx context.Context, req *driver.DeleteDriverReq) (*driver.DeleteDriverResp, error) {
	return h.svc.Delete(ctx, req)
}

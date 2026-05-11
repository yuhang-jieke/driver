package handler

import (
	"context"
	driver "driver/taketaxi/common/kitexGen"
	"driver/taketaxi/pkg/logger"
	"time"

	"go.uber.org/zap"
)

// ========== 基础 CRUD（管理后台使用） ==========

// Create 创建司机
func (h *DriverHandler) Create(ctx context.Context, req *driver.CreateDriverReq) (*driver.CreateDriverResp, error) {
	start := time.Now()
	resp, err := h.svc.Create(ctx, req)
	if err != nil {
		logger.Error("gRPC Create failed", zap.String("method", "Create"), zap.String("name", req.Name), zap.Error(err))
		return nil, err
	}
	logger.Info("gRPC Create success", zap.String("method", "Create"), zap.Int64("id", resp.Id), zap.Duration("duration", time.Since(start)))
	return resp, nil
}

// Get 查询单个司机
func (h *DriverHandler) Get(ctx context.Context, req *driver.GetDriverReq) (*driver.GetDriverResp, error) {
	start := time.Now()
	resp, err := h.svc.Get(ctx, req)
	if err != nil {
		logger.Error("gRPC Get failed", zap.String("method", "Get"), zap.Int64("id", req.Id), zap.Error(err))
		return nil, err
	}
	logger.Info("gRPC Get success", zap.String("method", "Get"), zap.Int64("id", req.Id), zap.Duration("duration", time.Since(start)))
	return resp, nil
}

// List 查询司机列表
func (h *DriverHandler) List(ctx context.Context, req *driver.ListDriverReq) (*driver.ListDriverResp, error) {
	start := time.Now()
	resp, err := h.svc.List(ctx, req)
	if err != nil {
		logger.Error("gRPC List failed", zap.String("method", "List"), zap.Error(err))
		return nil, err
	}
	logger.Info("gRPC List success", zap.String("method", "List"), zap.Int("count", len(resp.Items)), zap.Duration("duration", time.Since(start)))
	return resp, nil
}

// Update 更新司机
func (h *DriverHandler) Update(ctx context.Context, req *driver.UpdateDriverReq) (*driver.UpdateDriverResp, error) {
	start := time.Now()
	resp, err := h.svc.Update(ctx, req)
	if err != nil {
		logger.Error("gRPC Update failed", zap.String("method", "Update"), zap.Int64("id", req.Id), zap.Error(err))
		return nil, err
	}
	logger.Info("gRPC Update success", zap.String("method", "Update"), zap.Int64("id", req.Id), zap.Duration("duration", time.Since(start)))
	return resp, nil
}

// Delete 删除司机
func (h *DriverHandler) Delete(ctx context.Context, req *driver.DeleteDriverReq) (*driver.DeleteDriverResp, error) {
	start := time.Now()
	resp, err := h.svc.Delete(ctx, req)
	if err != nil {
		logger.Error("gRPC Delete failed", zap.String("method", "Delete"), zap.Int64("id", req.Id), zap.Error(err))
		return nil, err
	}
	logger.Info("gRPC Delete success", zap.String("method", "Delete"), zap.Int64("id", req.Id), zap.Duration("duration", time.Since(start)))
	return resp, nil
}

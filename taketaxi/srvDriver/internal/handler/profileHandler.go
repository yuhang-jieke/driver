package handler

import (
	"context"
	driver "driver/taketaxi/common/kitexGen"
	"driver/taketaxi/pkg/logger"
	"time"

	"go.uber.org/zap"
)

// ========== 个人信息 ==========

// GetProfile 查询司机个人信息与接单统计
func (h *DriverHandler) GetProfile(ctx context.Context, req *driver.GetDriverProfileReq) (*driver.GetDriverProfileResp, error) {
	start := time.Now()
	resp, err := h.svc.GetProfile(ctx, req)
	if err != nil {
		logger.Error("gRPC GetProfile failed", zap.String("method", "GetProfile"), zap.Int64("driver_id", req.DriverId), zap.Error(err))
		return nil, err
	}
	logger.Info("gRPC GetProfile success", zap.String("method", "GetProfile"), zap.Int64("driver_id", req.DriverId), zap.Duration("duration", time.Since(start)))
	return resp, nil
}

// UpdateProfile 更新司机个人资料（昵称、头像、性别）
func (h *DriverHandler) UpdateProfile(ctx context.Context, req *driver.UpdateProfileReq) (*driver.UpdateProfileResp, error) {
	start := time.Now()
	resp, err := h.svc.UpdateProfile(ctx, req)
	if err != nil {
		logger.Error("gRPC UpdateProfile failed", zap.String("method", "UpdateProfile"), zap.Int64("driver_id", req.DriverId), zap.Error(err))
		return nil, err
	}
	logger.Info("gRPC UpdateProfile success", zap.String("method", "UpdateProfile"), zap.Int64("driver_id", req.DriverId), zap.Duration("duration", time.Since(start)))
	return resp, nil
}

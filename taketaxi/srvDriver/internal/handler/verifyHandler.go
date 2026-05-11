package handler

import (
	"context"
	driver "driver/taketaxi/common/kitexGen"
	"driver/taketaxi/pkg/logger"
	"time"

	"go.uber.org/zap"
)

// ========== 资料 & 认证 ==========

// UpdateRealname 提交/更新实名认证信息
func (h *DriverHandler) UpdateRealname(ctx context.Context, req *driver.UpdateRealnameReq) (*driver.UpdateRealnameResp, error) {
	start := time.Now()
	resp, err := h.svc.UpdateRealname(ctx, req)
	if err != nil {
		logger.Error("gRPC UpdateRealname failed", zap.String("method", "UpdateRealname"), zap.Int64("driver_id", req.DriverId), zap.Error(err))
		return nil, err
	}
	logger.Info("gRPC UpdateRealname success", zap.String("method", "UpdateRealname"), zap.Int64("driver_id", req.DriverId), zap.Duration("duration", time.Since(start)))
	return resp, nil
}

// UpdateLicense 提交/更新驾驶证认证信息
func (h *DriverHandler) UpdateLicense(ctx context.Context, req *driver.UpdateLicenseReq) (*driver.UpdateLicenseResp, error) {
	start := time.Now()
	resp, err := h.svc.UpdateLicense(ctx, req)
	if err != nil {
		logger.Error("gRPC UpdateLicense failed", zap.String("method", "UpdateLicense"), zap.Int64("driver_id", req.DriverId), zap.Error(err))
		return nil, err
	}
	logger.Info("gRPC UpdateLicense success", zap.String("method", "UpdateLicense"), zap.Int64("driver_id", req.DriverId), zap.Duration("duration", time.Since(start)))
	return resp, nil
}

// UpdateVehicle 提交/更新车辆信息
func (h *DriverHandler) UpdateVehicle(ctx context.Context, req *driver.UpdateVehicleReq) (*driver.UpdateVehicleResp, error) {
	start := time.Now()
	resp, err := h.svc.UpdateVehicle(ctx, req)
	if err != nil {
		logger.Error("gRPC UpdateVehicle failed", zap.String("method", "UpdateVehicle"), zap.Int64("driver_id", req.DriverId), zap.Error(err))
		return nil, err
	}
	logger.Info("gRPC UpdateVehicle success", zap.String("method", "UpdateVehicle"), zap.Int64("driver_id", req.DriverId), zap.Duration("duration", time.Since(start)))
	return resp, nil
}

package handler

import (
	"context"
	driver "driver/taketaxi/common/kitexGen"
	"driver/taketaxi/pkg/logger"
	"time"

	"go.uber.org/zap"
)

// ========== 账号安全 ==========

// ChangeMobile 修改手机号
func (h *DriverHandler) ChangeMobile(ctx context.Context, req *driver.ChangeMobileReq) (*driver.ChangeMobileResp, error) {
	start := time.Now()
	resp, err := h.svc.ChangeMobile(ctx, req)
	if err != nil {
		logger.Error("gRPC ChangeMobile failed", zap.String("method", "ChangeMobile"), zap.Int64("driver_id", req.DriverId), zap.Error(err))
		return nil, err
	}
	logger.Info("gRPC ChangeMobile success", zap.String("method", "ChangeMobile"), zap.Int64("driver_id", req.DriverId), zap.Duration("duration", time.Since(start)))
	return resp, nil
}

// ChangePassword 修改密码
func (h *DriverHandler) ChangePassword(ctx context.Context, req *driver.ChangePasswordReq) (*driver.ChangePasswordResp, error) {
	start := time.Now()
	resp, err := h.svc.ChangePassword(ctx, req)
	if err != nil {
		logger.Error("gRPC ChangePassword failed", zap.String("method", "ChangePassword"), zap.Int64("driver_id", req.DriverId), zap.Error(err))
		return nil, err
	}
	logger.Info("gRPC ChangePassword success", zap.String("method", "ChangePassword"), zap.Int64("driver_id", req.DriverId), zap.Duration("duration", time.Since(start)))
	return resp, nil
}

// ResetPassword 重置密码（忘记密码场景）
func (h *DriverHandler) ResetPassword(ctx context.Context, req *driver.ResetPasswordReq) (*driver.ResetPasswordResp, error) {
	start := time.Now()
	resp, err := h.svc.ResetPassword(ctx, req)
	if err != nil {
		logger.Error("gRPC ResetPassword failed", zap.String("method", "ResetPassword"), zap.String("mobile", req.Mobile), zap.Error(err))
		return nil, err
	}
	logger.Info("gRPC ResetPassword success", zap.String("method", "ResetPassword"), zap.String("mobile", req.Mobile), zap.Duration("duration", time.Since(start)))
	return resp, nil
}

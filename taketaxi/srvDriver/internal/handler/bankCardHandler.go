package handler

import (
	"context"
	driver "driver/taketaxi/common/kitexGen"
	"driver/taketaxi/pkg/logger"
	"time"

	"go.uber.org/zap"
)

// ========== 银行卡 ==========

// BindBankCard 绑定银行卡
func (h *DriverHandler) BindBankCard(ctx context.Context, req *driver.BindBankCardReq) (*driver.BindBankCardResp, error) {
	start := time.Now()
	resp, err := h.svc.BindBankCard(ctx, req)
	if err != nil {
		logger.Error("gRPC BindBankCard failed", zap.String("method", "BindBankCard"), zap.Int64("driver_id", req.DriverId), zap.Error(err))
		return nil, err
	}
	logger.Info("gRPC BindBankCard success", zap.String("method", "BindBankCard"), zap.Int64("driver_id", req.DriverId), zap.Duration("duration", time.Since(start)))
	return resp, nil
}

// GetBankCard 查询绑定的银行卡信息
func (h *DriverHandler) GetBankCard(ctx context.Context, req *driver.GetBankCardReq) (*driver.GetBankCardResp, error) {
	start := time.Now()
	resp, err := h.svc.GetBankCard(ctx, req)
	if err != nil {
		logger.Error("gRPC GetBankCard failed", zap.String("method", "GetBankCard"), zap.Int64("driver_id", req.DriverId), zap.Error(err))
		return nil, err
	}
	logger.Info("gRPC GetBankCard success", zap.String("method", "GetBankCard"), zap.Int64("driver_id", req.DriverId), zap.Duration("duration", time.Since(start)))
	return resp, nil
}

// UpdateBankCard 更换银行卡
func (h *DriverHandler) UpdateBankCard(ctx context.Context, req *driver.UpdateBankCardReq) (*driver.UpdateBankCardResp, error) {
	start := time.Now()
	resp, err := h.svc.UpdateBankCard(ctx, req)
	if err != nil {
		logger.Error("gRPC UpdateBankCard failed", zap.String("method", "UpdateBankCard"), zap.Int64("driver_id", req.DriverId), zap.Error(err))
		return nil, err
	}
	logger.Info("gRPC UpdateBankCard success", zap.String("method", "UpdateBankCard"), zap.Int64("driver_id", req.DriverId), zap.Duration("duration", time.Since(start)))
	return resp, nil
}

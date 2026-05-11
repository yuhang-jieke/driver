package handler

import (
	"context"
	driver "driver/taketaxi/common/kitexGen"
	"driver/taketaxi/pkg/logger"
	"time"

	"go.uber.org/zap"
)

// ========== 钱包 & 提现 & 收入 ==========

// GetWallet 查询钱包概览
func (h *DriverHandler) GetWallet(ctx context.Context, req *driver.GetWalletReq) (*driver.GetWalletResp, error) {
	start := time.Now()
	resp, err := h.svc.GetWallet(ctx, req)
	if err != nil {
		logger.Error("gRPC GetWallet failed", zap.String("method", "GetWallet"), zap.Int64("driver_id", req.DriverId), zap.Error(err))
		return nil, err
	}
	logger.Info("gRPC GetWallet success", zap.String("method", "GetWallet"), zap.Int64("driver_id", req.DriverId), zap.Duration("duration", time.Since(start)))
	return resp, nil
}

// GetWithdrawPage 查询提现页信息（规则 + 资格 + 银行卡摘要）
func (h *DriverHandler) GetWithdrawPage(ctx context.Context, req *driver.GetWithdrawPageReq) (*driver.GetWithdrawPageResp, error) {
	start := time.Now()
	resp, err := h.svc.GetWithdrawPage(ctx, req)
	if err != nil {
		logger.Error("gRPC GetWithdrawPage failed", zap.String("method", "GetWithdrawPage"), zap.Int64("driver_id", req.DriverId), zap.Error(err))
		return nil, err
	}
	logger.Info("gRPC GetWithdrawPage success", zap.String("method", "GetWithdrawPage"), zap.Int64("driver_id", req.DriverId), zap.Duration("duration", time.Since(start)))
	return resp, nil
}

// ApplyWithdraw 申请提现
func (h *DriverHandler) ApplyWithdraw(ctx context.Context, req *driver.ApplyWithdrawReq) (*driver.ApplyWithdrawResp, error) {
	start := time.Now()
	resp, err := h.svc.ApplyWithdraw(ctx, req)
	if err != nil {
		logger.Error("gRPC ApplyWithdraw failed", zap.String("method", "ApplyWithdraw"), zap.Int64("driver_id", req.DriverId), zap.Error(err))
		return nil, err
	}
	logger.Info("gRPC ApplyWithdraw success", zap.String("method", "ApplyWithdraw"), zap.Int64("driver_id", req.DriverId), zap.Duration("duration", time.Since(start)))
	return resp, nil
}

// GetWithdrawRecords 分页查询提现记录
func (h *DriverHandler) GetWithdrawRecords(ctx context.Context, req *driver.GetWithdrawRecordsReq) (*driver.GetWithdrawRecordsResp, error) {
	start := time.Now()
	resp, err := h.svc.GetWithdrawRecords(ctx, req)
	if err != nil {
		logger.Error("gRPC GetWithdrawRecords failed", zap.String("method", "GetWithdrawRecords"), zap.Int64("driver_id", req.DriverId), zap.Error(err))
		return nil, err
	}
	logger.Info("gRPC GetWithdrawRecords success", zap.String("method", "GetWithdrawRecords"), zap.Int64("driver_id", req.DriverId), zap.Duration("duration", time.Since(start)))
	return resp, nil
}

// GetIncomeDetail 查询收入分类明细
func (h *DriverHandler) GetIncomeDetail(ctx context.Context, req *driver.GetIncomeDetailReq) (*driver.GetIncomeDetailResp, error) {
	start := time.Now()
	resp, err := h.svc.GetIncomeDetail(ctx, req)
	if err != nil {
		logger.Error("gRPC GetIncomeDetail failed", zap.String("method", "GetIncomeDetail"), zap.Int64("driver_id", req.DriverId), zap.Error(err))
		return nil, err
	}
	logger.Info("gRPC GetIncomeDetail success", zap.String("method", "GetIncomeDetail"), zap.Int64("driver_id", req.DriverId), zap.Duration("duration", time.Since(start)))
	return resp, nil
}

// GetIncome 查询司机收入明细（含趋势图数据）
func (h *DriverHandler) GetIncome(ctx context.Context, req *driver.GetDriverIncomeReq) (*driver.GetDriverIncomeResp, error) {
	start := time.Now()
	resp, err := h.svc.GetIncome(ctx, req)
	if err != nil {
		logger.Error("gRPC GetIncome failed", zap.String("method", "GetIncome"), zap.Int64("driver_id", req.DriverId), zap.Error(err))
		return nil, err
	}
	logger.Info("gRPC GetIncome success", zap.String("method", "GetIncome"), zap.Int64("driver_id", req.DriverId), zap.Duration("duration", time.Since(start)))
	return resp, nil
}

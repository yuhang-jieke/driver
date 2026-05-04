// Package handler 实现 BFF（Backend For Frontend）层的 HTTP 处理器，
// 作为前端和 gRPC 服务之间的适配层（Adapter / Anti-Corruption Layer）。
//
// 核心职责：
//   1. 解析 HTTP 请求（JSON body、Query params、Path params）
//   2. 参数校验与类型转换（string → int64 等）
//   3. 调用 RPC Client 将请求转发给 srvDriver（gRPC）
//   4. 将 gRPC 响应转换为 HTTP JSON 响应返回前端
//
// 每个方法遵循统一模式：
//   c.ShouldBindJSON(&req) → 参数绑定+校验
//   h.client.XXX(ctx, &pbReq)    → 调用 gRPC
//   c.JSON(code, resp)           → 返回 JSON
//
// 错误处理规范：
//   - 400 Bad Request: 参数校验失败（缺少必填字段、格式错误）
//   - 500 Internal Server Error: gRPC 调用失败或内部错误
//   所有响应都通过 logger 记录，便于排查问题
package handler

import (
	"driver/taketaxi/pkg/logger"
	"net/http"
	"strconv"

	"driver/taketaxi/bffDriver/internal/rpcClient"
	pb "driver/taketaxi/common/kitexGen"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// DriverHandler BFF 层司机 HTTP 处理器
// 持有 RPC 客户端引用，所有请求都通过它转发到后端 gRPC 服务
type DriverHandler struct {
	client *rpcclient.DriverClient // gRPC 客户端封装
}

// NewDriverHandler 创建 Handler 实例，注入 RPC 客户端
func NewDriverHandler(client *rpcclient.DriverClient) *DriverHandler {
	return &DriverHandler{client: client}
}

// ==================== 基础 CRUD（管理后台接口） ====================

// Create 创建司机
// POST /api/v1/drivers
// 请求体: {"name": "张三"}
func (h *DriverHandler) Create(c *gin.Context) {
	var req struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("创建司机参数错误", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.client.Create(c.Request.Context(), &pb.CreateDriverReq{Name: req.Name})
	if err != nil {
		logger.Error("创建司机失败", zap.String("name", req.Name), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	logger.Info("创建司机成功", zap.Int64("id", resp.Id), zap.String("name", req.Name))
	c.JSON(http.StatusOK, gin.H{"id": resp.Id})
}

// List 查询司机列表
// GET /api/v1/drivers
func (h *DriverHandler) List(c *gin.Context) {
	resp, err := h.client.List(c.Request.Context(), &pb.ListDriverReq{})
	if err != nil {
		logger.Error("获取司机列表失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	logger.Info("获取司机列表成功", zap.Int("count", len(resp.Items)))
	c.JSON(http.StatusOK, gin.H{"drivers": resp.Items})
}

// Get 查询单个司机
// GET /api/v1/drivers/:id
func (h *DriverHandler) Get(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		logger.Warn("获取司机参数错误", zap.String("id", c.Param("id")), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	resp, err := h.client.Get(c.Request.Context(), &pb.GetDriverReq{Id: id})
	if err != nil {
		logger.Error("获取司机失败", zap.Int64("id", id), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	logger.Info("获取司机成功", zap.Int64("id", id))
	c.JSON(http.StatusOK, resp)
}

// Update 更新司机
// PUT /api/v1/drivers/:id
// 请求体: {"name": "新名称"}
func (h *DriverHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		logger.Warn("更新司机参数错误", zap.String("id", c.Param("id")), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("更新司机参数错误", zap.Int64("id", id), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.client.Update(c.Request.Context(), &pb.UpdateDriverReq{Id: id, Name: req.Name})
	if err != nil {
		logger.Error("更新司机失败", zap.Int64("id", id), zap.String("name", req.Name), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	logger.Info("更新司机成功", zap.Int64("id", id), zap.String("name", req.Name))
	c.JSON(http.StatusOK, resp)
}

// Delete 删除司机
// DELETE /api/v1/drivers/:id
func (h *DriverHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		logger.Warn("删除司机参数错误", zap.String("id", c.Param("id")), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	resp, err := h.client.Delete(c.Request.Context(), &pb.DeleteDriverReq{Id: id})
	if err != nil {
		logger.Error("删除司机失败", zap.Int64("id", id), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	logger.Info("删除司机成功", zap.Int64("id", id))
	c.JSON(http.StatusOK, resp)
}

// ==================== 个人信息 & 收入 ====================

// Profile 查询司机个人信息与接单统计
// GET /api/v1/driver/profile?driver_id=200000001&date=2026-04-28&days=7
//
// 参数说明：
//   - driver_id (必填): 司机ID
//   - date (可选): 查询日期，默认当天，格式 YYYY-MM-DD
//   - days (可选): 统计天数范围，默认1天，最大30天
func (h *DriverHandler) Profile(c *gin.Context) {
	driverIDStr := c.Query("driver_id")
	if driverIDStr == "" {
		logger.Warn("Profile 缺少 driver_id 参数")
		c.JSON(http.StatusBadRequest, gin.H{"error": "driver_id is required"})
		return
	}
	driverID, err := strconv.ParseInt(driverIDStr, 10, 64)
	if err != nil {
		logger.Warn("Profile 参数错误", zap.String("driver_id", driverIDStr), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid driver_id"})
		return
	}

	date := c.Query("date")
	daysStr := c.Query("days")
	var days int32 = 1
	if daysStr != "" {
		d, err := strconv.ParseInt(daysStr, 10, 32)
		if err != nil {
			logger.Warn("Profile 参数错误", zap.String("days", daysStr), zap.Error(err))
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid days"})
			return
		}
		days = int32(d)
	}

	resp, err := h.client.GetProfile(c.Request.Context(), &pb.GetDriverProfileReq{
		DriverId: driverID,
		Date:     date,
		Days:     days,
	})
	if err != nil {
		logger.Error("Profile 查询失败", zap.Int64("driver_id", driverID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	logger.Info("Profile 查询成功", zap.Int64("driver_id", driverID))
	c.JSON(http.StatusOK, resp)
}

// Income 查询司机收入明细（含趋势图数据）
// GET /api/v1/driver/income?driver_id=&period=today|week|month
//
// period 取值说明：
//   - today: 今日收入统计 + 近7日趋势
//   - week: 本周收入统计 + 近7日趋势
//   - month: 本月收入统计 + 近7日趋势
func (h *DriverHandler) Income(c *gin.Context) {
	driverIDStr := c.Query("driver_id")
	if driverIDStr == "" {
		logger.Warn("Income 缺少 driver_id 参数")
		c.JSON(http.StatusBadRequest, gin.H{"error": "driver_id is required"})
		return
	}
	driverID, err := strconv.ParseInt(driverIDStr, 10, 64)
	if err != nil {
		logger.Warn("Income 参数错误", zap.String("driver_id", driverIDStr), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid driver_id"})
		return
	}

	period := c.Query("period")
	if period == "" {
		period = "today"
	}

	resp, err := h.client.GetIncome(c.Request.Context(), &pb.GetDriverIncomeReq{
		DriverId: driverID,
		Period:   period,
	})
	if err != nil {
		logger.Error("Income 查询失败", zap.Int64("driver_id", driverID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	logger.Info("Income 查询成功", zap.Int64("driver_id", driverID), zap.String("period", period))
	c.JSON(http.StatusOK, resp)
}

// ==================== 账号安全 ====================

// ChangeMobile 修改手机号
// PUT /api/v1/driver/mobile
// 请求体: {"driver_id":200000001,"new_mobile":"13800138000","verify_code":"123456"}
func (h *DriverHandler) ChangeMobile(c *gin.Context) {
	var req struct {
		DriverId  int64  `json:"driver_id" binding:"required"`
		NewMobile string `json:"new_mobile" binding:"required"`
		VerifyCode string `json:"verify_code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("ChangeMobile 参数错误", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.client.ChangeMobile(c.Request.Context(), &pb.ChangeMobileReq{
		DriverId:   req.DriverId,
		NewMobile:  req.NewMobile,
		VerifyCode: req.VerifyCode,
	})
	if err != nil {
		logger.Error("ChangeMobile 失败", zap.Int64("driver_id", req.DriverId), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	logger.Info("ChangeMobile 成功", zap.Int64("driver_id", req.DriverId))
	c.JSON(http.StatusOK, resp)
}

// ChangePassword 修改密码
// PUT /api/v1/driver/password
// 请求体: {"driver_id":200000001,"old_password":"xxx","new_password":"yyy"}
func (h *DriverHandler) ChangePassword(c *gin.Context) {
	var req struct {
		DriverId    int64  `json:"driver_id" binding:"required"`
		OldPassword string `json:"old_password" binding:"required"`
		NewPassword string `json:"new_password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("ChangePassword 参数错误", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.client.ChangePassword(c.Request.Context(), &pb.ChangePasswordReq{
		DriverId:    req.DriverId,
		OldPassword: req.OldPassword,
		NewPassword: req.NewPassword,
	})
	if err != nil {
		logger.Error("ChangePassword 失败", zap.Int64("driver_id", req.DriverId), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	logger.Info("ChangePassword 成功", zap.Int64("driver_id", req.DriverId))
	c.JSON(http.StatusOK, resp)
}

// ResetPassword 重置密码（忘记密码场景）
// PUT /api/v1/driver/password/reset
// 请求体: {"mobile":"13800138000","verify_code":"123456","new_password":"xxx"}
// 通过手机号反查司机ID，无需旧密码
func (h *DriverHandler) ResetPassword(c *gin.Context) {
	var req struct {
		Mobile      string `json:"mobile" binding:"required"`
		VerifyCode  string `json:"verify_code" binding:"required"`
		NewPassword string `json:"new_password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("ResetPassword 参数错误", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.client.ResetPassword(c.Request.Context(), &pb.ResetPasswordReq{
		Mobile:      req.Mobile,
		VerifyCode:  req.VerifyCode,
		NewPassword: req.NewPassword,
	})
	if err != nil {
		logger.Error("ResetPassword 失败", zap.String("mobile", req.Mobile), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	logger.Info("ResetPassword 成功", zap.String("mobile", req.Mobile))
	c.JSON(http.StatusOK, resp)
}

// ==================== 资料 & 认证（Upsert 模式） ====================
// 以下三个认证接口均使用 PUT 方法实现 Upsert 语义：
//   首次提交 → 创建记录
//   再次提交 → 更新已有记录

// UpdateProfile 更新个人资料
// PUT /api/v1/driver/profile
// 请求体: {"driver_id":200000001,"nickname":"张师傅","avatar":"http://...","gender":1}
func (h *DriverHandler) UpdateProfile(c *gin.Context) {
	var req struct {
		DriverId int64  `json:"driver_id" binding:"required"`
		Nickname string `json:"nickname"`
		Avatar   string `json:"avatar"`
		Gender   int32  `json:"gender"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("UpdateProfile 参数错误", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.client.UpdateProfile(c.Request.Context(), &pb.UpdateProfileReq{
		DriverId: req.DriverId,
		Nickname: req.Nickname,
		Avatar:   req.Avatar,
		Gender:   req.Gender,
	})
	if err != nil {
		logger.Error("UpdateProfile 失败", zap.Int64("driver_id", req.DriverId), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	logger.Info("UpdateProfile 成功", zap.Int64("driver_id", req.DriverId))
	c.JSON(http.StatusOK, resp)
}

// UpdateRealname 提交实名认证
// PUT /api/v1/driver/realname
func (h *DriverHandler) UpdateRealname(c *gin.Context) {
	var req struct {
		DriverId       int64  `json:"driver_id" binding:"required"`
		RealName       string `json:"real_name"`
		IdCardNo       string `json:"id_card_no"`
		IdCardFrontUrl string `json:"id_card_front_url"`
		IdCardBackUrl  string `json:"id_card_back_url"`
		Gender         int32  `json:"gender"`
		Birthday       string `json:"birthday"`
		Address        string `json:"address"`
		Nation         string `json:"nation"`
		ExpireDate     string `json:"expire_date"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("UpdateRealname 参数错误", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.client.UpdateRealname(c.Request.Context(), &pb.UpdateRealnameReq{
		DriverId:       req.DriverId,
		RealName:       req.RealName,
		IdCardNo:       req.IdCardNo,
		IdCardFrontUrl: req.IdCardFrontUrl,
		IdCardBackUrl:  req.IdCardBackUrl,
		Gender:         req.Gender,
		Birthday:       req.Birthday,
		Address:        req.Address,
		Nation:         req.Nation,
		ExpireDate:     req.ExpireDate,
	})
	if err != nil {
		logger.Error("UpdateRealname 失败", zap.Int64("driver_id", req.DriverId), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	logger.Info("UpdateRealname 成功", zap.Int64("driver_id", req.DriverId))
	c.JSON(http.StatusOK, resp)
}

// UpdateLicense 提交驾驶证认证
// PUT /api/v1/driver/license
func (h *DriverHandler) UpdateLicense(c *gin.Context) {
	var req struct {
		DriverId      int64  `json:"driver_id" binding:"required"`
		LicenseNo     string `json:"license_no"`
		LicenseType   string `json:"license_type"`
		LicenseUrl    string `json:"license_url"`
		FirstIssueDate string `json:"first_issue_date"`
		IssueDate     string `json:"issue_date"`
		ExpireDate    string `json:"expire_date"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("UpdateLicense 参数错误", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.client.UpdateLicense(c.Request.Context(), &pb.UpdateLicenseReq{
		DriverId:       req.DriverId,
		LicenseNo:      req.LicenseNo,
		LicenseType:    req.LicenseType,
		LicenseUrl:     req.LicenseUrl,
		FirstIssueDate: req.FirstIssueDate,
		IssueDate:      req.IssueDate,
		ExpireDate:     req.ExpireDate,
	})
	if err != nil {
		logger.Error("UpdateLicense 失败", zap.Int64("driver_id", req.DriverId), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	logger.Info("UpdateLicense 成功", zap.Int64("driver_id", req.DriverId))
	c.JSON(http.StatusOK, resp)
}

// UpdateVehicle 提交车辆信息认证
// PUT /api/v1/driver/vehicle
func (h *DriverHandler) UpdateVehicle(c *gin.Context) {
	var req struct {
		DriverId          int64  `json:"driver_id" binding:"required"`
		PlateNo           string `json:"plate_no"`
		VehicleModel      string `json:"vehicle_model"`
		VehicleBrand      string `json:"vehicle_brand"`
		VehicleColor      string `json:"vehicle_color"`
		VehicleColorCode  string `json:"vehicle_color_code"`
		SeatCount         int32  `json:"seat_count"`
		DrivingLicenseUrl string `json:"driving_license_url"`
		VehiclePhotoUrl   string `json:"vehicle_photo_url"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("UpdateVehicle 参数错误", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.client.UpdateVehicle(c.Request.Context(), &pb.UpdateVehicleReq{
		DriverId:          req.DriverId,
		PlateNo:           req.PlateNo,
		VehicleModel:      req.VehicleModel,
		VehicleBrand:      req.VehicleBrand,
		VehicleColor:      req.VehicleColor,
		VehicleColorCode:  req.VehicleColorCode,
		SeatCount:         req.SeatCount,
		DrivingLicenseUrl: req.DrivingLicenseUrl,
		VehiclePhotoUrl:   req.VehiclePhotoUrl,
	})
	if err != nil {
		logger.Error("UpdateVehicle 失败", zap.Int64("driver_id", req.DriverId), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	logger.Info("UpdateVehicle 成功", zap.Int64("driver_id", req.DriverId))
	c.JSON(http.StatusOK, resp)
}

// ==================== 银行卡管理 ====================

// BindBankCard 绑定银行卡（首次）
// PUT /api/v1/driver/bankcard
// 请求体: {"driver_id":200000001,"bank_name":"工商银行","bank_card_no":"6222021234567891215","account_name":"张三"}
func (h *DriverHandler) BindBankCard(c *gin.Context) {
	var req struct {
		DriverId   int64  `json:"driver_id" binding:"required"`
		BankName   string `json:"bank_name" binding:"required"`
		BankCode   string `json:"bank_code"`
		BankCardNo string `json:"bank_card_no" binding:"required"`
		AccountName string `json:"account_name" binding:"required"`
		CardType   int32  `json:"card_type"`
		BranchName string `json:"branch_name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("BindBankCard 参数错误", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.client.BindBankCard(c.Request.Context(), &pb.BindBankCardReq{
		DriverId:   req.DriverId,
		BankName:   req.BankName,
		BankCode:   req.BankCode,
		BankCardNo: req.BankCardNo,
		AccountName: req.AccountName,
		CardType:   req.CardType,
		BranchName: req.BranchName,
	})
	if err != nil {
		logger.Error("BindBankCard 失败", zap.Int64("driver_id", req.DriverId), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	logger.Info("BindBankCard 成功", zap.Int64("driver_id", req.DriverId))
	c.JSON(http.StatusOK, resp)
}

// GetBankCard 查询银行卡信息
// GET /api/v1/driver/bankcard?driver_id=200000001
// 返回脱敏后的卡号信息（不含原始卡号）
func (h *DriverHandler) GetBankCard(c *gin.Context) {
	driverIDStr := c.Query("driver_id")
	if driverIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "driver_id is required"})
		return
	}
	driverID, err := strconv.ParseInt(driverIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid driver_id"})
		return
	}

	resp, err := h.client.GetBankCard(c.Request.Context(), &pb.GetBankCardReq{DriverId: driverID})
	if err != nil {
		logger.Error("GetBankCard 失败", zap.Int64("driver_id", driverID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// UpdateBankCard 更换银行卡（每月限1次）
// PUT /api/v1/driver/bankcard/update
func (h *DriverHandler) UpdateBankCard(c *gin.Context) {
	var req struct {
		DriverId   int64  `json:"driver_id" binding:"required"`
		BankName   string `json:"bank_name" binding:"required"`
		BankCode   string `json:"bank_code"`
		BankCardNo string `json:"bank_card_no" binding:"required"`
		AccountName string `json:"account_name" binding:"required"`
		CardType   int32  `json:"card_type"`
		BranchName string `json:"branch_name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("UpdateBankCard 参数错误", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.client.UpdateBankCard(c.Request.Context(), &pb.UpdateBankCardReq{
		DriverId:   req.DriverId,
		BankName:   req.BankName,
		BankCode:   req.BankCode,
		BankCardNo: req.BankCardNo,
		AccountName: req.AccountName,
		CardType:   req.CardType,
		BranchName: req.BranchName,
	})
	if err != nil {
		logger.Error("UpdateBankCard 失败", zap.Int64("driver_id", req.DriverId), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	logger.Info("UpdateBankCard 成功", zap.Int64("driver_id", req.DriverId))
	c.JSON(http.StatusOK, resp)
}

// ==================== 钱包 & 提现 ====================

// GetWallet 查询钱包概览
// GET /api/v1/driver/wallet?driver_id=200000001
// 返回：余额、冻结金额、今日/周/月收入、提现次数、银行卡状态等聚合数据
func (h *DriverHandler) GetWallet(c *gin.Context) {
	driverIDStr := c.Query("driver_id")
	if driverIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "driver_id is required"})
		return
	}
	driverID, err := strconv.ParseInt(driverIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid driver_id"})
		return
	}

	resp, err := h.client.GetWallet(c.Request.Context(), &pb.GetWalletReq{DriverId: driverID})
	if err != nil {
		logger.Error("GetWallet 失败", zap.Int64("driver_id", driverID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetWithdrawPage 查询提现页信息
// GET /api/v1/driver/withdraw/page?driver_id=200000001
// 返回：提现规则、资格状态、银行卡摘要、推荐金额、须知
func (h *DriverHandler) GetWithdrawPage(c *gin.Context) {
	driverIDStr := c.Query("driver_id")
	if driverIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "driver_id is required"})
		return
	}
	driverID, err := strconv.ParseInt(driverIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid driver_id"})
		return
	}

	resp, err := h.client.GetWithdrawPage(c.Request.Context(), &pb.GetWithdrawPageReq{DriverId: driverID})
	if err != nil {
		logger.Error("GetWithdrawPage 失败", zap.Int64("driver_id", driverID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// ApplyWithdraw 申请提现
// POST /api/v1/driver/withdraw
// 请求体: {"driver_id":200000001,"amount":1200}
//
// 业务规则（由后端校验）：
//   - 单笔上限 5000 元
//   - 每日最多 3 次
//   - 需先绑定银行卡
//   - 可用余额 ≥ 申请金额
func (h *DriverHandler) ApplyWithdraw(c *gin.Context) {
	var req struct {
		DriverId int64   `json:"driver_id" binding:"required"`
		Amount   int64 `json:"amount" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("ApplyWithdraw 参数错误", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.client.ApplyWithdraw(c.Request.Context(), &pb.ApplyWithdrawReq{
		DriverId: req.DriverId,
		Amount:   req.Amount,
	})
	if err != nil {
		logger.Error("ApplyWithdraw 失败", zap.Int64("driver_id", req.DriverId), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	logger.Info("ApplyWithdraw 成功", zap.Int64("driver_id", req.DriverId))
	c.JSON(http.StatusOK, resp)
}

// GetWithdrawRecords 分页查询提现记录
// GET /api/v1/driver/withdraw/records?driver_id=200000001&page=1&page_size=20
func (h *DriverHandler) GetWithdrawRecords(c *gin.Context) {
	driverIDStr := c.Query("driver_id")
	if driverIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "driver_id is required"})
		return
	}
	driverID, err := strconv.ParseInt(driverIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid driver_id"})
		return
	}

	var page int32 = 1
	var pageSize int32 = 20
	if p := c.Query("page"); p != "" {
		if v, e := strconv.ParseInt(p, 10, 32); e == nil { page = int32(v) }
	}
	if ps := c.Query("page_size"); ps != "" {
		if v, e := strconv.ParseInt(ps, 10, 32); e == nil { pageSize = int32(v) }
	}

	resp, err := h.client.GetWithdrawRecords(c.Request.Context(), &pb.GetWithdrawRecordsReq{
		DriverId: driverID,
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		logger.Error("GetWithdrawRecords 失败", zap.Int64("driver_id", driverID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetIncomeDetail 查询收入分类明细
// GET /api/v1/driver/income/detail?driver_id=200000001&period=today
// 返回按类型分组的收入汇总（订单收入、奖励、空驶补偿等）
func (h *DriverHandler) GetIncomeDetail(c *gin.Context) {
	driverIDStr := c.Query("driver_id")
	if driverIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "driver_id is required"})
		return
	}
	driverID, err := strconv.ParseInt(driverIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid driver_id"})
		return
	}

	period := c.Query("period")
	if period == "" {
		period = "today"
	}

	resp, err := h.client.GetIncomeDetail(c.Request.Context(), &pb.GetIncomeDetailReq{
		DriverId: driverID,
		Period:   period,
	})
	if err != nil {
		logger.Error("GetIncomeDetail 失败", zap.Int64("driver_id", driverID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

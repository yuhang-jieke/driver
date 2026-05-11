// Package rpcclient 封装 BFF 层对 srvDriver 的 gRPC 远程调用，
// 提供 Go 友好的方法签名，屏蔽底层连接管理和序列化细节。
//
// 设计模式：Facade（外观模式）
// 对外暴露简洁的 Go 方法调用，内部委托给 protoc 生成的 gRPC stub。
// 这样 BFF 层代码不需要关心 gRPC 连接管理、重试、超时等底层细节。
//
// 使用方式：
//   1. 在 BFF main.go 中初始化：client, _ := rpcclient.NewDriverClient("localhost:8001")
//   2. 注入到 Handler：handler.NewDriverHandler(client)
//   3. 在 Handler 方法中调用：h.client.GetWallet(ctx, req)
//
// 连接生命周期：
//   - NewDriverClient: 建立 gRPC TCP 连接
//   - Close(): 关闭连接（通常在程序退出时 defer 调用）
//   - 当前实现为长连接（不按请求创建/关闭）
package rpcclient

import (
	"context"

	driver "driver/taketaxi/common/kitexGen"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// DriverClient gRPC 客户端封装
// 持有两个核心字段：
//   - conn: 底层 gRPC TCP 连接（管理连接的生命周期）
//   - client: protoc 自动生成的 DriverServiceClient stub（实际发送 RPC 请求）
type DriverClient struct {
	conn   *grpc.ClientConn            // gRPC 长连接
	client driver.DriverServiceClient  // gRPC 服务客户端桩
}

// NewDriverClient 创建 gRPC 客户端并建立连接
// addr 格式："host:port"，例如 "localhost:8001" 或 "127.0.0.1:8001"
//
// 安全说明：
//   当前使用 insecure.NewCredentials()（无 TLS 加密）
//   生产环境应替换为真实的 TLS 凭证
func NewDriverClient(addr string) (*DriverClient, error) {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return &DriverClient{
		conn:   conn,
		client: driver.NewDriverServiceClient(conn),
	}, nil
}

// Close 关闭 gRPC 连接
// 应在服务关闭时调用以释放资源：defer client.Close()
func (c *DriverClient) Close() { c.conn.Close() }

// ==================== 基础 CRUD ====================

// Create 创建司机
func (c *DriverClient) Create(ctx context.Context, req *driver.CreateDriverReq) (*driver.CreateDriverResp, error) {
	return c.client.Create(ctx, req)
}

// Get 查询单个司机
func (c *DriverClient) Get(ctx context.Context, req *driver.GetDriverReq) (*driver.GetDriverResp, error) {
	return c.client.Get(ctx, req)
}

// List 查询司机列表
func (c *DriverClient) List(ctx context.Context, req *driver.ListDriverReq) (*driver.ListDriverResp, error) {
	return c.client.List(ctx, req)
}

// Update 更新司机
func (c *DriverClient) Update(ctx context.Context, req *driver.UpdateDriverReq) (*driver.UpdateDriverResp, error) {
	return c.client.Update(ctx, req)
}

// Delete 删除司机
func (c *DriverClient) Delete(ctx context.Context, req *driver.DeleteDriverReq) (*driver.DeleteDriverResp, error) {
	return c.client.Delete(ctx, req)
}

// ==================== 个人信息 & 收入 ====================

// GetProfile 查询司机个人信息与接单统计
func (c *DriverClient) GetProfile(ctx context.Context, req *driver.GetDriverProfileReq) (*driver.GetDriverProfileResp, error) {
	return c.client.GetProfile(ctx, req)
}

// GetIncome 查询司机收入明细（含趋势图）
func (c *DriverClient) GetIncome(ctx context.Context, req *driver.GetDriverIncomeReq) (*driver.GetDriverIncomeResp, error) {
	return c.client.GetIncome(ctx, req)
}

// ==================== 账号安全 ====================

// ChangeMobile 修改手机号（含30天冷却）
func (c *DriverClient) ChangeMobile(ctx context.Context, req *driver.ChangeMobileReq) (*driver.ChangeMobileResp, error) {
	return c.client.ChangeMobile(ctx, req)
}

// ChangePassword 修改密码（含限流5次/h）
func (c *DriverClient) ChangePassword(ctx context.Context, req *driver.ChangePasswordReq) (*driver.ChangePasswordResp, error) {
	return c.client.ChangePassword(ctx, req)
}

// ResetPassword 重置密码（忘记密码场景）
func (c *DriverClient) ResetPassword(ctx context.Context, req *driver.ResetPasswordReq) (*driver.ResetPasswordResp, error) {
	return c.client.ResetPassword(ctx, req)
}

// ==================== 资料 & 认证 ====================

// UpdateProfile 更新个人资料
func (c *DriverClient) UpdateProfile(ctx context.Context, req *driver.UpdateProfileReq) (*driver.UpdateProfileResp, error) {
	return c.client.UpdateProfile(ctx, req)
}

// UpdateRealname 提交实名认证
func (c *DriverClient) UpdateRealname(ctx context.Context, req *driver.UpdateRealnameReq) (*driver.UpdateRealnameResp, error) {
	return c.client.UpdateRealname(ctx, req)
}

// UpdateLicense 提交驾驶证认证
func (c *DriverClient) UpdateLicense(ctx context.Context, req *driver.UpdateLicenseReq) (*driver.UpdateLicenseResp, error) {
	return c.client.UpdateLicense(ctx, req)
}

// UpdateVehicle 提交车辆信息
func (c *DriverClient) UpdateVehicle(ctx context.Context, req *driver.UpdateVehicleReq) (*driver.UpdateVehicleResp, error) {
	return c.client.UpdateVehicle(ctx, req)
}

// ==================== 银行卡 ====================

// BindBankCard 绑定银行卡（首次）
func (c *DriverClient) BindBankCard(ctx context.Context, req *driver.BindBankCardReq) (*driver.BindBankCardResp, error) {
	return c.client.BindBankCard(ctx, req)
}

// GetBankCard 查询银行卡信息
func (c *DriverClient) GetBankCard(ctx context.Context, req *driver.GetBankCardReq) (*driver.GetBankCardResp, error) {
	return c.client.GetBankCard(ctx, req)
}

// UpdateBankCard 更换银行卡（月频次限制）
func (c *DriverClient) UpdateBankCard(ctx context.Context, req *driver.UpdateBankCardReq) (*driver.UpdateBankCardResp, error) {
	return c.client.UpdateBankCard(ctx, req)
}

// ==================== 钱包 & 提现 ====================

// GetWallet 查询钱包概览（多维度聚合数据）
func (c *DriverClient) GetWallet(ctx context.Context, req *driver.GetWalletReq) (*driver.GetWalletResp, error) {
	return c.client.GetWallet(ctx, req)
}

// GetWithdrawPage 查询提现页信息（规则 + 资格 + 银行卡摘要）
func (c *DriverClient) GetWithdrawPage(ctx context.Context, req *driver.GetWithdrawPageReq) (*driver.GetWithdrawPageResp, error) {
	return c.client.GetWithdrawPage(ctx, req)
}

// ApplyWithdraw 申请提现（8步完整流程）
func (c *DriverClient) ApplyWithdraw(ctx context.Context, req *driver.ApplyWithdrawReq) (*driver.ApplyWithdrawResp, error) {
	return c.client.ApplyWithdraw(ctx, req)
}

// GetWithdrawRecords 分页查询提现记录
func (c *DriverClient) GetWithdrawRecords(ctx context.Context, req *driver.GetWithdrawRecordsReq) (*driver.GetWithdrawRecordsResp, error) {
	return c.client.GetWithdrawRecords(ctx, req)
}

// GetIncomeDetail 查询收入分类明细
func (c *DriverClient) GetIncomeDetail(ctx context.Context, req *driver.GetIncomeDetailReq) (*driver.GetIncomeDetailResp, error) {
	return c.client.GetIncomeDetail(ctx, req)
}

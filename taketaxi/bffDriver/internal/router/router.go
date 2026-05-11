// Package router 定义 BFF 层全部 HTTP 路由，
// 将 URL 路径映射到对应的 Handler 方法，并配置 CORS 中间件。
//
// 路由设计规范：
//   - RESTful 风格：GET 查询 / POST 创建 / PUT 更新 / DELETE 删除
//   - 版本化 URL 前缀：/api/v1/
//   - 管理端接口（复数名词）：/api/v1/drivers
//   - 司机端接口（单数名词）：/api/v1/driver/...
//
// 完整路由表：
//
//	┌──────────────────────┬─────────────────────────┬──────┐
//	│ 功能模块              │ HTTP 方法 + 路径          │ 处理器│
//	├──────────────────────┼─────────────────────────┼──────┤
//	│ 管理端 CRUD           │ GET    /drivers          │ List │
//	│                      │ GET    /drivers/:id       │ Get  │
//	│                      │ POST   /drivers          │ Create│
//	│                      │ PUT    /drivers/:id       │ Update│
//	│                      │ DELETE /drivers/:id       │ Delete│
//	├──────────────────────┼─────────────────────────┼──────┤
//	│ 个人信息              │ GET    /driver/profile     │ Profile│
//	│ 收入                  │ GET    /driver/income      │ Income│
//	├──────────────────────┼─────────────────────────┼──────┤
//	│ 修改手机号            │ PUT    /driver/mobile      │ ChangeMobile│
//	│ 修改密码              │ PUT    /driver/password    │ ChangePassword│
//	│ 重置密码              │ PUT    /driver/password/reset│ ResetPassword│
//	├──────────────────────┼─────────────────────────┼──────┤
//	│ 更新资料              │ PUT    /driver/profile     │ UpdateProfile│
//	│ 实名认证              │ PUT    /driver/realname    │ UpdateRealname│
//	│ 驾驶证认证            │ PUT    /driver/license     │ UpdateLicense│
//	│ 车辆信息              │ PUT    /driver/vehicle     │ UpdateVehicle│
//	├──────────────────────┼─────────────────────────┼──────┤
//	│ 查询银行卡            │ GET    /driver/bankcard    │ GetBankCard│
//	│ 绑定银行卡            │ PUT    /driver/bankcard    │ BindBankCard│
//	│ 更换银行卡            │ PUT    /driver/bankcard/update│ UpdateBankCard│
//	├──────────────────────┼─────────────────────────┼──────┤
//	│ 钱包概览              │ GET    /driver/wallet      │ GetWallet│
//	│ 提现页信息            │ GET    /driver/withdraw/page│ GetWithdrawPage│
//	│ 申请提现              │ POST   /driver/withdraw   │ ApplyWithdraw│
//	│ 提现记录              │ GET    /driver/withdraw/records│ GetWithdrawRecords│
//	│ 收入明细              │ GET    /driver/income/detail│ GetIncomeDetail│
//	├──────────────────────┼─────────────────────────┼──────┤
//	│ 文件上传              │ POST   /upload             │ Upload│
//	│ 文件删除              │ DELETE /upload             │ Delete│
//	└──────────────────────┴─────────────────────────┴──────┘
package router

import (
	"driver/taketaxi/bffDriver/internal/handler"
	"driver/taketaxi/bffDriver/internal/rpcClient"
	"driver/taketaxi/pkg/upload"
	"net/http"

	"github.com/gin-gonic/gin"
)

// NewRouter 创建并配置 Gin 引擎，注册所有路由和中间件
// 参数：
//   - client: gRPC 客户端实例，用于转发请求到 srvDriver
//   - storage: 文件存储服务实例，用于文件上传/删除
func NewRouter(client *rpcclient.DriverClient, storage upload.Storage) *gin.Engine {
	r := gin.Default()

	// ========== CORS 中间件 ==========
	// 允许前端跨域访问（开发环境使用 *，生产环境应限制为具体域名）
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	})

	// ========== 创建 Handler 实例 ==========
	driverHandler := handler.NewDriverHandler(client)
	uploadHandler := handler.NewUploadHandler(storage)

	// ========== 管理端 CRUD（复数路径） ==========
	r.GET("/api/v1/drivers", driverHandler.List)           // 列表
	r.GET("/api/v1/drivers/:id", driverHandler.Get)        // 详情
	r.POST("/api/v1/drivers", driverHandler.Create)        // 创建
	r.PUT("/api/v1/drivers/:id", driverHandler.Update)     // 更新
	r.DELETE("/api/v1/drivers/:id", driverHandler.Delete)  // 删除

	// ========== 司机端：个人信息 & 收入 ==========
	r.GET("/api/v1/driver/profile", driverHandler.Profile)  // 个人信息与接单统计
	r.GET("/api/v1/driver/income", driverHandler.Income)    // 收入明细(含趋势图)

	// ========== 司机端：账号安全 ==========
	r.PUT("/api/v1/driver/mobile", driverHandler.ChangeMobile)         // 修改手机号
	r.PUT("/api/v1/driver/password", driverHandler.ChangePassword)     // 修改密码
	r.PUT("/api/v1/driver/password/reset", driverHandler.ResetPassword) // 重置密码

	// ========== 司机端：资料 & 认证（Upsert） ==========
	r.PUT("/api/v1/driver/profile", driverHandler.UpdateProfile)  // 更新昵称/头像/性别
	r.PUT("/api/v1/driver/realname", driverHandler.UpdateRealname) // 实名认证
	r.PUT("/api/v1/driver/license", driverHandler.UpdateLicense)   // 驾驶证认证
	r.PUT("/api/v1/driver/vehicle", driverHandler.UpdateVehicle)   // 车辆信息

	// ========== 司机端：银行卡 ==========
	r.GET("/api/v1/driver/bankcard", driverHandler.GetBankCard)             // 查询银行卡
	r.PUT("/api/v1/driver/bankcard", driverHandler.BindBankCard)            // 绑定银行卡
	r.PUT("/api/v1/driver/bankcard/update", driverHandler.UpdateBankCard)   // 更换银行卡

	// ========== 司机端：钱包 & 提现 ==========
	r.GET("/api/v1/driver/wallet", driverHandler.GetWallet)                 // 钱包概览
	r.GET("/api/v1/driver/withdraw/page", driverHandler.GetWithdrawPage)    // 提现页信息
	r.POST("/api/v1/driver/withdraw", driverHandler.ApplyWithdraw)          // 申请提现
	r.GET("/api/v1/driver/withdraw/records", driverHandler.GetWithdrawRecords) // 提现记录
	r.GET("/api/v1/driver/income/detail", driverHandler.GetIncomeDetail)    // 收入明细

	// ========== 文件上传 ==========
	r.POST("/api/v1/upload", uploadHandler.Upload)   // 上传图片/文件
	r.DELETE("/api/v1/upload", uploadHandler.Delete)  // 删除已上传文件

	return r
}

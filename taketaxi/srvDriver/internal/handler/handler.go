package handler

import (
	driver "driver/taketaxi/common/kitexGen"
	"driver/taketaxi/srvDriver/internal/repository"
	"driver/taketaxi/srvDriver/internal/service"

	"github.com/redis/go-redis/v9"
)

// DriverHandler gRPC 服务端 Handler 结构体
// 嵌入 UnimplementedDriverServiceServer 以实现向前兼容的接口实现
//
// 方法按业务域拆分到不同文件：
//   - crudHandler.go     : Create, Get, List, Update, Delete
//   - profileHandler.go  : GetProfile, UpdateProfile
//   - authHandler.go     : ChangeMobile, ChangePassword, ResetPassword
//   - verifyHandler.go   : UpdateRealname, UpdateLicense, UpdateVehicle
//   - walletHandler.go   : GetWallet, ApplyWithdraw, GetWithdrawRecords, GetIncomeDetail, GetIncome
//   - bankCardHandler.go : BindBankCard, GetBankCard, UpdateBankCard
type DriverHandler struct {
	driver.UnimplementedDriverServiceServer
	svc *service.DriverService
}

// NewDriverHandler 创建 Handler 实例
// 依赖注入：repo + rdb → service → handler
func NewDriverHandler(repo *repository.DriverRepo, rdb *redis.Client) *DriverHandler {
	return &DriverHandler{svc: service.NewDriverService(repo, rdb)}
}

package service

import (
	"driver/taketaxi/srvDriver/internal/repository"

	"github.com/redis/go-redis/v9"
)

// DriverService 司机业务服务，持有 repo（数据库）和 rdb（Redis）
//
// 方法按业务域拆分到不同文件：
//   - crudService.go     : Create, Get, List, Update, Delete
//   - profileService.go  : GetProfile, UpdateProfile
//   - authService.go     : ChangeMobile, ChangePassword, ResetPassword
//   - verifyService.go   : UpdateRealname, UpdateLicense, UpdateVehicle
//   - walletService.go   : GetWallet, ApplyWithdraw, GetWithdrawRecords, GetIncomeDetail, GetIncome
//   - bankCardService.go : BindBankCard, GetBankCard, UpdateBankCard
type DriverService struct {
	repo *repository.DriverRepo
	rdb  *redis.Client
}

// NewDriverService 创建 DriverService 实例
func NewDriverService(repo *repository.DriverRepo, rdb *redis.Client) *DriverService {
	return &DriverService{repo: repo, rdb: rdb}
}

// maskBankCard 银行卡脱敏：保留后4位
func maskBankCard(cardNo string) string {
	if len(cardNo) <= 4 {
		return cardNo
	}
	return "****" + cardNo[len(cardNo)-4:]
}

// getBankCode 根据银行名称返回银行代码
func getBankCode(bankName string) string {
	bankMap := map[string]string{
		"工商银行": "ICBC", "建设银行": "CCB", "农业银行": "ABC", "中国银行": "BOC",
		"交通银行": "BOCOM", "招商银行": "CMB", "邮储银行": "PSBC", "兴业银行": "CIB",
		"浦发银行": "SPDB", "民生银行": "CMBC",
	}
	if code, ok := bankMap[bankName]; ok {
		return code
	}
	return "OTHER"
}



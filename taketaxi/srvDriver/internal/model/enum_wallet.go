package model

// ==================== 钱包相关枚举 ====================

// WalletTransactionType 钱包交易类型
const (
	WalletTxTypeOrderIncome int8 = 1 // 订单收入
	WalletTxTypeBonusIncome int8 = 2 // 奖励收入
	WalletTxTypeWithdraw    int8 = 3 // 提现支出
	WalletTxTypePenalty     int8 = 4 // 罚款支出
	WalletTxTypeRefund      int8 = 5 // 退款
)

// WalletTransactionStatus 钱包交易状态
const (
	WalletTxStatusSuccess  int8 = 1 // 成功
	WalletTxStatusFailed   int8 = 2 // 失败
	WalletTxStatusPending  int8 = 3 // 处理中
)

// WithdrawStatus 提现状态
const (
	WithdrawStatusPending  int8 = 1 // 处理中
	WithdrawStatusSuccess  int8 = 2 // 成功
	WithdrawStatusFailed   int8 = 3 // 失败
)

// WithdrawStatusDesc 提现状态描述映射
var WithdrawStatusDesc = map[int8]string{
	WithdrawStatusPending: "处理中",
	WithdrawStatusSuccess: "成功",
	WithdrawStatusFailed:  "失败",
}

// ==================== 钱包业务约束常量 ====================

const (
	MaxWithdrawPerDay      = 3        // 每日最大提现次数
	MaxWithdrawPerOrder    = 500000   // 单笔提现上限（分）= 5000元
	WithdrawChannelDefault = "default" // 默认提现渠道
	WithdrawMinAmount      = 100      // 最低提现金额（分）= 1元
)

// WithdrawPageDisableReason 提现页禁用原因码
const (
	WithdrawDisableNoCard           = "NO_BANK_CARD"            // 未绑定银行卡
	WithdrawDisableCreditCard       = "CREDIT_CARD_NOT_ALLOWED" // 信用卡不允许提现
	WithdrawDisableVerify           = "VERIFY_PENDING"           // 认证审核中
	WithdrawDisableCardNameMismatch = "CARD_NAME_MISMATCH"      // 银行卡姓名与实名不一致
	WithdrawDisableCount            = "WITHDRAW_COUNT_LIMIT"    // 今日提现次数超限
	WithdrawDisableZero             = "AVAILABLE_AMOUNT_ZERO"   // 可提现余额为零
	WithdrawDisableAccountFrozen    = "ACCOUNT_FROZEN"          // 账号已冻结
	WithdrawDisableAccountClosed    = "ACCOUNT_CLOSED"          // 账号已注销
)

// WithdrawArrivalText 预计到账文案
const WithdrawArrivalText = "预计2小时到账"

// WithdrawNotice 提现须知
var WithdrawNotice = []string{
	"提现将转入您的银行卡，预计2小时到账",
	"每日最多提现3次，单笔上限5000元，最低1元起提",
	"仅支持提现至本人借记卡",
	"提现前请确认银行卡信息正确",
}

// IncomeDetailType 收入明细分类
const (
	IncomeTypeBaseFare  int8 = 1 // 基础车费
	IncomeTypeBonus     int8 = 2 // 奖励
	IncomeTypeEmptyComp int8 = 3 // 空驶补偿
	IncomeTypeToll      int8 = 4 // 高速费
	IncomeTypeOther     int8 = 5 // 其他
)

// IncomeDetailTypeDesc 收入明细类型描述映射
var IncomeDetailTypeDesc = map[int8]string{
	IncomeTypeBaseFare:  "基础车费",
	IncomeTypeBonus:     "奖励",
	IncomeTypeEmptyComp: "空驶补偿",
	IncomeTypeToll:      "高速费",
	IncomeTypeOther:     "其他",
}

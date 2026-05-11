package model

// ==================== 银行卡相关枚举 ====================

// BankCardType 银行卡类型
const (
	BankCardTypeDebit  int8 = 1 // 借记卡
	BankCardTypeCredit int8 = 2 // 信用卡
)

// BankCardStatus 银行卡状态
const (
	BankCardStatusActive  int8 = 1 // 有效
	BankCardStatusUnbound int8 = 2 // 已解绑
)

// BankCardNoMaxLength 银行卡号最大长度
const BankCardNoMaxLength = 20

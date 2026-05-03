package model

import "time"

// ==================== 钱包 & 提现 & 银行卡 (6张表) ====================
// 金额字段已统一使用分(int64)存储，避免浮点精度问题。

// DriverWallet 钱包表 (driver_wallet)
//
// 司机的资金账户，每个司机只有一条记录：
//   - Balance: 总余额（包含冻结部分）
//   - FrozenAmount: 冻结金额（T-3 结算中不可提现的部分，是 Balance 的子集）
//   - 可提现金额 = Balance - FrozenAmount
//   - Version: 乐观锁版本号，防止并发扣款冲突
type DriverWallet struct {
	Id            int64     `gorm:"column:id;primaryKey;autoIncrement;comment:主键ID" json:"id"`
	DriverId      int64     `gorm:"column:driver_id;comment:司机ID" json:"driver_id"`
	Balance       int64     `gorm:"column:balance;comment:总余额(分,含冻结)" json:"balance"`
	FrozenAmount  int64     `gorm:"column:frozen_amount;comment:冻结金额(分)" json:"frozen_amount"`
	TotalIncome   int64     `gorm:"column:total_income;comment:累计总收入(分)" json:"total_income"`
	TotalWithdraw int64     `gorm:"column:total_withdraw;comment:累计提现金额(分)" json:"total_withdraw"`
	Version       int       `gorm:"column:version;comment:乐观锁版本号" json:"version"`
	UpdatedAt     time.Time `gorm:"column:updated_at;comment:更新时间" json:"updated_at"`
}

func (DriverWallet) TableName() string { return "driver_wallet" }

// DriverIncomeLog 收入流水表 (driver_income_log)
// 记录每笔收入的来源和金额：订单收入、奖励、罚款、提现、退款
type DriverIncomeLog struct {
	Id            int64     `gorm:"column:id;primaryKey;autoIncrement;comment:主键ID" json:"id"`
	DriverId      int64     `gorm:"column:driver_id;comment:司机ID" json:"driver_id"`
	OrderId       int64     `gorm:"column:order_id;comment:关联订单ID" json:"order_id"`
	Amount        int64     `gorm:"column:amount;comment:金额(分,正为收入,负为支出)" json:"amount"`
	Type          int8      `gorm:"column:type;comment:类型: 1-订单收入 2-奖励 3-罚款 4-提现 5-退款" json:"type"`
	BalanceBefore int64     `gorm:"column:balance_before;comment:变更前余额(分)" json:"balance_before"`
	BalanceAfter  int64     `gorm:"column:balance_after;comment:变更后余额(分)" json:"balance_after"`
	Remark        string    `gorm:"column:remark;comment:备注说明" json:"remark"`
	CreatedAt     time.Time `gorm:"column:created_at;comment:创建时间" json:"created_at"`
}

func (DriverIncomeLog) TableName() string { return "driver_income_log" }

// WalletTransactionLog 钱包流水明细表 (wallet_transaction_log)
// 比 IncomeLog 更详细的流水记录，包含冻结金额变化和关联信息
type WalletTransactionLog struct {
	Id              int64     `gorm:"column:id;primaryKey;autoIncrement;comment:主键ID" json:"id"`
	DriverId        int64     `gorm:"column:driver_id;comment:司机ID" json:"driver_id"`
	TransactionNo   string    `gorm:"column:transaction_no;comment:流水号" json:"transaction_no"`
	TransactionType int8      `gorm:"column:transaction_type;comment:交易类型: 1-订单收入 2-奖励收入 3-提现支出 4-罚款支出 5-退款" json:"transaction_type"`
	Amount          int64     `gorm:"column:amount;comment:交易金额(分)" json:"amount"`
	BalanceBefore   int64     `gorm:"column:balance_before;comment:交易前余额(分)" json:"balance_before"`
	BalanceAfter    int64     `gorm:"column:balance_after;comment:交易后余额(分)" json:"balance_after"`
	FrozenBefore    int64     `gorm:"column:frozen_before;comment:交易前冻结金额(分)" json:"frozen_before"`
	FrozenAfter     int64     `gorm:"column:frozen_after;comment:交易后冻结金额(分)" json:"frozen_after"`
	RelatedId       int64     `gorm:"column:related_id;comment:关联ID(订单ID/提现ID等)" json:"related_id"`
	RelatedType     string    `gorm:"column:related_type;comment:关联类型" json:"related_type"`
	Status          int8      `gorm:"column:status;comment:状态: 1-成功 2-失败 3-处理中" json:"status"`
	Remark          string    `gorm:"column:remark;comment:备注" json:"remark"`
	CreatedAt       time.Time `gorm:"column:created_at;comment:创建时间" json:"created_at"`
}

func (WalletTransactionLog) TableName() string { return "wallet_transaction_log" }

// DriverWithdrawRecord 提现记录表 (driver_withdraw_record)
//
// 每次提现申请生成一条记录：
//   - Status 流转：1(处理中) → 2(成功)/3(失败)
//   - FinishTime: 提现完成（成功打款或失败退回）的时间
//   - Channel: 提现渠道预留字段（支付宝直连/微信代付/银联等）
//   - BankCardNo: 脱敏卡号用于展示，BankCardNoEncrypt 用于实际打款
//
// 注意：FinishTime 使用 *time.Time（指针），允许 NULL 表示尚未完成
type DriverWithdrawRecord struct {
	Id                int64      `gorm:"column:id;primaryKey;autoIncrement;comment:主键ID" json:"id"`
	WithdrawNo        string     `gorm:"column:withdraw_no;comment:提现单号" json:"withdraw_no"`
	DriverId          int64      `gorm:"column:driver_id;comment:司机ID" json:"driver_id"`
	Amount            int64      `gorm:"column:amount;comment:提现金额(分)" json:"amount"`
	Fee               int64      `gorm:"column:fee;comment:手续费(分)" json:"fee"`
	ActualAmount      int64      `gorm:"column:actual_amount;comment:实际到账金额(分)" json:"actual_amount"`
	BankName          string     `gorm:"column:bank_name;comment:银行名称" json:"bank_name"`
	BankCode          string     `gorm:"column:bank_code;comment:银行代码" json:"bank_code"`
	BankCardNo        string     `gorm:"column:bank_card_no;comment:银行卡号(脱敏)" json:"bank_card_no"`
	BankCardNoEncrypt string     `gorm:"column:bank_card_no_encrypt;comment:银行卡号(加密)" json:"bank_card_no_encrypt"`
	AccountName       string     `gorm:"column:account_name;comment:持卡人姓名" json:"account_name"`
	Status            int8       `gorm:"column:status;comment:状态: 1-处理中 2-成功 3-失败" json:"status"`
	FailReason        string     `gorm:"column:fail_reason;comment:失败原因" json:"fail_reason"`
	ApplyTime         time.Time  `gorm:"column:apply_time;comment:申请时间" json:"apply_time"`
	FinishTime        *time.Time `gorm:"column:finish_time;comment:完成时间" json:"finish_time"`
	Channel           string     `gorm:"column:channel;comment:提现渠道" json:"channel"`
	ChannelSerial     string     `gorm:"column:channel_serial;comment:渠道流水号" json:"channel_serial"`
	CreatedAt         time.Time  `gorm:"column:created_at;comment:创建时间" json:"created_at"`
	UpdatedAt         time.Time  `gorm:"column:updated_at;comment:更新时间" json:"updated_at"`
}

func (DriverWithdrawRecord) TableName() string { return "withdraw_record" }

// WithdrawRecord 旧版提现记录表 (withdraw_record)
// 保留兼容旧版本接口，新功能请使用 DriverWithdrawRecord
type WithdrawRecord struct {
	Id         int64     `gorm:"column:id;primaryKey;autoIncrement;comment:主键ID" json:"id"`
	WithdrawNo string    `gorm:"column:withdraw_no;comment:提现单号" json:"withdraw_no"`
	DriverId   int64     `gorm:"column:driver_id;comment:司机ID" json:"driver_id"`
	Amount     int64     `gorm:"column:amount;comment:提现金额(分)" json:"amount"`
	Fee        int64     `gorm:"column:fee;comment:手续费(分)" json:"fee"`
	BankName   string    `gorm:"column:bank_name;comment:银行名称" json:"bank_name"`
	BankCardNo string    `gorm:"column:bank_card_no;comment:银行卡号(脱敏)" json:"bank_card_no"`
	Status     int8      `gorm:"column:status;comment:状态: 1-处理中 2-成功 3-失败" json:"status"`
	FailReason string    `gorm:"column:fail_reason;comment:失败原因" json:"fail_reason"`
	ApplyTime  time.Time `gorm:"column:apply_time;comment:申请时间" json:"apply_time"`
	FinishTime time.Time `gorm:"column:finish_time;comment:完成时间" json:"finish_time"`
	CreatedAt  time.Time `gorm:"column:created_at;comment:创建时间" json:"created_at"`
}

func (WithdrawRecord) TableName() string { return "withdraw_record" }

// DriverBankCard 银行卡绑定表 (driver_bank_card)
//
// 每个司机只能绑定一张有效银行卡（uniqueIndex on driver_id）：
//   - BankCardNo: 脱敏卡号（****1215），用于前端展示
//   - BankCardNoEncrypt: 完整卡号（生产环境应 AES 加密），用于实际打款
//   - BankCode: 银行标准代码（ICBC/CCB等），对接支付渠道时使用
//   - LastModifiedAt: 最后修改时间，用于月频次限制校验（每月最多换1次）
//   - CardType: 1=借记卡 2=信用卡
type DriverBankCard struct {
	Id                int64     `gorm:"column:id;primaryKey;autoIncrement;comment:主键ID" json:"id"`
	DriverId          int64     `gorm:"column:driver_id;uniqueIndex;comment:司机ID" json:"driver_id"`
	BankName          string    `gorm:"column:bank_name;size:64;not null;comment:银行名称(如工商银行)" json:"bank_name"`
	BankCode          string    `gorm:"column:bank_code;size:32;comment:银行代码" json:"bank_code"`
	BankCardNo        string    `gorm:"column:bank_card_no;size:32;not null;comment:银行卡号(脱敏,如****8821)" json:"bank_card_no"`
	BankCardNoEncrypt string    `gorm:"column:bank_card_no_encrypt;size:255;not null;comment:银行卡号(AES加密)" json:"bank_card_no_encrypt"`
	AccountName       string    `gorm:"column:account_name;size:64;not null;comment:持卡人姓名" json:"account_name"`
	CardType          int8      `gorm:"column:card_type;default:1;comment:卡类型:1-借记卡 2-信用卡" json:"card_type"`
	BranchName        string    `gorm:"column:branch_name;size:128;comment:开户支行名称" json:"branch_name"`
	Status            int8      `gorm:"column:status;default:1;comment:状态:1-有效 2-已解绑" json:"status"`
	LastModifiedAt    time.Time `gorm:"column:last_modified_at;comment:最后修改时间(用于月频次校验)" json:"last_modified_at"`
	CreatedAt         time.Time `gorm:"column:created_at;comment:创建时间" json:"created_at"`
	UpdatedAt         time.Time `gorm:"column:updated_at;comment:更新时间" json:"updated_at"`
}

func (DriverBankCard) TableName() string { return "driver_bank_card" }

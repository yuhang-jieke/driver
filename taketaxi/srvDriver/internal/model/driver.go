// Package model 定义司机端全部数据库实体（GORM Entity），
// 每个结构体与 MySQL 表一一对应，通过 gorm tag 标注列名、类型、约束和注释。
//
// 文件组织（按业务域拆分）：
//
//	┌──────────────────┬───────────────────────────────────────────────────┐
//	│ driver.go        │ 基础实体: Driver, DriverS, Passenger             │
//	│ verify.go        │ 认证资料: Realname, License, Vehicle, Face       │
//	│ log.go           │ 审计日志: FaceAuthLog, StatusLog, OnlineLog, …  │
//	│ order.go         │ 订单模块: Order, Evaluation, Trip, Trajectory   │
//	│ wallet.go        │ 钱包提现: Wallet, IncomeLog, Withdraw, BankCard │
//	│ config.go        │ 配置统计: Statistics, Level, Pricing, Location  │
//	│ result.go        │ Repository 层查询结果 DTO                       │
//	└──────────────────┴───────────────────────────────────────────────────┘
package model

// ==================== 基础实体 ====================

// Driver GORM 通用演示模型（未绑定实际表，仅用于基础 CRUD 示例）
type Driver struct {
	ID     uint   `gorm:"primaryKey"` // 主键ID
	Name   string `gorm:"size:255"`  // 名称
	Status int    `gorm:"default:0"` // 状态
}

func (Driver) TableName() string { return "" }

// DriverS 司机基础信息表 (drivers)
//
// 这是司机的核心表，存储：
//   - 账号信息：手机号（+加密）、密码(bcrypt哈希)
//   - 个人信息：昵称、头像、性别
//   - 状态信息：账号状态、认证状态、服务分
//   - 统计数据：累计订单数、总收入
//
// 安全相关字段：
//   - Password: 存储 bcrypt 哈希值，非明文。bcrypt.DefaultCost=10
//   - MobileUpdatedAt/PasswordUpdatedAt: 用于安全审计，追踪敏感信息变更时间
type DriverS struct {
	DriverId          int64   `gorm:"column:driver_id;primaryKey;comment:司机ID，分布式ID" json:"driver_id"`
	Mobile            string  `gorm:"column:mobile;size:20;not null;comment:登录手机号" json:"mobile"`
	MobileEncrypt     string  `gorm:"column:mobile_encrypt;size:100;comment:手机号AES加密" json:"mobile_encrypt"`
	Password          string  `gorm:"column:password;size:255;comment:密码(bcrypt加密)" json:"password"`
	MobileUpdatedAt   string  `gorm:"column:mobile_updated_at;comment:手机号最后更新时间" json:"mobile_updated_at"`
	PasswordUpdatedAt string  `gorm:"column:password_updated_at;comment:密码最后更新时间" json:"password_updated_at"`
	Nickname          string  `gorm:"column:nickname;size:64;comment:司机昵称" json:"nickname"`
	Avatar            string  `gorm:"column:avatar;size:512;comment:头像URL" json:"avatar"`
	Gender            int8    `gorm:"column:gender;default:0;comment:性别: 0-未知 1-男 2-女" json:"gender"`
	Status            int8    `gorm:"column:status;not null;default:1;comment:账号状态: 1-正常 2-冻结 3-已注销" json:"status"`
	VerifyStatus      int8    `gorm:"column:verify_status;not null;default:0;comment:认证状态: 0-未认证 1-认证中 2-已认证 3-认证失败" json:"verify_status"`
	ServiceScore      float64 `gorm:"column:service_score;type:decimal(3,1);not null;default:80.0;comment:服务评分" json:"service_score"`
	OrderCount        int     `gorm:"column:order_count;not null;default:0;comment:累计完成订单数" json:"order_count"`
	TotalIncome       float64 `gorm:"column:total_income;type:decimal(12,2);not null;default:0.00;comment:累计收入金额" json:"total_income"`
	RegisterSource    string  `gorm:"column:register_source;size:32;comment:注册来源" json:"register_source"`
	CityId            int64   `gorm:"column:city_id;comment:服务城市ID" json:"city_id"`
	LastOnlineAt      string  `gorm:"column:last_online_at;comment:最后在线时间" json:"last_online_at"`
	CreatedAt         string  `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP;comment:创建时间" json:"created_at"`
	UpdatedAt         string  `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP;comment:更新时间" json:"updated_at"`
	DeletedAt         string  `gorm:"column:deleted_at;comment:删除时间" json:"deleted_at"`
}

func (DriverS) TableName() string { return "drivers" }

// Passenger 乘客基础信息表 (passenger)
//
// 字段设计要点：
//   - MobileEncrypt: 手机号 AES 加密存储，Mobile 为脱敏或原文
//   - Level: 会员等级系统，影响优惠力度和权益
//   - TotalConsumed: 使用 decimal(12,2) 保证金额精度
//   - RegisterSource: 追踪用户来源渠道（App/H5/小程序等）
type Passenger struct {
	PassengerId    int64   `gorm:"column:passenger_id;primaryKey;comment:乘客ID，分布式ID" json:"passenger_id"`
	Mobile         string  `gorm:"column:mobile;size:20;not null;comment:登录手机号" json:"mobile"`
	MobileEncrypt  string  `gorm:"column:mobile_encrypt;size:100;comment:手机号AES加密" json:"mobile_encrypt"`
	Nickname       string  `gorm:"column:nickname;size:64;comment:乘客昵称" json:"nickname"`
	Avatar         string  `gorm:"column:avatar;size:512;comment:头像URL" json:"avatar"`
	Gender         int8    `gorm:"column:gender;default:0;comment:性别: 0-未知 1-男 2-女" json:"gender"`
	Status         int8    `gorm:"column:status;not null;default:1;comment:账号状态: 1-正常 2-冻结 3-已注销" json:"status"`
	VerifyStatus   int8    `gorm:"column:verify_status;not null;default:0;comment:实名认证状态: 0-未认证 1-认证中 2-已认证" json:"verify_status"`
	Level          int8    `gorm:"column:level;not null;default:1;comment:会员等级: 1-普通 2-白银 3-黄金 4-铂金 5-钻石" json:"level"`
	TotalConsumed  float64 `gorm:"column:total_consumed;type:decimal(12,2);not null;default:0.00;comment:累计消费金额" json:"total_consumed"`
	OrderCount     int     `gorm:"column:order_count;not null;default:0;comment:累计订单数" json:"order_count"`
	RegisterSource string  `gorm:"column:register_source;size:32;comment:注册来源" json:"register_source"`
	CityId         int64   `gorm:"column:city_id;comment:常用城市ID" json:"city_id"`
	LastOrderAt    string  `gorm:"column:last_order_at;comment:最后下单时间" json:"last_order_at"`
	CreatedAt      string  `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP;comment:创建时间" json:"created_at"`
	UpdatedAt      string  `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP;comment:更新时间" json:"updated_at"`
	DeletedAt      string  `gorm:"column:deleted_at;comment:删除时间" json:"deleted_at"`
}

func (Passenger) TableName() string { return "passenger" }

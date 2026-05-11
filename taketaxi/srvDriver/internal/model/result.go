package model

// ==================== Repository 查询结果 DTO ====================
// 以下结构体用于 repository 层的 SQL 查询结果映射，
// 不对应任何数据库表，仅作为多表 JOIN / 聚合查询的承载对象。
// 从 repository 包迁入 model，实现分层解耦。

// DriverProfileResult 个人信息查询结果（多表 JOIN + 计算）
// 将 drivers 表字段 + vehicle 表字段组装为一个扁平化结构
type DriverProfileResult struct {
	Nickname     string  // 司机昵称
	Avatar       string  // 头像 URL
	ServiceScore float64 // 服务评分（80.0 起步）
	OrderCount   int     // 累计完单数
	VerifyStatus int8    // 认证状态
	Phone        string  // 手机号
	Plate        string  // 车牌号
	Car          string  // 车辆描述（型号·颜色）
	Online       bool    // 是否在线（5分钟内有活动）
	Status       string  // 状态字符串 idle/busy/offline
}

// OrderStatsResult 接单统计结果（聚合查询）
type OrderStatsResult struct {
	OrderCount     int     // 订单总数
	TotalIncome    float64 // 总收入
	OnlineDuration int     // 在线时长（秒）
}

// DailyIncomeResult 每日收入结果，用于趋势图渲染
type DailyIncomeResult struct {
	Date   string  // 日期 "2026-04-28"
	Income float64 // 当日收入
}

// VerifyInfoResult 认证信息聚合结果
// 一次性返回三类认证的状态和数据，供前端回显编辑页
type VerifyInfoResult struct {
	Realname RealnameInfoResult
	License  LicenseInfoResult
	Vehicle  VehicleInfoResult
}

// RealnameInfoResult 实名认证查询结果
type RealnameInfoResult struct {
	RealName       string // 姓名（脱敏）
	IdCardFrontUrl string // 身份证正面 URL
	IdCardBackUrl  string // 身份证反面 URL
	Status         int8   // 认证状态
}

// LicenseInfoResult 驾驶证认证查询结果
type LicenseInfoResult struct {
	LicenseNo   string // 驾驶证号
	LicenseType string // 准驾车型
	LicenseUrl  string // 驾驶证照片 URL
	Status      int8   // 认证状态
}

// VehicleInfoResult 车辆信息查询结果
type VehicleInfoResult struct {
	PlateNo           string // 车牌号
	VehicleBrand      string // 品牌
	VehicleModel      string // 型号
	VehicleColor      string // 颜色
	SeatCount         int8   // 座位数
	DrivingLicenseUrl string // 行驶证 URL
	VehiclePhotoUrl   string // 车辆实拍 URL
	Status            int8   // 认证状态
}

// IncomeDetailResult 收入分类汇总项
type IncomeDetailResult struct {
	TypeCode int8    // 类型编码（1=基础车费 2=奖励...）
	TypeName string // 类型中文名称
	Amount   float64 // 该类型总金额
	Count    int     // 该类型的笔数
}

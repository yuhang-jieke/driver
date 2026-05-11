package model

// ==================== 司机状态枚举 ====================
// 对应 drivers 表 status 字段

// DriverStatus 司机账号状态
const (
	DriverStatusNormal   int8 = 1 // 正常
	DriverStatusFrozen   int8 = 2 // 冻结
	DriverStatusClosed   int8 = 3 // 已注销
)

// DriverStatusDesc 状态描述映射
var DriverStatusDesc = map[int8]string{
	DriverStatusNormal: "正常",
	DriverStatusFrozen: "冻结",
	DriverStatusClosed: "已注销",
}

// ==================== 认证状态枚举 ====================
// 对应 drivers.verify_status / driver_realname.status / driver_license.status / driver_vehicle.status

// VerifyStatus 认证状态（实名/驾驶证/车辆通用）
const (
	VerifyStatusNone     int8 = 0 // 未认证
	VerifyStatusPending  int8 = 1 // 认证中/审核中
	VerifyStatusApproved int8 = 2 // 已认证/审核通过
	VerifyStatusRejected int8 = 3 // 认证失败/审核驳回
)

// VerifyStatusDesc 认证状态描述映射
var VerifyStatusDesc = map[int8]string{
	VerifyStatusNone:     "未认证",
	VerifyStatusPending:  "审核中",
	VerifyStatusApproved: "已认证",
	VerifyStatusRejected: "认证失败",
}

// IsVerifyEditable 审核中状态禁止重复提交，驳回后允许重新提交
func IsVerifyEditable(status int8) bool {
	return status != VerifyStatusPending
}

// ==================== 性别枚举 ====================

const (
	GenderUnknown int8 = 0 // 未知
	GenderMale    int8 = 1 // 男
	GenderFemale  int8 = 2 // 女
)

package model

import "time"

// ==================== 审计日志模块 (4张表) ====================

// DriverFaceAuthLog 人脸核验记录表 (driver_face_auth_log)
//
// 记录每次人脸核验的详情：
//   - AuthType: 1=出车验证 2=提现验证 3=登录验证
//   - Similarity: 人脸相似度百分比（如 95.6%）
//   - DeviceId/IP: 安全审计信息
type DriverFaceAuthLog struct {
	Id         int64     `gorm:"column:id;primaryKey;autoIncrement;comment:主键ID" json:"id"`
	DriverId   int64     `gorm:"column:driver_id;comment:司机ID" json:"driver_id"`
	AuthType   int8      `gorm:"column:auth_type;comment:核验类型: 1-出车验证 2-提现验证 3-登录验证" json:"auth_type"`
	AuthScene  string    `gorm:"column:auth_scene;comment:核验场景" json:"auth_scene"`
	FaceUrl    string    `gorm:"column:face_url;comment:人脸照片URL" json:"face_url"`
	Similarity float64   `gorm:"column:similarity;comment:相似度(%)" json:"similarity"`
	Status     int8      `gorm:"column:status;comment:核验结果: 1-通过 2-失败" json:"status"`
	FailReason string    `gorm:"column:fail_reason;comment:失败原因" json:"fail_reason"`
	DeviceId   string    `gorm:"column:device_id;comment:设备ID" json:"device_id"`
	Ip         string    `gorm:"column:ip;comment:IP地址" json:"ip"`
	CreatedAt  time.Time `gorm:"column:created_at;comment:创建时间" json:"created_at"`
}

func (DriverFaceAuthLog) TableName() string { return "driver_face_auth_log" }

// DriverStatusLog 司机状态变更日志 (driver_status_log)
// 记录司机账号状态的每一次变更（正常→冻结→注销等），支持审计追溯
type DriverStatusLog struct {
	Id         int64     `gorm:"column:id;primaryKey;autoIncrement;comment:主键ID" json:"id"`
	DriverId   int64     `gorm:"column:driver_id;comment:司机ID" json:"driver_id"`
	FromStatus int8      `gorm:"column:from_status;comment:变更前状态" json:"from_status"`
	ToStatus   int8      `gorm:"column:to_status;comment:变更后状态" json:"to_status"`
	Reason     string    `gorm:"column:reason;comment:变更原因" json:"reason"`
	Operator   string    `gorm:"column:operator;comment:操作人" json:"operator"`
	CreatedAt  time.Time `gorm:"column:created_at;comment:创建时间" json:"created_at"`
}

func (DriverStatusLog) TableName() string { return "driver_status_log" }

// DriverOnlineLog 出车记录表 (driver_online_log)
//
// 记录每次出车的完整周期：
//   - OnlineTime/OfflineTime: 上线/收车时间
//   - OnlineDuration: 在线时长（秒）
//   - OrderCount/Income: 该次出车完成的订单数和收入
//   - StartLat/Lng ~ EndLat/Lng: 起止位置坐标
type DriverOnlineLog struct {
	Id             int64     `gorm:"column:id;primaryKey;autoIncrement;comment:主键ID" json:"id"`
	DriverId       int64     `gorm:"column:driver_id;comment:司机ID" json:"driver_id"`
	OnlineTime     time.Time `gorm:"column:online_time;comment:出车时间" json:"online_time"`
	OfflineTime    time.Time `gorm:"column:offline_time;comment:收车时间" json:"offline_time"`
	OnlineDuration int       `gorm:"column:online_duration;comment:在线时长(秒)" json:"online_duration"`
	OrderCount     int       `gorm:"column:order_count;comment:完成订单数" json:"order_count"`
	Income         float64   `gorm:"column:income;comment:当日收入" json:"income"`
	StartLat       float64   `gorm:"column:start_lat;comment:起始纬度" json:"start_lat"`
	StartLng       float64   `gorm:"column:start_lng;comment:起始经度" json:"start_lng"`
	EndLat         float64   `gorm:"column:end_lat;comment:结束纬度" json:"end_lat"`
	EndLng         float64   `gorm:"column:end_lng;comment:结束经度" json:"end_lng"`
	CityId         int64     `gorm:"column:city_id;comment:服务城市ID" json:"city_id"`
	CreatedAt      time.Time `gorm:"column:created_at;comment:创建时间" json:"created_at"`
}

func (DriverOnlineLog) TableName() string { return "driver_online_log" }

// ServiceScoreLog 服务分变动记录 (service_score_log)
// 记录每次服务分的增减及原因：好评加分、投诉扣分、违规扣分等
type ServiceScoreLog struct {
	Id          int64     `gorm:"column:id;primaryKey;autoIncrement;comment:主键ID" json:"id"`
	DriverId    int64     `gorm:"column:driver_id;comment:司机ID" json:"driver_id"`
	ScoreBefore float64   `gorm:"column:score_before;comment:变更前分数" json:"score_before"`
	ScoreChange float64   `gorm:"column:score_change;comment:变更分数(正负)" json:"score_change"`
	ScoreAfter  float64   `gorm:"column:score_after;comment:变更后分数" json:"score_after"`
	ChangeType  int8      `gorm:"column:change_type;comment:变更类型: 1-好评加分 2-投诉扣分 3-违规扣分 4-系统调整" json:"change_type"`
	OrderId     int64     `gorm:"column:order_id;comment:关联订单ID" json:"orderId"`
	Remark      string    `gorm:"column:remark;comment:变更说明" json:"remark"`
	CreatedAt   time.Time `gorm:"column:created_at;comment:创建时间" json:"created_at"`
}

func (ServiceScoreLog) TableName() string { return "service_score_log" }

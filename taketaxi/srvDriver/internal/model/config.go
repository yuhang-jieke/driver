package model

import "time"

// ==================== 统计 & 配置 & 定位派单 (6张表) ====================

// DriverStatisticsSummary 司机统计汇总表 (driver_statistics_summary)
//
// 按天预聚合的统计数据，用于快速展示：
//   - StatDate: 统计日期（每天一条记录）
//   - 包含：在线时长、订单数（完成/取消/拒绝）、收入明细、里程、评分等
//   - 由离线任务或定时任务计算写入
type DriverStatisticsSummary struct {
	Id             int64     `gorm:"column:id;primaryKey;autoIncrement;comment:主键ID" json:"id"`
	DriverId       int64     `gorm:"column:driver_id;comment:司机ID" json:"driver_id"`
	StatDate       time.Time `gorm:"column:stat_date;comment:统计日期" json:"stat_date"`
	OnlineDuration int       `gorm:"column:online_duration;comment:在线时长(秒)" json:"online_duration"`
	OrderCount     int       `gorm:"column:order_count;comment:订单数" json:"order_count"`
	CompleteCount  int       `gorm:"column:complete_count;comment:完成订单数" json:"complete_count"`
	CancelCount    int       `gorm:"column:cancel_count;comment:取消订单数" json:"cancel_count"`
	RejectCount    int       `gorm:"column:reject_count;comment:拒绝订单数" json:"reject_count"`
	TotalIncome    float64   `gorm:"column:total_income;comment:总收入" json:"total_income"`
	OrderIncome    float64   `gorm:"column:order_income;comment:订单收入" json:"order_income"`
	BonusIncome    float64   `gorm:"column:bonus_income;comment:奖励收入" json:"bonus_income"`
	TotalDistance  int       `gorm:"column:total_distance;comment:总里程(米)" json:"total_distance"`
	TotalDuration  int       `gorm:"column:total_duration;comment:总时长(秒)" json:"total_duration"`
	AvgScore       float64   `gorm:"column:avg_score;comment:平均评分" json:"avg_score"`
	PraiseRate     float64   `gorm:"column:praise_rate;comment:好评率(%)" json:"praise_rate"`
	ComplaintCount int       `gorm:"column:complaint_count;comment:投诉数" json:"complaint_count"`
	CreatedAt      time.Time `gorm:"column:created_at;comment:创建时间" json:"created_at"`
	UpdatedAt      time.Time `gorm:"column:updated_at;comment:更新时间" json:"updated_at"`
}

func (DriverStatisticsSummary) TableName() string { return "driver_statistics_summary" }

// DriverLevelConfig 等级配置表 (driver_level_config)
// 配置各等级的门槛条件（最低服务分、最低订单数）和权益（佣金比例等）
// 全局配置，所有司机共享
type DriverLevelConfig struct {
	Id             int64     `gorm:"column:id;primaryKey;autoIncrement;comment:主键ID" json:"id"`
	Level          int8      `gorm:"column:level;comment:等级: 1-5" json:"level"`
	LevelName      string    `gorm:"column:level_name;comment:等级名称" json:"level_name"`
	MinScore       float64   `gorm:"column:min_score;comment:最低服务分" json:"min_score"`
	MaxScore       float64   `gorm:"column:max_score;comment:最高服务分" json:"max_score"`
	MinOrderCount  int       `gorm:"column:min_order_count;comment:最低订单数" json:"min_order_count"`
	CommissionRate float64   `gorm:"column:commission_rate;comment:佣金比例(%)" json:"commission_rate"`
	Benefits       string    `gorm:"column:benefits;comment:等级权益(JSON)" json:"benefits"`
	Status         int8      `gorm:"column:status;comment:状态: 1-启用 2-禁用" json:"status"`
	CreatedAt      time.Time `gorm:"column:created_at;comment:创建时间" json:"created_at"`
	UpdatedAt      time.Time `gorm:"column:updated_at;comment:更新时间" json:"updated_at"`
}

func (DriverLevelConfig) TableName() string { return "driver_level_config" }

// DriverLevelRecord 等级变动记录 (driver_level_record)
// 记录每个司机的等级升降历史
type DriverLevelRecord struct {
	Id         int64     `gorm:"column:id;primaryKey;autoIncrement;comment:主键ID" json:"id"`
	DriverId   int64     `gorm:"column:driver_id;comment:司机ID" json:"driver_id"`
	FromLevel  int8      `gorm:"column:from_level;comment:变更前等级" json:"from_level"`
	ToLevel    int8      `gorm:"column:to_level;comment:变更后等级" json:"to_level"`
	ChangeType int8      `gorm:"column:change_type;comment:变更类型: 1-升级 2-降级" json:"change_type"`
	Reason     string    `gorm:"column:reason;comment:变更原因" json:"reason"`
	CreatedAt  time.Time `gorm:"column:created_at;comment:创建时间" json:"created_at"`
}

func (DriverLevelRecord) TableName() string { return "driver_level_record" }

// PricingRuleConfig 计费规则配置表 (pricing_rule_config)
//
// 按 城市+服务类型 配置不同的计费规则：
//   - BasePrice/BaseDistance/BaseDuration: 起步价及包含的里程/时长
//   - DistancePrice/DurationPrice/WaitPrice: 各项单价
//   - NightStartTime/EndTime/NightRate: 夜间加价时段和倍率
//   - PeakRate: 高峰期动态调价倍率
//   - EffectiveTime/ExpireTime: 规则生效时间段
type PricingRuleConfig struct {
	Id               int64     `gorm:"column:id;primaryKey;autoIncrement;comment:主键ID" json:"id"`
	CityId           int64     `gorm:"column:city_id;comment:城市ID" json:"city_id"`
	ServiceType      int8      `gorm:"column:service_type;comment:服务类型: 1-快车 2-特惠快车" json:"service_type"`
	RuleName         string    `gorm:"column:rule_name;comment:规则名称" json:"rule_name"`
	BasePrice        float64   `gorm:"column:base_price;comment:起步价" json:"base_price"`
	BaseDistance     int       `gorm:"column:base_distance;comment:起步里程(米)" json:"base_distance"`
	BaseDuration     int       `gorm:"column:base_duration;comment:起步时长(秒)" json:"base_duration"`
	DistancePrice    float64   `gorm:"column:distance_price;comment:里程单价(元/公里)" json:"distance_price"`
	DurationPrice    float64   `gorm:"column:duration_price;comment:时长单价(元/分钟)" json:"duration_price"`
	WaitPrice        float64   `gorm:"column:wait_price;comment:等待单价(元/分钟)" json:"wait_price"`
	WaitFreeDuration int       `gorm:"column:wait_free_duration;comment:免费等待时长(秒)" json:"wait_free_duration"`
	NightStartTime   time.Time `gorm:"column:night_start_time;comment:夜间开始时间" json:"night_start_time"`
	NightEndTime     time.Time `gorm:"column:night_end_time;comment:夜间结束时间" json:"night_end_time"`
	NightRate        float64   `gorm:"column:night_rate;comment:夜间加价倍率" json:"night_rate"`
	PeakRate         float64   `gorm:"column:peak_rate;comment:高峰加价倍率" json:"peak_rate"`
	MinPrice         float64   `gorm:"column:min_price;comment:最低消费" json:"min_price"`
	DynamicPricing   int8      `gorm:"column:dynamic_pricing;comment:是否动态调价: 0-否 1-是" json:"dynamic_pricing"`
	Status           int8      `gorm:"column:status;comment:状态: 1-启用 2-禁用" json:"status"`
	EffectiveTime    time.Time `gorm:"column:effective_time;comment:生效时间" json:"effective_time"`
	ExpireTime       time.Time `gorm:"column:expire_time;comment:失效时间" json:"expire_time"`
	CreatedAt        time.Time `gorm:"column:created_at;comment:创建时间" json:"created_at"`
	UpdatedAt        time.Time `gorm:"column:updated_at;comment:更新时间" json:"updated_at"`
}

func (PricingRuleConfig) TableName() string { return "pricing_rule_config" }

// DriverLocationCache 司机实时位置表 (driver_location_cache)
//
// 高频更新表（每3~5秒更新一次），用于：
//   - 附近的司机查询（派单匹配）
//   - 司机实时轨迹追踪
//   - Status: 1=空车(可派单) 2=有客(行程中) 3=离线
type DriverLocationCache struct {
	Id        int64     `gorm:"column:id;primaryKey;autoIncrement;comment:主键ID" json:"id"`
	DriverId  int64     `gorm:"column:driver_id;comment:司机ID" json:"driver_id"`
	Lat       float64   `gorm:"column:lat;comment:当前纬度" json:"lat"`
	Lng       float64   `gorm:"column:lng;comment:当前经度" json:"lng"`
	Heading   float64   `gorm:"column:heading;comment:航向角(度)" json:"heading"`
	Speed     float64   `gorm:"column:speed;comment:速度(km/h)" json:"speed"`
	Accuracy  float64   `gorm:"column:accuracy;comment:精度(米)" json:"accuracy"`
	Status    int8      `gorm:"column:status;comment:状态: 1-空车 2-有客 3-离线" json:"status"`
	OrderId   int64     `gorm:"column:order_id;comment:当前订单ID" json:"order_id"`
	CityId    int64     `gorm:"column:city_id;comment:当前城市ID" json:"city_id"`
	UpdatedAt time.Time `gorm:"column:updated_at;comment:更新时间" json:"updated_at"`
}

func (DriverLocationCache) TableName() string { return "driver_location_cache" }

// DispatchLog 派单日志表 (dispatch_log)
//
// 记录每次派单的全生命周期：
//   - DispatchType: 1=指派(平台分配) 2=抢单(司机抢)
//   - Result: 1=接受 2=拒绝 3=超时未响应
//   - Distance: 司机到上车点的距离（米）
type DispatchLog struct {
	Id           int64     `gorm:"column:id;primaryKey;autoIncrement;comment:主键ID" json:"id"`
	OrderId      int64     `gorm:"column:order_id;comment:订单ID" json:"order_id"`
	DriverId     int64     `gorm:"column:driver_id;comment:司机ID" json:"driver_id"`
	DispatchType int8      `gorm:"column:dispatch_type;comment:派单类型: 1-指派 2-抢单" json:"dispatch_type"`
	DispatchTime time.Time `gorm:"column:dispatch_time;comment:派单时间" json:"dispatch_time"`
	ExpireTime   time.Time `gorm:"column:expire_time;comment:响应截止时间" json:"expire_time"`
	Result       int8      `gorm:"column:result;comment:结果: 1-接受 2-拒绝 3-超时" json:"result"`
	ResponseTime time.Time `gorm:"column:response_time;comment:司机响应时间" json:"response_time"`
	RejectReason string    `gorm:"column:reject_reason;comment:拒绝原因" json:"reject_reason"`
	DriverLat    float64   `gorm:"column:driver_lat;comment:司机当时纬度" json:"driver_lat"`
	DriverLng    float64   `gorm:"column:driver_lng;comment:司机当时经度" json:"driver_lng"`
	Distance     int       `gorm:"column:distance;comment:距离上车点(米)" json:"distance"`
	CreatedAt    time.Time `gorm:"column:created_at;comment:创建时间" json:"created_at"`
}

func (DispatchLog) TableName() string { return "dispatch_log" }

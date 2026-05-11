package model

import "time"

// ==================== 订单模块 (4张表) ====================

// Order 订单主表 (order)
//
// 核心业务表，包含完整的订单生命周期数据：
//   - 行程信息：起终点地址/坐标、预估/实际里程时长费用
//   - 费用明细：基础费+里程费+时长费+等待费-优惠券
//   - 支付信息：支付方式/支付时间/支付状态
//   - 取消信息：取消原因/取消方/取消时间
//
// 状态流转：0待派单→1已派单→2司机到达→3行程中→4已完成 / 5已取消
type Order struct {
	OrderId          int64     `gorm:"column:order_id;primaryKey;comment:订单ID" json:"order_id"`
	OrderNo          string    `gorm:"column:order_no;comment:订单编号" json:"order_no"`
	OrderType        int8      `gorm:"column:order_type;comment:订单类型: 1-即时单 2-预约单 3-拼车单" json:"order_type"`
	ServiceType      int8      `gorm:"column:service_type;comment:服务类型: 1-快车 2-特惠快车" json:"service_type"`
	Status           int8      `gorm:"column:status;comment:订单状态: 0-待派单 1-已派单 2-司机已到达 3-行程中 4-已完成 5-已取消" json:"status"`
	DriverId         int64     `gorm:"column:driver_id;comment:司机ID" json:"driver_id"`
	PassengerId      int64     `gorm:"column:passenger_id;comment:乘客ID" json:"passenger_id"`
	PassengerMobile  string    `gorm:"column:passenger_mobile;comment:乘客手机号(脱敏)" json:"passenger_mobile"`
	PassengerName    string    `gorm:"column:passenger_name;comment:乘客姓名" json:"passenger_name"`
	OriginAddress    string    `gorm:"column:origin_address;comment:起点地址" json:"origin_address"`
	OriginLat        float64   `gorm:"column:origin_lat;comment:起点纬度" json:"origin_lat"`
	OriginLng        float64   `gorm:"column:origin_lng;comment:起点经度" json:"origin_lng"`
	OriginPoi        string    `gorm:"column:origin_poi;comment:起点POI名称" json:"origin_poi"`
	DestAddress      string    `gorm:"column:dest_address;comment:终点地址" json:"dest_address"`
	DestLat          float64   `gorm:"column:dest_lat;comment:终点纬度" json:"dest_lat"`
	DestLng          float64   `gorm:"column:dest_lng;comment:终点经度" json:"dest_lng"`
	DestPoi          string    `gorm:"column:dest_poi;comment:终点POI名称" json:"dest_poi"`
	EstimateDistance int       `gorm:"column:estimate_distance;comment:预估里程(米)" json:"estimate_distance"`
	EstimateDuration int       `gorm:"column:estimate_duration;comment:预估时长(秒)" json:"estimate_duration"`
	EstimateFee      float64   `gorm:"column:estimate_fee;comment:预估费用" json:"estimate_fee"`
	ActualDistance   int       `gorm:"column:actual_distance;comment:实际里程(米)" json:"actual_distance"`
	ActualDuration   int       `gorm:"column:actual_duration;comment:实际时长(秒)" json:"actual_duration"`
	ActualFee        float64   `gorm:"column:actual_fee;comment:实际费用" json:"actual_fee"`
	BaseFee          float64   `gorm:"column:base_fee;comment:基础费用" json:"base_fee"`
	DistanceFee      float64   `gorm:"column:distance_fee;comment:里程费" json:"distance_fee"`
	DurationFee      float64   `gorm:"column:duration_fee;comment:时长费" json:"duration_fee"`
	WaitFee          float64   `gorm:"column:wait_fee;comment:等待费" json:"wait_fee"`
	CouponId         int64     `gorm:"column:coupon_id;comment:优惠券ID" json:"coupon_id"`
	CouponAmount     float64   `gorm:"column:coupon_amount;comment:优惠券金额" json:"coupon_amount"`
	PayStatus        int8      `gorm:"column:pay_status;comment:支付状态: 1-待支付 2-已支付 3-已退款" json:"pay_status"`
	PayType          int8      `gorm:"column:pay_type;comment:支付方式: 1-微信 2-支付宝 3-余额" json:"pay_type"`
	PayTime          time.Time `gorm:"column:pay_time;comment:支付时间" json:"pay_time"`
	CancelReason     string    `gorm:"column:cancel_reason;comment:取消原因" json:"cancel_reason"`
	CancelBy         int8      `gorm:"column:cancel_by;comment:取消方: 1-乘客 2-司机 3-系统" json:"cancel_by"`
	CancelTime       time.Time `gorm:"column:cancel_time;comment:取消时间" json:"cancel_time"`
	CityId           int64     `gorm:"column:city_id;comment:城市ID" json:"city_id"`
	CreatedAt        time.Time `gorm:"column:created_at;comment:创建时间" json:"created_at"`
	UpdatedAt        time.Time `gorm:"column:updated_at;comment:更新时间" json:"updated_at"`
}

func (Order) TableName() string { return "order" }

// OrderEvaluation 订单评价表 (order_evaluation)
// 双向评价：乘客评司机 + 司机评乘客
type OrderEvaluation struct {
	Id               int64     `gorm:"column:id;primaryKey;autoIncrement;comment:主键ID" json:"id"`
	OrderId          int64     `gorm:"column:order_id;comment:订单ID" json:"order_id"`
	DriverId         int64     `gorm:"column:driver_id;comment:司机ID" json:"driver_id"`
	PassengerId      int64     `gorm:"column:passenger_id;comment:乘客ID" json:"passenger_id"`
	DriverScore      int8      `gorm:"column:driver_score;comment:乘客评司机分数(1-5)" json:"driver_score"`
	DriverComment    string    `gorm:"column:driver_comment;comment:乘客评司机评价内容" json:"driver_comment"`
	DriverTags       string    `gorm:"column:driver_tags;comment:乘客评司机标签(JSON数组)" json:"driver_tags"`
	PassengerScore   int8      `gorm:"column:passenger_score;comment:司机评乘客分数(1-5)" json:"passenger_score"`
	PassengerComment string    `gorm:"column:passenger_comment;comment:司机评乘客评价内容" json:"passenger_comment"`
	IsAnonymous      int8      `gorm:"column:is_anonymous;comment:是否匿名评价: 0-否 1-是" json:"is_anonymous"`
	CreatedAt        time.Time `gorm:"column:created_at;comment:创建时间" json:"created_at"`
}

func (OrderEvaluation) TableName() string { return "order_evaluation" }

// TripService 行程服务表 (trip_service)
//
// 记录订单执行过程中的关键时间节点：
//   AcceptTime(接单) → ArriveTime(到达) → StartTime(开始行程) → EndTime(结束)
// 用于计算各环节耗时和服务质量评估
type TripService struct {
	Id           int64     `gorm:"column:id;primaryKey;autoIncrement;comment:主键ID" json:"id"`
	TripId       int64     `gorm:"column:trip_id;comment:行程ID" json:"trip_id"`
	OrderId      int64     `gorm:"column:order_id;comment:订单ID" json:"order_id"`
	DriverId     int64     `gorm:"column:driver_id;comment:司机ID" json:"driver_id"`
	PassengerId  int64     `gorm:"column:passenger_id;comment:乘客ID" json:"passenger_id"`
	AcceptTime   time.Time `gorm:"column:accept_time;comment:接单时间" json:"accept_time"`
	ArriveTime   time.Time `gorm:"column:arrive_time;comment:到达上车点时间" json:"arrive_time"`
	StartTime    time.Time `gorm:"column:start_time;comment:行程开始时间" json:"start_time"`
	EndTime      time.Time `gorm:"column:end_time;comment:行程结束时间" json:"end_time"`
	WaitDuration int       `gorm:"column:wait_duration;comment:等待乘客时长(秒)" json:"wait_duration"`
	TripDuration int       `gorm:"column:trip_duration;comment:行程时长(秒)" json:"trip_duration"`
	TripDistance int       `gorm:"column:trip_distance;comment:行程里程(米)" json:"trip_distance"`
	StartLat     float64   `gorm:"column:start_lat;comment:起点纬度" json:"start_lat"`
	StartLng     float64   `gorm:"column:start_lng;comment:起点经度" json:"start_lng"`
	EndLat       float64   `gorm:"column:end_lat;comment:终点纬度" json:"end_lat"`
	EndLng       float64   `gorm:"column:end_lng;comment:终点经度" json:"end_lng"`
	Status       int8      `gorm:"column:status;comment:状态: 1-前往上车点 2-已到达 3-行程中 4-已完成" json:"status"`
	CreatedAt    time.Time `gorm:"column:created_at;comment:创建时间" json:"created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at;comment:更新时间" json:"updated_at"`
}

func (TripService) TableName() string { return "trip_service" }

// TripTrajectory 行程轨迹归档表 (trip_trajectory)
//
// 存储完整行程的 GPS 轨迹数据：
//   - TrajectoryData: JSON 格式或 GZIP 压缩的轨迹点数组
//   - 大量轨迹数据可存至文件系统/OSS，此处只存 FileUrl
type TripTrajectory struct {
	Id             int64     `gorm:"column:id;primaryKey;autoIncrement;comment:主键ID" json:"id"`
	TripId         int64     `gorm:"column:trip_id;comment:行程ID" json:"trip_id"`
	OrderId        int64     `gorm:"column:order_id;comment:订单ID" json:"order_id"`
	DriverId       int64     `gorm:"column:driver_id;comment:司机ID" json:"driver_id"`
	TrajectoryData string    `gorm:"column:trajectory_data;comment:轨迹数据(JSON/GZIP压缩)" json:"trajectory_data"`
	PointCount     int       `gorm:"column:point_count;comment:轨迹点数量" json:"point_count"`
	StartTime      time.Time `gorm:"column:start_time;comment:轨迹开始时间" json:"start_time"`
	EndTime        time.Time `gorm:"column:end_time;comment:轨迹结束时间" json:"end_time"`
	Distance       int       `gorm:"column:distance;comment:轨迹总距离(米)" json:"distance"`
	FileUrl        string    `gorm:"column:file_url;comment:轨迹文件存储URL(大数据量时存文件)" json:"file_url"`
	CreatedAt      time.Time `gorm:"column:created_at;comment:创建时间" json:"created_at"`
}

func (TripTrajectory) TableName() string { return "trip_trajectory" }

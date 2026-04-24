package model

type Driver struct {
	ID     uint   `gorm:"primaryKey"`
	Name   string `gorm:"size:255"`
	Status int    `gorm:"default:0"`
}

func (Driver) TableName() string { return "" }

// Driver 司机基础信息表
type DriverS struct {
	DriverID              int64   `gorm:"column:driver_id;primaryKey;comment:司机唯一标识，分布式雪花算法生成" json:"driver_id"`
	Mobile                string  `gorm:"column:mobile;comment:登录手机号，用于账号登录和短信验证" json:"mobile"`
	MobileEncrypt         string  `gorm:"column:mobile_encrypt;comment:手机号AES加密存储，用于数据安全保护" json:"mobile_encrypt"`
	Nickname              string  `gorm:"column:nickname;comment:司机昵称，显示在乘客端和司机端" json:"nickname"`
	Avatar                string  `gorm:"column:avatar;comment:头像图片URL地址" json:"avatar"`
	Gender                int8    `gorm:"column:gender;comment:性别：0-未知，1-男，2-女" json:"gender"`
	Status                int8    `gorm:"column:status;comment:账号状态：1-正常，2-冻结，3-已注销" json:"status"`
	VerifyStatus          int8    `gorm:"column:verify_status;comment:认证状态：0-未认证，1-认证中，2-已认证，3-认证失败" json:"verify_status"`
	ServiceScore          float64 `gorm:"column:service_score;comment:服务分，综合评估司机服务质量的评分体系" json:"service_score"`
	OrderCount            int     `gorm:"column:order_count;comment:累计完成订单总数，用于统计展示" json:"order_count"`
	TotalIncome           float64 `gorm:"column:total_income;comment:累计收入总额，用于统计展示" json:"total_income"`
	RegisterSource        string  `gorm:"column:register_source;comment:注册来源渠道：APP-官方App，INVITE-邀请注册，ADS-广告投放" json:"register_source"`
	CityID                int64   `gorm:"column:city_id;comment:主要服务城市ID，关联城市表" json:"city_id"`
	LastOnlineAt          string  `gorm:"column:last_online_at;comment:最后出车时间，用于活跃度统计" json:"last_online_at"`
	DailyRejectCount      int     `gorm:"column:daily_reject_count;comment:当日拒绝订单次数，每日零点重置" json:"daily_reject_count"`
	ContinuousRejectCount int     `gorm:"column:continuous_reject_count;comment:连续拒绝订单次数，用于触发暂停派单惩罚" json:"continuous_reject_count"`
	CreatedAt             string  `gorm:"column:created_at;comment:记录创建时间" json:"created_at"`
	UpdatedAt             string  `gorm:"column:updated_at;comment:记录更新时间" json:"updated_at"`
	DeletedAt             string  `gorm:"column:deleted_at;comment:软删除时间，非空表示已删除" json:"deleted_at"`
}

func (DriverS) TableName() string {
	return "driver"
}

// DriverRealname 司机实名认证表
type DriverRealname struct {
	RealnameID      int64  `gorm:"column:realname_id;primaryKey;autoIncrement;comment:实名认证记录唯一标识" json:"realname_id"`
	DriverID        int64  `gorm:"column:driver_id;comment:关联司机ID，外键关联driver表" json:"driver_id"`
	RealName        string `gorm:"column:real_name;comment:真实姓名，从身份证OCR识别" json:"real_name"`
	RealNameEncrypt string `gorm:"column:real_name_encrypt;comment:姓名AES加密存储，用于数据安全保护" json:"real_name_encrypt"`
	IdCardNo        string `gorm:"column:id_card_no;comment:身份证号码，18位标准格式" json:"id_card_no"`
	IdCardNoEncrypt string `gorm:"column:id_card_no_encrypt;comment:身份证号AES加密存储" json:"id_card_no_encrypt"`
	IdCardFrontUrl  string `gorm:"column:id_card_front_url;comment:身份证正面照片URL，存储在OSS" json:"id_card_front_url"`
	IdCardBackUrl   string `gorm:"column:id_card_back_url;comment:身份证反面照片URL，存储在OSS" json:"id_card_back_url"`
	Gender          int8   `gorm:"column:gender;comment:性别，从身份证解析：1-男，2-女" json:"gender"`
	Birthday        string `gorm:"column:birthday;comment:出生日期，从身份证解析" json:"birthday"`
	Address         string `gorm:"column:address;comment:身份证登记地址" json:"address"`
	Nation          string `gorm:"column:nation;comment:民族，如：汉、回、满等" json:"nation"`
	ExpireDate      string `gorm:"column:expire_date;comment:身份证有效期截止日期" json:"expire_date"`
	Status          int8   `gorm:"column:status;comment:认证状态：0-未认证，1-认证中，2-已认证，3-认证失败" json:"status"`
	FailReason      string `gorm:"column:fail_reason;comment:认证失败原因，审核拒绝时填写" json:"fail_reason"`
	VerifyTime      string `gorm:"column:verify_time;comment:认证完成时间，审核通过或拒绝的时间" json:"verify_time"`
	CreatedAt       string `gorm:"column:created_at;comment:记录创建时间" json:"created_at"`
	UpdatedAt       string `gorm:"column:updated_at;comment:记录更新时间" json:"updated_at"`
}

func (DriverRealname) TableName() string {
	return "driver_realname"
}

// DriverLicense 司机驾驶证表
type DriverLicense struct {
	LicenseID      int64  `gorm:"column:license_id;primaryKey;autoIncrement;comment:驾驶证认证记录唯一标识" json:"license_id"`
	DriverID       int64  `gorm:"column:driver_id;comment:关联司机ID" json:"driver_id"`
	LicenseNo      string `gorm:"column:license_no;comment:驾驶证编号，18位格式" json:"license_no"`
	LicenseType    string `gorm:"column:license_type;comment:准驾车型：C1-小型汽车，C2-小型自动挡，B1-中型客车，B2-大型货车，A1-大型客车，A2-牵引车" json:"license_type"`
	LicenseUrl     string `gorm:"column:license_url;comment:驾驶证照片URL，存储在OSS" json:"license_url"`
	FirstIssueDate string `gorm:"column:first_issue_date;comment:初次领证日期，用于计算驾龄，平台要求满3年" json:"first_issue_date"`
	IssueDate      string `gorm:"column:issue_date;comment:当前证件发证日期，换证后更新" json:"issue_date"`
	ExpireDate     string `gorm:"column:expire_date;comment:驾驶证有效期截止日期" json:"expire_date"`
	DrivingYears   int8   `gorm:"column:driving_years;comment:驾龄年数，系统自动计算，平台要求>=3年" json:"driving_years"`
	Status         int8   `gorm:"column:status;comment:认证状态：0-未认证，1-认证中，2-已认证，3-认证失败" json:"status"`
	FailReason     string `gorm:"column:fail_reason;comment:认证失败原因，如驾龄不足" json:"fail_reason"`
	VerifyTime     string `gorm:"column:verify_time;comment:认证完成时间" json:"verify_time"`
	CreatedAt      string `gorm:"column:created_at;comment:记录创建时间" json:"created_at"`
	UpdatedAt      string `gorm:"column:updated_at;comment:记录更新时间" json:"updated_at"`
}

func (DriverLicense) TableName() string {
	return "driver_license"
}

// DriverVehicle 司机车辆信息表
type DriverVehicle struct {
	VehicleID         int64  `gorm:"column:vehicle_id;primaryKey;autoIncrement;comment:车辆信息记录唯一标识" json:"vehicle_id"`
	DriverID          int64  `gorm:"column:driver_id;comment:关联司机ID" json:"driver_id"`
	PlateNo           string `gorm:"column:plate_no;comment:车牌号码，如：京A12345" json:"plate_no"`
	PlateNoEncrypt    string `gorm:"column:plate_no_encrypt;comment:车牌号AES加密存储" json:"plate_no_encrypt"`
	VehicleModel      string `gorm:"column:vehicle_model;comment:车型名称，如：大众朗逸" json:"vehicle_model"`
	VehicleBrand      string `gorm:"column:vehicle_brand;comment:车辆品牌，如：大众" json:"vehicle_brand"`
	VehicleColor      string `gorm:"column:vehicle_color;comment:车身颜色，如：白色、黑色" json:"vehicle_color"`
	VehicleColorCode  string `gorm:"column:vehicle_color_code;comment:颜色十六进制代码，如：#FFFFFF" json:"vehicle_color_code"`
	SeatCount         int8   `gorm:"column:seat_count;comment:核定载人数，平台要求7座及以下" json:"seat_count"`
	RegisterDate      string `gorm:"column:register_date;comment:车辆注册登记日期，用于计算车龄" json:"register_date"`
	VehicleAge        int8   `gorm:"column:vehicle_age;comment:车龄年数，系统自动计算，平台要求<=8年" json:"vehicle_age"`
	DrivingLicenseUrl string `gorm:"column:driving_license_url;comment:行驶证照片URL，存储在OSS" json:"driving_license_url"`
	VehiclePhotoUrl   string `gorm:"column:vehicle_photo_url;comment:车辆外观照片URL" json:"vehicle_photo_url"`
	Status            int8   `gorm:"column:status;comment:认证状态：0-未认证，1-认证中，2-已认证，3-认证失败" json:"status"`
	FailReason        string `gorm:"column:fail_reason;comment:认证失败原因，如车龄超限" json:"fail_reason"`
	VerifyTime        string `gorm:"column:verify_time;comment:认证完成时间" json:"verify_time"`
	CreatedAt         string `gorm:"column:created_at;comment:记录创建时间" json:"created_at"`
	UpdatedAt         string `gorm:"column:updated_at;comment:记录更新时间" json:"updated_at"`
}

func (DriverVehicle) TableName() string {
	return "driver_vehicle"
}

// DriverFace 司机人脸信息表
type DriverFace struct {
	FaceID      int64  `gorm:"column:face_id;primaryKey;autoIncrement;comment:人脸信息记录唯一标识" json:"face_id"`
	DriverID    int64  `gorm:"column:driver_id;comment:关联司机ID" json:"driver_id"`
	FaceUrl     string `gorm:"column:face_url;comment:人脸照片URL，存储在OSS" json:"face_url"`
	FaceFeature string `gorm:"column:face_feature;comment:人脸特征向量，加密存储用于人脸比对" json:"face_feature"`
	Status      int8   `gorm:"column:status;comment:状态：1-有效，2-失效" json:"status"`
	VerifyTime  string `gorm:"column:verify_time;comment:人脸验证时间" json:"verify_time"`
	ExpireTime  string `gorm:"column:expire_time;comment:验证结果过期时间，有效期7天" json:"expire_time"`
	CreatedAt   string `gorm:"column:created_at;comment:记录创建时间" json:"created_at"`
	UpdatedAt   string `gorm:"column:updated_at;comment:记录更新时间" json:"updated_at"`
}

func (DriverFace) TableName() string {
	return "driver_face"
}

// DriverStatusLog 司机状态变更日志表
type DriverStatusLog struct {
	StatusLogID int64  `gorm:"column:status_log_id;primaryKey;autoIncrement;comment:状态日志唯一标识" json:"status_log_id"`
	DriverID    int64  `gorm:"column:driver_id;comment:关联司机ID" json:"driver_id"`
	FromStatus  int8   `gorm:"column:from_status;comment:变更前状态：1-正常，2-冻结，3-已注销" json:"from_status"`
	ToStatus    int8   `gorm:"column:to_status;comment:变更后状态" json:"to_status"`
	Reason      string `gorm:"column:reason;comment:状态变更原因" json:"reason"`
	Operator    string `gorm:"column:operator;comment:操作人，系统操作或管理员账号" json:"operator"`
	CreatedAt   string `gorm:"column:created_at;comment:记录创建时间" json:"created_at"`
}

func (DriverStatusLog) TableName() string {
	return "driver_status_log"
}

// DriverOnlineLog 司机出车记录表
type DriverOnlineLog struct {
	OnlineLogID    int64   `gorm:"column:online_log_id;primaryKey;autoIncrement;comment:出车记录唯一标识" json:"online_log_id"`
	DriverID       int64   `gorm:"column:driver_id;comment:关联司机ID" json:"driver_id"`
	OnlineTime     string  `gorm:"column:online_time;comment:出车时间，司机开始接单的时间" json:"online_time"`
	OfflineTime    string  `gorm:"column:offline_time;comment:收车时间，司机结束接单的时间" json:"offline_time"`
	OnlineDuration int     `gorm:"column:online_duration;comment:在线时长(秒)，单次出车最长12小时" json:"online_duration"`
	OrderCount     int     `gorm:"column:order_count;comment:当次出车完成订单数" json:"order_count"`
	Income         float64 `gorm:"column:income;comment:当次出车收入总额" json:"income"`
	StartLat       float64 `gorm:"column:start_lat;comment:出车位置纬度" json:"start_lat"`
	StartLng       float64 `gorm:"column:start_lng;comment:出车位置经度" json:"start_lng"`
	EndLat         float64 `gorm:"column:end_lat;comment:收车位置纬度" json:"end_lat"`
	EndLng         float64 `gorm:"column:end_lng;comment:收车位置经度" json:"end_lng"`
	CityID         int64   `gorm:"column:city_id;comment:服务城市ID" json:"city_id"`
	CreatedAt      string  `gorm:"column:created_at;comment:记录创建时间" json:"created_at"`
}

func (DriverOnlineLog) TableName() string {
	return "driver_online_log"
}

// DriverRejectLog 司机拒单记录表
type DriverRejectLog struct {
	RejectLogID  int64  `gorm:"column:reject_log_id;primaryKey;autoIncrement;comment:拒单记录唯一标识" json:"reject_log_id"`
	DriverID     int64  `gorm:"column:driver_id;comment:关联司机ID" json:"driver_id"`
	OrderID      int64  `gorm:"column:order_id;comment:被拒绝的订单ID" json:"order_id"`
	RejectReason string `gorm:"column:reject_reason;comment:拒绝原因，司机选择或填写" json:"reject_reason"`
	RejectTime   string `gorm:"column:reject_time;comment:拒绝时间" json:"reject_time"`
	CreatedAt    string `gorm:"column:created_at;comment:记录创建时间" json:"created_at"`
}

func (DriverRejectLog) TableName() string {
	return "driver_reject_log"
}

// TripService 行程服务表
type TripService struct {
	TripServiceID int64   `gorm:"column:trip_service_id;primaryKey;autoIncrement;comment:行程服务记录唯一标识" json:"trip_service_id"`
	TripID        int64   `gorm:"column:trip_id;comment:关联行程ID，与行程系统对接" json:"trip_id"`
	OrderID       int64   `gorm:"column:order_id;comment:关联订单ID" json:"order_id"`
	DriverID      int64   `gorm:"column:driver_id;comment:关联司机ID" json:"driver_id"`
	PassengerID   int64   `gorm:"column:passenger_id;comment:关联乘客ID" json:"passenger_id"`
	AcceptTime    string  `gorm:"column:accept_time;comment:接单时间，司机确认接受订单的时间" json:"accept_time"`
	ArriveTime    string  `gorm:"column:arrive_time;comment:到达时间，司机到达上车点的时间" json:"arrive_time"`
	StartTime     string  `gorm:"column:start_time;comment:开始时间，乘客上车开始行程的时间" json:"start_time"`
	EndTime       string  `gorm:"column:end_time;comment:结束时间，到达目的地结束行程的时间" json:"end_time"`
	WaitDuration  int     `gorm:"column:wait_duration;comment:等待时长(秒)，等待乘客上车的时长" json:"wait_duration"`
	TripDuration  int     `gorm:"column:trip_duration;comment:行程时长(秒)，从开始到结束的时长" json:"trip_duration"`
	TripDistance  int     `gorm:"column:trip_distance;comment:行程里程(米)，从起点到终点的距离" json:"trip_distance"`
	StartLat      float64 `gorm:"column:start_lat;comment:行程起点纬度" json:"start_lat"`
	StartLng      float64 `gorm:"column:start_lng;comment:行程起点经度" json:"start_lng"`
	EndLat        float64 `gorm:"column:end_lat;comment:行程终点纬度" json:"end_lat"`
	EndLng        float64 `gorm:"column:end_lng;comment:行程终点经度" json:"end_lng"`
	Status        int8    `gorm:"column:status;comment:行程状态：1-前往上车点，2-已到达，3-行程中，4-已完成，5-已取消" json:"status"`
	CancelReason  string  `gorm:"column:cancel_reason;comment:取消原因，行程取消时填写" json:"cancel_reason"`
	CreatedAt     string  `gorm:"column:created_at;comment:记录创建时间" json:"created_at"`
	UpdatedAt     string  `gorm:"column:updated_at;comment:记录更新时间" json:"updated_at"`
}

func (TripService) TableName() string {
	return "trip_service"
}

// TripTrack 行程轨迹表
type TripTrack struct {
	TrackID    int64   `gorm:"column:track_id;primaryKey;autoIncrement;comment:轨迹点唯一标识" json:"track_id"`
	TripID     int64   `gorm:"column:trip_id;comment:关联行程ID" json:"trip_id"`
	DriverID   int64   `gorm:"column:driver_id;comment:关联司机ID" json:"driver_id"`
	Latitude   float64 `gorm:"column:latitude;comment:轨迹点纬度" json:"latitude"`
	Longitude  float64 `gorm:"column:longitude;comment:轨迹点经度" json:"longitude"`
	Speed      int     `gorm:"column:speed;comment:速度(km/h)" json:"speed"`
	Direction  int     `gorm:"column:direction;comment:方向角度(0-360)" json:"direction"`
	RecordTime string  `gorm:"column:record_time;comment:记录时间" json:"record_time"`
	CreatedAt  string  `gorm:"column:created_at;comment:记录创建时间" json:"created_at"`
}

func (TripTrack) TableName() string {
	return "trip_track"
}

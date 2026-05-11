package model

import "time"

// ==================== 认证资料模块 (5张表) ====================
// 三级认证体系：实名 → 驾驶证 → 车辆 → 人脸（可选）
// 每级认证独立审核，可并行进行

// DriverRealname 实名认证表 (driver_realname)
//
// 存储司机实名认证所需的信息：
//   - RealName + IdCardNo: 身份证姓名和号码
//   - IdCardFrontUrl / IdCardBackUrl: 身份证正反面照片（上传至 OSS）
//   - ExpireDate: 身份证有效期，用于判断是否过期需重新认证
//
// 安全提示：生产环境中 RealName 和 IdCardNo 应在入库前做 AES 加密
type DriverRealname struct {
	Id              int64     `gorm:"column:id;primaryKey;autoIncrement;comment:主键ID" json:"id"`
	DriverId        int64     `gorm:"column:driver_id;comment:司机ID" json:"driver_id"`
	RealName        string    `gorm:"column:real_name;comment:真实姓名" json:"real_name"`
	RealNameEncrypt string    `gorm:"column:real_name_encrypt;comment:姓名AES加密" json:"real_name_encrypt"`
	IdCardNo        string    `gorm:"column:id_card_no;comment:身份证号" json:"id_card_no"`
	IdCardNoEncrypt string    `gorm:"column:id_card_no_encrypt;comment:身份证号AES加密" json:"id_card_no_encrypt"`
	IdCardFrontUrl  string    `gorm:"column:id_card_front_url;comment:身份证正面照片URL" json:"id_card_front_url"`
	IdCardBackUrl   string    `gorm:"column:id_card_back_url;comment:身份证反面照片URL" json:"id_card_back_url"`
	Gender          int8      `gorm:"column:gender;comment:性别: 1-男 2-女" json:"gender"`
	Birthday        time.Time `gorm:"column:birthday;comment:出生日期" json:"birthday"`
	Address         string    `gorm:"column:address;comment:身份证地址" json:"address"`
	Nation          string    `gorm:"column:nation;comment:民族" json:"nation"`
	ExpireDate      time.Time `gorm:"column:expire_date;comment:身份证有效期" json:"expire_date"`
	Status          int8      `gorm:"column:status;comment:认证状态: 0-未认证 1-认证中 2-已认证 3-认证失败" json:"status"`
	FailReason      string    `gorm:"column:fail_reason;comment:认证失败原因" json:"fail_reason"`
	VerifyTime      time.Time `gorm:"column:verify_time;comment:认证完成时间" json:"verify_time"`
	CreatedAt       time.Time `gorm:"column:created_at;comment:创建时间" json:"created_at"`
	UpdatedAt       time.Time `gorm:"column:updated_at;comment:更新时间" json:"updated_at"`
}

func (DriverRealname) TableName() string { return "driver_realname" }

// DriverLicense 驾驶证认证表 (driver_license)
//
// 存储驾驶证相关信息：
//   - LicenseType: 准驾车型（C1/C2/B1/B2/A1/A2 等）
//   - FirstIssueDate: 初次领证日期 → 可计算驾龄
//   - IssueDate / ExpireDate: 当前证件有效期范围
type DriverLicense struct {
	Id             int64     `gorm:"column:id;primaryKey;autoIncrement;comment:主键ID" json:"id"`
	DriverId       int64     `gorm:"column:driver_id;comment:司机ID" json:"driver_id"`
	LicenseNo      string    `gorm:"column:license_no;comment:驾驶证编号" json:"license_no"`
	LicenseType    string    `gorm:"column:license_type;comment:准驾车型: C1/C2/B1/B2/A1/A2" json:"license_type"`
	LicenseUrl     string    `gorm:"column:license_url;comment:驾驶证照片URL" json:"license_url"`
	FirstIssueDate time.Time `gorm:"column:first_issue_date;comment:初次领证日期" json:"first_issue_date"`
	IssueDate      time.Time `gorm:"column:issue_date;comment:当前证件发证日期" json:"issue_date"`
	ExpireDate     time.Time `gorm:"column:expire_date;comment:驾驶证有效期" json:"expire_date"`
	DrivingYears   int8      `gorm:"column:driving_years;comment:驾龄(年)" json:"driving_years"`
	Status         int8      `gorm:"column:status;comment:认证状态: 0-未认证 1-认证中 2-已认证 3-认证失败" json:"status"`
	FailReason     string    `gorm:"column:fail_reason;comment:认证失败原因" json:"fail_reason"`
	VerifyTime     time.Time `gorm:"column:verify_time;comment:认证完成时间" json:"verify_time"`
	CreatedAt      time.Time `gorm:"column:created_at;comment:创建时间" json:"created_at"`
	UpdatedAt      time.Time `gorm:"column:updated_at;comment:更新时间" json:"updated_at"`
}

func (DriverLicense) TableName() string { return "driver_license" }

// DriverVehicle 车辆信息表 (driver_vehicle)
//
// 存储司机用于接单的车辆信息：
//   - PlateNo: 车牌号（唯一性校验）
//   - VehicleBrand/Model/Color: 车辆外观描述
//   - SeatCount: 座位数影响最大载客数
//   - DrivingLicenseUrl/VehiclePhotoUrl: 行驶证和实拍照片
type DriverVehicle struct {
	Id                int64     `gorm:"column:id;primaryKey;autoIncrement;comment:主键ID" json:"id"`
	DriverId          int64     `gorm:"column:driver_id;comment:司机ID" json:"driver_id"`
	PlateNo           string    `gorm:"column:plate_no;comment:车牌号" json:"plate_no"`
	PlateNoEncrypt    string    `gorm:"column:plate_no_encrypt;comment:车牌号AES加密" json:"plate_no_encrypt"`
	VehicleModel      string    `gorm:"column:vehicle_model;comment:车型名称" json:"vehicle_model"`
	VehicleBrand      string    `gorm:"column:vehicle_brand;comment:车辆品牌" json:"vehicle_brand"`
	VehicleColor      string    `gorm:"column:vehicle_color;comment:车身颜色" json:"vehicle_color"`
	VehicleColorCode  string    `gorm:"column:vehicle_color_code;comment:颜色代码(十六进制)" json:"vehicle_color_code"`
	SeatCount         int8      `gorm:"column:seat_count;comment:核定载人数" json:"seat_count"`
	RegisterDate      time.Time `gorm:"column:register_date;comment:车辆注册日期" json:"register_date"`
	VehicleAge        int8      `gorm:"column:vehicle_age;comment:车龄(年)" json:"vehicle_age"`
	DrivingLicenseUrl string    `gorm:"column:driving_license_url;comment:行驶证照片URL" json:"driving_license_url"`
	VehiclePhotoUrl   string    `gorm:"column:vehicle_photo_url;comment:车辆外观照片URL" json:"vehicle_photo_url"`
	Status            int8      `gorm:"column:status;comment:认证状态: 0-未认证 1-认证中 2-已认证 3-认证失败" json:"status"`
	FailReason        string    `gorm:"column:fail_reason;comment:认证失败原因" json:"fail_reason"`
	VerifyTime        time.Time `gorm:"column:verify_time;comment:认证完成时间" json:"verify_time"`
	CreatedAt         time.Time `gorm:"column:created_at;comment:创建时间" json:"created_at"`
	UpdatedAt         time.Time `gorm:"column:updated_at;comment:更新时间" json:"updated_at"`
}

func (DriverVehicle) TableName() string { return "driver_vehicle" }

// DriverVehicleInfo 车辆详细信息表 (driver_vehicle_info)
// 存储车辆的详细技术参数，与 DriverVehicle 是 1:1 关系
type DriverVehicleInfo struct {
	Id                   int64     `gorm:"column:id;primaryKey;autoIncrement;comment:主键ID" json:"id"`
	DriverId             int64     `gorm:"column:driver_id;comment:司机ID" json:"driver_id"`
	VehicleId            int64     `gorm:"column:vehicle_id;comment:关联车辆ID" json:"vehicle_id"`
	Vin                  string    `gorm:"column:vin;comment:车辆识别代号VIN" json:"vin"`
	EngineNo             string    `gorm:"column:engine_no;comment:发动机号" json:"engine_no"`
	VehicleType          string    `gorm:"column:vehicle_type;comment:车辆类型" json:"vehicle_type"`
	Displacement         float64   `gorm:"column:displacement;comment:排量(L)" json:"displacement"`
	FuelType             string    `gorm:"column:fuel_type;comment:燃料类型: 汽油/柴油/电动/混动" json:"fuel_type"`
	InspectionExpireDate time.Time `gorm:"column:inspection_expire_date;comment:年检有效期" json:"inspection_expire_date"`
	InsuranceCompany     string    `gorm:"column:insurance_company;comment:保险公司" json:"insurance_company"`
	InsuranceExpireDate  time.Time `gorm:"column:insurance_expire_date;comment:保险有效期" json:"insurance_expire_date"`
	InsuranceUrl         string    `gorm:"column:insurance_url;comment:保险单照片URL" json:"insurance_url"`
	CreatedAt            time.Time `gorm:"column:created_at;comment:创建时间" json:"created_at"`
	UpdatedAt            time.Time `gorm:"column:updated_at;comment:更新时间" json:"updated_at"`
}

func (DriverVehicleInfo) TableName() string { return "driver_vehicle_info" }

// DriverFace 人脸信息表 (driver_face)
//
// 存储司机人脸识别模板：
//   - FaceFeature: 人脸特征向量（通常为二进制/加密存储），用于出车/提现时的人脸核验
//   - FaceUrl: 最近一次成功采集的人脸照片
//   - ExpireTime: 人脸模板过期时间，过期后需重新采集
type DriverFace struct {
	Id          int64     `gorm:"column:id;primaryKey;autoIncrement;comment:主键ID" json:"id"`
	DriverId    int64     `gorm:"column:driver_id;comment:司机ID" json:"driver_id"`
	FaceUrl     string    `gorm:"column:face_url;comment:人脸照片URL" json:"face_url"`
	FaceFeature string    `gorm:"column:face_feature;comment:人脸特征向量(加密)" json:"face_feature"`
	Status      int8      `gorm:"column:status;comment:状态: 1-有效 2-失效" json:"status"`
	VerifyTime  time.Time `gorm:"column:verify_time;comment:人脸验证时间" json:"verify_time"`
	ExpireTime  time.Time `gorm:"column:expire_time;comment:人脸验证结果过期时间" json:"expire_time"`
	CreatedAt   time.Time `gorm:"column:created_at;comment:创建时间" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at;comment:更新时间" json:"updated_at"`
}

func (DriverFace) TableName() string { return "driver_face" }

// Package errcode 定义司机端统一错误码体系，
// 遵循 go-driver-dev-spec.md §9 错误码规范：
//
//	0      : 成功
//	4xxxx  : 参数与请求错误
//	5xxxx  : 业务规则错误
//	6xxxx  : 依赖服务错误
//	9xxxx  : 系统内部错误
package errcode

import (
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ==================== 错误码定义 ====================

// Code 错误码类型
type Code int

const (
	// Success 成功
	Success Code = 0

	// ---- 4xxxx 参数与请求错误 ----
	ErrInvalidParam      Code = 40001 // 通用参数非法
	ErrInvalidDriverID   Code = 40002 // 司机ID非法
	ErrInvalidDate       Code = 40003 // 日期格式非法
	ErrInvalidDays       Code = 40004 // 天数参数非法
	ErrMissingField      Code = 40005 // 必填字段缺失
	ErrBankCardNoTooLong  Code = 40006 // 银行卡号超长
	ErrInvalidAccountName Code = 40007 // 持卡人姓名格式不正确

	// ---- 5xxxx 业务规则错误 ----
	ErrDriverNotFound       Code = 50001 // 司机不存在
	ErrDriverStatusInvalid  Code = 50002 // 司机状态不允许操作
	ErrVerifyPending        Code = 50003 // 认证审核中，禁止重复提交
	ErrVerifyNotApproved    Code = 50004 // 认证未通过
	ErrBankCardAlreadyBound Code = 50005 // 银行卡已绑定
	ErrBankCardNotBound     Code = 50006 // 未绑定银行卡
	ErrBankCardChangeLimit  Code = 50007 // 银行卡更换频次超限
	ErrWalletNotFound       Code = 50008 // 钱包账户不存在
	ErrInsufficientBalance  Code = 50009 // 可提现余额不足
	ErrFrozenAmountLimit    Code = 50010 // 冻结金额限制
	ErrWithdrawAmountLimit  Code = 50011 // 提现金额超限
	ErrWithdrawCountLimit   Code = 50012 // 今日提现次数超限
	ErrMobileChangeLimit    Code = 50013 // 手机号修改频次超限
	ErrPasswordNotSet       Code = 50014 // 密码未设置
	ErrOldPasswordWrong     Code = 50015 // 原密码错误
	ErrPasswordSameAsOld    Code = 50016 // 新密码与原密码相同
	ErrNoUpdateFields       Code = 50017 // 无更新字段
	ErrMobileNotRegistered  Code = 50018 // 手机号未注册
	ErrNoUploadFields        Code = 50019 // 无上传内容
	ErrWithdrawMinAmount       Code = 50020 // 提现金额低于最低限额
	ErrWithdrawRealnameNeeded  Code = 50021 // 提现需完成实名认证
	ErrWithdrawPageUnavailable Code = 50022 // 提现页面不可用
	ErrWithdrawCreditCard      Code = 50023 // 信用卡不允许提现
	ErrWithdrawCardNameMismatch Code = 50024 // 银行卡姓名与实名不一致

	// ---- 6xxxx 依赖服务错误 ----
	ErrRedisError    Code = 60001 // Redis 异常
	ErrSMSSendFailed Code = 60002 // 短信发送失败

	// ---- 9xxxx 系统内部错误 ----
	ErrInternal      Code = 90001 // 系统内部错误
	ErrPasswordHash  Code = 90002 // 密码加密失败
	ErrCreateRecord  Code = 90003 // 创建记录失败
	ErrUpdateBalance Code = 90004 // 余额更新失败
)

// ==================== 错误码描述映射 ====================

var codeMessages = map[Code]string{
	Success:                "成功",
	ErrInvalidParam:        "参数非法",
	ErrInvalidDriverID:     "司机ID非法",
	ErrInvalidDate:         "日期格式非法",
	ErrInvalidDays:         "天数参数非法",
	ErrMissingField:        "必填字段缺失",
	ErrBankCardNoTooLong:   "银行卡号超长",
	ErrInvalidAccountName:  "持卡人姓名格式不正确，必须为2-6个汉字",
	ErrDriverNotFound:      "司机不存在",
	ErrDriverStatusInvalid: "司机状态不允许操作",
	ErrVerifyPending:       "认证审核中，禁止重复提交",
	ErrVerifyNotApproved:   "认证未通过",
	ErrBankCardAlreadyBound: "银行卡已绑定",
	ErrBankCardNotBound:    "未绑定银行卡",
	ErrBankCardChangeLimit: "银行卡更换频次超限",
	ErrWalletNotFound:      "钱包账户不存在",
	ErrInsufficientBalance: "可提现余额不足",
	ErrFrozenAmountLimit:   "冻结金额限制",
	ErrWithdrawAmountLimit: "提现金额超限",
	ErrWithdrawCountLimit:  "今日提现次数超限",
	ErrMobileChangeLimit:   "手机号修改频次超限",
	ErrPasswordNotSet:      "密码未设置",
	ErrOldPasswordWrong:    "原密码错误",
	ErrPasswordSameAsOld:   "新密码与原密码相同",
	ErrNoUpdateFields:      "无更新字段",
	ErrMobileNotRegistered: "手机号未注册",
	ErrNoUploadFields:        "无上传内容",
	ErrWithdrawMinAmount:        "提现金额低于最低限额",
	ErrWithdrawRealnameNeeded:   "提现需完成实名认证",
	ErrWithdrawPageUnavailable:  "提现页面不可用",
	ErrWithdrawCreditCard:       "信用卡不允许提现",
	ErrWithdrawCardNameMismatch: "银行卡姓名与实名不一致",
	ErrRedisError:          "缓存服务异常",
	ErrSMSSendFailed:       "短信发送失败",
	ErrInternal:            "系统内部错误",
	ErrPasswordHash:        "密码加密失败",
	ErrCreateRecord:        "创建记录失败",
	ErrUpdateBalance:       "余额更新失败",
}

// Message 获取错误码描述
func (c Code) Message() string {
	if msg, ok := codeMessages[c]; ok {
		return msg
	}
	return "未知错误"
}

// Int 返回错误码整数值
func (c Code) Int() int {
	return int(c)
}

// ==================== BusinessError 业务错误 ====================

// BusinessError 业务错误，携带错误码和描述
type BusinessError struct {
	Code    Code   // 错误码
	Message string // 错误描述（默认取 Code.Message()，可覆盖）
	Detail  string // 详细信息（可选，用于日志，不返回前端）
}

// New 创建业务错误
func New(code Code) *BusinessError {
	return &BusinessError{Code: code, Message: code.Message()}
}

// NewWithDetail 创建带详细信息的业务错误
func NewWithDetail(code Code, detail string) *BusinessError {
	return &BusinessError{Code: code, Message: code.Message(), Detail: detail}
}

// NewWithMessage 创建带自定义消息的业务错误
func NewWithMessage(code Code, msg string) *BusinessError {
	return &BusinessError{Code: code, Message: msg}
}

// Error 实现 error 接口
func (e *BusinessError) Error() string {
	if e.Detail != "" {
		return fmt.Sprintf("[%d] %s: %s", e.Code, e.Message, e.Detail)
	}
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

// GRPCStatus 转换为 gRPC status（用于 gRPC 传输）
func (e *BusinessError) GRPCStatus() *status.Status {
	return status.New(codes.Code(e.Code.Int()%1000+100), e.Error())
}

// Is 判断是否为指定错误码
func (e *BusinessError) Is(code Code) bool {
	return e.Code == code
}

// ==================== 便捷函数 ====================

// IsCode 判断 error 是否为指定错误码的 BusinessError
func IsCode(err error, code Code) bool {
	if be, ok := err.(*BusinessError); ok {
		return be.Code == code
	}
	return false
}

// FromError 从 error 提取 BusinessError，若不是则返回 nil
func FromError(err error) *BusinessError {
	if be, ok := err.(*BusinessError); ok {
		return be
	}
	return nil
}

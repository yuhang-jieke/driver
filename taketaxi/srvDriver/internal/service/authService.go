package service

import (
	"context"
	"fmt"
	"time"

	driver "driver/taketaxi/common/kitexGen"
	"driver/taketaxi/pkg/errcode"

	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

// ========== 账号安全 ==========

// ChangeMobile 修改手机号
func (s *DriverService) ChangeMobile(ctx context.Context, req *driver.ChangeMobileReq) (*driver.ChangeMobileResp, error) {
	if req.DriverId <= 0 {
		return nil, errcode.New(errcode.ErrInvalidDriverID)
	}
	if req.NewMobile == "" {
		return nil, errcode.NewWithDetail(errcode.ErrMissingField, "new_mobile")
	}
	if req.VerifyCode == "" {
		return nil, errcode.NewWithDetail(errcode.ErrMissingField, "verify_code")
	}

	// TODO: 验证短信验证码

	// 检查是否1个月内已修改
	key := fmt.Sprintf("driver:change_mobile:%d", req.DriverId)
	lastChange, err := s.rdb.Get(ctx, key).Result()
	if err != nil && err != redis.Nil {
		return nil, errcode.New(errcode.ErrRedisError)
	}
	if lastChange != "" {
		return nil, errcode.New(errcode.ErrMobileChangeLimit)
	}

	// 更新手机号
	if err := s.repo.UpdateMobile(ctx, req.DriverId, req.NewMobile); err != nil {
		return nil, errcode.NewWithDetail(errcode.ErrInternal, err.Error())
	}

	// 设置1个月过期标记
	s.rdb.Set(ctx, key, time.Now().Format("2006-01-02 15:04:05"), 30*24*time.Hour)

	return &driver.ChangeMobileResp{Success: true, Message: errcode.Success.Message()}, nil
}

// ChangePassword 修改密码
func (s *DriverService) ChangePassword(ctx context.Context, req *driver.ChangePasswordReq) (*driver.ChangePasswordResp, error) {
	if req.DriverId <= 0 {
		return nil, errcode.New(errcode.ErrInvalidDriverID)
	}
	if req.OldPassword == "" || req.NewPassword == "" {
		return nil, errcode.NewWithDetail(errcode.ErrMissingField, "password")
	}

	// 检查1小时内修改次数（限流防暴力破解）
	key := fmt.Sprintf("driver:change_pwd:%d", req.DriverId)
	count, err := s.rdb.Incr(ctx, key).Result()
	if err != nil {
		return nil, errcode.New(errcode.ErrRedisError)
	}
	if count == 1 {
		s.rdb.Expire(ctx, key, time.Hour)
	}
	if count > 5 {
		return nil, errcode.New(errcode.ErrMobileChangeLimit) // 复用频次限制码
	}

	// 获取司机信息
	drv, err := s.repo.GetDriverByID(ctx, req.DriverId)
	if err != nil {
		return nil, errcode.New(errcode.ErrDriverNotFound)
	}

	// 校验旧密码
	if drv.Password == "" {
		return nil, errcode.New(errcode.ErrPasswordNotSet)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(drv.Password), []byte(req.OldPassword)); err != nil {
		return nil, errcode.New(errcode.ErrOldPasswordWrong)
	}

	// 检查新旧密码是否相同
	if err := bcrypt.CompareHashAndPassword([]byte(drv.Password), []byte(req.NewPassword)); err == nil {
		return nil, errcode.New(errcode.ErrPasswordSameAsOld)
	}

	// 加密新密码
	hashedPwd, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, errcode.New(errcode.ErrPasswordHash)
	}

	// 更新密码
	if err := s.repo.UpdatePassword(ctx, req.DriverId, string(hashedPwd)); err != nil {
		return nil, errcode.NewWithDetail(errcode.ErrInternal, err.Error())
	}

	return &driver.ChangePasswordResp{Success: true, Message: errcode.Success.Message()}, nil
}

// ResetPassword 重置密码（忘记密码）
func (s *DriverService) ResetPassword(ctx context.Context, req *driver.ResetPasswordReq) (*driver.ResetPasswordResp, error) {
	if req.Mobile == "" {
		return nil, errcode.NewWithDetail(errcode.ErrMissingField, "mobile")
	}
	if req.VerifyCode == "" {
		return nil, errcode.NewWithDetail(errcode.ErrMissingField, "verify_code")
	}
	if req.NewPassword == "" {
		return nil, errcode.NewWithDetail(errcode.ErrMissingField, "new_password")
	}

	// TODO: 验证短信验证码

	// 通过手机号查找司机
	drv, err := s.repo.GetDriverByMobile(ctx, req.Mobile)
	if err != nil {
		return nil, errcode.New(errcode.ErrMobileNotRegistered)
	}

	// 加密新密码
	hashedPwd, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, errcode.New(errcode.ErrPasswordHash)
	}

	// 更新密码
	if err := s.repo.UpdatePassword(ctx, drv.DriverId, string(hashedPwd)); err != nil {
		return nil, errcode.NewWithDetail(errcode.ErrInternal, err.Error())
	}

	return &driver.ResetPasswordResp{Success: true, Message: errcode.Success.Message()}, nil
}

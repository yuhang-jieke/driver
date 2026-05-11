package service

import (
	"context"
	"time"

	driver "driver/taketaxi/common/kitexGen"
	"driver/taketaxi/srvDriver/internal/model"
	"driver/taketaxi/pkg/errcode"
)

// ========== 认证资料 ==========

// UpdateRealname 更新/提交实名认证
func (s *DriverService) UpdateRealname(ctx context.Context, req *driver.UpdateRealnameReq) (*driver.UpdateRealnameResp, error) {
	if req.DriverId <= 0 {
		return nil, errcode.New(errcode.ErrInvalidDriverID)
	}

	rn := &model.DriverRealname{
		DriverId:       req.DriverId,
		RealName:       req.RealName,
		IdCardNo:       req.IdCardNo,
		IdCardFrontUrl: req.IdCardFrontUrl,
		IdCardBackUrl:  req.IdCardBackUrl,
		Address:        req.Address,
		Nation:         req.Nation,
	}
	if req.Gender > 0 {
		rn.Gender = int8(req.Gender)
	}
	if req.Birthday != "" {
		if t, err := time.Parse("2006-01-02", req.Birthday); err == nil {
			rn.Birthday = t
		}
	}
	if req.ExpireDate != "" {
		if t, err := time.Parse("2006-01-02", req.ExpireDate); err == nil {
			rn.ExpireDate = t
		}
	}
	// 提交后状态变为审核中
	rn.Status = model.VerifyStatusPending

	if err := s.repo.UpdateRealname(ctx, rn); err != nil {
		return nil, errcode.NewWithDetail(errcode.ErrInternal, err.Error())
	}
	return &driver.UpdateRealnameResp{Success: true, Message: errcode.Success.Message()}, nil
}

// UpdateLicense 更新/提交驾驶证认证
func (s *DriverService) UpdateLicense(ctx context.Context, req *driver.UpdateLicenseReq) (*driver.UpdateLicenseResp, error) {
	if req.DriverId <= 0 {
		return nil, errcode.New(errcode.ErrInvalidDriverID)
	}

	lic := &model.DriverLicense{
		DriverId:    req.DriverId,
		LicenseNo:   req.LicenseNo,
		LicenseType: req.LicenseType,
		LicenseUrl:  req.LicenseUrl,
	}
	if req.FirstIssueDate != "" {
		if t, err := time.Parse("2006-01-02", req.FirstIssueDate); err == nil {
			lic.FirstIssueDate = t
		}
	}
	if req.IssueDate != "" {
		if t, err := time.Parse("2006-01-02", req.IssueDate); err == nil {
			lic.IssueDate = t
		}
	}
	if req.ExpireDate != "" {
		if t, err := time.Parse("2006-01-02", req.ExpireDate); err == nil {
			lic.ExpireDate = t
		}
	}
	// 提交后状态变为审核中
	lic.Status = model.VerifyStatusPending

	if err := s.repo.UpdateLicense(ctx, lic); err != nil {
		return nil, errcode.NewWithDetail(errcode.ErrInternal, err.Error())
	}
	return &driver.UpdateLicenseResp{Success: true, Message: errcode.Success.Message()}, nil
}

// UpdateVehicle 更新/提交车辆信息
func (s *DriverService) UpdateVehicle(ctx context.Context, req *driver.UpdateVehicleReq) (*driver.UpdateVehicleResp, error) {
	if req.DriverId <= 0 {
		return nil, errcode.New(errcode.ErrInvalidDriverID)
	}

	v := &model.DriverVehicle{
		DriverId:          req.DriverId,
		PlateNo:           req.PlateNo,
		VehicleModel:      req.VehicleModel,
		VehicleBrand:      req.VehicleBrand,
		VehicleColor:      req.VehicleColor,
		VehicleColorCode:  req.VehicleColorCode,
		DrivingLicenseUrl: req.DrivingLicenseUrl,
		VehiclePhotoUrl:   req.VehiclePhotoUrl,
		Status:            model.VerifyStatusPending, // 提交后状态变为审核中
	}
	if req.SeatCount > 0 {
		v.SeatCount = int8(req.SeatCount)
	}

	if err := s.repo.UpdateVehicle(ctx, v); err != nil {
		return nil, errcode.NewWithDetail(errcode.ErrInternal, err.Error())
	}
	return &driver.UpdateVehicleResp{Success: true, Message: errcode.Success.Message()}, nil
}

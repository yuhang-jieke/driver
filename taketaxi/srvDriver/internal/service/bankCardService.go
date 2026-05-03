package service

import (
	"context"
	"fmt"
	"regexp"
	"time"

	driver "driver/taketaxi/common/kitexGen"
	"driver/taketaxi/srvDriver/internal/model"
	"driver/taketaxi/pkg/errcode"
)

// ========== 银行卡 ==========

// BindBankCard 绑定银行卡
// 业务规则：
//   - 仅允许借记卡（信用卡不能接收转账）
//   - 持卡人姓名必须与实名认证姓名一致
func (s *DriverService) BindBankCard(ctx context.Context, req *driver.BindBankCardReq) (*driver.BindBankCardResp, error) {
	if req.DriverId <= 0 {
		return nil, errcode.New(errcode.ErrInvalidDriverID)
	}
	if req.BankName == "" || req.BankCardNo == "" || req.AccountName == "" {
		return nil, errcode.NewWithDetail(errcode.ErrMissingField, "bank_name/bank_card_no/account_name")
	}
	if len(req.BankCardNo) > model.BankCardNoMaxLength {
		return nil, errcode.New(errcode.ErrBankCardNoTooLong)
	}

	// 验证持卡人姓名格式（2-6个汉字）
	if !isValidChineseName(req.AccountName) {
		return nil, errcode.New(errcode.ErrInvalidAccountName)
	}

	// 拒绝信用卡（信用卡不能接收转账，打款必失败）
	if int8(req.CardType) == model.BankCardTypeCredit {
		return nil, errcode.New(errcode.ErrWithdrawCreditCard)
	}

	// 检查是否已有绑定
	existing, err := s.repo.GetBankCard(ctx, req.DriverId)
	if err != nil {
		return nil, errcode.NewWithDetail(errcode.ErrInternal, err.Error())
	}
	if existing != nil {
		return nil, errcode.New(errcode.ErrBankCardAlreadyBound)
	}

	// 持卡人姓名与实名认证姓名一致性校验
	realname, _ := s.repo.GetDriverRealname(ctx, req.DriverId)
	if realname == nil || realname.Status != model.VerifyStatusApproved {
		return nil, errcode.New(errcode.ErrWithdrawRealnameNeeded)
	}
	if req.AccountName != realname.RealName {
		return nil, errcode.New(errcode.ErrWithdrawCardNameMismatch)
	}

	// 脱敏卡号: 保留后4位
	maskedNo := maskBankCard(req.BankCardNo)

	card := &model.DriverBankCard{
		DriverId:          req.DriverId,
		BankName:          req.BankName,
		BankCode:          getBankCode(req.BankName),
		BankCardNo:        maskedNo,
		BankCardNoEncrypt: req.BankCardNo, // TODO: AES加密存储
		AccountName:       req.AccountName,
		CardType:          model.BankCardTypeDebit,
		Status:            model.BankCardStatusActive,
		LastModifiedAt:    time.Now(),
	}

	if err := s.repo.CreateBankCard(ctx, card); err != nil {
		return nil, errcode.NewWithDetail(errcode.ErrCreateRecord, err.Error())
	}
	return &driver.BindBankCardResp{Success: true, Message: errcode.Success.Message()}, nil
}

// GetBankCard 查询银行卡信息
func (s *DriverService) GetBankCard(ctx context.Context, req *driver.GetBankCardReq) (*driver.GetBankCardResp, error) {
	if req.DriverId <= 0 {
		return nil, errcode.New(errcode.ErrInvalidDriverID)
	}

	card, err := s.repo.GetBankCard(ctx, req.DriverId)
	if err != nil {
		return nil, err
	}

	resp := &driver.GetBankCardResp{HasCard: card != nil}
	if card != nil {
		resp.BankName = card.BankName
		resp.BankCode = card.BankCode
		resp.BankCardNo = card.BankCardNo // 已脱敏
		resp.AccountName = card.AccountName
		resp.CardType = int32(card.CardType)
		resp.BranchName = card.BranchName
		resp.LastModifiedAt = card.LastModifiedAt.Format("2006-01-02 15:04:05")
	}
	return resp, nil
}

// UpdateBankCard 更换银行卡（每月最多1次）
// 业务规则：
//   - 仅允许借记卡
//   - 持卡人姓名必须与实名认证姓名一致
func (s *DriverService) UpdateBankCard(ctx context.Context, req *driver.UpdateBankCardReq) (*driver.UpdateBankCardResp, error) {
	if req.DriverId <= 0 {
		return nil, errcode.New(errcode.ErrInvalidDriverID)
	}
	if req.BankName == "" || req.BankCardNo == "" || req.AccountName == "" {
		return nil, errcode.NewWithDetail(errcode.ErrMissingField, "bank_name/bank_card_no/account_name")
	}
	if len(req.BankCardNo) > model.BankCardNoMaxLength {
		return nil, errcode.New(errcode.ErrBankCardNoTooLong)
	}

	// 验证持卡人姓名格式（2-6个汉字）
	if !isValidChineseName(req.AccountName) {
		return nil, errcode.New(errcode.ErrInvalidAccountName)
	}

	// 拒绝信用卡
	if int8(req.CardType) == model.BankCardTypeCredit {
		return nil, errcode.New(errcode.ErrWithdrawCreditCard)
	}

	existing, err := s.repo.GetBankCard(ctx, req.DriverId)
	if err != nil {
		return nil, errcode.NewWithDetail(errcode.ErrInternal, err.Error())
	}
	if existing == nil {
		return nil, errcode.New(errcode.ErrBankCardNotBound)
	}

	// 持卡人姓名与实名认证姓名一致性校验
	realname, _ := s.repo.GetDriverRealname(ctx, req.DriverId)
	if realname == nil || realname.Status != model.VerifyStatusApproved {
		return nil, errcode.New(errcode.ErrWithdrawRealnameNeeded)
	}
	if req.AccountName != realname.RealName {
		return nil, errcode.New(errcode.ErrWithdrawCardNameMismatch)
	}

	// 检查月频次限制
	if !existing.LastModifiedAt.IsZero() {
		now := time.Now()
		if existing.LastModifiedAt.Year() == now.Year() && existing.LastModifiedAt.Month() == now.Month() {
			nextMonth := time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, time.Local)
			return nil, errcode.NewWithDetail(errcode.ErrBankCardChangeLimit,
				fmt.Sprintf("next_available=%s", nextMonth.Format("2006-01-02")))
		}
	}

	maskedNo := maskBankCard(req.BankCardNo)
	updates := map[string]interface{}{
		"bank_name":            req.BankName,
		"bank_code":            getBankCode(req.BankName),
		"bank_card_no":         maskedNo,
		"bank_card_no_encrypt": req.BankCardNo, // TODO: AES加密
		"account_name":         req.AccountName,
		"card_type":            model.BankCardTypeDebit, // 强制借记卡
		"branch_name":          req.BranchName,
		"last_modified_at":     time.Now(),
	}

	if err := s.repo.UpdateBankCard(ctx, req.DriverId, updates); err != nil {
		return nil, errcode.NewWithDetail(errcode.ErrInternal, err.Error())
	}
	return &driver.UpdateBankCardResp{Success: true, Message: errcode.Success.Message()}, nil
}

// isValidChineseName 验证姓名是否为2-6个汉字
func isValidChineseName(name string) bool {
	matched, _ := regexp.MatchString(`^[\x{4e00}-\x{9fa5}]{2,6}$`, name)
	return matched
}

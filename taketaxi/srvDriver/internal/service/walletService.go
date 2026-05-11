package service

import (
	"context"
	"fmt"
	"time"

	driver "driver/taketaxi/common/kitexGen"
	"driver/taketaxi/srvDriver/internal/model"
	"driver/taketaxi/srvDriver/internal/repository"
	"driver/taketaxi/pkg/errcode"
)

// ========== 钱包 & 提现 ==========

// withdrawEligibility 提现资格评估结果
type withdrawEligibility struct {
	CanWithdraw       bool   // 是否可提现
	DisableReasonCode string // 禁用原因码（供前端做差异化 UI）
	DisableReasonText string // 禁用原因描述（兜底展示）
}

// evaluateWithdrawEligibility 提现资格统一校验
// 提现页渲染和提现提交共用同一条校验路径，保证页态与提交态一致
//
// 校验优先级：账号状态 → 银行卡 → 信用卡拒绝 → 实名认证 → 姓名一致性 → 每日次数 → 可提现余额
//
// 金额语义：wallet.Balance = 总余额，wallet.FrozenAmount = 冻结子集，
// 可提现金额 = Balance - FrozenAmount
func (s *DriverService) evaluateWithdrawEligibility(
	ctx context.Context,
	driverID int64,
	wallet *model.DriverWallet,
) withdrawEligibility {
	// 0. 司机账号状态校验（冻结/注销账号不允许提现）
	driverInfo, _ := s.repo.GetDriverByID(ctx, driverID)
	if driverInfo != nil {
		if driverInfo.Status == model.DriverStatusFrozen {
			return withdrawEligibility{
				CanWithdraw:       false,
				DisableReasonCode: model.WithdrawDisableAccountFrozen,
				DisableReasonText: "账号已冻结，无法提现",
			}
		}
		if driverInfo.Status == model.DriverStatusClosed {
			return withdrawEligibility{
				CanWithdraw:       false,
				DisableReasonCode: model.WithdrawDisableAccountClosed,
				DisableReasonText: "账号已注销，无法提现",
			}
		}
	}

	// 1. 银行卡校验
	card, _ := s.repo.GetBankCard(ctx, driverID)
	if card == nil {
		return withdrawEligibility{
			CanWithdraw:       false,
			DisableReasonCode: model.WithdrawDisableNoCard,
			DisableReasonText: "请先绑定银行卡",
		}
	}

	// 2. 仅允许借记卡（信用卡不能接收转账，打款必失败）
	if card.CardType == model.BankCardTypeCredit {
		return withdrawEligibility{
			CanWithdraw:       false,
			DisableReasonCode: model.WithdrawDisableCreditCard,
			DisableReasonText: "仅支持提现至借记卡",
		}
	}

	// 3. 认证状态校验（需实名认证通过才能提现）
	realname, _ := s.repo.GetDriverRealname(ctx, driverID)
	if realname == nil || realname.Status != model.VerifyStatusApproved {
		return withdrawEligibility{
			CanWithdraw:       false,
			DisableReasonCode: model.WithdrawDisableVerify,
			DisableReasonText: "需完成实名认证",
		}
	}

	// 4. 银行卡持卡人姓名与实名姓名一致性校验
	if card.AccountName != realname.RealName {
		return withdrawEligibility{
			CanWithdraw:       false,
			DisableReasonCode: model.WithdrawDisableCardNameMismatch,
			DisableReasonText: "银行卡持卡人须与实名认证姓名一致",
		}
	}

	// 5. 今日提现次数校验（只统计处理中+成功的记录，失败不计入）
	todayCount, _ := s.repo.GetTodayWithdrawCount(ctx, driverID)
	if todayCount >= int64(model.MaxWithdrawPerDay) {
		return withdrawEligibility{
			CanWithdraw:       false,
			DisableReasonCode: model.WithdrawDisableCount,
			DisableReasonText: fmt.Sprintf("今日提现次数已达上限(%d次)", model.MaxWithdrawPerDay),
		}
	}

	// 6. 可提现余额校验（Balance=总余额，FrozenAmount=冻结子集）
	available := wallet.Balance - wallet.FrozenAmount
	if available <= 0 {
		return withdrawEligibility{
			CanWithdraw:       false,
			DisableReasonCode: model.WithdrawDisableZero,
			DisableReasonText: "可提现余额为零",
		}
	}

	return withdrawEligibility{CanWithdraw: true}
}

// GetWithdrawPage 查询提现页信息（规则 + 资格 + 银行卡摘要）
// 所有金额字段单位为分(int64)
func (s *DriverService) GetWithdrawPage(ctx context.Context, req *driver.GetWithdrawPageReq) (*driver.GetWithdrawPageResp, error) {
	if req.DriverId <= 0 {
		return nil, errcode.New(errcode.ErrInvalidDriverID)
	}

	// 1. 查询钱包
	wallet, err := s.repo.GetWallet(ctx, req.DriverId)
	if err != nil {
		return nil, errcode.NewWithDetail(errcode.ErrWalletNotFound, err.Error())
	}

	// 2. 评估提现资格（核心：统一校验路径）
	eligibility := s.evaluateWithdrawEligibility(ctx, req.DriverId, wallet)

	// 3. 查询银行卡摘要
	card, _ := s.repo.GetBankCard(ctx, req.DriverId)
	bankCardInfo := &driver.WithdrawPageBankCard{
		HasBankCard: card != nil,
	}
	if card != nil {
		bankCardInfo.BankName = card.BankName
		bankCardInfo.BankCardNo = card.BankCardNo // 已脱敏
	}

	// 4. 今日提现次数（只统计处理中+成功的记录）
	todayCount, _ := s.repo.GetTodayWithdrawCount(ctx, req.DriverId)

	// 5. 计算可提现金额（Balance=总余额，FrozenAmount=冻结子集）
	available := wallet.Balance - wallet.FrozenAmount
	if available < 0 {
		available = 0
	}

	// 6. 推荐提现金额（单位：分）
	suggestedAmounts := s.computeSuggestedAmounts(available)

	// 7. 组装规则信息（金额单位：分）
	ruleInfo := &driver.WithdrawRuleInfo{
		MinWithdrawAmount:    model.WithdrawMinAmount,
		MaxWithdrawAmount:    model.MaxWithdrawPerOrder,
		TodayWithdrawCount:   int32(todayCount),
		TodayWithdrawLimit:   int32(model.MaxWithdrawPerDay),
		EstimatedArrivalText: model.WithdrawArrivalText,
		FeeFree:              true,
		FeeAmount:            0,
		FeeDesc:              "免手续费",
	}

	// 8. 组装操作状态
	actionState := &driver.WithdrawPageActionState{
		CanWithdraw:       eligibility.CanWithdraw,
		DisableReasonCode: eligibility.DisableReasonCode,
		DisableReasonText: eligibility.DisableReasonText,
	}

	return &driver.GetWithdrawPageResp{
		WalletBalance:           wallet.Balance,
		FrozenAmount:            wallet.FrozenAmount,
		AvailableWithdrawAmount: available,
		RuleInfo:                ruleInfo,
		BankCard:                bankCardInfo,
		ActionState:             actionState,
		SuggestedAmounts:        suggestedAmounts,
		WithdrawNotice:          model.WithdrawNotice,
	}, nil
}

// computeSuggestedAmounts 根据可用余额计算推荐提现金额（单位：分）
func (s *DriverService) computeSuggestedAmounts(available int64) []int64 {
	if available <= 0 {
		return nil
	}
	// 预设金额：50元, 100元, 200元, 500元, 1000元, 2000元
	presets := []int64{5000, 10000, 20000, 50000, 100000, 200000}
	var result []int64
	for _, p := range presets {
		if p <= available {
			result = append(result, p)
		}
	}
	// 如果可用余额大于最大预设值，追加"全部提现"
	if available > 200000 {
		result = append(result, available)
	}
	return result
}

// ApplyWithdraw 申请提现（事务 + 乐观锁 + 幂等）
// 金额单位：分(int64)
func (s *DriverService) ApplyWithdraw(ctx context.Context, req *driver.ApplyWithdrawReq) (*driver.ApplyWithdrawResp, error) {
	if req.DriverId <= 0 {
		return nil, errcode.New(errcode.ErrInvalidDriverID)
	}
	if req.Amount <= 0 {
		return nil, errcode.NewWithDetail(errcode.ErrInvalidParam, "amount must be positive")
	}

	// 1. 最低金额校验
	if req.Amount < model.WithdrawMinAmount {
		return nil, errcode.New(errcode.ErrWithdrawMinAmount)
	}
	if req.Amount > model.MaxWithdrawPerOrder {
		return nil, errcode.New(errcode.ErrWithdrawAmountLimit)
	}

	// 2. 幂等校验：30秒内相同司机相同金额的请求视为重复
	if existing, _ := s.repo.FindRecentPendingWithdraw(ctx, req.DriverId, req.Amount, 30*time.Second); existing != nil {
		return &driver.ApplyWithdrawResp{
			Success:    true,
			Message:    model.WithdrawArrivalText,
			WithdrawNo: existing.WithdrawNo,
		}, nil
	}

	// 3. 查询钱包
	wallet, err := s.repo.GetWallet(ctx, req.DriverId)
	if err != nil {
		return nil, errcode.New(errcode.ErrWalletNotFound)
	}

	// 4. 统一资格校验（与提现页共用）
	eligibility := s.evaluateWithdrawEligibility(ctx, req.DriverId, wallet)
	if !eligibility.CanWithdraw {
		return nil, errcode.NewWithMessage(errcode.ErrWithdrawPageUnavailable, eligibility.DisableReasonText)
	}

	// 5. 可提现余额校验（Balance=总余额，FrozenAmount=冻结子集）
	available := wallet.Balance - wallet.FrozenAmount
	if available < req.Amount {
		if wallet.Balance < req.Amount {
			return nil, errcode.NewWithDetail(errcode.ErrInsufficientBalance,
				fmt.Sprintf("balance=%d,available=%d", wallet.Balance, available))
		}
		return nil, errcode.NewWithDetail(errcode.ErrFrozenAmountLimit,
			fmt.Sprintf("frozen=%d,available=%d", wallet.FrozenAmount, available))
	}

	// 6. 查询银行卡（资格校验已保证 card != nil 且为借记卡）
	card, _ := s.repo.GetBankCard(ctx, req.DriverId)

	// 7. 生成提现单号
	withdrawNo := fmt.Sprintf("WD%s%d", time.Now().Format("20060102150405"), req.DriverId%10000)

	// 8. 在事务中执行：创建提现记录 + 扣减余额（乐观锁）+ 记录流水
	err = s.repo.RunInTx(ctx, func(txRepo *repository.DriverRepo) error {
		// 8a. 事务内重新读取钱包（防止并发冲突）
		txWallet, err := txRepo.GetWallet(ctx, req.DriverId)
		if err != nil {
			return err
		}

		// 8b. 事务内再次校验可提现余额
		txAvailable := txWallet.Balance - txWallet.FrozenAmount
		if txAvailable < req.Amount {
			return errcode.New(errcode.ErrInsufficientBalance)
		}

		// 8c. 创建提现记录
		record := &model.DriverWithdrawRecord{
			WithdrawNo:        withdrawNo,
			DriverId:          req.DriverId,
			Amount:            req.Amount,
			Fee:               0,
			ActualAmount:      req.Amount,
			BankName:          card.BankName,
			BankCode:          card.BankCode,
			BankCardNo:        card.BankCardNo,
			BankCardNoEncrypt: card.BankCardNoEncrypt,
			AccountName:       card.AccountName,
			Status:            model.WithdrawStatusPending,
			ApplyTime:         time.Now(),
			Channel:           model.WithdrawChannelDefault,
		}
		if err := txRepo.CreateWithdrawRecord(ctx, record); err != nil {
			return errcode.NewWithDetail(errcode.ErrCreateRecord, err.Error())
		}

		// 8d. 乐观锁扣减余额
		txWallet.Balance -= req.Amount
		txWallet.TotalWithdraw += req.Amount
		if err := txRepo.UpdateWallet(ctx, txWallet); err != nil {
			return errcode.New(errcode.ErrUpdateBalance)
		}

		// 8e. 记录流水
		_ = txRepo.CreateWalletTransactionLog(ctx, &model.WalletTransactionLog{
			DriverId:        req.DriverId,
			TransactionNo:   withdrawNo,
			TransactionType: model.WalletTxTypeWithdraw,
			Amount:          req.Amount,
			BalanceBefore:   txWallet.Balance + req.Amount,
			BalanceAfter:    txWallet.Balance,
			FrozenBefore:    txWallet.FrozenAmount,
			FrozenAfter:     txWallet.FrozenAmount,
			RelatedId:       record.Id,
			RelatedType:     "withdraw",
			Status:          model.WalletTxStatusSuccess,
			Remark:          "withdraw to " + card.BankCode,
		})

		return nil
	})

	if err != nil {
		// 如果是 BusinessError 直接返回
		if _, ok := err.(*errcode.BusinessError); ok {
			return nil, err
		}
		return nil, errcode.NewWithDetail(errcode.ErrInternal, err.Error())
	}

	return &driver.ApplyWithdrawResp{
		Success:    true,
		Message:    model.WithdrawArrivalText,
		WithdrawNo: withdrawNo,
	}, nil
}

// GetWallet 查询钱包概览
// 金额单位：分(int64)
// 支持两种查询模式：
//   1. 固定周期模式（默认）：返回今日/本周/本月收入
//   2. 自定义范围模式：传入 start_date/end_date，返回指定范围内收入
func (s *DriverService) GetWallet(ctx context.Context, req *driver.GetWalletReq) (*driver.GetWalletResp, error) {
	if req.DriverId <= 0 {
		return nil, errcode.New(errcode.ErrInvalidDriverID)
	}

	wallet, err := s.repo.GetWallet(ctx, req.DriverId)
	if err != nil {
		return nil, err
	}

	// 查询今日提现次数（只统计处理中+成功的记录）
	todayCount, _ := s.repo.GetTodayWithdrawCount(ctx, req.DriverId)

	// 查询银行卡
	card, _ := s.repo.GetBankCard(ctx, req.DriverId)
	bankCardNo := ""
	if card != nil {
		bankCardNo = card.BankCardNo
	}

	resp := &driver.GetWalletResp{
		Balance:            wallet.Balance,
		FrozenAmount:       wallet.FrozenAmount,
		TotalIncome:        wallet.TotalIncome,
		TotalWithdraw:      wallet.TotalWithdraw,
		TodayWithdrawCount: int32(todayCount),
		BankCardNo:         bankCardNo,
		HasBankCard:        card != nil,
	}

	// 判断查询模式
	if req.StartDate != "" && req.EndDate != "" {
		// 自定义范围模式
		return s.getWalletWithDateRange(ctx, req, resp)
	}

	// 固定周期模式（原有逻辑）
	return s.getWalletWithFixedPeriod(ctx, req, resp)
}

// getWalletWithDateRange 自定义日期范围查询模式
func (s *DriverService) getWalletWithDateRange(ctx context.Context, req *driver.GetWalletReq, resp *driver.GetWalletResp) (*driver.GetWalletResp, error) {
	// 1. 解析日期参数
	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return nil, errcode.NewWithDetail(errcode.ErrInvalidParam, "invalid start_date format, expected YYYY-MM-DD")
	}
	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		return nil, errcode.NewWithDetail(errcode.ErrInvalidParam, "invalid end_date format, expected YYYY-MM-DD")
	}

	// 2. 日期范围校验
	if endDate.Before(startDate) {
		return nil, errcode.NewWithDetail(errcode.ErrInvalidParam, "end_date must be after or equal to start_date")
	}

	// 3. 最多查询 90 天
	daysDiff := int(endDate.Sub(startDate).Hours()/24) + 1 // +1 包含结束日期当天
	if daysDiff > 90 {
		return nil, errcode.NewWithDetail(errcode.ErrInvalidParam, "date range cannot exceed 90 days")
	}

	// 4. 补充时间边界：startDate 00:00:00, endDate 23:59:59
	startDateTime := time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, time.Local)
	endDateTime := time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 23, 59, 59, 999999999, time.Local)

	// 5. 查询指定范围内的统计数据
	stats, _ := s.repo.GetOrderStats(ctx, req.DriverId, startDateTime, endDateTime)

	// 6. 填充响应（复用 today_income 字段为"范围内收入"）
	resp.TodayIncome = int64(stats.TotalIncome)
	resp.WeekIncome = 0  // 自定义范围模式下，周/月收入字段无意义
	resp.MonthIncome = 0
	resp.QueryStartDate = req.StartDate
	resp.QueryEndDate = req.EndDate

	return resp, nil
}

// getWalletWithFixedPeriod 固定周期查询模式（原有逻辑）
func (s *DriverService) getWalletWithFixedPeriod(ctx context.Context, req *driver.GetWalletReq, resp *driver.GetWalletResp) (*driver.GetWalletResp, error) {
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
	weekday := int(now.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	weekStart := todayStart.AddDate(0, 0, -weekday+1) // 本周一（修复周日=0的Bug）
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.Local)

	todayStats, _ := s.repo.GetOrderStats(ctx, req.DriverId, todayStart, now)
	weekStats, _ := s.repo.GetOrderStats(ctx, req.DriverId, weekStart, now)
	monthStats, _ := s.repo.GetOrderStats(ctx, req.DriverId, monthStart, now)

	resp.TodayIncome = int64(todayStats.TotalIncome)
	resp.WeekIncome = int64(weekStats.TotalIncome)
	resp.MonthIncome = int64(monthStats.TotalIncome)
	resp.QueryStartDate = todayStart.Format("2006-01-02")
	resp.QueryEndDate = now.Format("2006-01-02")

	return resp, nil
}

// GetWithdrawRecords 查询提现记录
// 金额单位：分(int64)
func (s *DriverService) GetWithdrawRecords(ctx context.Context, req *driver.GetWithdrawRecordsReq) (*driver.GetWithdrawRecordsResp, error) {
	if req.DriverId <= 0 {
		return nil, errcode.New(errcode.ErrInvalidDriverID)
	}

	page := req.Page
	if page <= 0 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}

	records, total, err := s.repo.GetWithdrawRecords(ctx, req.DriverId, page, pageSize)
	if err != nil {
		return nil, err
	}

	var items []*driver.WithdrawRecordItem
	for _, r := range records {
		finishTime := ""
		if r.FinishTime != nil {
			finishTime = r.FinishTime.Format("2006-01-02 15:04:05")
		}
		items = append(items, &driver.WithdrawRecordItem{
			Id:           r.Id,
			WithdrawNo:   r.WithdrawNo,
			Amount:       r.Amount,
			Fee:          r.Fee,
			ActualAmount: r.ActualAmount,
			BankName:     r.BankName,
			BankCardNo:   r.BankCardNo,
			Status:       int32(r.Status),
			FailReason:   r.FailReason,
			ApplyTime:    r.ApplyTime.Format("2006-01-02 15:04:05"),
			FinishTime:   finishTime,
		})
	}

	return &driver.GetWithdrawRecordsResp{
		Records: items,
		Total:   int32(total),
	}, nil
}

// GetIncomeDetail 查询收入明细
func (s *DriverService) GetIncomeDetail(ctx context.Context, req *driver.GetIncomeDetailReq) (*driver.GetIncomeDetailResp, error) {
	if req.DriverId <= 0 {
		return nil, errcode.New(errcode.ErrInvalidDriverID)
	}

	period := req.Period
	if period == "" {
		period = "today"
	}
	if period != "today" && period != "week" && period != "month" {
		period = "today"
	}

	results, err := s.repo.GetIncomeDetail(ctx, req.DriverId, period)
	if err != nil {
		return nil, err
	}

	var totalAmount int64
	var items []*driver.IncomeDetailItem
	for _, r := range results {
		totalAmount += int64(r.Amount)
		items = append(items, &driver.IncomeDetailItem{
			TypeName: r.TypeName,
			TypeCode: int32(r.TypeCode),
			Amount:   int64(r.Amount),
			Count:    int32(r.Count),
		})
	}

	return &driver.GetIncomeDetailResp{
		Items:       items,
		TotalAmount: totalAmount,
	}, nil
}

// HandleWithdrawCallback 提现回调处理（支付渠道通知提现成功/失败）
// 业务规则（P1-提现失败退回机制）：
//   - 提现成功(STATUS=2)：更新提现记录状态为成功，记录完成时间
//   - 提现失败(STATUS=3)：退回余额（Balance += Amount），更新提现记录状态为失败
//   - 仅处理当前状态为"处理中(1)"的记录，避免重复回调
//
// 金额语义：Balance = 总余额（含冻结），提现失败退回时只加 Balance，不动 FrozenAmount
func (s *DriverService) HandleWithdrawCallback(ctx context.Context, withdrawNo string, success bool, failReason string) error {
	// 1. 查询提现记录
	record, err := s.repo.GetWithdrawRecordByNo(ctx, withdrawNo)
	if err != nil {
		return errcode.NewWithDetail(errcode.ErrInternal, fmt.Sprintf("withdraw record not found: %s", withdrawNo))
	}

	// 2. 仅处理处理中的记录（幂等：重复回调不重复退回）
	if record.Status != model.WithdrawStatusPending {
		return nil // 已处理过，忽略
	}

	now := time.Now()

	if success {
		// 3a. 提现成功：更新状态 + 记录完成时间
		updates := map[string]interface{}{
			"status":      model.WithdrawStatusSuccess,
			"finish_time": now,
		}
		return s.repo.UpdateWithdrawStatusByNo(ctx, withdrawNo, updates)
	}

	// 3b. 提现失败：退回余额 + 更新提现记录状态
	return s.repo.RunInTx(ctx, func(txRepo *repository.DriverRepo) error {
		// 退回余额
		wallet, err := txRepo.GetWallet(ctx, record.DriverId)
		if err != nil {
			return err
		}
		wallet.Balance += record.Amount
		wallet.TotalWithdraw -= record.Amount
		if err := txRepo.UpdateWallet(ctx, wallet); err != nil {
			return err
		}

		// 更新提现记录
		updates := map[string]interface{}{
			"status":      model.WithdrawStatusFailed,
			"fail_reason": failReason,
			"finish_time": now,
		}
		if err := txRepo.UpdateWithdrawStatusByNo(ctx, withdrawNo, updates); err != nil {
			return err
		}

		// 记录退回流水
		_ = txRepo.CreateWalletTransactionLog(ctx, &model.WalletTransactionLog{
			DriverId:        record.DriverId,
			TransactionNo:   fmt.Sprintf("RF%s", withdrawNo[2:]),
			TransactionType: model.WalletTxTypeRefund,
			Amount:          record.Amount,
			BalanceBefore:   wallet.Balance - record.Amount,
			BalanceAfter:    wallet.Balance,
			FrozenBefore:    wallet.FrozenAmount,
			FrozenAfter:     wallet.FrozenAmount,
			RelatedId:       record.Id,
			RelatedType:     "withdraw_refund",
			Status:          model.WalletTxStatusSuccess,
			Remark:          "withdraw failed refund: " + failReason,
		})

		return nil
	})
}

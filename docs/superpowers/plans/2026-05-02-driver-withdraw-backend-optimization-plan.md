# 提现后端代码优化实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 优化提现后端代码：金额精度迁移(float64→int64分)、修正校验顺序、实现换卡滚动周期限制、实现提现失败回退机制。

**Architecture:** 分三阶段实施：Phase 1 数据模型层 → Phase 2 业务逻辑层 → Phase 3 接口层。每阶段完成后可独立测试验证。

**Tech Stack:** Go 1.26, gRPC/protobuf, Gin, GORM

---

## Phase 1: 数据模型层

### Task 1: 更新 Proto 金额字段类型

**Files:**
- Modify: `taketaxi/common/idl/driver.proto`

- [ ] **Step 1: 更新 GetWalletResp 金额字段**

找到 `GetWalletResp` message，将所有金额字段从 `double` 改为 `int64`，添加注释标注单位为分：

```proto
message GetWalletResp {
  int64 balance = 1;            // 可提现余额(分)
  int64 frozen_amount = 2;      // 冻结金额(分)
  int64 today_income = 3;       // 今日收入(分)
  int64 week_income = 4;        // 本周收入(分)
  int64 month_income = 5;       // 本月收入(分)
  int64 total_income = 6;       // 累计总收入(分)
  int64 total_withdraw = 7;     // 累计提现金额(分)
  int32 today_withdraw_count = 8;
  string bank_card_no = 9;
  bool has_bank_card = 10;
}
```

- [ ] **Step 2: 更新 ApplyWithdrawReq 金额字段**

```proto
message ApplyWithdrawReq {
  int64 driver_id = 1;
  int64 amount = 2;             // 提现金额(分)
}
```

- [ ] **Step 3: 更新 ApplyWithdrawResp 添加金额字段**

```proto
message ApplyWithdrawResp {
  bool success = 1;
  string message = 2;
  string withdraw_no = 3;
}
```

- [ ] **Step 4: 更新 WithdrawRecordItem 金额字段**

```proto
message WithdrawRecordItem {
  int64 id = 1;
  string withdraw_no = 2;
  int64 amount = 3;             // 提现金额(分)
  int64 fee = 4;                // 手续费(分)
  int64 actual_amount = 5;      // 实际到账(分)
  string bank_name = 6;
  string bank_card_no = 7;
  int32 status = 8;
  string fail_reason = 9;
  string apply_time = 10;
  string finish_time = 11;
}
```

- [ ] **Step 5: 更新 GetWithdrawPageResp 金额字段**

```proto
message WithdrawRuleInfo {
  int64 min_withdraw_amount = 1;   // 最低提现金额(分)
  int64 max_withdraw_amount = 2;   // 最高提现金额(分)
  int32 today_withdraw_count = 3;
  int32 today_withdraw_limit = 4;
  string estimated_arrival_text = 5;
  bool fee_free = 6;
  int64 fee_amount = 7;            // 手续费(分)
  string fee_desc = 8;
}

message GetWithdrawPageResp {
  int64 wallet_balance = 1;           // 钱包余额(分)
  int64 frozen_amount = 2;            // 冻结金额(分)
  int64 available_withdraw_amount = 3; // 可提现金额(分)
  WithdrawRuleInfo rule_info = 4;
  WithdrawPageBankCard bank_card = 5;
  WithdrawPageActionState action_state = 6;
  repeated int64 suggested_amounts = 7; // 推荐金额(分)
  repeated string withdraw_notice = 8;
}
```

- [ ] **Step 6: 更新 IncomeDetailItem 和 GetIncomeDetailResp**

```proto
message IncomeDetailItem {
  string type_name = 1;
  int32 type_code = 2;
  int64 amount = 3;               // 金额(分)
  int32 count = 4;
}

message GetIncomeDetailResp {
  repeated IncomeDetailItem items = 1;
  int64 total_amount = 2;         // 总金额(分)
}
```

- [ ] **Step 7: 更新 GetDriverIncomeResp 相关金额字段**

```proto
message DailyIncome {
  string date = 1;
  int64 income = 2;              // 收入(分)
}

message IncomeSummary {
  int32 order_count = 1;
  int64 income = 2;              // 收入(分)
  int32 online_duration = 3;
}

message GetDriverIncomeResp {
  IncomeSummary summary = 1;
  repeated DailyIncome trend = 2;
}
```

- [ ] **Step 8: 更新 WithdrawPageActionState 金额字段（如有）**

确认 `WithdrawPageActionState` 无金额字段，跳过。

- [ ] **Step 9: 提交 Proto 变更**

```bash
git add taketaxi/common/idl/driver.proto
git commit -m "feat(proto): migrate amount fields to int64 (cents)"
```

---

### Task 2: 重新生成 Proto 代码

**Files:**
- Modify: `taketaxi/common/kitexGen/driver.pb.go`
- Modify: `taketaxi/common/kitexGen/driver_grpc.pb.go`

- [ ] **Step 1: 运行 proto 生成脚本**

```bash
cd D:/software/GoWork/src/driver/taketaxi
./scripts/gen_proto.sh
```

Expected: 脚本成功执行，无报错

- [ ] **Step 2: 验证生成代码包含 int64 类型**

```bash
grep -n "int64.*balance\|int64.*amount\|int64.*income" taketaxi/common/kitexGen/driver.pb.go | head -20
```

Expected: 输出显示 `int64` 类型的金额字段

- [ ] **Step 3: 提交生成代码**

```bash
git add taketaxi/common/kitexGen/driver.pb.go taketaxi/common/kitexGen/driver_grpc.pb.go
git commit -m "chore: regenerate proto code with int64 amount fields"
```

---

### Task 3: 更新 Model 金额字段类型

**Files:**
- Modify: `taketaxi/srvDriver/internal/model/wallet.go`

- [ ] **Step 1: 更新 DriverWallet 结构体金额字段**

将 `DriverWallet` 中的金额字段从 `float64` 改为 `int64`：

```go
type DriverWallet struct {
    Id            int64     `gorm:"column:id;primaryKey;autoIncrement;comment:主键ID" json:"id"`
    DriverId      int64     `gorm:"column:driver_id;comment:司机ID" json:"driver_id"`
    Balance       int64     `gorm:"column:balance;comment:可用余额(分)" json:"balance"`
    FrozenAmount  int64     `gorm:"column:frozen_amount;comment:冻结金额(分)" json:"frozen_amount"`
    TotalIncome   int64     `gorm:"column:total_income;comment:累计总收入(分)" json:"total_income"`
    TotalWithdraw int64     `gorm:"column:total_withdraw;comment:累计提现金额(分)" json:"total_withdraw"`
    Version       int       `gorm:"column:version;comment:乐观锁版本号" json:"version"`
    UpdatedAt     time.Time `gorm:"column:updated_at;comment:更新时间" json:"updated_at"`
}
```

- [ ] **Step 2: 更新 DriverIncomeLog 结构体金额字段**

```go
type DriverIncomeLog struct {
    Id            int64     `gorm:"column:id;primaryKey;autoIncrement;comment:主键ID" json:"id"`
    DriverId      int64     `gorm:"column:driver_id;comment:司机ID" json:"driver_id"`
    OrderId       int64     `gorm:"column:order_id;comment:关联订单ID" json:"order_id"`
    Amount        int64     `gorm:"column:amount;comment:金额(分,正为收入,负为支出)" json:"amount"`
    Type          int8      `gorm:"column:type;comment:类型: 1-订单收入 2-奖励 3-罚款 4-提现 5-退款" json:"type"`
    BalanceBefore int64     `gorm:"column:balance_before;comment:变更前余额(分)" json:"balance_before"`
    BalanceAfter  int64     `gorm:"column:balance_after;comment:变更后余额(分)" json:"balance_after"`
    Remark        string    `gorm:"column:remark;comment:备注说明" json:"remark"`
    CreatedAt     time.Time `gorm:"column:created_at;comment:创建时间" json:"created_at"`
}
```

- [ ] **Step 3: 更新 WalletTransactionLog 结构体金额字段**

```go
type WalletTransactionLog struct {
    Id              int64     `gorm:"column:id;primaryKey;autoIncrement;comment:主键ID" json:"id"`
    DriverId        int64     `gorm:"column:driver_id;comment:司机ID" json:"driver_id"`
    TransactionNo   string    `gorm:"column:transaction_no;comment:流水号" json:"transaction_no"`
    TransactionType int8      `gorm:"column:transaction_type;comment:交易类型" json:"transaction_type"`
    Amount          int64     `gorm:"column:amount;comment:交易金额(分)" json:"amount"`
    BalanceBefore   int64     `gorm:"column:balance_before;comment:交易前余额(分)" json:"balance_before"`
    BalanceAfter    int64     `gorm:"column:balance_after;comment:交易后余额(分)" json:"balance_after"`
    FrozenBefore    int64     `gorm:"column:frozen_before;comment:交易前冻结金额(分)" json:"frozen_before"`
    FrozenAfter     int64     `gorm:"column:frozen_after;comment:交易后冻结金额(分)" json:"frozen_after"`
    RelatedId       int64     `gorm:"column:related_id;comment:关联ID" json:"related_id"`
    RelatedType     string    `gorm:"column:related_type;comment:关联类型" json:"related_type"`
    Status          int8      `gorm:"column:status;comment:状态" json:"status"`
    Remark          string    `gorm:"column:remark;comment:备注" json:"remark"`
    CreatedAt       time.Time `gorm:"column:created_at;comment:创建时间" json:"created_at"`
}
```

- [ ] **Step 4: 更新 DriverWithdrawRecord 结构体金额字段**

```go
type DriverWithdrawRecord struct {
    Id                int64      `gorm:"column:id;primaryKey;autoIncrement;comment:主键ID" json:"id"`
    WithdrawNo        string     `gorm:"column:withdraw_no;comment:提现单号" json:"withdraw_no"`
    DriverId          int64      `gorm:"column:driver_id;comment:司机ID" json:"driver_id"`
    Amount            int64      `gorm:"column:amount;comment:提现金额(分)" json:"amount"`
    Fee               int64      `gorm:"column:fee;comment:手续费(分)" json:"fee"`
    ActualAmount      int64      `gorm:"column:actual_amount;comment:实际到账金额(分)" json:"actual_amount"`
    BankName          string     `gorm:"column:bank_name;comment:银行名称" json:"bank_name"`
    BankCode          string     `gorm:"column:bank_code;comment:银行代码" json:"bank_code"`
    BankCardNo        string     `gorm:"column:bank_card_no;comment:银行卡号(脱敏)" json:"bank_card_no"`
    BankCardNoEncrypt string     `gorm:"column:bank_card_no_encrypt;comment:银行卡号(加密)" json:"bank_card_no_encrypt"`
    AccountName       string     `gorm:"column:account_name;comment:持卡人姓名" json:"account_name"`
    Status            int8       `gorm:"column:status;comment:状态: 1-处理中 2-成功 3-失败" json:"status"`
    FailReason        string     `gorm:"column:fail_reason;comment:失败原因" json:"fail_reason"`
    ApplyTime         time.Time  `gorm:"column:apply_time;comment:申请时间" json:"apply_time"`
    FinishTime        *time.Time `gorm:"column:finish_time;comment:完成时间" json:"finish_time"`
    Channel           string     `gorm:"column:channel;comment:提现渠道" json:"channel"`
    ChannelSerial     string     `gorm:"column:channel_serial;comment:渠道流水号" json:"channel_serial"`
    CreatedAt         time.Time  `gorm:"column:created_at;comment:创建时间" json:"created_at"`
    UpdatedAt         time.Time  `gorm:"column:updated_at;comment:更新时间" json:"updated_at"`
}
```

- [ ] **Step 5: 提交 Model 变更**

```bash
git add taketaxi/srvDriver/internal/model/wallet.go
git commit -m "feat(model): migrate wallet amount fields to int64 (cents)"
```

---

### Task 4: 更新钱包相关常量值

**Files:**
- Modify: `taketaxi/srvDriver/internal/model/enum_wallet.go`

- [ ] **Step 1: 更新提现限制常量（元转分）**

```go
const (
    MaxWithdrawPerDay    = 3       // 每日最大提现次数
    MaxWithdrawPerOrder  = 500000  // 单笔提现上限（分）= 5000元
    WithdrawChannelDefault = "default" // 默认提现渠道
    WithdrawMinAmount    = 100     // 最低提现金额（分）= 1元
)
```

- [ ] **Step 2: 提交常量变更**

```bash
git add taketaxi/srvDriver/internal/model/enum_wallet.go
git commit -m "feat: update withdraw constants to cents (×100)"
```

---

### Task 5: 新增 Repository 方法 - 提现记录查询与更新

**Files:**
- Modify: `taketaxi/srvDriver/internal/repository/withdrawRepo.go`

- [ ] **Step 1: 添加 GetWithdrawRecordByNo 方法**

在 `withdrawRepo.go` 末尾添加：

```go
// GetWithdrawRecordByNo 根据提现单号查询记录
func (r *DriverRepo) GetWithdrawRecordByNo(ctx context.Context, withdrawNo string) (*model.DriverWithdrawRecord, error) {
    var record model.DriverWithdrawRecord
    err := r.db.WithContext(ctx).Where("withdraw_no = ?", withdrawNo).First(&record).Error
    if err != nil {
        return nil, err
    }
    return &record, nil
}
```

- [ ] **Step 2: 添加 UpdateWithdrawStatus 方法**

```go
// UpdateWithdrawStatus 更新提现记录状态
func (r *DriverRepo) UpdateWithdrawStatus(ctx context.Context, withdrawNo string, status int8, failReason string) error {
    now := time.Now()
    return r.db.WithContext(ctx).Model(&model.DriverWithdrawRecord{}).
        Where("withdraw_no = ?", withdrawNo).
        Updates(map[string]interface{}{
            "status":      status,
            "fail_reason": failReason,
            "finish_time": &now,
            "updated_at":  now,
        }).Error
}
```

- [ ] **Step 3: 编译验证**

```bash
cd D:/software/GoWork/src/driver
go build ./taketaxi/srvDriver/...
```

Expected: 编译成功，无报错

- [ ] **Step 4: 提交 Repository 变更**

```bash
git add taketaxi/srvDriver/internal/repository/withdrawRepo.go
git commit -m "feat(repo): add GetWithdrawRecordByNo and UpdateWithdrawStatus"
```

---

### Task 6: 新增 Repository 方法 - 银行卡换卡频次校验

**Files:**
- Modify: `taketaxi/srvDriver/internal/repository/bankCardRepo.go`

- [ ] **Step 1: 添加 CheckBankCardChangeLimit 方法**

在 `bankCardRepo.go` 末尾添加：

```go
import (
    "errors"
    "time"
    // ... 其他已有的 import
)

// CheckBankCardChangeLimit 检查银行卡更换是否超过频次限制
// 规则：滚动周期 30 天内只能换 1 次
// 返回：ok=true 允许更换，ok=false 拒绝，nextAvailable 下次可换时间
func (r *DriverRepo) CheckBankCardChangeLimit(ctx context.Context, driverID int64) (ok bool, nextAvailable time.Time, err error) {
    var card model.DriverBankCard
    err = r.db.WithContext(ctx).
        Where("driver_id = ? AND status = ?", driverID, model.BankCardStatusActive).
        First(&card).Error
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            // 首次绑卡，无限制
            return true, time.Time{}, nil
        }
        return false, time.Time{}, err
    }

    // 滚动周期检查：距离上次修改需满 30 天
    nextAvailable = card.LastModifiedAt.AddDate(0, 0, 30)
    if time.Now().Before(nextAvailable) {
        return false, nextAvailable, nil
    }
    return true, time.Time{}, nil
}
```

- [ ] **Step 2: 编译验证**

```bash
cd D:/software/GoWork/src/driver
go build ./taketaxi/srvDriver/...
```

Expected: 编译成功，无报错

- [ ] **Step 3: 提交 Repository 变更**

```bash
git add taketaxi/srvDriver/internal/repository/bankCardRepo.go
git commit -m "feat(repo): add CheckBankCardChangeLimit for 30-day rolling period"
```

---

### Task 7: 更新 Wallet Repository 适配 int64

**Files:**
- Modify: `taketaxi/srvDriver/internal/repository/walletRepo.go`

- [ ] **Step 1: 检查 UpdateWallet 方法**

确认 `UpdateWallet` 方法中 `balance`、`frozen_amount` 等字段的 Updates 语句无需修改（GORM 会根据结构体字段类型自动处理）。

当前代码已正确，无需修改。

- [ ] **Step 2: 检查 CreateWalletTransactionLog 方法**

确认方法签名接受 `*model.WalletTransactionLog`，字段类型已随 model 更新。

当前代码已正确，无需修改。

- [ ] **Step 3: 编译验证**

```bash
go build ./taketaxi/srvDriver/...
```

Expected: 编译成功

---

## Phase 2: 业务逻辑层

### Task 8: 重构提现资格校验顺序

**Files:**
- Modify: `taketaxi/srvDriver/internal/service/walletService.go`

- [ ] **Step 1: 重写 evaluateWithdrawEligibility 方法**

将校验顺序调整为：实名 → 银行卡 → 次数 → 余额

```go
func (s *DriverService) evaluateWithdrawEligibility(
    ctx context.Context,
    driverID int64,
    wallet *model.DriverWallet,
) withdrawEligibility {
    // 1. 实名认证校验（优先级最高）
    realname, _ := s.repo.GetDriverRealname(ctx, driverID)
    if realname == nil || realname.Status != model.VerifyStatusApproved {
        return withdrawEligibility{
            CanWithdraw:       false,
            DisableReasonCode: model.WithdrawDisableVerify,
            DisableReasonText: "需完成实名认证",
        }
    }

    // 2. 银行卡校验
    card, _ := s.repo.GetBankCard(ctx, driverID)
    if card == nil {
        return withdrawEligibility{
            CanWithdraw:       false,
            DisableReasonCode: model.WithdrawDisableNoCard,
            DisableReasonText: "请先绑定银行卡",
        }
    }

    // 3. 今日提现次数校验
    todayCount, _ := s.repo.GetTodayWithdrawCount(ctx, driverID)
    if todayCount >= int64(model.MaxWithdrawPerDay) {
        return withdrawEligibility{
            CanWithdraw:       false,
            DisableReasonCode: model.WithdrawDisableCount,
            DisableReasonText: fmt.Sprintf("今日提现次数已达上限(%d次)", model.MaxWithdrawPerDay),
        }
    }

    // 4. 可提现余额校验
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
```

- [ ] **Step 2: 编译验证**

```bash
go build ./taketaxi/srvDriver/...
```

Expected: 编译成功

- [ ] **Step 3: 提交变更**

```bash
git add taketaxi/srvDriver/internal/service/walletService.go
git commit -m "fix: reorder withdraw eligibility checks (realname → bankcard → count)"
```

---

### Task 9: 更新 ApplyWithdraw 适配 int64

**Files:**
- Modify: `taketaxi/srvDriver/internal/service/walletService.go`

- [ ] **Step 1: 更新 ApplyWithdraw 方法中的金额比较逻辑**

找到 `ApplyWithdraw` 方法，更新金额比较逻辑（移除 float64 比较）：

```go
// 1. 最低金额校验
if req.Amount < model.WithdrawMinAmount {
    return nil, errcode.New(errcode.ErrWithdrawMinAmount)
}
if req.Amount > model.MaxWithdrawPerOrder {
    return nil, errcode.New(errcode.ErrWithdrawAmountLimit)
}

// ... 省略中间代码 ...

// 4. 余额校验
if wallet.Balance < req.Amount {
    return nil, errcode.NewWithDetail(errcode.ErrInsufficientBalance,
        fmt.Sprintf("balance=%d", wallet.Balance))
}

// 5. T-3 冻结校验
if wallet.FrozenAmount > 0 && (wallet.Balance-wallet.FrozenAmount) < req.Amount {
    return nil, errcode.NewWithDetail(errcode.ErrFrozenAmountLimit,
        fmt.Sprintf("frozen=%d", wallet.FrozenAmount))
}
```

- [ ] **Step 2: 更新提现记录创建时的金额字段**

```go
record := &model.DriverWithdrawRecord{
    WithdrawNo:        withdrawNo,
    DriverId:          req.DriverId,
    Amount:            req.Amount,
    Fee:               0,
    ActualAmount:      req.Amount,
    // ... 其他字段不变
}
```

- [ ] **Step 3: 更新流水记录创建时的金额字段**

```go
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
```

- [ ] **Step 4: 编译验证**

```bash
go build ./taketaxi/srvDriver/...
```

Expected: 编译成功

- [ ] **Step 5: 提交变更**

```bash
git add taketaxi/srvDriver/internal/service/walletService.go
git commit -m "feat: adapt ApplyWithdraw for int64 amount (cents)"
```

---

### Task 10: 更新 GetWithdrawPage 适配 int64

**Files:**
- Modify: `taketaxi/srvDriver/internal/service/walletService.go`

- [ ] **Step 1: 更新 GetWithdrawPage 方法返回值**

更新返回结构中的金额字段：

```go
// 5. 计算可提现金额
available := wallet.Balance - wallet.FrozenAmount
if available < 0 {
    available = 0
}

// 6. 推荐提现金额（分）
suggestedAmounts := s.computeSuggestedAmounts(available)

// 7. 组装规则信息
ruleInfo := &driver.WithdrawRuleInfo{
    MinWithdrawAmount:   int64(model.WithdrawMinAmount),
    MaxWithdrawAmount:   int64(model.MaxWithdrawPerOrder),
    TodayWithdrawCount:  int32(todayCount),
    TodayWithdrawLimit:  int32(model.MaxWithdrawPerDay),
    EstimatedArrivalText: model.WithdrawArrivalText,
    FeeFree:             true,
    FeeAmount:           0,
    FeeDesc:             "免手续费",
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
```

- [ ] **Step 2: 更新 computeSuggestedAmounts 方法**

```go
func (s *DriverService) computeSuggestedAmounts(available int64) []int64 {
    if available <= 0 {
        return nil
    }
    presets := []int64{5000, 10000, 20000, 50000, 100000, 200000} // 50元、100元...（分）
    var result []int64
    for _, p := range presets {
        if p <= available {
            result = append(result, p)
        }
    }
    if available > 200000 {
        result = append(result, available)
    }
    return result
}
```

- [ ] **Step 3: 编译验证**

```bash
go build ./taketaxi/srvDriver/...
```

Expected: 编译成功

- [ ] **Step 4: 提交变更**

```bash
git add taketaxi/srvDriver/internal/service/walletService.go
git commit -m "feat: adapt GetWithdrawPage for int64 amount (cents)"
```

---

### Task 11: 更新 GetWallet 适配 int64

**Files:**
- Modify: `taketaxi/srvDriver/internal/service/walletService.go`

- [ ] **Step 1: 更新 GetWallet 方法返回值**

找到 `GetWallet` 方法，更新返回结构：

```go
return &driver.GetWalletResp{
    Balance:            wallet.Balance,
    FrozenAmount:       wallet.FrozenAmount,
    TodayIncome:        todayStats.TotalIncome,
    WeekIncome:         weekStats.TotalIncome,
    MonthIncome:        monthStats.TotalIncome,
    TotalIncome:        wallet.TotalIncome,
    TotalWithdraw:      wallet.TotalWithdraw,
    TodayWithdrawCount: int32(todayCount),
    BankCardNo:         bankCardNo,
    HasBankCard:        card != nil,
}, nil
```

- [ ] **Step 2: 更新 GetOrderStats 返回值类型**

检查 `GetOrderStats` 方法返回的 `TotalIncome` 类型是否为 `int64`。如果 model 中 `IncomeDetailResult.Amount` 是 `float64`，需要更新。

找到 `model/result.go` 或相关文件，确认类型：

```go
type IncomeDetailResult struct {
    TypeCode int8
    TypeName string
    Amount   int64  // 改为 int64
    Count    int
}
```

- [ ] **Step 3: 更新 GetIncomeDetail 方法**

确认方法中金额累加逻辑：

```go
var totalAmount int64
for _, r := range results {
    totalAmount += r.Amount
    // ...
}
```

- [ ] **Step 4: 编译验证**

```bash
go build ./taketaxi/srvDriver/...
```

Expected: 编译成功

- [ ] **Step 5: 提交变更**

```bash
git add taketaxi/srvDriver/internal/model/result.go taketaxi/srvDriver/internal/service/walletService.go
git commit -m "feat: adapt GetWallet and GetIncomeDetail for int64 amount"
```

---

### Task 12: 新增 MarkWithdrawFailed 服务方法

**Files:**
- Modify: `taketaxi/srvDriver/internal/service/walletService.go`

- [ ] **Step 1: 添加 MarkWithdrawFailed 方法**

在 `walletService.go` 末尾添加：

```go
// MarkWithdrawFailed 标记提现失败并回退余额
// 供后台管理接口调用，支持人工干预
func (s *DriverService) MarkWithdrawFailed(ctx context.Context, withdrawNo string, failReason string) error {
    // 1. 查询提现记录
    record, err := s.repo.GetWithdrawRecordByNo(ctx, withdrawNo)
    if err != nil {
        return errcode.NewWithDetail(errcode.ErrRecordNotFound, "提现记录不存在")
    }

    // 2. 状态校验：仅"处理中"可标记失败
    if record.Status != model.WithdrawStatusPending {
        return errcode.NewWithDetail(errcode.ErrInvalidStatus,
            fmt.Sprintf("当前状态不可变更: %d", record.Status))
    }

    // 3. 事务：更新状态 + 回补余额 + 写流水
    err = s.repo.RunInTx(ctx, func(txRepo *repository.DriverRepo) error {
        // 3a. 更新提现记录状态
        if err := txRepo.UpdateWithdrawStatus(ctx, withdrawNo, model.WithdrawStatusFailed, failReason); err != nil {
            return err
        }

        // 3b. 查询钱包并回补余额
        wallet, err := txRepo.GetWallet(ctx, record.DriverId)
        if err != nil {
            return err
        }
        wallet.Balance += record.Amount
        wallet.TotalWithdraw -= record.Amount
        if err := txRepo.UpdateWallet(ctx, wallet); err != nil {
            return err
        }

        // 3c. 写入退款流水
        refundLog := &model.WalletTransactionLog{
            DriverId:        record.DriverId,
            TransactionNo:   fmt.Sprintf("RF%s", time.Now().Format("20060102150405")),
            TransactionType: model.WalletTxTypeRefund,
            Amount:          record.Amount,
            BalanceBefore:   wallet.Balance - record.Amount,
            BalanceAfter:    wallet.Balance,
            FrozenBefore:    wallet.FrozenAmount,
            FrozenAfter:     wallet.FrozenAmount,
            RelatedId:       record.Id,
            RelatedType:     "withdraw_refund",
            Status:          model.WalletTxStatusSuccess,
            Remark:          fmt.Sprintf("提现失败回退: %s", failReason),
        }
        return txRepo.CreateWalletTransactionLog(ctx, refundLog)
    })

    if err != nil {
        if be, ok := err.(*errcode.BusinessError); ok {
            return be
        }
        return errcode.NewWithDetail(errcode.ErrInternal, err.Error())
    }

    logger.Info("提现失败回退成功",
        zap.String("withdraw_no", withdrawNo),
        zap.Int64("driver_id", record.DriverId),
        zap.Int64("refund_amount", record.Amount),
        zap.String("fail_reason", failReason),
    )

    return nil
}
```

- [ ] **Step 2: 添加缺失的错误码**

检查 `errcode/errcode.go` 是否有 `ErrRecordNotFound` 和 `ErrInvalidStatus`，若无则添加：

```go
const (
    // ... 已有错误码
    ErrRecordNotFound Code = 50023 // 记录不存在
    ErrInvalidStatus  Code = 50024 // 状态不允许操作
)

var codeMessages = map[Code]string{
    // ... 已有映射
    ErrRecordNotFound: "记录不存在",
    ErrInvalidStatus:  "状态不允许操作",
}
```

- [ ] **Step 3: 编译验证**

```bash
go build ./taketaxi/srvDriver/...
```

Expected: 编译成功

- [ ] **Step 4: 提交变更**

```bash
git add taketaxi/srvDriver/internal/service/walletService.go taketaxi/pkg/errcode/errcode.go
git commit -m "feat: add MarkWithdrawFailed service method with refund transaction"
```

---

### Task 13: 更新银行卡绑定服务添加换卡频次校验

**Files:**
- Modify: `taketaxi/srvDriver/internal/handler/bankCardHandler.go` 或对应 service 文件

- [ ] **Step 1: 找到银行卡绑定/更换的服务方法**

检查是否存在 `bankCardService.go` 或在 handler 中直接处理。

假设在 `srvDriver/internal/handler/bankCardHandler.go` 或对应的 service 层。

- [ ] **Step 2: 在更换银行卡时添加频次校验**

在 `UpdateBankCard` 或 `BindBankCard`（当已有卡时）方法中添加：

```go
// 检查是否已有银行卡
existingCard, _ := s.repo.GetBankCard(ctx, req.DriverId)

if existingCard != nil {
    // 更换银行卡：检查滚动周期限制
    ok, nextAvailable, err := s.repo.CheckBankCardChangeLimit(ctx, req.DriverId)
    if err != nil {
        return nil, err
    }
    if !ok {
        return nil, errcode.NewWithDetail(errcode.ErrBankCardChangeLimit,
            fmt.Sprintf("下次可换卡时间: %s", nextAvailable.Format("2006-01-02")))
    }
}
```

- [ ] **Step 3: 编译验证**

```bash
go build ./taketaxi/srvDriver/...
```

Expected: 编译成功

- [ ] **Step 4: 提交变更**

```bash
git add taketaxi/srvDriver/internal/handler/bankCardHandler.go
git commit -m "feat: add bank card change limit check (30-day rolling period)"
```

---

## Phase 3: 接口层

### Task 14: 新增 Proto 定义 - MarkWithdrawFailed

**Files:**
- Modify: `taketaxi/common/idl/driver.proto`

- [ ] **Step 1: 添加 MarkWithdrawFailed RPC 和消息定义**

在 `DriverService` 中添加新 RPC：

```proto
service DriverService {
  // ... 已有方法
  rpc MarkWithdrawFailed(MarkWithdrawFailedReq) returns (MarkWithdrawFailedResp);
}

message MarkWithdrawFailedReq {
  string withdraw_no = 1;
  string fail_reason = 2;
}

message MarkWithdrawFailedResp {
  bool success = 1;
  string message = 2;
}
```

- [ ] **Step 2: 重新生成 Proto 代码**

```bash
cd D:/software/GoWork/src/driver/taketaxi
./scripts/gen_proto.sh
```

- [ ] **Step 3: 提交变更**

```bash
git add taketaxi/common/idl/driver.proto taketaxi/common/kitexGen/driver.pb.go taketaxi/common/kitexGen/driver_grpc.pb.go
git commit -m "feat(proto): add MarkWithdrawFailed RPC"
```

---

### Task 15: 新增 srvDriver Handler - MarkWithdrawFailed

**Files:**
- Modify: `taketaxi/srvDriver/internal/handler/walletHandler.go`

- [ ] **Step 1: 添加 MarkWithdrawFailed Handler 方法**

```go
// MarkWithdrawFailed 标记提现失败（后台管理接口）
func (h *DriverHandler) MarkWithdrawFailed(ctx context.Context, req *driver.MarkWithdrawFailedReq) (*driver.MarkWithdrawFailedResp, error) {
    start := time.Now()
    err := h.svc.MarkWithdrawFailed(ctx, req.WithdrawNo, req.FailReason)
    if err != nil {
        logger.Error("gRPC MarkWithdrawFailed failed",
            zap.String("method", "MarkWithdrawFailed"),
            zap.String("withdraw_no", req.WithdrawNo),
            zap.Error(err))
        return nil, err
    }
    logger.Info("gRPC MarkWithdrawFailed success",
        zap.String("method", "MarkWithdrawFailed"),
        zap.String("withdraw_no", req.WithdrawNo),
        zap.Duration("duration", time.Since(start)))
    return &driver.MarkWithdrawFailedResp{Success: true, Message: "标记成功"}, nil
}
```

- [ ] **Step 2: 编译验证**

```bash
go build ./taketaxi/srvDriver/...
```

Expected: 编译成功

- [ ] **Step 3: 提交变更**

```bash
git add taketaxi/srvDriver/internal/handler/walletHandler.go
git commit -m "feat(handler): add MarkWithdrawFailed gRPC handler"
```

---

### Task 16: 新增 RPC Client 方法

**Files:**
- Modify: `taketaxi/bffDriver/internal/rpcClient/driverClient.go`

- [ ] **Step 1: 添加 MarkWithdrawFailed 方法**

```go
// MarkWithdrawFailed 标记提现失败（后台管理）
func (c *DriverClient) MarkWithdrawFailed(ctx context.Context, req *driver.MarkWithdrawFailedReq) (*driver.MarkWithdrawFailedResp, error) {
    return c.client.MarkWithdrawFailed(ctx, req)
}
```

- [ ] **Step 2: 编译验证**

```bash
go build ./taketaxi/bffDriver/...
```

Expected: 编译成功

- [ ] **Step 3: 提交变更**

```bash
git add taketaxi/bffDriver/internal/rpcClient/driverClient.go
git commit -m "feat(rpc): add MarkWithdrawFailed client method"
```

---

### Task 17: 新增 BFF Handler - MarkWithdrawFailed

**Files:**
- Modify: `taketaxi/bffDriver/internal/handler/driverHandler.go`

- [ ] **Step 1: 添加金额转换辅助函数**

在文件顶部添加：

```go
// 金额转换辅助函数
func centsToYuan(cents int64) float64 {
    return float64(cents) / 100.0
}

func yuanToCents(yuan float64) int64 {
    return int64(yuan * 100)
}

// 银行卡脱敏（取后4位）
func maskBankCard(cardNo string) string {
    if len(cardNo) <= 4 {
        return cardNo
    }
    return cardNo[len(cardNo)-4:]
}
```

- [ ] **Step 2: 添加 MarkWithdrawFailed Handler 方法**

```go
// MarkWithdrawFailed 后台标记提现失败
// POST /api/v1/admin/withdraw/mark-failed
// 请求体: {"withdraw_no":"WD20260502143015345","fail_reason":"银行卡已冻结"}
func (h *DriverHandler) MarkWithdrawFailed(c *gin.Context) {
    var req struct {
        WithdrawNo string `json:"withdraw_no" binding:"required"`
        FailReason string `json:"fail_reason" binding:"required"`
    }
    if err := c.ShouldBindJSON(&req); err != nil {
        logger.Warn("MarkWithdrawFailed 参数错误", zap.Error(err))
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    resp, err := h.client.MarkWithdrawFailed(c.Request.Context(), &pb.MarkWithdrawFailedReq{
        WithdrawNo: req.WithdrawNo,
        FailReason: req.FailReason,
    })
    if err != nil {
        logger.Error("MarkWithdrawFailed 失败",
            zap.String("withdraw_no", req.WithdrawNo),
            zap.Error(err))
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    logger.Info("MarkWithdrawFailed 成功", zap.String("withdraw_no", req.WithdrawNo))
    c.JSON(http.StatusOK, resp)
}
```

- [ ] **Step 3: 编译验证**

```bash
go build ./taketaxi/bffDriver/...
```

Expected: 编译成功

- [ ] **Step 4: 提交变更**

```bash
git add taketaxi/bffDriver/internal/handler/driverHandler.go
git commit -m "feat(bff): add MarkWithdrawFailed HTTP handler and amount helpers"
```

---

### Task 18: 更新 BFF 金额转换 - GetWallet

**Files:**
- Modify: `taketaxi/bffDriver/internal/handler/driverHandler.go`

- [ ] **Step 1: 更新 GetWallet 方法添加金额转换**

```go
// GetWallet 查询钱包概览
func (h *DriverHandler) GetWallet(c *gin.Context) {
    driverIDStr := c.Query("driver_id")
    if driverIDStr == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "driver_id is required"})
        return
    }
    driverID, err := strconv.ParseInt(driverIDStr, 10, 64)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid driver_id"})
        return
    }

    pbResp, err := h.client.GetWallet(c.Request.Context(), &pb.GetWalletReq{DriverId: driverID})
    if err != nil {
        logger.Error("GetWallet 失败", zap.Int64("driver_id", driverID), zap.Error(err))
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    // 金额转换：分 → 元
    resp := gin.H{
        "balance":              centsToYuan(pbResp.Balance),
        "frozen_amount":        centsToYuan(pbResp.FrozenAmount),
        "today_income":         centsToYuan(pbResp.TodayIncome),
        "week_income":          centsToYuan(pbResp.WeekIncome),
        "month_income":         centsToYuan(pbResp.MonthIncome),
        "total_income":         centsToYuan(pbResp.TotalIncome),
        "total_withdraw":       centsToYuan(pbResp.TotalWithdraw),
        "today_withdraw_count": pbResp.TodayWithdrawCount,
        "bank_card_no":         pbResp.BankCardNo,
        "has_bank_card":        pbResp.HasBankCard,
    }
    c.JSON(http.StatusOK, resp)
}
```

- [ ] **Step 2: 编译验证**

```bash
go build ./taketaxi/bffDriver/...
```

Expected: 编译成功

- [ ] **Step 3: 提交变更**

```bash
git add taketaxi/bffDriver/internal/handler/driverHandler.go
git commit -m "feat(bff): add amount conversion in GetWallet (cents to yuan)"
```

---

### Task 19: 更新 BFF 金额转换 - ApplyWithdraw

**Files:**
- Modify: `taketaxi/bffDriver/internal/handler/driverHandler.go`

- [ ] **Step 1: 更新 ApplyWithdraw 方法添加金额转换**

```go
// ApplyWithdraw 申请提现
func (h *DriverHandler) ApplyWithdraw(c *gin.Context) {
    var req struct {
        DriverId int64   `json:"driver_id" binding:"required"`
        Amount   float64 `json:"amount" binding:"required"` // 前端传入元
    }
    if err := c.ShouldBindJSON(&req); err != nil {
        logger.Warn("ApplyWithdraw 参数错误", zap.Error(err))
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    resp, err := h.client.ApplyWithdraw(c.Request.Context(), &pb.ApplyWithdrawReq{
        DriverId: req.DriverId,
        Amount:   yuanToCents(req.Amount), // 元 → 分
    })
    if err != nil {
        logger.Error("ApplyWithdraw 失败", zap.Int64("driver_id", req.DriverId), zap.Error(err))
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    logger.Info("ApplyWithdraw 成功", zap.Int64("driver_id", req.DriverId))
    c.JSON(http.StatusOK, resp)
}
```

- [ ] **Step 2: 编译验证**

```bash
go build ./taketaxi/bffDriver/...
```

Expected: 编译成功

- [ ] **Step 3: 提交变更**

```bash
git add taketaxi/bffDriver/internal/handler/driverHandler.go
git commit -m "feat(bff): add amount conversion in ApplyWithdraw (yuan to cents)"
```

---

### Task 20: 更新 BFF 金额转换 - GetWithdrawPage

**Files:**
- Modify: `taketaxi/bffDriver/internal/handler/driverHandler.go`

- [ ] **Step 1: 更新 GetWithdrawPage 方法添加金额转换**

```go
// GetWithdrawPage 查询提现页信息
func (h *DriverHandler) GetWithdrawPage(c *gin.Context) {
    driverIDStr := c.Query("driver_id")
    if driverIDStr == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "driver_id is required"})
        return
    }
    driverID, err := strconv.ParseInt(driverIDStr, 10, 64)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid driver_id"})
        return
    }

    pbResp, err := h.client.GetWithdrawPage(c.Request.Context(), &pb.GetWithdrawPageReq{DriverId: driverID})
    if err != nil {
        logger.Error("GetWithdrawPage 失败", zap.Int64("driver_id", driverID), zap.Error(err))
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    // 金额转换：分 → 元
    suggestedYuan := make([]float64, len(pbResp.SuggestedAmounts))
    for i, v := range pbResp.SuggestedAmounts {
        suggestedYuan[i] = centsToYuan(v)
    }

    resp := gin.H{
        "wallet_balance":             centsToYuan(pbResp.WalletBalance),
        "frozen_amount":              centsToYuan(pbResp.FrozenAmount),
        "available_withdraw_amount":  centsToYuan(pbResp.AvailableWithdrawAmount),
        "rule_info": gin.H{
            "min_withdraw_amount":   centsToYuan(pbResp.RuleInfo.MinWithdrawAmount),
            "max_withdraw_amount":   centsToYuan(pbResp.RuleInfo.MaxWithdrawAmount),
            "today_withdraw_count":  pbResp.RuleInfo.TodayWithdrawCount,
            "today_withdraw_limit":  pbResp.RuleInfo.TodayWithdrawLimit,
            "estimated_arrival_text": pbResp.RuleInfo.EstimatedArrivalText,
            "fee_free":              pbResp.RuleInfo.FeeFree,
            "fee_amount":            centsToYuan(pbResp.RuleInfo.FeeAmount),
            "fee_desc":              pbResp.RuleInfo.FeeDesc,
        },
        "bank_card":      pbResp.BankCard,
        "action_state":   pbResp.ActionState,
        "suggested_amounts": suggestedYuan,
        "withdraw_notice":   pbResp.WithdrawNotice,
    }
    c.JSON(http.StatusOK, resp)
}
```

- [ ] **Step 2: 编译验证**

```bash
go build ./taketaxi/bffDriver/...
```

Expected: 编译成功

- [ ] **Step 3: 提交变更**

```bash
git add taketaxi/bffDriver/internal/handler/driverHandler.go
git commit -m "feat(bff): add amount conversion in GetWithdrawPage"
```

---

### Task 21: 新增 Admin 路由

**Files:**
- Modify: `taketaxi/bffDriver/internal/router/router.go`

- [ ] **Step 1: 添加 Admin 路由组**

在 `NewRouter` 函数中，文件上传路由之后添加：

```go
// ========== 后台管理接口 ==========
admin := r.Group("/api/v1/admin")
{
    admin.POST("/withdraw/mark-failed", driverHandler.MarkWithdrawFailed)
}
```

- [ ] **Step 2: 编译验证**

```bash
go build ./taketaxi/bffDriver/...
```

Expected: 编译成功

- [ ] **Step 3: 提交变更**

```bash
git add taketaxi/bffDriver/internal/router/router.go
git commit -m "feat(router): add admin routes for withdraw management"
```

---

### Task 22: 全量编译验证

**Files:**
- N/A

- [ ] **Step 1: 编译所有包**

```bash
cd D:/software/GoWork/src/driver
go build ./...
```

Expected: 无编译错误

- [ ] **Step 2: 运行现有测试**

```bash
go test ./taketaxi/... -v
```

Expected: 所有测试通过（或列出已知的失败测试）

- [ ] **Step 3: 最终提交**

```bash
git add -A
git status
```

检查是否有未提交的变更，如有则提交。

---

## 数据库迁移说明

以上代码改动完成后，需要 DBA 在部署前执行数据库迁移脚本（见设计文档 Section 2.2）。

迁移脚本将金额字段从 FLOAT/DOUBLE 迁移到 BIGINT，并乘以 100 转换为分。

# 提现后端代码优化设计

## 1. 概述

**背景：** 根据规格书 `2026-05-02-driver-withdraw-rules-boundary-design.md` 审计现有代码，发现多处实现与规则不一致或存在隐患。

**目标：** 
1. 金额精度从 float64 迁移到 int64（分）
2. 修正提现资格校验顺序
3. 实现银行卡换卡滚动周期限制
4. 实现提现失败回退机制
5. 补全审计日志

## 2. 改动范围

### 2.1 文件改动清单

| 阶段 | 文件 | 改动类型 |
|------|------|----------|
| Phase 1 | `common/idl/driver.proto` | 修改金额字段类型 |
| Phase 1 | `srvDriver/internal/model/wallet.go` | 修改字段类型 float64 → int64 |
| Phase 1 | `srvDriver/internal/model/enum_wallet.go` | 常量值乘以 100 |
| Phase 1 | `srvDriver/internal/repository/withdrawRepo.go` | 新增方法 |
| Phase 1 | `srvDriver/internal/repository/bankCardRepo.go` | 新增方法 |
| Phase 2 | `srvDriver/internal/service/walletService.go` | 重构校验顺序 + 新增失败回退 |
| Phase 2 | `srvDriver/internal/service/bankCardService.go` | 新增换卡频次校验 |
| Phase 3 | `bffDriver/internal/handler/driverHandler.go` | 新增后台接口 + 金额转换 |
| Phase 3 | `bffDriver/internal/router/router.go` | 新增路由 |
| Phase 3 | `srvDriver/internal/handler/walletHandler.go` | 新增 gRPC handler |

### 2.2 数据库迁移（DBA 执行）

```sql
-- 金额字段迁移：元 → 分
ALTER TABLE driver_wallet 
  MODIFY COLUMN balance BIGINT NOT NULL DEFAULT 0 COMMENT '可用余额(分)',
  MODIFY COLUMN frozen_amount BIGINT NOT NULL DEFAULT 0 COMMENT '冻结金额(分)',
  MODIFY COLUMN total_income BIGINT NOT NULL DEFAULT 0 COMMENT '累计总收入(分)',
  MODIFY COLUMN total_withdraw BIGINT NOT NULL DEFAULT 0 COMMENT '累计提现金额(分)';

UPDATE driver_wallet SET 
  balance = FLOOR(balance * 100),
  frozen_amount = FLOOR(frozen_amount * 100),
  total_income = FLOOR(total_income * 100),
  total_withdraw = FLOOR(total_withdraw * 100);

ALTER TABLE driver_withdraw_record
  MODIFY COLUMN amount BIGINT NOT NULL COMMENT '提现金额(分)',
  MODIFY COLUMN fee BIGINT NOT NULL DEFAULT 0 COMMENT '手续费(分)',
  MODIFY COLUMN actual_amount BIGINT NOT NULL COMMENT '实际到账金额(分)';

UPDATE driver_withdraw_record SET
  amount = FLOOR(amount * 100),
  fee = FLOOR(fee * 100),
  actual_amount = FLOOR(actual_amount * 100);

ALTER TABLE wallet_transaction_log
  MODIFY COLUMN amount BIGINT NOT NULL COMMENT '交易金额(分)',
  MODIFY COLUMN balance_before BIGINT NOT NULL COMMENT '交易前余额(分)',
  MODIFY COLUMN balance_after BIGINT NOT NULL COMMENT '交易后余额(分)',
  MODIFY COLUMN frozen_before BIGINT NOT NULL COMMENT '交易前冻结金额(分)',
  MODIFY COLUMN frozen_after BIGINT NOT NULL COMMENT '交易后冻结金额(分)';

UPDATE wallet_transaction_log SET
  amount = FLOOR(amount * 100),
  balance_before = FLOOR(balance_before * 100),
  balance_after = FLOOR(balance_after * 100),
  frozen_before = FLOOR(frozen_before * 100),
  frozen_after = FLOOR(frozen_after * 100);
```

## 3. Phase 1: 数据模型层

### 3.1 金额精度迁移

**Proto 定义更新：**

```proto
message GetWalletResp {
  int64 balance = 1;            // 可提现余额(分)
  int64 frozen_amount = 2;      // 冻结金额(分)
  int64 today_income = 3;       // 今日收入(分)
  // ... 其他金额字段
}

message ApplyWithdrawReq {
  int64 driver_id = 1;
  int64 amount = 2;             // 提现金额(分)
}
```

**Model 字段更新：**

```go
// model/wallet.go
type DriverWallet struct {
    Balance       int64 `gorm:"column:balance;comment:可用余额(分)"`
    FrozenAmount  int64 `gorm:"column:frozen_amount;comment:冻结金额(分)"`
    TotalIncome   int64 `gorm:"column:total_income;comment:累计总收入(分)"`
    TotalWithdraw int64 `gorm:"column:total_withdraw;comment:累计提现金额(分)"`
    // ...
}

// model/enum_wallet.go
const (
    MaxWithdrawPerOrder  = 500000 // 单笔提现上限（分）= 5000元
    WithdrawMinAmount    = 100    // 最低提现金额（分）= 1元
)
```

### 3.2 新增 Repository 方法

**withdrawRepo.go：**

```go
// UpdateWithdrawStatus 更新提现记录状态
func (r *DriverRepo) UpdateWithdrawStatus(ctx context.Context, withdrawNo string, status int8, failReason string) error

// GetWithdrawRecordByNo 根据提现单号查询记录
func (r *DriverRepo) GetWithdrawRecordByNo(ctx context.Context, withdrawNo string) (*model.DriverWithdrawRecord, error)
```

**bankCardRepo.go：**

```go
// CheckBankCardChangeLimit 检查银行卡更换是否超过频次限制
// 规则：滚动周期 30 天内只能换 1 次
// 返回：ok=true 允许更换，ok=false 拒绝，nextAvailable 下次可换时间
func (r *DriverRepo) CheckBankCardChangeLimit(ctx context.Context, driverID int64) (ok bool, nextAvailable time.Time, err error)
```

## 4. Phase 2: 业务逻辑层

### 4.1 重构提现资格校验顺序

**规格书定义顺序：** 实名 → 银行卡 → 次数 → 余额

```go
func (s *DriverService) evaluateWithdrawEligibility(...) withdrawEligibility {
    // 1. 实名认证校验（优先级最高）
    // 2. 银行卡校验
    // 3. 今日提现次数校验
    // 4. 可提现余额校验
}
```

### 4.2 提现失败回退事务

```go
// MarkWithdrawFailed 标记提现失败并回退余额
func (s *DriverService) MarkWithdrawFailed(ctx context.Context, withdrawNo string, failReason string) error {
    // 事务：
    //   1. 更新提现记录状态为失败
    //   2. 回补钱包余额
    //   3. 写入退款流水
}
```

### 4.3 审计日志规范

| 操作 | 必须记录字段 |
|------|-------------|
| 提现申请 | driver_id, amount(分), bank_card_tail(4位), withdraw_no, result |
| 提现失败 | withdraw_no, fail_reason, refund_amount, operator |
| 银行卡绑定 | driver_id, bank_code, card_tail(4位), is_first_bind |
| 余额变更 | driver_id, amount, balance_before, balance_after, reason |

## 5. Phase 3: 接口层

### 5.1 新增后台接口

```
POST /api/v1/admin/withdraw/mark-failed
请求体: {"withdraw_no":"WD20260502143015345","fail_reason":"银行卡已冻结"}
响应: {"success":true}
```

### 5.2 BFF 金额转换

**请求（前端元 → 后端分）：**
```go
func yuanToCents(yuan float64) int64 {
    return int64(yuan * 100)
}
```

**响应（后端分 → 前端元）：**
```go
func centsToYuan(cents int64) float64 {
    return float64(cents) / 100.0
}
```

### 5.3 路由配置

```go
admin := r.Group("/api/v1/admin")
{
    admin.POST("/withdraw/mark-failed", driverHandler.MarkWithdrawFailed)
}
```

## 6. 实施顺序

1. **Phase 1** - 数据模型层（需协调 DBA 执行数据库迁移）
2. **Phase 2** - 业务逻辑层
3. **Phase 3** - 接口层

每个阶段完成后进行回归测试，确保提现流程正常。

## 7. 风险与回滚

### 7.1 风险点

| 风险 | 影响 | 缓解措施 |
|------|------|----------|
| 金额迁移精度丢失 | 数据不一致 | 先在测试环境验证迁移脚本 |
| 校验顺序变更 | 部分用户提现行为改变 | 发布前通知运营 |
| 新接口权限泄露 | 误操作标记失败 | 接口加权限校验 |

### 7.2 回滚方案

- 代码回滚：Git revert
- 数据库回滚：执行逆向迁移脚本（分 → 元）

# 司机提现规则与边界设计（测试工程师视角）

> 作者：10年网约车测试工程师
> 日期：2026-05-02
> 基于当前项目规则，对标花小猪行业实践做边界梳理和缺口分析

---

## 一、当前项目提现规则总览

| 规则项 | 当前值 | 存储 |
|---|---|---|
| 最低提现金额 | 1元（100分） | `WithdrawMinAmount = 100` |
| 单笔上限 | 5000元（500000分） | `MaxWithdrawPerOrder = 500000` |
| 每日次数上限 | 3次 | `MaxWithdrawPerDay = 3` |
| 手续费 | 免费 | `FeeFree = true, FeeAmount = 0` |
| 预计到账 | 2小时 | `WithdrawArrivalText` |
| 结算周期 | T-3（3天前结算才可提） | `FrozenAmount` 字段 |
| 前置条件 | 银行卡 + 实名认证通过 | `evaluateWithdrawEligibility()` |
| 提现时间窗口 | 24h 随时 | 无限制 |

---

## 二、严重 Bug（编译级 / 逻辑级）

### BUG-1：金额单位不一致（编译级 🔴）

**Proto 生成代码已全部改为 `int64 分`**，但 `walletService.go` 仍用 `float64` 比较：

```go
// ❌ 当前代码（walletService.go）
if req.Amount < float64(model.WithdrawMinAmount) { ... }
if req.Amount > float64(model.MaxWithdrawPerOrder) { ... }
suggestedAmounts := []float64{50, 100, 200, 500, 1000, 2000}  // 元！
MinWithdrawAmount: float64(model.WithdrawMinAmount)            // 100.0 分 → 前端以为100元
```

```go
// ✅ 应该是
if req.Amount < model.WithdrawMinAmount { ... }               // int64 分
if req.Amount > model.MaxWithdrawPerOrder { ... }
suggestedAmounts := []int64{5000, 10000, 20000, 50000, 100000, 200000}  // 分
MinWithdrawAmount: model.WithdrawMinAmount                     // 100 分 = 1元
```

**影响范围**：`GetWithdrawPage` + `ApplyWithdraw` + `computeSuggestedAmounts` + `GetWallet`，全部金额字段需从 `float64` 改为 `int64`。

### BUG-2：T-3 冻结逻辑语义冲突（逻辑级 🔴）

Model 注释说 `Balance` 是"可用余额"（已扣除冻结），但代码又 `Balance - FrozenAmount`，等于二次扣减：

```go
// ❌ walletService.go:61
available := wallet.Balance - wallet.FrozenAmount

// ❌ walletService.go:196
if wallet.FrozenAmount > 0 && (wallet.Balance-wallet.FrozenAmount) < req.Amount {
```

**正确语义**（需与 DB 对齐二选一）：

| 方案 | Balance 含义 | available 计算 | 说明 |
|---|---|---|---|
| A（推荐） | 总余额 | `Balance - FrozenAmount` | FrozenAmount 是 Balance 的子集 |
| B | 可用余额 | `Balance` | FrozenAmount 是独立字段 |

当前代码按**方案 A**计算但注释写的是"可用余额"，需统一。

### BUG-3：GetTodayWithdrawCount 不区分状态（逻辑级 🟡）

```go
// withdrawRepo.go:20 — 只过滤了日期，没过滤状态
Where("driver_id = ? AND DATE(apply_time) = ?", driverID, today)
```

**问题**：失败的提现也会计入每日次数，导致司机被错误限流。应只统计 `status IN (1,2)`（处理中+成功）。

### BUG-4：允许信用卡提现（业务级 🟡）

`DriverBankCard.CardType = 2`（信用卡）可绑定且可提现，但国内信用卡不能接收转账，打款必失败。应在绑卡或提现时拒绝信用卡。

---

## 三、边界值分析（等价类划分）

### 3.1 提现金额边界

| 等价类 | 输入（分） | 预期结果 |
|---|---|---|
| 零值 | 0 | ❌ `ErrInvalidParam` |
| 负值 | -100 | ❌ `ErrInvalidParam` |
| 低于最低限额 | 99（0.99元） | ❌ `ErrWithdrawMinAmount` |
| **最低限额** | **100（1元）** | **✅ 通过** |
| 正常范围 | 50000（500元） | ✅ 通过 |
| **单笔上限** | **500000（5000元）** | **✅ 通过** |
| 超过单笔上限 | 500001 | ❌ `ErrWithdrawAmountLimit` |
| 超过可提现余额 | Balance=30000, Amount=40000 | ❌ `ErrInsufficientBalance` |
| 等于可提现余额 | Balance=30000, Amount=30000 | ✅ 全部提现 |
| 非整数分 | 101（1.01元）| ⚠️ 应允许（分是最小单位，101分合法） |

### 3.2 每日次数边界

| 等价类 | 今日已提次数 | 预期 |
|---|---|---|
| 未提现 | 0 | ✅ 可提 |
| 接近上限 | 2 | ✅ 可提（第3次） |
| **达到上限** | **3** | **❌ `WITHDRAW_COUNT_LIMIT`** |
| 超过上限（异常数据） | 4 | ❌ `WITHDRAW_COUNT_LIMIT` |

### 3.3 资格校验优先级

当多个条件同时不满足时，按以下顺序返回第一个失败原因：

| 优先级 | 条件 | 禁用原因码 |
|---|---|---|
| 1 | 未绑定银行卡 | `NO_BANK_CARD` |
| 2 | 未通过实名认证 | `VERIFY_PENDING` |
| 3 | 今日提现次数超限 | `WITHDRAW_COUNT_LIMIT` |
| 4 | 可提现余额为零 | `AVAILABLE_AMOUNT_ZERO` |

### 3.4 银行卡与实名关联

| 场景 | 预期 |
|---|---|
| 有银行卡 + 实名通过 | ✅ 可提现 |
| 有银行卡 + 实名未提交 | ❌ `VERIFY_PENDING` |
| 有银行卡 + 实名审核中 | ❌ `VERIFY_PENDING` |
| 有银行卡 + 实名被拒 | ❌ `VERIFY_PENDING` |
| 无银行卡 + 实名通过 | ❌ `NO_BANK_CARD`（优先级1） |
| 银行卡持卡人 ≠ 实名姓名 | ⚠️ 当前未校验（**缺口**） |

---

## 四、并发与幂等边界

### 4.1 并发提现竞态

| 场景 | 当前防护 | 风险 |
|---|---|---|
| 同一司机同时发起2笔提现 | 乐观锁（version） | ✅ 安全（第二笔会 version conflict） |
| 提现页查询与提交的间隙余额变化 | 无 | ⚠️ 页面显示可用但提交时可能余额不足（可接受，提交时再校验） |
| 事务内重读钱包 | ✅ `txRepo.GetWallet` | 防止事务外变更 |

### 4.2 缺失的幂等保护

| 问题 | 说明 |
|---|---|
| 无提现单号去重 | 同一司机短时间重复提交会生成多条记录，每条扣一次余额 |
| 建议 | 基于金额+时间窗口（如30秒内相同金额）做幂等，或前端传 idempotent_key |

---

## 五、花小猪对标缺口分析

当前项目规则已足够简化，以下为行业常见但本项目暂未实现的规则（标记优先级）：

| 缺口 | 花小猪实践 | 优先级 | 建议 |
|---|---|---|---|
| 银行卡姓名一致性校验 | 持卡人必须=实名姓名 | **P0** | 在 `evaluateWithdrawEligibility` 中新增校验 |
| 信用卡拒绝 | 只允许借记卡 | **P0** | 绑卡/提现时检查 `CardType == 1` |
| 每周累计限额 | 周累计15000元 | P2 | 暂不需要，当前项目3次/天已够用 |
| 提现时间窗口 | 周二/四 9:00-22:00 | P2 | 暂不需要，24h可提体验更好 |
| 提现失败退回 | 失败后自动退回余额 | **P1** | 需提现回调接口，当前模拟打款无此逻辑 |
| 手续费 | >10元收2% | P3 | 当前免费策略更优 |

---

## 六、推荐金额预设（分）

当前 `computeSuggestedAmounts` 返回的是元，需改为分：

```go
// ✅ 修正后（单位：分）
var presets = []int64{5000, 10000, 20000, 50000, 100000, 200000}
// 对应：50元, 100元, 200元, 500元, 1000元, 2000元
```

当可用余额 > 200000分(2000元)时，追加"全部提现"选项。

---

## 七、提现须知文案修正

```go
var WithdrawNotice = []string{
    "提现将转入您的银行卡，预计2小时到账",
    "每日最多提现3次，单笔上限5000元，最低1元起提",  // 补充最低金额
    "提现前请确认银行卡信息正确",
    "仅支持提现至本人借记卡",  // 新增：信用卡/非本人卡提醒
}
```

---

## 八、完整测试场景矩阵

### 8.1 GetWithdrawPage 正向

| # | 场景 | 前置条件 | 预期关键字段 |
|---|---|---|---|
| WP-01 | 正常可提现 | 有卡+实名通过+余额>0+次数<3 | `can_withdraw=true` |
| WP-02 | 无银行卡 | 无卡 | `can_withdraw=false, reason=NO_BANK_CARD` |
| WP-03 | 实名未通过 | 实名status≠2 | `can_withdraw=false, reason=VERIFY_PENDING` |
| WP-04 | 今日次数已满 | today_count=3 | `can_withdraw=false, reason=WITHDRAW_COUNT_LIMIT` |
| WP-05 | 余额为零 | Balance=0 | `can_withdraw=false, reason=AVAILABLE_AMOUNT_ZERO` |
| WP-06 | 冻结>可用 | FrozenAmount>Balance | `available_withdraw_amount=0` |
| WP-07 | 钱包不存在 | 新司机 | 自动创建钱包，余额=0 |

### 8.2 ApplyWithdraw 正向

| # | 场景 | 输入 | 预期 |
|---|---|---|---|
| AW-01 | 最低金额提现 | amount=100(1元) | ✅ 成功 |
| AW-02 | 单笔上限提现 | amount=500000(5000元) | ✅ 成功 |
| AW-03 | 全部提现 | amount=available | ✅ 成功，余额变0 |
| AW-04 | 部分提现 | amount<available | ✅ 成功，余额减少 |

### 8.3 ApplyWithdraw 反向

| # | 场景 | 输入 | 预期错误码 |
|---|---|---|---|
| AW-10 | 无效driver_id | driver_id=0 | `ErrInvalidDriverID` |
| AW-11 | 零金额 | amount=0 | `ErrInvalidParam` |
| AW-12 | 负金额 | amount=-100 | `ErrInvalidParam` |
| AW-13 | 低于最低限额 | amount=99 | `ErrWithdrawMinAmount` |
| AW-14 | 超过单笔上限 | amount=500001 | `ErrWithdrawAmountLimit` |
| AW-15 | 余额不足 | amount>Balance | `ErrInsufficientBalance` |
| AW-16 | 冻结限制 | amount>(Balance-Frozen) | `ErrFrozenAmountLimit` |
| AW-17 | 今日次数已满 | today_count>=3 | `ErrWithdrawPageUnavailable` |
| AW-18 | 无银行卡 | 无卡 | `ErrWithdrawPageUnavailable` |
| AW-19 | 未实名 | 实名未通过 | `ErrWithdrawPageUnavailable` |

### 8.4 并发与竞态

| # | 场景 | 预期 |
|---|---|---|
| CC-01 | 同一司机并发2笔5000元，余额10000 | 一笔成功，一笔 version conflict |
| CC-02 | 提现页查询后余额变化再提交 | 提交时重新校验，可能 `ErrInsufficientBalance` |

### 8.5 提现记录与流水

| # | 场景 | 预期 |
|---|---|---|
| RL-01 | 提现成功后查记录 | 状态=处理中(1)，单号格式 WD20260502xxxx |
| RL-02 | 流水记录正确 | BalanceBefore - Amount = BalanceAfter |
| RL-03 | 余额扣减正确 | wallet.Balance -= Amount, wallet.TotalWithdraw += Amount |

---

## 九、修复优先级排序

| 优先级 | Bug/缺口 | 修复工作量 |
|---|---|---|
| **P0** | BUG-1：金额单位 float64→int64 | 中（walletService.go 全面修改） |
| **P0** | 银行卡姓名一致性校验 | 小（加一个比较） |
| **P0** | 信用卡拒绝 | 小（加 CardType 校验） |
| **P1** | BUG-2：T-3 冻结语义统一 | 小（统一注释+逻辑） |
| **P1** | BUG-3：提现次数只统计有效状态 | 小（加 status 过滤） |
| **P1** | 幂等保护 | 中（idempotent_key 方案） |
| **P2** | 提现失败退回机制 | 大（需回调接口） |

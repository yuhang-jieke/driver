# 司机端提现产品规则设计

## 1. 背景

当前司机端提现能力为简化版实现，仅提供：

- `GetWallet`：返回钱包概览
- `ApplyWithdraw`：按基础余额与绑卡规则发起提现
- `GetWithdrawRecords`：查询提现记录

现状问题：

1. 提现页所需的产品规则字段缺失，前端无法直接按真实产品方式展示
2. 提现申请规则较薄，仅覆盖余额、绑卡、单日次数，缺少更细的禁提原因与状态语义
3. 当前提现状态仅有 `处理中/成功/失败`，无法表达“预计 2 小时到账”的真实产品体验
4. `ApplyWithdraw` 当前为“先落单再扣余额”的非事务实现，规则与数据一致性都有隐患
5. `driver.proto` 尚未定义“提现页查询接口”，无法把产品规则集中在服务端统一输出

本次目标不是接入真实代付渠道，而是将司机端提现产品规则、提现页展示能力、提现申请规则与记录状态设计成接近真实网约车司机端的体验。

## 2. 目标与范围

### 2.1 目标

- 提供可直接驱动提现页渲染的后端接口
- 将“是否能提现”判断统一收口到服务端
- 将提现申请规则做成稳定、可扩展、可测试的业务能力
- 保持当前技术栈下最小必要改造，不引入真实支付通道

### 2.2 本次范围

包含：

- 提现页查询接口设计
- 提现申请接口规则升级
- 提现记录状态表达升级
- BFF / proto / srvDriver / repository 的实现边界调整
- 错误码、日志、测试方案

不包含：

- 第三方代付渠道接入
- 实际银行到账回调
- 风控系统联调
- 金额字段从 `float64` 迁移到“分”的彻底改造

## 3. 方案对比

### 方案 A：在现有 `GetWallet` 上继续堆字段

做法：

- 继续使用 `GetWallet` 作为钱包页与提现页统一入口
- 在现有返回上增加提现页相关字段

优点：

- 改动小
- 前端接入快

缺点：

- 钱包概览与提现页规则语义混杂
- 后续继续扩展手续费、禁提原因、到账说明时会越来越重
- 违背“一个接口一个清晰职责”

### 方案 B：新增独立“提现页查询接口”

做法：

- 保留 `GetWallet` 负责钱包概览
- 新增 `GetWithdrawPage` 专门服务提现页
- `ApplyWithdraw` 只负责发起提现，不承担页面规则拼装

优点：

- 职责清晰
- 前后端边界稳定
- 更接近真实产品分层
- 便于后续扩展手续费、风控、活动规则

缺点：

- 需要新增 proto、BFF、service 与测试

### 方案 C：直接做资金域整体重构

做法：

- 一次性重构钱包、提现、银行卡、流水、资格校验所有边界

优点：

- 架构最完整

缺点：

- 范围过大
- 会拖慢当前交付
- 对现有代码影响面过宽

### 结论

推荐采用 **方案 B：新增独立“提现页查询接口”**。

原因：

1. 既能满足“提现页规则”和“提现动作规则”同时产品化
2. 改动可控，不需要一次性重构整个钱包域
3. 能自然承接后续真实代付与风控能力

## 4. 业务规则设计

### 4.1 提现产品规则

本次提现统一采用以下产品语义：

- 预计到账时间：`2 小时到账`
- 司机必须已绑卡
- 司机必须完成实名相关前置校验
- 提现按“可提现金额”而不是“钱包总余额”判断
- 单日最多提现 `3` 次
- 单笔提现最小金额 `1` 元
- 单笔提现最大金额 `5000` 元
- 默认不收手续费，保留手续费字段与规则扩展位

### 4.2 可提现金额口径

当前阶段沿用现有钱包模型：

- `balance`：当前可用余额
- `frozen_amount`：冻结金额

提现页展示口径：

- `wallet_balance`：钱包显示余额
- `frozen_amount`：暂不可提现金额
- `available_withdraw_amount`：本次实际可提金额

计算规则：

- 当前模型下 `available_withdraw_amount = balance`
- 若后续引入更细的在途冻结、风控冻结，可继续在此字段上收口，不影响前端

### 4.3 提现前置校验顺序

服务端按以下顺序校验：

1. 司机 ID 合法
2. 已实名且状态允许提现
3. 已绑定银行卡
4. 单日提现次数未超限
5. 提现金额格式合法
6. 金额达到最小提现门槛
7. 金额未超过单笔上限
8. 金额未超过可提现金额
9. 若存在额外冻结规则，校验冻结限制

这样可以保证前端和后端禁提原因一致，且优先返回最可理解的失败原因。

### 4.4 提现页禁提原因

提现页统一输出以下禁提原因码：

- `NO_BANK_CARD`
- `REALNAME_REQUIRED`
- `VERIFY_PENDING`
- `WITHDRAW_COUNT_LIMIT`
- `AVAILABLE_AMOUNT_ZERO`
- `BELOW_MIN_AMOUNT`
- `SYSTEM_RESTRICTED`

同时返回对应文案，BFF 可直接透传给前端展示。

### 4.5 到账时效规则

产品固定展示：

- `预计 2 小时到账`

记录状态说明：

- 申请成功后进入“处理中”
- 系统异步任务在 2 小时窗口内推进为“到账成功”或“到账失败”
- 若失败，需要有失败原因，并支持司机重新发起提现

## 5. 接口设计

## 5.1 新增 RPC：`GetWithdrawPage`

职责：

- 为提现页提供完整规则信息
- 前端无需再从 `GetWallet`、`GetBankCard`、`GetProfile` 拼装提现能力

建议 proto：

```proto
message GetWithdrawPageReq {
  int64 driver_id = 1;
}

message WithdrawRuleInfo {
  double min_withdraw_amount = 1;
  double max_withdraw_amount = 2;
  int32 today_withdraw_count = 3;
  int32 today_withdraw_limit = 4;
  string estimated_arrival_text = 5;
  bool fee_free = 6;
  double fee_amount = 7;
  string fee_desc = 8;
}

message WithdrawPageBankCard {
  bool has_bank_card = 1;
  string bank_name = 2;
  string bank_card_no = 3;
}

message WithdrawPageActionState {
  bool can_withdraw = 1;
  string disable_reason_code = 2;
  string disable_reason_text = 3;
}

message GetWithdrawPageResp {
  double wallet_balance = 1;
  double frozen_amount = 2;
  double available_withdraw_amount = 3;
  WithdrawRuleInfo rule_info = 4;
  WithdrawPageBankCard bank_card = 5;
  WithdrawPageActionState action_state = 6;
  repeated double suggested_amounts = 7;
  repeated string withdraw_notice = 8;
}
```

### 5.2 升级 RPC：`ApplyWithdraw`

保留现有职责，但补充规则：

- 以 `GetWithdrawPage` 同一口径校验提现资格
- 创建提现单时写入产品化状态与预计到账时间
- 使用事务更新提现记录与钱包余额
- 失败时返回稳定错误码

建议响应仍保持兼容：

```proto
message ApplyWithdrawResp {
  bool success = 1;
  string message = 2;
  string withdraw_no = 3;
}
```

如需要增强体验，可在后续增加：

- `status`
- `estimated_arrival_text`
- `actual_amount`

### 5.3 升级 RPC：`GetWithdrawRecords`

目标：

- 更贴合“真实产品记录页”
- 返回预计到账与失败原因等展示语义

建议在现有基础上补充以下字段：

- `status_text`
- `estimated_arrival_text`
- `processing_deadline`

若本次希望控制改动面，可先不改 proto，只在现有 `status` 与 `fail_reason` 基础上由 BFF 做有限展示。

## 6. 状态流转设计

### 6.1 提现记录状态

当前状态：

- `1`: 处理中
- `2`: 成功
- `3`: 失败

本次保留该状态枚举，避免大面积兼容风险，但补充清晰语义：

- `处理中`：申请已受理，预计 2 小时到账
- `成功`：已到账
- `失败`：未到账，余额已回退或未扣减完成

### 6.2 状态流转

```text
申请提现
  -> 处理中
     -> 成功
     -> 失败
```

### 6.3 失败处理

若后续异步任务判定失败：

1. 更新提现记录为失败
2. 回补钱包余额
3. 写入失败流水
4. 记录失败原因

本次即使不做真实异步渠道，也应把状态设计成支持该流程。

## 7. 架构与组件改动

### 7.1 `common/idl/driver.proto`

新增：

- `GetWithdrawPageReq`
- `GetWithdrawPageResp`
- 提现页相关 message
- `DriverService.GetWithdrawPage` RPC

### 7.2 `bffDriver`

新增或调整：

- 在 HTTP Handler 中新增提现页接口
- 请求仅接收当前登录司机身份，不再依赖页面自己推导资格
- 将 gRPC 响应转为前端 JSON 响应

### 7.3 `srvDriver/handler`

新增：

- `GetWithdrawPage`

调整：

- `ApplyWithdraw` 中统一调用提现规则校验逻辑

### 7.4 `srvDriver/service`

新增私有能力：

- `buildWithdrawPage`
- `evaluateWithdrawEligibility`
- `validateWithdrawApply`

调整：

- `ApplyWithdraw` 改造成事务式实现
- 统一提现页与提现申请的规则口径

### 7.5 `srvDriver/repository`

新增或调整：

- 获取实名状态方法
- 获取提现页所需银行卡信息方法
- 事务内更新钱包与创建提现记录
- 预留“更新提现记录状态”方法，便于后续异步任务推进

## 8. 数据流设计

### 8.1 提现页查询

1. BFF 接收提现页请求
2. 调用 `GetWithdrawPage`
3. srvDriver 查询钱包、绑卡、实名、今日提现次数
4. service 计算规则、禁提原因、建议金额
5. 返回页面可直接渲染的数据

### 8.2 提现申请

1. BFF 传入 `driver_id` 与提现金额
2. srvDriver 统一执行资格校验
3. 事务内创建提现记录并扣减钱包
4. 写钱包流水
5. 返回提现单号和成功结果

## 9. 错误处理

建议新增或明确以下错误码：

- `ErrWithdrawRealnameRequired`
- `ErrWithdrawVerifyPending`
- `ErrWithdrawMinAmount`
- `ErrWithdrawMaxAmount`
- `ErrWithdrawCountLimit`
- `ErrWithdrawAmountExceedAvailable`
- `ErrWithdrawCreateRecordFailed`
- `ErrWithdrawWalletConflict`

要求：

- 提现页禁用原因与提现申请失败错误码语义一致
- 前端展示文案由后端稳定输出

## 10. 日志与审计

关键日志：

- 查询提现页：司机 ID、可提现金额、禁提原因
- 发起提现：司机 ID、金额、银行卡尾号、提现单号、结果
- 提现失败：失败原因、是否已回退余额

日志注意：

- 禁止打印银行卡全号
- 金额、状态、提现单号必须可追踪

## 11. 测试策略

### 11.1 Service 单测

至少覆盖：

1. 未绑卡不能提现
2. 未实名不能提现
3. 审核中不能提现
4. 可提现金额为 0
5. 单日次数超限
6. 小于最小提现金额
7. 大于单笔上限
8. 余额充足且申请成功

### 11.2 Handler / RPC 测试

- 参数缺失
- 非法金额
- 正常返回提现页
- 正常发起提现

### 11.3 回归重点

- 不能影响现有 `GetWallet`
- 不能破坏提现记录列表查询
- 保持现有 proto 调用兼容性

## 12. 实施建议

建议按以下顺序落地：

1. 扩展 `driver.proto`
2. 生成 gRPC 代码
3. 实现 `GetWithdrawPage`
4. 重构 `ApplyWithdraw` 为统一规则 + 事务实现
5. 补充 BFF 提现页接口
6. 增加单测

## 13. 风险与约束

### 风险

- 当前金额仍为 `float64`，存在精度风险
- 当前缺少真实支付回调，`2 小时到账` 只能先作为产品语义与状态设计
- 若实名状态数据来源分散，需要补充 repository 收口

### 应对

- 本次不改金额模型，但避免扩大 `float64` 使用范围
- 提前预留异步状态推进接口
- 将提现资格判断统一放在 service 层

## 14. 结论

本次提现能力应采用“提现页查询接口 + 提现申请接口”双接口模式：

- `GetWithdrawPage` 负责页面规则与资格展示
- `ApplyWithdraw` 负责提现动作与数据变更

这能在现有架构下，以最小必要改动实现更接近真实网约车司机端的提现产品能力，并为后续真实打款通道、异步状态回调、手续费与风控扩展留下清晰边界。

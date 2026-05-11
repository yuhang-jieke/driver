# 钱包统计数据按日期范围查询功能设计

## 1. 需求背景

### 1.1 当前问题
现有 `GetWallet` API 的统计数据（今日/本周/本月收入）使用固定时间范围计算：
- 今日：当天 00:00:00 到当前时间
- 本周：本周一 00:00:00 到当前时间
- 本月：本月 1 日 00:00:00 到当前时间

这导致在查询历史数据时，无法灵活指定日期范围。例如：
- 用户在 5 月 4 日想查询"5 月 1 日～5 月 3 日"的数据，当前 API 无法满足
- 用户想查询"4 月全月"的数据，当前 API 只能查询"往前推 30 天"

### 1.2 目标用户
司机端用户，通过前端日期选择器查询指定日期范围内的收入统计数据。

### 1.3 核心需求
- 支持用户自定义日期范围查询（startDate, endDate）
- 保持向后兼容：未传日期参数时，保持原有逻辑
- 日期范围限制：最多查询 90 天
- 严格按用户选择的日期范围查询，不做时间偏移计算

---

## 2. 技术方案

### 2.1 方案概述

**修改现有 `GetWallet` API**，增加可选的日期范围参数：
- 如果传了 `start_date` 和 `end_date`，使用自定义范围查询
- 如果未传，保持原有逻辑（今日/本周/本月）

**优点**：
- 向后兼容，不影响现有调用方
- 复用现有代码结构，改动最小
- 前端可以灵活选择使用固定周期或自定义范围

**缺点**：
- API 语义略显复杂（既支持固定周期，又支持自定义范围）

### 2.2 Proto 定义修改

**文件**：`taketaxi/common/idl/driver.proto`

```protobuf
message GetWalletReq {
  int64 driver_id = 1;
  string start_date = 2;  // 可选，格式 YYYY-MM-DD，查询开始日期
  string end_date = 3;    // 可选，格式 YYYY-MM-DD，查询结束日期
}

message GetWalletResp {
  int64 balance = 1;             // 可提现余额(单位:分)
  int64 frozen_amount = 2;       // 冻结金额(未结算/在途)(单位:分)
  int64 today_income = 3;        // 今日收入(单位:分) - 如果传了日期范围，此字段为范围内收入
  int64 week_income = 4;         // 本周收入(单位:分) - 如果传了日期范围，此字段为 0
  int64 month_income = 5;        // 本月收入(单位:分) - 如果传了日期范围，此字段为 0
  int64 total_income = 6;        // 累计总收入(单位:分)
  int64 total_withdraw = 7;      // 累计提现金额(单位:分)
  int32 today_withdraw_count = 8;// 今日已提现次数(最多3次)
  string bank_card_no = 9;       // 已绑定银行卡(脱敏)
  bool has_bank_card = 10;       // 是否已绑定银行卡
  string query_start_date = 11;  // 实际查询的开始日期（回显用）
  string query_end_date = 12;    // 实际查询的结束日期（回显用）
}
```

**字段说明**：
- `start_date` / `end_date`：可选参数，格式 `YYYY-MM-DD`
- `query_start_date` / `query_end_date`：返回实际查询的日期范围，方便前端回显
- 当传了日期范围时：
  - `today_income` 字段复用为"范围内收入"
  - `week_income` 和 `month_income` 字段设为 0（避免混淆）

### 2.3 Service 层修改

**文件**：`taketaxi/srvDriver/internal/service/walletService.go`

**修改 `GetWallet` 方法**：

```go
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
		return nil, errcode.NewWithDetail(errcode.ErrInvalidParam, "end_date must be after start_date")
	}

	// 3. 最多查询 90 天
	daysDiff := int(endDate.Sub(startDate).Hours() / 24)
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
	weekStart := todayStart.AddDate(0, 0, -int(now.Weekday()))
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
```

### 2.4 Repository 层修改

**无需修改**。现有的 `GetOrderStats` 方法已经支持任意日期范围查询：

```go
// GetOrderStats 查询指定日期范围内的接单统计数据
func (r *DriverRepo) GetOrderStats(ctx context.Context, driverID int64, startDate, endDate time.Time) (*model.OrderStatsResult, error) {
	var result model.OrderStatsResult
	err := r.db.WithContext(ctx).
		Model(&model.DriverStatisticsSummary{}).
		Select("COALESCE(SUM(order_count), 0) as order_count, COALESCE(SUM(total_income), 0) as total_income, COALESCE(SUM(online_duration), 0) as online_duration").
		Where("driver_id = ? AND stat_date BETWEEN ? AND ?", driverID, startDate, endDate).
		Scan(&result).Error
	if err != nil {
		return nil, err
	}
	return &result, nil
}
```

**关键点**：
- `stat_date BETWEEN ? AND ?` 会严格按传入的时间范围查询
- Service 层已经将 `startDate` 补充为 `00:00:00`，`endDate` 补充为 `23:59:59`
- 因此 SQL 查询会覆盖完整的日期范围

### 2.5 BFF 层修改

**文件**：`taketaxi/bffDriver/internal/handler/walletHandler.go`

**修改 HTTP 接口**，支持可选的查询参数：

```go
// GetWallet 查询钱包概览
// GET /api/v1/driver/wallet?start_date=2026-05-01&end_date=2026-05-31
func (h *WalletHandler) GetWallet(c *gin.Context) {
	driverID := c.GetInt64("driver_id") // 从 JWT 中获取

	// 可选参数：日期范围
	startDate := c.Query("start_date") // 格式 YYYY-MM-DD
	endDate := c.Query("end_date")

	req := &driver.GetWalletReq{
		DriverId:  driverID,
		StartDate: startDate,
		EndDate:   endDate,
	}

	resp, err := h.driverClient.GetWallet(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}
```

### 2.6 前端调用示例

**固定周期模式（原有逻辑）**：
```typescript
// 查询今日/本周/本月收入
const wallet = await fetch('http://localhost:8080/api/v1/driver/wallet');
// 返回：today_income, week_income, month_income
```

**自定义范围模式（新功能）**：
```typescript
// 查询 5 月 1 日～5 月 31 日的收入
const wallet = await fetch(
  'http://localhost:8080/api/v1/driver/wallet?start_date=2026-05-01&end_date=2026-05-31'
);
// 返回：today_income（实际为范围内收入），week_income=0, month_income=0
```

---

## 3. 错误码设计

复用现有错误码：

| 错误码 | 说明 | 触发条件 |
|--------|------|---------|
| `ErrInvalidParam` (40001) | 参数错误 | 日期格式错误、结束日期早于开始日期、范围超过 90 天 |
| `ErrInvalidDriverID` (40002) | 司机 ID 无效 | driver_id <= 0 |

---

## 4. 数据库影响

**无需修改数据库表结构**。

现有表 `driver_statistics_summary` 已满足需求：
- `stat_date` 字段存储统计日期（datetime 类型）
- 查询时使用 `BETWEEN` 条件即可

---

## 5. 测试计划

### 5.1 单元测试

**Service 层测试**：
```go
func TestGetWallet_WithDateRange(t *testing.T) {
	// 测试用例 1：正常查询 30 天范围
	req := &driver.GetWalletReq{
		DriverId:  200000001,
		StartDate: "2026-05-01",
		EndDate:   "2026-05-31",
	}
	resp, err := service.GetWallet(ctx, req)
	assert.NoError(t, err)
	assert.Equal(t, "2026-05-01", resp.QueryStartDate)
	assert.Equal(t, "2026-05-31", resp.QueryEndDate)

	// 测试用例 2：日期格式错误
	req.StartDate = "2026/05/01" // 错误格式
	_, err = service.GetWallet(ctx, req)
	assert.Error(t, err)

	// 测试用例 3：范围超过 90 天
	req.StartDate = "2026-01-01"
	req.EndDate = "2026-05-01" // 121 天
	_, err = service.GetWallet(ctx, req)
	assert.Error(t, err)

	// 测试用例 4：结束日期早于开始日期
	req.StartDate = "2026-05-31"
	req.EndDate = "2026-05-01"
	_, err = service.GetWallet(ctx, req)
	assert.Error(t, err)
}

func TestGetWallet_WithoutDateRange(t *testing.T) {
	// 测试用例 5：不传日期参数，保持原有逻辑
	req := &driver.GetWalletReq{
		DriverId: 200000001,
	}
	resp, err := service.GetWallet(ctx, req)
	assert.NoError(t, err)
	assert.NotZero(t, resp.TodayIncome)
	assert.NotZero(t, resp.WeekIncome)
	assert.NotZero(t, resp.MonthIncome)
}
```

### 5.2 集成测试

**HTTP API 测试**：
```bash
# 测试 1：查询 5 月全月数据
curl "http://localhost:8080/api/v1/driver/wallet?start_date=2026-05-01&end_date=2026-05-31"

# 测试 2：不传日期参数（原有逻辑）
curl "http://localhost:8080/api/v1/driver/wallet"

# 测试 3：日期格式错误
curl "http://localhost:8080/api/v1/driver/wallet?start_date=2026/05/01&end_date=2026/05/31"
# 预期：返回 400 错误

# 测试 4：范围超过 90 天
curl "http://localhost:8080/api/v1/driver/wallet?start_date=2026-01-01&end_date=2026-05-01"
# 预期：返回 400 错误
```

### 5.3 边界测试

| 测试场景 | 输入 | 预期输出 |
|---------|------|---------|
| 查询单日数据 | start_date=2026-05-01, end_date=2026-05-01 | 返回 5 月 1 日当天数据 |
| 查询 90 天边界 | start_date=2026-02-01, end_date=2026-05-01 | 正常返回（89 天） |
| 查询 91 天 | start_date=2026-01-31, end_date=2026-05-01 | 返回错误 |
| 跨月查询 | start_date=2026-04-25, end_date=2026-05-05 | 正常返回 |
| 跨年查询 | start_date=2025-12-25, end_date=2026-01-05 | 正常返回 |

---

## 6. 上线计划

### 6.1 发布步骤

1. **修改 Proto 定义**：
   - 编辑 `taketaxi/common/idl/driver.proto`
   - 运行 `./taketaxi/scripts/gen_proto.sh` 生成代码

2. **修改 Service 层**：
   - 编辑 `taketaxi/srvDriver/internal/service/walletService.go`
   - 添加 `getWalletWithDateRange` 和 `getWalletWithFixedPeriod` 方法

3. **修改 BFF 层**：
   - 编辑 `taketaxi/bffDriver/internal/handler/walletHandler.go`
   - 支持 `start_date` 和 `end_date` 查询参数

4. **运行测试**：
   - 单元测试：`go test ./taketaxi/srvDriver/internal/service/...`
   - 集成测试：启动服务，使用 curl 测试

5. **部署**：
   - 先部署 srvDriver（gRPC 服务）
   - 再部署 bffDriver（HTTP 服务）

### 6.2 回滚方案

如果发现问题，可以快速回滚：
- Proto 定义向后兼容（新增字段为可选）
- Service 层逻辑向后兼容（未传日期参数时保持原有逻辑）
- 回滚只需重新部署旧版本代码即可

---

## 7. 后续优化

### 7.1 性能优化

如果 `driver_statistics_summary` 表数据量较大，建议：
- 在 `stat_date` 字段上添加索引（如果尚未添加）
- 考虑分区表（按月分区）

### 7.2 功能扩展

未来可以考虑：
- 支持更多统计维度（订单数、在线时长等）
- 支持导出 Excel 报表
- 支持按周/月聚合查询（GROUP BY WEEK/MONTH）

---

## 8. 附录

### 8.1 相关文件清单

| 文件路径 | 修改内容 |
|---------|---------|
| `taketaxi/common/idl/driver.proto` | 修改 `GetWalletReq` 和 `GetWalletResp` 定义 |
| `taketaxi/srvDriver/internal/service/walletService.go` | 修改 `GetWallet` 方法，新增两个辅助方法 |
| `taketaxi/bffDriver/internal/handler/walletHandler.go` | 修改 HTTP 接口，支持查询参数 |

### 8.2 时间复杂度分析

**SQL 查询**：
```sql
SELECT COALESCE(SUM(order_count), 0) as order_count, 
       COALESCE(SUM(total_income), 0) as total_income, 
       COALESCE(SUM(online_duration), 0) as online_duration 
FROM `driver_statistics_summary` 
WHERE driver_id = ? AND stat_date BETWEEN ? AND ?
```

**复杂度**：
- 如果 `(driver_id, stat_date)` 有联合索引：O(log N + M)，其中 M 为范围内记录数
- 如果只有 `driver_id` 索引：O(log N + K)，其中 K 为该司机所有记录数
- 最坏情况（无索引）：O(N)，全表扫描

**建议**：确保 `(driver_id, stat_date)` 有联合索引。

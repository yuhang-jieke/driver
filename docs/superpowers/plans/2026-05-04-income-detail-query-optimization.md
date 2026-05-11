# 收入明细查询优化 实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将收入明细 API 的 `days` 参数改为 `period` 枚举，实现正确的时间范围计算（今天/本周/本月）。

**Architecture:** Proto 层定义 period 枚举 → Repository 层计算时间范围 → Service 层参数校验 → BFF Handler 解析参数 → 前端 API 调用。

**Tech Stack:** Go (gRPC/GORM), TypeScript (React/Fetch API)

---

## 文件结构

| 文件 | 职责 | 改动类型 |
|------|------|---------|
| `taketaxi/common/idl/driver.proto` | Proto 定义 | 修改字段 |
| `taketaxi/srvDriver/internal/repository/withdrawRepo.go` | 时间范围计算 + 数据查询 | 修改函数签名和逻辑 |
| `taketaxi/srvDriver/internal/service/walletService.go` | 参数校验 | 修改参数处理 |
| `taketaxi/bffDriver/internal/handler/driverHandler.go` | HTTP 参数解析 | 修改参数读取 |
| `driverfrontend/src/app/api/wallet.ts` | 前端 API 调用 | 修改函数签名 |
| `driverfrontend/src/app/components/DriverApp.tsx` | 收入明细组件 | 修改 API 调用（已使用 period，无需改动）|

---

## Task 1: 修改 Proto 定义

**Files:**
- Modify: `taketaxi/common/idl/driver.proto:368-371`

- [ ] **Step 1: 修改 GetIncomeDetailReq 消息**

将 `days` 字段改为 `period` 字段：

```protobuf
// GetIncomeDetail 收入明细(分类汇总)
message GetIncomeDetailReq {
  int64 driver_id = 1;
  string period = 2;             // 查询周期: today/week/month
}
```

- [ ] **Step 2: 重新生成 Proto 代码**

```bash
cd D:/software/GoWork/src/driver
./taketaxi/scripts/gen_proto.sh
```

Expected: 生成新的 `driver.pb.go` 和 `driver_grpc.pb.go`

- [ ] **Step 3: 提交 Proto 变更**

```bash
git add taketaxi/common/idl/driver.proto taketaxi/common/kitexGen/
git commit -m "feat(proto): change GetIncomeDetailReq.days to period enum"
```

---

## Task 2: 修改 Repository 层时间计算逻辑

**Files:**
- Modify: `taketaxi/srvDriver/internal/repository/withdrawRepo.go:87-127`

- [ ] **Step 1: 修改 GetIncomeDetail 函数签名和实现**

将原有的 `days` 参数改为 `period` 参数，并实现正确的时间范围计算：

```go
// GetIncomeDetail 查询收入分类明细（GROUP BY type）
// period: today/week/month
func (r *DriverRepo) GetIncomeDetail(ctx context.Context, driverID int64, period string) ([]model.IncomeDetailResult, error) {
	var results []model.IncomeDetailResult
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)

	var startDate time.Time
	switch period {
	case "today":
		startDate = today
	case "week":
		// 中国习惯：周一为一周开始
		weekday := int(now.Weekday())
		if weekday == 0 {
			weekday = 7 // 周日转为7
		}
		startDate = today.AddDate(0, 0, -weekday+1)
	case "month":
		startDate = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.Local)
	default:
		startDate = today // 默认今天
	}
	endDate := today.Add(24*time.Hour - time.Second)

	typeNames := map[int8]string{
		1: "基础车费", 2: "奖励", 3: "空驶补偿", 4: "高速费", 5: "其他",
	}

	rows, err := r.db.WithContext(ctx).Model(&model.DriverIncomeLog{}).
		Select("type, COALESCE(SUM(amount), 0) as amount, COUNT(*) as count").
		Where("driver_id = ? AND created_at BETWEEN ? AND ? AND amount > 0", driverID, startDate, endDate).
		Group("type").
		Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var typeCode int8
		var amount float64
		var count int
		if err := rows.Scan(&typeCode, &amount, &count); err != nil {
			continue
		}
		name, ok := typeNames[typeCode]
		if !ok {
			name = "其他"
		}
		results = append(results, model.IncomeDetailResult{
			TypeCode: typeCode,
			TypeName: name,
			Amount:   amount,
			Count:    count,
		})
	}
	return results, nil
}
```

- [ ] **Step 2: 提交 Repository 层变更**

```bash
git add taketaxi/srvDriver/internal/repository/withdrawRepo.go
git commit -m "feat(repo): change GetIncomeDetail to use period param for correct time range"
```

---

## Task 3: 修改 Service 层参数校验

**Files:**
- Modify: `taketaxi/srvDriver/internal/service/walletService.go:429-464`

- [ ] **Step 1: 修改 GetIncomeDetail 方法**

将 `days` 参数校验改为 `period` 参数校验：

```go
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
```

- [ ] **Step 2: 提交 Service 层变更**

```bash
git add taketaxi/srvDriver/internal/service/walletService.go
git commit -m "feat(service): change GetIncomeDetail to use period param"
```

---

## Task 4: 修改 BFF Handler 参数解析

**Files:**
- Modify: `taketaxi/bffDriver/internal/handler/driverHandler.go:724-755`

- [ ] **Step 1: 修改 GetIncomeDetail Handler**

将 `days` 参数解析改为 `period` 参数：

```go
// GetIncomeDetail 查询收入分类明细
// GET /api/v1/driver/income/detail?driver_id=200000001&period=today
// 返回按类型分组的收入汇总（订单收入、奖励、空驶补偿等）
func (h *DriverHandler) GetIncomeDetail(c *gin.Context) {
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

	period := c.Query("period")
	if period == "" {
		period = "today"
	}

	resp, err := h.client.GetIncomeDetail(c.Request.Context(), &pb.GetIncomeDetailReq{
		DriverId: driverID,
		Period:   period,
	})
	if err != nil {
		logger.Error("GetIncomeDetail 失败", zap.Int64("driver_id", driverID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}
```

- [ ] **Step 2: 提交 BFF Handler 变更**

```bash
git add taketaxi/bffDriver/internal/handler/driverHandler.go
git commit -m "feat(bff): change GetIncomeDetail API param from days to period"
```

---

## Task 5: 修改前端 API 函数

**Files:**
- Modify: `driverfrontend/src/app/api/wallet.ts:184-205`

- [ ] **Step 1: 修改 getIncomeDetail 函数签名**

```typescript
export async function getIncomeDetail(
  driverId: number,
  period: 'today' | 'week' | 'month'
): Promise<{ items: IncomeDetailItem[]; total_amount: number } | null> {
  try {
    const res = await fetch(
      `${API_BASE}/api/v1/driver/income/detail?driver_id=${driverId}&period=${period}`
    );
    if (!res.ok) return null;
    return await res.json();
  } catch (err) {
    console.error("GetIncomeDetail API failed:", err);
    return null;
  }
}
```

- [ ] **Step 2: 提交前端 API 变更**

```bash
git add driverfrontend/src/app/api/wallet.ts
git commit -m "feat(frontend): change getIncomeDetail param from days to period"
```

---

## Task 6: 验证前端组件调用

**Files:**
- Verify: `driverfrontend/src/app/components/DriverApp.tsx`

- [ ] **Step 1: 确认 DriverIncomeView 已使用 period 参数**

检查 `DriverIncomeView` 组件中的 API 调用是否已使用 `period` 参数：

```typescript
// 第 935 行附近，确认已有代码：
const data = await fetch(
  `http://localhost:8080/api/v1/driver/income?driver_id=${TEST_DRIVER_ID}&period=${period}`
)
```

**注意**: 该组件已使用 `period` 参数，无需修改。但需确认该接口是 `/api/v1/driver/income`（收入趋势）而非 `/api/v1/driver/income/detail`（收入明细分类）。

- [ ] **Step 2: 如果需要调用收入明细分类 API，确认调用方式**

如果组件需要调用 `getIncomeDetail` 获取分类明细，确认使用方式：

```typescript
// 如果需要调用收入明细分类 API
const detailData = await getIncomeDetail(TEST_DRIVER_ID, period);
```

---

## Task 7: 编译验证

- [ ] **Step 1: 编译后端服务**

```bash
cd D:/software/GoWork/src/driver
go build ./taketaxi/srvDriver/cmd/main.go
go build ./taketaxi/bffDriver/cmd/main.go
```

Expected: 编译成功，无错误

- [ ] **Step 2: 编译前端**

```bash
cd D:/software/GoWork/src/driver/driverfrontend
npm run build
```

Expected: 编译成功，无 TypeScript 错误

---

## Task 8: 最终提交

- [ ] **Step 1: 确认所有变更**

```bash
git status
git log --oneline -5
```

- [ ] **Step 2: 推送变更（如需要）**

```bash
git push origin qh
```

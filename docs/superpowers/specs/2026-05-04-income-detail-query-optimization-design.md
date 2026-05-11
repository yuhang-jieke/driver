---
name: 收入明细查询优化
description: 将 days 参数改为 period 枚举，实现正确的时间范围计算
type: project
---

# 收入明细查询优化设计

## 背景

当前收入明细 API 使用 `days` 参数（1/7/30），存在以下问题：
- "本周"查询的是"往前7天"，而非"周一到今天"
- "本月"查询的是"往前30天"，而非"当月1日到今天"

## 需求

| period | 时间范围 |
|--------|---------|
| today | 今天 00:00:00 ~ 今天 23:59:59 |
| week | 本周一 00:00:00 ~ 今天 23:59:59（周一为一周开始） |
| month | 当月1日 00:00:00 ~ 今天 23:59:59（仅当月，不能查上个月） |

## 改动方案

### 1. Proto 定义修改

**文件**: `taketaxi/common/idl/driver.proto`

```diff
  // GetIncomeDetail 收入明细(分类汇总)
  message GetIncomeDetailReq {
    int64 driver_id = 1;
-   int32 days = 2;                // 查询天数: 1-今日 7-本周 30-本月
+   string period = 2;             // 查询周期: today/week/month
  }
```

### 2. Repository 层修改

**文件**: `taketaxi/srvDriver/internal/repository/withdrawRepo.go`

修改 `GetIncomeDetail` 函数签名和实现：

```go
// GetIncomeDetail 查询收入分类明细（GROUP BY type）
// period: today/week/month
func (r *DriverRepo) GetIncomeDetail(ctx context.Context, driverID int64, period string) ([]model.IncomeDetailResult, error) {
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

    // ... 查询逻辑不变
}
```

### 3. Service 层修改

**文件**: `taketaxi/srvDriver/internal/service/walletService.go`

```diff
  func (s *DriverService) GetIncomeDetail(ctx context.Context, req *driver.GetIncomeDetailReq) (*driver.GetIncomeDetailResp, error) {
      if req.DriverId <= 0 {
          return nil, errcode.New(errcode.ErrInvalidDriverID)
      }

-     days := int(req.Days)
-     if days <= 0 {
-         days = 1
-     }
-     if days > 30 {
-         days = 30
-     }
+     period := req.Period
+     if period == "" {
+         period = "today"
+     }
+     if period != "today" && period != "week" && period != "month" {
+         period = "today"
+     }

-     results, err := s.repo.GetIncomeDetail(ctx, req.DriverId, days)
+     results, err := s.repo.GetIncomeDetail(ctx, req.DriverId, period)
      // ...
  }
```

### 4. BFF Handler 修改

**文件**: `taketaxi/bffDriver/internal/handler/driverHandler.go`

```diff
  func (h *DriverHandler) GetIncomeDetail(c *gin.Context) {
      driverIDStr := c.Query("driver_id")
      // ...

-     daysStr := c.Query("days")
-     days := int32(1)
-     if daysStr != "" {
-         if d, err := strconv.ParseInt(daysStr, 10, 32); err == nil {
-             days = int32(d)
-         }
-     }
+     period := c.Query("period")
+     if period == "" {
+         period = "today"
+     }

      resp, err := h.client.GetIncomeDetail(c.Request.Context(), &pb.GetIncomeDetailReq{
          DriverId: driverID,
-         Days:     days,
+         Period:   period,
      })
      // ...
  }
```

### 5. 前端 API 修改

**文件**: `driverfrontend/src/app/api/wallet.ts`

```diff
  export async function getIncomeDetail(
    driverId: number,
-   days: number // 1-今日 7-本周 30-本月
+   period: 'today' | 'week' | 'month'
  ): Promise<{ items: IncomeDetailItem[]; total_amount: number } | null> {
    try {
      const res = await fetch(
-       `${API_BASE}/api/v1/driver/income/detail?driver_id=${driverId}&days=${days}`
+       `${API_BASE}/api/v1/driver/income/detail?driver_id=${driverId}&period=${period}`
      );
      // ...
    }
  }
```

## 时间范围计算示例

假设今天是 **2026年5月9日（周六）**：

| period | startDate | endDate |
|--------|-----------|---------|
| today | 2026-05-09 00:00:00 | 2026-05-09 23:59:59 |
| week | 2026-05-04 00:00:00（周一） | 2026-05-09 23:59:59 |
| month | 2026-05-01 00:00:00 | 2026-05-09 23:59:59 |

假设今天是 **2026年5月3日（周日）**：

| period | startDate | endDate |
|--------|-----------|---------|
| today | 2026-05-03 00:00:00 | 2026-05-03 23:59:59 |
| week | 2026-04-27 00:00:00（周一） | 2026-05-03 23:59:59 |
| month | 2026-05-01 00:00:00 | 2026-05-03 23:59:59 |

## 兼容性

- API 参数从 `days` 改为 `period`，属于**不兼容变更**
- 前后端需同步发布

## 文件改动清单

| 文件 | 改动类型 |
|------|---------|
| `taketaxi/common/idl/driver.proto` | 修改 |
| `taketaxi/srvDriver/internal/repository/withdrawRepo.go` | 修改 |
| `taketaxi/srvDriver/internal/service/walletService.go` | 修改 |
| `taketaxi/bffDriver/internal/handler/driverHandler.go` | 修改 |
| `driverfrontend/src/app/api/wallet.ts` | 修改 |
| `driverfrontend/src/app/components/DriverApp.tsx` | 修改（调用处） |

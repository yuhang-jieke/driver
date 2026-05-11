---
name: Income API 集成设计
description: 后端收入明细 API 与前端对接
type: project
---

# 收入明细 API 集成设计

## 需求

前端收入明细页面需要对接后端，实现：
1. 根据时间段（今日/本周/本月）查询统计数据
2. 显示近7天收入趋势图

## 时间段定义

| 前端 period | 后端 days | 说明 |
|------------|-----------|------|
| today | 1 | 今天 |
| week | 7 | 今天往前7天（含今天） |
| month | 30 | 今天往前30天（含今天） |

## 数据来源

表：`driver_statistics_summary`

| 字段 | 说明 |
|-----|------|
| `stat_date` | 统计日期 |
| `order_count` | 订单数 |
| `online_duration` | 在线时长(秒) |
| `total_income` | 总收入 |

## 后端改动

### 新增接口

**路由**：`GET /api/v1/driver/income`

**请求参数**：
| 参数 | 类型 | 必填 | 说明 |
|-----|------|-----|------|
| driver_id | int64 | 是 | 司机ID |
| period | string | 是 | today/week/month |

**响应结构**：
```json
{
  "summary": {
    "order_count": 8,
    "income": 328.5,
    "online_duration": 22320
  },
  "trend": [
    { "date": "2026-04-21", "income": 289.5 },
    { "date": "2026-04-22", "income": 312.0 },
    { "date": "2026-04-23", "income": 245.8 },
    { "date": "2026-04-24", "income": 356.2 },
    { "date": "2026-04-25", "income": 298.0 },
    { "date": "2026-04-26", "income": 412.5 },
    { "date": "2026-04-27", "income": 328.5 }
  ]
}
```

### 文件改动

1. **新增 proto 定义**：`taketaxi/common/idl/driver.proto`
   - `GetDriverIncomeReq` / `GetDriverIncomeResp` / `DailyIncome`

2. **新增 Repository 方法**：`taketaxi/srvDriver/internal/repository/driverRepo.go`
   - `GetIncomeStats(driverID, startDate, endDate)` - 汇总统计
   - `GetDailyIncome(driverID, startDate, endDate)` - 每日数据

3. **新增 Service 方法**：`taketaxi/srvDriver/internal/service/driverService.go`
   - `GetIncome(ctx, req)` - 业务逻辑

4. **新增 RPC Client 方法**：`taketaxi/bffDriver/internal/rpcClient/driverClient.go`
   - `GetIncome(ctx, req)`

5. **新增 Handler**：`taketaxi/bffDriver/internal/handler/driverHandler.go`
   - `Income(c *gin.Context)`

6. **新增路由**：`taketaxi/bffDriver/internal/router/router.go`
   - `GET /api/v1/driver/income`

## 前端改动

### 新增文件

**`driverfrontend/src/app/api/income.ts`**

```typescript
interface IncomeSummary {
  order_count: number;
  income: number;
  online_duration: number;
}

interface DailyIncome {
  date: string;
  income: number;
}

interface IncomeResponse {
  summary: IncomeSummary;
  trend: DailyIncome[];
}

export function fetchDriverIncome(driverId: number, period: 'today' | 'week' | 'month'): Promise<IncomeResponse | null>
```

### 修改文件

**`driverfrontend/src/app/components/DriverApp.tsx`**

修改 `DriverIncomeView` 组件：
1. 添加 `useEffect` 在 period 变化时调用 API
2. 用 API 返回数据替换 mock 数据
3. 计算时均收入 = income / (online_duration / 3600)

## 数据映射

| 后端字段 | 前端显示 | 转换 |
|---------|---------|------|
| `summary.order_count` | 完单数 | 直接显示 |
| `summary.online_duration` | 出车时长 | 秒 → 小时 (÷3600) |
| `summary.income / hours` | 时均收入 | 前端计算 |
| `summary.income` | 总收入 | 直接显示 |
| `trend[]` | 近7日趋势 | 柱状图渲染 |

## 错误处理

- API 调用失败时保持 mock 数据
- 空数据时显示 0

## 测试验证

1. 启动后端服务
2. 访问前端收入明细页面
3. 切换今日/本周/本月，验证数据变化
4. 检查趋势图是否显示近7天数据

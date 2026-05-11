# 司机个人信息与接单统计查询接口设计

## 概述

新增一个聚合接口，供司机端 APP 首页调用，返回个人信息、今日接单统计和认证状态。

## API 接口设计

### 接口定义

```
GET /api/v1/driver/profile?date=2026-04-26&days=1
```

### 请求参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| date | string | 否 | 查询日期，格式 `YYYY-MM-DD`，默认当天 |
| days | int | 否 | 统计天数，默认1，最大30 |

### 响应结构

```json
{
  "personal_info": {
    "nickname": "张师傅",
    "avatar": "https://xxx.com/avatar.jpg",
    "service_score": 85.5,
    "order_count": 128
  },
  "order_stats": {
    "order_count": 5,
    "income": 150.00,
    "online_duration": 18000
  },
  "verify_status": 2
}
```

### 字段说明

| 字段 | 类型 | 说明 |
|------|------|------|
| personal_info.nickname | string | 司机昵称 |
| personal_info.avatar | string | 头像URL |
| personal_info.service_score | double | 服务评分 |
| personal_info.order_count | int32 | 累计完成订单数 |
| order_stats.order_count | int32 | 统计周期内完成订单数 |
| order_stats.income | double | 统计周期内收入金额 |
| order_stats.online_duration | int32 | 统计周期内在线时长（秒） |
| verify_status | int32 | 认证状态：0-未认证 1-认证中 2-已认证 3-认证失败 |

## 架构设计

### 调用链路

```
bffDriver (HTTP Handler)
    → srvDriver (gRPC Service)
        → Repository (数据库查询)
```

### 数据来源

| 数据 | 数据表 | 说明 |
|------|--------|------|
| 个人信息 | `drivers` | 按 `driver_id` 查询 |
| 认证状态 | `drivers.verify_status` | 同一张表 |
| 接单统计 | `driver_statistics_summary` | 按 `stat_date` 范围查询并汇总 |

### 统计查询逻辑

```sql
SELECT
  SUM(order_count) as order_count,
  SUM(total_income) as income,
  SUM(online_duration) as online_duration
FROM driver_statistics_summary
WHERE driver_id = ?
  AND stat_date BETWEEN ? AND ?
```

## Proto 定义

```protobuf
message GetDriverProfileReq {
  int64 driver_id = 1;
  string date = 2;
  int32 days = 3;
}

message GetDriverProfileResp {
  PersonalInfo personal_info = 1;
  OrderStats order_stats = 2;
  int32 verify_status = 3;
}

message PersonalInfo {
  string nickname = 1;
  string avatar = 2;
  double service_score = 3;
  int32 order_count = 4;
}

message OrderStats {
  int32 order_count = 1;
  double income = 2;
  int32 online_duration = 3;
}
```

在 `DriverService` 中新增 `GetProfile` 方法。

## 错误处理

### 错误码

| 场景 | HTTP 状态码 | 错误信息 |
|------|-------------|----------|
| 司机不存在 | 404 | "driver not found" |
| 日期格式错误 | 400 | "invalid date format" |
| days 参数非法 | 400 | "invalid days parameter" |

### 边界情况

1. **统计表无数据**：返回零值 `{"order_count": 0, "income": 0, "online_duration": 0}`
2. **days 参数过大**：限制最大值 30 天，超出返回 400 错误
3. **date 为未来日期**：返回 400 错误

## 测试策略

### 单元测试

- `DriverService.GetDriverProfile` 方法测试
- 覆盖：正常查询、司机不存在、统计表无数据、日期边界

### 集成测试

- HTTP 接口端到端测试
- 验证完整调用链路

## 文件变更清单

| 文件 | 变更类型 | 说明 |
|------|----------|------|
| `taketaxi/common/kitexGen/driver.proto` | 新增 | 添加 GetDriverProfile 相关消息定义 |
| `taketaxi/common/kitexGen/driver.pb.go` | 重新生成 | protoc 编译生成 |
| `taketaxi/common/kitexGen/driver_grpc.pb.go` | 重新生成 | protoc 编译生成 |
| `taketaxi/srvDriver/internal/repository/driverRepo.go` | 新增方法 | GetDriverProfile 查询方法 |
| `taketaxi/srvDriver/internal/service/driverService.go` | 新增方法 | GetProfile 业务逻辑 |
| `taketaxi/srvDriver/internal/handler/driverHandler.go` | 新增方法 | GetProfile gRPC 处理器 |
| `taketaxi/bffDriver/internal/rpcClient/driverClient.go` | 新增方法 | GetProfile RPC 调用 |
| `taketaxi/bffDriver/internal/handler/driverHandler.go` | 新增方法 | Profile HTTP 处理器 |
| `taketaxi/bffDriver/internal/router/router.go` | 新增路由 | GET /api/v1/driver/profile |

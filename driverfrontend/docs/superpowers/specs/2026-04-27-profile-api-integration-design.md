---
name: Profile API Integration
description: 前端对接后端 Profile API，替换 mock 数据
type: project
---

## 背景

前端目前使用 mock store 存储司机数据。需要对接后端 `GET /api/v1/driver/profile` 接口，获取真实数据。

## 目标

- 创建 API 服务模块，封装 HTTP 请求
- 通过环境变量配置 API baseURL
- 改造 Store，从 API 获取司机信息
- 处理加载和错误状态

## 技术方案

### 1. 环境变量配置

创建 `.env.development`:
```
VITE_API_BASE_URL=http://localhost:8080
```

### 2. API 服务模块

创建 `src/app/api/profile.ts`:
- 定义 `ProfileResponse` 类型（与后端 proto 一致）
- 定义 `PersonalInfo` 类型
- 定义 `OrderStats` 类型
- 实现 `fetchProfile(driverId: number)` 函数

### 3. Store 改造

修改 `src/app/store.tsx`:
- 删除 `initialDrivers` mock 数据
- 添加 `loading` 和 `error` 状态
- 添加 `fetchProfile()` 异步方法
- 字段映射逻辑

### 4. 组件调整

修改 `src/app/components/DriverApp.tsx`:
- 在 `useEffect` 中调用 `fetchProfile()`
- 添加加载中 UI
- 添加错误处理 UI

## 字段映射

| 后端字段 (proto) | 前端字段 (Driver) |
|-----------------|------------------|
| `personal_info.nickname` | `name` |
| `personal_info.phone` | `phone` |
| `personal_info.plate` | `plate` |
| `personal_info.car` | `car` |
| `personal_info.service_score` | `rating` |
| `personal_info.online` | `online` |
| `personal_info.status` | `status` |
| `order_stats.order_count` | `totalOrders` |
| `order_stats.income` | `todayEarnings` |

## 接口详情

**请求:**
```
GET /api/v1/driver/profile?driver_id=1&date=&days=1
```

**响应:**
```json
{
  "personal_info": {
    "nickname": "王师傅",
    "avatar": "",
    "service_score": 4.92,
    "order_count": 2341,
    "phone": "138****2341",
    "plate": "苏N·8F23K",
    "car": "轩逸 · 银色",
    "online": true,
    "status": "idle"
  },
  "order_stats": {
    "order_count": 8,
    "income": 328.5,
    "online_duration": 360
  },
  "verify_status": 1
}
```

## 暂时保留的功能

- 订单相关的 mock 数据（pending orders、history 等）—— 后端暂无对应接口
- `orders` 和 `complaints` 相关的 store 逻辑

## 风险与约束

- driver_id 暂时写死为 `1`，后续接入认证时替换
- 错误处理使用简单的 toast 提示
- 暂不实现 token 认证

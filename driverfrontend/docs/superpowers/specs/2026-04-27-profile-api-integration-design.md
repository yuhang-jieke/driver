---
name: Backend-Frontend Profile API Integration
description: 对接后端 Profile API 到前端 React 应用
type: project
---

# Backend-Frontend Profile API 集成设计

## 背景

前端 React 应用目前使用 mock 数据，需要对接后端已有的 `/api/v1/driver/profile` 接口获取司机真实数据。

## 范围

- 仅对接 Profile 接口
- 订单、抢单等功能保持 mock 数据
- 使用硬编码 driver_id = 1777200100810 进行测试

## 后端接口

### GET /api/v1/driver/profile

**请求参数**：
| 参数 | 类型 | 必填 | 说明 |
|-----|------|-----|------|
| driver_id | int64 | 是 | 司机ID |
| date | string | 否 | 查询日期，格式 YYYY-MM-DD |
| days | int32 | 否 | 查询天数，默认1 |

**响应结构**：
```json
{
  "personal_info": {
    "nickname": "string",
    "avatar": "string",
    "service_score": 80.0,
    "order_count": 100,
    "phone": "138****2341",
    "plate": "苏N·8F23K",
    "car": "轩逸 · 银色",
    "online": true,
    "status": "idle"
  },
  "order_stats": {
    "order_count": 8,
    "income": 328.5,
    "online_duration": 22320
  },
  "verify_status": 2
}
```

## 数据映射

| 后端字段 | 前端字段 | 转换规则 |
|---------|---------|---------|
| `personal_info.nickname` | `name` | 直接映射 |
| `personal_info.phone` | `phone` | 直接映射 |
| `personal_info.plate` | `plate` | 直接映射 |
| `personal_info.car` | `car` | 直接映射 |
| `personal_info.service_score` | `rating` | service_score / 20 (转成 0-5 分) |
| `personal_info.online` | `online` | 直接映射 |
| `personal_info.order_count` | `totalOrders` | 直接映射 |
| `order_stats.income` | `todayEarnings` | 直接映射 |
| `personal_info.status` | `status` | 直接映射 (idle/busy/offline) |

## 文件改动

### 新增文件

**`driverfrontend/src/app/api/profile.ts`**

职责：
- 定义后端 Profile 响应类型
- 封装 fetchDriverProfile 函数
- 实现数据转换函数

### 修改文件

**`driverfrontend/src/app/store.tsx`**

改动：
1. 导入 profile API
2. 添加 loading state
3. 在 StoreProvider 中用 useEffect 调用 API
4. API 成功时更新 drivers[0] 数据
5. API 失败时保持 mock 数据，打印错误日志

## API 调用流程

```
App 启动
    ↓
StoreProvider 初始化
    ↓
useEffect 调用 fetchDriverProfile(1777200100810)
    ↓
GET http://localhost:8080/api/v1/driver/profile?driver_id=1777200100810
    ↓
成功 → transformProfileData() → 更新 drivers state
失败 → console.error → 保持 mock 数据
    ↓
UI 渲染
```

## 错误处理

- 网络错误：保持 mock 数据，打印错误
- 后端错误响应：保持 mock 数据，打印错误
- 数据解析错误：保持 mock 数据，打印错误

## 测试验证

1. 启动后端服务 (端口 8080)
2. 启动前端开发服务器
3. 打开浏览器访问前端
4. 检查 Network 面板确认 API 调用
5. 验证首页司机信息是否来自后端

## 后续扩展

当需要对接更多接口时：
- 扩展 `api/` 目录结构
- 添加 axios 实例统一配置
- 实现登录认证获取 driver_id

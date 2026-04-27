# Profile API Integration Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 前端对接后端 Profile API，用真实数据替换 mock store 中的司机信息

**Architecture:** 创建独立的 API 服务层封装 HTTP 请求，Store 层调用 API 获取数据并处理状态，组件层消费 Store 数据并处理加载/错误状态

**Tech Stack:** React 18, TypeScript, Vite, fetch API

---

## File Structure

| File | Action | Purpose |
|------|--------|---------|
| `.env.development` | Create | 环境变量配置 API baseURL |
| `src/app/api/profile.ts` | Create | API 服务模块，类型定义和请求封装 |
| `src/app/store.tsx` | Modify | 添加 loading/error 状态，fetchProfile 方法 |
| `src/app/components/DriverApp.tsx` | Modify | 调用 API，处理加载和错误状态 |

---

### Task 1: 创建环境变量配置

**Files:**
- Create: `.env.development`

- [ ] **Step 1: 创建 .env.development 文件**

```bash
VITE_API_BASE_URL=http://localhost:8080
```

---

### Task 2: 创建 API 服务模块

**Files:**
- Create: `src/app/api/profile.ts`

- [ ] **Step 1: 创建 API 类型定义和请求函数**

```typescript
// API 响应类型定义（与后端 proto 一致）
export interface PersonalInfo {
  nickname: string;
  avatar: string;
  service_score: number;
  order_count: number;
  phone: string;
  plate: string;
  car: string;
  online: boolean;
  status: string;
}

export interface OrderStats {
  order_count: number;
  income: number;
  online_duration: number;
}

export interface ProfileResponse {
  personal_info: PersonalInfo;
  order_stats: OrderStats;
  verify_status: number;
}

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || '';

export async function fetchProfile(driverId: number): Promise<ProfileResponse> {
  const url = `${API_BASE_URL}/api/v1/driver/profile?driver_id=${driverId}&date=&days=1`;
  const response = await fetch(url);
  
  if (!response.ok) {
    throw new Error(`API error: ${response.status}`);
  }
  
  return response.json();
}
```

---

### Task 3: 改造 Store 添加 API 调用

**Files:**
- Modify: `src/app/store.tsx`

- [ ] **Step 1: 添加 loading 和 error 状态，修改 Store 接口**

在 `Store` interface 中添加：
```typescript
interface Store {
  orders: Order[];
  drivers: Driver[];
  complaints: Complaint[];
  currentDriverId: string;
  loading: boolean;
  error: string | null;
  createOrder: (o: Omit<Order, "id" | "createdAt" | "status">) => string;
  acceptOrder: (orderId: string, driverId: string) => void;
  updateOrderStatus: (orderId: string, status: OrderStatus) => void;
  cancelOrder: (orderId: string) => void;
  setDriverOnline: (driverId: string, online: boolean) => void;
  fetchProfile: () => Promise<void>;
}
```

- [ ] **Step 2: 修改 StoreProvider 添加状态和 fetchProfile 方法**

删除 `initialDrivers` 数组，修改 StoreProvider：

```typescript
export function StoreProvider({ children }: { children: ReactNode }) {
  const [orders, setOrders] = useState<Order[]>(initialOrders);
  const [drivers, setDrivers] = useState<Driver[]>([]);
  const [complaints, setComplaints] = useState<Complaint[]>(initialComplaints);
  const [currentDriverId] = useState("D001");
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchProfile = async () => {
    setLoading(true);
    setError(null);
    try {
      const resp = await fetchProfileApi(1); // driver_id 暂时写死为 1
      const info = resp.personal_info;
      const stats = resp.order_stats;
      
      const driver: Driver = {
        id: currentDriverId,
        name: info.nickname,
        phone: info.phone,
        plate: info.plate,
        car: info.car,
        rating: info.service_score,
        online: info.online,
        totalOrders: stats.order_count,
        todayEarnings: stats.income,
        status: info.status as "idle" | "busy" | "offline",
      };
      
      setDrivers([driver]);
    } catch (e) {
      setError(e instanceof Error ? e.message : "获取数据失败");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchProfile();
  }, []);

  // ... rest of the code remains the same
```

- [ ] **Step 3: 更新 Provider value 包含新状态**

```typescript
return (
  <StoreCtx.Provider value={{
    orders, drivers, complaints, currentDriverId, loading, error,
    createOrder, acceptOrder, updateOrderStatus, cancelOrder, setDriverOnline, fetchProfile,
  }}>
    {children}
  </StoreCtx.Provider>
);
```

- [ ] **Step 4: 添加导入语句**

在文件顶部添加：
```typescript
import { fetchProfile as fetchProfileApi } from "./api/profile";
```

---

### Task 4: 更新组件处理加载和错误状态

**Files:**
- Modify: `src/app/components/DriverApp.tsx`

- [ ] **Step 1: 从 Store 解构 loading 和 error**

修改 `DriverApp` 函数开头的解构：
```typescript
const { orders, drivers, currentDriverId, loading, error, setDriverOnline, acceptOrder, updateOrderStatus } = useStore();
```

- [ ] **Step 2: 添加加载中和错误状态 UI**

在 `DriverApp` 返回的 JSX 中，`PhoneFrame` 内部的 `div` 开头添加：
```typescript
{loading && (
  <div className="h-full flex items-center justify-center bg-gray-50">
    <div className="text-center">
      <div className="text-3xl mb-2">⏳</div>
      <div className="text-sm text-gray-400">加载中...</div>
    </div>
  </div>
)}
{error && (
  <div className="h-full flex items-center justify-center bg-gray-50">
    <div className="text-center p-6">
      <div className="text-3xl mb-2">😔</div>
      <div className="text-sm text-gray-600 mb-3">{error}</div>
      <button 
        onClick={() => window.location.reload()} 
        className="px-4 py-2 bg-emerald-500 text-white text-sm rounded-full"
      >
        重试
      </button>
    </div>
  </div>
)}
{!loading && !error && (
  <>
    {/* 原有的内容 */}
  </>
)}
```

- [ ] **Step 3: 调整原有内容包裹**

将原来的内容用 `<>...</>` 包裹，放在 `{!loading && !error && (...)}` 内部。

---

### Task 5: 提交代码

- [ ] **Step 1: 提交所有改动**

```bash
git add .env.development src/app/api/profile.ts src/app/store.tsx src/app/components/DriverApp.tsx
git commit -m "$(cat <<'EOF'
feat: integrate Profile API with frontend

- Add API service module for profile endpoint
- Configure API base URL via environment variable
- Replace mock driver data with real API response
- Add loading and error state handling

Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>
EOF
)"
```

---

## 完成标准

- [ ] `npm run dev` 启动无报错
- [ ] 页面显示加载状态
- [ ] API 成功返回时显示真实司机数据
- [ ] API 失败时显示错误提示和重试按钮
- [ ] 代码已提交

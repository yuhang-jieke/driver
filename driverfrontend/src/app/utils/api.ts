const BASE = "/api/v1";

async function get<T = any>(path: string, query?: Record<string, string>): Promise<T> {
  const url = query ? `${path}?${new URLSearchParams(query)}` : path;
  const res = await fetch(`${BASE}${url}`);
  if (!res.ok) throw new Error(`HTTP ${res.status}`);
  return res.json();
}

async function post<T = any>(path: string, body?: Record<string, any>): Promise<T> {
  const res = await fetch(`${BASE}${path}`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: body ? JSON.stringify(body) : undefined,
  });
  if (!res.ok) throw new Error(`HTTP ${res.status}`);
  return res.json();
}

// ---------- 基础接口 ----------

export interface ApiOrder {
  success: boolean;
  order_no: string;
  status: number;
  created_at: number;
  completed_at?: number;
  origin_address: string;
  dest_address: string;
  distance_km?: number;
  duration_min?: number;
  passenger_name: string;
  passenger_mobile: string;
  passenger_score?: number;
  passenger_comment?: string;
  total_fee?: number;
  platform_commission?: number;
  driver_income?: number;
  pay_type?: number;
  service_type?: number;
  nodes?: { name: string; time: number }[];
}

export interface ApiDriver {
  id: number;
  name: string;
  mobile: string;
  nickname: string;
  work_status: number;
  service_score: number;
  order_count: number;
  total_income: number;
}

export interface ApiOrderItem {
  order_no: string;
  service_type: number;
  origin_address: string;
  dest_address: string;
  distance_km: number;
  duration_min: number;
  status: number;
  created_at: number;
}

export interface ApiListOrdersResp {
  success: boolean;
  items: ApiOrderItem[];
}

export interface ApiGetOrderResp extends ApiOrder {}

// ---------- 司机相关 ----------

export const apiDriver = {
  get(id: number) {
    return get<ApiDriver>(`/drivers/${id}`);
  },
  goOnline(id: number) {
    return post(`/drivers/${id}/go-online`);
  },
  goOffline(id: number) {
    return post(`/drivers/${id}/go-offline`);
  },
  startListening(id: number, lat: number, lng: number) {
    return post(`/drivers/${id}/start-listening`, { lat, lng });
  },
  reportLocation(id: number, lat: number, lng: number, heading: number, speed: number, status: number) {
    return post(`/drivers/${id}/report-location`, { lat, lng, heading, speed, status });
  },
};

// ---------- 订单相关 ----------

export const apiOrder = {
  dispatch(body: { order_id: number; service_type: number; origin_lat: number; origin_lng: number; passenger_id: number }) {
    return post("/orders/dispatch", body);
  },
  accept(orderId: number, driverId: number) {
    return post(`/orders/${orderId}/accept`, { driver_id: driverId });
  },
  reject(orderId: number, driverId: number) {
    return post(`/orders/${orderId}/reject`, { driver_id: driverId });
  },
  cancel(orderId: number, driverId: number, cancel_reason: string) {
    return post(`/orders/${orderId}/cancel`, { driver_id: driverId, cancel_reason });
  },
  arrive(orderId: number, driverId: number) {
    return post(`/orders/${orderId}/arrive`, { driver_id: driverId });
  },
  verifyPassenger(orderId: number, driverId: number, phone_last4: string) {
    return post(`/orders/${orderId}/verify-passenger`, { driver_id: driverId, phone_last4 });
  },
  startTrip(orderId: number, driverId: number) {
    return post(`/orders/${orderId}/start-trip`, { driver_id: driverId });
  },
  endTrip(orderId: number, driverId: number) {
    return post(`/orders/${orderId}/end-trip`, { driver_id: driverId });
  },
};

// ---------- 订单查询 ----------

export const apiOrderQuery = {
  list(driverId: number, params?: { date?: string; cursor?: number; is_all?: boolean }) {
    const q: Record<string, string> = {};
    if (params?.date) q.date = params.date;
    if (params?.cursor) q.cursor = String(params.cursor);
    if (params?.is_all) q.is_all = "true";
    return get<ApiListOrdersResp>(`/drivers/${driverId}/orders`, q);
  },
  detail(orderId: number, driverId: number) {
    return get<ApiGetOrderResp>(`/orders/${orderId}/detail`, { driver_id: String(driverId) });
  },
};

// ---------- 抢单池 ----------

export const apiPool = {
  list(driverId: number, page = 1, pageSize = 20) {
    return post("/pool/list", { driver_id: driverId, page, page_size: pageSize });
  },
  grab(orderId: number, driverId: number) {
    return post(`/orders/${orderId}/grab`, { driver_id: driverId });
  },
};

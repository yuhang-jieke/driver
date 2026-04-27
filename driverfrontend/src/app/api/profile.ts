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

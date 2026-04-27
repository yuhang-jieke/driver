import type { Driver } from "../store";

const API_BASE = "http://localhost:8080";

interface PersonalInfo {
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

interface OrderStats {
  order_count: number;
  income: number;
  online_duration: number;
}

interface ProfileResponse {
  personal_info: PersonalInfo;
  order_stats: OrderStats;
  verify_status: number;
}

export async function fetchDriverProfile(driverId: number): Promise<ProfileResponse | null> {
  try {
    const res = await fetch(
      `${API_BASE}/api/v1/driver/profile?driver_id=${driverId}`
    );
    if (!res.ok) {
      console.error("Profile API error:", res.status, res.statusText);
      return null;
    }
    return await res.json();
  } catch (err) {
    console.error("Profile API fetch failed:", err);
    return null;
  }
}

export function transformProfileData(data: ProfileResponse, driverId: string): Partial<Driver> {
  const { personal_info, order_stats } = data;
  return {
    id: driverId,
    name: personal_info.nickname || "司机",
    phone: personal_info.phone || "",
    plate: personal_info.plate || "",
    car: personal_info.car || "",
    rating: personal_info.service_score ? personal_info.service_score / 20 : 4.5,
    online: personal_info.online ?? false,
    totalOrders: personal_info.order_count || 0,
    todayEarnings: order_stats?.income || 0,
    status: (personal_info.status as Driver["status"]) || "offline",
  };
}

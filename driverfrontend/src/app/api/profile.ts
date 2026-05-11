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

interface RealnameInfo {
  real_name: string;
  id_card_front_url: string;
  id_card_back_url: string;
  status: number;
}

interface LicenseInfo {
  license_no: string;
  license_type: string;
  license_url: string;
  status: number;
}

interface VehicleInfo {
  plate_no: string;
  vehicle_brand: string;
  vehicle_model: string;
  vehicle_color: string;
  seat_count: number;
  driving_license_url: string;
  vehicle_photo_url: string;
  status: number;
}

interface ProfileResponse {
  personal_info: PersonalInfo;
  order_stats: OrderStats;
  verify_status: number;
  realname_info?: RealnameInfo;
  license_info?: LicenseInfo;
  vehicle_info?: VehicleInfo;
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

// 更新司机个人资料
export async function updateDriverProfile(params: {
  driver_id: number;
  nickname?: string;
  avatar?: string;
  gender?: number;
}): Promise<{ success: boolean; message: string } | null> {
  try {
    const res = await fetch(`${API_BASE}/api/v1/driver/profile`, {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(params),
    });
    if (!res.ok) return null;
    return await res.json();
  } catch (err) {
    console.error("UpdateProfile API fetch failed:", err);
    return null;
  }
}

export function transformProfileData(data: ProfileResponse, driverId: string): Partial<Driver> {
  const { personal_info, order_stats, realname_info, license_info, vehicle_info } = data;
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
    avatar: personal_info.avatar || "",
    // 认证信息（供认证页回填）
    realnameInfo: realname_info,
    licenseInfo: license_info,
    vehicleInfo: vehicle_info,
  };
}

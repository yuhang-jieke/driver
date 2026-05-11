const API_BASE = "http://localhost:8080";

export interface ApiResponse {
  success: boolean;
  message: string;
}

// 修改手机号
export async function changeMobile(
  driverId: number,
  newMobile: string,
  verifyCode: string
): Promise<ApiResponse | null> {
  try {
    const res = await fetch(`${API_BASE}/api/v1/driver/mobile`, {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ driver_id: driverId, new_mobile: newMobile, verify_code: verifyCode }),
    });
    if (!res.ok) {
      console.error("ChangeMobile API error:", res.status);
      return null;
    }
    return await res.json();
  } catch (err) {
    console.error("ChangeMobile API fetch failed:", err);
    return null;
  }
}

// 修改密码
export async function changePassword(
  driverId: number,
  oldPassword: string,
  newPassword: string
): Promise<ApiResponse | null> {
  try {
    const res = await fetch(`${API_BASE}/api/v1/driver/password`, {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ driver_id: driverId, old_password: oldPassword, new_password: newPassword }),
    });
    if (!res.ok) {
      console.error("ChangePassword API error:", res.status);
      return null;
    }
    return await res.json();
  } catch (err) {
    console.error("ChangePassword API fetch failed:", err);
    return null;
  }
}

// 重置密码（忘记密码）
export async function resetPassword(
  mobile: string,
  verifyCode: string,
  newPassword: string
): Promise<ApiResponse | null> {
  try {
    const res = await fetch(`${API_BASE}/api/v1/driver/password/reset`, {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ mobile, verify_code: verifyCode, new_password: newPassword }),
    });
    if (!res.ok) {
      console.error("ResetPassword API error:", res.status);
      return null;
    }
    return await res.json();
  } catch (err) {
    console.error("ResetPassword API fetch failed:", err);
    return null;
  }
}

const API_BASE = "http://localhost:8080";

export interface UpdateRealnameParams {
  driver_id: number;
  real_name?: string;
  id_card_no?: string;
  id_card_front_url?: string;
  id_card_back_url?: string;
  gender?: number;
  birthday?: string;
  address?: string;
  nation?: string;
  expire_date?: string;
}

export async function updateRealname(params: UpdateRealnameParams): Promise<{ success: boolean; message: string } | null> {
  try {
    const res = await fetch(`${API_BASE}/api/v1/driver/realname`, {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(params),
    });
    if (!res.ok) return null;
    return await res.json();
  } catch (err) {
    console.error("UpdateRealname API fetch failed:", err);
    return null;
  }
}

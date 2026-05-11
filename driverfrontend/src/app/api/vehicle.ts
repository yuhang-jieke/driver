const API_BASE = "http://localhost:8080";

export interface UpdateVehicleParams {
  driver_id: number;
  plate_no?: string;
  vehicle_model?: string;
  vehicle_brand?: string;
  vehicle_color?: string;
  vehicle_color_code?: string;
  seat_count?: number;
  driving_license_url?: string;
  vehicle_photo_url?: string;
}

export async function updateVehicle(params: UpdateVehicleParams): Promise<{ success: boolean; message: string } | null> {
  try {
    const res = await fetch(`${API_BASE}/api/v1/driver/vehicle`, {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(params),
    });
    if (!res.ok) return null;
    return await res.json();
  } catch (err) {
    console.error("UpdateVehicle API fetch failed:", err);
    return null;
  }
}

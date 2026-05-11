const API_BASE = "http://localhost:8080";

export interface UpdateLicenseParams {
  driver_id: number;
  license_no?: string;
  license_type?: string;
  license_url?: string;
  first_issue_date?: string;
  issue_date?: string;
  expire_date?: string;
}

export async function updateLicense(params: UpdateLicenseParams): Promise<{ success: boolean; message: string } | null> {
  try {
    const res = await fetch(`${API_BASE}/api/v1/driver/license`, {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(params),
    });
    if (!res.ok) return null;
    return await res.json();
  } catch (err) {
    console.error("UpdateLicense API fetch failed:", err);
    return null;
  }
}

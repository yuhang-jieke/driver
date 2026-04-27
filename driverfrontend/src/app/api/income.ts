const API_BASE = "http://localhost:8080";

export interface IncomeSummary {
  order_count: number;
  income: number;
  online_duration: number;
}

export interface DailyIncome {
  date: string;
  income: number;
}

export interface IncomeResponse {
  summary: IncomeSummary;
  trend: DailyIncome[];
}

export async function fetchDriverIncome(
  driverId: number,
  period: "today" | "week" | "month"
): Promise<IncomeResponse | null> {
  try {
    const res = await fetch(
      `${API_BASE}/api/v1/driver/income?driver_id=${driverId}&period=${period}`
    );
    if (!res.ok) {
      console.error("Income API error:", res.status, res.statusText);
      return null;
    }
    return await res.json();
  } catch (err) {
    console.error("Income API fetch failed:", err);
    return null;
  }
}

export function formatDuration(seconds: number): string {
  const hours = seconds / 3600;
  return hours.toFixed(1);
}

export function calculateHourlyIncome(income: number, seconds: number): number {
  if (seconds <= 0) return 0;
  const hours = seconds / 3600;
  if (hours <= 0) return 0;
  return income / hours;
}

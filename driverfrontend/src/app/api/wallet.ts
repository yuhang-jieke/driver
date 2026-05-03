const API_BASE = "http://localhost:8080";

// 验证持卡人姓名格式（2-6个汉字）
function validateAccountName(name: string): boolean {
  const regex = /^[一-龥]{2,6}$/;
  return regex.test(name);
}

// ==================== 银行卡 ====================

export interface BankCardInfo {
  bank_name: string;
  bank_code: string;
  bank_card_no: string;    // 脱敏
  account_name: string;
  card_type: number;
  branch_name: string;
  has_card: boolean;
  last_modified_at: string;
}

export async function getBankCard(
  driverId: number
): Promise<BankCardInfo | null> {
  try {
    const res = await fetch(
      `${API_BASE}/api/v1/driver/bankcard?driver_id=${driverId}`
    );
    if (!res.ok) return null;
    return await res.json();
  } catch (err) {
    console.error("GetBankCard API failed:", err);
    return null;
  }
}

export async function bindBankCard(params: {
  driver_id: number;
  bank_name: string;
  bank_code?: string;
  bank_card_no: string;
  account_name: string;
  card_type?: number;
  branch_name?: string;
}): Promise<{ success: boolean; message: string } | null> {
  // 客户端验证：持卡人姓名格式
  if (!validateAccountName(params.account_name)) {
    return {
      success: false,
      message: "持卡人姓名必须为2-6个汉字"
    };
  }

  try {
    const res = await fetch(`${API_BASE}/api/v1/driver/bankcard`, {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(params),
    });
    if (!res.ok) return null;
    return await res.json();
  } catch (err) {
    console.error("BindBankCard API failed:", err);
    return null;
  }
}

export async function updateBankCard(params: {
  driver_id: number;
  bank_name: string;
  bank_code?: string;
  bank_card_no: string;
  account_name: string;
  card_type?: number;
  branch_name?: string;
}): Promise<{ success: boolean; message: string } | null> {
  // 客户端验证：持卡人姓名格式
  if (!validateAccountName(params.account_name)) {
    return {
      success: false,
      message: "持卡人姓名必须为2-6个汉字"
    };
  }

  try {
    const res = await fetch(`${API_BASE}/api/v1/driver/bankcard/update`, {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(params),
    });
    if (!res.ok) return null;
    return await res.json();
  } catch (err) {
    console.error("UpdateBankCard API failed:", err);
    return null;
  }
}

// ==================== 钱包 ====================

export interface WalletInfo {
  balance: number;              // 可提现余额
  frozen_amount: number;        // 冻结金额
  today_income: number;         // 今日收入
  week_income: number;          // 本周收入
  month_income: number;         // 本月收入
  total_income: number;         // 累计总收入
  total_withdraw: number;       // 累计提现
  today_withdraw_count: number; // 今日已提现次数
  bank_card_no: string;         // 已绑定银行卡(脱敏)
  has_bank_card: boolean;       // 是否已绑定银行卡
}

export async function getWallet(
  driverId: number
): Promise<WalletInfo | null> {
  try {
    const res = await fetch(
      `${API_BASE}/api/v1/driver/wallet?driver_id=${driverId}`
    );
    if (!res.ok) return null;
    return await res.json();
  } catch (err) {
    console.error("GetWallet API failed:", err);
    return null;
  }
}

// ==================== 提现 ====================

export async function applyWithdraw(
  driverId: number,
  amount: number
): Promise<{ success: boolean; message: string; withdraw_no?: string } | null> {
  try {
    const res = await fetch(`${API_BASE}/api/v1/driver/withdraw`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ driver_id: driverId, amount }),
    });
    if (!res.ok) return null;
    return await res.json();
  } catch (err) {
    console.error("ApplyWithdraw API failed:", err);
    return null;
  }
}

// ==================== 提现记录 ====================

export interface WithdrawRecord {
  id: number;
  withdraw_no: string;
  amount: number;
  fee: number;
  actual_amount: number;
  bank_name: string;
  bank_card_no: string;
  status: number;  // 1-处理中 2-成功 3-失败
  fail_reason: string;
  apply_time: string;
  finish_time: string;
}

export async function getWithdrawRecords(
  driverId: number,
  page = 1,
  pageSize = 20
): Promise<{ records: WithdrawRecord[]; total: number } | null> {
  try {
    const res = await fetch(
      `${API_BASE}/api/v1/driver/withdraw/records?driver_id=${driverId}&page=${page}&page_size=${pageSize}`
    );
    if (!res.ok) return null;
    return await res.json();
  } catch (err) {
    console.error("GetWithdrawRecords API failed:", err);
    return null;
  }
}

// ==================== 收入明细 ====================

export interface IncomeDetailItem {
  type_name: string;
  type_code: number;
  amount: number;
  count: number;
}

export async function getIncomeDetail(
  driverId: number,
  days: number // 1-今日 7-本周 30-本月
): Promise<{ items: IncomeDetailItem[]; total_amount: number } | null> {
  try {
    const res = await fetch(
      `${API_BASE}/api/v1/driver/income/detail?driver_id=${driverId}&days=${days}`
    );
    if (!res.ok) return null;
    return await res.json();
  } catch (err) {
    console.error("GetIncomeDetail API failed:", err);
    return null;
  }
}

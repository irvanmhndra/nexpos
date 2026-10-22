export type DailySettlementStatus = 'draft' | 'finalized';

export type PaymentMethod =
  | 'cash'
  | 'debit_card'
  | 'credit_card'
  | 'e_wallet'
  | 'bank_transfer'
  | 'qris';

export interface DailySettlementItem {
  id: number;
  payment_method: PaymentMethod;
  expected_amount: number;
  actual_amount: number | null;
  variance_amount: number;
  notes: string | null;
}

export interface DailySettlement {
  id: number;
  branch_id: number;
  settlement_date: string; // YYYY-MM-DD
  status: DailySettlementStatus;
  total_sales: number;
  total_refunds: number;
  total_expenses: number;
  total_expected: number;
  total_actual: number;
  total_variance: number;
  notes: string | null;
  recorded_by: number | null;
  finalized_by: number | null;
  finalized_at: string | null;
  items?: DailySettlementItem[];
  created_at: string;
  updated_at: string;
}

export interface CreateDailySettlementRequest {
  branch_id: number;
  settlement_date: string;
  notes?: string;
}

export interface UpdateSettlementItemRequest {
  actual_amount: number;
  notes?: string;
}

export interface BulkUpdateSettlementItem {
  item_id: number;
  actual_amount: number;
  notes?: string;
}

export interface BulkUpdateSettlementItemsRequest {
  items: BulkUpdateSettlementItem[];
}

export interface FinalizeDailySettlementRequest {
  notes?: string;
}

export interface SettlementMethodBreakdown {
  payment_method: PaymentMethod;
  gross_sales: number;
  refunds: number;
  expenses_out: number;
  expected_amount: number;
}

export interface DailySettlementReport {
  branch_id: number;
  settlement_date: string;
  total_sales: number;
  total_refunds: number;
  total_expenses: number;
  total_expected: number;
  by_method: SettlementMethodBreakdown[];
}

export type ShiftStatus = 'open' | 'closed';

export interface Shift {
  id: number;
  company_id: number;
  branch_id: number;
  cashier_id: number;
  status: ShiftStatus;
  opening_float: number;
  closing_float: number | null;
  expected_cash: number | null;
  actual_cash: number | null;
  cash_difference: number | null;
  notes: string | null;
  opened_at: string;
  closed_at: string | null;
  cashier_name: string | null;
  created_at: string;
  updated_at: string;
}

export interface OpenShiftRequest {
  branch_id: number;
  opening_float: number;
}

export interface CloseShiftRequest {
  actual_cash: number;
  notes?: string;
}

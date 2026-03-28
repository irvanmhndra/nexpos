export interface ExpenseCategory {
  id: number;
  company_id: number;
  name: string;
  description: string | null;
  created_at: string;
  updated_at: string;
}

export interface Expense {
  id: number;
  company_id: number;
  branch_id: number | null;
  category_id: number | null;
  amount: number;
  description: string;
  reference_no: string | null;
  expense_date: string;
  recorded_by: number | null;
  notes: string | null;
  category?: ExpenseCategory;
  created_at: string;
  updated_at: string;
}

export interface CreateExpenseRequest {
  branch_id?: number;
  category_id?: number;
  amount: number;
  description: string;
  reference_no?: string;
  expense_date: string;
  notes?: string;
}

export interface UpdateExpenseRequest {
  branch_id?: number;
  category_id?: number;
  amount?: number;
  description?: string;
  reference_no?: string;
  expense_date?: string;
  notes?: string;
}

export interface ExpenseSummaryItem {
  category_name: string;
  total: number;
}

import { api } from './api';
import type {
  Expense,
  ExpenseCategory,
  CreateExpenseRequest,
  UpdateExpenseRequest,
  ExpenseSummaryItem,
} from '@/types/expense';

interface ApiResponse<T> {
  success: boolean;
  message: string;
  data: T;
  meta?: {
    pagination?: {
      total_records: number;
      total_pages: number;
      current_page: number;
      per_page: number;
      next_page: number | null;
      prev_page: number | null;
    };
  };
}

interface GetExpensesParams {
  page?: number;
  per_page?: number;
  category_id?: number;
  branch_id?: number;
  date_from?: string;
  date_to?: string;
}

export const expenseService = {
  getExpenses: async (params: GetExpensesParams = {}) => {
    const { data } = await api.get<ApiResponse<Expense[]>>('/expenses', { params });
    return {
      data: data.data,
      total: data.meta?.pagination?.total_records || 0,
      page: data.meta?.pagination?.current_page || 1,
      limit: data.meta?.pagination?.per_page || 10,
      totalPages: data.meta?.pagination?.total_pages || 1,
    };
  },

  getExpense: async (id: number) => {
    const { data } = await api.get<ApiResponse<Expense>>(`/expenses/${id}`);
    return data.data;
  },

  createExpense: async (expense: CreateExpenseRequest) => {
    const { data } = await api.post<ApiResponse<Expense>>('/expenses', expense);
    return data.data;
  },

  updateExpense: async (id: number, expense: UpdateExpenseRequest) => {
    const { data } = await api.put<ApiResponse<Expense>>(`/expenses/${id}`, expense);
    return data.data;
  },

  deleteExpense: async (id: number) => {
    const { data } = await api.delete<ApiResponse<null>>(`/expenses/${id}`);
    return data;
  },

  getExpenseCategories: async () => {
    const { data } = await api.get<ApiResponse<ExpenseCategory[]>>('/expense-categories');
    return data.data;
  },

  createExpenseCategory: async (category: { name: string; description?: string }) => {
    const { data } = await api.post<ApiResponse<ExpenseCategory>>('/expense-categories', category);
    return data.data;
  },

  updateExpenseCategory: async (id: number, category: { name?: string; description?: string }) => {
    const { data } = await api.put<ApiResponse<ExpenseCategory>>(`/expense-categories/${id}`, category);
    return data.data;
  },

  deleteExpenseCategory: async (id: number) => {
    const { data } = await api.delete<ApiResponse<null>>(`/expense-categories/${id}`);
    return data;
  },

  getExpenseSummary: async (params?: { date_from?: string; date_to?: string; branch_id?: number }) => {
    const { data } = await api.get<ApiResponse<ExpenseSummaryItem[]>>('/expenses/summary', { params });
    return data.data;
  },
};

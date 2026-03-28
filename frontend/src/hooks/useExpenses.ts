import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { expenseService } from '@/services/expenseService';
import type { CreateExpenseRequest, UpdateExpenseRequest } from '@/types/expense';

interface UseExpensesParams {
  page?: number;
  per_page?: number;
  category_id?: number;
  branch_id?: number;
  date_from?: string;
  date_to?: string;
}

export function useExpenses(params: UseExpensesParams = {}) {
  return useQuery({
    queryKey: ['expenses', params],
    queryFn: () => expenseService.getExpenses(params),
  });
}

export function useExpense(id: number) {
  return useQuery({
    queryKey: ['expenses', id],
    queryFn: () => expenseService.getExpense(id),
    enabled: !!id,
  });
}

export function useCreateExpense() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (expense: CreateExpenseRequest) => expenseService.createExpense(expense),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['expenses'] });
    },
  });
}

export function useUpdateExpense() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, expense }: { id: number; expense: UpdateExpenseRequest }) =>
      expenseService.updateExpense(id, expense),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['expenses'] });
    },
  });
}

export function useDeleteExpense() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: number) => expenseService.deleteExpense(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['expenses'] });
    },
  });
}

export function useExpenseCategories() {
  return useQuery({
    queryKey: ['expense-categories'],
    queryFn: () => expenseService.getExpenseCategories(),
  });
}

export function useCreateExpenseCategory() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (category: { name: string; description?: string }) =>
      expenseService.createExpenseCategory(category),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['expense-categories'] });
    },
  });
}

export function useUpdateExpenseCategory() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, category }: { id: number; category: { name?: string; description?: string } }) =>
      expenseService.updateExpenseCategory(id, category),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['expense-categories'] });
    },
  });
}

export function useDeleteExpenseCategory() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: number) => expenseService.deleteExpenseCategory(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['expense-categories'] });
    },
  });
}

export function useExpenseSummary(params?: { date_from?: string; date_to?: string; branch_id?: number }) {
  return useQuery({
    queryKey: ['expenses', 'summary', params],
    queryFn: () => expenseService.getExpenseSummary(params),
  });
}

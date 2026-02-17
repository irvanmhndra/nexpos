import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { customersService } from '@/services/customersService';
import type { CreateCustomerRequest, UpdateCustomerRequest } from '@/types/customer';

interface UseCustomersParams {
  page?: number;
  per_page?: number;
  search?: string;
  is_member?: boolean;
}

export function useCustomers(params: UseCustomersParams = {}) {
  return useQuery({
    queryKey: ['customers', params],
    queryFn: () => customersService.getCustomers(params),
  });
}

export function useCustomer(id: number) {
  return useQuery({
    queryKey: ['customers', id],
    queryFn: () => customersService.getCustomer(id),
    enabled: !!id,
  });
}

export function useCreateCustomer() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (customer: CreateCustomerRequest) => customersService.createCustomer(customer),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['customers'] });
    },
  });
}

export function useUpdateCustomer() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, customer }: { id: number; customer: UpdateCustomerRequest }) =>
      customersService.updateCustomer(id, customer),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['customers'] });
    },
  });
}

export function useDeleteCustomer() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: number) => customersService.deleteCustomer(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['customers'] });
    },
  });
}

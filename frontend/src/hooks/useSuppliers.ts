import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { supplierService } from '@/services/supplierService';
import type { CreateSupplierRequest, UpdateSupplierRequest } from '@/types/supplier';

interface UseSuppliersParams {
  page?: number;
  per_page?: number;
  search?: string;
  is_active?: boolean;
}

export function useSuppliers(params: UseSuppliersParams = {}) {
  return useQuery({
    queryKey: ['suppliers', params],
    queryFn: () => supplierService.getSuppliers(params),
  });
}

export function useSupplier(id: number) {
  return useQuery({
    queryKey: ['suppliers', id],
    queryFn: () => supplierService.getSupplier(id),
    enabled: !!id,
  });
}

export function useCreateSupplier() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (supplier: CreateSupplierRequest) => supplierService.createSupplier(supplier),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['suppliers'] });
    },
  });
}

export function useUpdateSupplier() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, supplier }: { id: number; supplier: UpdateSupplierRequest }) =>
      supplierService.updateSupplier(id, supplier),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['suppliers'] });
    },
  });
}

export function useDeleteSupplier() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: number) => supplierService.deleteSupplier(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['suppliers'] });
    },
  });
}

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { stockOpnameService } from '@/services/stockOpnameService';
import type {
  CreateStockOpnameRequest,
  UpdateOpnameItemRequest,
  BulkUpdateOpnameItemsRequest,
  CompleteStockOpnameRequest,
} from '@/types/stockOpname';

interface ListParams {
  page?: number;
  per_page?: number;
  status?: string;
  branch_id?: number;
}

export function useStockOpnames(params: ListParams = {}) {
  return useQuery({
    queryKey: ['stock-opnames', params],
    queryFn: () => stockOpnameService.list(params),
  });
}

export function useStockOpname(id: number) {
  return useQuery({
    queryKey: ['stock-opnames', id],
    queryFn: () => stockOpnameService.get(id),
    enabled: !!id,
  });
}

export function useCreateStockOpname() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (request: CreateStockOpnameRequest) => stockOpnameService.create(request),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['stock-opnames'] });
    },
  });
}

export function useUpdateOpnameItem(opnameId: number) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ itemId, request }: { itemId: number; request: UpdateOpnameItemRequest }) =>
      stockOpnameService.updateItem(opnameId, itemId, request),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['stock-opnames', opnameId] });
    },
  });
}

export function useBulkUpdateOpnameItems(opnameId: number) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (request: BulkUpdateOpnameItemsRequest) =>
      stockOpnameService.bulkUpdateItems(opnameId, request),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['stock-opnames', opnameId] });
    },
  });
}

export function useCompleteStockOpname(opnameId: number) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (request: CompleteStockOpnameRequest = {}) =>
      stockOpnameService.complete(opnameId, request),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['stock-opnames'] });
      queryClient.invalidateQueries({ queryKey: ['inventory'] });
    },
  });
}

export function useCancelStockOpname(opnameId: number) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: () => stockOpnameService.cancel(opnameId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['stock-opnames'] });
    },
  });
}

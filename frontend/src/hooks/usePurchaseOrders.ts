import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { purchaseOrderService } from '@/services/purchaseOrderService';
import type { CreatePORequest, ReceivePORequest } from '@/types/purchaseOrder';

interface UsePurchaseOrdersParams {
  page?: number;
  per_page?: number;
  status?: string;
  supplier_id?: number;
  branch_id?: number;
}

export function usePurchaseOrders(params: UsePurchaseOrdersParams = {}) {
  return useQuery({
    queryKey: ['purchase-orders', params],
    queryFn: () => purchaseOrderService.getPurchaseOrders(params),
  });
}

export function usePurchaseOrder(id: number) {
  return useQuery({
    queryKey: ['purchase-orders', id],
    queryFn: () => purchaseOrderService.getPurchaseOrder(id),
    enabled: !!id,
  });
}

export function useCreatePurchaseOrder() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (request: CreatePORequest) => purchaseOrderService.createPurchaseOrder(request),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['purchase-orders'] });
    },
  });
}

export function useReceivePurchaseOrder() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, request }: { id: number; request: ReceivePORequest }) =>
      purchaseOrderService.receivePurchaseOrder(id, request),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['purchase-orders'] });
      queryClient.invalidateQueries({ queryKey: ['inventory'] });
    },
  });
}

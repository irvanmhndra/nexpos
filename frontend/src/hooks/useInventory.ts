import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { inventoryService } from '@/services/inventoryService';
import { useBranch } from '@/contexts/BranchContext';
import type {
  AdjustStockRequest,
  UpdateMinStockRequest,
  GetInventoryParams,
  GetMovementsParams,
} from '@/types/inventory';

export function useInventory(params: GetInventoryParams = {}) {
  const { activeBranch } = useBranch();
  return useQuery({
    queryKey: ['inventory', params, activeBranch?.id],
    queryFn: () => inventoryService.getInventory({ ...params, branch_id: activeBranch?.id }),
  });
}

export function useInventoryStats() {
  const { activeBranch } = useBranch();
  return useQuery({
    queryKey: ['inventory', 'stats', activeBranch?.id],
    queryFn: () => inventoryService.getInventoryStats(activeBranch?.id),
  });
}

export function useStockMovements(params: GetMovementsParams = {}) {
  const { activeBranch } = useBranch();
  return useQuery({
    queryKey: ['inventory', 'movements', params, activeBranch?.id],
    queryFn: () => inventoryService.getMovements({ ...params, branch_id: activeBranch?.id }),
  });
}

export function useMovementStats() {
  const { activeBranch } = useBranch();
  return useQuery({
    queryKey: ['inventory', 'movements', 'stats', activeBranch?.id],
    queryFn: () => inventoryService.getMovementStats(activeBranch?.id),
  });
}

export function useAdjustStock() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (request: AdjustStockRequest) => inventoryService.adjustStock(request),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['inventory'] });
      queryClient.invalidateQueries({ queryKey: ['products'] });
    },
  });
}

export function useUpdateMinStock() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ variantId, request }: { variantId: number; request: UpdateMinStockRequest }) =>
      inventoryService.updateMinStock(variantId, request),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['inventory'] });
    },
  });
}

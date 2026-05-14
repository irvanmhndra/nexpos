import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { dailySettlementService } from '@/services/dailySettlementService';
import type {
  CreateDailySettlementRequest,
  UpdateSettlementItemRequest,
  BulkUpdateSettlementItemsRequest,
  FinalizeDailySettlementRequest,
} from '@/types/dailySettlement';

interface ListParams {
  page?: number;
  per_page?: number;
  status?: string;
  branch_id?: number;
  date_from?: string;
  date_to?: string;
}

export function useDailySettlements(params: ListParams = {}) {
  return useQuery({
    queryKey: ['daily-settlements', params],
    queryFn: () => dailySettlementService.list(params),
  });
}

export function useDailySettlement(id: number) {
  return useQuery({
    queryKey: ['daily-settlements', id],
    queryFn: () => dailySettlementService.get(id),
    enabled: !!id,
  });
}

export function useDailySettlementReport(branchId: number | undefined, date: string | undefined) {
  return useQuery({
    queryKey: ['daily-settlements', 'report', branchId, date],
    queryFn: () => dailySettlementService.report(branchId!, date!),
    enabled: !!branchId && !!date,
  });
}

export function useCreateDailySettlement() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (request: CreateDailySettlementRequest) => dailySettlementService.create(request),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['daily-settlements'] });
    },
  });
}

export function useUpdateSettlementItem(settlementId: number) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ itemId, request }: { itemId: number; request: UpdateSettlementItemRequest }) =>
      dailySettlementService.updateItem(settlementId, itemId, request),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['daily-settlements', settlementId] });
    },
  });
}

export function useBulkUpdateSettlementItems(settlementId: number) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (request: BulkUpdateSettlementItemsRequest) =>
      dailySettlementService.bulkUpdateItems(settlementId, request),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['daily-settlements', settlementId] });
    },
  });
}

export function useFinalizeDailySettlement(settlementId: number) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (request: FinalizeDailySettlementRequest = {}) =>
      dailySettlementService.finalize(settlementId, request),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['daily-settlements'] });
    },
  });
}

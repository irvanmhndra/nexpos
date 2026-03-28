import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { shiftService } from '@/services/shiftService';
import type { OpenShiftRequest, CloseShiftRequest } from '@/types/shift';
import { useBranch } from '@/contexts/BranchContext';

interface UseShiftsParams {
  page?: number;
  per_page?: number;
  status?: string;
  date_from?: string;
  date_to?: string;
}

export function useShifts(params: UseShiftsParams = {}) {
  const { activeBranch } = useBranch();
  return useQuery({
    queryKey: ['shifts', params, activeBranch?.id],
    queryFn: () => shiftService.getShifts({ ...params, branch_id: activeBranch?.id }),
  });
}

export function useCurrentShift() {
  const { activeBranch } = useBranch();
  return useQuery({
    queryKey: ['shifts', 'current', activeBranch?.id],
    queryFn: () => shiftService.getCurrentShift(activeBranch?.id),
    refetchInterval: 60000,
  });
}

export function useOpenShift() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (request: OpenShiftRequest) => shiftService.openShift(request),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['shifts'] });
    },
  });
}

export function useCloseShift() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, request }: { id: number; request: CloseShiftRequest }) =>
      shiftService.closeShift(id, request),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['shifts'] });
    },
  });
}

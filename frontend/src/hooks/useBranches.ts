import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { branchesService } from '@/services/branchesService';
import type { CreateBranchRequest, UpdateBranchRequest } from '@/types/branch';

interface UseBranchesParams {
  page?: number;
  per_page?: number;
  search?: string;
  status?: string;
  region?: string;
}

export function useBranches(params: UseBranchesParams = {}) {
  return useQuery({
    queryKey: ['branches', params],
    queryFn: () => branchesService.getBranches(params),
  });
}

export function useCreateBranch() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (branch: CreateBranchRequest) => branchesService.createBranch(branch),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['branches'] });
    },
  });
}

export function useUpdateBranch() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, branch }: { id: string; branch: UpdateBranchRequest }) =>
      branchesService.updateBranch(id, branch),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['branches'] });
    },
  });
}

export function useDeleteBranch() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: string) => branchesService.deleteBranch(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['branches'] });
    },
  });
}

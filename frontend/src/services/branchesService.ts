import { api } from './api';
import type { Branch, CreateBranchRequest, UpdateBranchRequest } from '@/types/branch';

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

interface GetBranchesParams {
  page?: number;
  per_page?: number;
  search?: string;
  is_active?: boolean;
}

export const branchesService = {
  getBranches: async (params: GetBranchesParams = {}) => {
    const { data } = await api.get<ApiResponse<Branch[]>>('/branches', { params });
    return {
      data: data.data,
      total: data.meta?.pagination?.total_records || 0,
      page: data.meta?.pagination?.current_page || 1,
      limit: data.meta?.pagination?.per_page || 10,
      totalPages: data.meta?.pagination?.total_pages || 1,
    };
  },

  getBranch: async (id: number) => {
    const { data } = await api.get<ApiResponse<Branch>>(`/branches/${id}`);
    return data.data;
  },

  createBranch: async (branch: CreateBranchRequest) => {
    const { data } = await api.post<ApiResponse<Branch>>('/branches', branch);
    return data.data;
  },

  updateBranch: async (id: number, branch: UpdateBranchRequest) => {
    const { data } = await api.put<ApiResponse<Branch>>(`/branches/${id}`, branch);
    return data.data;
  },

  deleteBranch: async (id: number) => {
    const { data } = await api.delete<ApiResponse<null>>(`/branches/${id}`);
    return data;
  },
};

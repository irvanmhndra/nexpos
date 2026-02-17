import { api } from './api';
import type { Branch, CreateBranchRequest, UpdateBranchRequest } from '@/types/branch';

interface BranchesListResponse {
  data: Branch[];
  total: number;
  page: number;
  limit: number;
  totalPages: number;
}

interface GetBranchesParams {
  page?: number;
  per_page?: number;
  search?: string;
  status?: string;
  region?: string;
}

export const branchesService = {
  getBranches: async (params: GetBranchesParams = {}) => {
    const { data } = await api.get<BranchesListResponse>('/branches', { params });
    return {
      data: data.data,
      total: data.total || 0,
      page: data.page || 1,
      limit: data.limit || 10,
      totalPages: data.totalPages || 1,
    };
  },

  getBranch: async (id: string) => {
    const { data } = await api.get<Branch>(`/branches/${id}`);
    return data;
  },

  createBranch: async (branch: CreateBranchRequest) => {
    const { data } = await api.post<Branch>('/branches', branch);
    return data;
  },

  updateBranch: async (id: string, branch: UpdateBranchRequest) => {
    const { data } = await api.put<Branch>(`/branches/${id}`, branch);
    return data;
  },

  deleteBranch: async (id: string) => {
    await api.delete(`/branches/${id}`);
  },
};

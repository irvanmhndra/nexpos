import { api } from './api';
import type { Shift, OpenShiftRequest, CloseShiftRequest } from '@/types/shift';

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

interface GetShiftsParams {
  page?: number;
  per_page?: number;
  status?: string;
  branch_id?: number;
  date_from?: string;
  date_to?: string;
}

export const shiftService = {
  openShift: async (request: OpenShiftRequest) => {
    const { data } = await api.post<ApiResponse<Shift>>('/shifts/open', request);
    return data.data;
  },

  getCurrentShift: async (branchId?: number) => {
    const { data } = await api.get<ApiResponse<Shift | null>>('/shifts/current', {
      params: branchId ? { branch_id: branchId } : undefined,
    });
    return data.data;
  },

  closeShift: async (id: number, request: CloseShiftRequest) => {
    const { data } = await api.post<ApiResponse<Shift>>(`/shifts/${id}/close`, request);
    return data.data;
  },

  getShift: async (id: number) => {
    const { data } = await api.get<ApiResponse<Shift>>(`/shifts/${id}`);
    return data.data;
  },

  getShifts: async (params: GetShiftsParams = {}) => {
    const { data } = await api.get<ApiResponse<Shift[]>>('/shifts', { params });
    return {
      data: data.data,
      total: data.meta?.pagination?.total_records || 0,
      page: data.meta?.pagination?.current_page || 1,
      limit: data.meta?.pagination?.per_page || 10,
      totalPages: data.meta?.pagination?.total_pages || 1,
    };
  },
};

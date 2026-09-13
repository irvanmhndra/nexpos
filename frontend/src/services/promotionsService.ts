import { api } from './api';
import type { Promotion, CreatePromotionRequest, UpdatePromotionRequest } from '@/types/promotion';

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

interface GetPromotionsParams {
  page?: number;
  per_page?: number;
  search?: string;
  is_active?: boolean;
  type?: string;
  status?: string;
}

export const promotionsService = {
  getPromotions: async (params: GetPromotionsParams = {}) => {
    const { data } = await api.get<ApiResponse<Promotion[]>>('/promotions', { params });
    return {
      data: data.data,
      total: data.meta?.pagination?.total_records || 0,
      page: data.meta?.pagination?.current_page || 1,
      limit: data.meta?.pagination?.per_page || 10,
      totalPages: data.meta?.pagination?.total_pages || 1,
    };
  },

  getPromotion: async (id: number) => {
    const { data } = await api.get<ApiResponse<Promotion>>(`/promotions/${id}`);
    return data.data;
  },

  createPromotion: async (promotion: CreatePromotionRequest) => {
    const { data } = await api.post<ApiResponse<Promotion>>('/promotions', promotion);
    return data.data;
  },

  updatePromotion: async (id: number, promotion: UpdatePromotionRequest) => {
    const { data } = await api.put<ApiResponse<Promotion>>(`/promotions/${id}`, promotion);
    return data.data;
  },

  deletePromotion: async (id: number) => {
    const { data } = await api.delete<ApiResponse<null>>(`/promotions/${id}`);
    return data;
  },
};

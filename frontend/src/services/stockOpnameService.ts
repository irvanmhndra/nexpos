import { api } from './api';
import type {
  StockOpname,
  StockOpnameItem,
  CreateStockOpnameRequest,
  UpdateOpnameItemRequest,
  BulkUpdateOpnameItemsRequest,
  CompleteStockOpnameRequest,
} from '@/types/stockOpname';

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

interface ListParams {
  page?: number;
  per_page?: number;
  status?: string;
  branch_id?: number;
}

export const stockOpnameService = {
  list: async (params: ListParams = {}) => {
    const { data } = await api.get<ApiResponse<StockOpname[]>>('/stock-opnames', { params });
    return {
      data: data.data,
      total: data.meta?.pagination?.total_records || 0,
      page: data.meta?.pagination?.current_page || 1,
      perPage: data.meta?.pagination?.per_page || 10,
      totalPages: data.meta?.pagination?.total_pages || 1,
    };
  },

  get: async (id: number) => {
    const { data } = await api.get<ApiResponse<StockOpname>>(`/stock-opnames/${id}`);
    return data.data;
  },

  create: async (request: CreateStockOpnameRequest) => {
    const { data } = await api.post<ApiResponse<StockOpname>>('/stock-opnames', request);
    return data.data;
  },

  updateItem: async (opnameId: number, itemId: number, request: UpdateOpnameItemRequest) => {
    const { data } = await api.patch<ApiResponse<StockOpnameItem>>(
      `/stock-opnames/${opnameId}/items/${itemId}`,
      request,
    );
    return data.data;
  },

  bulkUpdateItems: async (opnameId: number, request: BulkUpdateOpnameItemsRequest) => {
    const { data } = await api.patch<ApiResponse<StockOpname>>(
      `/stock-opnames/${opnameId}/items`,
      request,
    );
    return data.data;
  },

  complete: async (opnameId: number, request: CompleteStockOpnameRequest = {}) => {
    const { data } = await api.post<ApiResponse<StockOpname>>(
      `/stock-opnames/${opnameId}/complete`,
      request,
    );
    return data.data;
  },

  cancel: async (opnameId: number) => {
    const { data } = await api.post<ApiResponse<StockOpname>>(`/stock-opnames/${opnameId}/cancel`);
    return data.data;
  },
};

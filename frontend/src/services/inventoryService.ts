import { api } from './api';
import type {
  InventoryItem,
  InventoryStats,
  StockMovement,
  MovementStats,
  AdjustStockRequest,
  UpdateMinStockRequest,
  GetInventoryParams,
  GetMovementsParams,
} from '@/types/inventory';

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

export const inventoryService = {
  getInventory: async (params: GetInventoryParams = {}) => {
    const { data } = await api.get<ApiResponse<InventoryItem[]>>('/inventory', { params });
    return {
      data: data.data,
      total: data.meta?.pagination?.total_records || 0,
      page: data.meta?.pagination?.current_page || 1,
      limit: data.meta?.pagination?.per_page || 10,
      totalPages: data.meta?.pagination?.total_pages || 1,
    };
  },

  getInventoryStats: async (branchId?: number) => {
    const { data } = await api.get<ApiResponse<InventoryStats>>('/inventory/stats', {
      params: branchId ? { branch_id: branchId } : undefined,
    });
    return data.data;
  },

  getMovements: async (params: GetMovementsParams = {}) => {
    const { data } = await api.get<ApiResponse<StockMovement[]>>('/inventory/movements', { params });
    return {
      data: data.data,
      total: data.meta?.pagination?.total_records || 0,
      page: data.meta?.pagination?.current_page || 1,
      limit: data.meta?.pagination?.per_page || 10,
      totalPages: data.meta?.pagination?.total_pages || 1,
    };
  },

  getMovementStats: async (branchId?: number) => {
    const { data } = await api.get<ApiResponse<MovementStats>>('/inventory/movements/stats', {
      params: branchId ? { branch_id: branchId } : undefined,
    });
    return data.data;
  },

  adjustStock: async (request: AdjustStockRequest) => {
    const { data } = await api.post<ApiResponse<StockMovement>>('/inventory/adjust', request);
    return data.data;
  },

  updateMinStock: async (variantId: number, request: UpdateMinStockRequest) => {
    const { data } = await api.put<ApiResponse<null>>(`/inventory/${variantId}/min-stock`, request);
    return data.data;
  },
};

import { api } from './api';
import type { PurchaseOrder, CreatePORequest, ReceivePORequest } from '@/types/purchaseOrder';

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

interface GetPurchaseOrdersParams {
  page?: number;
  per_page?: number;
  status?: string;
  supplier_id?: number;
  branch_id?: number;
}

export const purchaseOrderService = {
  getPurchaseOrders: async (params: GetPurchaseOrdersParams = {}) => {
    const { data } = await api.get<ApiResponse<PurchaseOrder[]>>('/purchase-orders', { params });
    return {
      data: data.data,
      total: data.meta?.pagination?.total_records || 0,
      page: data.meta?.pagination?.current_page || 1,
      limit: data.meta?.pagination?.per_page || 10,
      totalPages: data.meta?.pagination?.total_pages || 1,
    };
  },

  getPurchaseOrder: async (id: number) => {
    const { data } = await api.get<ApiResponse<PurchaseOrder>>(`/purchase-orders/${id}`);
    return data.data;
  },

  createPurchaseOrder: async (request: CreatePORequest) => {
    const { data } = await api.post<ApiResponse<PurchaseOrder>>('/purchase-orders', request);
    return data.data;
  },

  receivePurchaseOrder: async (id: number, request: ReceivePORequest) => {
    const { data } = await api.post<ApiResponse<PurchaseOrder>>(`/purchase-orders/${id}/receive`, request);
    return data.data;
  },
};

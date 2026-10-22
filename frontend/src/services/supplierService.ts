import { api } from './api';
import type { Supplier, CreateSupplierRequest, UpdateSupplierRequest } from '@/types/supplier';

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

interface GetSuppliersParams {
  page?: number;
  per_page?: number;
  search?: string;
  is_active?: boolean;
}

export const supplierService = {
  getSuppliers: async (params: GetSuppliersParams = {}) => {
    const { data } = await api.get<ApiResponse<Supplier[]>>('/suppliers', { params });
    return {
      data: data.data,
      total: data.meta?.pagination?.total_records || 0,
      page: data.meta?.pagination?.current_page || 1,
      limit: data.meta?.pagination?.per_page || 10,
      totalPages: data.meta?.pagination?.total_pages || 1,
    };
  },

  getSupplier: async (id: number) => {
    const { data } = await api.get<ApiResponse<Supplier>>(`/suppliers/${id}`);
    return data.data;
  },

  createSupplier: async (supplier: CreateSupplierRequest) => {
    const { data } = await api.post<ApiResponse<Supplier>>('/suppliers', supplier);
    return data.data;
  },

  updateSupplier: async (id: number, supplier: UpdateSupplierRequest) => {
    const { data } = await api.put<ApiResponse<Supplier>>(`/suppliers/${id}`, supplier);
    return data.data;
  },

  deleteSupplier: async (id: number) => {
    const { data } = await api.delete<ApiResponse<null>>(`/suppliers/${id}`);
    return data;
  },
};

import { api } from './api';
import type { Customer, CreateCustomerRequest, UpdateCustomerRequest } from '@/types/customer';

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

interface GetCustomersParams {
  page?: number;
  per_page?: number;
  search?: string;
  is_member?: boolean;
}

export const customersService = {
  getCustomers: async (params: GetCustomersParams = {}) => {
    const { data } = await api.get<ApiResponse<Customer[]>>('/customers', { params });
    return {
      data: data.data,
      total: data.meta?.pagination?.total_records || 0,
      page: data.meta?.pagination?.current_page || 1,
      limit: data.meta?.pagination?.per_page || 20,
      totalPages: data.meta?.pagination?.total_pages || 1,
    };
  },

  getCustomer: async (id: number) => {
    const { data } = await api.get<ApiResponse<Customer>>(`/customers/${id}`);
    return data.data;
  },

  createCustomer: async (customer: CreateCustomerRequest) => {
    const { data } = await api.post<ApiResponse<Customer>>('/customers', customer);
    return data.data;
  },

  updateCustomer: async (id: number, customer: UpdateCustomerRequest) => {
    const { data } = await api.put<ApiResponse<Customer>>(`/customers/${id}`, customer);
    return data.data;
  },

  deleteCustomer: async (id: number) => {
    const { data } = await api.delete<ApiResponse<null>>(`/customers/${id}`);
    return data;
  },
};

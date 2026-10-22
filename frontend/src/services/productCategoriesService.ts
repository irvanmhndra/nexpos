import { api } from './api';
import type { ProductCategory, CreateProductCategoryRequest, UpdateProductCategoryRequest } from '@/types/product';

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

interface GetProductCategoriesParams {
  page?: number;
  per_page?: number;
  search?: string;
  is_active?: boolean;
}

export const productCategoriesService = {
  getCategories: async (params: GetProductCategoriesParams = {}) => {
    const { data } = await api.get<ApiResponse<ProductCategory[]>>('/product-categories', { params });
    return {
      data: data.data,
      total: data.meta?.pagination?.total_records || 0,
      page: data.meta?.pagination?.current_page || 1,
      limit: data.meta?.pagination?.per_page || 20,
      totalPages: data.meta?.pagination?.total_pages || 1,
    };
  },

  getAllCategories: async () => {
    const { data } = await api.get<ApiResponse<ProductCategory[]>>('/product-categories/all');
    return data.data;
  },

  getCategory: async (id: number) => {
    const { data } = await api.get<ApiResponse<ProductCategory>>(`/product-categories/${id}`);
    return data.data;
  },

  createCategory: async (category: CreateProductCategoryRequest) => {
    const { data } = await api.post<ApiResponse<ProductCategory>>('/product-categories', category);
    return data.data;
  },

  updateCategory: async (id: number, category: UpdateProductCategoryRequest) => {
    const { data } = await api.put<ApiResponse<ProductCategory>>(`/product-categories/${id}`, category);
    return data.data;
  },

  deleteCategory: async (id: number) => {
    const { data } = await api.delete<ApiResponse<null>>(`/product-categories/${id}`);
    return data;
  },
};

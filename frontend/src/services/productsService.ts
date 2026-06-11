import { api } from './api';
import type {
  Product,
  CreateProductRequest,
  UpdateProductRequest,
  ProductLookupResponse,
} from '@/types/product';

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

interface GetProductsParams {
  page?: number;
  per_page?: number;
  search?: string;
  category_id?: number;
  is_active?: boolean;
}

export const productsService = {
  getProducts: async (params: GetProductsParams = {}) => {
    const { data } = await api.get<ApiResponse<Product[]>>('/products', { params });
    return {
      data: data.data,
      total: data.meta?.pagination?.total_records || 0,
      page: data.meta?.pagination?.current_page || 1,
      limit: data.meta?.pagination?.per_page || 20,
      totalPages: data.meta?.pagination?.total_pages || 1,
    };
  },

  getProduct: async (id: number) => {
    const { data } = await api.get<ApiResponse<Product>>(`/products/${id}`);
    return data.data;
  },

  // Resolve a scanned code (barcode first, SKU fallback) to its product +
  // matched variant. Throws (404) when nothing matches.
  lookupByCode: async (code: string) => {
    const { data } = await api.get<ApiResponse<ProductLookupResponse>>('/products/lookup', {
      params: { code },
    });
    return data.data;
  },

  createProduct: async (product: CreateProductRequest) => {
    const { data } = await api.post<ApiResponse<Product>>('/products', product);
    return data.data;
  },

  updateProduct: async (id: number, product: UpdateProductRequest) => {
    const { data } = await api.put<ApiResponse<Product>>(`/products/${id}`, product);
    return data.data;
  },

  deleteProduct: async (id: number) => {
    const { data } = await api.delete<ApiResponse<null>>(`/products/${id}`);
    return data;
  },
};

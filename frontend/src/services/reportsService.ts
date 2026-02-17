import { api } from './api';
import type { ApiResponse } from '@/types/api';
import type {
  ReportSummary,
  SalesTrendResponse,
  TopProductsResponse,
  CategoryRevenueResponse,
  PaymentMethodResponse,
  HourlySalesResponse,
  ReportParams,
} from '@/types/report';

export const reportsService = {
  getSummary: async (params: ReportParams) => {
    const { data } = await api.get<ApiResponse<ReportSummary>>('/reports/summary', { params });
    return data.data;
  },

  getSalesTrend: async (params: ReportParams) => {
    const { data } = await api.get<ApiResponse<SalesTrendResponse>>('/reports/sales-trend', { params });
    return data.data;
  },

  getTopProducts: async (params: ReportParams & { limit?: number }) => {
    const { data } = await api.get<ApiResponse<TopProductsResponse>>('/reports/top-products', { params });
    return data.data;
  },

  getCategoryRevenue: async (params: ReportParams) => {
    const { data } = await api.get<ApiResponse<CategoryRevenueResponse>>('/reports/category-revenue', { params });
    return data.data;
  },

  getPaymentMethods: async (params: ReportParams) => {
    const { data } = await api.get<ApiResponse<PaymentMethodResponse>>('/reports/payment-methods', { params });
    return data.data;
  },

  getHourlySales: async (params: ReportParams) => {
    const { data } = await api.get<ApiResponse<HourlySalesResponse>>('/reports/hourly-sales', { params });
    return data.data;
  },
};

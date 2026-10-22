import { useQuery } from '@tanstack/react-query';
import { reportsService } from '@/services/reportsService';
import type { ReportParams } from '@/types/report';

export function useReportSummary(params: ReportParams) {
  return useQuery({
    queryKey: ['reports', 'summary', params],
    queryFn: () => reportsService.getSummary(params),
    enabled: !!params.date_from && !!params.date_to,
  });
}

export function useSalesTrend(params: ReportParams) {
  return useQuery({
    queryKey: ['reports', 'sales-trend', params],
    queryFn: () => reportsService.getSalesTrend(params),
    enabled: !!params.date_from && !!params.date_to,
  });
}

export function useTopProducts(params: ReportParams & { limit?: number }) {
  return useQuery({
    queryKey: ['reports', 'top-products', params],
    queryFn: () => reportsService.getTopProducts(params),
    enabled: !!params.date_from && !!params.date_to,
  });
}

export function useCategoryRevenue(params: ReportParams) {
  return useQuery({
    queryKey: ['reports', 'category-revenue', params],
    queryFn: () => reportsService.getCategoryRevenue(params),
    enabled: !!params.date_from && !!params.date_to,
  });
}

export function usePaymentMethods(params: ReportParams) {
  return useQuery({
    queryKey: ['reports', 'payment-methods', params],
    queryFn: () => reportsService.getPaymentMethods(params),
    enabled: !!params.date_from && !!params.date_to,
  });
}

export function useHourlySales(params: ReportParams) {
  return useQuery({
    queryKey: ['reports', 'hourly-sales', params],
    queryFn: () => reportsService.getHourlySales(params),
    enabled: !!params.date_from && !!params.date_to,
  });
}

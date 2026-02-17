import { useQuery } from '@tanstack/react-query';
import { dashboardService } from '@/services/dashboardService';

export function useDashboardStats() {
  return useQuery({
    queryKey: ['dashboard', 'stats'],
    queryFn: dashboardService.getStats,
  });
}

export function useRecentOrders() {
  return useQuery({
    queryKey: ['dashboard', 'recent-orders'],
    queryFn: dashboardService.getRecentOrders,
  });
}

export function useSalesTrend() {
  return useQuery({
    queryKey: ['dashboard', 'sales-trend'],
    queryFn: dashboardService.getSalesTrend,
  });
}

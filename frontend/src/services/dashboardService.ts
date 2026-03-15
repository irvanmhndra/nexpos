import { api } from './api';
import type { ApiResponse } from '@/types/api';

// Types for dashboard data
export interface DashboardStats {
  totalSales: number;
  salesGrowth: number;
  newOrders: number;
  ordersGrowth: number;
  newCustomers: number;
  customersGrowth: number;
  netIncome: number;
  incomeGrowth: number;
}

export interface SalesTrendData {
  date: string;
  sales: number;
  orders: number;
}

interface ReportSummary {
  total_revenue: number;
  total_orders: number;
  total_items_sold: number;
  avg_order_value: number;
  total_customers: number;
  new_customers: number;
  completed_orders: number;
  cancelled_orders: number;
  pending_orders: number;
  total_discount: number;
  total_tax: number;
  gross_profit: number;
  gross_profit_margin: number;
}

interface SalesTrendItem {
  date: string;
  sales: number;
  orders: number;
  items: number;
}

interface SalesTrendResponse {
  data: SalesTrendItem[];
}

// Helper to format date as YYYY-MM-DD
function formatDate(date: Date): string {
  return date.toISOString().split('T')[0];
}

// Helper to get date range
function getDateRange(days: number): { dateFrom: string; dateTo: string } {
  const dateTo = new Date();
  const dateFrom = new Date();
  dateFrom.setDate(dateFrom.getDate() - days);
  return {
    dateFrom: formatDate(dateFrom),
    dateTo: formatDate(dateTo),
  };
}

// Helper to get previous month range
function getPreviousMonthRange(): { dateFrom: string; dateTo: string } {
  const now = new Date();
  const firstDayLastMonth = new Date(now.getFullYear(), now.getMonth() - 1, 1);
  const lastDayLastMonth = new Date(now.getFullYear(), now.getMonth(), 0);
  return {
    dateFrom: formatDate(firstDayLastMonth),
    dateTo: formatDate(lastDayLastMonth),
  };
}

// Helper to get current month range
function getCurrentMonthRange(): { dateFrom: string; dateTo: string } {
  const now = new Date();
  const firstDayThisMonth = new Date(now.getFullYear(), now.getMonth(), 1);
  return {
    dateFrom: formatDate(firstDayThisMonth),
    dateTo: formatDate(now),
  };
}

export const dashboardService = {
  getStats: async (): Promise<DashboardStats> => {
    // Get current month data
    const currentRange = getCurrentMonthRange();
    const previousRange = getPreviousMonthRange();

    try {
      // Fetch current and previous month data in parallel
      const [currentRes, previousRes] = await Promise.all([
        api.get<ApiResponse<ReportSummary>>('/reports/summary', {
          params: { date_from: currentRange.dateFrom, date_to: currentRange.dateTo },
        }),
        api.get<ApiResponse<ReportSummary>>('/reports/summary', {
          params: { date_from: previousRange.dateFrom, date_to: previousRange.dateTo },
        }),
      ]);

      const current = currentRes.data.data;
      const previous = previousRes.data.data;

      // Calculate growth percentages
      const calcGrowth = (curr: number, prev: number): number => {
        if (prev === 0) return curr > 0 ? 100 : 0;
        return Math.round(((curr - prev) / prev) * 100);
      };

      return {
        totalSales: current.total_revenue,
        salesGrowth: calcGrowth(current.total_revenue, previous.total_revenue),
        newOrders: current.total_orders,
        ordersGrowth: calcGrowth(current.total_orders, previous.total_orders),
        newCustomers: current.new_customers,
        customersGrowth: calcGrowth(current.new_customers, previous.new_customers),
        netIncome: current.gross_profit,
        incomeGrowth: calcGrowth(current.gross_profit, previous.gross_profit),
      };
    } catch {
      // Return default stats if API fails
      return {
        totalSales: 0,
        salesGrowth: 0,
        newOrders: 0,
        ordersGrowth: 0,
        newCustomers: 0,
        customersGrowth: 0,
        netIncome: 0,
        incomeGrowth: 0,
      };
    }
  },

  getRecentOrders: async () => {
    try {
      const { data } = await api.get('/orders', {
        params: { per_page: 5, page: 1 },
      });
      // Map to expected format for DataTable
      interface RawOrder { id: number; order_no: string; customer_name?: string; created_at: string; status: string; grand_total: number; }
      return (data.data || []).map((order: RawOrder) => ({
        id: order.id,
        orderNumber: order.order_no,
        customerName: order.customer_name || 'Walk-in Customer',
        createdAt: order.created_at,
        status: order.status,
        total: order.grand_total,
      }));
    } catch {
      return [];
    }
  },

  getSalesTrend: async (): Promise<SalesTrendData[]> => {
    const { dateFrom, dateTo } = getDateRange(30);

    try {
      const { data } = await api.get<ApiResponse<SalesTrendResponse>>('/reports/sales-trend', {
        params: { date_from: dateFrom, date_to: dateTo },
      });

      return (data.data?.data || []).map((item) => ({
        date: item.date,
        sales: item.sales,
        orders: item.orders,
      }));
    } catch {
      return [];
    }
  },
};

import { useState } from 'react';
import { Card, CardTitle, CardDescription, Icon, DateRangePicker } from '@/components/ui';
import { SalesChart, BarChart, PieChart } from '@/components/common';
import { formatCurrency } from '@/utils';
import {
  useReportSummary,
  useSalesTrend,
  useTopProducts,
  useCategoryRevenue,
  usePaymentMethods,
} from '@/hooks/useReports';

// Get default date range (last 30 days)
const getDefaultDateRange = () => {
  const end = new Date();
  const start = new Date();
  start.setDate(end.getDate() - 30);
  return {
    startDate: start.toISOString().split('T')[0],
    endDate: end.toISOString().split('T')[0],
  };
};

export function ReportsPage() {
  const defaultRange = getDefaultDateRange();
  const [startDate, setStartDate] = useState(defaultRange.startDate);
  const [endDate, setEndDate] = useState(defaultRange.endDate);

  const reportParams = { date_from: startDate, date_to: endDate };

  const { data: summary, isLoading: summaryLoading } = useReportSummary(reportParams);
  const { data: salesTrend, isLoading: trendLoading } = useSalesTrend(reportParams);
  const { data: topProducts, isLoading: productsLoading } = useTopProducts({ ...reportParams, limit: 5 });
  const { data: categoryRevenue, isLoading: categoryLoading } = useCategoryRevenue(reportParams);
  const { data: paymentMethods } = usePaymentMethods(reportParams);

  const handleDateRangeChange = (start: string, end: string) => {
    setStartDate(start);
    setEndDate(end);
  };

  // Transform data for charts
  const salesTrendData = salesTrend?.data?.map(item => ({
    date: item.date,
    sales: item.sales,
    orders: item.orders,
  })) || [];

  const topProductsData = topProducts?.data?.map(item => ({
    name: item.product_name,
    value: item.total_amount,
  })) || [];

  const categoryRevenueData = categoryRevenue?.data?.map(item => ({
    name: item.category_name,
    value: item.total_amount,
  })) || [];

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h1 className="text-2xl md:text-3xl font-bold text-text-main dark:text-white">
            Laporan & Analitik
          </h1>
          <p className="text-text-sub dark:text-gray-400 text-sm mt-1">
            Visualisasi performa bisnis dan insight mendalam
          </p>
        </div>
        <DateRangePicker
          startDate={startDate}
          endDate={endDate}
          onChange={handleDateRangeChange}
          className="w-full sm:w-72"
        />
      </div>

      {/* Key Metrics */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <Card variant="bordered" padding="md">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-xs text-text-sub dark:text-gray-400 mb-1">Total Pendapatan</p>
              {summaryLoading ? (
                <div className="h-8 w-32 bg-gray-200 dark:bg-gray-700 rounded animate-pulse" />
              ) : (
                <p className="text-2xl font-bold text-text-main dark:text-white">
                  {formatCurrency(summary?.total_revenue || 0)}
                </p>
              )}
              <p className="text-xs text-text-sub dark:text-gray-400 mt-1">
                Margin: {summary?.gross_profit_margin?.toFixed(1) || 0}%
              </p>
            </div>
            <div className="w-12 h-12 rounded-full bg-blue-50 dark:bg-blue-900/20 flex items-center justify-center">
              <Icon name="payments" size={24} className="text-primary" />
            </div>
          </div>
        </Card>

        <Card variant="bordered" padding="md">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-xs text-text-sub dark:text-gray-400 mb-1">Total Pesanan</p>
              {summaryLoading ? (
                <div className="h-8 w-20 bg-gray-200 dark:bg-gray-700 rounded animate-pulse" />
              ) : (
                <p className="text-2xl font-bold text-text-main dark:text-white">
                  {summary?.total_orders || 0}
                </p>
              )}
              <p className="text-xs text-text-sub dark:text-gray-400 mt-1">
                Selesai: {summary?.completed_orders || 0}
              </p>
            </div>
            <div className="w-12 h-12 rounded-full bg-green-50 dark:bg-green-900/20 flex items-center justify-center">
              <Icon name="shopping_cart" size={24} className="text-green-600" />
            </div>
          </div>
        </Card>

        <Card variant="bordered" padding="md">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-xs text-text-sub dark:text-gray-400 mb-1">Rata-rata Nilai Order</p>
              {summaryLoading ? (
                <div className="h-8 w-28 bg-gray-200 dark:bg-gray-700 rounded animate-pulse" />
              ) : (
                <p className="text-2xl font-bold text-text-main dark:text-white">
                  {formatCurrency(summary?.avg_order_value || 0)}
                </p>
              )}
              <p className="text-xs text-text-sub dark:text-gray-400 mt-1">
                Item terjual: {summary?.total_items_sold || 0}
              </p>
            </div>
            <div className="w-12 h-12 rounded-full bg-yellow-50 dark:bg-yellow-900/20 flex items-center justify-center">
              <Icon name="account_balance_wallet" size={24} className="text-yellow-600" />
            </div>
          </div>
        </Card>

        <Card variant="bordered" padding="md">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-xs text-text-sub dark:text-gray-400 mb-1">Pelanggan</p>
              {summaryLoading ? (
                <div className="h-8 w-16 bg-gray-200 dark:bg-gray-700 rounded animate-pulse" />
              ) : (
                <p className="text-2xl font-bold text-text-main dark:text-white">
                  {summary?.total_customers || 0}
                </p>
              )}
              <p className="text-xs text-green-600 dark:text-green-400 mt-1">
                +{summary?.new_customers || 0} pelanggan baru
              </p>
            </div>
            <div className="w-12 h-12 rounded-full bg-purple-50 dark:bg-purple-900/20 flex items-center justify-center">
              <Icon name="groups" size={24} className="text-purple-600" />
            </div>
          </div>
        </Card>
      </div>

      {/* Charts Row 1 */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Sales Trend */}
        <Card variant="bordered" padding="md">
          <div className="mb-4">
            <CardTitle>Tren Penjualan</CardTitle>
            <CardDescription className="mt-1">
              Grafik penjualan harian dalam periode yang dipilih
            </CardDescription>
          </div>
          {trendLoading ? (
            <div className="h-64 flex items-center justify-center">
              <Icon name="progress_activity" size={32} className="animate-spin text-primary" />
            </div>
          ) : salesTrendData.length > 0 ? (
            <SalesChart data={salesTrendData} />
          ) : (
            <div className="h-64 flex items-center justify-center text-text-sub dark:text-gray-400">
              Tidak ada data penjualan
            </div>
          )}
        </Card>

        {/* Top Products */}
        <Card variant="bordered" padding="md">
          <div className="mb-4">
            <CardTitle>Produk Terlaris</CardTitle>
            <CardDescription className="mt-1">
              5 produk dengan penjualan tertinggi
            </CardDescription>
          </div>
          {productsLoading ? (
            <div className="h-64 flex items-center justify-center">
              <Icon name="progress_activity" size={32} className="animate-spin text-primary" />
            </div>
          ) : topProductsData.length > 0 ? (
            <BarChart data={topProductsData} color="#10b981" valuePrefix="Rp" />
          ) : (
            <div className="h-64 flex items-center justify-center text-text-sub dark:text-gray-400">
              Tidak ada data produk
            </div>
          )}
        </Card>
      </div>

      {/* Charts Row 2 */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Revenue by Category */}
        <Card variant="bordered" padding="md">
          <div className="mb-4">
            <CardTitle>Pendapatan per Kategori</CardTitle>
            <CardDescription className="mt-1">
              Distribusi pendapatan berdasarkan kategori produk
            </CardDescription>
          </div>
          {categoryLoading ? (
            <div className="h-64 flex items-center justify-center">
              <Icon name="progress_activity" size={32} className="animate-spin text-primary" />
            </div>
          ) : categoryRevenueData.length > 0 ? (
            <PieChart data={categoryRevenueData} />
          ) : (
            <div className="h-64 flex items-center justify-center text-text-sub dark:text-gray-400">
              Tidak ada data kategori
            </div>
          )}
        </Card>

        {/* Performance Summary */}
        <Card variant="bordered" padding="md">
          <div className="mb-4">
            <CardTitle>Ringkasan Performa</CardTitle>
            <CardDescription className="mt-1">
              Metrik kunci performa bisnis
            </CardDescription>
          </div>
          <div className="space-y-4">
            <div className="flex items-center justify-between p-4 bg-background-light dark:bg-background-dark rounded-lg">
              <div className="flex items-center gap-3">
                <div className="w-10 h-10 rounded-full bg-blue-50 dark:bg-blue-900/20 flex items-center justify-center">
                  <Icon name="shopping_bag" size={20} className="text-primary" />
                </div>
                <div>
                  <p className="text-sm font-medium text-text-main dark:text-white">
                    Produk Terjual
                  </p>
                  <p className="text-xs text-text-sub dark:text-gray-400">Total unit terjual</p>
                </div>
              </div>
              <p className="text-xl font-bold text-text-main dark:text-white">
                {summary?.total_items_sold?.toLocaleString('id-ID') || 0}
              </p>
            </div>

            <div className="flex items-center justify-between p-4 bg-background-light dark:bg-background-dark rounded-lg">
              <div className="flex items-center gap-3">
                <div className="w-10 h-10 rounded-full bg-green-50 dark:bg-green-900/20 flex items-center justify-center">
                  <Icon name="trending_up" size={20} className="text-green-600" />
                </div>
                <div>
                  <p className="text-sm font-medium text-text-main dark:text-white">
                    Gross Profit
                  </p>
                  <p className="text-xs text-text-sub dark:text-gray-400">Laba kotor</p>
                </div>
              </div>
              <p className="text-xl font-bold text-text-main dark:text-white">
                {formatCurrency(summary?.gross_profit || 0)}
              </p>
            </div>

            {/* Payment Methods Summary */}
            {paymentMethods?.data && paymentMethods.data.length > 0 && (
              <div className="flex items-center justify-between p-4 bg-background-light dark:bg-background-dark rounded-lg">
                <div className="flex items-center gap-3">
                  <div className="w-10 h-10 rounded-full bg-purple-50 dark:bg-purple-900/20 flex items-center justify-center">
                    <Icon name="credit_card" size={20} className="text-purple-600" />
                  </div>
                  <div>
                    <p className="text-sm font-medium text-text-main dark:text-white">
                      Metode Pembayaran Terpopuler
                    </p>
                    <p className="text-xs text-text-sub dark:text-gray-400 capitalize">
                      {paymentMethods.data[0]?.method.replace('_', ' ')}
                    </p>
                  </div>
                </div>
                <p className="text-xl font-bold text-text-main dark:text-white">
                  {paymentMethods.data[0]?.count || 0}x
                </p>
              </div>
            )}
          </div>
        </Card>
      </div>
    </div>
  );
}

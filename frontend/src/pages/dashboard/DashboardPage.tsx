import { Card, CardTitle, CardDescription, Badge, Avatar } from '@/components/ui';
import { StatsCard, DataTable, SalesChart } from '@/components/common';
import { useDashboardStats, useRecentOrders, useSalesTrend } from '@/hooks/useDashboard';
import { formatCurrency, formatDate } from '@/utils';

// Type for recent orders from dashboard service
interface RecentOrder {
  id: number;
  orderNumber: string;
  customerName: string;
  createdAt: string;
  status: string;
  total: number;
}

export function DashboardPage() {
  const { data: stats, isLoading: statsLoading } = useDashboardStats();
  const { data: recentOrders, isLoading: ordersLoading } = useRecentOrders();
  const { data: salesTrend, isLoading: trendLoading } = useSalesTrend();

  const statsConfig = stats
    ? [
        {
          title: 'Total Penjualan',
          value: formatCurrency(stats.totalSales),
          icon: 'payments',
          trend: {
            value: `${stats.salesGrowth > 0 ? '+' : ''}${stats.salesGrowth}%`,
            isPositive: stats.salesGrowth > 0,
            label: 'dari bulan lalu',
          },
        },
        {
          title: 'Pesanan Baru',
          value: stats.newOrders,
          icon: 'shopping_bag',
          trend: {
            value: `${stats.ordersGrowth > 0 ? '+' : ''}${stats.ordersGrowth}%`,
            isPositive: stats.ordersGrowth > 0,
            label: 'dari bulan lalu',
          },
        },
        {
          title: 'Pelanggan Baru',
          value: stats.newCustomers,
          icon: 'person_add',
          trend: {
            value: `${stats.customersGrowth > 0 ? '+' : ''}${stats.customersGrowth}%`,
            isPositive: stats.customersGrowth > 0,
            label: 'dari bulan lalu',
          },
        },
        {
          title: 'Pendapatan Bersih',
          value: formatCurrency(stats.netIncome),
          icon: 'account_balance_wallet',
          trend: {
            value: `${stats.incomeGrowth > 0 ? '+' : ''}${stats.incomeGrowth}%`,
            isPositive: stats.incomeGrowth > 0,
            label: 'dari bulan lalu',
          },
        },
      ]
    : [];

  const orderColumns = [
    {
      header: 'ID Pesanan',
      accessor: (row: RecentOrder) => (
        <span className="font-medium text-text-main dark:text-white">{row.orderNumber}</span>
      ),
    },
    {
      header: 'Pelanggan',
      accessor: (row: RecentOrder) => (
        <div className="flex items-center gap-2">
          <Avatar name={row.customerName} size="xs" />
          <span className="text-text-main dark:text-gray-200">{row.customerName}</span>
        </div>
      ),
    },
    {
      header: 'Tanggal',
      accessor: (row: RecentOrder) => (
        <span className="text-text-sub dark:text-gray-400">
          {formatDate(row.createdAt)}
        </span>
      ),
    },
    {
      header: 'Status',
      accessor: (row: RecentOrder) => {
        type BadgeVariant = 'default' | 'primary' | 'success' | 'warning' | 'danger' | 'info' | 'gray';
        const statusConfig: Record<string, { variant: BadgeVariant; label: string }> = {
          draft: { variant: 'default', label: 'Draft' },
          confirmed: { variant: 'info', label: 'Dikonfirmasi' },
          completed: { variant: 'success', label: 'Selesai' },
          cancelled: { variant: 'danger', label: 'Dibatalkan' },
          voided: { variant: 'warning', label: 'Void' },
        };
        const config = statusConfig[row.status] || statusConfig.draft;
        return (
          <Badge variant={config.variant} dot>
            {config.label}
          </Badge>
        );
      },
    },
    {
      header: 'Jumlah',
      accessor: (row: RecentOrder) => (
        <span className="font-medium text-text-main dark:text-white">
          {formatCurrency(row.total)}
        </span>
      ),
      className: 'text-right',
    },
  ];

  return (
    <div className="space-y-6">
      {/* Header */}
      <div>
        <h1 className="text-2xl md:text-3xl font-bold text-text-main dark:text-white">
          Selamat Datang kembali! 👋
        </h1>
        <p className="text-text-sub dark:text-gray-400 text-sm mt-1">
          Berikut adalah ringkasan performa bisnis Anda hari ini.
        </p>
      </div>

      {/* Stats Grid */}
      {statsLoading ? (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
          {[...Array(4)].map((_, i) => (
            <div
              key={i}
              className="h-32 bg-card-light dark:bg-card-dark rounded-xl border border-border-light dark:border-border-dark animate-pulse"
            />
          ))}
        </div>
      ) : (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
          {statsConfig.map((stat, index) => (
            <StatsCard key={index} {...stat} />
          ))}
        </div>
      )}

      {/* Sales Trend Chart */}
      <Card variant="bordered" padding="md">
        <div className="mb-4">
          <CardTitle>Tren Penjualan 30 Hari Terakhir</CardTitle>
          <CardDescription className="mt-1">
            Grafik penjualan dan jumlah pesanan dalam 30 hari terakhir
          </CardDescription>
        </div>
        {trendLoading ? (
          <div className="h-[350px] flex items-center justify-center">
            <div className="flex flex-col items-center gap-3">
              <span className="material-symbols-outlined text-4xl animate-spin text-primary">
                progress_activity
              </span>
              <p className="text-text-sub dark:text-gray-400 text-sm">Loading chart...</p>
            </div>
          </div>
        ) : salesTrend && salesTrend.length > 0 ? (
          <SalesChart data={salesTrend} />
        ) : (
          <div className="h-[350px] flex items-center justify-center text-text-sub dark:text-gray-400">
            <p>Tidak ada data tren penjualan</p>
          </div>
        )}
      </Card>

      {/* Recent Orders */}
      <Card variant="bordered" padding="none">
        <div className="p-6 pb-4">
          <div className="flex items-center justify-between">
            <div>
              <CardTitle>Pesanan Terbaru</CardTitle>
              <CardDescription className="mt-1">
                Daftar pesanan terbaru dari pelanggan Anda
              </CardDescription>
            </div>
            <a
              href="/orders"
              className="text-primary hover:text-blue-700 text-sm font-bold flex items-center gap-1 transition-colors"
            >
              Lihat Semua
              <span className="material-symbols-outlined text-[16px]">arrow_forward</span>
            </a>
          </div>
        </div>

        <DataTable
          data={recentOrders || []}
          columns={orderColumns}
          loading={ordersLoading}
          emptyMessage="Belum ada pesanan"
          onRowClick={order => console.log('Order clicked:', order)}
        />
      </Card>
    </div>
  );
}

import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Card, Badge, Button, Dropdown, Icon } from '@/components/ui';
import { usePurchaseOrders } from '@/hooks/usePurchaseOrders';
import { formatCurrency, formatDate } from '@/utils';
import type { PurchaseOrder, POStatus } from '@/types/purchaseOrder';

const STATUS_BADGE_MAP: Record<POStatus, 'gray' | 'warning' | 'info' | 'success' | 'danger'> = {
  draft: 'gray',
  ordered: 'warning',
  partial: 'info',
  received: 'success',
  cancelled: 'danger',
};

const STATUS_LABELS: Record<POStatus, string> = {
  draft: 'Draft',
  ordered: 'Dipesan',
  partial: 'Sebagian',
  received: 'Diterima',
  cancelled: 'Dibatalkan',
};

export function PurchaseOrdersPage() {
  const navigate = useNavigate();
  const [page, setPage] = useState(1);
  const [statusFilter, setStatusFilter] = useState<string>('');

  const { data, isLoading } = usePurchaseOrders({
    page,
    per_page: 10,
    status: statusFilter || undefined,
  });

  const orders = data?.data || [];

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h1 className="text-2xl md:text-3xl font-bold text-text-main dark:text-white">
            Purchase Order
          </h1>
          <p className="text-text-sub dark:text-gray-400 text-sm mt-1">
            Kelola pembelian dan penerimaan barang
          </p>
        </div>
        <Button
          leftIcon={<Icon name="add" size={20} />}
          onClick={() => navigate('/purchase-orders/new')}
        >
          Buat PO Baru
        </Button>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-2 sm:grid-cols-5 gap-4">
        {(['draft', 'ordered', 'partial', 'received', 'cancelled'] as POStatus[]).map(status => (
          <Card key={status} variant="bordered" padding="md">
            <p className="text-xs text-text-sub dark:text-gray-400 mb-1">{STATUS_LABELS[status]}</p>
            <p className="text-2xl font-bold text-text-main dark:text-white">
              {orders.filter(o => o.status === status).length}
            </p>
          </Card>
        ))}
      </div>

      {/* Filters */}
      <Card variant="bordered" padding="md">
        <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
          <Dropdown
            value={statusFilter}
            onChange={v => { setStatusFilter(v); setPage(1); }}
            options={[
              { value: '', label: 'Semua Status' },
              { value: 'draft', label: 'Draft' },
              { value: 'ordered', label: 'Dipesan' },
              { value: 'partial', label: 'Sebagian Diterima' },
              { value: 'received', label: 'Diterima' },
              { value: 'cancelled', label: 'Dibatalkan' },
            ]}
            leftIcon={<Icon name="filter_list" size={20} />}
          />
          <Button
            variant="outline"
            leftIcon={<Icon name="refresh" size={20} />}
            onClick={() => { setStatusFilter(''); setPage(1); }}
          >
            Reset Filter
          </Button>
        </div>
      </Card>

      {/* Table */}
      <Card variant="bordered" padding="none">
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead>
              <tr className="bg-background-light dark:bg-background-dark border-b border-border-light dark:border-border-dark">
                <th className="px-4 py-3 text-left text-xs font-semibold text-text-sub dark:text-gray-400 uppercase tracking-wider">
                  No. PO
                </th>
                <th className="px-4 py-3 text-left text-xs font-semibold text-text-sub dark:text-gray-400 uppercase tracking-wider">
                  Supplier
                </th>
                <th className="px-4 py-3 text-left text-xs font-semibold text-text-sub dark:text-gray-400 uppercase tracking-wider">
                  Status
                </th>
                <th className="px-4 py-3 text-left text-xs font-semibold text-text-sub dark:text-gray-400 uppercase tracking-wider">
                  Total
                </th>
                <th className="px-4 py-3 text-left text-xs font-semibold text-text-sub dark:text-gray-400 uppercase tracking-wider">
                  Tanggal
                </th>
                <th className="px-4 py-3 text-right text-xs font-semibold text-text-sub dark:text-gray-400 uppercase tracking-wider">
                  Aksi
                </th>
              </tr>
            </thead>
            <tbody>
              {isLoading ? (
                <tr>
                  <td colSpan={6} className="px-4 py-12 text-center">
                    <Icon name="progress_activity" size={32} className="animate-spin text-primary mx-auto" />
                    <p className="text-text-sub dark:text-gray-400 mt-2">Memuat data...</p>
                  </td>
                </tr>
              ) : orders.length === 0 ? (
                <tr>
                  <td colSpan={6} className="px-4 py-12 text-center">
                    <Icon name="shopping_cart" size={48} className="text-text-sub dark:text-gray-400 mx-auto mb-2" />
                    <p className="text-text-sub dark:text-gray-400">Belum ada purchase order</p>
                  </td>
                </tr>
              ) : (
                orders.map((order: PurchaseOrder) => (
                  <tr
                    key={order.id}
                    className="hover:bg-background-light dark:hover:bg-background-dark transition-colors border-b border-border-light dark:border-border-dark last:border-0"
                  >
                    <td className="px-4 py-4">
                      <span className="font-mono text-sm font-medium text-primary">{order.po_number}</span>
                    </td>
                    <td className="px-4 py-4">
                      <p className="text-sm text-text-main dark:text-gray-200">
                        {order.supplier?.name || '-'}
                      </p>
                    </td>
                    <td className="px-4 py-4">
                      <Badge variant={STATUS_BADGE_MAP[order.status]} dot size="sm">
                        {STATUS_LABELS[order.status]}
                      </Badge>
                    </td>
                    <td className="px-4 py-4">
                      <span className="text-sm font-medium text-text-main dark:text-gray-200">
                        {formatCurrency(order.total_amount)}
                      </span>
                    </td>
                    <td className="px-4 py-4">
                      <span className="text-sm text-text-sub dark:text-gray-400">
                        {formatDate(new Date(order.created_at), 'dd MMM yyyy')}
                      </span>
                    </td>
                    <td className="px-4 py-4 text-right">
                      <button
                        className="p-1.5 hover:bg-background-light dark:hover:bg-background-dark rounded-lg transition-colors"
                        onClick={() => navigate(`/purchase-orders/${order.id}`)}
                        title="Lihat Detail"
                      >
                        <Icon name="visibility" size={18} className="text-text-sub dark:text-gray-400" />
                      </button>
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>

        {/* Pagination */}
        {data && data.totalPages > 1 && (
          <div className="px-6 py-4 border-t border-border-light dark:border-border-dark">
            <div className="flex items-center justify-between">
              <p className="text-sm text-text-sub dark:text-gray-400">
                Menampilkan {(page - 1) * 10 + 1} - {Math.min(page * 10, data.total)} dari{' '}
                {data.total} PO
              </p>
              <div className="flex items-center gap-2">
                <Button
                  variant="outline"
                  size="sm"
                  disabled={page === 1}
                  onClick={() => setPage(p => p - 1)}
                >
                  <Icon name="chevron_left" size={20} />
                </Button>
                <span className="text-sm text-text-main dark:text-white px-2">
                  {page} / {data.totalPages}
                </span>
                <Button
                  variant="outline"
                  size="sm"
                  disabled={page === data.totalPages}
                  onClick={() => setPage(p => p + 1)}
                >
                  <Icon name="chevron_right" size={20} />
                </Button>
              </div>
            </div>
          </div>
        )}
      </Card>
    </div>
  );
}

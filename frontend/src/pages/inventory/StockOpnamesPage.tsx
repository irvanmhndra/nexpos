import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Card, Badge, Button, Dropdown, Icon } from '@/components/ui';
import { useStockOpnames } from '@/hooks/useStockOpnames';
import { formatCurrency, formatDate } from '@/utils';
import type { StockOpname, StockOpnameStatus } from '@/types/stockOpname';

const STATUS_BADGE_MAP: Record<StockOpnameStatus, 'gray' | 'warning' | 'success' | 'danger'> = {
  in_progress: 'warning',
  completed: 'success',
  cancelled: 'danger',
};

const STATUS_LABELS: Record<StockOpnameStatus, string> = {
  in_progress: 'Berjalan',
  completed: 'Selesai',
  cancelled: 'Dibatalkan',
};

export function StockOpnamesPage() {
  const navigate = useNavigate();
  const [page, setPage] = useState(1);
  const [statusFilter, setStatusFilter] = useState<string>('');

  const { data, isLoading } = useStockOpnames({
    page,
    per_page: 10,
    status: statusFilter || undefined,
  });

  const opnames = data?.data || [];

  return (
    <div className="space-y-6">
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h1 className="text-2xl md:text-3xl font-bold text-text-main dark:text-white">
            Stock Opname
          </h1>
          <p className="text-text-sub dark:text-gray-400 text-sm mt-1">
            Hitung fisik stok dan cocokkan dengan catatan sistem
          </p>
        </div>
        <Button
          leftIcon={<Icon name="add" size={20} />}
          onClick={() => navigate('/stock-opnames/new')}
        >
          Mulai Opname Baru
        </Button>
      </div>

      <div className="grid grid-cols-3 gap-4">
        {(['in_progress', 'completed', 'cancelled'] as StockOpnameStatus[]).map(status => (
          <Card key={status} variant="bordered" padding="md">
            <p className="text-xs text-text-sub dark:text-gray-400 mb-1">{STATUS_LABELS[status]}</p>
            <p className="text-2xl font-bold text-text-main dark:text-white">
              {opnames.filter(o => o.status === status).length}
            </p>
          </Card>
        ))}
      </div>

      <Card variant="bordered" padding="md">
        <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
          <Dropdown
            value={statusFilter}
            onChange={v => { setStatusFilter(v); setPage(1); }}
            options={[
              { value: '', label: 'Semua Status' },
              { value: 'in_progress', label: 'Berjalan' },
              { value: 'completed', label: 'Selesai' },
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

      <Card variant="bordered" padding="none">
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead>
              <tr className="bg-background-light dark:bg-background-dark border-b border-border-light dark:border-border-dark">
                <th className="px-4 py-3 text-left text-xs font-semibold text-text-sub dark:text-gray-400 uppercase tracking-wider">No. Opname</th>
                <th className="px-4 py-3 text-left text-xs font-semibold text-text-sub dark:text-gray-400 uppercase tracking-wider">Status</th>
                <th className="px-4 py-3 text-left text-xs font-semibold text-text-sub dark:text-gray-400 uppercase tracking-wider">Progress</th>
                <th className="px-4 py-3 text-left text-xs font-semibold text-text-sub dark:text-gray-400 uppercase tracking-wider">Selisih Qty</th>
                <th className="px-4 py-3 text-left text-xs font-semibold text-text-sub dark:text-gray-400 uppercase tracking-wider">Selisih Nilai</th>
                <th className="px-4 py-3 text-left text-xs font-semibold text-text-sub dark:text-gray-400 uppercase tracking-wider">Tanggal Mulai</th>
                <th className="px-4 py-3 text-right text-xs font-semibold text-text-sub dark:text-gray-400 uppercase tracking-wider">Aksi</th>
              </tr>
            </thead>
            <tbody>
              {isLoading ? (
                <tr>
                  <td colSpan={7} className="px-4 py-12 text-center">
                    <Icon name="progress_activity" size={32} className="animate-spin text-primary mx-auto" />
                    <p className="text-text-sub dark:text-gray-400 mt-2">Memuat data...</p>
                  </td>
                </tr>
              ) : opnames.length === 0 ? (
                <tr>
                  <td colSpan={7} className="px-4 py-12 text-center">
                    <Icon name="inventory" size={48} className="text-text-sub dark:text-gray-400 mx-auto mb-2" />
                    <p className="text-text-sub dark:text-gray-400">Belum ada stock opname</p>
                  </td>
                </tr>
              ) : (
                opnames.map((opname: StockOpname) => {
                  const progressPct = opname.total_items > 0
                    ? Math.round((opname.counted_items / opname.total_items) * 100)
                    : 0;
                  const varianceColor = opname.total_variance_qty < 0
                    ? 'text-red-600 dark:text-red-400'
                    : opname.total_variance_qty > 0
                      ? 'text-green-600 dark:text-green-400'
                      : 'text-text-sub dark:text-gray-400';
                  return (
                    <tr
                      key={opname.id}
                      className="hover:bg-background-light dark:hover:bg-background-dark transition-colors border-b border-border-light dark:border-border-dark last:border-0"
                    >
                      <td className="px-4 py-4">
                        <span className="font-mono text-sm font-medium text-primary">
                          {opname.opname_number}
                        </span>
                      </td>
                      <td className="px-4 py-4">
                        <Badge variant={STATUS_BADGE_MAP[opname.status]} dot size="sm">
                          {STATUS_LABELS[opname.status]}
                        </Badge>
                      </td>
                      <td className="px-4 py-4">
                        <div className="flex items-center gap-2 min-w-[140px]">
                          <div className="flex-1 h-1.5 bg-background-light dark:bg-background-dark rounded-full overflow-hidden">
                            <div
                              className="h-full bg-primary transition-all"
                              style={{ width: `${progressPct}%` }}
                            />
                          </div>
                          <span className="text-xs text-text-sub dark:text-gray-400">
                            {opname.counted_items}/{opname.total_items}
                          </span>
                        </div>
                      </td>
                      <td className={`px-4 py-4 text-sm font-medium ${varianceColor}`}>
                        {opname.status === 'completed'
                          ? (opname.total_variance_qty > 0 ? '+' : '') + opname.total_variance_qty
                          : '-'}
                      </td>
                      <td className={`px-4 py-4 text-sm font-medium ${varianceColor}`}>
                        {opname.status === 'completed' ? formatCurrency(opname.total_variance_value) : '-'}
                      </td>
                      <td className="px-4 py-4">
                        <span className="text-sm text-text-sub dark:text-gray-400">
                          {formatDate(new Date(opname.started_at), 'dd MMM yyyy HH:mm')}
                        </span>
                      </td>
                      <td className="px-4 py-4 text-right">
                        <button
                          className="p-1.5 hover:bg-background-light dark:hover:bg-background-dark rounded-lg transition-colors"
                          onClick={() => navigate(`/stock-opnames/${opname.id}`)}
                          title="Lihat Detail"
                        >
                          <Icon name="visibility" size={18} className="text-text-sub dark:text-gray-400" />
                        </button>
                      </td>
                    </tr>
                  );
                })
              )}
            </tbody>
          </table>
        </div>

        {data && data.totalPages > 1 && (
          <div className="px-6 py-4 border-t border-border-light dark:border-border-dark">
            <div className="flex items-center justify-between">
              <p className="text-sm text-text-sub dark:text-gray-400">
                Menampilkan {(page - 1) * 10 + 1} - {Math.min(page * 10, data.total)} dari {data.total} opname
              </p>
              <div className="flex items-center gap-2">
                <Button variant="outline" size="sm" disabled={page === 1} onClick={() => setPage(p => p - 1)}>
                  <Icon name="chevron_left" size={20} />
                </Button>
                <span className="text-sm text-text-main dark:text-white px-2">
                  {page} / {data.totalPages}
                </span>
                <Button variant="outline" size="sm" disabled={page === data.totalPages} onClick={() => setPage(p => p + 1)}>
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

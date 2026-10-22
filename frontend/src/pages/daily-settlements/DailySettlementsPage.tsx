import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Card, Badge, Button, Dropdown, Icon } from '@/components/ui';
import { useDailySettlements } from '@/hooks/useDailySettlements';
import { formatCurrency, formatDate } from '@/utils';
import type { DailySettlement, DailySettlementStatus } from '@/types/dailySettlement';

const STATUS_BADGE_MAP: Record<DailySettlementStatus, 'warning' | 'success'> = {
  draft: 'warning',
  finalized: 'success',
};

const STATUS_LABELS: Record<DailySettlementStatus, string> = {
  draft: 'Draft',
  finalized: 'Selesai',
};

export function DailySettlementsPage() {
  const navigate = useNavigate();
  const [page, setPage] = useState(1);
  const [statusFilter, setStatusFilter] = useState<string>('');

  const { data, isLoading } = useDailySettlements({
    page,
    per_page: 10,
    status: statusFilter || undefined,
  });

  const settlements = data?.data || [];

  return (
    <div className="space-y-6">
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h1 className="text-2xl md:text-3xl font-bold text-text-main dark:text-white">
            Rekonsiliasi Kas Harian
          </h1>
          <p className="text-text-sub dark:text-gray-400 text-sm mt-1">
            Cocokkan total penjualan per metode bayar dengan setoran kas / saldo bank tiap hari
          </p>
        </div>
        <Button
          leftIcon={<Icon name="add" size={20} />}
          onClick={() => navigate('/daily-settlements/new')}
        >
          Rekap Hari Baru
        </Button>
      </div>

      <Card variant="bordered" padding="md">
        <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
          <Dropdown
            value={statusFilter}
            onChange={v => { setStatusFilter(v); setPage(1); }}
            options={[
              { value: '', label: 'Semua Status' },
              { value: 'draft', label: 'Draft' },
              { value: 'finalized', label: 'Selesai' },
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
                <th className="px-4 py-3 text-left text-xs font-semibold text-text-sub dark:text-gray-400 uppercase tracking-wider">Tanggal</th>
                <th className="px-4 py-3 text-left text-xs font-semibold text-text-sub dark:text-gray-400 uppercase tracking-wider">Status</th>
                <th className="px-4 py-3 text-right text-xs font-semibold text-text-sub dark:text-gray-400 uppercase tracking-wider">Sales</th>
                <th className="px-4 py-3 text-right text-xs font-semibold text-text-sub dark:text-gray-400 uppercase tracking-wider">Expected</th>
                <th className="px-4 py-3 text-right text-xs font-semibold text-text-sub dark:text-gray-400 uppercase tracking-wider">Actual</th>
                <th className="px-4 py-3 text-right text-xs font-semibold text-text-sub dark:text-gray-400 uppercase tracking-wider">Selisih</th>
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
              ) : settlements.length === 0 ? (
                <tr>
                  <td colSpan={7} className="px-4 py-12 text-center">
                    <Icon name="account_balance" size={48} className="text-text-sub dark:text-gray-400 mx-auto mb-2" />
                    <p className="text-text-sub dark:text-gray-400">Belum ada rekonsiliasi harian</p>
                  </td>
                </tr>
              ) : (
                settlements.map((s: DailySettlement) => {
                  const varianceColor = s.total_variance < 0
                    ? 'text-red-600 dark:text-red-400'
                    : s.total_variance > 0
                      ? 'text-green-600 dark:text-green-400'
                      : 'text-text-sub dark:text-gray-400';
                  return (
                    <tr
                      key={s.id}
                      className="hover:bg-background-light dark:hover:bg-background-dark transition-colors border-b border-border-light dark:border-border-dark last:border-0"
                    >
                      <td className="px-4 py-4">
                        <p className="text-sm font-medium text-text-main dark:text-white">
                          {formatDate(new Date(s.settlement_date + 'T00:00:00'), 'dd MMM yyyy')}
                        </p>
                      </td>
                      <td className="px-4 py-4">
                        <Badge variant={STATUS_BADGE_MAP[s.status]} dot size="sm">
                          {STATUS_LABELS[s.status]}
                        </Badge>
                      </td>
                      <td className="px-4 py-4 text-right text-sm text-text-main dark:text-gray-200">
                        {formatCurrency(s.total_sales)}
                      </td>
                      <td className="px-4 py-4 text-right text-sm text-text-main dark:text-gray-200">
                        {formatCurrency(s.total_expected)}
                      </td>
                      <td className="px-4 py-4 text-right text-sm text-text-main dark:text-gray-200">
                        {formatCurrency(s.total_actual)}
                      </td>
                      <td className={`px-4 py-4 text-right text-sm font-medium ${varianceColor}`}>
                        {(s.total_variance > 0 ? '+' : '') + formatCurrency(s.total_variance)}
                      </td>
                      <td className="px-4 py-4 text-right">
                        <button
                          className="p-1.5 hover:bg-background-light dark:hover:bg-background-dark rounded-lg transition-colors"
                          onClick={() => navigate(`/daily-settlements/${s.id}`)}
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
                Menampilkan {(page - 1) * 10 + 1} - {Math.min(page * 10, data.total)} dari {data.total} rekap
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

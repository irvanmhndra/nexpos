import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Card, Badge, Button, Input, Dropdown, Icon } from '@/components/ui';
import { useStockMovements, useMovementStats } from '@/hooks/useInventory';
import { useBranch } from '@/contexts/BranchContext';
import { MOVEMENT_TYPE_LABELS, ROUTES } from '@/utils/constants';
import { format } from 'date-fns';

const TYPE_BADGE_MAP: Record<string, 'success' | 'danger' | 'info'> = {
  IN: 'success',
  OUT: 'danger',
  ADJUST: 'info',
};

export function StockMovementPage() {
  const navigate = useNavigate();
  const [page, setPage] = useState(1);
  const [search, setSearch] = useState('');
  const [typeFilter, setTypeFilter] = useState('');
  const [startDate, setStartDate] = useState('');
  const [endDate, setEndDate] = useState('');

  const { activeBranch } = useBranch();

  const { data, isLoading } = useStockMovements({
    page,
    per_page: 10,
    search: search || undefined,
    type: typeFilter as '' | 'IN' | 'OUT' | 'ADJUST' || undefined,
    start_date: startDate || undefined,
    end_date: endDate || undefined,
  });

  const { data: stats } = useMovementStats();

  const movements = data?.data || [];

  const typeOptions = [
    { value: '', label: 'Semua Tipe' },
    { value: 'IN', label: 'Stok Masuk' },
    { value: 'OUT', label: 'Stok Keluar' },
    { value: 'ADJUST', label: 'Penyesuaian' },
  ];

  const handleSearch = (value: string) => {
    setSearch(value);
    setPage(1);
  };

  const formatQuantity = (type: string, qty: number) => {
    if (type === 'IN') return `+${qty}`;
    if (type === 'OUT') return `-${qty}`;
    return `${qty}`;
  };

  const qtyColorClass = (type: string) => {
    if (type === 'IN') return 'text-green-600 dark:text-green-400';
    if (type === 'OUT') return 'text-red-600 dark:text-red-400';
    return 'text-blue-600 dark:text-blue-400';
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div className="flex items-center gap-3">
          <Button variant="ghost" size="sm" onClick={() => navigate(ROUTES.INVENTORY)}>
            <Icon name="arrow_back" size={20} />
          </Button>
          <div>
            <h1 className="text-2xl md:text-3xl font-bold text-text-main dark:text-white">Riwayat Pergerakan Stok</h1>
            <p className="text-text-sub dark:text-gray-400 text-sm mt-1">
              {activeBranch ? `Cabang: ${activeBranch.name}` : 'Catatan semua perubahan stok'}
            </p>
          </div>
        </div>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <Card variant="bordered" padding="md">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-xs text-text-sub dark:text-gray-400 mb-1">Total Pergerakan</p>
              <p className="text-2xl font-bold text-text-main dark:text-white">{stats?.total_movements ?? 0}</p>
              <p className="text-xs text-text-sub dark:text-gray-400">Bulan ini</p>
            </div>
            <div className="w-12 h-12 rounded-full bg-blue-50 dark:bg-blue-900/20 flex items-center justify-center">
              <Icon name="swap_vert" size={24} className="text-primary" />
            </div>
          </div>
        </Card>

        <Card variant="bordered" padding="md">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-xs text-text-sub dark:text-gray-400 mb-1">Stok Masuk</p>
              <p className="text-2xl font-bold text-green-600">{stats?.stock_in ?? 0}</p>
            </div>
            <div className="w-12 h-12 rounded-full bg-green-50 dark:bg-green-900/20 flex items-center justify-center">
              <Icon name="arrow_downward" size={24} className="text-green-600" />
            </div>
          </div>
        </Card>

        <Card variant="bordered" padding="md">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-xs text-text-sub dark:text-gray-400 mb-1">Stok Keluar</p>
              <p className="text-2xl font-bold text-red-600">{stats?.stock_out ?? 0}</p>
            </div>
            <div className="w-12 h-12 rounded-full bg-red-50 dark:bg-red-900/20 flex items-center justify-center">
              <Icon name="arrow_upward" size={24} className="text-red-600" />
            </div>
          </div>
        </Card>

        <Card variant="bordered" padding="md">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-xs text-text-sub dark:text-gray-400 mb-1">Penyesuaian</p>
              <p className="text-2xl font-bold text-blue-600">{stats?.adjustments ?? 0}</p>
            </div>
            <div className="w-12 h-12 rounded-full bg-blue-50 dark:bg-blue-900/20 flex items-center justify-center">
              <Icon name="sync_alt" size={24} className="text-blue-600" />
            </div>
          </div>
        </Card>
      </div>

      {/* Filters */}
      <Card variant="bordered" padding="md">
        <div className="grid grid-cols-1 sm:grid-cols-5 gap-4">
          <Input
            placeholder="Cari produk / SKU..."
            leftIcon={<Icon name="search" size={20} />}
            value={search}
            onChange={e => handleSearch(e.target.value)}
          />
          <Input
            type="date"
            value={startDate}
            onChange={e => { setStartDate(e.target.value); setPage(1); }}
          />
          <Input
            type="date"
            value={endDate}
            onChange={e => { setEndDate(e.target.value); setPage(1); }}
          />
          <Dropdown
            value={typeFilter}
            onChange={v => { setTypeFilter(v); setPage(1); }}
            options={typeOptions}
            leftIcon={<Icon name="filter_list" size={20} />}
          />
          <Button
            variant="outline"
            leftIcon={<Icon name="refresh" size={20} />}
            onClick={() => { setSearch(''); setTypeFilter(''); setStartDate(''); setEndDate(''); setPage(1); }}
          >
            Reset
          </Button>
        </div>
      </Card>

      {/* Table */}
      <Card variant="bordered" padding="none">
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead>
              <tr className="bg-background-light dark:bg-background-dark border-b border-border-light dark:border-border-dark">
                <th className="px-4 py-3 text-left text-xs font-semibold text-text-sub dark:text-gray-400 uppercase tracking-wider">Tanggal</th>
                <th className="px-4 py-3 text-left text-xs font-semibold text-text-sub dark:text-gray-400 uppercase tracking-wider">Produk</th>
                <th className="px-4 py-3 text-left text-xs font-semibold text-text-sub dark:text-gray-400 uppercase tracking-wider">SKU</th>
                <th className="px-4 py-3 text-left text-xs font-semibold text-text-sub dark:text-gray-400 uppercase tracking-wider">Tipe</th>
                <th className="px-4 py-3 text-right text-xs font-semibold text-text-sub dark:text-gray-400 uppercase tracking-wider">Jumlah</th>
                <th className="px-4 py-3 text-right text-xs font-semibold text-text-sub dark:text-gray-400 uppercase tracking-wider">Sebelum</th>
                <th className="px-4 py-3 text-right text-xs font-semibold text-text-sub dark:text-gray-400 uppercase tracking-wider">Sesudah</th>
                <th className="px-4 py-3 text-left text-xs font-semibold text-text-sub dark:text-gray-400 uppercase tracking-wider">Catatan</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border-light dark:divide-border-dark">
              {isLoading ? (
                <tr>
                  <td colSpan={8} className="px-4 py-12 text-center">
                    <Icon name="progress_activity" size={32} className="animate-spin text-primary mx-auto" />
                    <p className="text-text-sub dark:text-gray-400 mt-2">Memuat data...</p>
                  </td>
                </tr>
              ) : movements.length === 0 ? (
                <tr>
                  <td colSpan={8} className="px-4 py-12 text-center">
                    <Icon name="history" size={48} className="text-text-sub dark:text-gray-400 mx-auto mb-2" />
                    <p className="text-text-sub dark:text-gray-400">Belum ada riwayat pergerakan stok</p>
                  </td>
                </tr>
              ) : (
                movements.map(movement => (
                  <tr key={movement.id} className="hover:bg-background-light dark:hover:bg-background-dark transition-colors">
                    <td className="px-4 py-4 text-sm text-text-main dark:text-gray-200 whitespace-nowrap">
                      {format(new Date(movement.created_at), 'dd MMM yyyy HH:mm')}
                    </td>
                    <td className="px-4 py-4">
                      <p className="text-sm font-medium text-text-main dark:text-white">{movement.product_name}</p>
                      {movement.variant_name !== 'Default' && (
                        <p className="text-xs text-text-sub dark:text-gray-400">{movement.variant_name}</p>
                      )}
                    </td>
                    <td className="px-4 py-4 font-mono text-xs text-text-sub dark:text-gray-400">{movement.sku}</td>
                    <td className="px-4 py-4">
                      <Badge variant={TYPE_BADGE_MAP[movement.type]} size="sm">
                        {MOVEMENT_TYPE_LABELS[movement.type]}
                      </Badge>
                    </td>
                    <td className={`px-4 py-4 text-right font-medium ${qtyColorClass(movement.type)}`}>
                      {formatQuantity(movement.type, movement.quantity)}
                    </td>
                    <td className="px-4 py-4 text-right text-sm text-text-sub dark:text-gray-400">{movement.stock_before}</td>
                    <td className="px-4 py-4 text-right text-sm text-text-main dark:text-white font-medium">{movement.stock_after}</td>
                    <td className="px-4 py-4 text-sm text-text-sub dark:text-gray-400 max-w-48 truncate">
                      {movement.notes || '-'}
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
                Menampilkan {(page - 1) * 10 + 1} - {Math.min(page * 10, data.total)} dari {data.total} pergerakan
              </p>
              <div className="flex items-center gap-2">
                <Button variant="outline" size="sm" disabled={page === 1} onClick={() => setPage(p => p - 1)}>
                  <Icon name="chevron_left" size={20} />
                </Button>
                <span className="text-sm text-text-main dark:text-white px-2">{page} / {data.totalPages}</span>
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

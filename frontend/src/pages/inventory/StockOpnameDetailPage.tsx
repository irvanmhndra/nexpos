import { useEffect, useMemo, useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import toast from 'react-hot-toast';
import { Card, Badge, Button, Input, Icon, Dropdown } from '@/components/ui';
import {
  useStockOpname,
  useBulkUpdateOpnameItems,
  useCompleteStockOpname,
  useCancelStockOpname,
} from '@/hooks/useStockOpnames';
import { formatCurrency, formatDate, getApiErrorMessage } from '@/utils';
import type { StockOpnameItem, StockOpnameStatus } from '@/types/stockOpname';

const STATUS_BADGE_MAP: Record<StockOpnameStatus, 'warning' | 'success' | 'danger'> = {
  in_progress: 'warning',
  completed: 'success',
  cancelled: 'danger',
};

const STATUS_LABELS: Record<StockOpnameStatus, string> = {
  in_progress: 'Berjalan',
  completed: 'Selesai',
  cancelled: 'Dibatalkan',
};

type FilterMode = 'all' | 'uncounted' | 'variance';

export function StockOpnameDetailPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const opnameId = Number(id);

  const { data: opname, isLoading } = useStockOpname(opnameId);
  const bulkUpdate = useBulkUpdateOpnameItems(opnameId);
  const complete = useCompleteStockOpname(opnameId);
  const cancel = useCancelStockOpname(opnameId);

  // Local count drafts keyed by item id; only dirty values are sent on save
  const [drafts, setDrafts] = useState<Record<number, string>>({});
  const [search, setSearch] = useState('');
  const [filterMode, setFilterMode] = useState<FilterMode>('all');
  const [showCompleteConfirm, setShowCompleteConfirm] = useState(false);
  const [showCancelConfirm, setShowCancelConfirm] = useState(false);

  // Reset drafts when underlying items refresh — keep edits the user hasn't saved yet
  useEffect(() => {
    if (opname?.items) {
      setDrafts(prev => {
        const merged: Record<number, string> = { ...prev };
        for (const item of opname.items!) {
          if (merged[item.id] === undefined && item.counted_stock !== null) {
            merged[item.id] = String(item.counted_stock);
          }
        }
        return merged;
      });
    }
  }, [opname?.items]);

  const items = opname?.items || [];

  const filteredItems = useMemo(() => {
    return items.filter(item => {
      if (search) {
        const q = search.toLowerCase();
        if (!item.product_name.toLowerCase().includes(q) && !item.sku.toLowerCase().includes(q)) {
          return false;
        }
      }
      if (filterMode === 'uncounted') return item.counted_stock === null;
      if (filterMode === 'variance') {
        return item.counted_stock !== null && item.variance_qty !== 0;
      }
      return true;
    });
  }, [items, search, filterMode]);

  const isReadOnly = opname?.status !== 'in_progress';

  const dirtyDrafts = useMemo(() => {
    const result: { item: StockOpnameItem; counted: number }[] = [];
    for (const item of items) {
      const raw = drafts[item.id];
      if (raw === undefined || raw === '') continue;
      const n = parseInt(raw);
      if (Number.isNaN(n) || n < 0) continue;
      if (item.counted_stock === n) continue;
      result.push({ item, counted: n });
    }
    return result;
  }, [items, drafts]);

  const handleSave = async () => {
    if (dirtyDrafts.length === 0) {
      toast('Belum ada perubahan untuk disimpan');
      return;
    }
    try {
      await bulkUpdate.mutateAsync({
        items: dirtyDrafts.map(d => ({ item_id: d.item.id, counted_stock: d.counted })),
      });
      toast.success(`${dirtyDrafts.length} item disimpan`);
    } catch (err) {
      toast.error(getApiErrorMessage(err));
    }
  };

  const handleComplete = async () => {
    try {
      // Save any pending drafts first
      if (dirtyDrafts.length > 0) {
        await bulkUpdate.mutateAsync({
          items: dirtyDrafts.map(d => ({ item_id: d.item.id, counted_stock: d.counted })),
        });
      }
      await complete.mutateAsync({});
      toast.success('Stock opname diselesaikan. Stok disesuaikan.');
      setShowCompleteConfirm(false);
    } catch (err) {
      toast.error(getApiErrorMessage(err));
    }
  };

  const handleCancel = async () => {
    try {
      await cancel.mutateAsync();
      toast.success('Stock opname dibatalkan');
      setShowCancelConfirm(false);
    } catch (err) {
      toast.error(getApiErrorMessage(err));
    }
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-20">
        <Icon name="progress_activity" size={40} className="animate-spin text-primary" />
      </div>
    );
  }

  if (!opname) {
    return (
      <div className="text-center py-20">
        <p className="text-text-sub dark:text-gray-400">Stock opname tidak ditemukan</p>
        <Button className="mt-4" onClick={() => navigate('/stock-opnames')}>Kembali</Button>
      </div>
    );
  }

  const totalItems = opname.total_items;
  const countedItems = items.filter(i => i.counted_stock !== null).length;
  const progressPct = totalItems > 0 ? Math.round((countedItems / totalItems) * 100) : 0;
  const varianceItems = items.filter(i => i.counted_stock !== null && i.variance_qty !== 0).length;

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-start sm:justify-between gap-4">
        <div className="flex items-start gap-3">
          <button
            onClick={() => navigate('/stock-opnames')}
            className="p-2 hover:bg-background-light dark:hover:bg-background-dark rounded-lg transition-colors"
          >
            <Icon name="arrow_back" size={20} className="text-text-sub" />
          </button>
          <div>
            <div className="flex items-center gap-3 flex-wrap">
              <h1 className="text-2xl md:text-3xl font-bold text-text-main dark:text-white">
                {opname.opname_number}
              </h1>
              <Badge variant={STATUS_BADGE_MAP[opname.status]} dot>
                {STATUS_LABELS[opname.status]}
              </Badge>
            </div>
            <p className="text-text-sub dark:text-gray-400 text-sm mt-1">
              Dimulai {formatDate(new Date(opname.started_at), 'dd MMM yyyy HH:mm')}
              {opname.completed_at && ` · Selesai ${formatDate(new Date(opname.completed_at), 'dd MMM yyyy HH:mm')}`}
            </p>
          </div>
        </div>
        {!isReadOnly && (
          <div className="flex items-center gap-2">
            <Button
              variant="outline"
              leftIcon={<Icon name="cancel" size={20} />}
              onClick={() => setShowCancelConfirm(true)}
            >
              Batalkan
            </Button>
            <Button
              variant="outline"
              leftIcon={<Icon name="save" size={20} />}
              onClick={handleSave}
              disabled={bulkUpdate.isPending || dirtyDrafts.length === 0}
            >
              Simpan ({dirtyDrafts.length})
            </Button>
            <Button
              leftIcon={<Icon name="check_circle" size={20} />}
              onClick={() => setShowCompleteConfirm(true)}
            >
              Selesaikan
            </Button>
          </div>
        )}
      </div>

      {/* Summary cards */}
      <div className="grid grid-cols-2 sm:grid-cols-4 gap-4">
        <Card variant="bordered" padding="md">
          <p className="text-xs text-text-sub dark:text-gray-400 mb-1">Progress</p>
          <p className="text-2xl font-bold text-text-main dark:text-white">{progressPct}%</p>
          <p className="text-xs text-text-sub dark:text-gray-400">{countedItems}/{totalItems} item</p>
        </Card>
        <Card variant="bordered" padding="md">
          <p className="text-xs text-text-sub dark:text-gray-400 mb-1">Item dengan Selisih</p>
          <p className="text-2xl font-bold text-text-main dark:text-white">{varianceItems}</p>
        </Card>
        <Card variant="bordered" padding="md">
          <p className="text-xs text-text-sub dark:text-gray-400 mb-1">
            Total Selisih Qty {opname.status === 'completed' ? '' : '(sementara)'}
          </p>
          <p className={`text-2xl font-bold ${
            (opname.status === 'completed' ? opname.total_variance_qty : items.reduce((s, i) => s + i.variance_qty, 0)) < 0
              ? 'text-red-600 dark:text-red-400'
              : 'text-text-main dark:text-white'
          }`}>
            {opname.status === 'completed'
              ? (opname.total_variance_qty > 0 ? '+' : '') + opname.total_variance_qty
              : items.reduce((s, i) => s + i.variance_qty, 0)}
          </p>
        </Card>
        <Card variant="bordered" padding="md">
          <p className="text-xs text-text-sub dark:text-gray-400 mb-1">
            Selisih Nilai {opname.status === 'completed' ? '' : '(sementara)'}
          </p>
          <p className={`text-2xl font-bold ${
            (opname.status === 'completed' ? opname.total_variance_value : items.reduce((s, i) => s + i.variance_value, 0)) < 0
              ? 'text-red-600 dark:text-red-400'
              : 'text-text-main dark:text-white'
          }`}>
            {formatCurrency(
              opname.status === 'completed'
                ? opname.total_variance_value
                : items.reduce((s, i) => s + i.variance_value, 0),
            )}
          </p>
        </Card>
      </div>

      {/* Filters */}
      <Card variant="bordered" padding="md">
        <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
          <Input
            placeholder="Cari produk atau SKU..."
            value={search}
            onChange={e => setSearch(e.target.value)}
            leftIcon={<Icon name="search" size={20} />}
          />
          <Dropdown
            value={filterMode}
            onChange={v => setFilterMode(v as FilterMode)}
            options={[
              { value: 'all', label: `Semua item (${items.length})` },
              { value: 'uncounted', label: `Belum dihitung (${items.length - countedItems})` },
              { value: 'variance', label: `Ada selisih (${varianceItems})` },
            ]}
            leftIcon={<Icon name="filter_list" size={20} />}
          />
          <Button
            variant="outline"
            leftIcon={<Icon name="refresh" size={20} />}
            onClick={() => { setSearch(''); setFilterMode('all'); }}
          >
            Reset Filter
          </Button>
        </div>
      </Card>

      {/* Items table */}
      <Card variant="bordered" padding="none">
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead>
              <tr className="bg-background-light dark:bg-background-dark border-b border-border-light dark:border-border-dark">
                <th className="px-4 py-3 text-left text-xs font-semibold text-text-sub dark:text-gray-400 uppercase tracking-wider">Produk</th>
                <th className="px-4 py-3 text-left text-xs font-semibold text-text-sub dark:text-gray-400 uppercase tracking-wider">SKU</th>
                <th className="px-4 py-3 text-right text-xs font-semibold text-text-sub dark:text-gray-400 uppercase tracking-wider">Stok Sistem</th>
                <th className="px-4 py-3 text-right text-xs font-semibold text-text-sub dark:text-gray-400 uppercase tracking-wider">Hitung Fisik</th>
                <th className="px-4 py-3 text-right text-xs font-semibold text-text-sub dark:text-gray-400 uppercase tracking-wider">Selisih</th>
                <th className="px-4 py-3 text-right text-xs font-semibold text-text-sub dark:text-gray-400 uppercase tracking-wider">Nilai Selisih</th>
              </tr>
            </thead>
            <tbody>
              {filteredItems.length === 0 ? (
                <tr>
                  <td colSpan={6} className="px-4 py-12 text-center">
                    <p className="text-text-sub dark:text-gray-400">Tidak ada item sesuai filter</p>
                  </td>
                </tr>
              ) : (
                filteredItems.map(item => {
                  const draftValue = drafts[item.id] ?? (item.counted_stock !== null ? String(item.counted_stock) : '');
                  const draftNumeric = draftValue === '' ? null : parseInt(draftValue);
                  const previewVariance =
                    !isReadOnly && draftNumeric !== null && !Number.isNaN(draftNumeric)
                      ? draftNumeric - item.system_stock
                      : item.variance_qty;
                  const previewValue =
                    !isReadOnly && draftNumeric !== null && !Number.isNaN(draftNumeric)
                      ? (draftNumeric - item.system_stock) * item.unit_cost
                      : item.variance_value;
                  const varianceColor = previewVariance < 0
                    ? 'text-red-600 dark:text-red-400'
                    : previewVariance > 0
                      ? 'text-green-600 dark:text-green-400'
                      : 'text-text-sub dark:text-gray-400';
                  const isDirty = drafts[item.id] !== undefined && draftNumeric !== item.counted_stock;
                  return (
                    <tr
                      key={item.id}
                      className={`border-b border-border-light dark:border-border-dark last:border-0 ${
                        isDirty ? 'bg-primary/5' : 'hover:bg-background-light dark:hover:bg-background-dark'
                      } transition-colors`}
                    >
                      <td className="px-4 py-3">
                        <p className="text-sm font-medium text-text-main dark:text-white">{item.product_name}</p>
                        <p className="text-xs text-text-sub dark:text-gray-400">{item.variant_name}</p>
                      </td>
                      <td className="px-4 py-3 font-mono text-xs text-text-sub dark:text-gray-400">{item.sku}</td>
                      <td className="px-4 py-3 text-right text-sm text-text-main dark:text-gray-200">
                        {item.system_stock}
                      </td>
                      <td className="px-4 py-3 text-right w-32">
                        {isReadOnly ? (
                          <span className="text-sm text-text-main dark:text-gray-200">
                            {item.counted_stock !== null ? item.counted_stock : '-'}
                          </span>
                        ) : (
                          <Input
                            type="number"
                            min="0"
                            value={draftValue}
                            onChange={e =>
                              setDrafts(prev => ({ ...prev, [item.id]: e.target.value }))
                            }
                            placeholder="-"
                            className="text-right"
                            fullWidth
                          />
                        )}
                      </td>
                      <td className={`px-4 py-3 text-right text-sm font-medium ${varianceColor}`}>
                        {item.counted_stock === null && !drafts[item.id] ? '-'
                          : (previewVariance > 0 ? '+' : '') + previewVariance}
                      </td>
                      <td className={`px-4 py-3 text-right text-sm font-medium ${varianceColor}`}>
                        {item.counted_stock === null && !drafts[item.id] ? '-' : formatCurrency(previewValue)}
                      </td>
                    </tr>
                  );
                })
              )}
            </tbody>
          </table>
        </div>
      </Card>

      {/* Complete confirmation modal */}
      {showCompleteConfirm && (
        <ConfirmModal
          title="Selesaikan Stock Opname?"
          message={
            countedItems < totalItems
              ? `${totalItems - countedItems} item belum dihitung. Item yang belum dihitung tidak akan disesuaikan. Lanjutkan?`
              : 'Sistem akan mencatat penyesuaian stok untuk setiap selisih. Tindakan ini tidak bisa dibatalkan.'
          }
          confirmLabel="Ya, Selesaikan"
          confirmVariant="primary"
          isLoading={complete.isPending || bulkUpdate.isPending}
          onConfirm={handleComplete}
          onCancel={() => setShowCompleteConfirm(false)}
        />
      )}

      {/* Cancel confirmation modal */}
      {showCancelConfirm && (
        <ConfirmModal
          title="Batalkan Stock Opname?"
          message="Stock opname ini akan dibatalkan dan tidak bisa diaktifkan kembali. Hitungan yang sudah masuk akan diabaikan."
          confirmLabel="Ya, Batalkan"
          confirmVariant="danger"
          isLoading={cancel.isPending}
          onConfirm={handleCancel}
          onCancel={() => setShowCancelConfirm(false)}
        />
      )}
    </div>
  );
}

function ConfirmModal({
  title,
  message,
  confirmLabel,
  confirmVariant,
  isLoading,
  onConfirm,
  onCancel,
}: {
  title: string;
  message: string;
  confirmLabel: string;
  confirmVariant: 'primary' | 'danger';
  isLoading: boolean;
  onConfirm: () => void;
  onCancel: () => void;
}) {
  return (
    <div className="fixed inset-0 z-50 overflow-y-auto">
      <div className="fixed inset-0 bg-black/50 transition-opacity" onClick={onCancel} />
      <div className="flex min-h-full items-center justify-center p-4">
        <div className="relative w-full max-w-md bg-white dark:bg-card-dark rounded-2xl shadow-xl">
          <div className="p-6">
            <h2 className="text-xl font-bold text-text-main dark:text-white">{title}</h2>
            <p className="text-sm text-text-sub dark:text-gray-400 mt-2">{message}</p>
          </div>
          <div className="flex items-center justify-end gap-3 p-6 pt-0">
            <Button variant="outline" onClick={onCancel} disabled={isLoading}>
              Tidak
            </Button>
            <Button
              variant={confirmVariant === 'danger' ? 'danger' : 'primary'}
              onClick={onConfirm}
              disabled={isLoading}
            >
              {isLoading ? 'Memproses...' : confirmLabel}
            </Button>
          </div>
        </div>
      </div>
    </div>
  );
}

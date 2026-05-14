import { useEffect, useMemo, useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import toast from 'react-hot-toast';
import { Card, Badge, Button, Input, Icon } from '@/components/ui';
import {
  useDailySettlement,
  useBulkUpdateSettlementItems,
  useFinalizeDailySettlement,
} from '@/hooks/useDailySettlements';
import { formatCurrency, formatDate, getApiErrorMessage } from '@/utils';
import type { DailySettlementStatus } from '@/types/dailySettlement';
import { PAYMENT_METHOD_LABELS } from './paymentMethodLabels';

const STATUS_BADGE_MAP: Record<DailySettlementStatus, 'warning' | 'success'> = {
  draft: 'warning',
  finalized: 'success',
};

const STATUS_LABELS: Record<DailySettlementStatus, string> = {
  draft: 'Draft',
  finalized: 'Selesai',
};

export function DailySettlementDetailPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const settlementId = Number(id);

  const { data: settlement, isLoading } = useDailySettlement(settlementId);
  const bulkUpdate = useBulkUpdateSettlementItems(settlementId);
  const finalize = useFinalizeDailySettlement(settlementId);

  const [drafts, setDrafts] = useState<Record<number, string>>({});
  const [showFinalizeConfirm, setShowFinalizeConfirm] = useState(false);

  useEffect(() => {
    if (settlement?.items) {
      setDrafts(prev => {
        const merged: Record<number, string> = { ...prev };
        for (const item of settlement.items!) {
          if (merged[item.id] === undefined && item.actual_amount !== null) {
            merged[item.id] = String(item.actual_amount);
          }
        }
        return merged;
      });
    }
  }, [settlement?.items]);

  const items = settlement?.items || [];
  const isReadOnly = settlement?.status !== 'draft';

  const dirtyDrafts = useMemo(() => {
    const result: { itemId: number; actual: number }[] = [];
    for (const item of items) {
      const raw = drafts[item.id];
      if (raw === undefined || raw === '') continue;
      const n = parseFloat(raw);
      if (Number.isNaN(n) || n < 0) continue;
      if (item.actual_amount === n) continue;
      result.push({ itemId: item.id, actual: n });
    }
    return result;
  }, [items, drafts]);

  const previewTotals = useMemo(() => {
    let expected = 0;
    let actual = 0;
    let variance = 0;
    for (const item of items) {
      const raw = drafts[item.id];
      const draftActual = raw !== undefined && raw !== ''
        ? parseFloat(raw)
        : item.actual_amount;
      const effectiveActual = Number.isFinite(draftActual as number) ? (draftActual as number) : null;
      expected += item.expected_amount;
      if (effectiveActual != null) {
        actual += effectiveActual;
        variance += effectiveActual - item.expected_amount;
      } else {
        variance += -item.expected_amount;
      }
    }
    return { expected, actual, variance };
  }, [items, drafts]);

  const handleSave = async () => {
    if (dirtyDrafts.length === 0) {
      toast('Belum ada perubahan');
      return;
    }
    try {
      await bulkUpdate.mutateAsync({
        items: dirtyDrafts.map(d => ({ item_id: d.itemId, actual_amount: d.actual })),
      });
      toast.success(`${dirtyDrafts.length} entry disimpan`);
    } catch (err) {
      toast.error(getApiErrorMessage(err));
    }
  };

  const handleFinalize = async () => {
    try {
      if (dirtyDrafts.length > 0) {
        await bulkUpdate.mutateAsync({
          items: dirtyDrafts.map(d => ({ item_id: d.itemId, actual_amount: d.actual })),
        });
      }
      await finalize.mutateAsync({});
      toast.success('Rekap difinalisasi');
      setShowFinalizeConfirm(false);
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

  if (!settlement) {
    return (
      <div className="text-center py-20">
        <p className="text-text-sub dark:text-gray-400">Rekap tidak ditemukan</p>
        <Button className="mt-4" onClick={() => navigate('/daily-settlements')}>Kembali</Button>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex flex-col sm:flex-row sm:items-start sm:justify-between gap-4">
        <div className="flex items-start gap-3">
          <button
            onClick={() => navigate('/daily-settlements')}
            className="p-2 hover:bg-background-light dark:hover:bg-background-dark rounded-lg transition-colors"
          >
            <Icon name="arrow_back" size={20} className="text-text-sub" />
          </button>
          <div>
            <div className="flex items-center gap-3 flex-wrap">
              <h1 className="text-2xl md:text-3xl font-bold text-text-main dark:text-white">
                Rekap {formatDate(new Date(settlement.settlement_date + 'T00:00:00'), 'dd MMM yyyy')}
              </h1>
              <Badge variant={STATUS_BADGE_MAP[settlement.status]} dot>
                {STATUS_LABELS[settlement.status]}
              </Badge>
            </div>
            {settlement.finalized_at && (
              <p className="text-text-sub dark:text-gray-400 text-sm mt-1">
                Difinalisasi {formatDate(new Date(settlement.finalized_at), 'dd MMM yyyy HH:mm')}
              </p>
            )}
          </div>
        </div>
        {!isReadOnly && (
          <div className="flex items-center gap-2">
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
              onClick={() => setShowFinalizeConfirm(true)}
            >
              Finalisasi
            </Button>
          </div>
        )}
      </div>

      <div className="grid grid-cols-2 sm:grid-cols-4 gap-4">
        <Card variant="bordered" padding="md">
          <p className="text-xs text-text-sub dark:text-gray-400 mb-1">Total Sales</p>
          <p className="text-xl font-bold text-text-main dark:text-white">
            {formatCurrency(settlement.total_sales)}
          </p>
        </Card>
        <Card variant="bordered" padding="md">
          <p className="text-xs text-text-sub dark:text-gray-400 mb-1">Expected</p>
          <p className="text-xl font-bold text-text-main dark:text-white">
            {formatCurrency(previewTotals.expected)}
          </p>
        </Card>
        <Card variant="bordered" padding="md">
          <p className="text-xs text-text-sub dark:text-gray-400 mb-1">Actual</p>
          <p className="text-xl font-bold text-text-main dark:text-white">
            {formatCurrency(previewTotals.actual)}
          </p>
        </Card>
        <Card variant="bordered" padding="md">
          <p className="text-xs text-text-sub dark:text-gray-400 mb-1">Selisih</p>
          <p className={`text-xl font-bold ${
            previewTotals.variance < 0
              ? 'text-red-600 dark:text-red-400'
              : previewTotals.variance > 0
                ? 'text-green-600 dark:text-green-400'
                : 'text-text-main dark:text-white'
          }`}>
            {(previewTotals.variance > 0 ? '+' : '') + formatCurrency(previewTotals.variance)}
          </p>
        </Card>
      </div>

      <Card variant="bordered" padding="none">
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead>
              <tr className="bg-background-light dark:bg-background-dark border-b border-border-light dark:border-border-dark">
                <th className="px-4 py-3 text-left text-xs font-semibold text-text-sub dark:text-gray-400 uppercase tracking-wider">Metode</th>
                <th className="px-4 py-3 text-right text-xs font-semibold text-text-sub dark:text-gray-400 uppercase tracking-wider">Expected</th>
                <th className="px-4 py-3 text-right text-xs font-semibold text-text-sub dark:text-gray-400 uppercase tracking-wider">Actual</th>
                <th className="px-4 py-3 text-right text-xs font-semibold text-text-sub dark:text-gray-400 uppercase tracking-wider">Selisih</th>
              </tr>
            </thead>
            <tbody>
              {items.map(item => {
                const draftValue = drafts[item.id] ?? (item.actual_amount != null ? String(item.actual_amount) : '');
                const draftNumeric = draftValue === '' ? null : parseFloat(draftValue);
                const previewVariance =
                  !isReadOnly && draftNumeric !== null && !Number.isNaN(draftNumeric)
                    ? draftNumeric - item.expected_amount
                    : item.variance_amount;
                const varianceColor = previewVariance < 0
                  ? 'text-red-600 dark:text-red-400'
                  : previewVariance > 0
                    ? 'text-green-600 dark:text-green-400'
                    : 'text-text-sub dark:text-gray-400';
                const isDirty = drafts[item.id] !== undefined && draftNumeric !== item.actual_amount;
                return (
                  <tr
                    key={item.id}
                    className={`border-b border-border-light dark:border-border-dark last:border-0 ${
                      isDirty ? 'bg-primary/5' : 'hover:bg-background-light dark:hover:bg-background-dark'
                    } transition-colors`}
                  >
                    <td className="px-4 py-3 text-sm font-medium text-text-main dark:text-white">
                      {PAYMENT_METHOD_LABELS[item.payment_method]}
                    </td>
                    <td className="px-4 py-3 text-right text-sm text-text-main dark:text-gray-200">
                      {formatCurrency(item.expected_amount)}
                    </td>
                    <td className="px-4 py-3 text-right w-44">
                      {isReadOnly ? (
                        <span className="text-sm text-text-main dark:text-gray-200">
                          {item.actual_amount != null ? formatCurrency(item.actual_amount) : '-'}
                        </span>
                      ) : (
                        <Input
                          type="number"
                          min="0"
                          step="any"
                          value={draftValue}
                          onChange={e =>
                            setDrafts(prev => ({ ...prev, [item.id]: e.target.value }))
                          }
                          placeholder="0"
                          className="text-right"
                          fullWidth
                        />
                      )}
                    </td>
                    <td className={`px-4 py-3 text-right text-sm font-medium ${varianceColor}`}>
                      {item.actual_amount == null && !drafts[item.id] ? '-'
                        : (previewVariance > 0 ? '+' : '') + formatCurrency(previewVariance)}
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>
      </Card>

      {settlement.notes && (
        <Card variant="bordered" padding="md">
          <p className="text-xs text-text-sub dark:text-gray-400 mb-1">Catatan</p>
          <p className="text-sm text-text-main dark:text-gray-200 whitespace-pre-wrap">{settlement.notes}</p>
        </Card>
      )}

      {showFinalizeConfirm && (
        <div className="fixed inset-0 z-50 overflow-y-auto">
          <div className="fixed inset-0 bg-black/50 transition-opacity" onClick={() => setShowFinalizeConfirm(false)} />
          <div className="flex min-h-full items-center justify-center p-4">
            <div className="relative w-full max-w-md bg-white dark:bg-card-dark rounded-2xl shadow-xl">
              <div className="p-6">
                <h2 className="text-xl font-bold text-text-main dark:text-white">Finalisasi Rekap?</h2>
                <p className="text-sm text-text-sub dark:text-gray-400 mt-2">
                  Setelah difinalisasi, rekap ini tidak bisa diedit lagi. Pastikan semua aktual sudah diinput.
                </p>
              </div>
              <div className="flex items-center justify-end gap-3 p-6 pt-0">
                <Button
                  variant="outline"
                  onClick={() => setShowFinalizeConfirm(false)}
                  disabled={finalize.isPending || bulkUpdate.isPending}
                >
                  Tidak
                </Button>
                <Button
                  onClick={handleFinalize}
                  disabled={finalize.isPending || bulkUpdate.isPending}
                >
                  {finalize.isPending || bulkUpdate.isPending ? 'Memproses...' : 'Ya, Finalisasi'}
                </Button>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

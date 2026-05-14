import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import toast from 'react-hot-toast';
import { Card, Button, Input, Icon } from '@/components/ui';
import {
  useCreateDailySettlement,
  useDailySettlementReport,
} from '@/hooks/useDailySettlements';
import { useBranch } from '@/contexts/BranchContext';
import { formatCurrency, getApiErrorMessage } from '@/utils';
import { PAYMENT_METHOD_LABELS } from './paymentMethodLabels';

function todayISO(): string {
  const d = new Date();
  const yyyy = d.getFullYear();
  const mm = String(d.getMonth() + 1).padStart(2, '0');
  const dd = String(d.getDate()).padStart(2, '0');
  return `${yyyy}-${mm}-${dd}`;
}

export function DailySettlementNewPage() {
  const navigate = useNavigate();
  const { activeBranch } = useBranch();
  const [date, setDate] = useState(todayISO());
  const [notes, setNotes] = useState('');

  const { data: report, isLoading: reportLoading } = useDailySettlementReport(activeBranch?.id, date);
  const createSettlement = useCreateDailySettlement();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!activeBranch) {
      toast.error('Pilih cabang dulu di sidebar');
      return;
    }
    try {
      const result = await createSettlement.mutateAsync({
        branch_id: activeBranch.id,
        settlement_date: date,
        notes: notes || undefined,
      });
      toast.success('Rekap dibuat. Masukkan setoran aktual.');
      navigate(`/daily-settlements/${result.id}`);
    } catch (err) {
      toast.error(getApiErrorMessage(err));
    }
  };

  return (
    <div className="space-y-6 max-w-3xl">
      <div className="flex items-center gap-4">
        <button
          onClick={() => navigate('/daily-settlements')}
          className="p-2 hover:bg-background-light dark:hover:bg-background-dark rounded-lg transition-colors"
        >
          <Icon name="arrow_back" size={20} className="text-text-sub" />
        </button>
        <div>
          <h1 className="text-2xl md:text-3xl font-bold text-text-main dark:text-white">
            Buat Rekap Harian
          </h1>
          <p className="text-text-sub dark:text-gray-400 text-sm mt-1">
            Sistem snapshot expected per metode bayar, kamu input aktualnya
          </p>
        </div>
      </div>

      <form onSubmit={handleSubmit} className="space-y-6">
        <Card variant="bordered" padding="lg">
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-text-main dark:text-white mb-1">
                Cabang
              </label>
              <div className="px-3 py-2 rounded-lg bg-background-light dark:bg-background-dark border border-border-light dark:border-border-dark text-sm text-text-main dark:text-white">
                {activeBranch ? activeBranch.name : <span className="text-text-sub">Belum dipilih</span>}
              </div>
            </div>
            <Input
              label="Tanggal Rekap"
              type="date"
              value={date}
              onChange={e => setDate(e.target.value)}
              fullWidth
            />
          </div>

          <div className="mt-4">
            <label className="block text-sm font-medium text-text-main dark:text-white mb-1">
              Catatan (opsional)
            </label>
            <textarea
              className="w-full rounded-lg border border-border-light dark:border-border-dark bg-white dark:bg-background-dark px-3 py-2 text-sm text-text-main dark:text-white placeholder:text-text-sub focus:outline-none focus:ring-2 focus:ring-primary"
              rows={2}
              placeholder="Misal: hari libur, ada promo khusus, dsb."
              value={notes}
              onChange={e => setNotes(e.target.value)}
            />
          </div>
        </Card>

        <Card variant="bordered" padding="none">
          <div className="p-4 border-b border-border-light dark:border-border-dark">
            <h2 className="text-lg font-semibold text-text-main dark:text-white">Preview Expected</h2>
            <p className="text-xs text-text-sub dark:text-gray-400 mt-1">
              Dihitung otomatis dari penjualan, refund, dan pengeluaran (cash) untuk tanggal ini
            </p>
          </div>
          {reportLoading ? (
            <div className="p-12 text-center">
              <Icon name="progress_activity" size={32} className="animate-spin text-primary mx-auto" />
            </div>
          ) : !report ? (
            <div className="p-12 text-center text-text-sub dark:text-gray-400">
              Pilih cabang dan tanggal untuk melihat preview
            </div>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full">
                <thead>
                  <tr className="bg-background-light dark:bg-background-dark border-b border-border-light dark:border-border-dark">
                    <th className="px-4 py-3 text-left text-xs font-semibold text-text-sub dark:text-gray-400 uppercase tracking-wider">Metode</th>
                    <th className="px-4 py-3 text-right text-xs font-semibold text-text-sub dark:text-gray-400 uppercase tracking-wider">Sales</th>
                    <th className="px-4 py-3 text-right text-xs font-semibold text-text-sub dark:text-gray-400 uppercase tracking-wider">Refund</th>
                    <th className="px-4 py-3 text-right text-xs font-semibold text-text-sub dark:text-gray-400 uppercase tracking-wider">Expense (kas)</th>
                    <th className="px-4 py-3 text-right text-xs font-semibold text-text-sub dark:text-gray-400 uppercase tracking-wider">Expected</th>
                  </tr>
                </thead>
                <tbody>
                  {report.by_method.map(m => (
                    <tr
                      key={m.payment_method}
                      className="border-b border-border-light dark:border-border-dark last:border-0"
                    >
                      <td className="px-4 py-3 text-sm font-medium text-text-main dark:text-white">
                        {PAYMENT_METHOD_LABELS[m.payment_method]}
                      </td>
                      <td className="px-4 py-3 text-right text-sm text-text-main dark:text-gray-200">
                        {formatCurrency(m.gross_sales)}
                      </td>
                      <td className="px-4 py-3 text-right text-sm text-text-sub dark:text-gray-400">
                        {m.refunds > 0 ? `-${formatCurrency(m.refunds)}` : '-'}
                      </td>
                      <td className="px-4 py-3 text-right text-sm text-text-sub dark:text-gray-400">
                        {m.expenses_out > 0 ? `-${formatCurrency(m.expenses_out)}` : '-'}
                      </td>
                      <td className="px-4 py-3 text-right text-sm font-bold text-text-main dark:text-white">
                        {formatCurrency(m.expected_amount)}
                      </td>
                    </tr>
                  ))}
                </tbody>
                <tfoot>
                  <tr className="bg-background-light dark:bg-background-dark">
                    <td className="px-4 py-3 text-sm font-semibold text-text-sub dark:text-gray-400">
                      Total
                    </td>
                    <td className="px-4 py-3 text-right text-sm font-semibold text-text-main dark:text-white">
                      {formatCurrency(report.total_sales)}
                    </td>
                    <td className="px-4 py-3 text-right text-sm text-text-sub dark:text-gray-400">
                      {report.total_refunds > 0 ? `-${formatCurrency(report.total_refunds)}` : '-'}
                    </td>
                    <td className="px-4 py-3 text-right text-sm text-text-sub dark:text-gray-400">
                      {report.total_expenses > 0 ? `-${formatCurrency(report.total_expenses)}` : '-'}
                    </td>
                    <td className="px-4 py-3 text-right text-sm font-bold text-text-main dark:text-white">
                      {formatCurrency(report.total_expected)}
                    </td>
                  </tr>
                </tfoot>
              </table>
            </div>
          )}
        </Card>

        <div className="flex items-center justify-end gap-3">
          <Button type="button" variant="outline" onClick={() => navigate('/daily-settlements')}>
            Batal
          </Button>
          <Button type="submit" disabled={createSettlement.isPending || !activeBranch}>
            {createSettlement.isPending ? 'Membuat...' : 'Buat Rekap'}
          </Button>
        </div>
      </form>
    </div>
  );
}

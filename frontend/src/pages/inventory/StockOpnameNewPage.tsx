import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import toast from 'react-hot-toast';
import { Card, Button, Icon } from '@/components/ui';
import { useCreateStockOpname } from '@/hooks/useStockOpnames';
import { useBranch } from '@/contexts/BranchContext';
import { getApiErrorMessage } from '@/utils';

export function StockOpnameNewPage() {
  const navigate = useNavigate();
  const { activeBranch } = useBranch();
  const [notes, setNotes] = useState('');

  const createOpname = useCreateStockOpname();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!activeBranch) {
      toast.error('Pilih cabang terlebih dahulu');
      return;
    }
    try {
      const result = await createOpname.mutateAsync({
        branch_id: activeBranch.id,
        notes: notes || undefined,
      });
      toast.success('Stock opname dibuat. Mulai menghitung fisik.');
      navigate(`/stock-opnames/${result.id}`);
    } catch (err) {
      toast.error(getApiErrorMessage(err));
    }
  };

  return (
    <div className="space-y-6 max-w-2xl">
      <div className="flex items-center gap-4">
        <button
          onClick={() => navigate('/stock-opnames')}
          className="p-2 hover:bg-background-light dark:hover:bg-background-dark rounded-lg transition-colors"
        >
          <Icon name="arrow_back" size={20} className="text-text-sub" />
        </button>
        <div>
          <h1 className="text-2xl md:text-3xl font-bold text-text-main dark:text-white">
            Mulai Stock Opname
          </h1>
          <p className="text-text-sub dark:text-gray-400 text-sm mt-1">
            Snapshot stok saat ini sebagai acuan, lalu hitung fisik per item
          </p>
        </div>
      </div>

      <form onSubmit={handleSubmit}>
        <Card variant="bordered" padding="lg">
          <div className="space-y-4">
            <div className="bg-primary/5 border border-primary/10 rounded-lg p-4 flex items-start gap-3">
              <Icon name="info" size={20} className="text-primary mt-0.5 flex-shrink-0" />
              <div className="text-sm">
                <p className="font-medium text-text-main dark:text-white">Cara kerja stock opname</p>
                <ul className="mt-2 space-y-1 text-text-sub dark:text-gray-400 list-disc list-inside">
                  <li>Saat dimulai, sistem akan ambil snapshot stok semua produk aktif di cabang ini.</li>
                  <li>Tim lapangan menghitung fisik dan input angka per item.</li>
                  <li>Saat dikonfirmasi selesai, sistem otomatis catat penyesuaian (movement ADJUST) untuk setiap selisih.</li>
                </ul>
              </div>
            </div>

            <div>
              <label className="block text-sm font-medium text-text-main dark:text-white mb-1">
                Cabang
              </label>
              <div className="px-3 py-2 rounded-lg bg-background-light dark:bg-background-dark border border-border-light dark:border-border-dark text-sm text-text-main dark:text-white">
                {activeBranch ? activeBranch.name : <span className="text-text-sub">Belum dipilih</span>}
              </div>
              <p className="text-xs text-text-sub dark:text-gray-400 mt-1">
                Ganti cabang aktif di sidebar bila ingin opname cabang lain.
              </p>
            </div>

            <div>
              <label className="block text-sm font-medium text-text-main dark:text-white mb-1">
                Catatan (opsional)
              </label>
              <textarea
                className="w-full rounded-lg border border-border-light dark:border-border-dark bg-white dark:bg-background-dark px-3 py-2 text-sm text-text-main dark:text-white placeholder:text-text-sub focus:outline-none focus:ring-2 focus:ring-primary"
                rows={3}
                placeholder="Misal: opname bulanan, atau pengecekan rutin"
                value={notes}
                onChange={e => setNotes(e.target.value)}
              />
            </div>
          </div>
        </Card>

        <div className="flex items-center justify-end gap-3 mt-6">
          <Button type="button" variant="outline" onClick={() => navigate('/stock-opnames')}>
            Batal
          </Button>
          <Button type="submit" disabled={createOpname.isPending || !activeBranch}>
            {createOpname.isPending ? 'Menyiapkan...' : 'Mulai Opname'}
          </Button>
        </div>
      </form>
    </div>
  );
}

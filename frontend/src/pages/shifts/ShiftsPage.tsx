import { useState } from 'react';
import toast from 'react-hot-toast';
import { Card, Badge, Button, Input, Dropdown, Icon } from '@/components/ui';
import { useShifts, useCurrentShift, useOpenShift, useCloseShift } from '@/hooks/useShifts';
import { useBranch } from '@/contexts/BranchContext';
import { formatCurrency, formatDate, getApiErrorMessage } from '@/utils';
import type { Shift, ShiftStatus } from '@/types/shift';

const STATUS_BADGE_MAP: Record<ShiftStatus, 'success' | 'gray'> = {
  open: 'success',
  closed: 'gray',
};

const STATUS_LABELS: Record<ShiftStatus, string> = {
  open: 'Buka',
  closed: 'Tutup',
};

function OpenShiftModal({
  isOpen,
  onClose,
  onSubmit,
  isLoading,
  branchId,
}: {
  isOpen: boolean;
  onClose: () => void;
  onSubmit: (branchId: number, openingFloat: number) => void;
  isLoading: boolean;
  branchId: number;
}) {
  const [openingFloat, setOpeningFloat] = useState('0');

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    onSubmit(branchId, parseFloat(openingFloat) || 0);
  };

  const handleClose = () => {
    setOpeningFloat('0');
    onClose();
  };

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 z-50 overflow-y-auto">
      <div className="fixed inset-0 bg-black/50 transition-opacity" onClick={handleClose} />
      <div className="flex min-h-full items-center justify-center p-4">
        <div className="relative w-full max-w-sm bg-white dark:bg-card-dark rounded-2xl shadow-xl">
          <div className="flex items-center justify-between p-6 border-b border-border-light dark:border-border-dark">
            <div>
              <h2 className="text-xl font-bold text-text-main dark:text-white">Buka Shift</h2>
              <p className="text-sm text-text-sub dark:text-gray-400 mt-1">
                Masukkan uang awal kasir
              </p>
            </div>
            <button
              onClick={handleClose}
              className="p-2 hover:bg-gray-100 dark:hover:bg-gray-800 rounded-lg transition-colors"
            >
              <Icon name="close" size={20} className="text-text-sub" />
            </button>
          </div>

          <form onSubmit={handleSubmit}>
            <div className="p-6">
              <Input
                label="Modal Awal (Opening Float)"
                type="number"
                min="0"
                placeholder="0"
                value={openingFloat}
                onChange={e => setOpeningFloat(e.target.value)}
                fullWidth
                required
              />
            </div>

            <div className="flex items-center justify-end gap-3 p-6 border-t border-border-light dark:border-border-dark">
              <Button type="button" variant="outline" onClick={handleClose}>
                Batal
              </Button>
              <Button type="submit" disabled={isLoading}>
                {isLoading ? 'Membuka...' : 'Buka Shift'}
              </Button>
            </div>
          </form>
        </div>
      </div>
    </div>
  );
}

function CloseShiftModal({
  isOpen,
  onClose,
  onSubmit,
  isLoading,
  shift,
}: {
  isOpen: boolean;
  onClose: () => void;
  onSubmit: (actualCash: number, notes?: string) => void;
  isLoading: boolean;
  shift: Shift;
}) {
  const [actualCash, setActualCash] = useState('');
  const [notes, setNotes] = useState('');

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    onSubmit(parseFloat(actualCash) || 0, notes || undefined);
  };

  const handleClose = () => {
    setActualCash('');
    setNotes('');
    onClose();
  };

  if (!isOpen) return null;

  const expectedCash = shift.expected_cash ?? 0;
  const actualNum = parseFloat(actualCash) || 0;
  const difference = actualNum - expectedCash;

  return (
    <div className="fixed inset-0 z-50 overflow-y-auto">
      <div className="fixed inset-0 bg-black/50 transition-opacity" onClick={handleClose} />
      <div className="flex min-h-full items-center justify-center p-4">
        <div className="relative w-full max-w-sm bg-white dark:bg-card-dark rounded-2xl shadow-xl">
          <div className="flex items-center justify-between p-6 border-b border-border-light dark:border-border-dark">
            <div>
              <h2 className="text-xl font-bold text-text-main dark:text-white">Tutup Shift</h2>
              <p className="text-sm text-text-sub dark:text-gray-400 mt-1">
                Hitung dan tutup shift kasir
              </p>
            </div>
            <button
              onClick={handleClose}
              className="p-2 hover:bg-gray-100 dark:hover:bg-gray-800 rounded-lg transition-colors"
            >
              <Icon name="close" size={20} className="text-text-sub" />
            </button>
          </div>

          <form onSubmit={handleSubmit}>
            <div className="p-6 space-y-4">
              <div className="bg-background-light dark:bg-background-dark rounded-lg p-4 space-y-2">
                <div className="flex justify-between text-sm">
                  <span className="text-text-sub dark:text-gray-400">Modal Awal</span>
                  <span className="font-medium text-text-main dark:text-white">
                    {formatCurrency(shift.opening_float)}
                  </span>
                </div>
                <div className="flex justify-between text-sm">
                  <span className="text-text-sub dark:text-gray-400">Kas yang Diharapkan</span>
                  <span className="font-medium text-text-main dark:text-white">
                    {formatCurrency(expectedCash)}
                  </span>
                </div>
                {actualCash && (
                  <div className="flex justify-between text-sm border-t border-border-light dark:border-border-dark pt-2">
                    <span className="text-text-sub dark:text-gray-400">Selisih</span>
                    <span
                      className={`font-bold ${
                        difference >= 0 ? 'text-green-600' : 'text-red-600'
                      }`}
                    >
                      {difference >= 0 ? '+' : ''}
                      {formatCurrency(difference)}
                    </span>
                  </div>
                )}
              </div>

              <Input
                label="Kas Aktual"
                type="number"
                min="0"
                placeholder="0"
                value={actualCash}
                onChange={e => setActualCash(e.target.value)}
                fullWidth
                required
              />

              <div>
                <label className="block text-sm font-medium text-text-main dark:text-white mb-1">
                  Catatan
                </label>
                <textarea
                  className="w-full rounded-lg border border-border-light dark:border-border-dark bg-white dark:bg-background-dark px-3 py-2 text-sm text-text-main dark:text-white placeholder:text-text-sub focus:outline-none focus:ring-2 focus:ring-primary"
                  rows={2}
                  placeholder="Catatan penutupan (opsional)"
                  value={notes}
                  onChange={e => setNotes(e.target.value)}
                />
              </div>
            </div>

            <div className="flex items-center justify-end gap-3 p-6 border-t border-border-light dark:border-border-dark">
              <Button type="button" variant="outline" onClick={handleClose}>
                Batal
              </Button>
              <Button type="submit" disabled={isLoading}>
                {isLoading ? 'Menutup...' : 'Tutup Shift'}
              </Button>
            </div>
          </form>
        </div>
      </div>
    </div>
  );
}

export function ShiftsPage() {
  const [page, setPage] = useState(1);
  const [statusFilter, setStatusFilter] = useState<string>('');
  const [dateFrom, setDateFrom] = useState('');
  const [dateTo, setDateTo] = useState('');
  const [isOpenShiftModalOpen, setIsOpenShiftModalOpen] = useState(false);
  const [isCloseShiftModalOpen, setIsCloseShiftModalOpen] = useState(false);

  const { activeBranch } = useBranch();

  const { data, isLoading } = useShifts({
    page,
    per_page: 10,
    status: statusFilter || undefined,
    date_from: dateFrom || undefined,
    date_to: dateTo || undefined,
  });

  const { data: currentShift, refetch: refetchCurrentShift } = useCurrentShift();
  const openShift = useOpenShift();
  const closeShift = useCloseShift();

  const shifts = data?.data || [];

  const handleOpenShift = async (branchId: number, openingFloat: number) => {
    try {
      await openShift.mutateAsync({ branch_id: branchId, opening_float: openingFloat });
      toast.success('Shift berhasil dibuka');
      setIsOpenShiftModalOpen(false);
      refetchCurrentShift();
    } catch (err) {
      toast.error(getApiErrorMessage(err));
    }
  };

  const handleCloseShift = async (actualCash: number, notes?: string) => {
    if (!currentShift) return;
    try {
      await closeShift.mutateAsync({
        id: currentShift.id,
        request: { actual_cash: actualCash, notes },
      });
      toast.success('Shift berhasil ditutup');
      setIsCloseShiftModalOpen(false);
      refetchCurrentShift();
    } catch (err) {
      toast.error(getApiErrorMessage(err));
    }
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h1 className="text-2xl md:text-3xl font-bold text-text-main dark:text-white">
            Shift Kasir
          </h1>
          <p className="text-text-sub dark:text-gray-400 text-sm mt-1">
            Kelola shift buka dan tutup kasir
          </p>
        </div>
        <div className="flex items-center gap-3">
          {currentShift?.status === 'open' ? (
            <Button
              variant="danger"
              leftIcon={<Icon name="lock" size={20} />}
              onClick={() => setIsCloseShiftModalOpen(true)}
            >
              Tutup Shift
            </Button>
          ) : (
            <Button
              leftIcon={<Icon name="lock_open" size={20} />}
              onClick={() => setIsOpenShiftModalOpen(true)}
              disabled={!activeBranch}
            >
              Buka Shift
            </Button>
          )}
        </div>
      </div>

      {/* Current Shift Card */}
      {currentShift?.status === 'open' && (
        <Card variant="bordered" padding="md">
          <div className="flex items-center gap-3">
            <div className="w-12 h-12 rounded-full bg-green-50 dark:bg-green-900/20 flex items-center justify-center flex-shrink-0">
              <Icon name="lock_open" size={24} className="text-green-600" />
            </div>
            <div className="flex-1">
              <div className="flex items-center gap-2">
                <p className="font-semibold text-text-main dark:text-white">Shift Sedang Berjalan</p>
                <Badge variant="success" dot size="sm">Aktif</Badge>
              </div>
              <p className="text-sm text-text-sub dark:text-gray-400">
                Dibuka {formatDate(new Date(currentShift.opened_at), 'dd MMM yyyy HH:mm')} &bull; Modal
                awal {formatCurrency(currentShift.opening_float)}
              </p>
            </div>
          </div>
        </Card>
      )}

      {/* Stats */}
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
        <Card variant="bordered" padding="md">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-xs text-text-sub dark:text-gray-400 mb-1">Total Shift</p>
              <p className="text-2xl font-bold text-text-main dark:text-white">{data?.total || 0}</p>
            </div>
            <div className="w-12 h-12 rounded-full bg-blue-50 dark:bg-blue-900/20 flex items-center justify-center">
              <Icon name="schedule" size={24} className="text-primary" />
            </div>
          </div>
        </Card>

        <Card variant="bordered" padding="md">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-xs text-text-sub dark:text-gray-400 mb-1">Shift Terbuka</p>
              <p className="text-2xl font-bold text-green-600">
                {shifts.filter(s => s.status === 'open').length}
              </p>
            </div>
            <div className="w-12 h-12 rounded-full bg-green-50 dark:bg-green-900/20 flex items-center justify-center">
              <Icon name="lock_open" size={24} className="text-green-600" />
            </div>
          </div>
        </Card>

        <Card variant="bordered" padding="md">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-xs text-text-sub dark:text-gray-400 mb-1">Shift Tertutup</p>
              <p className="text-2xl font-bold text-text-main dark:text-white">
                {shifts.filter(s => s.status === 'closed').length}
              </p>
            </div>
            <div className="w-12 h-12 rounded-full bg-gray-50 dark:bg-gray-800 flex items-center justify-center">
              <Icon name="lock" size={24} className="text-gray-600" />
            </div>
          </div>
        </Card>
      </div>

      {/* Filters */}
      <Card variant="bordered" padding="md">
        <div className="grid grid-cols-1 sm:grid-cols-4 gap-4">
          <Dropdown
            value={statusFilter}
            onChange={v => { setStatusFilter(v); setPage(1); }}
            options={[
              { value: '', label: 'Semua Status' },
              { value: 'open', label: 'Terbuka' },
              { value: 'closed', label: 'Tertutup' },
            ]}
            leftIcon={<Icon name="filter_list" size={20} />}
          />
          <Input
            type="date"
            label=""
            placeholder="Dari tanggal"
            value={dateFrom}
            onChange={e => { setDateFrom(e.target.value); setPage(1); }}
            fullWidth
          />
          <Input
            type="date"
            label=""
            placeholder="Sampai tanggal"
            value={dateTo}
            onChange={e => { setDateTo(e.target.value); setPage(1); }}
            fullWidth
          />
          <Button
            variant="outline"
            leftIcon={<Icon name="refresh" size={20} />}
            onClick={() => {
              setStatusFilter('');
              setDateFrom('');
              setDateTo('');
              setPage(1);
            }}
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
                  Kasir
                </th>
                <th className="px-4 py-3 text-left text-xs font-semibold text-text-sub dark:text-gray-400 uppercase tracking-wider">
                  Status
                </th>
                <th className="px-4 py-3 text-left text-xs font-semibold text-text-sub dark:text-gray-400 uppercase tracking-wider">
                  Modal Awal
                </th>
                <th className="px-4 py-3 text-left text-xs font-semibold text-text-sub dark:text-gray-400 uppercase tracking-wider">
                  Kas Aktual
                </th>
                <th className="px-4 py-3 text-left text-xs font-semibold text-text-sub dark:text-gray-400 uppercase tracking-wider">
                  Selisih
                </th>
                <th className="px-4 py-3 text-left text-xs font-semibold text-text-sub dark:text-gray-400 uppercase tracking-wider">
                  Dibuka
                </th>
                <th className="px-4 py-3 text-left text-xs font-semibold text-text-sub dark:text-gray-400 uppercase tracking-wider">
                  Ditutup
                </th>
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
              ) : shifts.length === 0 ? (
                <tr>
                  <td colSpan={7} className="px-4 py-12 text-center">
                    <Icon name="schedule" size={48} className="text-text-sub dark:text-gray-400 mx-auto mb-2" />
                    <p className="text-text-sub dark:text-gray-400">Belum ada data shift</p>
                  </td>
                </tr>
              ) : (
                shifts.map((shift: Shift) => {
                  const diff = shift.cash_difference;
                  return (
                    <tr
                      key={shift.id}
                      className="hover:bg-background-light dark:hover:bg-background-dark transition-colors border-b border-border-light dark:border-border-dark last:border-0"
                    >
                      <td className="px-4 py-4">
                        <p className="text-sm font-medium text-text-main dark:text-white">
                          {shift.cashier_name || `Kasir #${shift.cashier_id}`}
                        </p>
                      </td>
                      <td className="px-4 py-4">
                        <Badge variant={STATUS_BADGE_MAP[shift.status]} dot size="sm">
                          {STATUS_LABELS[shift.status]}
                        </Badge>
                      </td>
                      <td className="px-4 py-4 text-sm text-text-main dark:text-gray-200">
                        {formatCurrency(shift.opening_float)}
                      </td>
                      <td className="px-4 py-4 text-sm text-text-main dark:text-gray-200">
                        {shift.actual_cash != null ? formatCurrency(shift.actual_cash) : '-'}
                      </td>
                      <td className="px-4 py-4 text-sm">
                        {diff != null ? (
                          <span className={diff >= 0 ? 'text-green-600' : 'text-red-600'}>
                            {diff >= 0 ? '+' : ''}
                            {formatCurrency(diff)}
                          </span>
                        ) : (
                          '-'
                        )}
                      </td>
                      <td className="px-4 py-4 text-sm text-text-sub dark:text-gray-400">
                        {formatDate(new Date(shift.opened_at), 'dd MMM yyyy HH:mm')}
                      </td>
                      <td className="px-4 py-4 text-sm text-text-sub dark:text-gray-400">
                        {shift.closed_at
                          ? formatDate(new Date(shift.closed_at), 'dd MMM yyyy HH:mm')
                          : '-'}
                      </td>
                    </tr>
                  );
                })
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
                {data.total} shift
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

      {/* Open Shift Modal */}
      {activeBranch && (
        <OpenShiftModal
          isOpen={isOpenShiftModalOpen}
          onClose={() => setIsOpenShiftModalOpen(false)}
          onSubmit={handleOpenShift}
          isLoading={openShift.isPending}
          branchId={activeBranch.id}
        />
      )}

      {/* Close Shift Modal */}
      {currentShift && currentShift.status === 'open' && (
        <CloseShiftModal
          isOpen={isCloseShiftModalOpen}
          onClose={() => setIsCloseShiftModalOpen(false)}
          onSubmit={handleCloseShift}
          isLoading={closeShift.isPending}
          shift={currentShift}
        />
      )}
    </div>
  );
}

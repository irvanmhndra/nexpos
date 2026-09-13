import { useState } from 'react';
import { Button, Icon, Input, Badge } from '@/components/ui';
import { DeleteConfirmModal } from '@/components/common';
import {
  usePromotions,
  useCreatePromotion,
  useUpdatePromotion,
  useDeletePromotion,
} from '@/hooks/usePromotions';
import type { Promotion, CreatePromotionRequest, PromotionType, DiscountType, PromotionStatus } from '@/types/promotion';

const PROMO_TYPES: { value: PromotionType; label: string }[] = [
  { value: 'discount', label: 'Diskon' },
  { value: 'bundle', label: 'Paket' },
  { value: 'conditional', label: 'Bersyarat' },
];

// Derived status → label + badge color. Comes from the API (server time), so an
// expired promo shows "Kadaluarsa" even if is_active is still true.
const PROMO_STATUS: { value: PromotionStatus; label: string; variant: 'success' | 'info' | 'warning' | 'gray' }[] = [
  { value: 'active', label: 'Berjalan', variant: 'success' },
  { value: 'scheduled', label: 'Terjadwal', variant: 'info' },
  { value: 'expired', label: 'Kadaluarsa', variant: 'warning' },
  { value: 'inactive', label: 'Nonaktif', variant: 'gray' },
];

const DISCOUNT_TYPES: { value: DiscountType; label: string }[] = [
  { value: 'percentage', label: 'Persentase (%)' },
  { value: 'fixed', label: 'Nominal (Rp)' },
];

interface PromoFormData {
  code: string;
  name: string;
  description: string;
  type: PromotionType;
  discount_type: DiscountType;
  discount_value: string;
  min_purchase: string;
  max_discount: string;
  start_at: string;
  end_at: string;
  priority: string;
  is_active: boolean;
}

const EMPTY_FORM: PromoFormData = {
  code: '',
  name: '',
  description: '',
  type: 'discount',
  discount_type: 'percentage',
  discount_value: '',
  min_purchase: '',
  max_discount: '',
  start_at: new Date().toISOString().slice(0, 16),
  end_at: '',
  priority: '0',
  is_active: true,
};

function PromoFormModal({
  isOpen,
  onClose,
  onSubmit,
  initialData,
  isLoading,
}: {
  isOpen: boolean;
  onClose: () => void;
  onSubmit: (data: CreatePromotionRequest) => void;
  initialData?: Promotion;
  isLoading: boolean;
}) {
  const [form, setForm] = useState<PromoFormData>(
    initialData
      ? {
          code: initialData.code,
          name: initialData.name,
          description: initialData.description ?? '',
          type: initialData.type,
          discount_type: initialData.discount_type ?? 'percentage',
          discount_value: initialData.discount_value?.toString() ?? '',
          min_purchase: initialData.min_purchase?.toString() ?? '',
          max_discount: initialData.max_discount?.toString() ?? '',
          start_at: initialData.start_at.slice(0, 16),
          end_at: initialData.end_at ? initialData.end_at.slice(0, 16) : '',
          priority: initialData.priority.toString(),
          is_active: initialData.is_active,
        }
      : EMPTY_FORM
  );

  if (!isOpen) return null;

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    onSubmit({
      code: form.code,
      name: form.name,
      description: form.description || null,
      type: form.type,
      discount_type: form.type === 'discount' ? form.discount_type : null,
      discount_value: form.discount_value ? parseFloat(form.discount_value) : null,
      min_purchase: form.min_purchase ? parseFloat(form.min_purchase) : null,
      max_discount: form.max_discount ? parseFloat(form.max_discount) : null,
      start_at: new Date(form.start_at).toISOString(),
      end_at: form.end_at ? new Date(form.end_at).toISOString() : null,
      priority: parseInt(form.priority) || 0,
      is_active: form.is_active,
    });
  };

  const set =
    (key: keyof PromoFormData) => (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement | HTMLTextAreaElement>) =>
      setForm(prev => ({ ...prev, [key]: e.target.value }));

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 overflow-y-auto">
      <div className="absolute inset-0 bg-black/50" onClick={onClose} />
      <div className="relative bg-white dark:bg-card-dark rounded-xl shadow-xl w-full max-w-2xl my-8 overflow-hidden">
        {/* Header */}
        <div className="flex items-center justify-between px-6 py-5 border-b border-border-light dark:border-border-dark">
          <h3 className="text-lg font-bold text-text-main dark:text-white">
            {initialData ? 'Edit Promo' : 'Tambah Promo Baru'}
          </h3>
          <button
            onClick={onClose}
            className="size-8 flex items-center justify-center rounded-lg text-text-sub hover:bg-background-light dark:hover:bg-background-dark transition-colors"
          >
            <span className="material-symbols-outlined text-[20px]">close</span>
          </button>
        </div>

        {/* Form */}
        <form onSubmit={handleSubmit}>
          <div className="px-6 py-5 space-y-4 max-h-[70vh] overflow-y-auto">
            <div className="grid grid-cols-2 gap-4">
              <Input label="Kode Promo" placeholder="DISCOUNT50" value={form.code} onChange={set('code')} required fullWidth />
              <Input label="Nama Promo" placeholder="Diskon 50%" value={form.name} onChange={set('name')} required fullWidth />
            </div>

            <div className="flex flex-col gap-1.5">
              <label className="text-sm font-bold text-text-main dark:text-gray-200">Deskripsi (Opsional)</label>
              <textarea
                value={form.description}
                onChange={set('description')}
                placeholder="Deskripsi promo..."
                rows={2}
                className="w-full rounded-lg border border-border-light dark:border-gray-600 bg-white dark:bg-card-dark text-text-main dark:text-white focus:border-primary focus:ring-1 focus:ring-primary p-3 text-sm placeholder:text-text-sub dark:placeholder:text-gray-500 resize-none"
              />
            </div>

            <div className="grid grid-cols-2 gap-4">
              <div className="flex flex-col gap-1.5">
                <label className="text-sm font-bold text-text-main dark:text-gray-200">Tipe Promo</label>
                <select value={form.type} onChange={set('type')} className="h-11 rounded-lg border border-border-light dark:border-gray-600 bg-white dark:bg-card-dark text-text-main dark:text-white focus:border-primary focus:ring-1 focus:ring-primary px-4 text-sm">
                  {PROMO_TYPES.map(t => (
                    <option key={t.value} value={t.value}>{t.label}</option>
                  ))}
                </select>
              </div>

              {form.type === 'discount' && (
                <div className="flex flex-col gap-1.5">
                  <label className="text-sm font-bold text-text-main dark:text-gray-200">Jenis Diskon</label>
                  <select value={form.discount_type} onChange={set('discount_type')} className="h-11 rounded-lg border border-border-light dark:border-gray-600 bg-white dark:bg-card-dark text-text-main dark:text-white focus:border-primary focus:ring-1 focus:ring-primary px-4 text-sm">
                    {DISCOUNT_TYPES.map(t => (
                      <option key={t.value} value={t.value}>{t.label}</option>
                    ))}
                  </select>
                </div>
              )}
            </div>

            {form.type === 'discount' && (
              <div className="grid grid-cols-3 gap-4">
                <Input label={`Nilai Diskon ${form.discount_type === 'percentage' ? '(%)' : '(Rp)'}`} type="number" step="0.01" placeholder="0" value={form.discount_value} onChange={set('discount_value')} fullWidth />
                <Input label="Min. Pembelian (Rp)" type="number" step="0.01" placeholder="0" value={form.min_purchase} onChange={set('min_purchase')} fullWidth />
                {form.discount_type === 'percentage' && (
                  <Input label="Maks. Diskon (Rp)" type="number" step="0.01" placeholder="0" value={form.max_discount} onChange={set('max_discount')} fullWidth />
                )}
              </div>
            )}

            <div className="grid grid-cols-2 gap-4">
              <Input label="Mulai" type="datetime-local" value={form.start_at} onChange={set('start_at')} required fullWidth />
              <Input label="Berakhir (Opsional)" type="datetime-local" value={form.end_at} onChange={set('end_at')} fullWidth />
            </div>

            <div className="grid grid-cols-2 gap-4">
              <Input label="Prioritas" type="number" placeholder="0" value={form.priority} onChange={set('priority')} fullWidth />
              <div className="flex items-end pb-1">
                <div className="flex items-center gap-3 p-3 rounded-lg border border-border-light dark:border-gray-600">
                  <button
                    type="button"
                    onClick={() => setForm(prev => ({ ...prev, is_active: !prev.is_active }))}
                    className={`relative inline-flex h-6 w-11 flex-shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors ${form.is_active ? 'bg-primary' : 'bg-gray-300 dark:bg-gray-600'}`}
                  >
                    <span className={`pointer-events-none inline-block size-5 transform rounded-full bg-white shadow transition ${form.is_active ? 'translate-x-5' : 'translate-x-0'}`} />
                  </button>
                  <span className="text-sm font-semibold text-text-main dark:text-white">
                    {form.is_active ? 'Aktif' : 'Tidak Aktif'}
                  </span>
                </div>
              </div>
            </div>
          </div>

          {/* Footer */}
          <div className="flex flex-row-reverse gap-3 px-6 py-4 bg-gray-50 dark:bg-background-dark border-t border-border-light dark:border-border-dark">
            <Button type="submit" disabled={isLoading}>
              {isLoading ? 'Menyimpan...' : 'Simpan'}
            </Button>
            <Button type="button" variant="outline" onClick={onClose}>
              Batal
            </Button>
          </div>
        </form>
      </div>
    </div>
  );
}

export function PromotionsPage() {
  const [page, setPage] = useState(1);
  const [search, setSearch] = useState('');
  const [searchInput, setSearchInput] = useState('');
  const [statusFilter, setStatusFilter] = useState('');
  const [typeFilter, setTypeFilter] = useState('');

  const [isFormOpen, setIsFormOpen] = useState(false);
  const [editingPromo, setEditingPromo] = useState<Promotion | null>(null);
  const [deletingPromo, setDeletingPromo] = useState<Promotion | null>(null);

  const PER_PAGE = 10;

  const { data, isLoading } = usePromotions({
    page,
    per_page: PER_PAGE,
    search: search || undefined,
    status: statusFilter || undefined,
    type: typeFilter || undefined,
  });

  const createPromo = useCreatePromotion();
  const updatePromo = useUpdatePromotion();
  const deletePromo = useDeletePromotion();

  const promotions = data?.data ?? [];
  const total = data?.total ?? 0;
  const totalPages = data?.totalPages ?? 1;

  const handleSearch = () => {
    setSearch(searchInput);
    setPage(1);
  };

  const handleFormSubmit = async (formData: CreatePromotionRequest) => {
    if (editingPromo) {
      await updatePromo.mutateAsync({ id: editingPromo.id, promotion: formData });
    } else {
      await createPromo.mutateAsync(formData);
    }
    setIsFormOpen(false);
    setEditingPromo(null);
  };

  const handleEdit = (promo: Promotion) => {
    setEditingPromo(promo);
    setIsFormOpen(true);
  };

  const handleDeleteConfirm = async () => {
    if (!deletingPromo) return;
    await deletePromo.mutateAsync(deletingPromo.id);
    setDeletingPromo(null);
  };

  const startItem = total === 0 ? 0 : (page - 1) * PER_PAGE + 1;
  const endItem = Math.min(page * PER_PAGE, total);

  const formatCurrency = (val: number | null) => val ? `Rp ${val.toLocaleString('id-ID')}` : '-';
  const formatDate = (date: string) => new Date(date).toLocaleDateString('id-ID', { day: 'numeric', month: 'short', year: 'numeric' });

  return (
    <div className="space-y-6">
      {/* Page Heading */}
      <div className="flex flex-col md:flex-row md:items-end justify-between gap-4">
        <div>
          <h1 className="text-2xl md:text-3xl font-bold text-text-main dark:text-white">Manajemen Promo</h1>
          <p className="text-text-sub dark:text-gray-400 text-sm mt-1 max-w-2xl">
            Kelola semua promo, diskon, dan penawaran khusus untuk meningkatkan penjualan Anda.
          </p>
        </div>
        <Button leftIcon={<Icon name="add" size="sm" />} onClick={() => { setEditingPromo(null); setIsFormOpen(true); }} className="whitespace-nowrap">
          Tambah Promo Baru
        </Button>
      </div>

      {/* Filter Panel */}
      <div className="rounded-xl border border-border-light dark:border-border-dark bg-white dark:bg-card-dark shadow-sm p-4 lg:p-5">
        <div className="flex flex-col md:flex-row gap-4 items-end">
          <div className="flex flex-col w-full md:flex-[2] gap-1.5">
            <label className="text-sm font-semibold text-text-main dark:text-gray-200">Cari Promo</label>
            <div className="flex gap-2">
              <div className="relative flex-1">
                <span className="material-symbols-outlined absolute left-3 top-1/2 -translate-y-1/2 text-text-sub text-[20px]">search</span>
                <input
                  type="text"
                  value={searchInput}
                  onChange={e => setSearchInput(e.target.value)}
                  onKeyDown={e => e.key === 'Enter' && handleSearch()}
                  placeholder="Cari kode atau nama promo..."
                  className="w-full h-11 pl-10 pr-4 rounded-lg border border-border-light dark:border-border-dark bg-background-light dark:bg-background-dark text-text-main dark:text-white placeholder:text-text-sub dark:placeholder:text-gray-500 focus:outline-none focus:ring-2 focus:ring-primary/50 focus:border-primary text-sm"
                />
              </div>
              <Button onClick={handleSearch} variant="outline" className="shrink-0">Cari</Button>
            </div>
          </div>

          <div className="flex flex-col w-full md:w-48 gap-1.5">
            <label className="text-sm font-semibold text-text-main dark:text-gray-200">Tipe</label>
            <select value={typeFilter} onChange={e => { setTypeFilter(e.target.value); setPage(1); }} className="h-11 pl-4 pr-10 rounded-lg border border-border-light dark:border-border-dark bg-background-light dark:bg-background-dark text-text-main dark:text-white focus:outline-none focus:ring-2 focus:ring-primary/50 focus:border-primary appearance-none text-sm cursor-pointer">
              <option value="">Semua Tipe</option>
              {PROMO_TYPES.map(t => (
                <option key={t.value} value={t.value}>{t.label}</option>
              ))}
            </select>
          </div>

          <div className="flex flex-col w-full md:w-48 gap-1.5">
            <label className="text-sm font-semibold text-text-main dark:text-gray-200">Status</label>
            <select value={statusFilter} onChange={e => { setStatusFilter(e.target.value); setPage(1); }} className="h-11 pl-4 pr-10 rounded-lg border border-border-light dark:border-border-dark bg-background-light dark:bg-background-dark text-text-main dark:text-white focus:outline-none focus:ring-2 focus:ring-primary/50 focus:border-primary appearance-none text-sm cursor-pointer">
              <option value="">Semua Status</option>
              {PROMO_STATUS.map(s => (
                <option key={s.value} value={s.value}>{s.label}</option>
              ))}
            </select>
          </div>
        </div>
      </div>

      {/* Table */}
      <div className="rounded-xl border border-border-light dark:border-border-dark bg-white dark:bg-card-dark shadow-sm overflow-hidden">
        <div className="overflow-x-auto">
          <table className="w-full text-left border-collapse">
            <thead>
              <tr className="bg-gray-50 dark:bg-[#1e2d3b] border-b border-border-light dark:border-border-dark">
                <th className="p-4 pl-6 text-xs font-bold text-text-sub dark:text-gray-400 uppercase tracking-wider">Promo</th>
                <th className="p-4 text-xs font-bold text-text-sub dark:text-gray-400 uppercase tracking-wider">Tipe</th>
                <th className="p-4 text-xs font-bold text-text-sub dark:text-gray-400 uppercase tracking-wider">Diskon</th>
                <th className="p-4 text-xs font-bold text-text-sub dark:text-gray-400 uppercase tracking-wider">Periode</th>
                <th className="p-4 text-xs font-bold text-text-sub dark:text-gray-400 uppercase tracking-wider">Status</th>
                <th className="p-4 pr-6 text-xs font-bold text-text-sub dark:text-gray-400 uppercase tracking-wider text-right">Aksi</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border-light dark:divide-border-dark">
              {isLoading ? (
                Array.from({ length: 5 }).map((_, i) => (
                  <tr key={i}>
                    <td colSpan={6} className="p-4 pl-6">
                      <div className="h-10 bg-gray-100 dark:bg-gray-800 rounded animate-pulse" />
                    </td>
                  </tr>
                ))
              ) : promotions.length === 0 ? (
                <tr>
                  <td colSpan={6} className="py-16 text-center">
                    <div className="flex flex-col items-center gap-3 text-text-sub dark:text-gray-400">
                      <span className="material-symbols-outlined text-5xl">local_offer</span>
                      <p className="font-medium">Tidak ada promo ditemukan</p>
                    </div>
                  </td>
                </tr>
              ) : (
                promotions.map(promo => (
                  <tr key={promo.id} className="hover:bg-gray-50 dark:hover:bg-gray-800/40 transition-colors">
                    <td className="p-4 pl-6">
                      <div>
                        <p className="text-sm font-bold text-text-main dark:text-white">{promo.name}</p>
                        <p className="text-xs text-text-sub dark:text-gray-400">{promo.code}</p>
                      </div>
                    </td>
                    <td className="p-4">
                      <span className="text-sm text-text-main dark:text-gray-300">
                        {PROMO_TYPES.find(t => t.value === promo.type)?.label}
                      </span>
                    </td>
                    <td className="p-4">
                      {promo.type === 'discount' && promo.discount_value ? (
                        <span className="text-sm font-mono text-text-main dark:text-gray-300">
                          {promo.discount_type === 'percentage' ? `${promo.discount_value}%` : formatCurrency(promo.discount_value)}
                        </span>
                      ) : (
                        <span className="text-xs text-text-sub italic">-</span>
                      )}
                    </td>
                    <td className="p-4">
                      <div className="text-xs text-text-main dark:text-gray-300">
                        <div>{formatDate(promo.start_at)}</div>
                        {promo.end_at && <div className="text-text-sub">s/d {formatDate(promo.end_at)}</div>}
                      </div>
                    </td>
                    <td className="p-4">
                      {(() => {
                        const st = PROMO_STATUS.find(s => s.value === promo.status) ?? PROMO_STATUS[3];
                        return <Badge variant={st.variant} dot>{st.label}</Badge>;
                      })()}
                    </td>
                    <td className="p-4 pr-6 text-right">
                      <div className="flex items-center justify-end gap-1">
                        <button onClick={() => handleEdit(promo)} title="Edit Promo" className="size-8 flex items-center justify-center rounded-lg text-text-sub hover:text-primary hover:bg-blue-50 dark:hover:bg-blue-900/20 transition-all">
                          <span className="material-symbols-outlined text-[18px]">edit</span>
                        </button>
                        <button onClick={() => setDeletingPromo(promo)} title="Hapus Promo" className="size-8 flex items-center justify-center rounded-lg text-text-sub hover:text-red-600 hover:bg-red-50 dark:hover:bg-red-900/20 transition-all">
                          <span className="material-symbols-outlined text-[18px]">delete</span>
                        </button>
                      </div>
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>

        {/* Pagination */}
        {!isLoading && total > 0 && (
          <div className="px-6 py-4 flex items-center justify-between border-t border-border-light dark:border-border-dark bg-gray-50 dark:bg-card-dark/50">
            <p className="text-sm text-text-sub dark:text-gray-400">
              Menampilkan <span className="font-bold text-text-main dark:text-white">{startItem}–{endItem}</span> dari <span className="font-bold text-text-main dark:text-white">{total}</span> promo
            </p>
            <div className="flex items-center gap-1.5">
              <button onClick={() => setPage(p => Math.max(1, p - 1))} disabled={page === 1} className="flex items-center justify-center size-8 rounded-lg border border-border-light dark:border-border-dark bg-white dark:bg-card-dark text-text-sub hover:bg-gray-100 dark:hover:bg-gray-700 disabled:opacity-40 disabled:cursor-not-allowed transition-colors">
                <span className="material-symbols-outlined text-[18px]">chevron_left</span>
              </button>
              {Array.from({ length: Math.min(totalPages, 7) }, (_, i) => i + 1).map(p => (
                <button key={p} onClick={() => setPage(p)} className={`flex items-center justify-center size-8 rounded-lg border text-sm font-medium transition-colors ${p === page ? 'border-primary bg-primary text-white' : 'border-border-light dark:border-border-dark bg-white dark:bg-card-dark text-text-main dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700'}`}>
                  {p}
                </button>
              ))}
              <button onClick={() => setPage(p => Math.min(totalPages, p + 1))} disabled={page === totalPages} className="flex items-center justify-center size-8 rounded-lg border border-border-light dark:border-border-dark bg-white dark:bg-card-dark text-text-sub hover:bg-gray-100 dark:hover:bg-gray-700 disabled:opacity-40 disabled:cursor-not-allowed transition-colors">
                <span className="material-symbols-outlined text-[18px]">chevron_right</span>
              </button>
            </div>
          </div>
        )}
      </div>

      {/* Add / Edit Modal */}
      {/* key forces a remount when switching add/edit (or between promos) so the
          form's useState initializer re-seeds from initialData — otherwise the
          edit modal opens blank. */}
      <PromoFormModal
        key={editingPromo?.id ?? 'new'}
        isOpen={isFormOpen}
        onClose={() => { setIsFormOpen(false); setEditingPromo(null); }}
        onSubmit={handleFormSubmit}
        initialData={editingPromo ?? undefined}
        isLoading={createPromo.isPending || updatePromo.isPending}
      />

      {/* Delete Confirm Modal */}
      <DeleteConfirmModal
        isOpen={!!deletingPromo}
        onClose={() => setDeletingPromo(null)}
        onConfirm={handleDeleteConfirm}
        isLoading={deletePromo.isPending}
        title="Hapus Promo"
        message="Apakah Anda yakin ingin menghapus promo"
        itemName={deletingPromo?.name}
      />
    </div>
  );
}

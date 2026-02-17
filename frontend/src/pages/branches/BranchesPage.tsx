import { useState } from 'react';
import { Button, Icon, Input, Badge } from '@/components/ui';
import { DeleteConfirmModal } from '@/components/common';
import { useBranches, useCreateBranch, useUpdateBranch, useDeleteBranch } from '@/hooks/useBranches';
import type { Branch, BranchStatus, CreateBranchRequest } from '@/types/branch';

// ─── Status helpers ─────────────────────────────────────────────────────────

const STATUS_CONFIG: Record<BranchStatus, { label: string; variant: 'success' | 'default' | 'warning' }> = {
  active: { label: 'Aktif', variant: 'success' },
  inactive: { label: 'Non-Aktif', variant: 'default' },
  maintenance: { label: 'Renovasi', variant: 'warning' },
};

const ICON_COLORS = [
  'bg-blue-100 dark:bg-blue-900/30 text-primary',
  'bg-purple-100 dark:bg-purple-900/30 text-purple-600',
  'bg-orange-100 dark:bg-orange-900/30 text-orange-600',
  'bg-green-100 dark:bg-green-900/30 text-green-600',
  'bg-pink-100 dark:bg-pink-900/30 text-pink-600',
];

function getBranchIconColor(id: string) {
  const sum = id.split('').reduce((acc, c) => acc + c.charCodeAt(0), 0);
  return ICON_COLORS[sum % ICON_COLORS.length];
}

// ─── Form types ──────────────────────────────────────────────────────────────

interface BranchFormData {
  name: string;
  address: string;
  city: string;
  region: string;
  phone: string;
  manager: string;
  status: BranchStatus;
}

const EMPTY_FORM: BranchFormData = {
  name: '',
  address: '',
  city: '',
  region: '',
  phone: '',
  manager: '',
  status: 'active',
};

const REGIONS = ['DKI Jakarta', 'Jawa Barat', 'Jawa Timur', 'Sumatera Utara', 'Bali'];

// ─── Branch Form Modal ───────────────────────────────────────────────────────

function BranchFormModal({
  isOpen,
  onClose,
  onSubmit,
  initialData,
  isLoading,
}: {
  isOpen: boolean;
  onClose: () => void;
  onSubmit: (data: CreateBranchRequest) => void;
  initialData?: Branch;
  isLoading: boolean;
}) {
  const [form, setForm] = useState<BranchFormData>(
    initialData
      ? {
          name: initialData.name,
          address: initialData.address,
          city: initialData.city,
          region: initialData.region,
          phone: initialData.phone,
          manager: initialData.manager,
          status: initialData.status,
        }
      : EMPTY_FORM
  );

  if (!isOpen) return null;

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    onSubmit(form);
  };

  const set = (key: keyof BranchFormData) => (
    e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement | HTMLTextAreaElement>
  ) => setForm(prev => ({ ...prev, [key]: e.target.value }));

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
      <div className="absolute inset-0 bg-black/50" onClick={onClose} />
      <div className="relative bg-white dark:bg-card-dark rounded-xl shadow-xl w-full max-w-lg overflow-hidden">
        {/* Header */}
        <div className="flex items-center justify-between px-6 py-5 border-b border-border-light dark:border-border-dark">
          <h3 className="text-lg font-bold text-text-main dark:text-white">
            {initialData ? 'Edit Cabang' : 'Tambah Cabang Baru'}
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
          <div className="px-6 py-5 space-y-4">
            <Input
              label="Nama Cabang"
              placeholder="Contoh: Cabang Pusat Jakarta"
              value={form.name}
              onChange={set('name')}
              required
              fullWidth
            />

            <div className="flex flex-col gap-1.5">
              <label className="text-sm font-bold text-text-main dark:text-gray-200">
                Alamat Lengkap
              </label>
              <textarea
                value={form.address}
                onChange={set('address')}
                placeholder="Alamat lengkap cabang..."
                rows={3}
                required
                className="w-full rounded-lg border border-border-light dark:border-gray-600 bg-white dark:bg-card-dark text-text-main dark:text-white focus:border-primary focus:ring-1 focus:ring-primary p-3 text-sm placeholder:text-text-sub dark:placeholder:text-gray-500 resize-none"
              />
            </div>

            <div className="grid grid-cols-2 gap-4">
              <Input
                label="Kota"
                placeholder="Contoh: Jakarta Selatan"
                value={form.city}
                onChange={set('city')}
                required
                fullWidth
              />
              <div className="flex flex-col gap-1.5">
                <label className="text-sm font-bold text-text-main dark:text-gray-200">
                  Wilayah / Provinsi
                </label>
                <div className="relative">
                  <select
                    value={form.region}
                    onChange={set('region')}
                    required
                    className="w-full h-11 pl-4 pr-10 rounded-lg border border-border-light dark:border-gray-600 bg-white dark:bg-card-dark text-text-main dark:text-white focus:border-primary focus:ring-1 focus:ring-primary appearance-none text-sm"
                  >
                    <option value="">Pilih Wilayah</option>
                    {REGIONS.map(r => (
                      <option key={r} value={r}>{r}</option>
                    ))}
                  </select>
                  <span className="material-symbols-outlined absolute right-3 top-1/2 -translate-y-1/2 text-text-sub pointer-events-none text-[18px]">
                    expand_more
                  </span>
                </div>
              </div>
            </div>

            <div className="grid grid-cols-2 gap-4">
              <Input
                label="Nomor Telepon"
                placeholder="021-xxxxxxx"
                value={form.phone}
                onChange={set('phone')}
                required
                fullWidth
              />
              <div className="flex flex-col gap-1.5">
                <label className="text-sm font-bold text-text-main dark:text-gray-200">
                  Status
                </label>
                <div className="relative">
                  <select
                    value={form.status}
                    onChange={set('status')}
                    className="w-full h-11 pl-4 pr-10 rounded-lg border border-border-light dark:border-gray-600 bg-white dark:bg-card-dark text-text-main dark:text-white focus:border-primary focus:ring-1 focus:ring-primary appearance-none text-sm"
                  >
                    <option value="active">Aktif</option>
                    <option value="inactive">Non-Aktif</option>
                    <option value="maintenance">Renovasi</option>
                  </select>
                  <span className="material-symbols-outlined absolute right-3 top-1/2 -translate-y-1/2 text-text-sub pointer-events-none text-[18px]">
                    expand_more
                  </span>
                </div>
              </div>
            </div>

            <Input
              label="Manajer Cabang (Opsional)"
              placeholder="Nama penanggung jawab cabang"
              value={form.manager}
              onChange={set('manager')}
              fullWidth
            />
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

// ─── Page ─────────────────────────────────────────────────────────────────────

export function BranchesPage() {
  const [page, setPage] = useState(1);
  const [search, setSearch] = useState('');
  const [searchInput, setSearchInput] = useState('');
  const [statusFilter, setStatusFilter] = useState('');
  const [regionFilter, setRegionFilter] = useState('');

  const [isFormOpen, setIsFormOpen] = useState(false);
  const [editingBranch, setEditingBranch] = useState<Branch | null>(null);
  const [deletingBranch, setDeletingBranch] = useState<Branch | null>(null);

  const PER_PAGE = 5;

  const { data, isLoading } = useBranches({
    page,
    per_page: PER_PAGE,
    search: search || undefined,
    status: statusFilter || undefined,
    region: regionFilter || undefined,
  });

  const createBranch = useCreateBranch();
  const updateBranch = useUpdateBranch();
  const deleteBranch = useDeleteBranch();

  const branches = data?.data ?? [];
  const total = data?.total ?? 0;
  const totalPages = data?.totalPages ?? 1;

  const handleSearch = () => {
    setSearch(searchInput);
    setPage(1);
  };

  const handleStatusFilter = (e: React.ChangeEvent<HTMLSelectElement>) => {
    setStatusFilter(e.target.value);
    setPage(1);
  };

  const handleRegionFilter = (e: React.ChangeEvent<HTMLSelectElement>) => {
    setRegionFilter(e.target.value);
    setPage(1);
  };

  const handleFormSubmit = async (formData: CreateBranchRequest) => {
    if (editingBranch) {
      await updateBranch.mutateAsync({ id: editingBranch.id, branch: formData });
    } else {
      await createBranch.mutateAsync(formData);
    }
    setIsFormOpen(false);
    setEditingBranch(null);
  };

  const handleEdit = (branch: Branch) => {
    setEditingBranch(branch);
    setIsFormOpen(true);
  };

  const handleDeleteConfirm = async () => {
    if (!deletingBranch) return;
    await deleteBranch.mutateAsync(deletingBranch.id);
    setDeletingBranch(null);
  };

  const startItem = (page - 1) * PER_PAGE + 1;
  const endItem = Math.min(page * PER_PAGE, total);

  return (
    <div className="space-y-6">
      {/* Page Heading */}
      <div className="flex flex-col md:flex-row md:items-end justify-between gap-4">
        <div>
          <h1 className="text-2xl md:text-3xl font-bold text-text-main dark:text-white">
            Manajemen Cabang
          </h1>
          <p className="text-text-sub dark:text-gray-400 text-sm mt-1 max-w-2xl">
            Kelola semua lokasi outlet dan cabang bisnis Anda, pantau status operasional, dan atur informasi kontak.
          </p>
        </div>
        <Button
          leftIcon={<Icon name="add" size="sm" />}
          onClick={() => { setEditingBranch(null); setIsFormOpen(true); }}
          className="whitespace-nowrap"
        >
          Tambah Cabang Baru
        </Button>
      </div>

      {/* Filter Panel */}
      <div className="rounded-xl border border-border-light dark:border-border-dark bg-white dark:bg-card-dark shadow-sm p-4 lg:p-5">
        <div className="flex flex-col md:flex-row gap-4 items-end">
          {/* Search */}
          <div className="flex flex-col w-full md:flex-[2] gap-1.5">
            <label className="text-sm font-semibold text-text-main dark:text-gray-200">
              Cari Cabang
            </label>
            <div className="flex gap-2">
              <div className="relative flex-1">
                <span className="material-symbols-outlined absolute left-3 top-1/2 -translate-y-1/2 text-text-sub text-[20px]">
                  search
                </span>
                <input
                  type="text"
                  value={searchInput}
                  onChange={e => setSearchInput(e.target.value)}
                  onKeyDown={e => e.key === 'Enter' && handleSearch()}
                  placeholder="Cari nama cabang, alamat, atau telepon..."
                  className="w-full h-11 pl-10 pr-4 rounded-lg border border-border-light dark:border-border-dark bg-background-light dark:bg-background-dark text-text-main dark:text-white placeholder:text-text-sub dark:placeholder:text-gray-500 focus:outline-none focus:ring-2 focus:ring-primary/50 focus:border-primary text-sm"
                />
              </div>
            </div>
          </div>

          {/* Status Filter */}
          <div className="flex flex-col w-full md:flex-1 gap-1.5">
            <label className="text-sm font-semibold text-text-main dark:text-gray-200">
              Filter Status
            </label>
            <div className="relative">
              <select
                value={statusFilter}
                onChange={handleStatusFilter}
                className="w-full h-11 pl-4 pr-10 rounded-lg border border-border-light dark:border-border-dark bg-background-light dark:bg-background-dark text-text-main dark:text-white focus:outline-none focus:ring-2 focus:ring-primary/50 focus:border-primary appearance-none text-sm cursor-pointer"
              >
                <option value="">Semua Status</option>
                <option value="active">Aktif</option>
                <option value="inactive">Non-Aktif</option>
                <option value="maintenance">Renovasi</option>
              </select>
              <span className="material-symbols-outlined absolute right-3 top-1/2 -translate-y-1/2 text-text-sub pointer-events-none text-[18px]">
                expand_more
              </span>
            </div>
          </div>

          {/* Region Filter */}
          <div className="flex flex-col w-full md:flex-1 gap-1.5">
            <label className="text-sm font-semibold text-text-main dark:text-gray-200">
              Wilayah
            </label>
            <div className="relative">
              <select
                value={regionFilter}
                onChange={handleRegionFilter}
                className="w-full h-11 pl-4 pr-10 rounded-lg border border-border-light dark:border-border-dark bg-background-light dark:bg-background-dark text-text-main dark:text-white focus:outline-none focus:ring-2 focus:ring-primary/50 focus:border-primary appearance-none text-sm cursor-pointer"
              >
                <option value="">Semua Wilayah</option>
                {REGIONS.map(r => (
                  <option key={r} value={r}>{r}</option>
                ))}
              </select>
              <span className="material-symbols-outlined absolute right-3 top-1/2 -translate-y-1/2 text-text-sub pointer-events-none text-[18px]">
                expand_more
              </span>
            </div>
          </div>
        </div>
      </div>

      {/* Table */}
      <div className="rounded-xl border border-border-light dark:border-border-dark bg-white dark:bg-card-dark shadow-sm overflow-hidden">
        <div className="overflow-x-auto">
          <table className="w-full text-left border-collapse">
            <thead>
              <tr className="bg-gray-50 dark:bg-[#1e2d3b] border-b border-border-light dark:border-border-dark">
                <th className="p-4 pl-6 text-xs font-bold text-text-sub dark:text-gray-400 uppercase tracking-wider">
                  Nama Cabang
                </th>
                <th className="p-4 text-xs font-bold text-text-sub dark:text-gray-400 uppercase tracking-wider">
                  Alamat
                </th>
                <th className="p-4 text-xs font-bold text-text-sub dark:text-gray-400 uppercase tracking-wider">
                  Nomor Telepon
                </th>
                <th className="p-4 text-xs font-bold text-text-sub dark:text-gray-400 uppercase tracking-wider">
                  Status
                </th>
                <th className="p-4 pr-6 text-xs font-bold text-text-sub dark:text-gray-400 uppercase tracking-wider text-right">
                  Aksi
                </th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border-light dark:divide-border-dark">
              {isLoading ? (
                Array.from({ length: PER_PAGE }).map((_, i) => (
                  <tr key={i}>
                    <td colSpan={5} className="p-4 pl-6">
                      <div className="h-10 bg-gray-100 dark:bg-gray-800 rounded animate-pulse" />
                    </td>
                  </tr>
                ))
              ) : branches.length === 0 ? (
                <tr>
                  <td colSpan={5} className="py-16 text-center">
                    <div className="flex flex-col items-center gap-3 text-text-sub dark:text-gray-400">
                      <span className="material-symbols-outlined text-5xl">store_mall_directory</span>
                      <p className="font-medium">Tidak ada cabang ditemukan</p>
                    </div>
                  </td>
                </tr>
              ) : (
                branches.map(branch => {
                  const statusCfg = STATUS_CONFIG[branch.status] ?? STATUS_CONFIG.inactive;
                  const iconColor = getBranchIconColor(branch.id);
                  return (
                    <tr
                      key={branch.id}
                      className="hover:bg-gray-50 dark:hover:bg-gray-800/40 transition-colors"
                    >
                      {/* Name + Code */}
                      <td className="p-4 pl-6">
                        <div className="flex items-center gap-3">
                          <div className={`size-10 rounded-lg flex items-center justify-center shrink-0 ${iconColor}`}>
                            <span className="material-symbols-outlined text-[22px]">storefront</span>
                          </div>
                          <div>
                            <p className="text-sm font-bold text-text-main dark:text-white">
                              {branch.name}
                            </p>
                            <p className="text-xs text-text-sub dark:text-gray-400">
                              ID: {branch.code}
                            </p>
                          </div>
                        </div>
                      </td>

                      {/* Address */}
                      <td className="p-4">
                        <p
                          className="text-sm text-text-main dark:text-gray-300 max-w-[260px] truncate"
                          title={branch.address}
                        >
                          {branch.address}
                        </p>
                        {branch.region && (
                          <p className="text-xs text-text-sub dark:text-gray-500 mt-0.5">
                            {branch.city}, {branch.region}
                          </p>
                        )}
                      </td>

                      {/* Phone */}
                      <td className="p-4">
                        <p className="text-sm text-text-main dark:text-gray-300 font-mono">
                          {branch.phone}
                        </p>
                      </td>

                      {/* Status */}
                      <td className="p-4">
                        <Badge variant={statusCfg.variant} dot>
                          {statusCfg.label}
                        </Badge>
                      </td>

                      {/* Actions */}
                      <td className="p-4 pr-6 text-right">
                        <div className="flex items-center justify-end gap-1">
                          <button
                            onClick={() => handleEdit(branch)}
                            title="Edit Cabang"
                            className="size-8 flex items-center justify-center rounded-lg text-text-sub hover:text-primary hover:bg-blue-50 dark:hover:bg-blue-900/20 transition-all"
                          >
                            <span className="material-symbols-outlined text-[18px]">edit</span>
                          </button>
                          <button
                            onClick={() => setDeletingBranch(branch)}
                            title="Hapus Cabang"
                            className="size-8 flex items-center justify-center rounded-lg text-text-sub hover:text-red-600 hover:bg-red-50 dark:hover:bg-red-900/20 transition-all"
                          >
                            <span className="material-symbols-outlined text-[18px]">delete</span>
                          </button>
                        </div>
                      </td>
                    </tr>
                  );
                })
              )}
            </tbody>
          </table>
        </div>

        {/* Pagination */}
        {!isLoading && total > 0 && (
          <div className="px-6 py-4 flex items-center justify-between border-t border-border-light dark:border-border-dark bg-gray-50 dark:bg-card-dark/50">
            <p className="text-sm text-text-sub dark:text-gray-400">
              Menampilkan{' '}
              <span className="font-bold text-text-main dark:text-white">{startItem}–{endItem}</span>
              {' '}dari{' '}
              <span className="font-bold text-text-main dark:text-white">{total}</span> cabang
            </p>
            <div className="flex items-center gap-1.5">
              <button
                onClick={() => setPage(p => Math.max(1, p - 1))}
                disabled={page === 1}
                className="flex items-center justify-center size-8 rounded-lg border border-border-light dark:border-border-dark bg-white dark:bg-card-dark text-text-sub hover:bg-gray-100 dark:hover:bg-gray-700 disabled:opacity-40 disabled:cursor-not-allowed transition-colors"
              >
                <span className="material-symbols-outlined text-[18px]">chevron_left</span>
              </button>

              {Array.from({ length: totalPages }, (_, i) => i + 1).map(p => (
                <button
                  key={p}
                  onClick={() => setPage(p)}
                  className={`flex items-center justify-center size-8 rounded-lg border text-sm font-medium transition-colors ${
                    p === page
                      ? 'border-primary bg-primary text-white'
                      : 'border-border-light dark:border-border-dark bg-white dark:bg-card-dark text-text-main dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700'
                  }`}
                >
                  {p}
                </button>
              ))}

              <button
                onClick={() => setPage(p => Math.min(totalPages, p + 1))}
                disabled={page === totalPages}
                className="flex items-center justify-center size-8 rounded-lg border border-border-light dark:border-border-dark bg-white dark:bg-card-dark text-text-sub hover:bg-gray-100 dark:hover:bg-gray-700 disabled:opacity-40 disabled:cursor-not-allowed transition-colors"
              >
                <span className="material-symbols-outlined text-[18px]">chevron_right</span>
              </button>
            </div>
          </div>
        )}
      </div>

      {/* Add / Edit Modal */}
      <BranchFormModal
        isOpen={isFormOpen}
        onClose={() => { setIsFormOpen(false); setEditingBranch(null); }}
        onSubmit={handleFormSubmit}
        initialData={editingBranch ?? undefined}
        isLoading={createBranch.isPending || updateBranch.isPending}
      />

      {/* Delete Confirm Modal */}
      <DeleteConfirmModal
        isOpen={!!deletingBranch}
        onClose={() => setDeletingBranch(null)}
        onConfirm={handleDeleteConfirm}
        isLoading={deleteBranch.isPending}
        title="Hapus Cabang"
        message="Apakah Anda yakin ingin menghapus cabang"
        itemName={deletingBranch?.name}
      />
    </div>
  );
}

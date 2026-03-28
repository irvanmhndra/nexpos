import { useState } from 'react';
import toast from 'react-hot-toast';
import { Card, Badge, Button, Input, Dropdown, Icon } from '@/components/ui';
import { DataTable } from '@/components/common';
import {
  useSuppliers,
  useCreateSupplier,
  useUpdateSupplier,
  useDeleteSupplier,
} from '@/hooks/useSuppliers';
import { formatDate, getApiErrorMessage } from '@/utils';
import type { Supplier, CreateSupplierRequest, UpdateSupplierRequest } from '@/types/supplier';

interface SupplierFormData {
  name: string;
  code: string;
  contact_name: string;
  phone: string;
  email: string;
  address: string;
  notes: string;
}

interface SupplierFormErrors {
  name?: string;
  code?: string;
  email?: string;
}

function SupplierFormModal({
  isOpen,
  onClose,
  onSubmit,
  isLoading,
  initialData,
}: {
  isOpen: boolean;
  onClose: () => void;
  onSubmit: (data: CreateSupplierRequest) => void;
  isLoading: boolean;
  initialData?: Supplier;
}) {
  const [formData, setFormData] = useState<SupplierFormData>({
    name: initialData?.name ?? '',
    code: initialData?.code ?? '',
    contact_name: initialData?.contact_name ?? '',
    phone: initialData?.phone ?? '',
    email: initialData?.email ?? '',
    address: initialData?.address ?? '',
    notes: initialData?.notes ?? '',
  });
  const [errors, setErrors] = useState<SupplierFormErrors>({});

  const validateForm = (): boolean => {
    const newErrors: SupplierFormErrors = {};

    if (!formData.name.trim()) {
      newErrors.name = 'Nama supplier wajib diisi';
    }

    if (!formData.code.trim()) {
      newErrors.code = 'Kode supplier wajib diisi';
    }

    if (formData.email && !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(formData.email)) {
      newErrors.email = 'Format email tidak valid';
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleChange =
    (field: keyof SupplierFormData) => (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => {
      setFormData(prev => ({ ...prev, [field]: e.target.value }));
      if (errors[field as keyof SupplierFormErrors]) {
        setErrors(prev => ({ ...prev, [field]: undefined }));
      }
    };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (validateForm()) {
      onSubmit({
        name: formData.name,
        code: formData.code,
        contact_name: formData.contact_name || undefined,
        phone: formData.phone || undefined,
        email: formData.email || undefined,
        address: formData.address || undefined,
        notes: formData.notes || undefined,
      });
    }
  };

  const handleClose = () => {
    setFormData({ name: '', code: '', contact_name: '', phone: '', email: '', address: '', notes: '' });
    setErrors({});
    onClose();
  };

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 z-50 overflow-y-auto">
      <div className="fixed inset-0 bg-black/50 transition-opacity" onClick={handleClose} />
      <div className="flex min-h-full items-center justify-center p-4">
        <div className="relative w-full max-w-lg bg-white dark:bg-card-dark rounded-2xl shadow-xl">
          <div className="flex items-center justify-between p-6 border-b border-border-light dark:border-border-dark">
            <div>
              <h2 className="text-xl font-bold text-text-main dark:text-white">
                {initialData ? 'Edit Supplier' : 'Tambah Supplier Baru'}
              </h2>
              <p className="text-sm text-text-sub dark:text-gray-400 mt-1">
                {initialData
                  ? `Edit informasi ${initialData.name}`
                  : 'Isi informasi supplier di bawah ini'}
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
            <div className="p-6 space-y-5">
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                <Input
                  label="Nama Supplier"
                  placeholder="Contoh: PT. Sumber Makmur"
                  value={formData.name}
                  onChange={handleChange('name')}
                  error={errors.name}
                  fullWidth
                  required
                />
                <Input
                  label="Kode Supplier"
                  placeholder="Contoh: SUP001"
                  value={formData.code}
                  onChange={handleChange('code')}
                  error={errors.code}
                  fullWidth
                  required
                />
              </div>

              <Input
                label="Nama Kontak"
                placeholder="Nama penanggung jawab"
                value={formData.contact_name}
                onChange={handleChange('contact_name')}
                helperText="Opsional"
                fullWidth
              />

              <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                <Input
                  label="Telepon"
                  type="tel"
                  placeholder="08123456789"
                  value={formData.phone}
                  onChange={handleChange('phone')}
                  helperText="Opsional"
                  fullWidth
                />
                <Input
                  label="Email"
                  type="email"
                  placeholder="email@contoh.com"
                  value={formData.email}
                  onChange={handleChange('email')}
                  error={errors.email}
                  helperText="Opsional"
                  fullWidth
                />
              </div>

              <Input
                label="Alamat"
                placeholder="Alamat lengkap supplier"
                value={formData.address}
                onChange={handleChange('address')}
                helperText="Opsional"
                fullWidth
              />

              <div>
                <label className="block text-sm font-medium text-text-main dark:text-white mb-1">
                  Catatan
                </label>
                <textarea
                  className="w-full rounded-lg border border-border-light dark:border-border-dark bg-white dark:bg-background-dark px-3 py-2 text-sm text-text-main dark:text-white placeholder:text-text-sub focus:outline-none focus:ring-2 focus:ring-primary"
                  rows={3}
                  placeholder="Catatan tambahan (opsional)"
                  value={formData.notes}
                  onChange={handleChange('notes')}
                />
              </div>
            </div>

            <div className="flex items-center justify-end gap-3 p-6 border-t border-border-light dark:border-border-dark">
              <Button type="button" variant="outline" onClick={handleClose}>
                Batal
              </Button>
              <Button type="submit" disabled={isLoading}>
                {isLoading ? 'Menyimpan...' : initialData ? 'Simpan Perubahan' : 'Simpan Supplier'}
              </Button>
            </div>
          </form>
        </div>
      </div>
    </div>
  );
}

export function SuppliersPage() {
  const [page, setPage] = useState(1);
  const [search, setSearch] = useState('');
  const [activeFilter, setActiveFilter] = useState<string>('');
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [editingSupplier, setEditingSupplier] = useState<Supplier | null>(null);

  const { data, isLoading, refetch } = useSuppliers({
    page,
    per_page: 10,
    search: search || undefined,
    is_active: activeFilter === '' ? undefined : activeFilter === 'true',
  });

  const createSupplier = useCreateSupplier();
  const updateSupplier = useUpdateSupplier();
  const deleteSupplier = useDeleteSupplier();

  const handleAddSupplier = async (formData: CreateSupplierRequest) => {
    try {
      await createSupplier.mutateAsync(formData);
      toast.success('Supplier berhasil ditambahkan');
      setIsModalOpen(false);
    } catch (err) {
      toast.error(getApiErrorMessage(err));
    }
  };

  const handleEditSupplier = async (formData: CreateSupplierRequest) => {
    if (!editingSupplier) return;
    try {
      await updateSupplier.mutateAsync({ id: editingSupplier.id, supplier: formData });
      toast.success('Supplier berhasil diperbarui');
      setEditingSupplier(null);
    } catch (err) {
      toast.error(getApiErrorMessage(err));
    }
  };

  const handleToggleActive = async (supplier: Supplier) => {
    try {
      await updateSupplier.mutateAsync({
        id: supplier.id,
        supplier: { is_active: !supplier.is_active },
      });
      toast.success(
        supplier.is_active ? 'Supplier dinonaktifkan' : 'Supplier diaktifkan'
      );
    } catch (err) {
      toast.error(getApiErrorMessage(err));
    }
  };

  const handleDeleteSupplier = async (id: number) => {
    if (window.confirm('Apakah Anda yakin ingin menghapus supplier ini?')) {
      try {
        await deleteSupplier.mutateAsync(id);
        toast.success('Supplier berhasil dihapus');
      } catch (err) {
        toast.error(getApiErrorMessage(err));
      }
    }
  };

  const supplierColumns = [
    {
      header: 'Supplier',
      accessor: (row: Supplier) => (
        <div>
          <p className="font-medium text-text-main dark:text-white">{row.name}</p>
          <p className="text-xs text-text-sub dark:text-gray-400 font-mono">{row.code}</p>
        </div>
      ),
    },
    {
      header: 'Kontak',
      accessor: (row: Supplier) => (
        <div>
          <p className="text-sm text-text-main dark:text-gray-200">{row.contact_name || '-'}</p>
          <p className="text-xs text-text-sub dark:text-gray-400">{row.phone || '-'}</p>
        </div>
      ),
    },
    {
      header: 'Email',
      accessor: (row: Supplier) => (
        <span className="text-sm text-text-main dark:text-gray-200">{row.email || '-'}</span>
      ),
    },
    {
      header: 'Status',
      accessor: (row: Supplier) => (
        <Badge variant={row.is_active ? 'success' : 'gray'} dot>
          {row.is_active ? 'Aktif' : 'Nonaktif'}
        </Badge>
      ),
    },
    {
      header: 'Terdaftar',
      accessor: (row: Supplier) => (
        <span className="text-sm text-text-sub dark:text-gray-400">
          {formatDate(new Date(row.created_at), 'dd MMM yyyy')}
        </span>
      ),
    },
    {
      header: 'Aksi',
      accessor: (row: Supplier) => (
        <div className="flex items-center gap-2">
          <button
            className="p-1.5 hover:bg-background-light dark:hover:bg-background-dark rounded-lg transition-colors"
            onClick={e => {
              e.stopPropagation();
              setEditingSupplier(row);
            }}
            title="Edit"
          >
            <Icon name="edit" size={18} className="text-text-sub dark:text-gray-400" />
          </button>
          <button
            className="p-1.5 hover:bg-background-light dark:hover:bg-background-dark rounded-lg transition-colors"
            onClick={e => {
              e.stopPropagation();
              handleToggleActive(row);
            }}
            title={row.is_active ? 'Nonaktifkan' : 'Aktifkan'}
          >
            <Icon
              name={row.is_active ? 'toggle_on' : 'toggle_off'}
              size={18}
              className={row.is_active ? 'text-green-500' : 'text-gray-400'}
            />
          </button>
          <button
            className="p-1.5 hover:bg-red-50 dark:hover:bg-red-900/20 rounded-lg transition-colors"
            onClick={e => {
              e.stopPropagation();
              handleDeleteSupplier(row.id);
            }}
            title="Hapus"
          >
            <Icon name="delete" size={18} className="text-red-500" />
          </button>
        </div>
      ),
      className: 'text-right',
    },
  ];

  const handleSearch = (value: string) => {
    setSearch(value);
    setPage(1);
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h1 className="text-2xl md:text-3xl font-bold text-text-main dark:text-white">Supplier</h1>
          <p className="text-text-sub dark:text-gray-400 text-sm mt-1">
            Kelola data supplier dan pemasok
          </p>
        </div>
        <Button
          leftIcon={<Icon name="add_business" size={20} />}
          onClick={() => setIsModalOpen(true)}
        >
          Tambah Supplier
        </Button>
      </div>

      {/* Stats Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
        <Card variant="bordered" padding="md">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-xs text-text-sub dark:text-gray-400 mb-1">Total Supplier</p>
              <p className="text-2xl font-bold text-text-main dark:text-white">
                {data?.total || 0}
              </p>
            </div>
            <div className="w-12 h-12 rounded-full bg-blue-50 dark:bg-blue-900/20 flex items-center justify-center">
              <Icon name="business" size={24} className="text-primary" />
            </div>
          </div>
        </Card>

        <Card variant="bordered" padding="md">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-xs text-text-sub dark:text-gray-400 mb-1">Supplier Aktif</p>
              <p className="text-2xl font-bold text-text-main dark:text-white">
                {data?.data.filter(s => s.is_active).length || 0}
              </p>
            </div>
            <div className="w-12 h-12 rounded-full bg-green-50 dark:bg-green-900/20 flex items-center justify-center">
              <Icon name="check_circle" size={24} className="text-green-600" />
            </div>
          </div>
        </Card>

        <Card variant="bordered" padding="md">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-xs text-text-sub dark:text-gray-400 mb-1">Supplier Nonaktif</p>
              <p className="text-2xl font-bold text-text-main dark:text-white">
                {data?.data.filter(s => !s.is_active).length || 0}
              </p>
            </div>
            <div className="w-12 h-12 rounded-full bg-gray-50 dark:bg-gray-800 flex items-center justify-center">
              <Icon name="cancel" size={24} className="text-gray-600" />
            </div>
          </div>
        </Card>
      </div>

      {/* Filters */}
      <Card variant="bordered" padding="md">
        <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
          <Input
            placeholder="Cari nama atau kode supplier..."
            leftIcon={<Icon name="search" size={20} />}
            value={search}
            onChange={e => handleSearch(e.target.value)}
          />
          <Dropdown
            value={activeFilter}
            onChange={setActiveFilter}
            options={[
              { value: '', label: 'Semua Status' },
              { value: 'true', label: 'Aktif' },
              { value: 'false', label: 'Nonaktif' },
            ]}
            leftIcon={<Icon name="filter_list" size={20} />}
          />
          <Button
            variant="outline"
            leftIcon={<Icon name="refresh" size={20} />}
            onClick={() => {
              setSearch('');
              setActiveFilter('');
              setPage(1);
              refetch();
            }}
          >
            Reset Filter
          </Button>
        </div>
      </Card>

      {/* Suppliers Table */}
      <Card variant="bordered" padding="none">
        <DataTable
          data={data?.data || []}
          columns={supplierColumns}
          loading={isLoading}
          emptyMessage="Belum ada supplier"
        />

        {/* Pagination */}
        {data && data.totalPages > 1 && (
          <div className="px-6 py-4 border-t border-border-light dark:border-border-dark">
            <div className="flex items-center justify-between">
              <p className="text-sm text-text-sub dark:text-gray-400">
                Menampilkan {(page - 1) * 10 + 1} - {Math.min(page * 10, data.total)} dari{' '}
                {data.total} supplier
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

      {/* Add Supplier Modal */}
      <SupplierFormModal
        isOpen={isModalOpen}
        onClose={() => setIsModalOpen(false)}
        onSubmit={handleAddSupplier}
        isLoading={createSupplier.isPending}
      />

      {/* Edit Supplier Modal */}
      <SupplierFormModal
        key={editingSupplier?.id}
        isOpen={!!editingSupplier}
        onClose={() => setEditingSupplier(null)}
        onSubmit={handleEditSupplier}
        isLoading={updateSupplier.isPending}
        initialData={editingSupplier ?? undefined}
      />
    </div>
  );
}

import { useState } from 'react';
import { Card, Badge, Button, Input, Dropdown, Icon, Avatar } from '@/components/ui';
import { DataTable } from '@/components/common';
import { useCustomers, useCreateCustomer, useDeleteCustomer } from '@/hooks/useCustomers';
import { formatDate } from '@/utils';
import { isValidEmail, isValidPhoneNumber } from '@/utils/validators';
import type { Customer, CreateCustomerRequest } from '@/types/customer';

interface CustomerFormData {
  code: string;
  name: string;
  email: string;
  phone: string;
  is_member: boolean;
}

interface CustomerFormErrors {
  code?: string;
  name?: string;
  email?: string;
  phone?: string;
}

function CustomerFormModal({
  isOpen,
  onClose,
  onSubmit,
  isLoading,
}: {
  isOpen: boolean;
  onClose: () => void;
  onSubmit: (data: CreateCustomerRequest) => void;
  isLoading: boolean;
}) {
  const [formData, setFormData] = useState<CustomerFormData>({
    code: '',
    name: '',
    email: '',
    phone: '',
    is_member: false,
  });
  const [errors, setErrors] = useState<CustomerFormErrors>({});

  const validateForm = (): boolean => {
    const newErrors: CustomerFormErrors = {};

    if (!formData.code.trim()) {
      newErrors.code = 'Kode pelanggan wajib diisi';
    }

    if (!formData.name.trim()) {
      newErrors.name = 'Nama pelanggan wajib diisi';
    } else if (formData.name.trim().length < 2) {
      newErrors.name = 'Nama minimal 2 karakter';
    }

    if (formData.email && !isValidEmail(formData.email)) {
      newErrors.email = 'Format email tidak valid';
    }

    if (formData.phone && !isValidPhoneNumber(formData.phone)) {
      newErrors.phone = 'Format nomor telepon tidak valid';
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleChange = (field: keyof CustomerFormData) => (
    e: React.ChangeEvent<HTMLInputElement>
  ) => {
    const value = field === 'is_member' ? (e.target as HTMLInputElement).checked : e.target.value;
    setFormData(prev => ({ ...prev, [field]: value }));
    if (errors[field as keyof CustomerFormErrors]) {
      setErrors(prev => ({ ...prev, [field]: undefined }));
    }
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (validateForm()) {
      onSubmit({
        code: formData.code,
        name: formData.name,
        email: formData.email || undefined,
        phone: formData.phone || undefined,
        is_member: formData.is_member,
      });
    }
  };

  const handleClose = () => {
    setFormData({ code: '', name: '', email: '', phone: '', is_member: false });
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
                Tambah Pelanggan Baru
              </h2>
              <p className="text-sm text-text-sub dark:text-gray-400 mt-1">
                Isi informasi pelanggan di bawah ini
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
              <Input
                label="Kode Pelanggan"
                placeholder="Contoh: CUST001"
                value={formData.code}
                onChange={handleChange('code')}
                error={errors.code}
                fullWidth
                required
              />

              <Input
                label="Nama Pelanggan"
                placeholder="Masukkan nama lengkap"
                value={formData.name}
                onChange={handleChange('name')}
                error={errors.name}
                fullWidth
                required
              />

              <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
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
                <Input
                  label="Nomor Telepon"
                  type="tel"
                  placeholder="08123456789"
                  value={formData.phone}
                  onChange={handleChange('phone')}
                  error={errors.phone}
                  helperText="Opsional"
                  fullWidth
                />
              </div>

              <div className="flex items-center gap-3">
                <input
                  type="checkbox"
                  id="is_member"
                  checked={formData.is_member}
                  onChange={handleChange('is_member')}
                  className="w-4 h-4 text-primary border-gray-300 rounded focus:ring-primary"
                />
                <label htmlFor="is_member" className="text-sm font-medium text-text-main dark:text-white">
                  Member
                </label>
              </div>
            </div>

            <div className="flex items-center justify-end gap-3 p-6 border-t border-border-light dark:border-border-dark">
              <Button type="button" variant="outline" onClick={handleClose}>
                Batal
              </Button>
              <Button type="submit" disabled={isLoading}>
                {isLoading ? 'Menyimpan...' : 'Simpan Pelanggan'}
              </Button>
            </div>
          </form>
        </div>
      </div>
    </div>
  );
}

export function CustomersPage() {
  const [page, setPage] = useState(1);
  const [search, setSearch] = useState('');
  const [isMemberFilter, setIsMemberFilter] = useState<string>('');
  const [isModalOpen, setIsModalOpen] = useState(false);

  const { data, isLoading, refetch } = useCustomers({
    page,
    per_page: 10,
    search: search || undefined,
    is_member: isMemberFilter === '' ? undefined : isMemberFilter === 'true',
  });

  const createCustomer = useCreateCustomer();
  const deleteCustomer = useDeleteCustomer();

  const handleAddCustomer = async (formData: CreateCustomerRequest) => {
    try {
      await createCustomer.mutateAsync(formData);
      setIsModalOpen(false);
    } catch (error) {
      console.error('Failed to create customer:', error);
    }
  };

  const handleDeleteCustomer = async (id: number) => {
    if (window.confirm('Apakah Anda yakin ingin menghapus pelanggan ini?')) {
      try {
        await deleteCustomer.mutateAsync(id);
      } catch (error) {
        console.error('Failed to delete customer:', error);
      }
    }
  };

  const customerColumns = [
    {
      header: 'Pelanggan',
      accessor: (row: Customer) => (
        <div className="flex items-center gap-3">
          <Avatar name={row.name} size="md" />
          <div>
            <p className="font-medium text-text-main dark:text-white">{row.name}</p>
            <p className="text-xs text-text-sub dark:text-gray-400">{row.email || '-'}</p>
          </div>
        </div>
      ),
    },
    {
      header: 'Kode',
      accessor: (row: Customer) => (
        <span className="text-text-main dark:text-gray-200 font-mono text-sm">{row.code}</span>
      ),
    },
    {
      header: 'Telepon',
      accessor: (row: Customer) => (
        <span className="text-text-main dark:text-gray-200">{row.phone || '-'}</span>
      ),
    },
    {
      header: 'Status',
      accessor: (row: Customer) => (
        <Badge variant={row.is_member ? 'primary' : 'gray'} dot>
          {row.is_member ? 'Member' : 'Non-Member'}
        </Badge>
      ),
    },
    {
      header: 'Tanggal Daftar',
      accessor: (row: Customer) => (
        <span className="text-sm text-text-sub dark:text-gray-400">
          {formatDate(new Date(row.created_at), 'dd MMM yyyy')}
        </span>
      ),
    },
    {
      header: 'Aksi',
      accessor: (row: Customer) => (
        <div className="flex items-center gap-2">
          <button
            className="p-1.5 hover:bg-background-light dark:hover:bg-background-dark rounded-lg transition-colors"
            onClick={e => {
              e.stopPropagation();
              console.log('Edit customer:', row.id);
            }}
            title="Edit"
          >
            <Icon name="edit" size={18} className="text-text-sub dark:text-gray-400" />
          </button>
          <button
            className="p-1.5 hover:bg-red-50 dark:hover:bg-red-900/20 rounded-lg transition-colors"
            onClick={e => {
              e.stopPropagation();
              handleDeleteCustomer(row.id);
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

  const memberCount = data?.data.filter(c => c.is_member).length || 0;

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h1 className="text-2xl md:text-3xl font-bold text-text-main dark:text-white">
            Pelanggan
          </h1>
          <p className="text-text-sub dark:text-gray-400 text-sm mt-1">
            Kelola database pelanggan
          </p>
        </div>
        <Button
          leftIcon={<Icon name="person_add" size={20} />}
          onClick={() => setIsModalOpen(true)}
        >
          Tambah Pelanggan
        </Button>
      </div>

      {/* Stats Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
        <Card variant="bordered" padding="md">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-xs text-text-sub dark:text-gray-400 mb-1">Total Pelanggan</p>
              <p className="text-2xl font-bold text-text-main dark:text-white">
                {data?.total || 0}
              </p>
            </div>
            <div className="w-12 h-12 rounded-full bg-blue-50 dark:bg-blue-900/20 flex items-center justify-center">
              <Icon name="groups" size={24} className="text-primary" />
            </div>
          </div>
        </Card>

        <Card variant="bordered" padding="md">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-xs text-text-sub dark:text-gray-400 mb-1">Member</p>
              <p className="text-2xl font-bold text-text-main dark:text-white">{memberCount}</p>
            </div>
            <div className="w-12 h-12 rounded-full bg-green-50 dark:bg-green-900/20 flex items-center justify-center">
              <Icon name="card_membership" size={24} className="text-green-600" />
            </div>
          </div>
        </Card>

        <Card variant="bordered" padding="md">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-xs text-text-sub dark:text-gray-400 mb-1">Non-Member</p>
              <p className="text-2xl font-bold text-text-main dark:text-white">
                {(data?.total || 0) - memberCount}
              </p>
            </div>
            <div className="w-12 h-12 rounded-full bg-gray-50 dark:bg-gray-800 flex items-center justify-center">
              <Icon name="person" size={24} className="text-gray-600" />
            </div>
          </div>
        </Card>
      </div>

      {/* Filters */}
      <Card variant="bordered" padding="md">
        <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
          <Input
            placeholder="Cari nama, kode, atau email..."
            leftIcon={<Icon name="search" size={20} />}
            value={search}
            onChange={e => handleSearch(e.target.value)}
          />
          <Dropdown
            value={isMemberFilter}
            onChange={setIsMemberFilter}
            options={[
              { value: '', label: 'Semua Status' },
              { value: 'true', label: 'Member' },
              { value: 'false', label: 'Non-Member' },
            ]}
            leftIcon={<Icon name="filter_list" size={20} />}
          />
          <Button
            variant="outline"
            leftIcon={<Icon name="refresh" size={20} />}
            onClick={() => {
              setSearch('');
              setIsMemberFilter('');
              setPage(1);
              refetch();
            }}
          >
            Reset Filter
          </Button>
        </div>
      </Card>

      {/* Customers Table */}
      <Card variant="bordered" padding="none">
        <DataTable
          data={data?.data || []}
          columns={customerColumns}
          loading={isLoading}
          emptyMessage="Belum ada pelanggan"
        />

        {/* Pagination */}
        {data && data.totalPages > 1 && (
          <div className="px-6 py-4 border-t border-border-light dark:border-border-dark">
            <div className="flex items-center justify-between">
              <p className="text-sm text-text-sub dark:text-gray-400">
                Menampilkan {(page - 1) * 10 + 1} - {Math.min(page * 10, data.total)} dari{' '}
                {data.total} pelanggan
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

      {/* Add Customer Modal */}
      <CustomerFormModal
        isOpen={isModalOpen}
        onClose={() => setIsModalOpen(false)}
        onSubmit={handleAddCustomer}
        isLoading={createCustomer.isPending}
      />
    </div>
  );
}

import { useState, useEffect } from 'react';
import { Card, Badge, Button, Input, Dropdown, Icon } from '@/components/ui';
import { DataTable } from '@/components/common';
import {
  useProductCategories,
  useAllProductCategories,
  useCreateProductCategory,
  useUpdateProductCategory,
  useDeleteProductCategory,
} from '@/hooks/useProductCategories';
import { formatDate } from '@/utils';
import type { ProductCategory, CreateProductCategoryRequest, UpdateProductCategoryRequest } from '@/types/product';

interface CategoryFormData {
  code: string;
  name: string;
  parent_id: string;
  sort_order: string;
  is_active: boolean;
}

interface CategoryFormErrors {
  code?: string;
  name?: string;
}

function CategoryFormModal({
  isOpen,
  onClose,
  onSubmit,
  isLoading,
  editingCategory,
  categories,
}: {
  isOpen: boolean;
  onClose: () => void;
  onSubmit: (data: CreateProductCategoryRequest | UpdateProductCategoryRequest) => void;
  isLoading: boolean;
  editingCategory: ProductCategory | null;
  categories: ProductCategory[];
}) {
  const [formData, setFormData] = useState<CategoryFormData>({
    code: '',
    name: '',
    parent_id: '',
    sort_order: '0',
    is_active: true,
  });
  const [errors, setErrors] = useState<CategoryFormErrors>({});

  useEffect(() => {
    if (editingCategory) {
      setFormData({
        code: editingCategory.code,
        name: editingCategory.name,
        parent_id: editingCategory.parent_id?.toString() || '',
        sort_order: editingCategory.sort_order.toString(),
        is_active: editingCategory.is_active,
      });
    } else {
      setFormData({
        code: '',
        name: '',
        parent_id: '',
        sort_order: '0',
        is_active: true,
      });
    }
    setErrors({});
  }, [editingCategory, isOpen]);

  const validateForm = (): boolean => {
    const newErrors: CategoryFormErrors = {};

    if (!formData.code.trim()) {
      newErrors.code = 'Kode kategori wajib diisi';
    }

    if (!formData.name.trim()) {
      newErrors.name = 'Nama kategori wajib diisi';
    } else if (formData.name.trim().length < 2) {
      newErrors.name = 'Nama minimal 2 karakter';
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleChange = (field: keyof CategoryFormData) => (
    e: React.ChangeEvent<HTMLInputElement>
  ) => {
    const value = field === 'is_active' ? (e.target as HTMLInputElement).checked : e.target.value;
    setFormData(prev => ({ ...prev, [field]: value }));
    if (errors[field as keyof CategoryFormErrors]) {
      setErrors(prev => ({ ...prev, [field]: undefined }));
    }
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (validateForm()) {
      onSubmit({
        code: formData.code,
        name: formData.name,
        parent_id: formData.parent_id ? parseInt(formData.parent_id) : null,
        sort_order: parseInt(formData.sort_order) || 0,
        is_active: formData.is_active,
      });
    }
  };

  const handleClose = () => {
    setFormData({ code: '', name: '', parent_id: '', sort_order: '0', is_active: true });
    setErrors({});
    onClose();
  };

  // Filter out current category from parent options to prevent circular reference
  const parentOptions = categories.filter(c => c.id !== editingCategory?.id);

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 z-50 overflow-y-auto">
      <div className="fixed inset-0 bg-black/50 transition-opacity" onClick={handleClose} />
      <div className="flex min-h-full items-center justify-center p-4">
        <div className="relative w-full max-w-lg bg-white dark:bg-card-dark rounded-2xl shadow-xl">
          <div className="flex items-center justify-between p-6 border-b border-border-light dark:border-border-dark">
            <div>
              <h2 className="text-xl font-bold text-text-main dark:text-white">
                {editingCategory ? 'Edit Kategori' : 'Tambah Kategori Baru'}
              </h2>
              <p className="text-sm text-text-sub dark:text-gray-400 mt-1">
                Isi informasi kategori di bawah ini
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
                label="Kode Kategori"
                placeholder="Contoh: FOOD"
                value={formData.code}
                onChange={handleChange('code')}
                error={errors.code}
                fullWidth
                required
              />

              <Input
                label="Nama Kategori"
                placeholder="Masukkan nama kategori"
                value={formData.name}
                onChange={handleChange('name')}
                error={errors.name}
                fullWidth
                required
              />

              <div className="space-y-2">
                <label className="block text-sm font-medium text-text-main dark:text-white">
                  Kategori Induk
                </label>
                <select
                  value={formData.parent_id}
                  onChange={e => setFormData(prev => ({ ...prev, parent_id: e.target.value }))}
                  className="w-full px-4 py-2.5 rounded-lg border border-border-light dark:border-border-dark bg-white dark:bg-card-dark text-text-main dark:text-white focus:outline-none focus:ring-2 focus:ring-primary/50"
                >
                  <option value="">Tidak ada (Kategori Utama)</option>
                  {parentOptions.map(cat => (
                    <option key={cat.id} value={cat.id}>
                      {cat.name}
                    </option>
                  ))}
                </select>
                <p className="text-xs text-text-sub dark:text-gray-400">Opsional</p>
              </div>

              <Input
                label="Urutan"
                type="number"
                placeholder="0"
                value={formData.sort_order}
                onChange={handleChange('sort_order')}
                helperText="Semakin kecil, semakin di atas"
                fullWidth
              />

              <div className="flex items-center gap-3">
                <input
                  type="checkbox"
                  id="is_active"
                  checked={formData.is_active}
                  onChange={handleChange('is_active')}
                  className="w-4 h-4 text-primary border-gray-300 rounded focus:ring-primary"
                />
                <label htmlFor="is_active" className="text-sm font-medium text-text-main dark:text-white">
                  Aktif
                </label>
              </div>
            </div>

            <div className="flex items-center justify-end gap-3 p-6 border-t border-border-light dark:border-border-dark">
              <Button type="button" variant="outline" onClick={handleClose}>
                Batal
              </Button>
              <Button type="submit" disabled={isLoading}>
                {isLoading ? 'Menyimpan...' : editingCategory ? 'Simpan Perubahan' : 'Simpan Kategori'}
              </Button>
            </div>
          </form>
        </div>
      </div>
    </div>
  );
}

export function ProductCategoriesPage() {
  const [page, setPage] = useState(1);
  const [search, setSearch] = useState('');
  const [statusFilter, setStatusFilter] = useState<string>('');
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [editingCategory, setEditingCategory] = useState<ProductCategory | null>(null);

  const { data, isLoading, refetch } = useProductCategories({
    page,
    per_page: 10,
    search: search || undefined,
    is_active: statusFilter === '' ? undefined : statusFilter === 'true',
  });

  const { data: allCategories } = useAllProductCategories();

  const createCategory = useCreateProductCategory();
  const updateCategory = useUpdateProductCategory();
  const deleteCategory = useDeleteProductCategory();

  const handleAddCategory = async (formData: CreateProductCategoryRequest) => {
    try {
      await createCategory.mutateAsync(formData);
      setIsModalOpen(false);
    } catch (error) {
      console.error('Failed to create category:', error);
    }
  };

  const handleUpdateCategory = async (formData: UpdateProductCategoryRequest) => {
    if (!editingCategory) return;
    try {
      await updateCategory.mutateAsync({ id: editingCategory.id, category: formData });
      setIsModalOpen(false);
      setEditingCategory(null);
    } catch (error) {
      console.error('Failed to update category:', error);
    }
  };

  const handleDeleteCategory = async (id: number) => {
    if (window.confirm('Apakah Anda yakin ingin menghapus kategori ini?')) {
      try {
        await deleteCategory.mutateAsync(id);
      } catch (error) {
        console.error('Failed to delete category:', error);
      }
    }
  };

  const handleEdit = (category: ProductCategory) => {
    setEditingCategory(category);
    setIsModalOpen(true);
  };

  const handleCloseModal = () => {
    setIsModalOpen(false);
    setEditingCategory(null);
  };

  const categoryColumns = [
    {
      header: 'Kode',
      accessor: (row: ProductCategory) => (
        <span className="text-text-main dark:text-gray-200 font-mono text-sm">{row.code}</span>
      ),
    },
    {
      header: 'Nama Kategori',
      accessor: (row: ProductCategory) => (
        <div>
          <p className="font-medium text-text-main dark:text-white">{row.name}</p>
          {row.parent_name && (
            <p className="text-xs text-text-sub dark:text-gray-400">
              Induk: {row.parent_name}
            </p>
          )}
        </div>
      ),
    },
    {
      header: 'Urutan',
      accessor: (row: ProductCategory) => (
        <span className="text-text-main dark:text-gray-200">{row.sort_order}</span>
      ),
    },
    {
      header: 'Status',
      accessor: (row: ProductCategory) => (
        <Badge variant={row.is_active ? 'success' : 'gray'} dot>
          {row.is_active ? 'Aktif' : 'Nonaktif'}
        </Badge>
      ),
    },
    {
      header: 'Tanggal Dibuat',
      accessor: (row: ProductCategory) => (
        <span className="text-sm text-text-sub dark:text-gray-400">
          {formatDate(new Date(row.created_at), 'dd MMM yyyy')}
        </span>
      ),
    },
    {
      header: 'Aksi',
      accessor: (row: ProductCategory) => (
        <div className="flex items-center gap-2">
          <button
            className="p-1.5 hover:bg-background-light dark:hover:bg-background-dark rounded-lg transition-colors"
            onClick={e => {
              e.stopPropagation();
              handleEdit(row);
            }}
            title="Edit"
          >
            <Icon name="edit" size={18} className="text-text-sub dark:text-gray-400" />
          </button>
          <button
            className="p-1.5 hover:bg-red-50 dark:hover:bg-red-900/20 rounded-lg transition-colors"
            onClick={e => {
              e.stopPropagation();
              handleDeleteCategory(row.id);
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

  const activeCount = data?.data.filter(c => c.is_active).length || 0;

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h1 className="text-2xl md:text-3xl font-bold text-text-main dark:text-white">
            Kategori Produk
          </h1>
          <p className="text-text-sub dark:text-gray-400 text-sm mt-1">
            Kelola kategori produk
          </p>
        </div>
        <Button
          leftIcon={<Icon name="add" size={20} />}
          onClick={() => {
            setEditingCategory(null);
            setIsModalOpen(true);
          }}
        >
          Tambah Kategori
        </Button>
      </div>

      {/* Stats Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
        <Card variant="bordered" padding="md">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-xs text-text-sub dark:text-gray-400 mb-1">Total Kategori</p>
              <p className="text-2xl font-bold text-text-main dark:text-white">
                {data?.total || 0}
              </p>
            </div>
            <div className="w-12 h-12 rounded-full bg-blue-50 dark:bg-blue-900/20 flex items-center justify-center">
              <Icon name="category" size={24} className="text-primary" />
            </div>
          </div>
        </Card>

        <Card variant="bordered" padding="md">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-xs text-text-sub dark:text-gray-400 mb-1">Aktif</p>
              <p className="text-2xl font-bold text-text-main dark:text-white">{activeCount}</p>
            </div>
            <div className="w-12 h-12 rounded-full bg-green-50 dark:bg-green-900/20 flex items-center justify-center">
              <Icon name="check_circle" size={24} className="text-green-600" />
            </div>
          </div>
        </Card>

        <Card variant="bordered" padding="md">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-xs text-text-sub dark:text-gray-400 mb-1">Nonaktif</p>
              <p className="text-2xl font-bold text-text-main dark:text-white">
                {(data?.total || 0) - activeCount}
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
            placeholder="Cari kode atau nama..."
            leftIcon={<Icon name="search" size={20} />}
            value={search}
            onChange={e => handleSearch(e.target.value)}
          />
          <Dropdown
            value={statusFilter}
            onChange={setStatusFilter}
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
              setStatusFilter('');
              setPage(1);
              refetch();
            }}
          >
            Reset Filter
          </Button>
        </div>
      </Card>

      {/* Categories Table */}
      <Card variant="bordered" padding="none">
        <DataTable
          data={data?.data || []}
          columns={categoryColumns}
          loading={isLoading}
          emptyMessage="Belum ada kategori"
        />

        {/* Pagination */}
        {data && data.totalPages > 1 && (
          <div className="px-6 py-4 border-t border-border-light dark:border-border-dark">
            <div className="flex items-center justify-between">
              <p className="text-sm text-text-sub dark:text-gray-400">
                Menampilkan {(page - 1) * 10 + 1} - {Math.min(page * 10, data.total)} dari{' '}
                {data.total} kategori
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

      {/* Add/Edit Category Modal */}
      <CategoryFormModal
        isOpen={isModalOpen}
        onClose={handleCloseModal}
        onSubmit={editingCategory ? handleUpdateCategory : handleAddCategory}
        isLoading={editingCategory ? updateCategory.isPending : createCategory.isPending}
        editingCategory={editingCategory}
        categories={allCategories || []}
      />
    </div>
  );
}

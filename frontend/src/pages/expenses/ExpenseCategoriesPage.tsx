import { useState } from 'react';
import toast from 'react-hot-toast';
import { Card, Button, Input, Icon } from '@/components/ui';
import { DataTable } from '@/components/common';
import {
  useExpenseCategories,
  useCreateExpenseCategory,
  useUpdateExpenseCategory,
  useDeleteExpenseCategory,
} from '@/hooks/useExpenses';
import { formatDate, getApiErrorMessage } from '@/utils';
import type { ExpenseCategory } from '@/types/expense';

interface CategoryFormData {
  name: string;
  description: string;
}

interface CategoryFormErrors {
  name?: string;
}

function CategoryFormModal({
  isOpen,
  onClose,
  onSubmit,
  isLoading,
  initialData,
}: {
  isOpen: boolean;
  onClose: () => void;
  onSubmit: (data: { name: string; description?: string }) => void;
  isLoading: boolean;
  initialData?: ExpenseCategory;
}) {
  const [formData, setFormData] = useState<CategoryFormData>({
    name: initialData?.name ?? '',
    description: initialData?.description ?? '',
  });
  const [errors, setErrors] = useState<CategoryFormErrors>({});

  const validateForm = (): boolean => {
    const newErrors: CategoryFormErrors = {};
    if (!formData.name.trim()) {
      newErrors.name = 'Nama kategori wajib diisi';
    }
    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleChange =
    (field: keyof CategoryFormData) =>
    (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => {
      setFormData(prev => ({ ...prev, [field]: e.target.value }));
      if (errors[field as keyof CategoryFormErrors]) {
        setErrors(prev => ({ ...prev, [field]: undefined }));
      }
    };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (validateForm()) {
      onSubmit({
        name: formData.name,
        description: formData.description || undefined,
      });
    }
  };

  const handleClose = () => {
    setFormData({ name: '', description: '' });
    setErrors({});
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
              <h2 className="text-xl font-bold text-text-main dark:text-white">
                {initialData ? 'Edit Kategori' : 'Tambah Kategori'}
              </h2>
              <p className="text-sm text-text-sub dark:text-gray-400 mt-1">
                {initialData ? `Edit ${initialData.name}` : 'Kategori pengeluaran baru'}
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
              <Input
                label="Nama Kategori"
                placeholder="Contoh: Operasional, Pemeliharaan"
                value={formData.name}
                onChange={handleChange('name')}
                error={errors.name}
                fullWidth
                required
              />

              <div>
                <label className="block text-sm font-medium text-text-main dark:text-white mb-1">
                  Deskripsi
                </label>
                <textarea
                  className="w-full rounded-lg border border-border-light dark:border-border-dark bg-white dark:bg-background-dark px-3 py-2 text-sm text-text-main dark:text-white placeholder:text-text-sub focus:outline-none focus:ring-2 focus:ring-primary"
                  rows={3}
                  placeholder="Deskripsi kategori (opsional)"
                  value={formData.description}
                  onChange={handleChange('description')}
                />
              </div>
            </div>

            <div className="flex items-center justify-end gap-3 p-6 border-t border-border-light dark:border-border-dark">
              <Button type="button" variant="outline" onClick={handleClose}>
                Batal
              </Button>
              <Button type="submit" disabled={isLoading}>
                {isLoading ? 'Menyimpan...' : initialData ? 'Simpan Perubahan' : 'Simpan Kategori'}
              </Button>
            </div>
          </form>
        </div>
      </div>
    </div>
  );
}

export function ExpenseCategoriesPage() {
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [editingCategory, setEditingCategory] = useState<ExpenseCategory | null>(null);

  const { data: categories, isLoading } = useExpenseCategories();
  const createCategory = useCreateExpenseCategory();
  const updateCategory = useUpdateExpenseCategory();
  const deleteCategory = useDeleteExpenseCategory();

  const categoriesList = categories || [];

  const handleAddCategory = async (data: { name: string; description?: string }) => {
    try {
      await createCategory.mutateAsync(data);
      toast.success('Kategori berhasil ditambahkan');
      setIsModalOpen(false);
    } catch (err) {
      toast.error(getApiErrorMessage(err));
    }
  };

  const handleEditCategory = async (data: { name: string; description?: string }) => {
    if (!editingCategory) return;
    try {
      await updateCategory.mutateAsync({ id: editingCategory.id, category: data });
      toast.success('Kategori berhasil diperbarui');
      setEditingCategory(null);
    } catch (err) {
      toast.error(getApiErrorMessage(err));
    }
  };

  const handleDeleteCategory = async (id: number) => {
    if (window.confirm('Apakah Anda yakin ingin menghapus kategori ini? Pengeluaran yang terkait akan kehilangan kategorinya.')) {
      try {
        await deleteCategory.mutateAsync(id);
        toast.success('Kategori berhasil dihapus');
      } catch (err) {
        toast.error(getApiErrorMessage(err));
      }
    }
  };

  const categoryColumns = [
    {
      header: 'Nama Kategori',
      accessor: (row: ExpenseCategory) => (
        <p className="font-medium text-text-main dark:text-white">{row.name}</p>
      ),
    },
    {
      header: 'Deskripsi',
      accessor: (row: ExpenseCategory) => (
        <span className="text-sm text-text-sub dark:text-gray-400">{row.description || '-'}</span>
      ),
    },
    {
      header: 'Dibuat',
      accessor: (row: ExpenseCategory) => (
        <span className="text-sm text-text-sub dark:text-gray-400">
          {formatDate(new Date(row.created_at), 'dd MMM yyyy')}
        </span>
      ),
    },
    {
      header: 'Aksi',
      accessor: (row: ExpenseCategory) => (
        <div className="flex items-center gap-2">
          <button
            className="p-1.5 hover:bg-background-light dark:hover:bg-background-dark rounded-lg transition-colors"
            onClick={e => {
              e.stopPropagation();
              setEditingCategory(row);
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

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h1 className="text-2xl md:text-3xl font-bold text-text-main dark:text-white">
            Kategori Pengeluaran
          </h1>
          <p className="text-text-sub dark:text-gray-400 text-sm mt-1">
            Kelola kategori untuk pengeluaran operasional
          </p>
        </div>
        <Button
          leftIcon={<Icon name="add" size={20} />}
          onClick={() => setIsModalOpen(true)}
        >
          Tambah Kategori
        </Button>
      </div>

      {/* Stats */}
      <Card variant="bordered" padding="md">
        <div className="flex items-center justify-between">
          <div>
            <p className="text-xs text-text-sub dark:text-gray-400 mb-1">Total Kategori</p>
            <p className="text-2xl font-bold text-text-main dark:text-white">
              {categoriesList.length}
            </p>
          </div>
          <div className="w-12 h-12 rounded-full bg-blue-50 dark:bg-blue-900/20 flex items-center justify-center">
            <Icon name="label" size={24} className="text-primary" />
          </div>
        </div>
      </Card>

      {/* Table */}
      <Card variant="bordered" padding="none">
        <DataTable
          data={categoriesList}
          columns={categoryColumns}
          loading={isLoading}
          emptyMessage="Belum ada kategori pengeluaran"
        />
      </Card>

      {/* Add Category Modal */}
      <CategoryFormModal
        isOpen={isModalOpen}
        onClose={() => setIsModalOpen(false)}
        onSubmit={handleAddCategory}
        isLoading={createCategory.isPending}
      />

      {/* Edit Category Modal */}
      <CategoryFormModal
        key={editingCategory?.id}
        isOpen={!!editingCategory}
        onClose={() => setEditingCategory(null)}
        onSubmit={handleEditCategory}
        isLoading={updateCategory.isPending}
        initialData={editingCategory ?? undefined}
      />
    </div>
  );
}

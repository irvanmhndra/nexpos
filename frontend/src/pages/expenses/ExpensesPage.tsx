import { useState } from 'react';
import toast from 'react-hot-toast';
import { Card, Button, Input, Dropdown, Icon } from '@/components/ui';
import { DataTable } from '@/components/common';
import {
  useExpenses,
  useCreateExpense,
  useUpdateExpense,
  useDeleteExpense,
  useExpenseCategories,
} from '@/hooks/useExpenses';
import { useBranch } from '@/contexts/BranchContext';
import { formatCurrency, formatDate, getApiErrorMessage } from '@/utils';
import type { Expense, CreateExpenseRequest, UpdateExpenseRequest } from '@/types/expense';

interface ExpenseFormData {
  category_id: string;
  amount: string;
  description: string;
  reference_no: string;
  expense_date: string;
  notes: string;
}

interface ExpenseFormErrors {
  amount?: string;
  description?: string;
  expense_date?: string;
}

function ExpenseFormModal({
  isOpen,
  onClose,
  onSubmit,
  isLoading,
  initialData,
  categories,
}: {
  isOpen: boolean;
  onClose: () => void;
  onSubmit: (data: CreateExpenseRequest) => void;
  isLoading: boolean;
  initialData?: Expense;
  categories: { id: number; name: string }[];
}) {
  const today = new Date().toISOString().split('T')[0];
  const [formData, setFormData] = useState<ExpenseFormData>({
    category_id: initialData?.category_id ? String(initialData.category_id) : '',
    amount: initialData?.amount ? String(initialData.amount) : '',
    description: initialData?.description ?? '',
    reference_no: initialData?.reference_no ?? '',
    expense_date: initialData?.expense_date ?? today,
    notes: initialData?.notes ?? '',
  });
  const [errors, setErrors] = useState<ExpenseFormErrors>({});

  const validateForm = (): boolean => {
    const newErrors: ExpenseFormErrors = {};

    if (!formData.amount || parseFloat(formData.amount) <= 0) {
      newErrors.amount = 'Jumlah pengeluaran wajib diisi';
    }

    if (!formData.description.trim()) {
      newErrors.description = 'Keterangan wajib diisi';
    }

    if (!formData.expense_date) {
      newErrors.expense_date = 'Tanggal wajib diisi';
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleChange =
    (field: keyof ExpenseFormData) =>
    (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>) => {
      setFormData(prev => ({ ...prev, [field]: e.target.value }));
      if (errors[field as keyof ExpenseFormErrors]) {
        setErrors(prev => ({ ...prev, [field]: undefined }));
      }
    };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (validateForm()) {
      onSubmit({
        category_id: formData.category_id ? Number(formData.category_id) : undefined,
        amount: parseFloat(formData.amount),
        description: formData.description,
        reference_no: formData.reference_no || undefined,
        expense_date: formData.expense_date,
        notes: formData.notes || undefined,
      });
    }
  };

  const handleClose = () => {
    setFormData({
      category_id: '',
      amount: '',
      description: '',
      reference_no: '',
      expense_date: today,
      notes: '',
    });
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
                {initialData ? 'Edit Pengeluaran' : 'Tambah Pengeluaran'}
              </h2>
              <p className="text-sm text-text-sub dark:text-gray-400 mt-1">
                {initialData ? 'Edit data pengeluaran' : 'Catat pengeluaran baru'}
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
                  label="Jumlah"
                  type="number"
                  min="0"
                  placeholder="0"
                  value={formData.amount}
                  onChange={handleChange('amount')}
                  error={errors.amount}
                  fullWidth
                  required
                />
                <Input
                  label="Tanggal"
                  type="date"
                  value={formData.expense_date}
                  onChange={handleChange('expense_date')}
                  error={errors.expense_date}
                  fullWidth
                  required
                />
              </div>

              <Input
                label="Keterangan"
                placeholder="Keterangan pengeluaran"
                value={formData.description}
                onChange={handleChange('description')}
                error={errors.description}
                fullWidth
                required
              />

              <div>
                <label className="block text-sm font-medium text-text-main dark:text-white mb-1">
                  Kategori
                </label>
                <select
                  value={formData.category_id}
                  onChange={handleChange('category_id')}
                  className="w-full rounded-lg border border-border-light dark:border-border-dark bg-white dark:bg-background-dark px-3 py-2 text-sm text-text-main dark:text-white focus:outline-none focus:ring-2 focus:ring-primary"
                >
                  <option value="">-- Tanpa Kategori --</option>
                  {categories.map(cat => (
                    <option key={cat.id} value={cat.id}>
                      {cat.name}
                    </option>
                  ))}
                </select>
              </div>

              <Input
                label="No. Referensi"
                placeholder="Nomor kwitansi atau referensi (opsional)"
                value={formData.reference_no}
                onChange={handleChange('reference_no')}
                helperText="Opsional"
                fullWidth
              />

              <div>
                <label className="block text-sm font-medium text-text-main dark:text-white mb-1">
                  Catatan
                </label>
                <textarea
                  className="w-full rounded-lg border border-border-light dark:border-border-dark bg-white dark:bg-background-dark px-3 py-2 text-sm text-text-main dark:text-white placeholder:text-text-sub focus:outline-none focus:ring-2 focus:ring-primary"
                  rows={2}
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
                {isLoading ? 'Menyimpan...' : initialData ? 'Simpan Perubahan' : 'Simpan Pengeluaran'}
              </Button>
            </div>
          </form>
        </div>
      </div>
    </div>
  );
}

export function ExpensesPage() {
  const [page, setPage] = useState(1);
  const [categoryFilter, setCategoryFilter] = useState<string>('');
  const [dateFrom, setDateFrom] = useState('');
  const [dateTo, setDateTo] = useState('');
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [editingExpense, setEditingExpense] = useState<Expense | null>(null);

  const { activeBranch } = useBranch();

  const { data, isLoading, refetch } = useExpenses({
    page,
    per_page: 10,
    category_id: categoryFilter ? Number(categoryFilter) : undefined,
    branch_id: activeBranch?.id,
    date_from: dateFrom || undefined,
    date_to: dateTo || undefined,
  });

  const { data: categoriesData } = useExpenseCategories();
  const createExpense = useCreateExpense();
  const updateExpense = useUpdateExpense();
  const deleteExpense = useDeleteExpense();

  const categories = categoriesData || [];
  const expenses = data?.data || [];

  const totalAmount = expenses.reduce((sum, e) => sum + e.amount, 0);

  const handleAddExpense = async (formData: CreateExpenseRequest) => {
    try {
      await createExpense.mutateAsync({
        ...formData,
        branch_id: activeBranch?.id,
      });
      toast.success('Pengeluaran berhasil dicatat');
      setIsModalOpen(false);
    } catch (err) {
      toast.error(getApiErrorMessage(err));
    }
  };

  const handleEditExpense = async (formData: UpdateExpenseRequest) => {
    if (!editingExpense) return;
    try {
      await updateExpense.mutateAsync({ id: editingExpense.id, expense: formData });
      toast.success('Pengeluaran berhasil diperbarui');
      setEditingExpense(null);
    } catch (err) {
      toast.error(getApiErrorMessage(err));
    }
  };

  const handleDeleteExpense = async (id: number) => {
    if (window.confirm('Apakah Anda yakin ingin menghapus pengeluaran ini?')) {
      try {
        await deleteExpense.mutateAsync(id);
        toast.success('Pengeluaran berhasil dihapus');
      } catch (err) {
        toast.error(getApiErrorMessage(err));
      }
    }
  };

  const categoryOptions = [
    { value: '', label: 'Semua Kategori' },
    ...categories.map(c => ({ value: String(c.id), label: c.name })),
  ];

  const expenseColumns = [
    {
      header: 'Tanggal',
      accessor: (row: Expense) => (
        <span className="text-sm text-text-sub dark:text-gray-400">
          {formatDate(new Date(row.expense_date), 'dd MMM yyyy')}
        </span>
      ),
    },
    {
      header: 'Keterangan',
      accessor: (row: Expense) => (
        <div>
          <p className="text-sm font-medium text-text-main dark:text-white">{row.description}</p>
          {row.reference_no && (
            <p className="text-xs text-text-sub dark:text-gray-400 font-mono">{row.reference_no}</p>
          )}
        </div>
      ),
    },
    {
      header: 'Kategori',
      accessor: (row: Expense) => (
        <span className="text-sm text-text-main dark:text-gray-200">
          {row.category?.name || '-'}
        </span>
      ),
    },
    {
      header: 'Jumlah',
      accessor: (row: Expense) => (
        <span className="text-sm font-semibold text-text-main dark:text-white">
          {formatCurrency(row.amount)}
        </span>
      ),
    },
    {
      header: 'Aksi',
      accessor: (row: Expense) => (
        <div className="flex items-center gap-2">
          <button
            className="p-1.5 hover:bg-background-light dark:hover:bg-background-dark rounded-lg transition-colors"
            onClick={e => {
              e.stopPropagation();
              setEditingExpense(row);
            }}
            title="Edit"
          >
            <Icon name="edit" size={18} className="text-text-sub dark:text-gray-400" />
          </button>
          <button
            className="p-1.5 hover:bg-red-50 dark:hover:bg-red-900/20 rounded-lg transition-colors"
            onClick={e => {
              e.stopPropagation();
              handleDeleteExpense(row.id);
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
            Pengeluaran
          </h1>
          <p className="text-text-sub dark:text-gray-400 text-sm mt-1">
            Catat dan kelola pengeluaran operasional
          </p>
        </div>
        <Button
          leftIcon={<Icon name="add" size={20} />}
          onClick={() => setIsModalOpen(true)}
        >
          Tambah Pengeluaran
        </Button>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
        <Card variant="bordered" padding="md">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-xs text-text-sub dark:text-gray-400 mb-1">Total Pengeluaran</p>
              <p className="text-2xl font-bold text-text-main dark:text-white">
                {data?.total || 0}
              </p>
            </div>
            <div className="w-12 h-12 rounded-full bg-red-50 dark:bg-red-900/20 flex items-center justify-center">
              <Icon name="receipt" size={24} className="text-red-500" />
            </div>
          </div>
        </Card>

        <Card variant="bordered" padding="md">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-xs text-text-sub dark:text-gray-400 mb-1">Total Nominal</p>
              <p className="text-2xl font-bold text-red-600">{formatCurrency(totalAmount)}</p>
            </div>
            <div className="w-12 h-12 rounded-full bg-red-50 dark:bg-red-900/20 flex items-center justify-center">
              <Icon name="payments" size={24} className="text-red-500" />
            </div>
          </div>
        </Card>

        <Card variant="bordered" padding="md">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-xs text-text-sub dark:text-gray-400 mb-1">Kategori</p>
              <p className="text-2xl font-bold text-text-main dark:text-white">
                {categories.length}
              </p>
            </div>
            <div className="w-12 h-12 rounded-full bg-blue-50 dark:bg-blue-900/20 flex items-center justify-center">
              <Icon name="label" size={24} className="text-primary" />
            </div>
          </div>
        </Card>
      </div>

      {/* Filters */}
      <Card variant="bordered" padding="md">
        <div className="grid grid-cols-1 sm:grid-cols-4 gap-4">
          <Dropdown
            value={categoryFilter}
            onChange={v => { setCategoryFilter(v); setPage(1); }}
            options={categoryOptions}
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
              setCategoryFilter('');
              setDateFrom('');
              setDateTo('');
              setPage(1);
              refetch();
            }}
          >
            Reset Filter
          </Button>
        </div>
      </Card>

      {/* Table */}
      <Card variant="bordered" padding="none">
        <DataTable
          data={expenses}
          columns={expenseColumns}
          loading={isLoading}
          emptyMessage="Belum ada data pengeluaran"
        />

        {/* Pagination */}
        {data && data.totalPages > 1 && (
          <div className="px-6 py-4 border-t border-border-light dark:border-border-dark">
            <div className="flex items-center justify-between">
              <p className="text-sm text-text-sub dark:text-gray-400">
                Menampilkan {(page - 1) * 10 + 1} - {Math.min(page * 10, data.total)} dari{' '}
                {data.total} pengeluaran
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

      {/* Add Expense Modal */}
      <ExpenseFormModal
        isOpen={isModalOpen}
        onClose={() => setIsModalOpen(false)}
        onSubmit={handleAddExpense}
        isLoading={createExpense.isPending}
        categories={categories}
      />

      {/* Edit Expense Modal */}
      <ExpenseFormModal
        key={editingExpense?.id}
        isOpen={!!editingExpense}
        onClose={() => setEditingExpense(null)}
        onSubmit={handleEditExpense}
        isLoading={updateExpense.isPending}
        initialData={editingExpense ?? undefined}
        categories={categories}
      />
    </div>
  );
}

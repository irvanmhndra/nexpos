import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Card, Badge, Button, Input, Dropdown, Icon } from '@/components/ui';
import { useInventory, useInventoryStats, useAdjustStock } from '@/hooks/useInventory';
import { useBranch } from '@/contexts/BranchContext';
import { formatCurrency } from '@/utils';
import { INVENTORY_STATUS_LABELS, MOVEMENT_TYPE_LABELS, ROUTES } from '@/utils/constants';
import type { InventoryItem, StockMovementType } from '@/types/inventory';

const STATUS_BADGE_MAP: Record<string, 'success' | 'warning' | 'danger'> = {
  normal: 'success',
  low: 'warning',
  out_of_stock: 'danger',
};

export function StockOverviewPage() {
  const navigate = useNavigate();
  const [page, setPage] = useState(1);
  const [search, setSearch] = useState('');
  const [categoryFilter, setCategoryFilter] = useState('');
  const [statusFilter, setStatusFilter] = useState('');
  const [expandedProducts, setExpandedProducts] = useState<Set<string>>(new Set());

  // Adjust stock modal
  const [isAdjustOpen, setIsAdjustOpen] = useState(false);
  const [selectedItem, setSelectedItem] = useState<InventoryItem | null>(null);
  const [adjustType, setAdjustType] = useState<StockMovementType>('IN');
  const [adjustQty, setAdjustQty] = useState('');
  const [adjustNotes, setAdjustNotes] = useState('');

  const { data, isLoading } = useInventory({
    page,
    per_page: 10,
    search: search || undefined,
    category: categoryFilter || undefined,
    status: statusFilter as InventoryItem['status'] || undefined,
  });

  const { activeBranch } = useBranch();
  const { data: stats } = useInventoryStats();
  const adjustStock = useAdjustStock();

  const items = data?.data || [];

  // Group items by product for expandable rows
  const groupedItems = items.reduce<Record<string, InventoryItem[]>>((acc, item) => {
    if (!acc[item.product_id]) acc[item.product_id] = [];
    acc[item.product_id].push(item);
    return acc;
  }, {});

  // Build unique categories from items
  const categories = [...new Set(items.map(i => i.category_name))].filter(Boolean);
  const categoryOptions = [
    { value: '', label: 'Semua Kategori' },
    ...categories.map(c => ({ value: c, label: c })),
  ];

  const statusOptions = [
    { value: '', label: 'Semua Status' },
    { value: 'normal', label: 'Normal' },
    { value: 'low', label: 'Menipis' },
    { value: 'out_of_stock', label: 'Habis' },
  ];

  const toggleExpand = (productId: string) => {
    const next = new Set(expandedProducts);
    if (next.has(productId)) next.delete(productId);
    else next.add(productId);
    setExpandedProducts(next);
  };

  const handleSearch = (value: string) => {
    setSearch(value);
    setPage(1);
  };

  const openAdjustModal = (item: InventoryItem) => {
    setSelectedItem(item);
    setAdjustType('IN');
    setAdjustQty('');
    setAdjustNotes('');
    setIsAdjustOpen(true);
  };

  const handleAdjustSubmit = async () => {
    if (!selectedItem || !adjustQty || !activeBranch) return;
    try {
      await adjustStock.mutateAsync({
        variant_id: selectedItem.product_variant_id,
        branch_id: activeBranch.id,
        type: adjustType,
        quantity: parseInt(adjustQty),
        note: adjustNotes || undefined,
      });
      setIsAdjustOpen(false);
    } catch (error) {
      console.error('Gagal adjust stok:', error);
    }
  };

  // For products with multiple variants: show one "parent" row and expand
  // For single-variant products: show directly
  const renderProductGroup = (productId: string, variants: InventoryItem[]) => {
    const hasMultiple = variants.length > 1 && variants[0].has_variants;
    const isExpanded = expandedProducts.has(productId);
    const first = variants[0];

    if (!hasMultiple) {
      // Single variant — render flat row
      return (
        <tr key={first.product_variant_id} className="hover:bg-background-light dark:hover:bg-background-dark transition-colors">
          <td className="px-4 py-4 w-10"></td>
          <td className="px-4 py-4">
            <div className="flex items-center gap-3">
              {first.image_data ? (
                <img src={first.image_data} alt={first.product_name} className="w-10 h-10 rounded-lg object-cover border border-border-light dark:border-border-dark" />
              ) : (
                <div className="w-10 h-10 rounded-lg bg-gray-100 dark:bg-gray-800 flex items-center justify-center">
                  <Icon name="image" size={20} className="text-gray-400" />
                </div>
              )}
              <div>
                <p className="font-medium text-text-main dark:text-white">{first.product_name}</p>
              </div>
            </div>
          </td>
          <td className="px-4 py-4 font-mono text-xs text-text-sub dark:text-gray-400">{first.sku}</td>
          <td className="px-4 py-4 text-text-main dark:text-gray-200">{first.category_name}</td>
          <td className="px-4 py-4 text-text-main dark:text-white font-medium">{first.current_stock}</td>
          <td className="px-4 py-4 text-text-sub dark:text-gray-400">{first.min_stock}</td>
          <td className="px-4 py-4">
            <Badge variant={STATUS_BADGE_MAP[first.status]} dot size="sm">
              {INVENTORY_STATUS_LABELS[first.status]}
            </Badge>
          </td>
          <td className="px-4 py-4 text-right">
            <div className="flex items-center justify-end gap-1">
              <button
                className="p-1.5 hover:bg-background-light dark:hover:bg-gray-700 rounded-lg transition-colors"
                onClick={() => navigate(`/products/${first.product_id}/edit`)}
                title="Edit Produk"
              >
                <Icon name="edit" size={18} className="text-text-sub dark:text-gray-400" />
              </button>
              <button
                className="p-1.5 hover:bg-background-light dark:hover:bg-gray-700 rounded-lg transition-colors"
                onClick={() => openAdjustModal(first)}
                title="Adjust Stok"
              >
                <Icon name="swap_vert" size={18} className="text-text-sub dark:text-gray-400" />
              </button>
            </div>
          </td>
        </tr>
      );
    }

    // Multi-variant: parent row + expandable children
    const totalStock = variants.reduce((sum, v) => sum + v.current_stock, 0);
    const worstStatus = variants.some(v => v.status === 'out_of_stock')
      ? 'out_of_stock'
      : variants.some(v => v.status === 'low')
        ? 'low'
        : 'normal';

    return (
      <tbody key={productId}>
        <tr className="hover:bg-background-light dark:hover:bg-background-dark transition-colors">
          <td className="px-4 py-4">
            <button
              onClick={() => toggleExpand(productId)}
              className="p-1 hover:bg-gray-200 dark:hover:bg-gray-700 rounded transition-colors"
            >
              <Icon
                name={isExpanded ? 'expand_less' : 'expand_more'}
                size={20}
                className="text-text-sub dark:text-gray-400"
              />
            </button>
          </td>
          <td className="px-4 py-4">
            <div className="flex items-center gap-3">
              {first.image_data ? (
                <img src={first.image_data} alt={first.product_name} className="w-10 h-10 rounded-lg object-cover border border-border-light dark:border-border-dark" />
              ) : (
                <div className="w-10 h-10 rounded-lg bg-gray-100 dark:bg-gray-800 flex items-center justify-center">
                  <Icon name="image" size={20} className="text-gray-400" />
                </div>
              )}
              <div>
                <p className="font-medium text-text-main dark:text-white">{first.product_name}</p>
                <button
                  onClick={() => toggleExpand(productId)}
                  className="inline-flex items-center gap-1 text-xs text-primary font-medium hover:underline"
                >
                  <Icon name="layers" size={14} />
                  {variants.length} varian
                </button>
              </div>
            </div>
          </td>
          <td className="px-4 py-4 text-xs text-text-sub dark:text-gray-400">-</td>
          <td className="px-4 py-4 text-text-main dark:text-gray-200">{first.category_name}</td>
          <td className="px-4 py-4 text-text-main dark:text-white font-medium">{totalStock}</td>
          <td className="px-4 py-4 text-text-sub dark:text-gray-400">-</td>
          <td className="px-4 py-4">
            <Badge variant={STATUS_BADGE_MAP[worstStatus]} dot size="sm">
              {INVENTORY_STATUS_LABELS[worstStatus]}
            </Badge>
          </td>
          <td className="px-4 py-4"></td>
        </tr>

        {isExpanded && variants.map(variant => (
          <tr key={variant.product_variant_id} className="bg-background-light dark:bg-background-dark">
            <td className="px-4 py-3"></td>
            <td className="px-4 py-3 pl-16">
              <p className="text-sm text-text-main dark:text-white">{variant.variant_name}</p>
            </td>
            <td className="px-4 py-3 font-mono text-xs text-text-sub dark:text-gray-400">{variant.sku}</td>
            <td className="px-4 py-3"></td>
            <td className="px-4 py-3 text-text-main dark:text-white font-medium">{variant.current_stock}</td>
            <td className="px-4 py-3 text-text-sub dark:text-gray-400">{variant.min_stock}</td>
            <td className="px-4 py-3">
              <Badge variant={STATUS_BADGE_MAP[variant.status]} dot size="sm">
                {INVENTORY_STATUS_LABELS[variant.status]}
              </Badge>
            </td>
            <td className="px-4 py-3 text-right">
              <button
                className="p-1.5 hover:bg-gray-200 dark:hover:bg-gray-700 rounded-lg transition-colors"
                onClick={() => openAdjustModal(variant)}
                title="Adjust Stok"
              >
                <Icon name="swap_vert" size={18} className="text-text-sub dark:text-gray-400" />
              </button>
            </td>
          </tr>
        ))}
      </tbody>
    );
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h1 className="text-2xl md:text-3xl font-bold text-text-main dark:text-white">Inventaris</h1>
          <p className="text-text-sub dark:text-gray-400 text-sm mt-1">
            {activeBranch ? `Cabang: ${activeBranch.name}` : 'Kelola stok dan inventori produk'}
          </p>
        </div>
        <Button
          variant="outline"
          leftIcon={<Icon name="history" size={20} />}
          onClick={() => navigate(ROUTES.INVENTORY_MOVEMENTS)}
        >
          Riwayat Pergerakan
        </Button>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <Card variant="bordered" padding="md">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-xs text-text-sub dark:text-gray-400 mb-1">Total SKU</p>
              <p className="text-2xl font-bold text-text-main dark:text-white">{stats?.total_sku ?? 0}</p>
            </div>
            <div className="w-12 h-12 rounded-full bg-blue-50 dark:bg-blue-900/20 flex items-center justify-center">
              <Icon name="inventory_2" size={24} className="text-primary" />
            </div>
          </div>
        </Card>

        <Card variant="bordered" padding="md">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-xs text-text-sub dark:text-gray-400 mb-1">Stok Menipis</p>
              <p className="text-2xl font-bold text-yellow-600">{stats?.low_stock ?? 0}</p>
            </div>
            <div className="w-12 h-12 rounded-full bg-yellow-50 dark:bg-yellow-900/20 flex items-center justify-center">
              <Icon name="warning" size={24} className="text-yellow-600" />
            </div>
          </div>
        </Card>

        <Card variant="bordered" padding="md">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-xs text-text-sub dark:text-gray-400 mb-1">Stok Habis</p>
              <p className="text-2xl font-bold text-red-600">{stats?.out_of_stock ?? 0}</p>
            </div>
            <div className="w-12 h-12 rounded-full bg-red-50 dark:bg-red-900/20 flex items-center justify-center">
              <Icon name="error" size={24} className="text-red-600" />
            </div>
          </div>
        </Card>

        <Card variant="bordered" padding="md">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-xs text-text-sub dark:text-gray-400 mb-1">Nilai Stok</p>
              <p className="text-2xl font-bold text-text-main dark:text-white">
                {formatCurrency(stats?.total_stock_value ?? 0)}
              </p>
            </div>
            <div className="w-12 h-12 rounded-full bg-green-50 dark:bg-green-900/20 flex items-center justify-center">
              <Icon name="paid" size={24} className="text-green-600" />
            </div>
          </div>
        </Card>
      </div>

      {/* Filters */}
      <Card variant="bordered" padding="md">
        <div className="grid grid-cols-1 sm:grid-cols-4 gap-4">
          <Input
            placeholder="Cari produk / SKU..."
            leftIcon={<Icon name="search" size={20} />}
            value={search}
            onChange={e => handleSearch(e.target.value)}
          />
          <Dropdown
            value={categoryFilter}
            onChange={v => { setCategoryFilter(v); setPage(1); }}
            options={categoryOptions}
            leftIcon={<Icon name="category" size={20} />}
          />
          <Dropdown
            value={statusFilter}
            onChange={v => { setStatusFilter(v); setPage(1); }}
            options={statusOptions}
            leftIcon={<Icon name="filter_list" size={20} />}
          />
          <Button
            variant="outline"
            leftIcon={<Icon name="refresh" size={20} />}
            onClick={() => { setSearch(''); setCategoryFilter(''); setStatusFilter(''); setPage(1); }}
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
                <th className="w-10 px-4 py-3"></th>
                <th className="px-4 py-3 text-left text-xs font-semibold text-text-sub dark:text-gray-400 uppercase tracking-wider">Produk</th>
                <th className="px-4 py-3 text-left text-xs font-semibold text-text-sub dark:text-gray-400 uppercase tracking-wider">SKU</th>
                <th className="px-4 py-3 text-left text-xs font-semibold text-text-sub dark:text-gray-400 uppercase tracking-wider">Kategori</th>
                <th className="px-4 py-3 text-left text-xs font-semibold text-text-sub dark:text-gray-400 uppercase tracking-wider">Stok</th>
                <th className="px-4 py-3 text-left text-xs font-semibold text-text-sub dark:text-gray-400 uppercase tracking-wider">Min Stok</th>
                <th className="px-4 py-3 text-left text-xs font-semibold text-text-sub dark:text-gray-400 uppercase tracking-wider">Status</th>
                <th className="px-4 py-3 text-right text-xs font-semibold text-text-sub dark:text-gray-400 uppercase tracking-wider">Aksi</th>
              </tr>
            </thead>
            {isLoading ? (
              <tbody>
                <tr>
                  <td colSpan={8} className="px-4 py-12 text-center">
                    <Icon name="progress_activity" size={32} className="animate-spin text-primary mx-auto" />
                    <p className="text-text-sub dark:text-gray-400 mt-2">Memuat data...</p>
                  </td>
                </tr>
              </tbody>
            ) : items.length === 0 ? (
              <tbody>
                <tr>
                  <td colSpan={8} className="px-4 py-12 text-center">
                    <Icon name="inventory_2" size={48} className="text-text-sub dark:text-gray-400 mx-auto mb-2" />
                    <p className="text-text-sub dark:text-gray-400">Belum ada data inventaris</p>
                  </td>
                </tr>
              </tbody>
            ) : (
              Object.entries(groupedItems).map(([productId, variants]) => {
                const hasMultiple = variants.length > 1 && variants[0].has_variants;
                if (!hasMultiple) {
                  return (
                    <tbody key={productId}>
                      {renderProductGroup(productId, variants)}
                    </tbody>
                  );
                }
                return renderProductGroup(productId, variants);
              })
            )}
          </table>
        </div>

        {/* Pagination */}
        {data && data.totalPages > 1 && (
          <div className="px-6 py-4 border-t border-border-light dark:border-border-dark">
            <div className="flex items-center justify-between">
              <p className="text-sm text-text-sub dark:text-gray-400">
                Menampilkan {(page - 1) * 10 + 1} - {Math.min(page * 10, data.total)} dari {data.total} item
              </p>
              <div className="flex items-center gap-2">
                <Button variant="outline" size="sm" disabled={page === 1} onClick={() => setPage(p => p - 1)}>
                  <Icon name="chevron_left" size={20} />
                </Button>
                <span className="text-sm text-text-main dark:text-white px-2">{page} / {data.totalPages}</span>
                <Button variant="outline" size="sm" disabled={page === data.totalPages} onClick={() => setPage(p => p + 1)}>
                  <Icon name="chevron_right" size={20} />
                </Button>
              </div>
            </div>
          </div>
        )}
      </Card>

      {/* Adjust Stock Modal */}
      {isAdjustOpen && selectedItem && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/50">
          <div className="bg-white dark:bg-card-dark rounded-xl shadow-xl w-full max-w-md">
            {/* Header */}
            <div className="flex items-center gap-3 p-6 border-b border-border-light dark:border-border-dark">
              <div className="flex items-center justify-center w-12 h-12 rounded-full bg-blue-100 dark:bg-blue-900/30">
                <Icon name="tune" size={24} className="text-primary" />
              </div>
              <div className="flex-1">
                <h2 className="text-lg font-bold text-text-main dark:text-white">Adjust Stok</h2>
                <p className="text-sm text-text-sub dark:text-gray-400">{selectedItem.product_name} — {selectedItem.variant_name}</p>
              </div>
              <button onClick={() => setIsAdjustOpen(false)} className="text-text-sub hover:text-text-main dark:hover:text-white transition-colors">
                <Icon name="close" size={24} />
              </button>
            </div>

            {/* Body */}
            <div className="p-6 space-y-4">
              <div>
                <p className="text-sm text-text-sub dark:text-gray-400 mb-1">Stok Saat Ini</p>
                <p className="text-2xl font-bold text-text-main dark:text-white">{selectedItem.current_stock}</p>
              </div>

              <div>
                <label className="block text-sm font-medium text-text-main dark:text-white mb-1">Tipe</label>
                <Dropdown
                  value={adjustType}
                  onChange={v => setAdjustType(v as StockMovementType)}
                  options={[
                    { value: 'IN', label: MOVEMENT_TYPE_LABELS.IN },
                    { value: 'OUT', label: MOVEMENT_TYPE_LABELS.OUT },
                    { value: 'ADJUST', label: MOVEMENT_TYPE_LABELS.ADJUST },
                  ]}
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-text-main dark:text-white mb-1">
                  {adjustType === 'ADJUST' ? 'Stok Baru' : 'Jumlah'}
                </label>
                <Input
                  type="number"
                  min="0"
                  placeholder={adjustType === 'ADJUST' ? 'Masukkan stok baru' : 'Masukkan jumlah'}
                  value={adjustQty}
                  onChange={e => setAdjustQty(e.target.value)}
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-text-main dark:text-white mb-1">Catatan</label>
                <textarea
                  className="w-full rounded-lg border border-border-light dark:border-border-dark bg-white dark:bg-background-dark px-3 py-2 text-sm text-text-main dark:text-white placeholder:text-text-sub focus:outline-none focus:ring-2 focus:ring-primary"
                  rows={3}
                  placeholder="Catatan opsional..."
                  value={adjustNotes}
                  onChange={e => setAdjustNotes(e.target.value)}
                />
              </div>
            </div>

            {/* Footer */}
            <div className="flex items-center justify-end gap-3 p-6 border-t border-border-light dark:border-border-dark">
              <Button variant="outline" onClick={() => setIsAdjustOpen(false)} disabled={adjustStock.isPending}>
                Batal
              </Button>
              <Button onClick={handleAdjustSubmit} disabled={adjustStock.isPending || !adjustQty || !activeBranch}>
                {adjustStock.isPending ? 'Menyimpan...' : 'Simpan'}
              </Button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

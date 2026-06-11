import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import toast from 'react-hot-toast';
import { Card, Badge, Button, Input, Dropdown, Icon } from '@/components/ui';
import { DeleteConfirmModal } from '@/components/common';
import { BarcodeScannerModal } from '@/components/features/pos/BarcodeScannerModal';
import { useProducts, useDeleteProduct } from '@/hooks/useProducts';
import { useAllProductCategories } from '@/hooks/useProductCategories';
import { productsService } from '@/services/productsService';
import { formatCurrency } from '@/utils';
import type { Product, ProductVariant } from '@/types/product';

// Helper functions
function getPriceRange(product: Product): { min: number; max: number; isSingle: boolean } {
  const prices = product.variants?.map(v => v.price) || [];
  if (prices.length === 0) return { min: 0, max: 0, isSingle: true };
  const min = Math.min(...prices);
  const max = Math.max(...prices);
  return { min, max, isSingle: min === max };
}

export function ProductsPage() {
  const navigate = useNavigate();
  const [page, setPage] = useState(1);
  const [search, setSearch] = useState('');
  const [categoryId, setCategoryId] = useState<string>('');
  const [expandedProducts, setExpandedProducts] = useState<Set<number>>(new Set());

  // Modal states
  const [isDeleteModalOpen, setIsDeleteModalOpen] = useState(false);
  const [productToDelete, setProductToDelete] = useState<Product | null>(null);
  const [isScanModalOpen, setIsScanModalOpen] = useState(false);

  const { data, isLoading } = useProducts({
    page,
    per_page: 10,
    search: search || undefined,
    category_id: categoryId ? parseInt(categoryId) : undefined,
  });

  const { data: categories } = useAllProductCategories();

  const deleteProduct = useDeleteProduct();

  const toggleExpand = (productId: number) => {
    const newExpanded = new Set(expandedProducts);
    if (newExpanded.has(productId)) {
      newExpanded.delete(productId);
    } else {
      newExpanded.add(productId);
    }
    setExpandedProducts(newExpanded);
  };

  const handleSearch = (value: string) => {
    setSearch(value);
    setPage(1);
  };

  const handleCategoryChange = (value: string) => {
    setCategoryId(value);
    setPage(1);
  };

  // CRUD handlers
  const handleAddProduct = () => {
    navigate('/products/new');
  };

  // Add product via barcode scan: if the code already maps to a product,
  // warn and jump to it (validation against duplicates); otherwise open the
  // create form with the barcode prefilled.
  const handleScanToAdd = async (code: string) => {
    setIsScanModalOpen(false);
    const trimmed = code.trim();
    if (!trimmed) return;

    try {
      const res = await productsService.lookupByCode(trimmed);
      const matchedVariant = res.product.variants?.find(v => v.id === res.matched_variant_id);
      const label = res.matched_by === 'barcode' ? 'barcode' : 'SKU';
      toast.error(
        `Produk "${res.product.name}"${matchedVariant ? ` (${matchedVariant.name})` : ''} sudah terdaftar dengan ${label} ini`,
      );
      navigate(`/products/${res.product.id}/edit`);
    } catch (err) {
      const status = (err as { response?: { status?: number } }).response?.status;
      if (status === 404) {
        toast.success('Barcode belum terdaftar — silakan lengkapi produk baru');
        navigate(`/products/new?barcode=${encodeURIComponent(trimmed)}`);
      } else {
        toast.error('Gagal memeriksa barcode. Coba lagi.');
      }
    }
  };

  const handleEditProduct = (product: Product) => {
    navigate(`/products/${product.id}/edit`);
  };

  const handleDeleteClick = (product: Product) => {
    setProductToDelete(product);
    setIsDeleteModalOpen(true);
  };

  const handleDeleteConfirm = async () => {
    if (!productToDelete) return;

    try {
      await deleteProduct.mutateAsync(productToDelete.id);
      setIsDeleteModalOpen(false);
      setProductToDelete(null);
    } catch (error) {
      console.error('Gagal menghapus produk:', error);
    }
  };

  // Calculate stats
  const products = data?.data || [];
  const totalProducts = data?.total || 0;
  const activeProducts = products.filter(p => p.is_active).length;
  const inactiveProducts = products.filter(p => !p.is_active).length;

  // Build category dropdown options
  const categoryOptions = [
    { value: '', label: 'Semua Kategori' },
    ...(categories || []).map(cat => ({
      value: cat.id.toString(),
      label: cat.name,
    })),
  ];

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h1 className="text-2xl md:text-3xl font-bold text-text-main dark:text-white">
            Produk
          </h1>
          <p className="text-text-sub dark:text-gray-400 text-sm mt-1">
            Kelola produk dan inventori Anda
          </p>
        </div>
        <div className="flex items-center gap-2">
          <Button
            variant="outline"
            leftIcon={<Icon name="qr_code_scanner" size={20} />}
            onClick={() => setIsScanModalOpen(true)}
          >
            Scan Barcode
          </Button>
          <Button leftIcon={<Icon name="add" size={20} />} onClick={handleAddProduct}>
            Tambah Produk
          </Button>
        </div>
      </div>

      {/* Stats Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
        <Card variant="bordered" padding="md">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-xs text-text-sub dark:text-gray-400 mb-1">Total Produk</p>
              <p className="text-2xl font-bold text-text-main dark:text-white">
                {totalProducts}
              </p>
            </div>
            <div className="w-12 h-12 rounded-full bg-blue-50 dark:bg-blue-900/20 flex items-center justify-center">
              <Icon name="inventory_2" size={24} className="text-primary" />
            </div>
          </div>
        </Card>

        <Card variant="bordered" padding="md">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-xs text-text-sub dark:text-gray-400 mb-1">Aktif</p>
              <p className="text-2xl font-bold text-text-main dark:text-white">
                {activeProducts}
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
              <p className="text-xs text-text-sub dark:text-gray-400 mb-1">Nonaktif</p>
              <p className="text-2xl font-bold text-text-main dark:text-white">
                {inactiveProducts}
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
            placeholder="Cari produk..."
            leftIcon={<Icon name="search" size={20} />}
            value={search}
            onChange={e => handleSearch(e.target.value)}
          />
          <Dropdown
            value={categoryId}
            onChange={handleCategoryChange}
            options={categoryOptions}
            leftIcon={<Icon name="category" size={20} />}
          />
          <Button
            variant="outline"
            leftIcon={<Icon name="refresh" size={20} />}
            onClick={() => {
              setSearch('');
              setCategoryId('');
              setPage(1);
            }}
          >
            Reset Filter
          </Button>
        </div>
      </Card>

      {/* Products Table with Expandable Rows */}
      <Card variant="bordered" padding="none">
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead>
              <tr className="bg-background-light dark:bg-background-dark border-b border-border-light dark:border-border-dark">
                <th className="w-10 px-4 py-3"></th>
                <th className="px-4 py-3 text-left text-xs font-semibold text-text-sub dark:text-gray-400 uppercase tracking-wider">
                  Produk
                </th>
                <th className="px-4 py-3 text-left text-xs font-semibold text-text-sub dark:text-gray-400 uppercase tracking-wider">
                  Kategori
                </th>
                <th className="px-4 py-3 text-left text-xs font-semibold text-text-sub dark:text-gray-400 uppercase tracking-wider">
                  Varian
                </th>
                <th className="px-4 py-3 text-left text-xs font-semibold text-text-sub dark:text-gray-400 uppercase tracking-wider">
                  Harga
                </th>
                <th className="px-4 py-3 text-left text-xs font-semibold text-text-sub dark:text-gray-400 uppercase tracking-wider">
                  Status
                </th>
                <th className="px-4 py-3 text-right text-xs font-semibold text-text-sub dark:text-gray-400 uppercase tracking-wider">
                  Aksi
                </th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border-light dark:divide-border-dark">
              {isLoading ? (
                <tr>
                  <td colSpan={7} className="px-4 py-12 text-center">
                    <Icon name="progress_activity" size={32} className="animate-spin text-primary mx-auto" />
                    <p className="text-text-sub dark:text-gray-400 mt-2">Memuat data...</p>
                  </td>
                </tr>
              ) : products.length === 0 ? (
                <tr>
                  <td colSpan={7} className="px-4 py-12 text-center">
                    <Icon name="inventory_2" size={48} className="text-text-sub dark:text-gray-400 mx-auto mb-2" />
                    <p className="text-text-sub dark:text-gray-400">Belum ada produk</p>
                  </td>
                </tr>
              ) : (
                products.map(product => (
                  <ProductRow
                    key={product.id}
                    product={product}
                    isExpanded={expandedProducts.has(product.id)}
                    onToggleExpand={() => toggleExpand(product.id)}
                    onEdit={() => handleEditProduct(product)}
                    onDelete={() => handleDeleteClick(product)}
                  />
                ))
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
                {data.total} produk
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

      {/* Delete Confirm Modal */}
      <DeleteConfirmModal
        isOpen={isDeleteModalOpen}
        onClose={() => {
          setIsDeleteModalOpen(false);
          setProductToDelete(null);
        }}
        onConfirm={handleDeleteConfirm}
        itemName={productToDelete?.name}
        message="Apakah Anda yakin ingin menghapus produk"
        isLoading={deleteProduct.isPending}
      />

      {/* Scan barcode to add / find a product */}
      <BarcodeScannerModal
        isOpen={isScanModalOpen}
        onClose={() => setIsScanModalOpen(false)}
        onScan={handleScanToAdd}
      />
    </div>
  );
}

// Product Row Component with Expandable Variants
function ProductRow({
  product,
  isExpanded,
  onToggleExpand,
  onEdit,
  onDelete,
}: {
  product: Product;
  isExpanded: boolean;
  onToggleExpand: () => void;
  onEdit: () => void;
  onDelete: () => void;
}) {
  const priceRange = getPriceRange(product);
  const variantCount = product.variants?.length || 0;
  const hasMultipleVariants = variantCount > 1;

  return (
    <>
      {/* Main Product Row */}
      <tr className="hover:bg-background-light dark:hover:bg-background-dark transition-colors">
        {/* Expand Button */}
        <td className="px-4 py-4">
          {hasMultipleVariants && (
            <button
              onClick={onToggleExpand}
              className="p-1 hover:bg-gray-200 dark:hover:bg-gray-700 rounded transition-colors"
            >
              <Icon
                name={isExpanded ? 'expand_less' : 'expand_more'}
                size={20}
                className="text-text-sub dark:text-gray-400"
              />
            </button>
          )}
        </td>

        {/* Product Info */}
        <td className="px-4 py-4">
          <div className="flex items-center gap-3">
            {product.image_data ? (
              <img
                src={product.image_data}
                alt={product.name}
                className="w-12 h-12 rounded-lg object-cover border border-border-light dark:border-border-dark"
              />
            ) : (
              <div className="w-12 h-12 rounded-lg bg-gray-100 dark:bg-gray-800 flex items-center justify-center">
                <Icon name="image" size={24} className="text-gray-400" />
              </div>
            )}
            <div>
              <p className="font-medium text-text-main dark:text-white">{product.name}</p>
              {!hasMultipleVariants && product.variants?.[0] && (
                <>
                  <p className="text-xs text-text-sub dark:text-gray-400">
                    SKU: {product.variants[0].sku}
                  </p>
                  {product.variants[0].barcode && (
                    <p className="text-xs text-text-sub dark:text-gray-400">
                      Barcode: {product.variants[0].barcode}
                    </p>
                  )}
                </>
              )}
            </div>
          </div>
        </td>

        {/* Category */}
        <td className="px-4 py-4">
          <span className="text-text-main dark:text-gray-200">
            {product.category_name || '-'}
          </span>
        </td>

        {/* Variant Count */}
        <td className="px-4 py-4">
          {hasMultipleVariants ? (
            <button
              onClick={onToggleExpand}
              className="inline-flex items-center gap-1 text-sm text-primary font-medium hover:underline"
            >
              <Icon name="layers" size={16} />
              {variantCount} varian
            </button>
          ) : (
            <span className="text-xs text-text-sub dark:text-gray-400">1 varian</span>
          )}
        </td>

        {/* Price */}
        <td className="px-4 py-4">
          {priceRange.isSingle ? (
            <span className="font-medium text-text-main dark:text-white">
              {formatCurrency(priceRange.min)}
            </span>
          ) : (
            <div>
              <span className="font-medium text-text-main dark:text-white">
                {formatCurrency(priceRange.min)}
              </span>
              <span className="text-text-sub dark:text-gray-400"> - </span>
              <span className="font-medium text-text-main dark:text-white">
                {formatCurrency(priceRange.max)}
              </span>
            </div>
          )}
        </td>

        {/* Status */}
        <td className="px-4 py-4">
          <Badge variant={product.is_active ? 'success' : 'gray'} dot>
            {product.is_active ? 'Aktif' : 'Nonaktif'}
          </Badge>
        </td>

        {/* Actions */}
        <td className="px-4 py-4 text-right">
          <div className="flex items-center justify-end gap-2">
            <button
              className="p-1.5 hover:bg-background-light dark:hover:bg-gray-700 rounded-lg transition-colors"
              onClick={onEdit}
            >
              <Icon name="edit" size={18} className="text-text-sub dark:text-gray-400" />
            </button>
            <button
              className="p-1.5 hover:bg-background-light dark:hover:bg-gray-700 rounded-lg transition-colors"
              onClick={onDelete}
            >
              <Icon name="delete" size={18} className="text-red-600 dark:text-red-400" />
            </button>
          </div>
        </td>
      </tr>

      {/* Expanded Variants */}
      {isExpanded && hasMultipleVariants && (
        <tr>
          <td colSpan={7} className="px-4 py-0 bg-background-light dark:bg-background-dark">
            <div className="py-4 pl-12">
              <table className="w-full">
                <thead>
                  <tr className="text-xs text-text-sub dark:text-gray-400">
                    <th className="pb-2 text-left font-medium">Nama Varian</th>
                    <th className="pb-2 text-left font-medium">SKU</th>
                    <th className="pb-2 text-left font-medium">Barcode</th>
                    <th className="pb-2 text-left font-medium">Harga</th>
                    <th className="pb-2 text-left font-medium">Modal</th>
                    <th className="pb-2 text-left font-medium">Status</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-border-light dark:divide-border-dark">
                  {product.variants?.map((variant: ProductVariant) => (
                    <tr key={variant.id} className="text-sm">
                      <td className="py-2 text-text-main dark:text-white">
                        {variant.name}
                        {variant.is_default && (
                          <Badge variant="primary" size="sm" className="ml-2">
                            Default
                          </Badge>
                        )}
                      </td>
                      <td className="py-2 text-text-sub dark:text-gray-400 font-mono text-xs">
                        {variant.sku}
                      </td>
                      <td className="py-2 text-text-sub dark:text-gray-400 font-mono text-xs">
                        {variant.barcode || '-'}
                      </td>
                      <td className="py-2 text-text-main dark:text-white font-medium">
                        {formatCurrency(variant.price)}
                      </td>
                      <td className="py-2 text-text-sub dark:text-gray-400">
                        {formatCurrency(variant.standard_cost)}
                      </td>
                      <td className="py-2">
                        <Badge variant={variant.is_active ? 'success' : 'gray'} size="sm">
                          {variant.is_active ? 'Aktif' : 'Nonaktif'}
                        </Badge>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </td>
        </tr>
      )}
    </>
  );
}

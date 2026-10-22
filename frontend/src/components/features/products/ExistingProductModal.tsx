import { Button, Badge, Icon } from '@/components/ui';
import { formatCurrency } from '@/utils';
import type { Product } from '@/types/product';

interface ExistingProductModalProps {
  isOpen: boolean;
  onClose: () => void;
  /** The product that already owns the scanned/entered code. */
  product: Product | null;
  /** Which variant matched (highlighted in the card). */
  matchedVariantId?: number;
  /** Whether the code matched a barcode or an SKU. */
  matchedBy?: 'barcode' | 'sku';
  /** The raw code that was scanned/typed. */
  code?: string;
  /** Open the existing product (e.g. navigate to its edit page). */
  onOpenProduct: () => void;
}

/**
 * Shown when a scanned/entered barcode is already registered to a product.
 * Instead of a transient toast, it surfaces *which* product owns the code so
 * the user can recognise it and jump straight to it.
 */
export function ExistingProductModal({
  isOpen,
  onClose,
  product,
  matchedVariantId,
  matchedBy = 'barcode',
  code,
  onOpenProduct,
}: ExistingProductModalProps) {
  if (!isOpen || !product) return null;

  const matchedVariant =
    product.variants?.find(v => v.id === matchedVariantId) ?? product.variants?.[0];
  const matchLabel = matchedBy === 'barcode' ? 'barcode' : 'SKU';

  return (
    <div className="fixed inset-0 z-[70] flex items-center justify-center p-4 bg-black/50">
      <div className="bg-white dark:bg-card-dark rounded-xl shadow-xl w-full max-w-md">
        {/* Header */}
        <div className="flex items-center gap-3 p-6 border-b border-border-light dark:border-border-dark">
          <div className="flex items-center justify-center w-12 h-12 rounded-full bg-amber-100 dark:bg-amber-900/30">
            <Icon name="inventory_2" size={24} className="text-amber-600 dark:text-amber-400" />
          </div>
          <div className="flex-1">
            <h2 className="text-lg font-bold text-text-main dark:text-white">
              Barcode Sudah Terdaftar
            </h2>
            <p className="text-sm text-text-sub dark:text-gray-400">
              {code ? (
                <>
                  Kode <span className="font-mono font-medium">{code}</span> sudah dipakai produk
                  berikut
                </>
              ) : (
                <>Kode ini sudah dipakai produk berikut</>
              )}
            </p>
          </div>
          <button
            onClick={onClose}
            className="text-text-sub hover:text-text-main dark:hover:text-white transition-colors"
          >
            <Icon name="close" size={24} />
          </button>
        </div>

        {/* Existing product card */}
        <div className="p-6">
          <div className="flex items-start gap-3 p-4 rounded-lg border border-border-light dark:border-border-dark bg-background-light dark:bg-background-dark">
            {(product.image_url ?? product.image_data) ? (
              <img
                src={(product.image_url ?? product.image_data) || undefined}
                alt={product.name}
                className="w-16 h-16 rounded-lg object-cover border border-border-light dark:border-border-dark shrink-0"
              />
            ) : (
              <div className="w-16 h-16 rounded-lg bg-gray-100 dark:bg-gray-800 flex items-center justify-center shrink-0">
                <Icon name="image" size={28} className="text-gray-400" />
              </div>
            )}

            <div className="flex-1 min-w-0">
              <div className="flex items-center gap-2">
                <p className="font-semibold text-text-main dark:text-white truncate">
                  {product.name}
                </p>
                <Badge variant={product.is_active ? 'success' : 'gray'} size="sm">
                  {product.is_active ? 'Aktif' : 'Nonaktif'}
                </Badge>
              </div>

              {product.category_name && (
                <p className="text-xs text-text-sub dark:text-gray-400 mt-0.5">
                  {product.category_name}
                </p>
              )}

              {matchedVariant && (
                <div className="mt-2 space-y-0.5 text-xs">
                  <div className="flex items-center gap-1.5">
                    <span className="text-text-sub dark:text-gray-400">Varian:</span>
                    <span className="text-text-main dark:text-gray-200 font-medium">
                      {matchedVariant.name}
                    </span>
                    {matchedVariant.is_default && (
                      <Badge variant="primary" size="sm">Default</Badge>
                    )}
                  </div>
                  <p className="text-text-sub dark:text-gray-400 font-mono">
                    SKU: {matchedVariant.sku}
                    {matchedBy === 'sku' && (
                      <span className="ml-1 text-amber-600 dark:text-amber-400">(cocok)</span>
                    )}
                  </p>
                  <p className="text-text-sub dark:text-gray-400 font-mono">
                    Barcode: {matchedVariant.barcode || '-'}
                    {matchedBy === 'barcode' && (
                      <span className="ml-1 text-amber-600 dark:text-amber-400">(cocok)</span>
                    )}
                  </p>
                  <p className="text-text-main dark:text-white font-semibold pt-0.5">
                    {formatCurrency(matchedVariant.price)}
                  </p>
                </div>
              )}
            </div>
          </div>

          <p className="text-xs text-text-sub dark:text-gray-400 mt-3">
            Cocok berdasarkan {matchLabel}. Setiap {matchLabel} hanya boleh dipakai satu produk.
          </p>
        </div>

        {/* Footer */}
        <div className="flex items-center justify-end gap-3 p-6 border-t border-border-light dark:border-border-dark">
          <Button variant="outline" onClick={onClose}>
            Tutup
          </Button>
          <Button onClick={onOpenProduct} leftIcon={<Icon name="open_in_new" size={18} />}>
            Buka Produk
          </Button>
        </div>
      </div>
    </div>
  );
}

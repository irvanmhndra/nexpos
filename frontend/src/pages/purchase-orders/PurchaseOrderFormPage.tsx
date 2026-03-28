import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import toast from 'react-hot-toast';
import { Card, Button, Input, Icon } from '@/components/ui';
import { useSuppliers } from '@/hooks/useSuppliers';
import { useCreatePurchaseOrder } from '@/hooks/usePurchaseOrders';
import { useBranch } from '@/contexts/BranchContext';
import { formatCurrency, getApiErrorMessage } from '@/utils';
import { productsService } from '@/services/productsService';
import type { CreatePOItemInput } from '@/types/purchaseOrder';

interface POItemRow extends CreatePOItemInput {
  _key: number;
}

let _keyCounter = 0;

function newRow(): POItemRow {
  return {
    _key: ++_keyCounter,
    product_variant_id: undefined,
    sku: '',
    variant_name: '',
    quantity: 1,
    unit_cost: 0,
  };
}

export function PurchaseOrderFormPage() {
  const navigate = useNavigate();
  const { activeBranch } = useBranch();

  const [supplierId, setSupplierId] = useState<string>('');
  const [notes, setNotes] = useState('');
  const [items, setItems] = useState<POItemRow[]>([newRow()]);
  const [skuSearching, setSkuSearching] = useState<Record<number, boolean>>({});

  const { data: suppliersData } = useSuppliers({ per_page: 200, is_active: true });
  const createPO = useCreatePurchaseOrder();

  const suppliers = suppliersData?.data || [];

  const handleSkuLookup = async (key: number, sku: string) => {
    if (!sku.trim()) return;
    setSkuSearching(prev => ({ ...prev, [key]: true }));
    try {
      const res = await productsService.getProducts({ search: sku, per_page: 10, is_active: true });
      let found: { variantId: number; variantName: string } | null = null;
      for (const p of res.data) {
        const v = p.variants?.find(v => v.sku === sku);
        if (v) {
          found = { variantId: v.id, variantName: `${p.name} - ${v.name}` };
          break;
        }
      }
      if (found) {
        setItems(prev =>
          prev.map(item =>
            item._key === key
              ? { ...item, product_variant_id: found!.variantId, variant_name: found!.variantName }
              : item
          )
        );
      } else {
        toast.error(`SKU "${sku}" tidak ditemukan dalam produk`);
      }
    } catch {
      toast.error('Gagal mencari produk');
    } finally {
      setSkuSearching(prev => ({ ...prev, [key]: false }));
    }
  };

  const updateItem = (key: number, field: keyof POItemRow, value: string | number) => {
    setItems(prev =>
      prev.map(item => (item._key === key ? { ...item, [field]: value } : item))
    );
  };

  const removeItem = (key: number) => {
    setItems(prev => prev.filter(item => item._key !== key));
  };

  const addItem = () => {
    setItems(prev => [...prev, newRow()]);
  };

  const totalAmount = items.reduce((sum, item) => sum + item.quantity * item.unit_cost, 0);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!supplierId) {
      toast.error('Pilih supplier terlebih dahulu');
      return;
    }

    if (!activeBranch) {
      toast.error('Pilih cabang terlebih dahulu');
      return;
    }

    const validItems = items.filter(item => item.sku.trim() && item.variant_name.trim());
    if (validItems.length === 0) {
      toast.error('Tambahkan minimal satu item');
      return;
    }

    try {
      const po = await createPO.mutateAsync({
        supplier_id: Number(supplierId),
        branch_id: activeBranch.id,
        notes: notes || undefined,
        items: validItems.map(({ _key, ...rest }) => rest),
      });
      toast.success('Purchase Order berhasil dibuat');
      navigate(`/purchase-orders/${po.id}`);
    } catch (err) {
      toast.error(getApiErrorMessage(err));
    }
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center gap-4">
        <button
          onClick={() => navigate('/purchase-orders')}
          className="p-2 hover:bg-background-light dark:hover:bg-background-dark rounded-lg transition-colors"
        >
          <Icon name="arrow_back" size={20} className="text-text-sub" />
        </button>
        <div>
          <h1 className="text-2xl md:text-3xl font-bold text-text-main dark:text-white">
            Buat Purchase Order
          </h1>
          <p className="text-text-sub dark:text-gray-400 text-sm mt-1">
            Isi detail pembelian barang dari supplier
          </p>
        </div>
      </div>

      <form onSubmit={handleSubmit} className="space-y-6">
        {/* PO Info */}
        <Card variant="bordered" padding="md">
          <h2 className="text-lg font-semibold text-text-main dark:text-white mb-4">
            Informasi PO
          </h2>
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-text-main dark:text-white mb-1">
                Supplier <span className="text-red-500">*</span>
              </label>
              <select
                value={supplierId}
                onChange={e => setSupplierId(e.target.value)}
                required
                className="w-full rounded-lg border border-border-light dark:border-border-dark bg-white dark:bg-background-dark px-3 py-2 text-sm text-text-main dark:text-white focus:outline-none focus:ring-2 focus:ring-primary"
              >
                <option value="">-- Pilih Supplier --</option>
                {suppliers.map(s => (
                  <option key={s.id} value={s.id}>
                    {s.name} ({s.code})
                  </option>
                ))}
              </select>
            </div>

            <div>
              <label className="block text-sm font-medium text-text-main dark:text-white mb-1">
                Cabang
              </label>
              <div className="w-full rounded-lg border border-border-light dark:border-border-dark bg-gray-50 dark:bg-background-dark px-3 py-2 text-sm text-text-main dark:text-white">
                {activeBranch?.name || 'Belum ada cabang aktif'}
              </div>
            </div>

            <div className="sm:col-span-2">
              <label className="block text-sm font-medium text-text-main dark:text-white mb-1">
                Catatan
              </label>
              <textarea
                className="w-full rounded-lg border border-border-light dark:border-border-dark bg-white dark:bg-background-dark px-3 py-2 text-sm text-text-main dark:text-white placeholder:text-text-sub focus:outline-none focus:ring-2 focus:ring-primary"
                rows={3}
                placeholder="Catatan PO (opsional)"
                value={notes}
                onChange={e => setNotes(e.target.value)}
              />
            </div>
          </div>
        </Card>

        {/* Items */}
        <Card variant="bordered" padding="none">
          <div className="flex items-center justify-between p-4 border-b border-border-light dark:border-border-dark">
            <h2 className="text-lg font-semibold text-text-main dark:text-white">Item Pembelian</h2>
            <Button type="button" variant="outline" size="sm" leftIcon={<Icon name="add" size={18} />} onClick={addItem}>
              Tambah Item
            </Button>
          </div>

          <div className="overflow-x-auto">
            <table className="w-full">
              <thead>
                <tr className="bg-background-light dark:bg-background-dark border-b border-border-light dark:border-border-dark">
                  <th className="px-4 py-3 text-left text-xs font-semibold text-text-sub dark:text-gray-400 uppercase tracking-wider">
                    SKU
                  </th>
                  <th className="px-4 py-3 text-left text-xs font-semibold text-text-sub dark:text-gray-400 uppercase tracking-wider">
                    Nama Produk
                  </th>
                  <th className="px-4 py-3 text-left text-xs font-semibold text-text-sub dark:text-gray-400 uppercase tracking-wider">
                    Qty
                  </th>
                  <th className="px-4 py-3 text-left text-xs font-semibold text-text-sub dark:text-gray-400 uppercase tracking-wider">
                    Harga Beli
                  </th>
                  <th className="px-4 py-3 text-left text-xs font-semibold text-text-sub dark:text-gray-400 uppercase tracking-wider">
                    Subtotal
                  </th>
                  <th className="px-4 py-3"></th>
                </tr>
              </thead>
              <tbody>
                {items.map(item => (
                  <tr
                    key={item._key}
                    className="border-b border-border-light dark:border-border-dark last:border-0"
                  >
                    <td className="px-4 py-3">
                      <div className="flex items-center gap-2">
                        <Input
                          placeholder="SKU..."
                          value={item.sku}
                          onChange={e => updateItem(item._key, 'sku', e.target.value)}
                          fullWidth
                        />
                        <button
                          type="button"
                          onClick={() => handleSkuLookup(item._key, item.sku)}
                          disabled={skuSearching[item._key]}
                          className="p-1.5 hover:bg-background-light dark:hover:bg-background-dark rounded-lg transition-colors flex-shrink-0"
                          title="Cari produk"
                        >
                          {skuSearching[item._key] ? (
                            <Icon name="progress_activity" size={18} className="animate-spin text-primary" />
                          ) : (
                            <Icon name="search" size={18} className="text-primary" />
                          )}
                        </button>
                      </div>
                    </td>
                    <td className="px-4 py-3">
                      <Input
                        placeholder="Nama produk/varian..."
                        value={item.variant_name}
                        onChange={e => updateItem(item._key, 'variant_name', e.target.value)}
                        fullWidth
                      />
                    </td>
                    <td className="px-4 py-3 w-24">
                      <Input
                        type="number"
                        min="1"
                        value={String(item.quantity)}
                        onChange={e => updateItem(item._key, 'quantity', parseInt(e.target.value) || 1)}
                        fullWidth
                      />
                    </td>
                    <td className="px-4 py-3 w-40">
                      <Input
                        type="number"
                        min="0"
                        placeholder="0"
                        value={String(item.unit_cost)}
                        onChange={e => updateItem(item._key, 'unit_cost', parseFloat(e.target.value) || 0)}
                        fullWidth
                      />
                    </td>
                    <td className="px-4 py-3 w-36">
                      <span className="text-sm font-medium text-text-main dark:text-gray-200">
                        {formatCurrency(item.quantity * item.unit_cost)}
                      </span>
                    </td>
                    <td className="px-4 py-3">
                      <button
                        type="button"
                        onClick={() => removeItem(item._key)}
                        disabled={items.length === 1}
                        className="p-1.5 hover:bg-red-50 dark:hover:bg-red-900/20 rounded-lg transition-colors disabled:opacity-40"
                      >
                        <Icon name="delete" size={18} className="text-red-500" />
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

          <div className="px-4 py-4 border-t border-border-light dark:border-border-dark flex items-center justify-between">
            <span className="text-sm font-semibold text-text-sub dark:text-gray-400 uppercase tracking-wider">
              Total
            </span>
            <span className="text-xl font-bold text-text-main dark:text-white">
              {formatCurrency(totalAmount)}
            </span>
          </div>
        </Card>

        {/* Actions */}
        <div className="flex items-center justify-end gap-3">
          <Button
            type="button"
            variant="outline"
            onClick={() => navigate('/purchase-orders')}
          >
            Batal
          </Button>
          <Button type="submit" disabled={createPO.isPending}>
            {createPO.isPending ? 'Menyimpan...' : 'Buat Purchase Order'}
          </Button>
        </div>
      </form>
    </div>
  );
}

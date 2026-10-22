import { useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import toast from 'react-hot-toast';
import { Card, Badge, Button, Input, Icon } from '@/components/ui';
import { usePurchaseOrder, useReceivePurchaseOrder } from '@/hooks/usePurchaseOrders';
import { formatCurrency, formatDate, getApiErrorMessage } from '@/utils';
import type { POStatus, PurchaseOrderItem, ReceivePOItemInput } from '@/types/purchaseOrder';

const STATUS_BADGE_MAP: Record<POStatus, 'gray' | 'warning' | 'info' | 'success' | 'danger'> = {
  draft: 'gray',
  ordered: 'warning',
  partial: 'info',
  received: 'success',
  cancelled: 'danger',
};

const STATUS_LABELS: Record<POStatus, string> = {
  draft: 'Draft',
  ordered: 'Dipesan',
  partial: 'Sebagian Diterima',
  received: 'Diterima',
  cancelled: 'Dibatalkan',
};

function ReceiveModal({
  isOpen,
  onClose,
  onSubmit,
  isLoading,
  items,
}: {
  isOpen: boolean;
  onClose: () => void;
  onSubmit: (data: { items: ReceivePOItemInput[]; notes?: string }) => void;
  isLoading: boolean;
  items: PurchaseOrderItem[];
}) {
  const [receivedQtys, setReceivedQtys] = useState<Record<number, string>>(
    Object.fromEntries(items.map(item => [item.id, String(item.quantity - item.received_quantity)]))
  );
  const [notes, setNotes] = useState('');

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    const receiveItems: ReceivePOItemInput[] = items.map(item => ({
      purchase_order_item_id: item.id,
      received_quantity: parseInt(receivedQtys[item.id] || '0') || 0,
    }));
    onSubmit({ items: receiveItems, notes: notes || undefined });
  };

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 z-50 overflow-y-auto">
      <div className="fixed inset-0 bg-black/50 transition-opacity" onClick={onClose} />
      <div className="flex min-h-full items-center justify-center p-4">
        <div className="relative w-full max-w-2xl bg-white dark:bg-card-dark rounded-2xl shadow-xl">
          <div className="flex items-center justify-between p-6 border-b border-border-light dark:border-border-dark">
            <div>
              <h2 className="text-xl font-bold text-text-main dark:text-white">Terima Barang</h2>
              <p className="text-sm text-text-sub dark:text-gray-400 mt-1">
                Masukkan jumlah barang yang diterima
              </p>
            </div>
            <button
              onClick={onClose}
              className="p-2 hover:bg-gray-100 dark:hover:bg-gray-800 rounded-lg transition-colors"
            >
              <Icon name="close" size={20} className="text-text-sub" />
            </button>
          </div>

          <form onSubmit={handleSubmit}>
            <div className="p-6 space-y-4">
              <div className="overflow-x-auto">
                <table className="w-full">
                  <thead>
                    <tr className="bg-background-light dark:bg-background-dark rounded-lg">
                      <th className="px-3 py-2 text-left text-xs font-semibold text-text-sub dark:text-gray-400 uppercase">
                        Item
                      </th>
                      <th className="px-3 py-2 text-left text-xs font-semibold text-text-sub dark:text-gray-400 uppercase">
                        Dipesan
                      </th>
                      <th className="px-3 py-2 text-left text-xs font-semibold text-text-sub dark:text-gray-400 uppercase">
                        Sudah Diterima
                      </th>
                      <th className="px-3 py-2 text-left text-xs font-semibold text-text-sub dark:text-gray-400 uppercase">
                        Diterima Kali Ini
                      </th>
                    </tr>
                  </thead>
                  <tbody>
                    {items.map(item => {
                      const remaining = item.quantity - item.received_quantity;
                      return (
                        <tr
                          key={item.id}
                          className="border-b border-border-light dark:border-border-dark last:border-0"
                        >
                          <td className="px-3 py-3">
                            <p className="text-sm text-text-main dark:text-white">{item.variant_name}</p>
                            <p className="text-xs text-text-sub dark:text-gray-400 font-mono">{item.sku}</p>
                          </td>
                          <td className="px-3 py-3 text-sm text-text-main dark:text-gray-200">
                            {item.quantity}
                          </td>
                          <td className="px-3 py-3 text-sm text-text-main dark:text-gray-200">
                            {item.received_quantity}
                          </td>
                          <td className="px-3 py-3 w-28">
                            <Input
                              type="number"
                              min="0"
                              max={remaining}
                              value={receivedQtys[item.id] || '0'}
                              onChange={e =>
                                setReceivedQtys(prev => ({ ...prev, [item.id]: e.target.value }))
                              }
                              fullWidth
                              disabled={remaining <= 0}
                            />
                          </td>
                        </tr>
                      );
                    })}
                  </tbody>
                </table>
              </div>

              <div>
                <label className="block text-sm font-medium text-text-main dark:text-white mb-1">
                  Catatan
                </label>
                <textarea
                  className="w-full rounded-lg border border-border-light dark:border-border-dark bg-white dark:bg-background-dark px-3 py-2 text-sm text-text-main dark:text-white placeholder:text-text-sub focus:outline-none focus:ring-2 focus:ring-primary"
                  rows={3}
                  placeholder="Catatan penerimaan (opsional)"
                  value={notes}
                  onChange={e => setNotes(e.target.value)}
                />
              </div>
            </div>

            <div className="flex items-center justify-end gap-3 p-6 border-t border-border-light dark:border-border-dark">
              <Button type="button" variant="outline" onClick={onClose}>
                Batal
              </Button>
              <Button type="submit" disabled={isLoading}>
                {isLoading ? 'Menyimpan...' : 'Konfirmasi Penerimaan'}
              </Button>
            </div>
          </form>
        </div>
      </div>
    </div>
  );
}

export function PurchaseOrderDetailPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [isReceiveOpen, setIsReceiveOpen] = useState(false);

  const { data: po, isLoading } = usePurchaseOrder(Number(id));
  const receivePO = useReceivePurchaseOrder();

  const handleReceive = async (data: { items: ReceivePOItemInput[]; notes?: string }) => {
    if (!po) return;
    try {
      await receivePO.mutateAsync({ id: po.id, request: data });
      toast.success('Barang berhasil diterima');
      setIsReceiveOpen(false);
    } catch (err) {
      toast.error(getApiErrorMessage(err));
    }
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-20">
        <Icon name="progress_activity" size={40} className="animate-spin text-primary" />
      </div>
    );
  }

  if (!po) {
    return (
      <div className="text-center py-20">
        <p className="text-text-sub dark:text-gray-400">Purchase Order tidak ditemukan</p>
        <Button className="mt-4" onClick={() => navigate('/purchase-orders')}>
          Kembali
        </Button>
      </div>
    );
  }

  const canReceive = po.status === 'ordered' || po.status === 'partial';

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div className="flex items-center gap-4">
          <button
            onClick={() => navigate('/purchase-orders')}
            className="p-2 hover:bg-background-light dark:hover:bg-background-dark rounded-lg transition-colors"
          >
            <Icon name="arrow_back" size={20} className="text-text-sub" />
          </button>
          <div>
            <div className="flex items-center gap-3">
              <h1 className="text-2xl md:text-3xl font-bold text-text-main dark:text-white">
                {po.po_number}
              </h1>
              <Badge variant={STATUS_BADGE_MAP[po.status]} dot>
                {STATUS_LABELS[po.status]}
              </Badge>
            </div>
            <p className="text-text-sub dark:text-gray-400 text-sm mt-1">
              Dibuat {formatDate(new Date(po.created_at), 'dd MMM yyyy HH:mm')}
            </p>
          </div>
        </div>
        {canReceive && (
          <Button
            leftIcon={<Icon name="inventory" size={20} />}
            onClick={() => setIsReceiveOpen(true)}
          >
            Terima Barang
          </Button>
        )}
      </div>

      {/* Info */}
      <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
        <Card variant="bordered" padding="md">
          <h2 className="text-sm font-semibold text-text-sub dark:text-gray-400 uppercase tracking-wider mb-3">
            Informasi Supplier
          </h2>
          <div className="space-y-2">
            <div>
              <p className="text-xs text-text-sub dark:text-gray-400">Nama</p>
              <p className="text-sm font-medium text-text-main dark:text-white">
                {po.supplier?.name || '-'}
              </p>
            </div>
            <div>
              <p className="text-xs text-text-sub dark:text-gray-400">Kontak</p>
              <p className="text-sm text-text-main dark:text-gray-200">
                {po.supplier?.contact_name || '-'}
              </p>
            </div>
            <div>
              <p className="text-xs text-text-sub dark:text-gray-400">Telepon</p>
              <p className="text-sm text-text-main dark:text-gray-200">
                {po.supplier?.phone || '-'}
              </p>
            </div>
          </div>
        </Card>

        <Card variant="bordered" padding="md">
          <h2 className="text-sm font-semibold text-text-sub dark:text-gray-400 uppercase tracking-wider mb-3">
            Informasi PO
          </h2>
          <div className="space-y-2">
            <div>
              <p className="text-xs text-text-sub dark:text-gray-400">Total Pembelian</p>
              <p className="text-xl font-bold text-text-main dark:text-white">
                {formatCurrency(po.total_amount)}
              </p>
            </div>
            {po.ordered_at && (
              <div>
                <p className="text-xs text-text-sub dark:text-gray-400">Tanggal Pemesanan</p>
                <p className="text-sm text-text-main dark:text-gray-200">
                  {formatDate(new Date(po.ordered_at), 'dd MMM yyyy')}
                </p>
              </div>
            )}
            {po.notes && (
              <div>
                <p className="text-xs text-text-sub dark:text-gray-400">Catatan</p>
                <p className="text-sm text-text-main dark:text-gray-200">{po.notes}</p>
              </div>
            )}
          </div>
        </Card>
      </div>

      {/* Items */}
      <Card variant="bordered" padding="none">
        <div className="p-4 border-b border-border-light dark:border-border-dark">
          <h2 className="text-lg font-semibold text-text-main dark:text-white">Item Pembelian</h2>
        </div>
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead>
              <tr className="bg-background-light dark:bg-background-dark border-b border-border-light dark:border-border-dark">
                <th className="px-4 py-3 text-left text-xs font-semibold text-text-sub dark:text-gray-400 uppercase tracking-wider">
                  Produk
                </th>
                <th className="px-4 py-3 text-left text-xs font-semibold text-text-sub dark:text-gray-400 uppercase tracking-wider">
                  SKU
                </th>
                <th className="px-4 py-3 text-left text-xs font-semibold text-text-sub dark:text-gray-400 uppercase tracking-wider">
                  Dipesan
                </th>
                <th className="px-4 py-3 text-left text-xs font-semibold text-text-sub dark:text-gray-400 uppercase tracking-wider">
                  Diterima
                </th>
                <th className="px-4 py-3 text-left text-xs font-semibold text-text-sub dark:text-gray-400 uppercase tracking-wider">
                  Harga Satuan
                </th>
                <th className="px-4 py-3 text-left text-xs font-semibold text-text-sub dark:text-gray-400 uppercase tracking-wider">
                  Subtotal
                </th>
              </tr>
            </thead>
            <tbody>
              {(po.items || []).map((item: PurchaseOrderItem) => {
                const remaining = item.quantity - item.received_quantity;
                return (
                  <tr
                    key={item.id}
                    className="border-b border-border-light dark:border-border-dark last:border-0 hover:bg-background-light dark:hover:bg-background-dark transition-colors"
                  >
                    <td className="px-4 py-4">
                      <p className="text-sm font-medium text-text-main dark:text-white">
                        {item.variant_name}
                      </p>
                      {remaining > 0 && (
                        <p className="text-xs text-yellow-600 dark:text-yellow-400 mt-0.5">
                          Sisa {remaining} belum diterima
                        </p>
                      )}
                    </td>
                    <td className="px-4 py-4 font-mono text-xs text-text-sub dark:text-gray-400">
                      {item.sku}
                    </td>
                    <td className="px-4 py-4 text-sm text-text-main dark:text-gray-200">
                      {item.quantity}
                    </td>
                    <td className="px-4 py-4 text-sm text-text-main dark:text-gray-200">
                      {item.received_quantity}
                    </td>
                    <td className="px-4 py-4 text-sm text-text-main dark:text-gray-200">
                      {formatCurrency(item.unit_cost)}
                    </td>
                    <td className="px-4 py-4 text-sm font-medium text-text-main dark:text-gray-200">
                      {formatCurrency(item.subtotal)}
                    </td>
                  </tr>
                );
              })}
            </tbody>
            <tfoot>
              <tr className="bg-background-light dark:bg-background-dark">
                <td colSpan={5} className="px-4 py-4 text-right font-semibold text-text-sub dark:text-gray-400">
                  Total
                </td>
                <td className="px-4 py-4 font-bold text-text-main dark:text-white">
                  {formatCurrency(po.total_amount)}
                </td>
              </tr>
            </tfoot>
          </table>
        </div>
      </Card>

      {/* Receive Modal */}
      {po.items && po.items.length > 0 && (
        <ReceiveModal
          isOpen={isReceiveOpen}
          onClose={() => setIsReceiveOpen(false)}
          onSubmit={handleReceive}
          isLoading={receivePO.isPending}
          items={po.items}
        />
      )}
    </div>
  );
}

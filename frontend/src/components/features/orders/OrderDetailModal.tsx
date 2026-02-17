import { useState } from 'react';
import { Button, Icon, Input } from '@/components/ui';
import { formatCurrency, formatDate } from '@/utils';
import type { Order, PaymentMethod } from '@/types/order';
import {
  useConfirmOrder,
  useAddPayment,
  useCompleteOrder,
  useCancelOrder,
  useVoidOrder,
} from '@/hooks/useOrders';

interface OrderDetailModalProps {
  isOpen: boolean;
  onClose: () => void;
  order: Order | null;
}

type ActionMode = 'view' | 'add-payment' | 'cancel' | 'void';

export function OrderDetailModal({ isOpen, onClose, order }: OrderDetailModalProps) {
  const [actionMode, setActionMode] = useState<ActionMode>('view');
  const [paymentAmount, setPaymentAmount] = useState('');
  const [paymentMethod, setPaymentMethod] = useState<PaymentMethod>('cash');
  const [paymentRef, setPaymentRef] = useState('');
  const [cancelReason, setCancelReason] = useState('');
  const [voidReason, setVoidReason] = useState('');

  const confirmOrder = useConfirmOrder();
  const addPayment = useAddPayment();
  const completeOrder = useCompleteOrder();
  const cancelOrder = useCancelOrder();
  const voidOrder = useVoidOrder();

  const isLoading =
    confirmOrder.isPending ||
    addPayment.isPending ||
    completeOrder.isPending ||
    cancelOrder.isPending ||
    voidOrder.isPending;

  const handleClose = () => {
    setActionMode('view');
    setPaymentAmount('');
    setPaymentMethod('cash');
    setPaymentRef('');
    setCancelReason('');
    setVoidReason('');
    onClose();
  };

  const handleConfirm = async () => {
    if (!order) return;
    try {
      await confirmOrder.mutateAsync(order.id);
      handleClose();
    } catch (error) {
      console.error('Error confirming order:', error);
    }
  };

  const handleAddPayment = async () => {
    if (!order) return;
    const amount = parseFloat(paymentAmount);
    if (isNaN(amount) || amount <= 0) return;

    try {
      await addPayment.mutateAsync({
        id: order.id,
        data: {
          payments: [
            {
              method: paymentMethod,
              amount,
              reference_no: paymentRef || undefined,
            },
          ],
        },
      });
      handleClose();
    } catch (error) {
      console.error('Error adding payment:', error);
    }
  };

  const handleComplete = async () => {
    if (!order) return;
    try {
      await completeOrder.mutateAsync({ id: order.id });
      handleClose();
    } catch (error) {
      console.error('Error completing order:', error);
    }
  };

  const handleCancel = async () => {
    if (!order || !cancelReason.trim()) return;

    try {
      await cancelOrder.mutateAsync({
        id: order.id,
        data: { reason: cancelReason },
      });
      handleClose();
    } catch (error) {
      console.error('Error cancelling order:', error);
    }
  };

  const handleVoid = async () => {
    if (!order || !voidReason.trim()) return;

    try {
      await voidOrder.mutateAsync({
        id: order.id,
        data: { reason: voidReason },
      });
      handleClose();
    } catch (error) {
      console.error('Error voiding order:', error);
    }
  };

  const getStatusBadge = (status: string) => {
    const badges: Record<string, string> = {
      draft: 'bg-gray-100 text-gray-700 dark:bg-gray-900/30 dark:text-gray-400',
      confirmed: 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400',
      completed: 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400',
      cancelled: 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400',
      voided: 'bg-orange-100 text-orange-700 dark:bg-orange-900/30 dark:text-orange-400',
    };
    return badges[status] || 'bg-gray-100 text-gray-700';
  };

  const getStatusLabel = (status: string) => {
    const labels: Record<string, string> = {
      draft: 'Draft',
      confirmed: 'Dikonfirmasi',
      completed: 'Selesai',
      cancelled: 'Dibatalkan',
      voided: 'Void',
    };
    return labels[status] || status;
  };

  const getPaymentStatusBadge = (status: string) => {
    const badges: Record<string, string> = {
      unpaid: 'bg-gray-100 text-gray-700 dark:bg-gray-900/30 dark:text-gray-400',
      partial: 'bg-yellow-100 text-yellow-700 dark:bg-yellow-900/30 dark:text-yellow-400',
      paid: 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400',
      refunded: 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400',
    };
    return badges[status] || 'bg-gray-100 text-gray-700';
  };

  const getPaymentStatusLabel = (status: string) => {
    const labels: Record<string, string> = {
      unpaid: 'Belum Bayar',
      partial: 'Sebagian',
      paid: 'Lunas',
      refunded: 'Dikembalikan',
    };
    return labels[status] || status;
  };

  // Determine available actions based on order state
  const canConfirm = order?.status === 'draft';
  const canAddPayment = (order?.status === 'draft' || order?.status === 'confirmed') && order?.payment_status !== 'paid';
  const canComplete = order?.status === 'confirmed' && order?.payment_status === 'paid';
  const canCancel = (order?.status === 'draft' || order?.status === 'confirmed') && order?.payment_status !== 'paid';
  const canVoid = (order?.status === 'confirmed' || order?.status === 'completed') && order?.payment_status === 'paid';

  const renderActionContent = () => {
    switch (actionMode) {
      case 'add-payment':
        return (
          <div className="space-y-4 p-4 bg-background-light dark:bg-background-dark rounded-lg">
            <h4 className="font-medium text-text-main dark:text-white">Tambah Pembayaran</h4>
            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-medium text-text-sub dark:text-gray-400 mb-1">
                  Jumlah
                </label>
                <Input
                  type="number"
                  value={paymentAmount}
                  onChange={(e) => setPaymentAmount(e.target.value)}
                  placeholder={`Sisa: ${formatCurrency(order?.balance_due ?? 0)}`}
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-text-sub dark:text-gray-400 mb-1">
                  Metode
                </label>
                <select
                  value={paymentMethod}
                  onChange={(e) => setPaymentMethod(e.target.value as PaymentMethod)}
                  className="w-full h-12 px-4 rounded-lg border border-border-light dark:border-border-dark bg-white dark:bg-card-dark text-text-main dark:text-white"
                >
                  <option value="cash">Cash</option>
                  <option value="debit_card">Debit Card</option>
                  <option value="credit_card">Credit Card</option>
                  <option value="e_wallet">E-Wallet</option>
                  <option value="bank_transfer">Transfer Bank</option>
                  <option value="qris">QRIS</option>
                </select>
              </div>
            </div>
            {paymentMethod !== 'cash' && (
              <div>
                <label className="block text-sm font-medium text-text-sub dark:text-gray-400 mb-1">
                  No. Referensi (Opsional)
                </label>
                <Input
                  value={paymentRef}
                  onChange={(e) => setPaymentRef(e.target.value)}
                  placeholder="No. transaksi/referensi"
                />
              </div>
            )}
            <div className="flex gap-2">
              <Button variant="outline" onClick={() => setActionMode('view')} disabled={isLoading}>
                Batal
              </Button>
              <Button onClick={handleAddPayment} disabled={isLoading || !paymentAmount}>
                {isLoading ? 'Memproses...' : 'Bayar'}
              </Button>
            </div>
          </div>
        );

      case 'cancel':
        return (
          <div className="space-y-4 p-4 bg-red-50 dark:bg-red-900/20 rounded-lg">
            <h4 className="font-medium text-red-700 dark:text-red-400">Batalkan Pesanan</h4>
            <div>
              <label className="block text-sm font-medium text-text-sub dark:text-gray-400 mb-1">
                Alasan Pembatalan
              </label>
              <textarea
                value={cancelReason}
                onChange={(e) => setCancelReason(e.target.value)}
                className="w-full h-24 px-4 py-3 rounded-lg border border-border-light dark:border-border-dark bg-white dark:bg-card-dark text-text-main dark:text-white resize-none"
                placeholder="Masukkan alasan pembatalan..."
              />
            </div>
            <div className="flex gap-2">
              <Button variant="outline" onClick={() => setActionMode('view')} disabled={isLoading}>
                Kembali
              </Button>
              <Button variant="danger" onClick={handleCancel} disabled={isLoading || !cancelReason.trim()}>
                {isLoading ? 'Memproses...' : 'Batalkan Pesanan'}
              </Button>
            </div>
          </div>
        );

      case 'void':
        return (
          <div className="space-y-4 p-4 bg-orange-50 dark:bg-orange-900/20 rounded-lg">
            <h4 className="font-medium text-orange-700 dark:text-orange-400">Void Pesanan</h4>
            <p className="text-sm text-text-sub dark:text-gray-400">
              Void akan membatalkan pesanan yang sudah dibayar. Pembayaran perlu direfund secara manual.
            </p>
            <div>
              <label className="block text-sm font-medium text-text-sub dark:text-gray-400 mb-1">
                Alasan Void
              </label>
              <textarea
                value={voidReason}
                onChange={(e) => setVoidReason(e.target.value)}
                className="w-full h-24 px-4 py-3 rounded-lg border border-border-light dark:border-border-dark bg-white dark:bg-card-dark text-text-main dark:text-white resize-none"
                placeholder="Masukkan alasan void..."
              />
            </div>
            <div className="flex gap-2">
              <Button variant="outline" onClick={() => setActionMode('view')} disabled={isLoading}>
                Kembali
              </Button>
              <Button variant="danger" onClick={handleVoid} disabled={isLoading || !voidReason.trim()}>
                {isLoading ? 'Memproses...' : 'Void Pesanan'}
              </Button>
            </div>
          </div>
        );

      default:
        return null;
    }
  };

  if (!isOpen || !order) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/50">
      <div className="bg-white dark:bg-card-dark rounded-xl shadow-xl w-full max-w-3xl max-h-[90vh] overflow-hidden flex flex-col">
        {/* Header */}
        <div className="flex items-center justify-between p-6 border-b border-border-light dark:border-border-dark">
          <h2 className="text-xl font-bold text-text-main dark:text-white">
            Pesanan {order.order_no}
          </h2>
          <button
            onClick={handleClose}
            className="text-text-sub hover:text-text-main dark:hover:text-white transition-colors"
          >
            <Icon name="close" size={24} />
          </button>
        </div>

        {/* Content */}
        <div className="flex-1 overflow-y-auto p-6">
          <div className="space-y-6">
            {/* Status Row */}
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            <span className={`inline-flex items-center px-3 py-1.5 rounded-full text-sm font-medium ${getStatusBadge(order.status)}`}>
              {getStatusLabel(order.status)}
            </span>
            <span className={`inline-flex items-center px-3 py-1.5 rounded-full text-sm font-medium ${getPaymentStatusBadge(order.payment_status)}`}>
              {getPaymentStatusLabel(order.payment_status)}
            </span>
          </div>
          <div className="text-sm text-text-sub dark:text-gray-400">
            {formatDate(order.created_at)}
          </div>
        </div>

        {/* Customer Info */}
        <div className="p-4 bg-background-light dark:bg-background-dark rounded-lg">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 rounded-full bg-primary/10 flex items-center justify-center">
              <Icon name="person" size={20} className="text-primary" />
            </div>
            <div>
              <p className="font-medium text-text-main dark:text-white">
                {order.customer_name || 'Walk-in Customer'}
              </p>
              <p className="text-sm text-text-sub dark:text-gray-400">
                {order.fulfillment_type === 'delivery' ? 'Delivery' : 'Counter'}
              </p>
            </div>
          </div>
          {order.shipping_address && (
            <div className="mt-3 pt-3 border-t border-border-light dark:border-border-dark">
              <p className="text-sm text-text-sub dark:text-gray-400">Alamat Pengiriman:</p>
              <p className="text-sm text-text-main dark:text-white">{order.shipping_address}</p>
            </div>
          )}
        </div>

        {/* Order Items */}
        <div>
          <h4 className="font-medium text-text-main dark:text-white mb-3">Item Pesanan</h4>
          <div className="border border-border-light dark:border-border-dark rounded-lg overflow-hidden">
            <table className="w-full">
              <thead className="bg-background-light dark:bg-background-dark">
                <tr>
                  <th className="px-4 py-2 text-left text-xs font-medium text-text-sub dark:text-gray-400">Produk</th>
                  <th className="px-4 py-2 text-center text-xs font-medium text-text-sub dark:text-gray-400">Qty</th>
                  <th className="px-4 py-2 text-right text-xs font-medium text-text-sub dark:text-gray-400">Harga</th>
                  <th className="px-4 py-2 text-right text-xs font-medium text-text-sub dark:text-gray-400">Subtotal</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-border-light dark:divide-border-dark">
                {order.items?.map((item) => (
                  <tr key={item.id}>
                    <td className="px-4 py-3">
                      <p className="text-sm text-text-main dark:text-white">{item.product_name || item.sku}</p>
                      {item.variant_name && (
                        <p className="text-xs text-text-sub dark:text-gray-400">{item.variant_name}</p>
                      )}
                    </td>
                    <td className="px-4 py-3 text-center text-sm text-text-main dark:text-white">
                      {item.quantity}
                    </td>
                    <td className="px-4 py-3 text-right text-sm text-text-main dark:text-white">
                      {formatCurrency(item.unit_price)}
                    </td>
                    <td className="px-4 py-3 text-right text-sm font-medium text-text-main dark:text-white">
                      {formatCurrency(item.subtotal)}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>

        {/* Totals */}
        <div className="p-4 bg-background-light dark:bg-background-dark rounded-lg space-y-2">
          <div className="flex justify-between text-sm">
            <span className="text-text-sub dark:text-gray-400">Subtotal</span>
            <span className="text-text-main dark:text-white">{formatCurrency(order.total_amount)}</span>
          </div>
          {order.total_discount > 0 && (
            <div className="flex justify-between text-sm">
              <span className="text-text-sub dark:text-gray-400">Diskon</span>
              <span className="text-red-600">-{formatCurrency(order.total_discount)}</span>
            </div>
          )}
          {order.total_tax > 0 && (
            <div className="flex justify-between text-sm">
              <span className="text-text-sub dark:text-gray-400">Pajak</span>
              <span className="text-text-main dark:text-white">{formatCurrency(order.total_tax)}</span>
            </div>
          )}
          <div className="flex justify-between text-base font-semibold pt-2 border-t border-border-light dark:border-border-dark">
            <span className="text-text-main dark:text-white">Total</span>
            <span className="text-text-main dark:text-white">{formatCurrency(order.grand_total)}</span>
          </div>
          {order.paid_amount > 0 && (
            <div className="flex justify-between text-sm">
              <span className="text-text-sub dark:text-gray-400">Dibayar</span>
              <span className="text-green-600">{formatCurrency(order.paid_amount)}</span>
            </div>
          )}
          {order.balance_due > 0 && (
            <div className="flex justify-between text-sm font-medium">
              <span className="text-red-600">Sisa Bayar</span>
              <span className="text-red-600">{formatCurrency(order.balance_due)}</span>
            </div>
          )}
        </div>

        {/* Payments */}
        {order.payments && order.payments.length > 0 && (
          <div>
            <h4 className="font-medium text-text-main dark:text-white mb-3">Riwayat Pembayaran</h4>
            <div className="space-y-2">
              {order.payments.map((payment) => (
                <div
                  key={payment.id}
                  className="flex items-center justify-between p-3 bg-background-light dark:bg-background-dark rounded-lg"
                >
                  <div className="flex items-center gap-3">
                    <div className="w-8 h-8 rounded-full bg-green-100 dark:bg-green-900/30 flex items-center justify-center">
                      <Icon name="payments" size={16} className="text-green-600" />
                    </div>
                    <div>
                      <p className="text-sm font-medium text-text-main dark:text-white capitalize">
                        {payment.method.replace('_', ' ')}
                      </p>
                      <p className="text-xs text-text-sub dark:text-gray-400">
                        {formatDate(payment.paid_at)}
                      </p>
                    </div>
                  </div>
                  <p className="font-medium text-text-main dark:text-white">
                    {formatCurrency(payment.amount)}
                  </p>
                </div>
              ))}
            </div>
          </div>
        )}

        {/* Notes */}
        {order.notes && (
          <div className="p-4 bg-yellow-50 dark:bg-yellow-900/20 rounded-lg">
            <p className="text-sm font-medium text-yellow-700 dark:text-yellow-400 mb-1">Catatan:</p>
            <p className="text-sm text-text-main dark:text-white">{order.notes}</p>
          </div>
        )}

        {/* Cancel/Void Reason */}
        {order.cancel_reason && (
          <div className="p-4 bg-red-50 dark:bg-red-900/20 rounded-lg">
            <p className="text-sm font-medium text-red-700 dark:text-red-400 mb-1">Alasan Pembatalan:</p>
            <p className="text-sm text-text-main dark:text-white">{order.cancel_reason}</p>
          </div>
        )}
        {order.void_reason && (
          <div className="p-4 bg-orange-50 dark:bg-orange-900/20 rounded-lg">
            <p className="text-sm font-medium text-orange-700 dark:text-orange-400 mb-1">Alasan Void:</p>
            <p className="text-sm text-text-main dark:text-white">{order.void_reason}</p>
          </div>
        )}

        {/* Action Content */}
        {actionMode !== 'view' && renderActionContent()}

        {/* Action Buttons */}
        {actionMode === 'view' && (canConfirm || canAddPayment || canComplete || canCancel || canVoid) && (
          <div className="flex flex-wrap gap-2 pt-4 border-t border-border-light dark:border-border-dark">
            {canConfirm && (
              <Button
                onClick={handleConfirm}
                disabled={isLoading}
                leftIcon={<Icon name="check" size={18} />}
              >
                {isLoading ? 'Memproses...' : 'Konfirmasi Pesanan'}
              </Button>
            )}
            {canAddPayment && (
              <Button
                variant="outline"
                onClick={() => {
                  setPaymentAmount(order.balance_due.toString());
                  setActionMode('add-payment');
                }}
                leftIcon={<Icon name="payments" size={18} />}
              >
                Tambah Pembayaran
              </Button>
            )}
            {canComplete && (
              <Button
                onClick={handleComplete}
                disabled={isLoading}
                leftIcon={<Icon name="check_circle" size={18} />}
                className="bg-green-600 hover:bg-green-700"
              >
                {isLoading ? 'Memproses...' : 'Selesaikan'}
              </Button>
            )}
            {canCancel && (
              <Button
                variant="outline"
                onClick={() => setActionMode('cancel')}
                leftIcon={<Icon name="cancel" size={18} />}
                className="text-red-600 border-red-300 hover:bg-red-50 dark:hover:bg-red-900/20"
              >
                Batalkan
              </Button>
            )}
            {canVoid && (
              <Button
                variant="outline"
                onClick={() => setActionMode('void')}
                leftIcon={<Icon name="block" size={18} />}
                className="text-orange-600 border-orange-300 hover:bg-orange-50 dark:hover:bg-orange-900/20"
              >
                Void
              </Button>
            )}
          </div>
        )}
          </div>
        </div>
      </div>
    </div>
  );
}

import { Icon } from '@/components/ui';
import { formatCurrency } from '@/utils';
import { useAuth } from '@/contexts/AuthContext';
import { useCompanySettings } from '@/hooks/useCompanySettings';
import type { Order } from '@/types/order';

interface ReceiptModalProps {
  isOpen: boolean;
  onClose: () => void;
  order: Order;
}

const PAYMENT_METHOD_LABELS: Record<string, string> = {
  cash: 'Tunai',
  debit_card: 'Kartu Debit',
  credit_card: 'Kartu Kredit',
  e_wallet: 'E-Wallet',
  bank_transfer: 'Transfer Bank',
  qris: 'QRIS',
};

function ReceiptDivider({ dashed }: { dashed?: boolean }) {
  return (
    <div
      className={`w-full border-t my-2 ${
        dashed
          ? 'border-dashed border-gray-400'
          : 'border-solid border-gray-700'
      }`}
    />
  );
}

function ReceiptRow({
  label,
  value,
  bold,
  valueClass,
}: {
  label: string;
  value: string;
  bold?: boolean;
  valueClass?: string;
}) {
  return (
    <div className={`flex justify-between gap-2 text-xs ${bold ? 'font-bold' : ''}`}>
      <span className="text-gray-700 shrink-0">{label}</span>
      <span className={`text-right ${valueClass ?? 'text-gray-900'}`}>{value}</span>
    </div>
  );
}

export function ReceiptModal({ isOpen, onClose, order }: ReceiptModalProps) {
  const { user } = useAuth();
  const { data: settings } = useCompanySettings();

  const handlePrint = () => {
    window.print();
  };

  // Determine change amount (only for cash payments)
  const cashPaid = (order.payments ?? [])
    .filter(p => p.method === 'cash' && p.status === 'completed')
    .reduce((sum, p) => sum + p.amount, 0);
  const change = order.payment_status === 'paid' && cashPaid > order.grand_total
    ? cashPaid - order.grand_total
    : 0;

  const formatDateTime = (dateStr: string) => {
    const d = new Date(dateStr);
    return d.toLocaleString('id-ID', {
      day: '2-digit',
      month: '2-digit',
      year: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    });
  };

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 z-[60] flex items-center justify-center bg-black/60 p-4">
      {/* No-print overlay controls */}
      <div className="no-print absolute top-4 right-4 flex items-center gap-2 z-10">
        <button
          onClick={handlePrint}
          className="flex items-center gap-2 px-4 py-2 bg-white rounded-lg shadow-lg text-sm font-medium text-gray-700 hover:bg-gray-50 transition-colors"
        >
          <Icon name="print" size={18} />
          Cetak
        </button>
        <button
          onClick={onClose}
          className="p-2 bg-white rounded-lg shadow-lg text-gray-500 hover:bg-gray-50 transition-colors"
        >
          <Icon name="close" size={20} />
        </button>
      </div>

      {/* Receipt paper */}
      <div
        className="receipt-content bg-white shadow-2xl rounded"
        style={{ width: '320px', fontFamily: 'monospace' }}
      >
        <div className="px-6 py-5">
          {/* Header */}
          {settings?.receipt_header && (
            <p className="text-xs text-center text-gray-600 whitespace-pre-line mb-2">
              {settings.receipt_header}
            </p>
          )}

          <div className="text-center mb-3">
            <h1 className="text-base font-bold text-gray-900 uppercase tracking-wide">
              {user?.company?.name || 'Toko'}
            </h1>
            <p className="text-xs text-gray-500 mt-0.5">STRUK PEMBAYARAN</p>
          </div>

          <ReceiptDivider />

          {/* Order meta */}
          <div className="space-y-0.5 my-2">
            <ReceiptRow label="No. Order" value={order.order_no} />
            <ReceiptRow
              label="Tanggal"
              value={formatDateTime(order.created_at)}
            />
            {order.cashier_name && (
              <ReceiptRow label="Kasir" value={order.cashier_name} />
            )}
            <ReceiptRow
              label="Pelanggan"
              value={order.customer_name || 'Walk-in'}
            />
            <ReceiptRow
              label="Tipe"
              value={order.fulfillment_type === 'delivery' ? 'Delivery' : 'Kasir'}
            />
          </div>

          <ReceiptDivider dashed />

          {/* Items */}
          <div className="my-2 space-y-2">
            {(order.items ?? []).map((item) => (
              <div key={item.id}>
                <p className="text-xs font-semibold text-gray-900 leading-tight">
                  {item.product_name || item.sku}
                  {item.variant_name && item.variant_name !== 'Default' && (
                    <span> - {item.variant_name}</span>
                  )}
                </p>
                <div className="flex justify-between text-xs text-gray-700">
                  <span>
                    {item.quantity} × {formatCurrency(item.unit_price)}
                    {item.discount_amount > 0 && (
                      <span className="text-gray-400 ml-1">
                        (diskon -{formatCurrency(item.discount_amount)})
                      </span>
                    )}
                  </span>
                  <span className="font-medium">{formatCurrency(item.subtotal)}</span>
                </div>
              </div>
            ))}
          </div>

          <ReceiptDivider dashed />

          {/* Totals */}
          <div className="my-2 space-y-0.5">
            <ReceiptRow label="Subtotal" value={formatCurrency(order.total_amount)} />
            {order.total_discount > 0 && (
              <ReceiptRow
                label={
                  order.applied_promotion
                    ? `Diskon (${order.applied_promotion.code})`
                    : 'Diskon'
                }
                value={`-${formatCurrency(order.total_discount)}`}
                valueClass="text-red-600"
              />
            )}
            {settings?.tax_enabled && order.total_tax > 0 && (
              <ReceiptRow
                label={`PPN (${settings.tax_rate}%)`}
                value={formatCurrency(order.total_tax)}
              />
            )}
          </div>

          <ReceiptDivider />

          <div className="my-2">
            <ReceiptRow
              label="TOTAL"
              value={formatCurrency(order.grand_total)}
              bold
            />
          </div>

          <ReceiptDivider dashed />

          {/* Payments */}
          <div className="my-2 space-y-0.5">
            {(order.payments ?? [])
              .filter(p => p.status === 'completed')
              .map((payment) => (
                <ReceiptRow
                  key={payment.id}
                  label={PAYMENT_METHOD_LABELS[payment.method] ?? payment.method}
                  value={formatCurrency(payment.amount)}
                />
              ))}
            {change > 0 && (
              <ReceiptRow
                label="Kembalian"
                value={formatCurrency(change)}
                bold
                valueClass="text-green-700 font-bold"
              />
            )}
            {order.balance_due > 0 && (
              <ReceiptRow
                label="Sisa Bayar"
                value={formatCurrency(order.balance_due)}
                bold
                valueClass="text-red-600 font-bold"
              />
            )}
          </div>

          <ReceiptDivider />

          {/* Notes */}
          {order.notes && (
            <p className="text-xs text-gray-500 italic text-center my-2">
              Catatan: {order.notes}
            </p>
          )}

          {/* Footer */}
          <div className="text-center mt-3 space-y-1">
            {settings?.receipt_footer ? (
              <p className="text-xs text-gray-600 whitespace-pre-line">
                {settings.receipt_footer}
              </p>
            ) : (
              <p className="text-xs text-gray-500">Terima kasih atas kunjungan Anda!</p>
            )}
            <p className="text-[10px] text-gray-400 mt-2">
              {order.order_no} · {formatDateTime(order.created_at)}
            </p>
          </div>
        </div>
      </div>
    </div>
  );
}

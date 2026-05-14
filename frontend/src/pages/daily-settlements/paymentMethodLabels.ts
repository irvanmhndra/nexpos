import type { PaymentMethod } from '@/types/dailySettlement';

export const PAYMENT_METHOD_LABELS: Record<PaymentMethod, string> = {
  cash: 'Tunai',
  debit_card: 'Kartu Debit',
  credit_card: 'Kartu Kredit',
  e_wallet: 'E-Wallet',
  bank_transfer: 'Transfer Bank',
  qris: 'QRIS',
};

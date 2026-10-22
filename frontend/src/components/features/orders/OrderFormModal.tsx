import { useState, useEffect } from 'react';
import { Button, Input, Icon, SearchableSelect } from '@/components/ui';
import type { Customer } from '@/types/customer';
import type { Product } from '@/types/product';
import type { FulfillmentType, PaymentMethod } from '@/types/order';
import { useProducts } from '@/hooks/useProducts';
import { useCustomers } from '@/hooks/useCustomers';

interface OrderFormModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSubmit: (data: OrderFormData) => void;
  isLoading?: boolean;
}

export interface OrderFormData {
  customerId?: string;
  customerName: string;
  customerEmail: string;
  customerPhone: string;
  items: OrderItem[];
  fulfillmentType: FulfillmentType;
  paymentMethod: PaymentMethod;
  paymentAmount: number;
  shippingAddress: string;
  notes?: string;
  promoCode?: string;
}

export interface OrderItem {
  productId: string;
  productVariantId: number | null;
  productName: string;
  quantity: number;
  price: number;
}

export function OrderFormModal({
  isOpen,
  onClose,
  onSubmit,
  isLoading = false,
}: OrderFormModalProps) {
  const [formData, setFormData] = useState<OrderFormData>({
    customerId: '',
    customerName: '',
    customerEmail: '',
    customerPhone: '',
    items: [{ productId: '', productVariantId: null, productName: '', quantity: 1, price: 0 }],
    fulfillmentType: 'counter',
    paymentMethod: 'cash',
    paymentAmount: 0,
    shippingAddress: '',
    notes: '',
  });

  const [isManualCustomer, setIsManualCustomer] = useState(false);
  const [payNow, setPayNow] = useState(true);

  const { data: productsData } = useProducts({ per_page: 100, is_active: true });
  const products: Product[] = productsData?.data || [];

  const { data: customersData } = useCustomers({ per_page: 100 });
  const customers: Customer[] = customersData?.data || [];

  const [errors, setErrors] = useState<Partial<Record<keyof OrderFormData, string>>>({});

  // Reset form when modal opens
  useEffect(() => {
    if (isOpen) {
      // eslint-disable-next-line react-hooks/set-state-in-effect
      setIsManualCustomer(false);
       
      setPayNow(true);
       
      setFormData({
        customerId: '',
        customerName: '',
        customerEmail: '',
        customerPhone: '',
        items: [{ productId: '', productVariantId: null, productName: '', quantity: 1, price: 0 }],
        fulfillmentType: 'counter',
        paymentMethod: 'cash',
        paymentAmount: 0,
        shippingAddress: '',
        notes: '',
        promoCode: '',
      });
       
      setErrors({});
    }
  }, [isOpen]);

  const validateForm = (): boolean => {
    const newErrors: Partial<Record<keyof OrderFormData, string>> = {};

    // Customer validation is optional for walk-in customers
    if (formData.items.length === 0) {
      newErrors.items = 'Minimal 1 item produk';
    }
    const hasEmptyProduct = formData.items.some(item => !item.productId);
    if (hasEmptyProduct) {
      newErrors.items = 'Semua item harus memilih produk';
    }

    // Shipping address required for delivery
    if (formData.fulfillmentType === 'delivery' && !formData.shippingAddress.trim()) {
      newErrors.shippingAddress = 'Alamat pengiriman wajib diisi untuk delivery';
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (validateForm()) {
      // Prepare final form data with correct payment amount
      const finalData: OrderFormData = {
        ...formData,
        paymentAmount: payNow ? (formData.paymentAmount || calculateTotal()) : 0,
      };
      onSubmit(finalData);
    }
  };

  const addItem = () => {
    setFormData({
      ...formData,
      items: [...formData.items, { productId: '', productVariantId: null, productName: '', quantity: 1, price: 0 }],
    });
  };

  const handleProductSelect = (index: number, productId: string) => {
    const selectedProduct = products.find(p => String(p.id) === productId);
    if (selectedProduct) {
      const firstVariant = selectedProduct.variants?.[0];
      const newItems = [...formData.items];
      newItems[index] = {
        ...newItems[index],
        productId: String(selectedProduct.id),
        productVariantId: firstVariant?.id || null,
        productName: selectedProduct.name,
        price: firstVariant?.price || 0,
      };
      setFormData({ ...formData, items: newItems });
    }
  };

  const handleCustomerSelect = (customerId: string) => {
    const selectedCustomer = customers.find(c => String(c.id) === customerId);
    if (selectedCustomer) {
      setFormData({
        ...formData,
        customerId: String(selectedCustomer.id),
        customerName: selectedCustomer.name,
        customerEmail: selectedCustomer.email || '',
        customerPhone: selectedCustomer.phone || '',
      });
    }
  };

  const toggleManualCustomer = () => {
    setIsManualCustomer(!isManualCustomer);
    if (!isManualCustomer) {
      // Switching to manual - clear customerId but keep other data
      setFormData({ ...formData, customerId: '' });
    }
  };

  const removeItem = (index: number) => {
    setFormData({
      ...formData,
      items: formData.items.filter((_, i) => i !== index),
    });
  };

  const updateItem = (index: number, field: keyof OrderItem, value: string | number) => {
    const newItems = [...formData.items];
    newItems[index] = { ...newItems[index], [field]: value };
    setFormData({ ...formData, items: newItems });
  };

  const calculateTotal = () => {
    return formData.items.reduce((sum, item) => sum + item.price * item.quantity, 0);
  };

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/50">
      <div className="bg-white dark:bg-card-dark rounded-xl shadow-xl w-full max-w-3xl max-h-[90vh] overflow-hidden flex flex-col">
        {/* Header */}
        <div className="flex items-center justify-between p-6 border-b border-border-light dark:border-border-dark">
          <h2 className="text-xl font-bold text-text-main dark:text-white">
            Buat Pesanan Baru
          </h2>
          <button
            onClick={onClose}
            className="text-text-sub hover:text-text-main dark:hover:text-white transition-colors"
          >
            <Icon name="close" size={24} />
          </button>
        </div>

        {/* Form */}
        <form onSubmit={handleSubmit} className="flex-1 overflow-y-auto p-6">
          <div className="space-y-6">
            {/* Customer Info */}
            <div>
              <div className="flex items-center justify-between mb-4">
                <h3 className="text-lg font-bold text-text-main dark:text-white">
                  Informasi Pelanggan
                </h3>
                <button
                  type="button"
                  onClick={toggleManualCustomer}
                  className="text-sm text-primary hover:text-primary/80 font-medium flex items-center gap-1"
                >
                  <Icon name={isManualCustomer ? 'person_search' : 'edit'} size={16} />
                  {isManualCustomer ? 'Pilih dari daftar' : 'Input manual'}
                </button>
              </div>

              {!isManualCustomer ? (
                <div className="space-y-4">
                  <div>
                    <label className="block text-sm font-bold text-text-main dark:text-gray-200 mb-2">
                      Pilih Pelanggan <span className="text-red-500">*</span>
                    </label>
                    <SearchableSelect
                      options={customers.map(c => ({
                        value: String(c.id),
                        label: c.name,
                        description: `${c.phone || '-'} - ${c.email || '-'}`,
                      }))}
                      value={formData.customerId || ''}
                      onChange={handleCustomerSelect}
                      placeholder="Cari pelanggan..."
                      searchPlaceholder="Ketik nama atau nomor telepon..."
                      error={errors.customerName}
                    />
                  </div>
                  {formData.customerId && (
                    <div className="p-3 bg-background-light dark:bg-background-dark rounded-lg">
                      <p className="text-sm text-text-main dark:text-white font-medium">
                        {formData.customerName}
                      </p>
                      <p className="text-xs text-text-sub dark:text-gray-400 mt-1">
                        {formData.customerPhone} {formData.customerEmail && `- ${formData.customerEmail}`}
                      </p>
                    </div>
                  )}
                </div>
              ) : (
                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                  <div>
                    <label className="block text-sm font-bold text-text-main dark:text-gray-200 mb-2">
                      Nama Pelanggan <span className="text-red-500">*</span>
                    </label>
                    <Input
                      value={formData.customerName}
                      onChange={e => setFormData({ ...formData, customerName: e.target.value })}
                      placeholder="John Doe"
                      error={errors.customerName}
                    />
                  </div>

                  <div>
                    <label className="block text-sm font-bold text-text-main dark:text-gray-200 mb-2">
                      Nomor Telepon <span className="text-red-500">*</span>
                    </label>
                    <Input
                      value={formData.customerPhone}
                      onChange={e => setFormData({ ...formData, customerPhone: e.target.value })}
                      placeholder="081234567890"
                      error={errors.customerPhone}
                    />
                  </div>

                  <div className="md:col-span-2">
                    <label className="block text-sm font-bold text-text-main dark:text-gray-200 mb-2">
                      Email
                    </label>
                    <Input
                      type="email"
                      value={formData.customerEmail}
                      onChange={e => setFormData({ ...formData, customerEmail: e.target.value })}
                      placeholder="john@example.com"
                    />
                  </div>
                </div>
              )}
            </div>

            {/* Order Items */}
            <div>
              <div className="flex items-center justify-between mb-4">
                <h3 className="text-lg font-bold text-text-main dark:text-white">
                  Item Pesanan <span className="text-red-500">*</span>
                </h3>
                <Button type="button" variant="outline" size="sm" onClick={addItem}>
                  <Icon name="add" size={16} />
                  Tambah Item
                </Button>
              </div>

              {errors.items && (
                <p className="text-sm text-red-500 mb-2">{errors.items}</p>
              )}

              {/* Column Headers */}
              <div className="hidden md:grid grid-cols-[2fr_1fr_1fr_1fr_auto] gap-3 px-4 mb-2">
                <span className="text-xs font-semibold text-text-sub dark:text-gray-400 uppercase tracking-wide">
                  Produk
                </span>
                <span className="text-xs font-semibold text-text-sub dark:text-gray-400 uppercase tracking-wide">
                  Jumlah
                </span>
                <span className="text-xs font-semibold text-text-sub dark:text-gray-400 uppercase tracking-wide text-right">
                  Harga
                </span>
                <span className="text-xs font-semibold text-text-sub dark:text-gray-400 uppercase tracking-wide text-right">
                  Subtotal
                </span>
                <span className="w-10"></span>
              </div>

              <div className="space-y-3">
                {formData.items.map((item, index) => (
                  <div
                    key={index}
                    className="flex items-start gap-3 p-4 bg-background-light dark:bg-background-dark rounded-lg"
                  >
                    <div className="flex-1 grid grid-cols-1 md:grid-cols-[2fr_1fr_1fr_1fr] gap-3 items-center">
                      <div>
                        <span className="text-xs font-semibold text-text-sub dark:text-gray-400 uppercase tracking-wide md:hidden mb-1 block">
                          Produk
                        </span>
                        <SearchableSelect
                          options={products.map(p => {
                            const firstVariant = p.variants?.[0];
                            return {
                              value: String(p.id),
                              label: p.name,
                              description: `Rp ${(firstVariant?.price || 0).toLocaleString('id-ID')}`,
                            };
                          })}
                          value={item.productId}
                          onChange={(productId) => handleProductSelect(index, productId)}
                          placeholder="Pilih produk..."
                          searchPlaceholder="Cari produk..."
                        />
                      </div>
                      <div>
                        <span className="text-xs font-semibold text-text-sub dark:text-gray-400 uppercase tracking-wide md:hidden mb-1 block">
                          Jumlah
                        </span>
                        <Input
                          type="number"
                          placeholder="Qty"
                          value={item.quantity}
                          onChange={e => updateItem(index, 'quantity', Math.max(1, Number(e.target.value)))}
                          min="1"
                        />
                      </div>
                      <div className="text-right">
                        <span className="text-xs font-semibold text-text-sub dark:text-gray-400 uppercase tracking-wide md:hidden mb-1 block">
                          Harga
                        </span>
                        <p className="text-sm text-text-main dark:text-white h-12 flex items-center justify-end">
                          {item.price > 0 ? `Rp ${item.price.toLocaleString('id-ID')}` : '-'}
                        </p>
                      </div>
                      <div className="text-right">
                        <span className="text-xs font-semibold text-text-sub dark:text-gray-400 uppercase tracking-wide md:hidden mb-1 block">
                          Subtotal
                        </span>
                        <p className="text-sm font-semibold text-text-main dark:text-white h-12 flex items-center justify-end">
                          {item.price > 0 ? `Rp ${(item.price * item.quantity).toLocaleString('id-ID')}` : '-'}
                        </p>
                      </div>
                    </div>
                    {formData.items.length > 1 ? (
                      <button
                        type="button"
                        onClick={() => removeItem(index)}
                        className="p-2 hover:bg-red-50 dark:hover:bg-red-900/20 rounded-lg transition-colors"
                      >
                        <Icon name="delete" size={20} className="text-red-600 dark:text-red-400" />
                      </button>
                    ) : (
                      <div className="w-10" />
                    )}
                  </div>
                ))}
              </div>

              {/* Total */}
              <div className="flex justify-end mt-4 p-4 bg-primary/5 rounded-lg">
                <div className="text-right">
                  <p className="text-sm text-text-sub dark:text-gray-400">Total</p>
                  <p className="text-2xl font-bold text-primary">
                    Rp {calculateTotal().toLocaleString('id-ID')}
                  </p>
                </div>
              </div>
            </div>

            {/* Fulfillment Type */}
            <div>
              <h3 className="text-lg font-bold text-text-main dark:text-white mb-4">
                Tipe Penjualan
              </h3>
              <div className="grid grid-cols-2 gap-3">
                <button
                  type="button"
                  onClick={() => setFormData({ ...formData, fulfillmentType: 'counter', shippingAddress: '' })}
                  className={`p-4 rounded-lg border-2 transition-all ${
                    formData.fulfillmentType === 'counter'
                      ? 'border-primary bg-primary/5'
                      : 'border-border-light dark:border-border-dark hover:border-primary/50'
                  }`}
                >
                  <div className="flex flex-col items-center gap-2">
                    <Icon
                      name="storefront"
                      size={28}
                      className={formData.fulfillmentType === 'counter' ? 'text-primary' : 'text-text-sub'}
                    />
                    <span className={`font-medium ${
                      formData.fulfillmentType === 'counter' ? 'text-primary' : 'text-text-main dark:text-white'
                    }`}>
                      Counter
                    </span>
                    <span className="text-xs text-text-sub dark:text-gray-400">
                      Penjualan langsung di toko
                    </span>
                  </div>
                </button>
                <button
                  type="button"
                  onClick={() => setFormData({ ...formData, fulfillmentType: 'delivery' })}
                  className={`p-4 rounded-lg border-2 transition-all ${
                    formData.fulfillmentType === 'delivery'
                      ? 'border-primary bg-primary/5'
                      : 'border-border-light dark:border-border-dark hover:border-primary/50'
                  }`}
                >
                  <div className="flex flex-col items-center gap-2">
                    <Icon
                      name="local_shipping"
                      size={28}
                      className={formData.fulfillmentType === 'delivery' ? 'text-primary' : 'text-text-sub'}
                    />
                    <span className={`font-medium ${
                      formData.fulfillmentType === 'delivery' ? 'text-primary' : 'text-text-main dark:text-white'
                    }`}>
                      Delivery
                    </span>
                    <span className="text-xs text-text-sub dark:text-gray-400">
                      Pengiriman ke alamat
                    </span>
                  </div>
                </button>
              </div>

              {/* Shipping Address - only for delivery */}
              {formData.fulfillmentType === 'delivery' && (
                <div className="mt-4">
                  <label className="block text-sm font-bold text-text-main dark:text-gray-200 mb-2">
                    Alamat Pengiriman <span className="text-red-500">*</span>
                  </label>
                  <textarea
                    value={formData.shippingAddress}
                    onChange={e => setFormData({ ...formData, shippingAddress: e.target.value })}
                    placeholder="Alamat lengkap pengiriman..."
                    rows={3}
                    className={`w-full rounded-lg border ${
                      errors.shippingAddress
                        ? 'border-red-500'
                        : 'border-border-light dark:border-gray-600'
                    } bg-white dark:bg-card-dark text-text-main dark:text-white focus:border-primary focus:ring-1 focus:ring-primary p-4 resize-none`}
                  />
                  {errors.shippingAddress && (
                    <p className="text-sm text-red-500 mt-1">{errors.shippingAddress}</p>
                  )}
                </div>
              )}
            </div>

            {/* Payment */}
            <div>
              <div className="flex items-center justify-between mb-4">
                <h3 className="text-lg font-bold text-text-main dark:text-white">
                  Pembayaran
                </h3>
                <label className="flex items-center gap-2 cursor-pointer">
                  <input
                    type="checkbox"
                    checked={payNow}
                    onChange={e => setPayNow(e.target.checked)}
                    className="w-4 h-4 rounded border-border-light text-primary focus:ring-primary"
                  />
                  <span className="text-sm text-text-main dark:text-white">Bayar Sekarang</span>
                </label>
              </div>

              {payNow ? (
                <div className="p-4 bg-green-50 dark:bg-green-900/20 rounded-lg space-y-4">
                  <div className="flex items-center gap-2 text-green-700 dark:text-green-400">
                    <Icon name="check_circle" size={20} />
                    <span className="font-medium">Pembayaran Langsung</span>
                  </div>
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                    <div>
                      <label className="block text-sm font-medium text-text-sub dark:text-gray-400 mb-2">
                        Metode Pembayaran
                      </label>
                      <select
                        value={formData.paymentMethod}
                        onChange={e => setFormData({ ...formData, paymentMethod: e.target.value as PaymentMethod })}
                        className="w-full rounded-lg border border-border-light dark:border-gray-600 bg-white dark:bg-card-dark text-text-main dark:text-white focus:border-primary focus:ring-1 focus:ring-primary h-12 px-4 text-base"
                      >
                        <option value="cash">Cash</option>
                        <option value="debit_card">Kartu Debit</option>
                        <option value="credit_card">Kartu Kredit</option>
                        <option value="e_wallet">E-Wallet</option>
                        <option value="bank_transfer">Transfer Bank</option>
                        <option value="qris">QRIS</option>
                      </select>
                    </div>
                    <div>
                      <label className="block text-sm font-medium text-text-sub dark:text-gray-400 mb-2">
                        Jumlah Bayar
                      </label>
                      <div className="relative">
                        <span className="absolute left-4 top-1/2 -translate-y-1/2 text-text-sub">Rp</span>
                        <input
                          type="number"
                          value={formData.paymentAmount || calculateTotal()}
                          onChange={e => setFormData({ ...formData, paymentAmount: Number(e.target.value) })}
                          className="w-full h-12 pl-12 pr-4 rounded-lg border border-border-light dark:border-gray-600 bg-white dark:bg-card-dark text-text-main dark:text-white focus:border-primary focus:ring-1 focus:ring-primary"
                        />
                      </div>
                    </div>
                  </div>
                  <p className="text-sm text-text-sub dark:text-gray-400">
                    {formData.fulfillmentType === 'counter'
                      ? 'Pesanan counter akan otomatis selesai setelah pembayaran lunas.'
                      : 'Pesanan delivery akan dikonfirmasi setelah pembayaran.'}
                  </p>
                </div>
              ) : (
                <div className="p-4 bg-yellow-50 dark:bg-yellow-900/20 rounded-lg">
                  <div className="flex items-center gap-2 text-yellow-700 dark:text-yellow-400">
                    <Icon name="schedule" size={20} />
                    <span className="font-medium">Bayar Nanti</span>
                  </div>
                  <p className="text-sm text-text-sub dark:text-gray-400 mt-2">
                    Pesanan akan dibuat sebagai draft. Anda bisa menambahkan pembayaran nanti melalui detail pesanan.
                  </p>
                </div>
              )}
            </div>

            {/* Promo Code */}
            <div>
              <label className="block text-sm font-bold text-text-main dark:text-gray-200 mb-2">
                Kode Promo (Opsional)
              </label>
              <Input
                value={formData.promoCode || ''}
                onChange={e => setFormData({ ...formData, promoCode: e.target.value.toUpperCase() })}
                placeholder="Masukkan kode promo..."
                leftIcon={<Icon name="local_offer" size={20} />}
              />
            </div>

            {/* Notes */}
            <div>
              <label className="block text-sm font-bold text-text-main dark:text-gray-200 mb-2">
                Catatan (Opsional)
              </label>
              <textarea
                value={formData.notes}
                onChange={e => setFormData({ ...formData, notes: e.target.value })}
                placeholder="Catatan tambahan untuk pesanan..."
                rows={2}
                className="w-full rounded-lg border border-border-light dark:border-gray-600 bg-white dark:bg-card-dark text-text-main dark:text-white focus:border-primary focus:ring-1 focus:ring-primary p-4 resize-none"
              />
            </div>
          </div>
        </form>

        {/* Footer */}
        <div className="flex items-center justify-end gap-3 p-6 border-t border-border-light dark:border-border-dark">
          <Button variant="outline" onClick={onClose} disabled={isLoading}>
            Batal
          </Button>
          <Button onClick={handleSubmit} disabled={isLoading}>
            {isLoading ? 'Memproses...' : payNow ? 'Buat & Bayar' : 'Simpan Draft'}
          </Button>
        </div>
      </div>
    </div>
  );
}

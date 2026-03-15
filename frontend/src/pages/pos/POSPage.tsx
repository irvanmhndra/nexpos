import { useState, useMemo, useRef } from 'react';
import { Button, Input, Icon, Badge } from '@/components/ui';
import { useProducts } from '@/hooks/useProducts';
import { useAllProductCategories } from '@/hooks/useProductCategories';
import { useCustomers } from '@/hooks/useCustomers';
import { useCreateOrder } from '@/hooks/useOrders';
import { useInventory } from '@/hooks/useInventory';
import { usePromotions } from '@/hooks/usePromotions';
import { useBranch } from '@/contexts/BranchContext';
import { formatCurrency, getApiErrorMessage } from '@/utils';
import { ReceiptModal } from '@/components/features/orders/ReceiptModal';
import type { Product, ProductVariant } from '@/types/product';
import type { Customer } from '@/types/customer';
import type { PaymentMethod, CreateOrderRequest, Order } from '@/types/order';
import type { Promotion } from '@/types/promotion';

interface CartItem {
  productVariantId: number;
  productId: number;
  productName: string;
  variantName: string;
  price: number;
  quantity: number;
}

interface OrderSuccess {
  orderNo: string;
  grandTotal: number;
  paid: number;
  change: number;
  promoApplied?: string;
}

const PAYMENT_METHODS: { value: PaymentMethod; label: string; icon: string }[] = [
  { value: 'cash', label: 'Cash', icon: 'payments' },
  { value: 'debit_card', label: 'Debit', icon: 'credit_card' },
  { value: 'qris', label: 'QRIS', icon: 'qr_code_2' },
  { value: 'e_wallet', label: 'E-Wallet', icon: 'account_balance_wallet' },
  { value: 'credit_card', label: 'Kredit', icon: 'credit_score' },
  { value: 'bank_transfer', label: 'Transfer', icon: 'account_balance' },
];

export function POSPage() {
  const [search, setSearch] = useState('');
  const [selectedCategoryId, setSelectedCategoryId] = useState<number | null>(null);
  const [cart, setCart] = useState<CartItem[]>([]);
  const [variantPickerProduct, setVariantPickerProduct] = useState<Product | null>(null);
  const [selectedCustomer, setSelectedCustomer] = useState<Customer | null>(null);
  const [customerSearch, setCustomerSearch] = useState('');
  const [showCustomerSearch, setShowCustomerSearch] = useState(false);
  const [promoCode, setPromoCode] = useState('');
  const [promoInput, setPromoInput] = useState('');
  const [matchedPromo, setMatchedPromo] = useState<Promotion | null>(null);
  const [promoError, setPromoError] = useState('');
  const [showPayment, setShowPayment] = useState(false);
  const [paymentMethod, setPaymentMethod] = useState<PaymentMethod>('cash');
  const [paymentAmount, setPaymentAmount] = useState('');
  const [orderSuccess, setOrderSuccess] = useState<OrderSuccess | null>(null);
  const [completedOrder, setCompletedOrder] = useState<Order | null>(null);
  const [isReceiptOpen, setIsReceiptOpen] = useState(false);

  const isSubmittingRef = useRef(false);

  const { activeBranch } = useBranch();
  const createOrder = useCreateOrder();
  const { data: inventoryData } = useInventory({ per_page: 200 });

  const { data: productsData } = useProducts({
    per_page: 200,
    is_active: true,
    ...(selectedCategoryId ? { category_id: selectedCategoryId } : {}),
    ...(search ? { search } : {}),
  });

  const { data: categories } = useAllProductCategories();

  const { data: customersData } = useCustomers({
    per_page: 20,
    ...(customerSearch ? { search: customerSearch } : {}),
  });

  const { data: promotionsData } = usePromotions({ is_active: true, per_page: 100 });

  const products = productsData?.data || [];
  const cats = categories || [];
  const customers = customersData?.data || [];
  const promotions = promotionsData?.data || [];

  const totalItems = useMemo(() => cart.reduce((s, i) => s + i.quantity, 0), [cart]);
  const subtotal = useMemo(() => cart.reduce((s, i) => s + i.price * i.quantity, 0), [cart]);

  const promoDiscount = useMemo(() => {
    if (!matchedPromo || !matchedPromo.discount_value) return 0;
    if (matchedPromo.discount_type === 'percentage') {
      const disc = (subtotal * matchedPromo.discount_value) / 100;
      return matchedPromo.max_discount ? Math.min(disc, matchedPromo.max_discount) : disc;
    }
    return Math.min(matchedPromo.discount_value, subtotal);
  }, [matchedPromo, subtotal]);

  const paymentAmountNum = parseFloat(paymentAmount) || 0;
  const changeAmount = paymentAmountNum - subtotal;
  const stockMap = useMemo(() => {
    const map: Record<number, number> = {};
    inventoryData?.data.forEach(item => { map[item.product_variant_id] = item.current_stock; });
    return map;
  }, [inventoryData]);

  const variantStock = (variantId: number): number | undefined => {
    if (!activeBranch) return undefined;
    return stockMap[variantId];
  };

  const isVariantOOS = (variantId: number): boolean => {
    const s = variantStock(variantId);
    return s !== undefined && s <= 0;
  };

  const addToCart = (product: Product, variant: ProductVariant) => {
    if (isVariantOOS(variant.id)) return;
    setCart(prev => {
      const existing = prev.find(i => i.productVariantId === variant.id);
      if (existing) {
        return prev.map(i =>
          i.productVariantId === variant.id ? { ...i, quantity: i.quantity + 1 } : i
        );
      }
      return [
        ...prev,
        {
          productVariantId: variant.id,
          productId: product.id,
          productName: product.name,
          variantName: variant.name,
          price: variant.price,
          quantity: 1,
        },
      ];
    });
    setVariantPickerProduct(null);
  };

  const handleProductClick = (product: Product) => {
    const variants = product.variants || [];
    if (variants.length === 0) return;
    if (variants.length === 1) {
      if (isVariantOOS(variants[0].id)) return;
      addToCart(product, variants[0]);
    } else {
      setVariantPickerProduct(product);
    }
  };

  const updateQty = (variantId: number, delta: number) => {
    setCart(prev =>
      prev
        .map(i => (i.productVariantId === variantId ? { ...i, quantity: i.quantity + delta } : i))
        .filter(i => i.quantity > 0)
    );
  };

  const removeItem = (variantId: number) => {
    setCart(prev => prev.filter(i => i.productVariantId !== variantId));
  };

  const handleCheckout = async () => {
    if (cart.length === 0 || isSubmittingRef.current) return;
    isSubmittingRef.current = true;

    const req: CreateOrderRequest = {
      customer_id: selectedCustomer?.id || null,
      fulfillment_type: 'counter',
      items: cart.map(i => ({
        product_variant_id: i.productVariantId,
        quantity: i.quantity,
        discount_amount: 0,
      })),
      payments:
        paymentAmountNum > 0
          ? [{ method: paymentMethod, amount: paymentAmountNum }]
          : undefined,
      promo_code: promoCode || undefined,
    };

    try {
      const order = (await createOrder.mutateAsync(req)) as Order;
      setCompletedOrder(order);
      setOrderSuccess({
        orderNo: order.order_no,
        grandTotal: order.grand_total,
        paid: paymentAmountNum,
        change: Math.max(0, paymentAmountNum - order.grand_total),
        promoApplied: order.applied_promotion?.name,
      });
    } catch {
      // error surfaced via createOrder.error + global toast
    } finally {
      isSubmittingRef.current = false;
    }
  };

  const handleApplyPromo = () => {
    const code = promoInput.trim().toUpperCase();
    if (!code) return;
    const found = promotions.find(p => p.code.toUpperCase() === code);
    if (!found) {
      setPromoError('Kode promo tidak valid');
      return;
    }
    if (found.min_purchase && subtotal < found.min_purchase) {
      setPromoError(`Minimum pembelian ${formatCurrency(found.min_purchase)}`);
      return;
    }
    setMatchedPromo(found);
    setPromoCode(code);
    setPromoError('');
  };

  const resetPOS = () => {
    setCart([]);
    setSelectedCustomer(null);
    setPromoCode('');
    setPromoInput('');
    setMatchedPromo(null);
    setPromoError('');
    setPaymentAmount('');
    setPaymentMethod('cash');
    setShowPayment(false);
    setOrderSuccess(null);
    setCompletedOrder(null);
    setSearch('');
    setSelectedCategoryId(null);
  };

  // Success screen
  if (orderSuccess) {
    return (
      <>
      <div className="h-full flex items-center justify-center bg-background-light dark:bg-background-dark">
        <div className="bg-white dark:bg-card-dark rounded-2xl p-10 shadow-lg max-w-sm w-full text-center space-y-5">
          <div className="w-20 h-20 rounded-full bg-green-100 dark:bg-green-900/30 flex items-center justify-center mx-auto">
            <Icon name="check_circle" size={48} className="text-green-600" />
          </div>
          <div>
            <h2 className="text-2xl font-bold text-text-main dark:text-white">Pembayaran Berhasil!</h2>
            <p className="text-text-sub font-mono text-sm mt-1">{orderSuccess.orderNo}</p>
          </div>

          <div className="bg-background-light dark:bg-background-dark rounded-xl p-4 text-left space-y-2.5">
            <div className="flex justify-between text-sm">
              <span className="text-text-sub dark:text-gray-400">Total</span>
              <span className="font-semibold text-text-main dark:text-white">
                {formatCurrency(orderSuccess.grandTotal)}
              </span>
            </div>
            <div className="flex justify-between text-sm">
              <span className="text-text-sub dark:text-gray-400">Dibayar</span>
              <span className="font-semibold text-text-main dark:text-white">
                {formatCurrency(orderSuccess.paid)}
              </span>
            </div>
            {orderSuccess.change > 0 && (
              <div className="flex justify-between text-lg font-bold text-green-600 dark:text-green-400 border-t border-border-light dark:border-border-dark pt-2.5">
                <span>Kembalian</span>
                <span>{formatCurrency(orderSuccess.change)}</span>
              </div>
            )}
            {orderSuccess.promoApplied && (
              <p className="text-xs text-primary text-center pt-1">
                ✓ Promo "{orderSuccess.promoApplied}" diterapkan
              </p>
            )}
          </div>

          <div className="flex flex-col gap-2 w-full">
            <Button
              variant="outline"
              onClick={() => setIsReceiptOpen(true)}
              className="w-full"
              leftIcon={<Icon name="receipt" size={18} />}
            >
              Cetak Struk
            </Button>
            <Button onClick={resetPOS} className="w-full" size="lg">
              <Icon name="add_shopping_cart" size={20} />
              Pesanan Baru
            </Button>
          </div>
        </div>
      </div>

      {completedOrder && (
        <ReceiptModal
          isOpen={isReceiptOpen}
          onClose={() => setIsReceiptOpen(false)}
          order={completedOrder}
        />
      )}
      </>
    );
  }

  return (
    <div className="h-full flex overflow-hidden">
      {/* Left: Product Browser */}
      <div className="flex-1 flex flex-col overflow-hidden border-r border-border-light dark:border-border-dark">
        {/* Search bar */}
        <div className="p-4 border-b border-border-light dark:border-border-dark bg-white dark:bg-card-dark">
          <Input
            placeholder="Cari produk atau SKU..."
            leftIcon={<Icon name="search" size={20} />}
            value={search}
            onChange={e => setSearch(e.target.value)}
          />
        </div>

        {/* Category filter tabs */}
        <div className="flex gap-2 px-4 py-2.5 overflow-x-auto border-b border-border-light dark:border-border-dark bg-white dark:bg-card-dark shrink-0">
          <button
            onClick={() => setSelectedCategoryId(null)}
            className={`px-3 py-1.5 rounded-full text-sm font-medium whitespace-nowrap transition-colors ${
              selectedCategoryId === null
                ? 'bg-primary text-white'
                : 'text-text-sub hover:text-text-main dark:hover:text-white bg-background-light dark:bg-background-dark'
            }`}
          >
            Semua
          </button>
          {cats.map(cat => (
            <button
              key={cat.id}
              onClick={() => setSelectedCategoryId(cat.id)}
              className={`px-3 py-1.5 rounded-full text-sm font-medium whitespace-nowrap transition-colors ${
                selectedCategoryId === cat.id
                  ? 'bg-primary text-white'
                  : 'text-text-sub hover:text-text-main dark:hover:text-white bg-background-light dark:bg-background-dark'
              }`}
            >
              {cat.name}
            </button>
          ))}
        </div>

        {/* Product grid */}
        <div className="flex-1 overflow-y-auto p-4 bg-background-light dark:bg-background-dark">
          {products.length === 0 ? (
            <div className="h-full flex items-center justify-center">
              <div className="text-center text-text-sub dark:text-gray-400">
                <Icon name="inventory_2" size={48} className="mx-auto mb-2" />
                <p>Tidak ada produk</p>
              </div>
            </div>
          ) : (
            <div className="grid grid-cols-2 sm:grid-cols-3 xl:grid-cols-4 gap-3">
              {products.map(product => {
                const firstVariant = product.variants?.[0];
                const hasMultiVariant = (product.variants?.length || 0) > 1;
                const cartQty = cart
                  .filter(i => i.productId === product.id)
                  .reduce((s, i) => s + i.quantity, 0);
                const singleOOS = !hasMultiVariant && !!firstVariant && isVariantOOS(firstVariant.id);
                const singleStock = !hasMultiVariant && firstVariant ? variantStock(firstVariant.id) : undefined;
                const multiTotalStock = hasMultiVariant
                  ? product.variants?.reduce((s, v) => s + (stockMap[v.id] ?? 0), 0)
                  : undefined;

                return (
                  <button
                    key={product.id}
                    onClick={() => handleProductClick(product)}
                    disabled={singleOOS}
                    className={`relative text-left p-3 rounded-xl border bg-white dark:bg-card-dark transition-all border-border-light dark:border-border-dark ${singleOOS ? 'opacity-60 cursor-not-allowed' : 'hover:border-primary hover:shadow-md active:scale-[0.98]'}`}
                  >
                    {cartQty > 0 && (
                      <span className="absolute top-2 right-2 min-w-[20px] h-5 rounded-full bg-primary text-white text-xs font-bold flex items-center justify-center px-1">
                        {cartQty}
                      </span>
                    )}
                    {singleOOS && (
                      <span className="absolute top-2 left-2 bg-red-500 text-white text-[10px] font-bold px-1.5 py-0.5 rounded-full">
                        Habis
                      </span>
                    )}
                    <div className="aspect-square rounded-lg bg-background-light dark:bg-background-dark flex items-center justify-center mb-3 overflow-hidden">
                      {product.image_data ? (
                        <img
                          src={product.image_data}
                          alt={product.name}
                          className="w-full h-full object-cover"
                        />
                      ) : (
                        <Icon name="inventory_2" size={32} className="text-text-sub dark:text-gray-500" />
                      )}
                    </div>
                    <p className="text-sm font-semibold text-text-main dark:text-white line-clamp-2 mb-1">
                      {product.name}
                    </p>
                    {hasMultiVariant ? (
                      <>
                        <p className="text-xs text-primary font-medium">
                          {product.variants?.length} varian
                        </p>
                        {multiTotalStock !== undefined && (
                          <p className={`text-xs mt-0.5 ${multiTotalStock === 0 ? 'text-red-500 font-medium' : multiTotalStock <= 5 ? 'text-amber-500' : 'text-text-sub dark:text-gray-400'}`}>
                            Stok: {multiTotalStock}
                          </p>
                        )}
                      </>
                    ) : (
                      <>
                        <p className="text-sm font-bold text-primary">
                          {firstVariant ? formatCurrency(firstVariant.price) : '-'}
                        </p>
                        {singleStock !== undefined && (
                          <p className={`text-xs mt-0.5 ${singleStock === 0 ? 'text-red-500 font-medium' : singleStock <= 5 ? 'text-amber-500' : 'text-text-sub dark:text-gray-400'}`}>
                            Stok: {singleStock}
                          </p>
                        )}
                      </>
                    )}
                  </button>
                );
              })}
            </div>
          )}
        </div>
      </div>

      {/* Right: Cart Panel */}
      <div className="w-80 xl:w-96 flex flex-col bg-white dark:bg-card-dark shrink-0">
        {/* Cart header + customer */}
        <div className="px-4 pt-4 pb-3 border-b border-border-light dark:border-border-dark space-y-3">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <h2 className="font-bold text-text-main dark:text-white">Keranjang</h2>
              {totalItems > 0 && (
                <Badge variant="primary" size="sm">{totalItems}</Badge>
              )}
            </div>
            <div className="flex items-center gap-3">
              {activeBranch && (
                <span className="text-xs text-text-sub dark:text-gray-400">{activeBranch.name}</span>
              )}
              {cart.length > 0 && (
                <button
                  onClick={() => setCart([])}
                  className="text-xs text-red-500 hover:text-red-700 font-medium"
                >
                  Hapus Semua
                </button>
              )}
            </div>
          </div>

          {/* Customer selector */}
          {selectedCustomer ? (
            <div className="flex items-center justify-between p-2.5 bg-blue-50 dark:bg-blue-900/20 rounded-lg">
              <div className="flex items-center gap-2">
                <Icon name="person" size={16} className="text-primary" />
                <span className="text-sm font-medium text-text-main dark:text-white">
                  {selectedCustomer.name}
                </span>
              </div>
              <button onClick={() => setSelectedCustomer(null)}>
                <Icon name="close" size={16} className="text-text-sub" />
              </button>
            </div>
          ) : (
            <div>
              <button
                onClick={() => setShowCustomerSearch(!showCustomerSearch)}
                className="w-full flex items-center gap-2 p-2.5 rounded-lg border border-dashed border-border-light dark:border-border-dark text-text-sub hover:text-primary hover:border-primary transition-colors text-sm"
              >
                <Icon name="person_add" size={16} />
                Tambah Pelanggan (Opsional)
              </button>
              {showCustomerSearch && (
                <div className="mt-2 space-y-1">
                  <Input
                    placeholder="Cari pelanggan..."
                    value={customerSearch}
                    onChange={e => setCustomerSearch(e.target.value)}
                  />
                  <div className="max-h-36 overflow-y-auto rounded-lg border border-border-light dark:border-border-dark bg-white dark:bg-card-dark">
                    {customers.map(c => (
                      <button
                        key={c.id}
                        onClick={() => {
                          setSelectedCustomer(c);
                          setShowCustomerSearch(false);
                          setCustomerSearch('');
                        }}
                        className="w-full text-left px-3 py-2 text-sm hover:bg-background-light dark:hover:bg-background-dark transition-colors border-b border-border-light dark:border-border-dark last:border-0"
                      >
                        <p className="font-medium text-text-main dark:text-white">{c.name}</p>
                        {c.phone && <p className="text-xs text-text-sub dark:text-gray-400">{c.phone}</p>}
                      </button>
                    ))}
                    {customers.length === 0 && (
                      <p className="px-3 py-2 text-xs text-text-sub dark:text-gray-400">
                        Pelanggan tidak ditemukan
                      </p>
                    )}
                  </div>
                </div>
              )}
            </div>
          )}
        </div>

        {/* Cart items list */}
        <div className="flex-1 overflow-y-auto px-4 py-2">
          {cart.length === 0 ? (
            <div className="h-full flex items-center justify-center">
              <div className="text-center text-text-sub dark:text-gray-400">
                <Icon name="shopping_cart" size={40} className="mx-auto mb-2" />
                <p className="text-sm font-medium">Keranjang kosong</p>
                <p className="text-xs mt-1">Klik produk untuk menambahkan</p>
              </div>
            </div>
          ) : (
            <div className="space-y-1">
              {cart.map(item => (
                <div
                  key={item.productVariantId}
                  className="flex items-center gap-2 py-2.5 border-b border-border-light dark:border-border-dark last:border-0"
                >
                  <div className="flex-1 min-w-0">
                    <p className="text-sm font-medium text-text-main dark:text-white truncate">
                      {item.productName}
                    </p>
                    {item.variantName !== 'Default' && (
                      <p className="text-xs text-text-sub dark:text-gray-400">{item.variantName}</p>
                    )}
                    <p className="text-xs font-semibold text-primary mt-0.5">
                      {formatCurrency(item.price)}
                    </p>
                  </div>

                  <div className="flex items-center gap-1 shrink-0">
                    <button
                      onClick={() => updateQty(item.productVariantId, -1)}
                      className="size-7 rounded-lg border border-border-light dark:border-border-dark flex items-center justify-center hover:bg-background-light dark:hover:bg-background-dark transition-colors"
                    >
                      <Icon name="remove" size={14} />
                    </button>
                    <span className="w-7 text-center text-sm font-bold text-text-main dark:text-white">
                      {item.quantity}
                    </span>
                    <button
                      onClick={() => updateQty(item.productVariantId, 1)}
                      className="size-7 rounded-lg border border-border-light dark:border-border-dark flex items-center justify-center hover:bg-background-light dark:hover:bg-background-dark transition-colors"
                    >
                      <Icon name="add" size={14} />
                    </button>
                  </div>

                  <div className="flex items-center gap-1 shrink-0">
                    <p className="text-sm font-bold text-text-main dark:text-white w-20 text-right">
                      {formatCurrency(item.price * item.quantity)}
                    </p>
                    <button
                      onClick={() => removeItem(item.productVariantId)}
                      className="size-7 flex items-center justify-center text-text-sub hover:text-red-500 transition-colors"
                    >
                      <Icon name="delete" size={16} />
                    </button>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>

        {/* Summary + payment */}
        <div className="border-t border-border-light dark:border-border-dark p-4 space-y-3">
          {/* Promo code (hidden during payment step) */}
          {!showPayment && (
            promoCode ? (
              <div className="flex items-center justify-between px-3 py-2 rounded-lg bg-primary/10 border border-primary/20">
                <div className="flex items-center gap-2">
                  <Icon name="local_offer" size={16} className="text-primary" />
                  <span className="text-sm font-medium text-primary">{promoCode}</span>
                  {promoDiscount > 0 ? (
                    <span className="text-xs font-medium text-green-600 dark:text-green-400">
                      · Hemat {formatCurrency(promoDiscount)}
                    </span>
                  ) : (
                    <span className="text-xs text-text-sub dark:text-gray-400">akan diterapkan saat checkout</span>
                  )}
                </div>
                <button
                  onClick={() => { setPromoCode(''); setPromoInput(''); setMatchedPromo(null); setPromoError(''); }}
                  className="text-text-sub hover:text-red-500 transition-colors"
                >
                  <Icon name="close" size={16} />
                </button>
              </div>
            ) : (
              <div className="space-y-1">
                <div className="flex gap-2">
                  <Input
                    placeholder="Kode promo (opsional)..."
                    leftIcon={<Icon name="local_offer" size={18} />}
                    value={promoInput}
                    onChange={e => { setPromoInput(e.target.value.toUpperCase()); setPromoError(''); }}
                    onKeyDown={e => { if (e.key === 'Enter') handleApplyPromo(); }}
                  />
                  <Button
                    variant="outline"
                    onClick={handleApplyPromo}
                    disabled={!promoInput.trim()}
                    className="shrink-0"
                  >
                    Terapkan
                  </Button>
                </div>
                {promoError && (
                  <p className="text-xs text-red-500 px-1">{promoError}</p>
                )}
              </div>
            )
          )}

          {/* Subtotal row */}
          <div className="flex justify-between items-center">
            <span className="text-sm text-text-sub dark:text-gray-400">
              Subtotal ({totalItems} item)
            </span>
            <span className="text-lg font-bold text-text-main dark:text-white">
              {formatCurrency(subtotal)}
            </span>
          </div>

          {!showPayment ? (
            <Button
              onClick={() => setShowPayment(true)}
              disabled={cart.length === 0}
              className="w-full"
              size="lg"
            >
              <Icon name="point_of_sale" size={20} />
              Bayar
            </Button>
          ) : (
            <div className="space-y-3">
              {/* Payment method grid */}
              <div className="grid grid-cols-3 gap-1.5">
                {PAYMENT_METHODS.map(m => (
                  <button
                    key={m.value}
                    onClick={() => setPaymentMethod(m.value)}
                    className={`flex flex-col items-center gap-1 py-2 px-1 rounded-lg border text-xs font-medium transition-all ${
                      paymentMethod === m.value
                        ? 'border-primary bg-primary/10 text-primary dark:bg-primary/20'
                        : 'border-border-light dark:border-border-dark text-text-sub hover:border-primary/50'
                    }`}
                  >
                    <Icon name={m.icon} size={20} />
                    {m.label}
                  </button>
                ))}
              </div>

              {/* Amount input */}
              <div>
                <label className="text-xs font-medium text-text-sub dark:text-gray-400 mb-1 block">
                  Jumlah Bayar
                </label>
                <div className="relative">
                  <span className="absolute left-3 top-1/2 -translate-y-1/2 text-sm text-text-sub font-medium">
                    Rp
                  </span>
                  <input
                    type="number"
                    value={paymentAmount}
                    onChange={e => setPaymentAmount(e.target.value)}
                    placeholder={String(subtotal)}
                    autoFocus
                    className="w-full h-12 pl-10 pr-4 rounded-lg border border-border-light dark:border-gray-600 bg-white dark:bg-card-dark text-text-main dark:text-white focus:border-primary focus:ring-1 focus:ring-primary text-lg font-bold"
                  />
                </div>
              </div>

              {/* Change display */}
              {paymentAmountNum > 0 && (
                <div
                  className={`flex justify-between items-center p-2.5 rounded-lg ${
                    changeAmount >= 0
                      ? 'bg-green-50 dark:bg-green-900/20'
                      : 'bg-red-50 dark:bg-red-900/20'
                  }`}
                >
                  <span className="text-sm font-medium text-text-main dark:text-white">
                    {changeAmount >= 0 ? 'Kembalian' : 'Kekurangan'}
                  </span>
                  <span
                    className={`text-base font-bold ${
                      changeAmount >= 0
                        ? 'text-green-600 dark:text-green-400'
                        : 'text-red-600 dark:text-red-400'
                    }`}
                  >
                    {formatCurrency(Math.abs(changeAmount))}
                  </span>
                </div>
              )}

              {createOrder.isError && (
                <div className="flex items-start gap-2 p-3 rounded-lg bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800">
                  <Icon name="error" size={16} className="text-red-500 shrink-0 mt-0.5" />
                  <p className="text-xs text-red-600 dark:text-red-400">
                    {getApiErrorMessage(createOrder.error)}
                  </p>
                </div>
              )}

              <div className="flex gap-2">
                <Button
                  variant="outline"
                  onClick={() => setShowPayment(false)}
                  className="flex-1"
                >
                  Batal
                </Button>
                <Button
                  onClick={handleCheckout}
                  disabled={createOrder.isPending || cart.length === 0 || !paymentAmount}
                  className="flex-[2]"
                >
                  {createOrder.isPending ? (
                    <>
                      <Icon name="progress_activity" size={18} className="animate-spin" />
                      Memproses...
                    </>
                  ) : (
                    'Konfirmasi'
                  )}
                </Button>
              </div>
            </div>
          )}
        </div>
      </div>

      {/* Variant Picker Modal */}
      {variantPickerProduct && (
        <div
          className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/50"
          onClick={() => setVariantPickerProduct(null)}
        >
          <div
            className="bg-white dark:bg-card-dark rounded-xl shadow-xl w-full max-w-sm"
            onClick={e => e.stopPropagation()}
          >
            <div className="flex items-center justify-between p-4 border-b border-border-light dark:border-border-dark">
              <div>
                <h3 className="font-bold text-text-main dark:text-white">
                  {variantPickerProduct.name}
                </h3>
                <p className="text-xs text-text-sub dark:text-gray-400 mt-0.5">Pilih varian</p>
              </div>
              <button
                onClick={() => setVariantPickerProduct(null)}
                className="size-8 flex items-center justify-center rounded-lg hover:bg-background-light dark:hover:bg-background-dark"
              >
                <Icon name="close" size={20} className="text-text-sub" />
              </button>
            </div>
            <div className="p-4 grid grid-cols-2 gap-2">
              {variantPickerProduct.variants?.map(v => {
                const oos = isVariantOOS(v.id);
                const stock = variantStock(v.id);
                return (
                  <button
                    key={v.id}
                    onClick={() => addToCart(variantPickerProduct, v)}
                    disabled={oos}
                    className={`p-3 rounded-lg border text-left transition-colors ${oos ? 'border-border-light dark:border-border-dark opacity-60 cursor-not-allowed' : 'border-border-light dark:border-border-dark hover:border-primary hover:bg-primary/5'}`}
                  >
                    <p className="text-sm font-semibold text-text-main dark:text-white">{v.name}</p>
                    <p className="text-sm text-primary font-bold mt-1">{formatCurrency(v.price)}</p>
                    {stock !== undefined && (
                      <p className={`text-xs mt-0.5 ${oos ? 'text-red-500 font-medium' : stock <= 5 ? 'text-amber-500' : 'text-text-sub dark:text-gray-400'}`}>
                        {oos ? 'Habis' : `Stok: ${stock}`}
                      </p>
                    )}
                  </button>
                );
              })}
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

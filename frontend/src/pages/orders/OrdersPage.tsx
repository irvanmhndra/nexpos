import { useState } from 'react';
import { Card, Button, Icon, Dropdown, DateRangePicker } from '@/components/ui';
import { OrderFormModal } from '@/components/features/orders/OrderFormModal';
import { OrderDetailModal } from '@/components/features/orders/OrderDetailModal';
import type { OrderFormData } from '@/components/features/orders/OrderFormModal';
import { useOrders, useCreateOrder } from '@/hooks/useOrders';
import { formatCurrency, formatDate } from '@/utils';
import type { Order, CreateOrderRequest, OrderStatus, OrderPaymentStatus, PaymentMethod } from '@/types/order';

export function OrdersPage() {
  const [page, setPage] = useState(1);
  const [status, setStatus] = useState<OrderStatus | ''>('');
  const [search, setSearch] = useState('');
  const [searchInput, setSearchInput] = useState('');
  const [paymentMethod, setPaymentMethod] = useState<PaymentMethod | ''>('');
  const [startDate, setStartDate] = useState('');
  const [endDate, setEndDate] = useState('');
  const [isFormModalOpen, setIsFormModalOpen] = useState(false);
  const [selectedOrder, setSelectedOrder] = useState<Order | null>(null);
  const [isDetailModalOpen, setIsDetailModalOpen] = useState(false);

  const { data, isLoading } = useOrders({
    page,
    per_page: 10,
    status: status || undefined,
    search,
    payment_method: paymentMethod || undefined,
    date_from: startDate || undefined,
    date_to: endDate || undefined,
  });

  const createOrder = useCreateOrder();

  const handleCreateOrder = () => {
    setIsFormModalOpen(true);
  };

  const handleFormSubmit = async (formData: OrderFormData) => {
    try {
      // Calculate total amount
      const totalAmount = formData.items.reduce(
        (sum, item) => sum + item.price * item.quantity,
        0
      );

      // Map OrderFormData to CreateOrderRequest
      const createOrderData: CreateOrderRequest = {
        customer_id: formData.customerId ? Number(formData.customerId) : null,
        fulfillment_type: formData.fulfillmentType,
        shipping_address: formData.fulfillmentType === 'delivery' ? formData.shippingAddress : undefined,
        items: formData.items
          .filter(item => item.productVariantId !== null)
          .map(item => ({
            product_variant_id: item.productVariantId!,
            quantity: item.quantity,
            discount_amount: 0,
          })),
        // Only include payment if paymentAmount > 0
        payments: formData.paymentAmount > 0
          ? [
              {
                method: formData.paymentMethod,
                amount: formData.paymentAmount || totalAmount,
              },
            ]
          : undefined,
        notes: formData.notes || undefined,
        promo_code: formData.promoCode || undefined,
      };

      await createOrder.mutateAsync(createOrderData);
      setIsFormModalOpen(false);
    } catch (error) {
      console.error('Error saving order:', error);
    }
  };

  const handleViewOrder = (order: Order) => {
    setSelectedOrder(order);
    setIsDetailModalOpen(true);
  };

  const handleDetailModalClose = () => {
    setIsDetailModalOpen(false);
    setSelectedOrder(null);
  };

  const handleStatusFilter = (value: string) => {
    setStatus(value as OrderStatus | '');
    setPage(1);
  };

  const handleSearch = () => {
    setSearch(searchInput);
    setPage(1);
  };

  const handleSearchKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') {
      handleSearch();
    }
  };

  const handlePaymentMethodFilter = (value: string) => {
    setPaymentMethod(value as PaymentMethod | '');
    setPage(1);
  };

  const handleDateRangeChange = (start: string, end: string) => {
    setStartDate(start);
    setEndDate(end);
    setPage(1);
  };

  const handleClearFilters = () => {
    setSearch('');
    setSearchInput('');
    setPaymentMethod('');
    setStartDate('');
    setEndDate('');
    setStatus('');
    setPage(1);
  };

  const hasActiveFilters = search || paymentMethod || startDate || endDate || status;

  const getStatusBadge = (orderStatus: OrderStatus) => {
    const badges: Record<OrderStatus, string> = {
      draft: 'bg-gray-100 text-gray-700 dark:bg-gray-900/30 dark:text-gray-400',
      confirmed: 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400',
      completed: 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400',
      cancelled: 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400',
      voided: 'bg-orange-100 text-orange-700 dark:bg-orange-900/30 dark:text-orange-400',
    };
    return badges[orderStatus] || 'bg-gray-100 text-gray-700';
  };

  const getStatusLabel = (orderStatus: OrderStatus) => {
    const labels: Record<OrderStatus, string> = {
      draft: 'Draft',
      confirmed: 'Dikonfirmasi',
      completed: 'Selesai',
      cancelled: 'Dibatalkan',
      voided: 'Void',
    };
    return labels[orderStatus] || orderStatus;
  };

  const getPaymentStatusBadge = (paymentStatus: OrderPaymentStatus) => {
    const badges: Record<OrderPaymentStatus, string> = {
      unpaid: 'bg-gray-100 text-gray-700 dark:bg-gray-900/30 dark:text-gray-400',
      partial: 'bg-yellow-100 text-yellow-700 dark:bg-yellow-900/30 dark:text-yellow-400',
      paid: 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400',
      refunded: 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400',
    };
    return badges[paymentStatus] || 'bg-gray-100 text-gray-700';
  };

  const getPaymentStatusLabel = (paymentStatus: OrderPaymentStatus) => {
    const labels: Record<OrderPaymentStatus, string> = {
      unpaid: 'Belum Bayar',
      partial: 'Sebagian',
      paid: 'Lunas',
      refunded: 'Dikembalikan',
    };
    return labels[paymentStatus] || paymentStatus;
  };

  // Count orders by status
  const draftCount = data?.data.filter(o => o.status === 'draft').length || 0;
  const confirmedCount = data?.data.filter(o => o.status === 'confirmed').length || 0;
  const completedCount = data?.data.filter(o => o.status === 'completed').length || 0;
  const cancelledCount = data?.data.filter(o => o.status === 'cancelled' || o.status === 'voided').length || 0;

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-text-main dark:text-white">
            Manajemen Pesanan
          </h1>
          <p className="text-text-sub dark:text-gray-400 text-sm mt-1">
            Lacak dan kelola semua pesanan penjualan Anda
          </p>
        </div>
        <Button leftIcon={<Icon name="add" size={20} />} onClick={handleCreateOrder}>
          Buat Pesanan
        </Button>
      </div>

      {/* Stats Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <Card variant="bordered" padding="md">
          <div className="flex justify-between items-start">
            <p className="text-sm font-medium text-text-sub dark:text-gray-400">Total Pesanan</p>
            <span className="p-1.5 rounded-lg bg-primary/10">
              <Icon name="shopping_bag" size={20} className="text-primary" />
            </span>
          </div>
          <div className="flex items-end gap-2 mt-2">
            <p className="text-2xl font-bold text-text-main dark:text-white tracking-tight">
              {data?.total || 0}
            </p>
          </div>
        </Card>

        <Card variant="bordered" padding="md">
          <div className="flex justify-between items-start">
            <p className="text-sm font-medium text-text-sub dark:text-gray-400">Menunggu</p>
            <span className="p-1.5 rounded-lg bg-blue-500/10">
              <Icon name="pending_actions" size={20} className="text-blue-500" />
            </span>
          </div>
          <div className="flex items-end gap-2 mt-2">
            <p className="text-2xl font-bold text-text-main dark:text-white tracking-tight">
              {draftCount + confirmedCount}
            </p>
            <span className="text-xs text-text-sub dark:text-gray-400 mb-1">
              ({draftCount} draft, {confirmedCount} dikonfirmasi)
            </span>
          </div>
        </Card>

        <Card variant="bordered" padding="md">
          <div className="flex justify-between items-start">
            <p className="text-sm font-medium text-text-sub dark:text-gray-400">Selesai</p>
            <span className="p-1.5 rounded-lg bg-green-600/10">
              <Icon name="check_circle" size={20} className="text-green-600" />
            </span>
          </div>
          <div className="flex items-end gap-2 mt-2">
            <p className="text-2xl font-bold text-text-main dark:text-white tracking-tight">
              {completedCount}
            </p>
          </div>
        </Card>

        <Card variant="bordered" padding="md">
          <div className="flex justify-between items-start">
            <p className="text-sm font-medium text-text-sub dark:text-gray-400">Dibatalkan</p>
            <span className="p-1.5 rounded-lg bg-red-500/10">
              <Icon name="cancel" size={20} className="text-red-500" />
            </span>
          </div>
          <div className="flex items-end gap-2 mt-2">
            <p className="text-2xl font-bold text-text-main dark:text-white tracking-tight">
              {cancelledCount}
            </p>
          </div>
        </Card>
      </div>

      {/* Filters */}
      <Card variant="bordered" padding="md">
        <div className="flex flex-col gap-4">
          {/* Filter Row */}
          <div className="flex flex-col lg:flex-row gap-3 items-stretch">
            {/* Search */}
            <div className="relative flex-1 min-w-0">
              <Icon
                name="search"
                size={20}
                className="absolute left-4 top-1/2 -translate-y-1/2 text-text-sub dark:text-gray-400 pointer-events-none"
              />
              <input
                type="text"
                placeholder="Cari ID Pesanan atau Pelanggan..."
                value={searchInput}
                onChange={e => setSearchInput(e.target.value)}
                onKeyDown={handleSearchKeyDown}
                className="w-full h-12 pl-12 pr-4 rounded-lg border border-border-light dark:border-border-dark bg-white dark:bg-card-dark text-text-main dark:text-white text-sm focus:outline-none focus:ring-2 focus:ring-primary/20 focus:border-primary"
              />
            </div>

            {/* Filter Dropdowns */}
            <div className="flex flex-wrap lg:flex-nowrap gap-3">
              <Dropdown
                value={status}
                onChange={handleStatusFilter}
                options={[
                  { value: '', label: 'Semua Status' },
                  { value: 'draft', label: 'Draft' },
                  { value: 'confirmed', label: 'Dikonfirmasi' },
                  { value: 'completed', label: 'Selesai' },
                  { value: 'cancelled', label: 'Dibatalkan' },
                  { value: 'voided', label: 'Void' },
                ]}
                leftIcon={<Icon name="local_shipping" size={20} />}
                className="w-full sm:w-40"
              />
              <Dropdown
                value={paymentMethod}
                onChange={handlePaymentMethodFilter}
                options={[
                  { value: '', label: 'Semua Pembayaran' },
                  { value: 'cash', label: 'Cash' },
                  { value: 'debit_card', label: 'Debit Card' },
                  { value: 'credit_card', label: 'Credit Card' },
                  { value: 'e_wallet', label: 'E-Wallet' },
                  { value: 'bank_transfer', label: 'Transfer Bank' },
                  { value: 'qris', label: 'QRIS' },
                ]}
                leftIcon={<Icon name="payment" size={20} />}
                className="w-full sm:w-44"
              />
              <DateRangePicker
                startDate={startDate}
                endDate={endDate}
                onChange={handleDateRangeChange}
                className="w-full sm:w-64"
              />
            </div>
          </div>

          {/* Active Filters & Actions */}
          {hasActiveFilters && (
            <div className="flex items-center gap-2 flex-wrap">
              <span className="text-sm text-text-sub dark:text-gray-400">Filter Aktif:</span>
              {status && (
                <span className="inline-flex items-center gap-1 px-2.5 py-1 rounded-full bg-primary/10 text-primary text-xs font-medium">
                  Status: {getStatusLabel(status)}
                  <button onClick={() => setStatus('')} className="hover:bg-primary/20 rounded-full p-0.5">
                    <Icon name="close" size={14} />
                  </button>
                </span>
              )}
              {paymentMethod && (
                <span className="inline-flex items-center gap-1 px-2.5 py-1 rounded-full bg-primary/10 text-primary text-xs font-medium">
                  Pembayaran: {paymentMethod}
                  <button onClick={() => setPaymentMethod('')} className="hover:bg-primary/20 rounded-full p-0.5">
                    <Icon name="close" size={14} />
                  </button>
                </span>
              )}
              {startDate && (
                <span className="inline-flex items-center gap-1 px-2.5 py-1 rounded-full bg-primary/10 text-primary text-xs font-medium">
                  Tanggal: {formatDate(startDate)} - {formatDate(endDate)}
                  <button onClick={() => { setStartDate(''); setEndDate(''); }} className="hover:bg-primary/20 rounded-full p-0.5">
                    <Icon name="close" size={14} />
                  </button>
                </span>
              )}
              <button
                onClick={handleClearFilters}
                className="text-sm text-red-600 dark:text-red-400 hover:underline ml-2"
              >
                Hapus Semua Filter
              </button>
            </div>
          )}
        </div>
      </Card>

      {/* Orders Table */}
      <Card variant="bordered" padding="none">
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead className="bg-background-light dark:bg-background-dark border-b border-border-light dark:border-border-dark">
              <tr>
                <th className="px-4 py-4 text-left text-sm font-semibold text-text-main dark:text-white">
                  ID Pesanan
                </th>
                <th className="px-4 py-4 text-left text-sm font-semibold text-text-main dark:text-white">
                  Pelanggan
                </th>
                <th className="px-4 py-4 text-left text-sm font-semibold text-text-main dark:text-white">
                  Produk
                </th>
                <th className="px-4 py-4 text-left text-sm font-semibold text-text-main dark:text-white">
                  Total
                </th>
                <th className="px-4 py-4 text-left text-sm font-semibold text-text-main dark:text-white">
                  Status
                </th>
                <th className="px-4 py-4 text-left text-sm font-semibold text-text-main dark:text-white">
                  Pembayaran
                </th>
                <th className="px-4 py-4 text-left text-sm font-semibold text-text-main dark:text-white">
                  Tanggal
                </th>
                <th className="px-4 py-4 text-left text-sm font-semibold text-text-main dark:text-white">
                  Aksi
                </th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border-light dark:divide-border-dark">
              {isLoading ? (
                <tr>
                  <td colSpan={8} className="px-4 py-12 text-center">
                    <Icon name="progress_activity" size={32} className="animate-spin text-primary mx-auto" />
                  </td>
                </tr>
              ) : data?.data.length === 0 ? (
                <tr>
                  <td colSpan={8} className="px-4 py-12 text-center">
                    <div className="flex flex-col items-center gap-3">
                      <div className="p-4 bg-background-light dark:bg-background-dark rounded-full">
                        <Icon name="shopping_cart" size={32} className="text-text-sub dark:text-gray-400" />
                      </div>
                      <div>
                        <p className="text-text-main dark:text-white font-medium">Tidak ada pesanan</p>
                        <p className="text-sm text-text-sub dark:text-gray-400 mt-1">
                          {hasActiveFilters ? 'Coba ubah filter pencarian Anda' : 'Buat pesanan pertama Anda'}
                        </p>
                      </div>
                    </div>
                  </td>
                </tr>
              ) : (
                data?.data.map((order) => (
                  <tr
                    key={order.id}
                    className="hover:bg-background-light dark:hover:bg-background-dark transition-colors cursor-pointer"
                    onClick={() => handleViewOrder(order)}
                  >
                    <td className="px-4 py-4">
                      <div className="font-medium text-text-main dark:text-white">
                        {order.order_no}
                      </div>
                    </td>
                    <td className="px-4 py-4">
                      <div className="flex items-center gap-3">
                        <div className="flex-shrink-0">
                          <div className="w-8 h-8 rounded-full bg-primary/10 flex items-center justify-center">
                            <Icon name="person" size={18} className="text-primary" />
                          </div>
                        </div>
                        <div className="min-w-0 flex-1">
                          <p className="text-sm font-medium text-text-main dark:text-white truncate">
                            {order.customer_name || 'Walk-in Customer'}
                          </p>
                        </div>
                      </div>
                    </td>
                    <td className="px-4 py-4">
                      <div className="text-sm text-text-main dark:text-white">
                        {order.items?.length || 0} item
                      </div>
                    </td>
                    <td className="px-4 py-4">
                      <div className="text-sm font-semibold text-text-main dark:text-white">
                        {formatCurrency(order.grand_total)}
                      </div>
                      {order.balance_due > 0 && (
                        <div className="text-xs text-red-600 dark:text-red-400">
                          Sisa: {formatCurrency(order.balance_due)}
                        </div>
                      )}
                    </td>
                    <td className="px-4 py-4">
                      <span className={`inline-flex items-center px-2.5 py-1 rounded-full text-xs font-medium ${getStatusBadge(order.status)}`}>
                        {getStatusLabel(order.status)}
                      </span>
                    </td>
                    <td className="px-4 py-4">
                      <span className={`inline-flex items-center px-2.5 py-1 rounded-full text-xs font-medium ${getPaymentStatusBadge(order.payment_status)}`}>
                        {getPaymentStatusLabel(order.payment_status)}
                      </span>
                    </td>
                    <td className="px-4 py-4">
                      <div className="text-sm text-text-sub dark:text-gray-400">
                        {formatDate(order.created_at)}
                      </div>
                    </td>
                    <td className="px-4 py-4">
                      <div className="flex items-center gap-1">
                        <button
                          className="p-2 hover:bg-primary/10 rounded-lg transition-colors"
                          title="Lihat Detail"
                          onClick={(e) => {
                            e.stopPropagation();
                            handleViewOrder(order);
                          }}
                        >
                          <Icon name="visibility" size={18} className="text-primary" />
                        </button>
                      </div>
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>

        {/* Pagination */}
        {data && data.totalPages > 1 && (
          <div className="flex items-center justify-between px-6 py-4 border-t border-border-light dark:border-border-dark">
            <div className="flex items-center gap-2">
              <span className="text-sm text-text-sub dark:text-gray-400">
                Menampilkan {((page - 1) * 10) + 1}-{Math.min(page * 10, data.total)} dari {data.total} pesanan
              </span>
            </div>
            <div className="flex items-center gap-2">
              <Button
                variant="outline"
                size="sm"
                onClick={() => setPage(p => Math.max(1, p - 1))}
                disabled={page === 1}
                leftIcon={<Icon name="chevron_left" size={18} />}
              >
                Sebelumnya
              </Button>
              <div className="flex items-center gap-1">
                {Array.from({ length: Math.min(5, data.totalPages) }, (_, i) => {
                  let pageNum = i + 1;
                  if (data.totalPages > 5) {
                    if (page > 3) pageNum = page - 2 + i;
                    if (page > data.totalPages - 3) pageNum = data.totalPages - 4 + i;
                  }
                  return (
                    <button
                      key={pageNum}
                      onClick={() => setPage(pageNum)}
                      className={`px-3 py-1.5 rounded-lg text-sm font-medium transition-colors ${
                        page === pageNum
                          ? 'bg-primary text-white'
                          : 'text-text-sub dark:text-gray-400 hover:bg-background-light dark:hover:bg-background-dark'
                      }`}
                    >
                      {pageNum}
                    </button>
                  );
                })}
              </div>
              <Button
                variant="outline"
                size="sm"
                onClick={() => setPage(p => Math.min(data.totalPages, p + 1))}
                disabled={page === data.totalPages}
                rightIcon={<Icon name="chevron_right" size={18} />}
              >
                Selanjutnya
              </Button>
            </div>
          </div>
        )}
      </Card>

      {/* Order Form Modal */}
      <OrderFormModal
        isOpen={isFormModalOpen}
        onClose={() => setIsFormModalOpen(false)}
        onSubmit={handleFormSubmit}
        isLoading={createOrder.isPending}
      />

      {/* Order Detail Modal */}
      <OrderDetailModal
        isOpen={isDetailModalOpen}
        onClose={handleDetailModalClose}
        order={selectedOrder}
      />
    </div>
  );
}

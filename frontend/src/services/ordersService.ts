import { api } from './api';
import type {
  Order,
  CreateOrderRequest,
  UpdateOrderRequest,
  AddPaymentRequest,
  CompleteOrderRequest,
  CancelOrderRequest,
  VoidOrderRequest,
  RefundPaymentRequest,
  OrderListParams,
} from '@/types/order';
import type { ApiResponse, PaginatedApiResponse } from '@/types/api';

export const ordersService = {
  // List orders with filters
  getOrders: async (params: OrderListParams = {}) => {
    const { data } = await api.get<PaginatedApiResponse<Order[]>>('/orders', { params });
    return {
      data: data.data,
      total: data.meta?.pagination?.total_records || 0,
      page: data.meta?.pagination?.current_page || 1,
      per_page: data.meta?.pagination?.per_page || 20,
      totalPages: data.meta?.pagination?.total_pages || 1,
    };
  },

  // Get single order
  getOrder: async (id: number) => {
    const { data } = await api.get<ApiResponse<Order>>(`/orders/${id}`);
    return data.data;
  },

  // Create new order (as draft)
  createOrder: async (orderData: CreateOrderRequest) => {
    const { data } = await api.post<ApiResponse<Order>>('/orders', orderData);
    return data.data;
  },

  // Update draft order
  updateOrder: async (id: number, orderData: UpdateOrderRequest) => {
    const { data } = await api.put<ApiResponse<Order>>(`/orders/${id}`, orderData);
    return data.data;
  },

  // Confirm order (draft -> confirmed)
  confirmOrder: async (id: number) => {
    const { data } = await api.post<ApiResponse<Order>>(`/orders/${id}/confirm`);
    return data.data;
  },

  // Add payment(s) to order
  addPayment: async (id: number, paymentData: AddPaymentRequest) => {
    const { data } = await api.post<ApiResponse<Order>>(`/orders/${id}/payments`, paymentData);
    return data.data;
  },

  // Complete order
  completeOrder: async (id: number, completeData?: CompleteOrderRequest) => {
    const { data } = await api.post<ApiResponse<Order>>(`/orders/${id}/complete`, completeData || {});
    return data.data;
  },

  // Cancel order (for unpaid orders)
  cancelOrder: async (id: number, cancelData: CancelOrderRequest) => {
    const { data } = await api.post<ApiResponse<Order>>(`/orders/${id}/cancel`, cancelData);
    return data.data;
  },

  // Void order (for paid orders)
  voidOrder: async (id: number, voidData: VoidOrderRequest) => {
    const { data } = await api.post<ApiResponse<Order>>(`/orders/${id}/void`, voidData);
    return data.data;
  },

  // Refund a payment
  refundPayment: async (id: number, refundData: RefundPaymentRequest) => {
    const { data } = await api.post<ApiResponse<Order>>(`/orders/${id}/refund`, refundData);
    return data.data;
  },
};

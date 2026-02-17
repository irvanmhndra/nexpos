// ============== Order Item ==============

export interface OrderItem {
  id: number;
  product_id: number | null;
  product_variant_id: number | null;
  sku: string;
  product_name: string;
  variant_name: string | null;
  variant_attributes: Record<string, unknown> | null;
  unit_price: number;
  unit_cost: number;
  quantity: number;
  discount_amount: number;
  tax_amount: number;
  subtotal: number;
  cogs_amount: number;
}

export interface OrderItemInput {
  product_variant_id: number;
  quantity: number;
  discount_amount?: number;
  notes?: string;
}

// ============== Payment ==============

export type PaymentMethod = 'cash' | 'debit_card' | 'credit_card' | 'e_wallet' | 'bank_transfer' | 'qris';
export type PaymentStatus = 'pending' | 'completed' | 'refunded' | 'partially_refunded' | 'failed';

export interface Payment {
  id: number;
  method: PaymentMethod;
  amount: number;
  reference_no: string | null;
  status: PaymentStatus;
  refunded_amount: number;
  refunded_at: string | null;
  refund_reason: string | null;
  paid_at: string;
}

export interface PaymentInput {
  method: PaymentMethod;
  amount: number;
  reference_no?: string;
}

// ============== Order Status Types ==============

export type OrderStatus = 'draft' | 'confirmed' | 'completed' | 'cancelled' | 'voided';
export type OrderPaymentStatus = 'unpaid' | 'partial' | 'paid' | 'refunded';
export type FulfillmentType = 'counter' | 'delivery';
export type FulfillmentStatus = 'pending' | 'processing' | 'shipped' | 'delivered';

// ============== Order ==============

export interface Order {
  id: number;
  order_no: string;
  customer_id: number | null;
  customer_name?: string;
  cashier_id: number;
  cashier_name?: string;
  status: OrderStatus;
  total_amount: number;
  total_discount: number;
  total_tax: number;
  grand_total: number;
  notes: string | null;

  // Lifecycle fields
  payment_status: OrderPaymentStatus;
  fulfillment_type: FulfillmentType;
  fulfillment_status: FulfillmentStatus | null;
  shipping_address: string | null;
  confirmed_at: string | null;
  paid_at: string | null;
  completed_at: string | null;
  cancelled_at: string | null;
  voided_at: string | null;
  cancel_reason: string | null;
  void_reason: string | null;

  // Offline sync
  offline_id: string | null;
  synced_at: string | null;

  // Computed fields
  paid_amount: number;
  balance_due: number;
  refunded_total: number;

  items: OrderItem[];
  payments: Payment[];
  created_at: string;
  updated_at: string;
}

// ============== Request Types ==============

export interface CreateOrderRequest {
  customer_id?: number | null;
  fulfillment_type?: FulfillmentType;
  shipping_address?: string;
  items: OrderItemInput[];
  payments?: PaymentInput[];
  notes?: string;
  offline_id?: string;
}

export interface UpdateOrderRequest {
  customer_id?: number | null;
  fulfillment_type?: FulfillmentType;
  shipping_address?: string;
  items?: OrderItemInput[];
  notes?: string;
}

export interface AddPaymentRequest {
  payments: PaymentInput[];
}

export interface CompleteOrderRequest {
  fulfillment_status?: FulfillmentStatus;
}

export interface CancelOrderRequest {
  reason: string;
}

export interface VoidOrderRequest {
  reason: string;
}

export interface RefundPaymentRequest {
  payment_id: number;
  amount: number;
  refund_reason: string;
}

// ============== List Params ==============

export interface OrderListParams {
  page?: number;
  per_page?: number;
  search?: string;
  status?: OrderStatus;
  payment_status?: OrderPaymentStatus;
  fulfillment_type?: FulfillmentType;
  fulfillment_status?: FulfillmentStatus;
  customer_id?: number;
  payment_method?: PaymentMethod;
  date_from?: string;
  date_to?: string;
}

// ============== List Responses ==============

export interface OrderListResponse {
  orders: Order[];
  pagination: {
    total_records: number;
    total_pages: number;
    current_page: number;
    per_page: number;
    next_page: number | null;
    prev_page: number | null;
  };
}

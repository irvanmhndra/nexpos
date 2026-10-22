// ============== Report Summary ==============

export interface ReportSummary {
  total_revenue: number;
  total_orders: number;
  total_items_sold: number;
  avg_order_value: number;
  total_customers: number;
  new_customers: number;
  completed_orders: number;
  cancelled_orders: number;
  pending_orders: number;
  total_discount: number;
  total_tax: number;
  gross_profit: number;
  gross_profit_margin: number;
}

// ============== Sales Trend ==============

export interface SalesTrendItem {
  date: string;
  sales: number;
  orders: number;
  items: number;
}

export interface SalesTrendResponse {
  data: SalesTrendItem[];
}

// ============== Top Products ==============

export interface TopProductItem {
  product_id: number;
  product_name: string;
  total_sold: number;
  total_amount: number;
}

export interface TopProductsResponse {
  data: TopProductItem[];
}

// ============== Category Revenue ==============

export interface CategoryRevenueItem {
  category_id: number;
  category_name: string;
  total_amount: number;
  order_count: number;
}

export interface CategoryRevenueResponse {
  data: CategoryRevenueItem[];
}

// ============== Payment Methods ==============

export interface PaymentMethodItem {
  method: string;
  total_amount: number;
  count: number;
}

export interface PaymentMethodResponse {
  data: PaymentMethodItem[];
}

// ============== Hourly Sales ==============

export interface HourlySalesItem {
  hour: number;
  total_amount: number;
  order_count: number;
}

export interface HourlySalesResponse {
  data: HourlySalesItem[];
}

// ============== Request Params ==============

export interface ReportParams {
  date_from: string;
  date_to: string;
}

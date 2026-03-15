export interface CompanySettings {
  id: number;
  company_id: number;

  // Tax
  tax_enabled: boolean;
  tax_rate: number;
  tax_inclusive: boolean;

  // Rounding
  rounding_enabled: boolean;
  rounding_amount: number;

  // Order
  auto_complete_counter_orders: boolean;
  require_customer_for_delivery: boolean;

  // Receipt
  receipt_header: string | null;
  receipt_footer: string | null;
  show_tax_on_receipt: boolean;

  // Offline
  offline_mode_enabled: boolean;
  max_offline_days: number;

  created_at: string;
  updated_at: string;
}

export interface UpdateCompanySettingsRequest {
  tax_enabled?: boolean;
  tax_rate?: number;
  tax_inclusive?: boolean;
  rounding_enabled?: boolean;
  rounding_amount?: number;
  auto_complete_counter_orders?: boolean;
  require_customer_for_delivery?: boolean;
  receipt_header?: string | null;
  receipt_footer?: string | null;
  show_tax_on_receipt?: boolean;
  offline_mode_enabled?: boolean;
  max_offline_days?: number;
}

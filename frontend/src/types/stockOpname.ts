export type StockOpnameStatus = 'in_progress' | 'completed' | 'cancelled';

export interface StockOpnameItem {
  id: number;
  stock_opname_id: number;
  product_variant_id: number;
  sku: string;
  product_name: string;
  variant_name: string;
  system_stock: number;
  counted_stock: number | null;
  variance_qty: number;
  unit_cost: number;
  variance_value: number;
  notes: string | null;
  counted_at: string | null;
}

export interface StockOpname {
  id: number;
  branch_id: number;
  opname_number: string;
  status: StockOpnameStatus;
  notes: string | null;
  total_items: number;
  counted_items: number;
  total_variance_qty: number;
  total_variance_value: number;
  started_by: number | null;
  completed_by: number | null;
  started_at: string;
  completed_at: string | null;
  cancelled_at: string | null;
  items?: StockOpnameItem[];
  created_at: string;
  updated_at: string;
}

export interface CreateStockOpnameRequest {
  branch_id: number;
  notes?: string;
  category_id?: number;
}

export interface UpdateOpnameItemRequest {
  counted_stock: number;
  notes?: string;
}

export interface BulkUpdateOpnameItem {
  item_id: number;
  counted_stock: number;
  notes?: string;
}

export interface BulkUpdateOpnameItemsRequest {
  items: BulkUpdateOpnameItem[];
}

export interface CompleteStockOpnameRequest {
  notes?: string;
}

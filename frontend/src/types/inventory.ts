// ============== Stock Status & Movement Types ==============

export type StockStatus = 'normal' | 'low' | 'out_of_stock';
export type StockMovementType = 'IN' | 'OUT' | 'ADJUST';

// ============== Inventory Item ==============

export interface InventoryItem {
  product_id: number;
  product_variant_id: number;
  product_name: string;
  variant_name: string;
  sku: string;
  category_name: string;
  image_data: string | null;
  image_url: string | null;
  branch_id: number;
  branch_name: string;
  current_stock: number;
  min_stock: number;
  stock_value: number;
  status: StockStatus;
  has_variants: boolean;
}

// ============== Stock Movement ==============

export interface StockMovement {
  id: string;
  product_id: string;
  variant_id: string;
  product_name: string;
  variant_name: string;
  sku: string;
  type: StockMovementType;
  quantity: number;
  stock_before: number;
  stock_after: number;
  reference: string | null;
  notes: string | null;
  created_at: string;
  created_by: string;
}

// ============== Stats ==============

export interface InventoryStats {
  total_sku: number;
  low_stock: number;
  out_of_stock: number;
  total_stock_value: number;
}

export interface MovementStats {
  total_movements: number;
  stock_in: number;
  stock_out: number;
  adjustments: number;
}

// ============== Requests ==============

export interface AdjustStockRequest {
  variant_id: number;
  branch_id: number;
  type: StockMovementType;
  quantity: number;
  note?: string;
}

export interface UpdateMinStockRequest {
  min_stock: number;
}

// ============== Query Params ==============

export interface GetInventoryParams {
  page?: number;
  per_page?: number;
  search?: string;
  category?: string;
  status?: StockStatus | '';
  branch_id?: number;
}

export interface GetMovementsParams {
  page?: number;
  per_page?: number;
  search?: string;
  type?: StockMovementType | '';
  start_date?: string;
  end_date?: string;
  branch_id?: number;
}

import type { Supplier } from './supplier';

export type POStatus = 'draft' | 'ordered' | 'partial' | 'received' | 'cancelled';

export interface PurchaseOrderItem {
  id: number;
  purchase_order_id: number;
  product_variant_id: number | null;
  sku: string;
  variant_name: string;
  quantity: number;
  received_quantity: number;
  unit_cost: number;
  subtotal: number;
}

export interface PurchaseOrder {
  id: number;
  company_id: number;
  branch_id: number;
  supplier_id: number;
  po_number: string;
  status: POStatus;
  notes: string | null;
  total_amount: number;
  ordered_at: string | null;
  received_at: string | null;
  cancelled_at: string | null;
  items?: PurchaseOrderItem[];
  supplier?: Supplier;
  created_at: string;
  updated_at: string;
}

export interface CreatePOItemInput {
  product_variant_id?: number;
  sku: string;
  variant_name: string;
  quantity: number;
  unit_cost: number;
}

export interface CreatePORequest {
  supplier_id: number;
  branch_id: number;
  notes?: string;
  items: CreatePOItemInput[];
}

export interface ReceivePOItemInput {
  purchase_order_item_id: number;
  received_quantity: number;
}

export interface ReceivePORequest {
  items: ReceivePOItemInput[];
  notes?: string;
}

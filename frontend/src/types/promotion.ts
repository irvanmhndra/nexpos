export type PromotionType = 'discount' | 'bundle' | 'conditional';
export type DiscountType = 'percentage' | 'fixed';
// Derived (server-computed) lifecycle status.
export type PromotionStatus = 'inactive' | 'scheduled' | 'active' | 'expired';

export interface Promotion {
  id: number;
  code: string;
  name: string;
  description: string | null;
  type: PromotionType;
  discount_type: DiscountType | null;
  discount_value: number | null;
  min_purchase: number | null;
  max_discount: number | null;
  start_at: string;
  end_at: string | null;
  priority: number;
  is_active: boolean;
  status: PromotionStatus;
  created_at: string;
  updated_at: string;
}

export interface CreatePromotionRequest {
  code: string;
  name: string;
  description?: string | null;
  type: PromotionType;
  discount_type?: DiscountType | null;
  discount_value?: number | null;
  min_purchase?: number | null;
  max_discount?: number | null;
  start_at: string; // ISO 8601
  end_at?: string | null;
  priority: number;
  is_active: boolean;
}

export type UpdatePromotionRequest = CreatePromotionRequest;

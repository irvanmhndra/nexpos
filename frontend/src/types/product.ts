// ============== Product Category ==============

export interface ProductCategory {
  id: number;
  code: string;
  name: string;
  parent_id: number | null;
  parent_name?: string;
  sort_order: number;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface CreateProductCategoryRequest {
  code: string;
  name: string;
  parent_id?: number | null;
  sort_order?: number;
  is_active?: boolean;
}

export interface UpdateProductCategoryRequest {
  code: string;
  name: string;
  parent_id?: number | null;
  sort_order?: number;
  is_active?: boolean;
}

// ============== Product Variant ==============

export interface ProductVariant {
  id: number;
  sku: string;
  name: string;
  attributes: Record<string, unknown> | null;
  price: number;
  standard_cost: number;
  last_purchase_cost: number;
  is_default: boolean;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface ProductVariantInput {
  id?: number;
  sku: string;
  name: string;
  attributes?: Record<string, unknown>;
  price: number;
  standard_cost?: number;
  last_purchase_cost?: number;
  is_default?: boolean;
  is_active?: boolean;
}

// ============== Product ==============

export interface Product {
  id: number;
  name: string;
  category_id: number | null;
  category_name?: string;
  description: string | null;
  image_data: string | null;
  is_active: boolean;
  variants: ProductVariant[];
  created_at: string;
  updated_at: string;
}

export interface CreateProductRequest {
  name: string;
  category_id?: number | null;
  description?: string;
  image_data?: string;
  is_active?: boolean;
  variants: ProductVariantInput[];
}

export interface UpdateProductRequest {
  name: string;
  category_id?: number | null;
  description?: string;
  image_data?: string;
  is_active?: boolean;
  variants: ProductVariantInput[];
}

// ============== List Responses ==============

export interface ProductCategoryListResponse {
  categories: ProductCategory[];
  pagination: {
    total_records: number;
    total_pages: number;
    current_page: number;
    per_page: number;
    next_page: number | null;
    prev_page: number | null;
  };
}

export interface ProductListResponse {
  products: Product[];
  pagination: {
    total_records: number;
    total_pages: number;
    current_page: number;
    per_page: number;
    next_page: number | null;
    prev_page: number | null;
  };
}

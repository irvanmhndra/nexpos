export interface Customer {
  id: number;
  code: string;
  name: string;
  phone: string | null;
  email: string | null;
  is_member: boolean;
  created_at: string;
  updated_at: string;
}

export interface CreateCustomerRequest {
  code: string;
  name: string;
  phone?: string;
  email?: string;
  is_member: boolean;
}

export interface UpdateCustomerRequest {
  code: string;
  name: string;
  phone?: string;
  email?: string;
  is_member: boolean;
}

export interface CustomerListResponse {
  customers: Customer[];
  pagination: {
    total_records: number;
    total_pages: number;
    current_page: number;
    per_page: number;
    next_page: number | null;
    prev_page: number | null;
  };
}

export interface Branch {
  id: number;
  code: string;
  name: string;
  address: string | null;
  phone: string | null;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface CreateBranchRequest {
  code: string;
  name: string;
  address?: string | null;
  phone?: string | null;
  is_active: boolean;
}

export type UpdateBranchRequest = CreateBranchRequest;

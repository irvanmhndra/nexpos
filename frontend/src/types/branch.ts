export type BranchStatus = 'active' | 'inactive' | 'maintenance';

export interface Branch {
  id: string;
  name: string;
  code: string;
  address: string;
  city: string;
  region: string;
  phone: string;
  manager: string;
  status: BranchStatus;
  created_at: string;
}

export interface CreateBranchRequest {
  name: string;
  address: string;
  city: string;
  region: string;
  phone: string;
  manager?: string;
  status: BranchStatus;
}

export type UpdateBranchRequest = Partial<CreateBranchRequest>;

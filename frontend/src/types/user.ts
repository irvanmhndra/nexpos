// ============== Role ==============

export interface Role {
  id: number;
  code: string;
  name: string;
}

// ============== User ==============

export type UserStatus = 'active' | 'inactive';

export interface User {
  id: number;
  company_id: number;
  role_id: number | null;
  email: string;
  name: string;
  status: UserStatus;
  role?: Role;
  created_at: string;
  updated_at: string;
}

// ============== Request Types ==============

export interface CreateUserRequest {
  email: string;
  password: string;
  name: string;
  role_id?: number;
}

export interface UpdateUserRequest {
  email?: string;
  name?: string;
  role_id?: number;
}

export interface UpdateUserStatusRequest {
  status: UserStatus;
}

export interface ListUsersParams {
  page?: number;
  limit?: number;
  search?: string;
  status?: UserStatus;
}

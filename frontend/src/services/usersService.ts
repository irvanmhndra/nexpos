import { api } from './api';
import type {
  User,
  CreateUserRequest,
  UpdateUserRequest,
  UpdateUserStatusRequest,
  ListUsersParams,
} from '@/types/user';
import type { ApiResponse, PaginatedApiResponse } from '@/types/api';

export const usersService = {
  // List users with filters
  getUsers: async (params: ListUsersParams = {}) => {
    const { data } = await api.get<PaginatedApiResponse<User[]>>('/users', { params });
    return {
      data: data.data,
      total: data.meta?.pagination?.total_records || 0,
      page: data.meta?.pagination?.current_page || 1,
      per_page: data.meta?.pagination?.per_page || 10,
      totalPages: data.meta?.pagination?.total_pages || 1,
    };
  },

  // Get single user
  getUser: async (id: number) => {
    const { data } = await api.get<ApiResponse<User>>(`/users/${id}`);
    return data.data;
  },

  // Create new user
  createUser: async (userData: CreateUserRequest) => {
    const { data } = await api.post<ApiResponse<User>>('/users', userData);
    return data.data;
  },

  // Update user
  updateUser: async (id: number, userData: UpdateUserRequest) => {
    const { data } = await api.put<ApiResponse<User>>(`/users/${id}`, userData);
    return data.data;
  },

  // Update user status
  updateUserStatus: async (id: number, statusData: UpdateUserStatusRequest) => {
    const { data } = await api.patch<ApiResponse<null>>(`/users/${id}/status`, statusData);
    return data.data;
  },

  // Delete user
  deleteUser: async (id: number) => {
    await api.delete(`/users/${id}`);
  },
};

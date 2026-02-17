export interface ApiResponse<T> {
  success: boolean;
  message: string;
  data: T;
  error_code?: string;
  errors?: unknown;
  meta?: {
    server_time: string;
  };
}

export interface PaginatedApiResponse<T> {
  success: boolean;
  message: string;
  data: T;
  meta?: {
    pagination?: {
      total_records: number;
      total_pages: number;
      current_page: number;
      per_page: number;
      count: number;
      next_page: number | null;
      prev_page: number | null;
    };
    server_time: string;
  };
}

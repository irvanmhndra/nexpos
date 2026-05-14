import { api } from './api';
import type {
  DailySettlement,
  DailySettlementItem,
  DailySettlementReport,
  CreateDailySettlementRequest,
  UpdateSettlementItemRequest,
  BulkUpdateSettlementItemsRequest,
  FinalizeDailySettlementRequest,
} from '@/types/dailySettlement';

interface ApiResponse<T> {
  success: boolean;
  message: string;
  data: T;
  meta?: {
    pagination?: {
      total_records: number;
      total_pages: number;
      current_page: number;
      per_page: number;
      next_page: number | null;
      prev_page: number | null;
    };
  };
}

interface ListParams {
  page?: number;
  per_page?: number;
  status?: string;
  branch_id?: number;
  date_from?: string;
  date_to?: string;
}

export const dailySettlementService = {
  list: async (params: ListParams = {}) => {
    const { data } = await api.get<ApiResponse<DailySettlement[]>>('/daily-settlements', { params });
    return {
      data: data.data,
      total: data.meta?.pagination?.total_records || 0,
      page: data.meta?.pagination?.current_page || 1,
      perPage: data.meta?.pagination?.per_page || 10,
      totalPages: data.meta?.pagination?.total_pages || 1,
    };
  },

  get: async (id: number) => {
    const { data } = await api.get<ApiResponse<DailySettlement>>(`/daily-settlements/${id}`);
    return data.data;
  },

  report: async (branchId: number, date: string) => {
    const { data } = await api.get<ApiResponse<DailySettlementReport>>('/daily-settlements/report', {
      params: { branch_id: branchId, date },
    });
    return data.data;
  },

  create: async (request: CreateDailySettlementRequest) => {
    const { data } = await api.post<ApiResponse<DailySettlement>>('/daily-settlements', request);
    return data.data;
  },

  updateItem: async (settlementId: number, itemId: number, request: UpdateSettlementItemRequest) => {
    const { data } = await api.patch<ApiResponse<DailySettlementItem>>(
      `/daily-settlements/${settlementId}/items/${itemId}`,
      request,
    );
    return data.data;
  },

  bulkUpdateItems: async (settlementId: number, request: BulkUpdateSettlementItemsRequest) => {
    const { data } = await api.patch<ApiResponse<DailySettlement>>(
      `/daily-settlements/${settlementId}/items`,
      request,
    );
    return data.data;
  },

  finalize: async (settlementId: number, request: FinalizeDailySettlementRequest = {}) => {
    const { data } = await api.post<ApiResponse<DailySettlement>>(
      `/daily-settlements/${settlementId}/finalize`,
      request,
    );
    return data.data;
  },
};

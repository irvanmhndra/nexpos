import { api } from './api';
import type { CompanySettings, UpdateCompanySettingsRequest } from '@/types/companySettings';

interface ApiResponse<T> {
  success: boolean;
  message: string;
  data: T;
}

export const companySettingsService = {
  getSettings: async () => {
    const { data } = await api.get<ApiResponse<CompanySettings>>('/company-settings');
    return data.data;
  },

  updateSettings: async (req: UpdateCompanySettingsRequest) => {
    const { data } = await api.put<ApiResponse<CompanySettings>>('/company-settings', req);
    return data.data;
  },
};

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { companySettingsService } from '@/services/companySettingsService';
import type { UpdateCompanySettingsRequest } from '@/types/companySettings';

export function useCompanySettings() {
  return useQuery({
    queryKey: ['company-settings'],
    queryFn: companySettingsService.getSettings,
    staleTime: 5 * 60 * 1000,
  });
}

export function useUpdateCompanySettings() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (req: UpdateCompanySettingsRequest) => companySettingsService.updateSettings(req),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['company-settings'] });
    },
  });
}

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { promotionsService } from '@/services/promotionsService';
import type { CreatePromotionRequest, UpdatePromotionRequest } from '@/types/promotion';

interface UsePromotionsParams {
  page?: number;
  per_page?: number;
  search?: string;
  is_active?: boolean;
  type?: string;
  status?: string;
}

export function usePromotions(params: UsePromotionsParams = {}) {
  return useQuery({
    queryKey: ['promotions', params],
    queryFn: () => promotionsService.getPromotions(params),
  });
}

export function useCreatePromotion() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (promotion: CreatePromotionRequest) => promotionsService.createPromotion(promotion),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['promotions'] });
    },
  });
}

export function useUpdatePromotion() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, promotion }: { id: number; promotion: UpdatePromotionRequest }) =>
      promotionsService.updatePromotion(id, promotion),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['promotions'] });
    },
  });
}

export function useDeletePromotion() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: number) => promotionsService.deletePromotion(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['promotions'] });
    },
  });
}

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { productCategoriesService } from '@/services/productCategoriesService';
import type { CreateProductCategoryRequest, UpdateProductCategoryRequest } from '@/types/product';

interface UseProductCategoriesParams {
  page?: number;
  per_page?: number;
  search?: string;
  is_active?: boolean;
}

export function useProductCategories(params: UseProductCategoriesParams = {}) {
  return useQuery({
    queryKey: ['product-categories', params],
    queryFn: () => productCategoriesService.getCategories(params),
  });
}

export function useAllProductCategories() {
  return useQuery({
    queryKey: ['product-categories', 'all'],
    queryFn: () => productCategoriesService.getAllCategories(),
  });
}

export function useProductCategory(id: number) {
  return useQuery({
    queryKey: ['product-categories', id],
    queryFn: () => productCategoriesService.getCategory(id),
    enabled: !!id,
  });
}

export function useCreateProductCategory() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (category: CreateProductCategoryRequest) => productCategoriesService.createCategory(category),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['product-categories'] });
    },
  });
}

export function useUpdateProductCategory() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, category }: { id: number; category: UpdateProductCategoryRequest }) =>
      productCategoriesService.updateCategory(id, category),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['product-categories'] });
    },
  });
}

export function useDeleteProductCategory() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: number) => productCategoriesService.deleteCategory(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['product-categories'] });
    },
  });
}

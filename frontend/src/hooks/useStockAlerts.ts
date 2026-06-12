import { useQuery } from '@tanstack/react-query';
import { inventoryService } from '@/services/inventoryService';
import { useBranch } from '@/contexts/BranchContext';

/**
 * Fetches the actual out-of-stock and low-stock items for the active branch,
 * used to populate the notification dropdown. Lazy: only runs when `enabled`
 * (i.e. when the dropdown is open) to avoid polling inventory on every page.
 */
export function useStockAlerts(enabled: boolean) {
  const { activeBranch } = useBranch();
  return useQuery({
    queryKey: ['inventory', 'alerts', activeBranch?.id],
    queryFn: async () => {
      const [out, low] = await Promise.all([
        inventoryService.getInventory({
          status: 'out_of_stock',
          branch_id: activeBranch?.id,
          per_page: 20,
        }),
        inventoryService.getInventory({
          status: 'low',
          branch_id: activeBranch?.id,
          per_page: 20,
        }),
      ]);
      return { outOfStock: out.data, lowStock: low.data };
    },
    enabled,
    staleTime: 60 * 1000,
  });
}

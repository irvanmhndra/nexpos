import { createContext, useContext, useState, useEffect } from 'react';
import { useQuery } from '@tanstack/react-query';
import type { ReactNode } from 'react';
import type { Branch } from '@/types/branch';
import { branchesService } from '@/services/branchesService';
import { useAuth } from '@/contexts/AuthContext';

const STORAGE_KEY = 'nexpos_active_branch_id';

interface BranchContextType {
  activeBranch: Branch | null;
  setActiveBranch: (branch: Branch) => void;
  branches: Branch[];
  isLoading: boolean;
}

const BranchContext = createContext<BranchContextType | undefined>(undefined);

export function BranchProvider({ children }: { children: ReactNode }) {
  const [activeBranch, setActiveBranchState] = useState<Branch | null>(null);
  const { user } = useAuth();

  const { data, isLoading } = useQuery({
    queryKey: ['branches', 'all'],
    queryFn: () => branchesService.getBranches({ per_page: 100, is_active: true }),
    staleTime: 5 * 60 * 1000, // 5 minutes
    enabled: !!user, // Don't fetch when unauthenticated (prevents 401 loop on /login page)
  });

  const branches = data?.data || [];

  // Once branches are loaded, resolve active branch:
  // 1. localStorage (user's explicit choice)
  // 2. user.default_branch from login response
  // 3. first branch in list
  useEffect(() => {
    if (branches.length === 0) return;

    const storedId = localStorage.getItem(STORAGE_KEY);
    if (storedId) {
      const found = branches.find(b => String(b.id) === storedId);
      if (found) {
        setActiveBranchState(found);
        return;
      }
    }

    if (user?.default_branch) {
      const found = branches.find(b => b.id === user.default_branch!.id);
      if (found) {
        setActiveBranchState(found);
        return;
      }
    }

    setActiveBranchState(branches[0]);
  }, [branches, user?.default_branch?.id]);

  const setActiveBranch = (branch: Branch) => {
    localStorage.setItem(STORAGE_KEY, String(branch.id));
    setActiveBranchState(branch);
  };

  return (
    <BranchContext.Provider value={{ activeBranch, setActiveBranch, branches, isLoading }}>
      {children}
    </BranchContext.Provider>
  );
}

export function useBranch() {
  const context = useContext(BranchContext);
  if (!context) throw new Error('useBranch must be used within BranchProvider');
  return context;
}

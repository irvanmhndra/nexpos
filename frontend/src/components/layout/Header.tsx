import { useState, useRef, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { Icon, Avatar } from '@/components/ui';
import { useAuth } from '@/contexts/AuthContext';
import { useBranch } from '@/contexts/BranchContext';
import { useInventoryStats } from '@/hooks/useInventory';
import { useStockAlerts } from '@/hooks/useStockAlerts';

interface HeaderProps {
  title?: string;
  onMenuClick?: () => void;
}

export function Header({ title = 'Dashboard', onMenuClick }: HeaderProps) {
  const [isDropdownOpen, setIsDropdownOpen] = useState(false);
  const [isBranchOpen, setIsBranchOpen] = useState(false);
  const [isNotifOpen, setIsNotifOpen] = useState(false);
  const dropdownRef = useRef<HTMLDivElement>(null);
  const branchRef = useRef<HTMLDivElement>(null);
  const notifRef = useRef<HTMLDivElement>(null);
  const navigate = useNavigate();
  const { user, logout } = useAuth();
  const { activeBranch, setActiveBranch, branches, isLoading: isBranchLoading } = useBranch();

  // Notifications: stock alerts for the active branch.
  const { data: stats } = useInventoryStats();
  const alertCount = (stats?.low_stock ?? 0) + (stats?.out_of_stock ?? 0);
  const { data: alerts } = useStockAlerts(isNotifOpen);

  // Close dropdown when clicking outside
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
        setIsDropdownOpen(false);
      }
    };

    if (isDropdownOpen) {
      document.addEventListener('mousedown', handleClickOutside);
    }

    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
    };
  }, [isDropdownOpen]);

  // Close branch selector when clicking outside
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (branchRef.current && !branchRef.current.contains(event.target as Node)) {
        setIsBranchOpen(false);
      }
    };

    if (isBranchOpen) {
      document.addEventListener('mousedown', handleClickOutside);
    }

    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
    };
  }, [isBranchOpen]);

  // Close notifications when clicking outside
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (notifRef.current && !notifRef.current.contains(event.target as Node)) {
        setIsNotifOpen(false);
      }
    };

    if (isNotifOpen) {
      document.addEventListener('mousedown', handleClickOutside);
    }

    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
    };
  }, [isNotifOpen]);

  const handleLogout = () => {
    logout();
    setIsDropdownOpen(false);
    navigate('/login');
  };

  return (
    <header className="flex items-center justify-between whitespace-nowrap border-b border-border-light dark:border-border-dark bg-card-light dark:bg-card-dark px-4 md:px-6 py-3 flex-shrink-0 z-10">
      <div className="flex items-center gap-4 lg:gap-8 flex-1">
        {/* Mobile Menu Button */}
        <button
          onClick={onMenuClick}
          className="lg:hidden text-text-main dark:text-white p-2 hover:bg-background-light dark:hover:bg-background-dark rounded-lg transition-colors"
        >
          <Icon name="menu" size="md" />
        </button>

        {/* Page Title */}
        <div className="hidden md:flex flex-col">
          <h2 className="text-text-main dark:text-white text-lg font-bold leading-tight tracking-[-0.015em]">
            {title}
          </h2>
        </div>

        {/* Branch Selector */}
        <div className="hidden md:flex relative" ref={branchRef}>
          <button
            onClick={() => setIsBranchOpen(open => !open)}
            disabled={isBranchLoading || branches.length === 0}
            className="flex items-center gap-2 px-3 py-1.5 bg-background-light dark:bg-background-dark border border-border-light dark:border-border-dark rounded-lg text-sm font-medium text-text-main dark:text-white hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors disabled:opacity-60 disabled:cursor-not-allowed"
          >
            <Icon name="store" size="sm" className="text-text-sub dark:text-gray-400" />
            <span className="max-w-[160px] truncate">
              {isBranchLoading
                ? 'Memuat…'
                : activeBranch?.name ?? (branches.length === 0 ? 'Tidak ada cabang' : 'Pilih Cabang')}
            </span>
            <Icon
              name="expand_more"
              size="sm"
              className={`text-text-sub dark:text-gray-400 transition-transform ${isBranchOpen ? 'rotate-180' : ''}`}
            />
          </button>

          {isBranchOpen && branches.length > 0 && (
            <div className="absolute left-0 top-full mt-2 w-64 bg-white dark:bg-card-dark rounded-xl shadow-lg border border-border-light dark:border-border-dark overflow-hidden z-50 py-2 max-h-80 overflow-y-auto">
              {branches.map(branch => (
                <button
                  key={branch.id}
                  onClick={() => {
                    setActiveBranch(branch);
                    setIsBranchOpen(false);
                  }}
                  className="w-full flex items-center justify-between gap-3 px-4 py-2.5 hover:bg-background-light dark:hover:bg-background-dark text-left transition-colors"
                >
                  <span className="flex items-center gap-2.5 min-w-0">
                    <Icon name="store" size={18} className="text-text-sub dark:text-gray-400 shrink-0" />
                    <span className="min-w-0">
                      <span className="block text-sm font-medium text-text-main dark:text-white truncate">
                        {branch.name}
                      </span>
                      <span className="block text-xs text-text-sub dark:text-gray-400 truncate">
                        {branch.code}
                      </span>
                    </span>
                  </span>
                  {activeBranch?.id === branch.id && (
                    <Icon name="check" size={18} className="text-primary shrink-0" />
                  )}
                </button>
              ))}
            </div>
          )}
        </div>
      </div>

      {/* Right Section */}
      <div className="flex items-center gap-3 md:gap-4 pl-4">
        {/* Notifications */}
        <div className="relative" ref={notifRef}>
          <button
            onClick={() => setIsNotifOpen(open => !open)}
            className="flex items-center justify-center size-10 rounded-full hover:bg-background-light dark:hover:bg-background-dark text-text-main dark:text-white transition-colors relative"
            aria-label="Notifikasi"
          >
            <Icon name="notifications" size="md" />
            {alertCount > 0 && (
              <span className="absolute -top-0.5 -right-0.5 min-w-[18px] h-[18px] px-1 bg-red-500 text-white text-[10px] font-bold rounded-full border border-white dark:border-card-dark flex items-center justify-center">
                {alertCount > 9 ? '9+' : alertCount}
              </span>
            )}
          </button>

          {isNotifOpen && (
            <div className="absolute right-0 mt-2 w-80 max-w-[calc(100vw-2rem)] bg-white dark:bg-card-dark rounded-xl shadow-lg border border-border-light dark:border-border-dark overflow-hidden z-50">
              <div className="flex items-center justify-between px-4 py-3 border-b border-border-light dark:border-border-dark">
                <p className="text-sm font-bold text-text-main dark:text-white">Notifikasi</p>
                {alertCount > 0 && (
                  <span className="text-xs font-medium text-text-sub dark:text-gray-400">
                    {alertCount} stok perlu perhatian
                  </span>
                )}
              </div>

              <div className="max-h-80 overflow-y-auto">
                {alertCount === 0 ? (
                  <div className="px-4 py-8 text-center">
                    <Icon name="check_circle" size={28} className="text-green-500 mx-auto mb-2" />
                    <p className="text-sm text-text-sub dark:text-gray-400">Tidak ada notifikasi</p>
                  </div>
                ) : (
                  <>
                    {alerts?.outOfStock.map(item => (
                      <button
                        key={`out-${item.product_variant_id}`}
                        onClick={() => {
                          setIsNotifOpen(false);
                          navigate('/inventory');
                        }}
                        className="w-full flex items-start gap-3 px-4 py-3 hover:bg-background-light dark:hover:bg-background-dark text-left transition-colors border-b border-border-light dark:border-border-dark last:border-0"
                      >
                        <Icon name="error" size={20} className="text-red-500 shrink-0 mt-0.5" />
                        <div className="min-w-0">
                          <p className="text-sm font-medium text-text-main dark:text-white truncate">
                            {item.product_name}
                            {item.variant_name && item.variant_name !== 'Default' && ` · ${item.variant_name}`}
                          </p>
                          <p className="text-xs text-red-600 dark:text-red-400">Stok habis</p>
                        </div>
                      </button>
                    ))}
                    {alerts?.lowStock.map(item => (
                      <button
                        key={`low-${item.product_variant_id}`}
                        onClick={() => {
                          setIsNotifOpen(false);
                          navigate('/inventory');
                        }}
                        className="w-full flex items-start gap-3 px-4 py-3 hover:bg-background-light dark:hover:bg-background-dark text-left transition-colors border-b border-border-light dark:border-border-dark last:border-0"
                      >
                        <Icon name="warning" size={20} className="text-amber-500 shrink-0 mt-0.5" />
                        <div className="min-w-0">
                          <p className="text-sm font-medium text-text-main dark:text-white truncate">
                            {item.product_name}
                            {item.variant_name && item.variant_name !== 'Default' && ` · ${item.variant_name}`}
                          </p>
                          <p className="text-xs text-amber-600 dark:text-amber-400">
                            Stok menipis ({item.current_stock}/{item.min_stock})
                          </p>
                        </div>
                      </button>
                    ))}
                    {!alerts && (
                      <div className="px-4 py-6 text-center text-sm text-text-sub dark:text-gray-400">
                        Memuat…
                      </div>
                    )}
                  </>
                )}
              </div>

              <button
                onClick={() => {
                  setIsNotifOpen(false);
                  navigate('/inventory');
                }}
                className="w-full px-4 py-2.5 text-sm font-medium text-primary hover:bg-background-light dark:hover:bg-background-dark border-t border-border-light dark:border-border-dark transition-colors"
              >
                Lihat inventaris
              </button>
            </div>
          )}
        </div>

        {/* User Profile */}
        <div className="relative" ref={dropdownRef}>
          <div
            className="flex items-center gap-3 cursor-pointer group"
            onClick={() => setIsDropdownOpen(!isDropdownOpen)}
          >
            <div className="hidden sm:block text-right">
              <p className="text-sm font-bold text-text-main dark:text-white">{user?.name || 'User'}</p>
              <p className="text-xs text-text-sub dark:text-gray-400">{user?.email || ''}</p>
            </div>
            <Avatar
              name={user?.name || 'User'}
              size="md"
              className="border-2 border-border-light dark:border-border-dark"
            />
          </div>

          {/* Dropdown Menu */}
          {isDropdownOpen && (
            <div className="absolute right-0 mt-2 w-64 bg-white dark:bg-card-dark rounded-xl shadow-lg border border-border-light dark:border-border-dark overflow-hidden z-50">
              {/* User Info */}
              <div className="px-4 py-3 border-b border-border-light dark:border-border-dark">
                <p className="text-sm font-bold text-text-main dark:text-white">{user?.name || 'User'}</p>
                <p className="text-xs text-text-sub dark:text-gray-400">{user?.email || ''}</p>
              </div>

              {/* Menu Items */}
              <div className="py-2">
                <button
                  onClick={() => {
                    setIsDropdownOpen(false);
                    navigate('/settings');
                  }}
                  className="w-full flex items-center gap-3 px-4 py-2.5 hover:bg-background-light dark:hover:bg-background-dark text-left transition-colors"
                >
                  <Icon name="settings" size={20} className="text-text-sub dark:text-gray-400" />
                  <span className="text-sm font-medium text-text-main dark:text-white">Pengaturan</span>
                </button>

                <button
                  onClick={() => {
                    setIsDropdownOpen(false);
                    navigate('/settings');
                  }}
                  className="w-full flex items-center gap-3 px-4 py-2.5 hover:bg-background-light dark:hover:bg-background-dark text-left transition-colors"
                >
                  <Icon name="person" size={20} className="text-text-sub dark:text-gray-400" />
                  <span className="text-sm font-medium text-text-main dark:text-white">Profil Saya</span>
                </button>

                <button
                  onClick={() => {
                    setIsDropdownOpen(false);
                  }}
                  className="w-full flex items-center gap-3 px-4 py-2.5 hover:bg-background-light dark:hover:bg-background-dark text-left transition-colors"
                >
                  <Icon name="help" size={20} className="text-text-sub dark:text-gray-400" />
                  <span className="text-sm font-medium text-text-main dark:text-white">Bantuan</span>
                </button>
              </div>

              {/* Logout */}
              <div className="border-t border-border-light dark:border-border-dark py-2">
                <button
                  onClick={handleLogout}
                  className="w-full flex items-center gap-3 px-4 py-2.5 hover:bg-red-50 dark:hover:bg-red-900/20 text-left transition-colors"
                >
                  <Icon name="logout" size={20} className="text-red-600 dark:text-red-400" />
                  <span className="text-sm font-medium text-red-600 dark:text-red-400">Keluar</span>
                </button>
              </div>
            </div>
          )}
        </div>
      </div>
    </header>
  );
}

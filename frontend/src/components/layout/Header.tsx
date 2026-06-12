import { useState, useRef, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { Icon, Avatar } from '@/components/ui';
import { useAuth } from '@/contexts/AuthContext';
import { useBranch } from '@/contexts/BranchContext';

interface HeaderProps {
  title?: string;
  onMenuClick?: () => void;
}

export function Header({ title = 'Dashboard', onMenuClick }: HeaderProps) {
  const [isDropdownOpen, setIsDropdownOpen] = useState(false);
  const [isBranchOpen, setIsBranchOpen] = useState(false);
  const dropdownRef = useRef<HTMLDivElement>(null);
  const branchRef = useRef<HTMLDivElement>(null);
  const navigate = useNavigate();
  const { user, logout } = useAuth();
  const { activeBranch, setActiveBranch, branches, isLoading: isBranchLoading } = useBranch();

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
        <button className="flex items-center justify-center size-10 rounded-full hover:bg-background-light dark:hover:bg-background-dark text-text-main dark:text-white transition-colors relative">
          <Icon name="notifications" size="md" />
          <span className="absolute top-2 right-2 size-2 bg-red-500 rounded-full border border-white dark:border-card-dark"></span>
        </button>

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

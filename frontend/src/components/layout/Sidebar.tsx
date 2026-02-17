import { Link, useLocation, useNavigate } from 'react-router-dom';
import { Icon, Button } from '@/components/ui';
import { cn } from '@/utils';
import { useAuth } from '@/contexts/AuthContext';

interface NavItem {
  label: string;
  icon: string;
  path: string;
}

const navItems: NavItem[] = [
  { label: 'Dashboard', icon: 'dashboard', path: '/' },
  { label: 'Produk', icon: 'inventory_2', path: '/products' },
  { label: 'Kategori', icon: 'category', path: '/categories' },
  { label: 'Pesanan', icon: 'shopping_cart', path: '/orders' },
  { label: 'Pelanggan', icon: 'group', path: '/customers' },
  { label: 'Cabang', icon: 'store', path: '/branches' },
  { label: 'Laporan', icon: 'description', path: '/reports' },
  { label: 'Pengaturan', icon: 'settings', path: '/settings' },
];

export function Sidebar() {
  const location = useLocation();
  const navigate = useNavigate();
  const { logout } = useAuth();

  const isActive = (path: string) => {
    return location.pathname === path;
  };

  const handleLogout = () => {
    logout();
    navigate('/login');
  };

  return (
    <aside className="hidden lg:flex w-72 flex-col bg-card-light dark:bg-card-dark border-r border-border-light dark:border-border-dark flex-shrink-0 transition-colors duration-200">
      <div className="flex h-full flex-col justify-between p-4">
        {/* Logo & Brand */}
        <div className="flex flex-col gap-6">
          <Link to="/" className="flex items-center gap-3 px-2 group">
            <div className="bg-primary/10 p-2 rounded-lg">
              <Icon name="analytics" filled className="text-primary" size={28} />
            </div>
            <div className="flex flex-col">
              <h1 className="text-text-main dark:text-white text-lg font-bold leading-normal">
                Aplikabizz
              </h1>
              <p className="text-text-sub dark:text-gray-400 text-sm font-medium leading-normal">
                Business Manager
              </p>
            </div>
          </Link>

          {/* Navigation */}
          <nav className="flex flex-col gap-2">
            {navItems.map(item => (
              <Link
                key={item.path}
                to={item.path}
                className={cn(
                  'flex items-center gap-3 px-3 py-2.5 rounded-lg transition-colors group',
                  isActive(item.path)
                    ? 'bg-primary/10 text-primary'
                    : 'text-text-main dark:text-gray-300 hover:bg-background-light dark:hover:bg-background-dark'
                )}
              >
                <Icon
                  name={item.icon}
                  filled={isActive(item.path)}
                  size="md"
                  className={cn(!isActive(item.path) && 'group-hover:text-primary')}
                />
                <p
                  className={cn(
                    'text-sm leading-normal',
                    isActive(item.path) ? 'font-bold' : 'font-medium'
                  )}
                >
                  {item.label}
                </p>
              </Link>
            ))}
          </nav>
        </div>

        {/* Logout Button */}
        <Button
          variant="ghost"
          className="w-full justify-start"
          leftIcon={<Icon name="logout" size="sm" />}
          onClick={handleLogout}
        >
          Log Out
        </Button>
      </div>
    </aside>
  );
}

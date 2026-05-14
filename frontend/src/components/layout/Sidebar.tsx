import { Link, useLocation } from 'react-router-dom';
import { Icon } from '@/components/ui';
import { cn } from '@/utils';
import { useBranch } from '@/contexts/BranchContext';

interface NavItem {
  label: string;
  icon: string;
  path: string;
}

interface NavGroup {
  title: string;
  items: NavItem[];
}

const navGroups: NavGroup[] = [
  {
    title: '',
    items: [
      { label: 'Dashboard', icon: 'dashboard', path: '/' },
      { label: 'Kasir', icon: 'point_of_sale', path: '/pos' },
      { label: 'Pesanan', icon: 'receipt_long', path: '/orders' },
    ],
  },
  {
    title: 'Katalog',
    items: [
      { label: 'Produk', icon: 'inventory_2', path: '/products' },
      { label: 'Kategori', icon: 'category', path: '/categories' },
      { label: 'Inventaris', icon: 'warehouse', path: '/inventory' },
      { label: 'Stock Opname', icon: 'fact_check', path: '/stock-opnames' },
    ],
  },
  {
    title: 'Pelanggan',
    items: [
      { label: 'Pelanggan', icon: 'group', path: '/customers' },
      { label: 'Promo', icon: 'local_offer', path: '/promotions' },
    ],
  },
  {
    title: 'Pembelian',
    items: [
      { label: 'Supplier', icon: 'local_shipping', path: '/suppliers' },
      { label: 'Purchase Order', icon: 'shopping_bag', path: '/purchase-orders' },
    ],
  },
  {
    title: 'Keuangan',
    items: [
      { label: 'Shift', icon: 'schedule', path: '/shifts' },
      { label: 'Pengeluaran', icon: 'receipt', path: '/expenses' },
      { label: 'Kategori Pengeluaran', icon: 'label', path: '/expense-categories' },
    ],
  },
  {
    title: 'Lainnya',
    items: [
      { label: 'Laporan', icon: 'bar_chart', path: '/reports' },
      { label: 'Cabang', icon: 'store', path: '/branches' },
      { label: 'Pengaturan', icon: 'settings', path: '/settings' },
    ],
  },
];

export function Sidebar() {
  const location = useLocation();
  const { activeBranch, branches, setActiveBranch } = useBranch();

  const isActive = (path: string) => {
    if (path === '/') return location.pathname === '/';
    return location.pathname === path || location.pathname.startsWith(path + '/');
  };

  return (
    <aside className="hidden lg:flex w-72 flex-col bg-card-light dark:bg-card-dark border-r border-border-light dark:border-border-dark flex-shrink-0 transition-colors duration-200">
      <div className="flex h-full min-h-0 flex-col justify-between p-4">
        {/* Logo & Brand */}
        <div className="flex flex-col gap-6 flex-1 min-h-0 overflow-y-auto">
          <Link to="/" className="flex items-center gap-3 px-2 group">
            <div className="bg-primary/10 p-2 rounded-lg">
              <Icon name="analytics" filled className="text-primary" size={28} />
            </div>
            <div className="flex flex-col">
              <h1 className="text-text-main dark:text-white text-lg font-bold leading-normal">
                Nexpos
              </h1>
              <p className="text-text-sub dark:text-gray-400 text-sm font-medium leading-normal">
                Business Manager
              </p>
            </div>
          </Link>

          {/* Branch Switcher */}
          {activeBranch && (
            <div className="px-2">
              <div className="flex items-center gap-2 px-3 py-2.5 rounded-lg bg-primary/5 border border-primary/10">
                <Icon name="store" size={16} className="text-primary flex-shrink-0" />
                <div className="flex-1 min-w-0">
                  <p className="text-xs text-text-sub dark:text-gray-400">Cabang Aktif</p>
                  {branches.length > 1 ? (
                    <select
                      value={activeBranch.id}
                      onChange={e => {
                        const branch = branches.find(b => b.id === Number(e.target.value));
                        if (branch) setActiveBranch(branch);
                      }}
                      className="w-full bg-transparent text-sm font-medium text-text-main dark:text-white focus:outline-none cursor-pointer truncate"
                    >
                      {branches.map(b => (
                        <option key={b.id} value={b.id}>{b.name}</option>
                      ))}
                    </select>
                  ) : (
                    <p className="text-sm font-medium text-text-main dark:text-white truncate">{activeBranch.name}</p>
                  )}
                </div>
              </div>
            </div>
          )}

          {/* Navigation */}
          <nav className="flex flex-col gap-1">
            {navGroups.map((group, groupIdx) => (
              <div key={group.title || groupIdx}>
                {groupIdx > 0 && (
                  <div className="my-2 mx-2 border-t border-border-light dark:border-border-dark" />
                )}
                {group.title && (
                  <p className="px-3 pt-1 pb-1.5 text-[11px] font-semibold uppercase tracking-wider text-text-sub/60 dark:text-gray-500">
                    {group.title}
                  </p>
                )}
                {group.items.map(item => (
                  <Link
                    key={item.path}
                    to={item.path}
                    className={cn(
                      'flex items-center gap-3 px-3 py-2 rounded-lg transition-colors group',
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
              </div>
            ))}
          </nav>
        </div>

      </div>
    </aside>
  );
}

import { Icon } from '@/components/ui';

interface NotificationToggle {
  email: boolean;
  app: boolean;
}

interface NotificationItem {
  id: string;
  label: string;
  description: string;
  badge?: { text: string; color: 'yellow' | 'red' };
  toggles: NotificationToggle;
}

interface NotificationSection {
  title: string;
  items: NotificationItem[];
}

const defaultSections: NotificationSection[] = [
  {
    title: 'Penjualan & Pesanan',
    items: [
      {
        id: 'new_order',
        label: 'Pesanan Masuk Baru',
        description: 'Setiap ada pesanan baru yang masuk',
        badge: { text: 'Penting', color: 'yellow' },
        toggles: { email: true, app: true },
      },
      {
        id: 'order_status',
        label: 'Perubahan Status Pesanan',
        description: 'Pesanan diproses, dikirim, atau selesai',
        toggles: { email: false, app: true },
      },
      {
        id: 'payment_received',
        label: 'Pembayaran Diterima',
        description: 'Konfirmasi pembayaran berhasil',
        toggles: { email: true, app: true },
      },
    ],
  },
  {
    title: 'Inventaris & Stok',
    items: [
      {
        id: 'low_stock',
        label: 'Stok Hampir Habis',
        description: 'Stok produk di bawah batas minimum',
        badge: { text: 'Penting', color: 'yellow' },
        toggles: { email: true, app: true },
      },
      {
        id: 'out_of_stock',
        label: 'Stok Habis',
        description: 'Produk sudah tidak tersedia',
        toggles: { email: false, app: true },
      },
    ],
  },
  {
    title: 'Sistem & Akun',
    items: [
      {
        id: 'new_device_login',
        label: 'Login Perangkat Baru',
        description: 'Akses dari perangkat yang belum dikenal',
        badge: { text: 'Keamanan', color: 'red' },
        toggles: { email: true, app: true },
      },
      {
        id: 'system_update',
        label: 'Pembaruan Sistem',
        description: 'Rilis fitur dan pembaruan aplikasi',
        toggles: { email: false, app: false },
      },
    ],
  },
];

function Toggle({ checked, disabled }: { checked: boolean; disabled?: boolean }) {
  return (
    <button
      type="button"
      role="switch"
      aria-checked={checked}
      disabled={disabled}
      className={`relative inline-flex h-6 w-11 flex-shrink-0 rounded-full border-2 border-transparent transition-colors duration-200 focus:outline-none ${
        checked ? 'bg-primary' : 'bg-gray-200 dark:bg-gray-700'
      } ${disabled ? 'opacity-50 cursor-not-allowed' : 'cursor-pointer'}`}
    >
      <span
        className={`pointer-events-none inline-block size-5 transform rounded-full bg-white shadow ring-0 transition duration-200 ease-in-out ${
          checked ? 'translate-x-5' : 'translate-x-0'
        }`}
      />
    </button>
  );
}

export function NotificationSettings() {
  return (
    <div className="flex flex-col bg-white dark:bg-card-dark rounded-xl shadow-sm border border-border-light dark:border-border-dark overflow-hidden">
      {/* Header */}
      <div className="flex flex-col gap-2 p-6 md:p-8 border-b border-border-light dark:border-border-dark">
        <div className="flex items-center gap-3 flex-wrap">
          <h2 className="text-2xl md:text-[32px] font-bold text-text-main dark:text-white">
            Notifikasi
          </h2>
          <span className="inline-flex items-center px-2.5 py-1 rounded-full text-xs font-semibold bg-primary/10 text-primary">
            Segera hadir
          </span>
        </div>
        <p className="text-sm md:text-base text-text-sub dark:text-gray-400">
          Atur preferensi notifikasi untuk setiap jenis aktivitas
        </p>
      </div>

      {/* Coming-soon banner */}
      <div className="flex items-start gap-3 px-6 md:px-8 py-4 bg-amber-50 dark:bg-amber-900/20 border-b border-amber-200/60 dark:border-amber-800/40">
        <Icon name="info" size={20} className="text-amber-600 dark:text-amber-400 shrink-0 mt-0.5" />
        <div className="text-sm text-amber-800 dark:text-amber-200">
          <p className="font-medium">Preferensi notifikasi belum aktif.</p>
          <p className="mt-0.5 text-amber-700 dark:text-amber-300">
            Pengaturan di bawah ini masih dalam pengembangan. Untuk saat ini, peringatan stok
            menipis &amp; habis sudah muncul otomatis di ikon lonceng 🔔 pada bar atas.
          </p>
        </div>
      </div>

      {/* Column Headers */}
      <div className="hidden md:flex items-center px-6 md:px-8 py-3 bg-gray-50 dark:bg-gray-800/50 border-b border-border-light dark:border-border-dark">
        <div className="flex-1" />
        <div className="flex gap-10 pr-2">
          <span className="text-xs font-semibold text-text-sub dark:text-gray-400 uppercase tracking-wide w-16 text-center">
            Email
          </span>
          <span className="text-xs font-semibold text-text-sub dark:text-gray-400 uppercase tracking-wide w-16 text-center">
            Aplikasi
          </span>
        </div>
      </div>

      {/* Sections — preview only, controls disabled until the feature ships */}
      <div className="flex flex-col divide-y divide-border-light dark:divide-border-dark opacity-70">
        {defaultSections.map(section => (
          <div key={section.title}>
            <div className="px-6 md:px-8 py-3 bg-background-light dark:bg-background-dark">
              <h3 className="text-sm font-bold text-text-main dark:text-white">{section.title}</h3>
            </div>
            <div className="flex flex-col divide-y divide-border-light dark:divide-border-dark">
              {section.items.map(item => (
                <div
                  key={item.id}
                  className="flex flex-col md:flex-row md:items-center gap-3 px-6 md:px-8 py-4"
                >
                  <div className="flex-1 flex flex-col gap-0.5">
                    <div className="flex items-center gap-2">
                      <span className="text-sm font-semibold text-text-main dark:text-white">
                        {item.label}
                      </span>
                      {item.badge && (
                        <span
                          className={`inline-flex items-center px-2 py-0.5 rounded-full text-xs font-semibold ${
                            item.badge.color === 'yellow'
                              ? 'bg-amber-100 dark:bg-amber-900/30 text-amber-700 dark:text-amber-400'
                              : 'bg-red-100 dark:bg-red-900/30 text-red-700 dark:text-red-400'
                          }`}
                        >
                          {item.badge.text}
                        </span>
                      )}
                    </div>
                    <p className="text-xs text-text-sub dark:text-gray-400">{item.description}</p>
                  </div>
                  <div className="flex items-center gap-6 md:gap-10">
                    <div className="flex flex-col items-center gap-1 w-16">
                      <span className="text-xs text-text-sub dark:text-gray-400 md:hidden">
                        Email
                      </span>
                      <Toggle checked={item.toggles.email} disabled />
                    </div>
                    <div className="flex flex-col items-center gap-1 w-16">
                      <span className="text-xs text-text-sub dark:text-gray-400 md:hidden">
                        Aplikasi
                      </span>
                      <Toggle checked={item.toggles.app} disabled />
                    </div>
                  </div>
                </div>
              ))}
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

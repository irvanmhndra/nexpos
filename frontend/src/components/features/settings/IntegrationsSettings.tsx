import { useState } from 'react';
import { Icon } from '@/components/ui';

type IntegrationStatus = 'connected' | 'disconnected' | 'coming_soon';
type IntegrationCategory = 'payment' | 'shipping' | 'accounting' | 'messaging';

interface Integration {
  id: string;
  name: string;
  description: string;
  icon: string;
  category: IntegrationCategory;
  status: IntegrationStatus;
}

const integrations: Integration[] = [
  {
    id: 'midtrans',
    name: 'Midtrans',
    description: 'Payment gateway terpercaya untuk bisnis Indonesia',
    icon: 'payments',
    category: 'payment',
    status: 'connected',
  },
  {
    id: 'jne',
    name: 'JNE',
    description: 'Layanan pengiriman dan logistik terkemuka',
    icon: 'local_shipping',
    category: 'shipping',
    status: 'disconnected',
  },
  {
    id: 'jurnal',
    name: 'Jurnal.id',
    description: 'Software akuntansi online untuk UMKM Indonesia',
    icon: 'menu_book',
    category: 'accounting',
    status: 'disconnected',
  },
  {
    id: 'whatsapp',
    name: 'WhatsApp Business',
    description: 'Kirim notifikasi pesanan & invoice via WhatsApp',
    icon: 'chat',
    category: 'messaging',
    status: 'connected',
  },
  {
    id: 'xendit',
    name: 'Xendit',
    description: 'Solusi pembayaran untuk startup dan enterprise',
    icon: 'credit_card',
    category: 'payment',
    status: 'disconnected',
  },
  {
    id: 'jnt',
    name: 'J&T Express',
    description: 'Pengiriman ekspres ke seluruh Indonesia',
    icon: 'airport_shuttle',
    category: 'shipping',
    status: 'coming_soon',
  },
];

type TabFilter = 'all' | IntegrationCategory;

const tabs: { id: TabFilter; label: string }[] = [
  { id: 'all', label: 'Semua' },
  { id: 'payment', label: 'Pembayaran' },
  { id: 'shipping', label: 'Pengiriman' },
  { id: 'accounting', label: 'Akuntansi' },
];

const categoryLabels: Record<IntegrationCategory, string> = {
  payment: 'Pembayaran',
  shipping: 'Pengiriman',
  accounting: 'Akuntansi',
  messaging: 'Pesan',
};

export function IntegrationsSettings() {
  const [activeTab, setActiveTab] = useState<TabFilter>('all');
  const [search, setSearch] = useState('');

  const filtered = integrations.filter(item => {
    const matchesTab = activeTab === 'all' || item.category === activeTab;
    const matchesSearch =
      !search ||
      item.name.toLowerCase().includes(search.toLowerCase()) ||
      item.description.toLowerCase().includes(search.toLowerCase());
    return matchesTab && matchesSearch;
  });

  return (
    <div className="flex flex-col gap-6">
      {/* Main Card */}
      <div className="flex flex-col bg-white dark:bg-card-dark rounded-xl shadow-sm border border-border-light dark:border-border-dark overflow-hidden">
        {/* Header */}
        <div className="flex flex-col gap-2 p-6 md:p-8 border-b border-border-light dark:border-border-dark">
          <h2 className="text-2xl md:text-[32px] font-bold text-text-main dark:text-white">
            Integrasi
          </h2>
          <p className="text-sm md:text-base text-text-sub dark:text-gray-400">
            Hubungkan aplikasi dan layanan pihak ketiga untuk memperluas fungsionalitas bisnis Anda
          </p>
        </div>

        {/* Search & Tabs */}
        <div className="flex flex-col gap-4 p-6 md:p-8 border-b border-border-light dark:border-border-dark">
          <div className="relative max-w-sm">
            <Icon
              name="search"
              size={18}
              className="absolute left-3 top-1/2 -translate-y-1/2 text-text-sub dark:text-gray-400"
            />
            <input
              type="text"
              value={search}
              onChange={e => setSearch(e.target.value)}
              placeholder="Cari integrasi..."
              className="w-full rounded-lg border border-border-light dark:border-gray-600 bg-white dark:bg-card-dark text-text-main dark:text-white focus:border-primary focus:ring-1 focus:ring-primary h-10 pl-10 pr-4 text-sm placeholder:text-text-sub dark:placeholder:text-gray-500"
            />
          </div>
          <div className="flex gap-2 overflow-x-auto pb-1">
            {tabs.map(tab => (
              <button
                key={tab.id}
                onClick={() => setActiveTab(tab.id)}
                className={`flex-shrink-0 px-4 py-1.5 rounded-full text-sm font-medium transition-colors ${
                  activeTab === tab.id
                    ? 'bg-primary text-white'
                    : 'bg-background-light dark:bg-background-dark text-text-sub dark:text-gray-400 hover:text-text-main dark:hover:text-white'
                }`}
              >
                {tab.label}
              </button>
            ))}
          </div>
        </div>

        {/* Integration Cards Grid */}
        <div className="p-6 md:p-8">
          {filtered.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-12 gap-3">
              <Icon name="search_off" size={40} className="text-text-sub dark:text-gray-600" />
              <p className="text-text-sub dark:text-gray-400">
                Tidak ada integrasi yang ditemukan
              </p>
            </div>
          ) : (
            <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4">
              {filtered.map(integration => (
                <IntegrationCard key={integration.id} integration={integration} />
              ))}
            </div>
          )}
        </div>
      </div>

      {/* Help Section */}
      <div className="flex flex-col sm:flex-row gap-4 items-start p-6 bg-primary/5 rounded-xl border border-primary/10">
        <div className="size-10 rounded-lg bg-primary/10 flex items-center justify-center flex-shrink-0">
          <Icon name="help_outline" size={20} className="text-primary" />
        </div>
        <div className="flex flex-col gap-1">
          <p className="text-sm font-bold text-text-main dark:text-white">
            Butuh integrasi khusus?
          </p>
          <p className="text-xs text-text-sub dark:text-gray-400">
            Hubungi tim kami untuk mendiskusikan kebutuhan integrasi yang belum tersedia di daftar
            ini.
          </p>
        </div>
        <button className="flex-shrink-0 text-sm font-semibold text-primary hover:underline mt-1 sm:mt-0">
          Hubungi Kami
        </button>
      </div>
    </div>
  );
}

function IntegrationCard({ integration }: { integration: Integration }) {
  const [connected, setConnected] = useState(integration.status === 'connected');
  const isComingSoon = integration.status === 'coming_soon';

  return (
    <div
      className={`flex flex-col gap-4 p-5 rounded-xl border-2 transition-colors ${
        connected
          ? 'border-emerald-400 dark:border-emerald-600 bg-emerald-50/30 dark:bg-emerald-900/10'
          : isComingSoon
            ? 'border-border-light dark:border-border-dark bg-gray-50 dark:bg-gray-800/30 opacity-60'
            : 'border-border-light dark:border-border-dark hover:border-primary/40'
      }`}
    >
      {/* Card Top */}
      <div className="flex items-start justify-between gap-3">
        <div
          className={`size-12 rounded-xl flex items-center justify-center flex-shrink-0 ${
            connected
              ? 'bg-emerald-100 dark:bg-emerald-900/30'
              : 'bg-background-light dark:bg-background-dark'
          }`}
        >
          <Icon
            name={integration.icon}
            size={24}
            className={connected ? 'text-emerald-600 dark:text-emerald-400' : 'text-text-sub dark:text-gray-400'}
          />
        </div>
        {connected ? (
          <span className="inline-flex items-center gap-1 px-2 py-1 rounded-full bg-emerald-100 dark:bg-emerald-900/30 text-emerald-700 dark:text-emerald-400 text-xs font-semibold">
            <span className="size-1.5 rounded-full bg-emerald-500" />
            Terhubung
          </span>
        ) : isComingSoon ? (
          <span className="inline-flex items-center px-2 py-1 rounded-full bg-gray-100 dark:bg-gray-800 text-gray-500 dark:text-gray-400 text-xs font-semibold">
            Segera
          </span>
        ) : null}
      </div>

      {/* Card Info */}
      <div className="flex flex-col gap-1 flex-1">
        <p className="text-sm font-bold text-text-main dark:text-white">{integration.name}</p>
        <p className="text-xs text-text-sub dark:text-gray-400 leading-relaxed">
          {integration.description}
        </p>
        <span className="text-xs text-text-sub dark:text-gray-500 mt-1">
          {categoryLabels[integration.category]}
        </span>
      </div>

      {/* Card Action */}
      <div className="flex gap-2">
        {isComingSoon ? (
          <button
            disabled
            className="flex-1 h-9 rounded-lg border border-border-light dark:border-gray-600 text-sm font-medium text-text-sub dark:text-gray-500 cursor-not-allowed"
          >
            Segera Hadir
          </button>
        ) : connected ? (
          <>
            <button className="flex-1 h-9 rounded-lg border border-border-light dark:border-gray-600 text-sm font-medium text-text-main dark:text-white hover:bg-background-light dark:hover:bg-background-dark transition-colors">
              Konfigurasi
            </button>
            <button
              onClick={() => setConnected(false)}
              className="h-9 px-3 rounded-lg border border-red-200 dark:border-red-800 text-sm font-medium text-red-600 dark:text-red-400 hover:bg-red-50 dark:hover:bg-red-900/20 transition-colors"
            >
              Putuskan
            </button>
          </>
        ) : (
          <button
            onClick={() => setConnected(true)}
            className="flex-1 h-9 rounded-lg bg-primary text-white text-sm font-semibold hover:bg-primary/90 transition-colors"
          >
            Hubungkan
          </button>
        )}
      </div>
    </div>
  );
}

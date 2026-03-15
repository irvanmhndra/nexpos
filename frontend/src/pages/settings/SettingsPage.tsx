import { useState } from 'react';
import { Button, Icon, Input } from '@/components/ui';
import { useAuth } from '@/contexts/AuthContext';
import { UserManagement } from '@/components/features/settings/UserManagement';
import { NotificationSettings } from '@/components/features/settings/NotificationSettings';
import { IntegrationsSettings } from '@/components/features/settings/IntegrationsSettings';
import { GeneralSettings } from '@/components/features/settings/GeneralSettings';
import { BusinessSettings } from '@/components/features/settings/BusinessSettings';

type SettingsSection = 'user-profile' | 'company-profile' | 'users' | 'business' | 'notifications' | 'integrations' | 'general';

export function SettingsPage() {
  const { user } = useAuth();
  const [activeSection, setActiveSection] = useState<SettingsSection>('user-profile');

  // User profile state
  const [userProfile, setUserProfile] = useState({
    name: user?.name || '',
    email: user?.email || '',
    phone: '',
  });

  // Company profile state
  const [companyProfile, setCompanyProfile] = useState({
    name: user?.company?.name || '',
    code: user?.company?.code || '',
    category: 'fnb',
    address: '',
    phone: '',
    email: '',
    description: '',
  });

  const [isUserSaving, setIsUserSaving] = useState(false);
  const [isCompanySaving, setIsCompanySaving] = useState(false);

  const settingsSections = [
    { id: 'user-profile' as const, label: 'Profil Saya', icon: 'person' },
    { id: 'company-profile' as const, label: 'Profil Perusahaan', icon: 'business' },
    { id: 'users' as const, label: 'Pengguna & Tim', icon: 'groups' },
    { id: 'business' as const, label: 'Pengaturan Bisnis', icon: 'storefront' },
    { id: 'notifications' as const, label: 'Notifikasi', icon: 'notifications' },
    { id: 'integrations' as const, label: 'Integrasi', icon: 'extension' },
    { id: 'general' as const, label: 'Umum', icon: 'settings' },
  ];

  const handleSaveUserProfile = async () => {
    setIsUserSaving(true);
    // Simulate API call
    await new Promise(resolve => setTimeout(resolve, 1000));
    setIsUserSaving(false);
  };

  const handleSaveCompanyProfile = async () => {
    setIsCompanySaving(true);
    // Simulate API call
    await new Promise(resolve => setTimeout(resolve, 1000));
    setIsCompanySaving(false);
  };

  const renderUserProfile = () => (
    <div className="flex flex-col gap-6 bg-white dark:bg-card-dark rounded-xl shadow-sm border border-border-light dark:border-border-dark overflow-hidden">
      {/* Header */}
      <div className="flex flex-col gap-2 p-6 md:p-8 border-b border-border-light dark:border-border-dark">
        <h2 className="text-2xl md:text-[32px] font-bold text-text-main dark:text-white">
          Profil Saya
        </h2>
        <p className="text-sm md:text-base text-text-sub dark:text-gray-400">
          Kelola informasi akun pribadi Anda
        </p>
      </div>

      {/* Avatar Section */}
      <div className="flex flex-col p-6 md:p-8 gap-6 border-b border-border-light dark:border-border-dark">
        <div className="flex flex-col sm:flex-row gap-6 items-start sm:items-center">
          <div className="bg-primary/10 rounded-full size-24 md:size-32 border-2 border-primary/20 flex items-center justify-center">
            <span className="text-4xl md:text-5xl font-bold text-primary">
              {userProfile.name.charAt(0).toUpperCase() || 'U'}
            </span>
          </div>
          <div className="flex flex-col gap-2 flex-1">
            <h3 className="text-lg font-bold text-text-main dark:text-white">
              {userProfile.name || 'User'}
            </h3>
            <p className="text-sm text-text-sub dark:text-gray-400">
              {userProfile.email}
            </p>
            <div className="flex items-center gap-2 mt-1">
              <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-primary/10 text-primary capitalize">
                {user?.role || 'User'}
              </span>
            </div>
          </div>
        </div>
      </div>

      {/* User Form Fields */}
      <div className="flex flex-col p-6 md:p-8 gap-6">
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          <Input
            label="Nama Lengkap"
            value={userProfile.name}
            onChange={(e) => setUserProfile({ ...userProfile, name: e.target.value })}
            placeholder="Masukkan nama lengkap"
            fullWidth
          />
          <Input
            label="Email"
            type="email"
            value={userProfile.email}
            onChange={(e) => setUserProfile({ ...userProfile, email: e.target.value })}
            placeholder="email@contoh.com"
            fullWidth
          />
        </div>

        <div className="flex flex-col gap-2">
          <label className="text-sm font-bold text-text-main dark:text-gray-200">
            Nomor Telepon <span className="text-text-sub font-normal">(Opsional)</span>
          </label>
          <div className="relative max-w-md">
            <span className="absolute left-4 top-1/2 -translate-y-1/2 text-text-sub dark:text-gray-400">
              +62
            </span>
            <input
              type="tel"
              value={userProfile.phone}
              onChange={(e) => setUserProfile({ ...userProfile, phone: e.target.value })}
              placeholder="8123456789"
              className="form-input w-full rounded-lg border border-border-light dark:border-gray-600 bg-white dark:bg-card-dark text-text-main dark:text-white focus:border-primary focus:ring-1 focus:ring-primary h-12 pl-12 pr-4 text-base placeholder:text-text-sub dark:placeholder:text-gray-500"
            />
          </div>
          <p className="text-xs text-text-sub dark:text-gray-400">
            Nomor telepon untuk notifikasi dan verifikasi akun
          </p>
        </div>

        {/* Company Info (Read Only) */}
        <div className="h-px w-full bg-border-light dark:bg-border-dark my-2"></div>

        <div className="flex flex-col gap-4">
          <h3 className="text-base font-bold text-text-main dark:text-white">
            Informasi Perusahaan
          </h3>
          <div className="flex items-center gap-4 p-4 bg-gray-50 dark:bg-gray-800/50 rounded-lg border border-border-light dark:border-gray-700">
            <div className="size-12 rounded-lg bg-primary/10 flex items-center justify-center">
              <Icon name="business" size={24} className="text-primary" />
            </div>
            <div className="flex-1">
              <p className="font-bold text-text-main dark:text-white">
                {user?.company?.name || 'Nama Perusahaan'}
              </p>
              <p className="text-sm text-text-sub dark:text-gray-400">
                Kode: {user?.company?.code || '-'}
              </p>
            </div>
            <Button
              size="sm"
              variant="ghost"
              onClick={() => setActiveSection('company-profile')}
              rightIcon={<Icon name="arrow_forward" size={16} />}
            >
              Lihat Detail
            </Button>
          </div>
        </div>
      </div>

      {/* Footer Actions */}
      <div className="sticky bottom-0 z-10 bg-white dark:bg-card-dark border-t border-border-light dark:border-border-dark p-6 flex flex-col-reverse sm:flex-row items-center justify-end gap-3">
        <Button variant="outline" className="w-full sm:w-auto">
          Batal
        </Button>
        <Button
          className="w-full sm:w-auto"
          onClick={handleSaveUserProfile}
          disabled={isUserSaving}
        >
          {isUserSaving ? 'Menyimpan...' : 'Simpan Perubahan'}
        </Button>
      </div>
    </div>
  );

  const renderCompanyProfile = () => (
    <div className="flex flex-col gap-6 bg-white dark:bg-card-dark rounded-xl shadow-sm border border-border-light dark:border-border-dark overflow-hidden">
      {/* Header */}
      <div className="flex flex-col gap-2 p-6 md:p-8 border-b border-border-light dark:border-border-dark">
        <div className="flex items-center justify-between">
          <div>
            <h2 className="text-2xl md:text-[32px] font-bold text-text-main dark:text-white">
              Profil Perusahaan
            </h2>
            <p className="text-sm md:text-base text-text-sub dark:text-gray-400 mt-1">
              Kelola informasi dasar dan branding bisnis Anda
            </p>
          </div>
          <div className="hidden sm:flex items-center gap-2 bg-background-light dark:bg-background-dark px-3 py-1.5 rounded-full border border-border-light dark:border-gray-700">
            <Icon name="store" size={18} className="text-text-sub" />
            <span className="text-sm font-medium text-text-main dark:text-gray-200">
              Pusat (HQ)
            </span>
            <Icon name="expand_more" size={18} className="text-text-sub" />
          </div>
        </div>
      </div>

      {/* Company Avatar Section */}
      <div className="flex flex-col p-6 md:p-8 gap-6 border-b border-border-light dark:border-border-dark">
        <div className="flex flex-col sm:flex-row gap-6 items-start sm:items-center">
          <div className="bg-primary/10 rounded-xl size-24 md:size-32 border-2 border-primary/20 flex items-center justify-center">
            <span className="text-3xl md:text-4xl font-bold text-primary">
              {companyProfile.name
                .split(' ')
                .map(word => word.charAt(0))
                .join('')
                .slice(0, 2)
                .toUpperCase() || 'CO'}
            </span>
          </div>
          <div className="flex flex-col gap-2 flex-1">
            <h3 className="text-lg font-bold text-text-main dark:text-white">
              {companyProfile.name || 'Nama Perusahaan'}
            </h3>
            <p className="text-sm text-text-sub dark:text-gray-400">
              Kode: {companyProfile.code || '-'}
            </p>
          </div>
        </div>
      </div>

      {/* Company Form Fields */}
      <div className="flex flex-col p-6 md:p-8 gap-6">
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          <Input
            label="Nama Perusahaan"
            value={companyProfile.name}
            onChange={(e) => setCompanyProfile({ ...companyProfile, name: e.target.value })}
            placeholder="Contoh: PT Maju Jaya"
            fullWidth
          />
          <div className="flex flex-col gap-2">
            <label className="text-sm font-bold text-text-main dark:text-gray-200">
              Kode Perusahaan
            </label>
            <div className="flex items-center gap-2 h-12 px-4 rounded-lg border border-border-light dark:border-gray-600 bg-gray-50 dark:bg-gray-800">
              <span className="text-text-main dark:text-white font-mono">
                {companyProfile.code || '-'}
              </span>
            </div>
            <p className="text-xs text-text-sub dark:text-gray-400">
              Kode unik perusahaan tidak dapat diubah
            </p>
          </div>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          <div className="flex flex-col gap-2">
            <label className="text-sm font-bold text-text-main dark:text-gray-200">
              Kategori Bisnis
            </label>
            <select
              value={companyProfile.category}
              onChange={(e) => setCompanyProfile({ ...companyProfile, category: e.target.value })}
              className="form-select w-full rounded-lg border border-border-light dark:border-gray-600 bg-white dark:bg-card-dark text-text-main dark:text-white focus:border-primary focus:ring-1 focus:ring-primary h-12 px-4 text-base"
            >
              <option value="fnb">Makanan & Minuman (F&B)</option>
              <option value="retail">Ritel</option>
              <option value="service">Jasa</option>
              <option value="manufacturing">Manufaktur</option>
              <option value="other">Lainnya</option>
            </select>
          </div>
        </div>

        <div className="flex flex-col gap-2">
          <label className="text-sm font-bold text-text-main dark:text-gray-200">
            Alamat
          </label>
          <textarea
            value={companyProfile.address}
            onChange={(e) => setCompanyProfile({ ...companyProfile, address: e.target.value })}
            placeholder="Masukkan alamat lengkap kantor atau toko..."
            className="form-textarea w-full rounded-lg border border-border-light dark:border-gray-600 bg-white dark:bg-card-dark text-text-main dark:text-white focus:border-primary focus:ring-1 focus:ring-primary min-h-[80px] p-4 placeholder:text-text-sub dark:placeholder:text-gray-500 text-base resize-y"
          />
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          <div className="flex flex-col gap-2">
            <label className="text-sm font-bold text-text-main dark:text-gray-200">
              Nomor Telepon Perusahaan
            </label>
            <div className="relative">
              <span className="absolute left-4 top-1/2 -translate-y-1/2 text-text-sub dark:text-gray-400">
                +62
              </span>
              <input
                type="tel"
                value={companyProfile.phone}
                onChange={(e) => setCompanyProfile({ ...companyProfile, phone: e.target.value })}
                placeholder="2112345678"
                className="form-input w-full rounded-lg border border-border-light dark:border-gray-600 bg-white dark:bg-card-dark text-text-main dark:text-white focus:border-primary focus:ring-1 focus:ring-primary h-12 pl-12 pr-4 text-base placeholder:text-text-sub dark:placeholder:text-gray-500"
              />
            </div>
          </div>
          <Input
            label="Email Perusahaan"
            type="email"
            value={companyProfile.email}
            onChange={(e) => setCompanyProfile({ ...companyProfile, email: e.target.value })}
            placeholder="info@perusahaan.com"
            fullWidth
          />
        </div>

        <div className="flex flex-col gap-2">
          <label className="text-sm font-bold text-text-main dark:text-gray-200">
            Deskripsi Perusahaan
          </label>
          <textarea
            value={companyProfile.description}
            onChange={(e) => setCompanyProfile({ ...companyProfile, description: e.target.value })}
            placeholder="Ceritakan sedikit tentang bisnis Anda..."
            className="form-textarea w-full rounded-lg border border-border-light dark:border-gray-600 bg-white dark:bg-card-dark text-text-main dark:text-white focus:border-primary focus:ring-1 focus:ring-primary min-h-[100px] p-4 placeholder:text-text-sub dark:placeholder:text-gray-500 text-base resize-y"
          />
          <p className="text-xs text-text-sub dark:text-gray-400">
            Deskripsi ini akan muncul di profil publik perusahaan Anda.
          </p>
        </div>

      </div>

      {/* Footer Actions */}
      <div className="sticky bottom-0 z-10 bg-white dark:bg-card-dark border-t border-border-light dark:border-border-dark p-6 flex flex-col-reverse sm:flex-row items-center justify-end gap-3">
        <Button variant="outline" className="w-full sm:w-auto">
          Batal
        </Button>
        <Button
          className="w-full sm:w-auto"
          onClick={handleSaveCompanyProfile}
          disabled={isCompanySaving}
        >
          {isCompanySaving ? 'Menyimpan...' : 'Simpan Perubahan'}
        </Button>
      </div>
    </div>
  );


  const renderContent = () => {
    switch (activeSection) {
      case 'user-profile':
        return renderUserProfile();
      case 'company-profile':
        return renderCompanyProfile();
      case 'users':
        return <UserManagement />;
      case 'business':
        return <BusinessSettings />;
      case 'notifications':
        return <NotificationSettings />;
      case 'integrations':
        return <IntegrationsSettings />;
      case 'general':
        return <GeneralSettings />;
      default:
        return renderUserProfile();
    }
  };

  return (
    <div className="flex gap-6 h-full">
      {/* Sidebar Navigation */}
      <aside className="hidden lg:flex flex-col w-64 gap-4">
        <div className="flex flex-col gap-1">
          {settingsSections.map((section) => (
            <button
              key={section.id}
              onClick={() => setActiveSection(section.id)}
              className={`flex items-center gap-3 px-3 py-2.5 rounded-lg transition-colors group border-l-4 ${
                activeSection === section.id
                  ? 'bg-background-light dark:bg-background-dark border-primary'
                  : 'border-transparent hover:bg-background-light dark:hover:bg-background-dark'
              }`}
            >
              <Icon
                name={section.icon}
                size={20}
                className={
                  activeSection === section.id
                    ? 'text-primary'
                    : 'text-text-sub dark:text-gray-400 group-hover:text-text-main dark:group-hover:text-white'
                }
              />
              <p
                className={`text-sm font-medium ${
                  activeSection === section.id
                    ? 'text-primary'
                    : 'text-text-sub dark:text-gray-400 group-hover:text-text-main dark:group-hover:text-white'
                }`}
              >
                {section.label}
              </p>
            </button>
          ))}
        </div>

        {/* Help Card */}
        <div className="flex flex-col gap-2 p-4 bg-primary/5 rounded-xl mt-auto border border-primary/10">
          <Icon name="support_agent" size={24} className="text-primary" />
          <p className="text-sm font-bold text-text-main dark:text-white">Butuh bantuan?</p>
          <p className="text-xs text-text-sub dark:text-gray-400">
            Hubungi tim support kami untuk kendala teknis.
          </p>
        </div>
      </aside>

      {/* Mobile Navigation */}
      <div className="lg:hidden fixed bottom-0 left-0 right-0 z-50 bg-white dark:bg-card-dark border-t border-border-light dark:border-border-dark p-2">
        <div className="flex justify-around">
          {settingsSections.slice(0, 4).map((section) => (
            <button
              key={section.id}
              onClick={() => setActiveSection(section.id)}
              className={`flex flex-col items-center gap-1 p-2 rounded-lg ${
                activeSection === section.id
                  ? 'text-primary'
                  : 'text-text-sub'
              }`}
            >
              <Icon name={section.icon} size={20} />
              <span className="text-[10px] font-medium">{section.label.split(' ')[0]}</span>
            </button>
          ))}
        </div>
      </div>

      {/* Main Content */}
      <main className="flex-1 min-w-0 pb-20 lg:pb-0">
        {renderContent()}
      </main>
    </div>
  );
}

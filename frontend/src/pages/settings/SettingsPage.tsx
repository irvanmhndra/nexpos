import { useState } from 'react';
import toast from 'react-hot-toast';
import { Button, Icon, Input } from '@/components/ui';
import { useAuth } from '@/contexts/AuthContext';
import { useUpdateUser } from '@/hooks/useUsers';
import { UserManagement } from '@/components/features/settings/UserManagement';
import { NotificationSettings } from '@/components/features/settings/NotificationSettings';
import { GeneralSettings } from '@/components/features/settings/GeneralSettings';
import { BusinessSettings } from '@/components/features/settings/BusinessSettings';
import { ScannerDevicesSettings } from '@/components/features/settings/ScannerDevicesSettings';

type SettingsSection = 'user-profile' | 'users' | 'business' | 'devices' | 'notifications' | 'general';

export function SettingsPage() {
  const { user } = useAuth();
  const [activeSection, setActiveSection] = useState<SettingsSection>('user-profile');
  const updateUser = useUpdateUser();

  // User profile state
  const [userProfile, setUserProfile] = useState({
    name: user?.name || '',
    email: user?.email || '',
  });

  const settingsSections = [
    { id: 'user-profile' as const, label: 'Profil Saya', icon: 'person' },
    { id: 'users' as const, label: 'Pengguna & Tim', icon: 'groups' },
    { id: 'business' as const, label: 'Pengaturan Bisnis', icon: 'storefront' },
    { id: 'devices' as const, label: 'Perangkat', icon: 'barcode_scanner' },
    { id: 'notifications' as const, label: 'Notifikasi', icon: 'notifications' },
    { id: 'general' as const, label: 'Umum', icon: 'settings' },
  ];

  const handleSaveUserProfile = async () => {
    if (!user) return;
    try {
      await updateUser.mutateAsync({
        id: user.id,
        data: { name: userProfile.name, email: userProfile.email },
      });
      toast.success('Profil berhasil disimpan');
    } catch {
      toast.error('Gagal menyimpan profil');
    }
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
              onClick={() => setActiveSection('business')}
              rightIcon={<Icon name="arrow_forward" size={16} />}
            >
              Lihat Detail
            </Button>
          </div>
        </div>
      </div>

      {/* Footer Actions */}
      <div className="sticky bottom-0 z-10 bg-white dark:bg-card-dark border-t border-border-light dark:border-border-dark p-6 flex flex-col-reverse sm:flex-row items-center justify-end gap-3">
        <Button
          className="w-full sm:w-auto"
          onClick={handleSaveUserProfile}
          disabled={updateUser.isPending}
        >
          {updateUser.isPending ? 'Menyimpan...' : 'Simpan Perubahan'}
        </Button>
      </div>
    </div>
  );

  const renderContent = () => {
    switch (activeSection) {
      case 'user-profile':
        return renderUserProfile();
      case 'users':
        return <UserManagement />;
      case 'business':
        return <BusinessSettings />;
      case 'devices':
        return <ScannerDevicesSettings />;
      case 'notifications':
        return <NotificationSettings />;
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

      {/* Mobile Navigation — horizontally scrollable so every section is reachable */}
      <div className="lg:hidden fixed bottom-0 left-0 right-0 z-50 bg-white dark:bg-card-dark border-t border-border-light dark:border-border-dark p-2">
        <div className="flex gap-1 overflow-x-auto scrollbar-hide">
          {settingsSections.map((section) => (
            <button
              key={section.id}
              onClick={() => setActiveSection(section.id)}
              className={`flex flex-col items-center gap-1 p-2 rounded-lg shrink-0 min-w-[68px] ${
                activeSection === section.id
                  ? 'text-primary bg-primary/5'
                  : 'text-text-sub'
              }`}
            >
              <Icon name={section.icon} size={20} />
              <span className="text-[10px] font-medium whitespace-nowrap">
                {section.label.split(' ')[0]}
              </span>
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

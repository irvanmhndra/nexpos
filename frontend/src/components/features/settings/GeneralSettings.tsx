import { useState } from 'react';
import { Button, Icon } from '@/components/ui';
import { useThemeStore } from '@/store/themeStore';

type DateFormat = 'DD/MM/YYYY' | 'YYYY-MM-DD' | 'MM/DD/YYYY';

export function GeneralSettings() {
  const { theme, setTheme } = useThemeStore();

  const [language, setLanguage] = useState('id');
  const [timezone, setTimezone] = useState('WIB');
  const [dateFormat, setDateFormat] = useState<DateFormat>('DD/MM/YYYY');
  const [isSaving, setIsSaving] = useState(false);

  const [passwords, setPasswords] = useState({
    current: '',
    new: '',
    confirm: '',
  });
  const [isChangingPassword, setIsChangingPassword] = useState(false);
  const [showPasswords, setShowPasswords] = useState({
    current: false,
    new: false,
    confirm: false,
  });

  const handleSave = async () => {
    setIsSaving(true);
    await new Promise(resolve => setTimeout(resolve, 1000));
    setIsSaving(false);
  };

  const handleChangePassword = async () => {
    if (!passwords.current || !passwords.new || !passwords.confirm) return;
    if (passwords.new !== passwords.confirm) return;
    setIsChangingPassword(true);
    await new Promise(resolve => setTimeout(resolve, 1000));
    setIsChangingPassword(false);
    setPasswords({ current: '', new: '', confirm: '' });
  };

  const themeOptions = [
    {
      id: 'light' as const,
      label: 'Terang',
      description: 'Tampilan terang untuk penggunaan siang hari',
      icon: 'light_mode',
    },
    {
      id: 'dark' as const,
      label: 'Gelap',
      description: 'Tampilan gelap yang nyaman di malam hari',
      icon: 'dark_mode',
    },
    {
      id: 'system' as const,
      label: 'Ikuti Sistem',
      description: 'Menyesuaikan dengan pengaturan perangkat',
      icon: 'devices',
    },
  ];

  const dateFormats: { value: DateFormat; example: string }[] = [
    { value: 'DD/MM/YYYY', example: '17/02/2026' },
    { value: 'YYYY-MM-DD', example: '2026-02-17' },
    { value: 'MM/DD/YYYY', example: '02/17/2026' },
  ];

  return (
    <div className="flex flex-col gap-6">
      {/* Preferences Card */}
      <div className="flex flex-col bg-white dark:bg-card-dark rounded-xl shadow-sm border border-border-light dark:border-border-dark overflow-hidden">
        {/* Header */}
        <div className="flex flex-col gap-2 p-6 md:p-8 border-b border-border-light dark:border-border-dark">
          <h2 className="text-2xl md:text-[32px] font-bold text-text-main dark:text-white">
            Pengaturan Umum
          </h2>
          <p className="text-sm md:text-base text-text-sub dark:text-gray-400">
            Atur preferensi tampilan, bahasa, dan regional sesuai kebutuhan Anda
          </p>
        </div>

        <div className="flex flex-col divide-y divide-border-light dark:divide-border-dark">
          {/* Language & Timezone */}
          <div className="flex flex-col gap-6 p-6 md:p-8">
            <h3 className="text-base font-bold text-text-main dark:text-white">Regional</h3>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              <div className="flex flex-col gap-2">
                <label className="text-sm font-bold text-text-main dark:text-gray-200">
                  Bahasa
                </label>
                <select
                  value={language}
                  onChange={e => setLanguage(e.target.value)}
                  className="rounded-lg border border-border-light dark:border-gray-600 bg-white dark:bg-card-dark text-text-main dark:text-white focus:border-primary focus:ring-1 focus:ring-primary h-12 px-4 text-base"
                >
                  <option value="id">Bahasa Indonesia</option>
                  <option value="en">English</option>
                </select>
              </div>
              <div className="flex flex-col gap-2">
                <label className="text-sm font-bold text-text-main dark:text-gray-200">
                  Zona Waktu
                </label>
                <select
                  value={timezone}
                  onChange={e => setTimezone(e.target.value)}
                  className="rounded-lg border border-border-light dark:border-gray-600 bg-white dark:bg-card-dark text-text-main dark:text-white focus:border-primary focus:ring-1 focus:ring-primary h-12 px-4 text-base"
                >
                  <option value="WIB">WIB – Waktu Indonesia Barat (UTC+7)</option>
                  <option value="WITA">WITA – Waktu Indonesia Tengah (UTC+8)</option>
                  <option value="WIT">WIT – Waktu Indonesia Timur (UTC+9)</option>
                  <option value="UTC">UTC – Coordinated Universal Time</option>
                </select>
              </div>
            </div>
          </div>

          {/* Date Format */}
          <div className="flex flex-col gap-4 p-6 md:p-8">
            <h3 className="text-base font-bold text-text-main dark:text-white">Format Tanggal</h3>
            <div className="flex flex-col sm:flex-row gap-3">
              {dateFormats.map(fmt => (
                <label
                  key={fmt.value}
                  className={`flex items-center gap-3 flex-1 p-4 rounded-lg border-2 cursor-pointer transition-colors ${
                    dateFormat === fmt.value
                      ? 'border-primary bg-primary/5'
                      : 'border-border-light dark:border-border-dark hover:border-primary/40'
                  }`}
                >
                  <input
                    type="radio"
                    name="dateFormat"
                    value={fmt.value}
                    checked={dateFormat === fmt.value}
                    onChange={() => setDateFormat(fmt.value)}
                    className="sr-only"
                  />
                  <div
                    className={`size-4 rounded-full border-2 flex items-center justify-center flex-shrink-0 ${
                      dateFormat === fmt.value
                        ? 'border-primary'
                        : 'border-border-light dark:border-gray-600'
                    }`}
                  >
                    {dateFormat === fmt.value && (
                      <div className="size-2 rounded-full bg-primary" />
                    )}
                  </div>
                  <div className="flex flex-col gap-0.5">
                    <span className="text-sm font-semibold text-text-main dark:text-white font-mono">
                      {fmt.value}
                    </span>
                    <span className="text-xs text-text-sub dark:text-gray-400">{fmt.example}</span>
                  </div>
                </label>
              ))}
            </div>
          </div>

          {/* Theme */}
          <div className="flex flex-col gap-4 p-6 md:p-8">
            <h3 className="text-base font-bold text-text-main dark:text-white">Tema Tampilan</h3>
            <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
              {themeOptions.map(option => (
                <button
                  key={option.id}
                  onClick={() => setTheme(option.id)}
                  className={`flex flex-col gap-3 p-4 rounded-xl border-2 text-left transition-colors ${
                    theme === option.id
                      ? 'border-primary bg-primary/5'
                      : 'border-border-light dark:border-border-dark hover:border-primary/40'
                  }`}
                >
                  <div
                    className={`size-10 rounded-lg flex items-center justify-center ${
                      theme === option.id
                        ? 'bg-primary/10'
                        : 'bg-background-light dark:bg-background-dark'
                    }`}
                  >
                    <Icon
                      name={option.icon}
                      size={22}
                      className={theme === option.id ? 'text-primary' : 'text-text-sub dark:text-gray-400'}
                    />
                  </div>
                  <div className="flex flex-col gap-0.5">
                    <span
                      className={`text-sm font-bold ${
                        theme === option.id
                          ? 'text-primary'
                          : 'text-text-main dark:text-white'
                      }`}
                    >
                      {option.label}
                    </span>
                    <span className="text-xs text-text-sub dark:text-gray-400 leading-relaxed">
                      {option.description}
                    </span>
                  </div>
                  {theme === option.id && (
                    <div className="flex items-center gap-1 mt-auto">
                      <Icon name="check_circle" size={14} className="text-primary" filled />
                      <span className="text-xs text-primary font-semibold">Aktif</span>
                    </div>
                  )}
                </button>
              ))}
            </div>
          </div>
        </div>

        {/* Save Footer */}
        <div className="sticky bottom-0 z-10 bg-white dark:bg-card-dark border-t border-border-light dark:border-border-dark p-6 flex flex-col-reverse sm:flex-row items-center justify-end gap-3">
          <Button variant="outline" className="w-full sm:w-auto">
            Batal
          </Button>
          <Button className="w-full sm:w-auto" onClick={handleSave} disabled={isSaving}>
            {isSaving ? 'Menyimpan...' : 'Simpan Perubahan'}
          </Button>
        </div>
      </div>

      {/* Change Password Card */}
      <div className="flex flex-col bg-white dark:bg-card-dark rounded-xl shadow-sm border border-border-light dark:border-border-dark overflow-hidden">
        <div className="flex flex-col gap-2 p-6 md:p-8 border-b border-border-light dark:border-border-dark">
          <h3 className="text-lg font-bold text-text-main dark:text-white">Ubah Kata Sandi</h3>
          <p className="text-sm text-text-sub dark:text-gray-400">
            Perbarui kata sandi untuk menjaga keamanan akun Anda
          </p>
        </div>

        <div className="flex flex-col gap-6 p-6 md:p-8">
          <div className="flex flex-col gap-2 max-w-md">
            <label className="text-sm font-bold text-text-main dark:text-gray-200">
              Kata Sandi Saat Ini
            </label>
            <div className="relative">
              <input
                type={showPasswords.current ? 'text' : 'password'}
                value={passwords.current}
                onChange={e => setPasswords(p => ({ ...p, current: e.target.value }))}
                placeholder="Masukkan kata sandi saat ini"
                className="w-full rounded-lg border border-border-light dark:border-gray-600 bg-white dark:bg-card-dark text-text-main dark:text-white focus:border-primary focus:ring-1 focus:ring-primary h-12 px-4 pr-12 text-base placeholder:text-text-sub dark:placeholder:text-gray-500"
              />
              <button
                type="button"
                onClick={() => setShowPasswords(p => ({ ...p, current: !p.current }))}
                className="absolute right-3 top-1/2 -translate-y-1/2 text-text-sub dark:text-gray-400 hover:text-text-main dark:hover:text-white"
              >
                <Icon name={showPasswords.current ? 'visibility_off' : 'visibility'} size={20} />
              </button>
            </div>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-6 max-w-2xl">
            <div className="flex flex-col gap-2">
              <label className="text-sm font-bold text-text-main dark:text-gray-200">
                Kata Sandi Baru
              </label>
              <div className="relative">
                <input
                  type={showPasswords.new ? 'text' : 'password'}
                  value={passwords.new}
                  onChange={e => setPasswords(p => ({ ...p, new: e.target.value }))}
                  placeholder="Minimal 8 karakter"
                  className="w-full rounded-lg border border-border-light dark:border-gray-600 bg-white dark:bg-card-dark text-text-main dark:text-white focus:border-primary focus:ring-1 focus:ring-primary h-12 px-4 pr-12 text-base placeholder:text-text-sub dark:placeholder:text-gray-500"
                />
                <button
                  type="button"
                  onClick={() => setShowPasswords(p => ({ ...p, new: !p.new }))}
                  className="absolute right-3 top-1/2 -translate-y-1/2 text-text-sub dark:text-gray-400 hover:text-text-main dark:hover:text-white"
                >
                  <Icon name={showPasswords.new ? 'visibility_off' : 'visibility'} size={20} />
                </button>
              </div>
            </div>
            <div className="flex flex-col gap-2">
              <label className="text-sm font-bold text-text-main dark:text-gray-200">
                Konfirmasi Kata Sandi Baru
              </label>
              <div className="relative">
                <input
                  type={showPasswords.confirm ? 'text' : 'password'}
                  value={passwords.confirm}
                  onChange={e => setPasswords(p => ({ ...p, confirm: e.target.value }))}
                  placeholder="Ulangi kata sandi baru"
                  className={`w-full rounded-lg border bg-white dark:bg-card-dark text-text-main dark:text-white focus:ring-1 focus:ring-primary h-12 px-4 pr-12 text-base placeholder:text-text-sub dark:placeholder:text-gray-500 ${
                    passwords.confirm && passwords.new !== passwords.confirm
                      ? 'border-red-400 dark:border-red-600 focus:border-red-400'
                      : 'border-border-light dark:border-gray-600 focus:border-primary'
                  }`}
                />
                <button
                  type="button"
                  onClick={() => setShowPasswords(p => ({ ...p, confirm: !p.confirm }))}
                  className="absolute right-3 top-1/2 -translate-y-1/2 text-text-sub dark:text-gray-400 hover:text-text-main dark:hover:text-white"
                >
                  <Icon name={showPasswords.confirm ? 'visibility_off' : 'visibility'} size={20} />
                </button>
              </div>
              {passwords.confirm && passwords.new !== passwords.confirm && (
                <p className="text-xs text-red-500 dark:text-red-400">
                  Kata sandi tidak cocok
                </p>
              )}
            </div>
          </div>
        </div>

        <div className="p-6 md:p-8 pt-0 flex justify-start">
          <Button
            onClick={handleChangePassword}
            disabled={
              isChangingPassword ||
              !passwords.current ||
              !passwords.new ||
              !passwords.confirm ||
              passwords.new !== passwords.confirm
            }
            leftIcon={<Icon name="lock_reset" size={16} />}
          >
            {isChangingPassword ? 'Memperbarui...' : 'Perbarui Kata Sandi'}
          </Button>
        </div>
      </div>
    </div>
  );
}

import { useState, useEffect } from 'react';
import { Button, Icon } from '@/components/ui';
import { useCompanySettings, useUpdateCompanySettings } from '@/hooks/useCompanySettings';

function ToggleSwitch({
  checked,
  onChange,
  disabled,
}: {
  checked: boolean;
  onChange: (v: boolean) => void;
  disabled?: boolean;
}) {
  return (
    <button
      type="button"
      role="switch"
      aria-checked={checked}
      disabled={disabled}
      onClick={() => onChange(!checked)}
      className={`relative inline-flex h-6 w-11 items-center rounded-full transition-colors focus:outline-none focus:ring-2 focus:ring-primary/30 ${
        checked ? 'bg-primary' : 'bg-gray-300 dark:bg-gray-600'
      } ${disabled ? 'opacity-50 cursor-not-allowed' : 'cursor-pointer'}`}
    >
      <span
        className={`inline-block h-4 w-4 transform rounded-full bg-white shadow transition-transform ${
          checked ? 'translate-x-6' : 'translate-x-1'
        }`}
      />
    </button>
  );
}

function SettingRow({
  label,
  description,
  children,
}: {
  label: string;
  description?: string;
  children: React.ReactNode;
}) {
  return (
    <div className="flex items-start justify-between gap-6 py-4">
      <div className="flex-1">
        <p className="text-sm font-semibold text-text-main dark:text-white">{label}</p>
        {description && (
          <p className="text-xs text-text-sub dark:text-gray-400 mt-0.5">{description}</p>
        )}
      </div>
      <div className="flex-shrink-0 flex items-center">{children}</div>
    </div>
  );
}

function SectionHeader({ title, description }: { title: string; description: string }) {
  return (
    <div className="pb-3 border-b border-border-light dark:border-border-dark">
      <h3 className="text-base font-bold text-text-main dark:text-white">{title}</h3>
      <p className="text-sm text-text-sub dark:text-gray-400 mt-0.5">{description}</p>
    </div>
  );
}

export function BusinessSettings() {
  const { data: settings, isLoading } = useCompanySettings();
  const updateSettings = useUpdateCompanySettings();

  const [form, setForm] = useState({
    tax_enabled: false,
    tax_rate: 11,
    tax_inclusive: true,
    rounding_enabled: false,
    rounding_amount: 100,
    auto_complete_counter_orders: true,
    require_customer_for_delivery: true,
    receipt_header: '',
    receipt_footer: '',
    show_tax_on_receipt: true,
    offline_mode_enabled: false,
    max_offline_days: 7,
  });

  const [isDirty, setIsDirty] = useState(false);

  // Populate form when settings load
  useEffect(() => {
    if (settings) {
      // eslint-disable-next-line react-hooks/set-state-in-effect
      setForm({
        tax_enabled: settings.tax_enabled,
        tax_rate: settings.tax_rate,
        tax_inclusive: settings.tax_inclusive,
        rounding_enabled: settings.rounding_enabled,
        rounding_amount: settings.rounding_amount,
        auto_complete_counter_orders: settings.auto_complete_counter_orders,
        require_customer_for_delivery: settings.require_customer_for_delivery,
        receipt_header: settings.receipt_header ?? '',
        receipt_footer: settings.receipt_footer ?? '',
        show_tax_on_receipt: settings.show_tax_on_receipt,
        offline_mode_enabled: settings.offline_mode_enabled,
        max_offline_days: settings.max_offline_days,
      });
       
      setIsDirty(false);
    }
  }, [settings]);

  const update = (patch: Partial<typeof form>) => {
    setForm(prev => ({ ...prev, ...patch }));
    setIsDirty(true);
  };

  const handleSave = async () => {
    await updateSettings.mutateAsync({
      ...form,
      receipt_header: form.receipt_header || null,
      receipt_footer: form.receipt_footer || null,
    });
    setIsDirty(false);
  };

  const handleDiscard = () => {
    if (settings) {
      setForm({
        tax_enabled: settings.tax_enabled,
        tax_rate: settings.tax_rate,
        tax_inclusive: settings.tax_inclusive,
        rounding_enabled: settings.rounding_enabled,
        rounding_amount: settings.rounding_amount,
        auto_complete_counter_orders: settings.auto_complete_counter_orders,
        require_customer_for_delivery: settings.require_customer_for_delivery,
        receipt_header: settings.receipt_header ?? '',
        receipt_footer: settings.receipt_footer ?? '',
        show_tax_on_receipt: settings.show_tax_on_receipt,
        offline_mode_enabled: settings.offline_mode_enabled,
        max_offline_days: settings.max_offline_days,
      });
    }
    setIsDirty(false);
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-24">
        <Icon name="progress_activity" size={32} className="animate-spin text-primary" />
      </div>
    );
  }

  return (
    <div className="flex flex-col gap-6">
      {/* Tax Configuration */}
      <div className="flex flex-col gap-4 bg-white dark:bg-card-dark rounded-xl shadow-sm border border-border-light dark:border-border-dark p-6 md:p-8">
        <SectionHeader
          title="Pajak"
          description="Konfigurasi PPN dan cara pajak diterapkan pada transaksi"
        />

        <div className="flex flex-col divide-y divide-border-light dark:divide-border-dark">
          <SettingRow
            label="Aktifkan Pajak (PPN)"
            description="Terapkan pajak secara otomatis pada setiap transaksi"
          >
            <ToggleSwitch
              checked={form.tax_enabled}
              onChange={v => update({ tax_enabled: v })}
            />
          </SettingRow>

          {form.tax_enabled && (
            <>
              <SettingRow
                label="Tarif Pajak (%)"
                description="Persentase pajak yang dikenakan (misal: 11 untuk PPN 11%)"
              >
                <div className="flex items-center gap-2">
                  <input
                    type="number"
                    min="0"
                    max="100"
                    step="0.1"
                    value={form.tax_rate}
                    onChange={e => update({ tax_rate: parseFloat(e.target.value) || 0 })}
                    className="w-20 h-10 px-3 rounded-lg border border-border-light dark:border-gray-600 bg-white dark:bg-card-dark text-text-main dark:text-white text-sm text-center focus:outline-none focus:ring-2 focus:ring-primary/30 focus:border-primary"
                  />
                  <span className="text-sm text-text-sub dark:text-gray-400">%</span>
                </div>
              </SettingRow>

              <SettingRow
                label="Pajak Sudah Termasuk Harga"
                description="Harga produk sudah termasuk pajak (tax inclusive)"
              >
                <ToggleSwitch
                  checked={form.tax_inclusive}
                  onChange={v => update({ tax_inclusive: v })}
                />
              </SettingRow>

              <SettingRow
                label="Tampilkan Pajak di Struk"
                description="Cetak rincian pajak pada struk pelanggan"
              >
                <ToggleSwitch
                  checked={form.show_tax_on_receipt}
                  onChange={v => update({ show_tax_on_receipt: v })}
                />
              </SettingRow>
            </>
          )}
        </div>
      </div>

      {/* Rounding */}
      <div className="flex flex-col gap-4 bg-white dark:bg-card-dark rounded-xl shadow-sm border border-border-light dark:border-border-dark p-6 md:p-8">
        <SectionHeader
          title="Pembulatan"
          description="Bulatkan total transaksi ke kelipatan tertentu"
        />

        <div className="flex flex-col divide-y divide-border-light dark:divide-border-dark">
          <SettingRow
            label="Aktifkan Pembulatan"
            description="Bulatkan grand total ke kelipatan yang ditentukan"
          >
            <ToggleSwitch
              checked={form.rounding_enabled}
              onChange={v => update({ rounding_enabled: v })}
            />
          </SettingRow>

          {form.rounding_enabled && (
            <SettingRow
              label="Kelipatan Pembulatan (Rp)"
              description="Total akan dibulatkan ke kelipatan angka ini (misal: 100 → Rp 15.300 menjadi Rp 15.300)"
            >
              <div className="flex items-center gap-2">
                <span className="text-sm text-text-sub dark:text-gray-400">Rp</span>
                <input
                  type="number"
                  min="1"
                  step="1"
                  value={form.rounding_amount}
                  onChange={e => update({ rounding_amount: parseInt(e.target.value) || 1 })}
                  className="w-24 h-10 px-3 rounded-lg border border-border-light dark:border-gray-600 bg-white dark:bg-card-dark text-text-main dark:text-white text-sm text-center focus:outline-none focus:ring-2 focus:ring-primary/30 focus:border-primary"
                />
              </div>
            </SettingRow>
          )}
        </div>
      </div>

      {/* Order Settings */}
      <div className="flex flex-col gap-4 bg-white dark:bg-card-dark rounded-xl shadow-sm border border-border-light dark:border-border-dark p-6 md:p-8">
        <SectionHeader
          title="Pengaturan Order"
          description="Konfigurasi alur pemrosesan order dan pengiriman"
        />

        <div className="flex flex-col divide-y divide-border-light dark:divide-border-dark">
          <SettingRow
            label="Auto-Complete Order Kasir"
            description="Order di kasir langsung diselesaikan otomatis setelah pembayaran penuh"
          >
            <ToggleSwitch
              checked={form.auto_complete_counter_orders}
              onChange={v => update({ auto_complete_counter_orders: v })}
            />
          </SettingRow>

          <SettingRow
            label="Wajib Data Pelanggan untuk Pengiriman"
            description="Order dengan tipe delivery harus memilih pelanggan"
          >
            <ToggleSwitch
              checked={form.require_customer_for_delivery}
              onChange={v => update({ require_customer_for_delivery: v })}
            />
          </SettingRow>
        </div>
      </div>

      {/* Receipt Settings */}
      <div className="flex flex-col gap-4 bg-white dark:bg-card-dark rounded-xl shadow-sm border border-border-light dark:border-border-dark p-6 md:p-8">
        <SectionHeader
          title="Struk / Nota"
          description="Kustomisasi teks header dan footer pada struk pelanggan"
        />

        <div className="flex flex-col gap-4">
          <div className="flex flex-col gap-2">
            <label className="text-sm font-semibold text-text-main dark:text-gray-200">
              Header Struk
            </label>
            <textarea
              value={form.receipt_header}
              onChange={e => update({ receipt_header: e.target.value })}
              placeholder="Contoh: Selamat berbelanja di toko kami!"
              rows={3}
              className="w-full rounded-lg border border-border-light dark:border-gray-600 bg-white dark:bg-card-dark text-text-main dark:text-white focus:border-primary focus:ring-1 focus:ring-primary p-3 text-sm placeholder:text-text-sub dark:placeholder:text-gray-500 resize-y"
            />
            <p className="text-xs text-text-sub dark:text-gray-400">
              Ditampilkan di bagian atas struk
            </p>
          </div>

          <div className="flex flex-col gap-2">
            <label className="text-sm font-semibold text-text-main dark:text-gray-200">
              Footer Struk
            </label>
            <textarea
              value={form.receipt_footer}
              onChange={e => update({ receipt_footer: e.target.value })}
              placeholder="Contoh: Terima kasih atas kunjungan Anda. Barang yang sudah dibeli tidak dapat dikembalikan."
              rows={3}
              className="w-full rounded-lg border border-border-light dark:border-gray-600 bg-white dark:bg-card-dark text-text-main dark:text-white focus:border-primary focus:ring-1 focus:ring-primary p-3 text-sm placeholder:text-text-sub dark:placeholder:text-gray-500 resize-y"
            />
            <p className="text-xs text-text-sub dark:text-gray-400">
              Ditampilkan di bagian bawah struk
            </p>
          </div>
        </div>
      </div>

      {/* Sticky footer */}
      {isDirty && (
        <div className="sticky bottom-0 z-10 bg-white dark:bg-card-dark border border-border-light dark:border-border-dark rounded-xl shadow-lg p-4 flex items-center justify-between gap-4">
          <p className="text-sm text-text-sub dark:text-gray-400">
            Ada perubahan yang belum disimpan
          </p>
          <div className="flex items-center gap-3">
            <Button variant="outline" onClick={handleDiscard} disabled={updateSettings.isPending}>
              Batal
            </Button>
            <Button onClick={handleSave} disabled={updateSettings.isPending}>
              {updateSettings.isPending ? 'Menyimpan...' : 'Simpan Perubahan'}
            </Button>
          </div>
        </div>
      )}
    </div>
  );
}

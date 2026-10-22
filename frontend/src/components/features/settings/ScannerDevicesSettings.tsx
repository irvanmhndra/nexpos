import { useCallback, useEffect, useRef, useState } from 'react';
import { Button, Icon } from '@/components/ui';
import { BarcodeScannerModal } from '@/components/features/pos/BarcodeScannerModal';
import { useBarcodeScanner } from '@/hooks/useBarcodeScanner';

// --- Minimal WebHID typings (not in the default DOM lib) -------------------
interface HidLikeDevice {
  productName?: string;
  vendorId?: number;
  productId?: number;
  opened?: boolean;
}
interface HidApi {
  getDevices(): Promise<HidLikeDevice[]>;
  requestDevice(options: { filters: unknown[] }): Promise<HidLikeDevice[]>;
  addEventListener(type: 'connect' | 'disconnect', listener: () => void): void;
  removeEventListener(type: 'connect' | 'disconnect', listener: () => void): void;
}
function getHid(): HidApi | null {
  if (typeof navigator === 'undefined') return null;
  const hid = (navigator as unknown as { hid?: HidApi }).hid;
  return hid ?? null;
}

type PermState = 'granted' | 'denied' | 'prompt' | 'unknown';

function StatusDot({ tone }: { tone: 'green' | 'amber' | 'gray' | 'red' }) {
  const color =
    tone === 'green'
      ? 'bg-green-500'
      : tone === 'amber'
        ? 'bg-amber-500'
        : tone === 'red'
          ? 'bg-red-500'
          : 'bg-gray-400';
  return (
    <span className="relative flex h-2.5 w-2.5">
      {tone === 'green' && (
        <span className="absolute inline-flex h-full w-full animate-ping rounded-full bg-green-400 opacity-75" />
      )}
      <span className={`relative inline-flex h-2.5 w-2.5 rounded-full ${color}`} />
    </span>
  );
}

function isBackCamera(label: string) {
  const l = label.toLowerCase();
  return l.includes('back') || l.includes('rear') || l.includes('environment');
}

function formatTime(ts: number) {
  return new Date(ts).toLocaleTimeString('id-ID', {
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
  });
}

export function ScannerDevicesSettings() {
  // --- Cameras -------------------------------------------------------------
  const cameraSupported =
    typeof navigator !== 'undefined' && !!navigator.mediaDevices?.enumerateDevices;
  const [cameras, setCameras] = useState<MediaDeviceInfo[]>([]);
  const [cameraPermission, setCameraPermission] = useState<PermState>('unknown');
  const [requestingCamera, setRequestingCamera] = useState(false);
  const [isTestScanOpen, setIsTestScanOpen] = useState(false);
  const [cameraTestResult, setCameraTestResult] = useState<{ code: string; at: number } | null>(
    null,
  );

  const enumerateCameras = useCallback(async () => {
    if (!cameraSupported) return;
    try {
      const devices = await navigator.mediaDevices.enumerateDevices();
      setCameras(devices.filter(d => d.kind === 'videoinput'));
    } catch {
      // ignore — listing can fail on insecure origins
    }
  }, [cameraSupported]);

  useEffect(() => {
    if (!cameraSupported) return;
    enumerateCameras();

    // Best-effort permission state (not supported in every browser).
    let permResult: PermissionStatus | null = null;
    navigator.permissions
      ?.query({ name: 'camera' as PermissionName })
      .then(result => {
        setCameraPermission(result.state as PermState);
        result.onchange = () => setCameraPermission(result.state as PermState);
        permResult = result;
      })
      .catch(() => setCameraPermission('unknown'));

    const onChange = () => enumerateCameras();
    navigator.mediaDevices.addEventListener?.('devicechange', onChange);
    return () => {
      navigator.mediaDevices.removeEventListener?.('devicechange', onChange);
      if (permResult) permResult.onchange = null;
    };
  }, [cameraSupported, enumerateCameras]);

  const requestCameraAccess = async () => {
    setRequestingCamera(true);
    try {
      const stream = await navigator.mediaDevices.getUserMedia({ video: true });
      stream.getTracks().forEach(t => t.stop());
      setCameraPermission('granted');
      await enumerateCameras(); // labels populate once permission is granted
    } catch {
      setCameraPermission('denied');
    } finally {
      setRequestingCamera(false);
    }
  };

  // --- USB / Bluetooth scanner (keyboard wedge) ----------------------------
  const [wedgeScans, setWedgeScans] = useState<{ code: string; at: number }[]>([]);
  const [now, setNow] = useState(Date.now());
  const handleWedgeScan = useCallback((code: string) => {
    setWedgeScans(prev => [{ code, at: Date.now() }, ...prev].slice(0, 5));
  }, []);
  useBarcodeScanner({ onScan: handleWedgeScan, enabled: true });

  // Heartbeat so "x detik lalu" and the live dot stay fresh.
  useEffect(() => {
    const id = setInterval(() => setNow(Date.now()), 2000);
    return () => clearInterval(id);
  }, []);

  const lastWedge = wedgeScans[0];
  const wedgeActive = !!lastWedge && now - lastWedge.at < 15000;

  // --- WebHID (HID-POS mode scanners) --------------------------------------
  const hidApiRef = useRef<HidApi | null>(getHid());
  const hidSupported = !!hidApiRef.current;
  const [hidDevices, setHidDevices] = useState<HidLikeDevice[]>([]);

  const refreshHid = useCallback(async () => {
    const hid = hidApiRef.current;
    if (!hid) return;
    try {
      setHidDevices(await hid.getDevices());
    } catch {
      // ignore
    }
  }, []);

  useEffect(() => {
    const hid = hidApiRef.current;
    if (!hid) return;
    refreshHid();
    const onChange = () => refreshHid();
    hid.addEventListener('connect', onChange);
    hid.addEventListener('disconnect', onChange);
    return () => {
      hid.removeEventListener('connect', onChange);
      hid.removeEventListener('disconnect', onChange);
    };
  }, [refreshHid]);

  const connectHid = async () => {
    const hid = hidApiRef.current;
    if (!hid) return;
    try {
      await hid.requestDevice({ filters: [] });
      await refreshHid();
    } catch {
      // user dismissed the chooser or no device selected
    }
  };

  const cameraReady = cameraSupported && cameras.length > 0 && cameraPermission === 'granted';

  return (
    <div className="flex flex-col bg-white dark:bg-card-dark rounded-xl shadow-sm border border-border-light dark:border-border-dark overflow-hidden">
      {/* Header */}
      <div className="flex flex-col gap-2 p-6 md:p-8 border-b border-border-light dark:border-border-dark">
        <div className="flex items-center gap-3">
          <Icon name="barcode_scanner" size={24} className="text-primary" />
          <h2 className="text-xl font-bold text-text-main dark:text-white">Perangkat Pemindai</h2>
        </div>
        <p className="text-sm text-text-sub dark:text-gray-400">
          Daftar perangkat scanner barcode yang terhubung beserta statusnya. Kamera dapat dideteksi
          langsung; scanner USB/Bluetooth tipe keyboard terdeteksi otomatis saat memindai.
        </p>
      </div>

      <div className="p-6 md:p-8 space-y-8">
        {/* ============================ CAMERAS ============================ */}
        <section>
          <div className="flex items-center justify-between gap-3 mb-3">
            <div className="flex items-center gap-2">
              <Icon name="photo_camera" size={20} className="text-text-sub dark:text-gray-400" />
              <h3 className="font-semibold text-text-main dark:text-white">Kamera</h3>
              <span className="text-xs text-text-sub dark:text-gray-400">
                ({cameras.length} terdeteksi)
              </span>
            </div>
            <div className="flex items-center gap-2">
              {cameraReady && (
                <Button
                  variant="outline"
                  size="sm"
                  leftIcon={<Icon name="qr_code_scanner" size={16} />}
                  onClick={() => setIsTestScanOpen(true)}
                >
                  Tes Pindai
                </Button>
              )}
              {cameraSupported && cameraPermission !== 'granted' && (
                <Button
                  size="sm"
                  leftIcon={<Icon name="videocam" size={16} />}
                  onClick={requestCameraAccess}
                  disabled={requestingCamera}
                >
                  {requestingCamera ? 'Meminta…' : 'Izinkan & Deteksi'}
                </Button>
              )}
            </div>
          </div>

          {!cameraSupported ? (
            <p className="text-sm text-text-sub dark:text-gray-400 p-4 rounded-lg bg-background-light dark:bg-background-dark">
              Browser ini tidak mendukung deteksi kamera (butuh koneksi HTTPS).
            </p>
          ) : cameraPermission === 'denied' ? (
            <div className="flex items-start gap-2 text-sm p-4 rounded-lg bg-red-50 dark:bg-red-900/20 text-red-700 dark:text-red-300">
              <Icon name="block" size={18} className="shrink-0 mt-0.5" />
              <span>Akses kamera ditolak. Izinkan kamera di pengaturan browser untuk mendeteksi perangkat.</span>
            </div>
          ) : cameras.length === 0 ? (
            <p className="text-sm text-text-sub dark:text-gray-400 p-4 rounded-lg bg-background-light dark:bg-background-dark">
              Belum ada kamera terdeteksi. Klik "Izinkan & Deteksi" untuk menampilkan kamera yang terpasang.
            </p>
          ) : (
            <ul className="space-y-2">
              {cameras.map((cam, i) => (
                <li
                  key={cam.deviceId || i}
                  className="flex items-center gap-3 p-3 rounded-lg border border-border-light dark:border-border-dark bg-background-light dark:bg-background-dark"
                >
                  <StatusDot tone="green" />
                  <Icon name="videocam" size={20} className="text-text-sub dark:text-gray-400" />
                  <div className="flex-1 min-w-0">
                    <p className="text-sm font-medium text-text-main dark:text-white truncate">
                      {cam.label || `Kamera ${i + 1}`}
                      {cam.label && isBackCamera(cam.label) && (
                        <span className="ml-2 text-xs font-normal text-text-sub dark:text-gray-400">
                          (belakang)
                        </span>
                      )}
                    </p>
                    {!cam.label && (
                      <p className="text-xs text-text-sub dark:text-gray-400">
                        Izinkan akses kamera untuk melihat nama perangkat
                      </p>
                    )}
                  </div>
                  <span className="text-xs font-medium text-green-600 dark:text-green-400">Tersedia</span>
                </li>
              ))}
            </ul>
          )}

          {cameraTestResult && (
            <div className="mt-3 flex items-center gap-2 text-sm p-3 rounded-lg bg-green-50 dark:bg-green-900/20 text-green-700 dark:text-green-300">
              <Icon name="check_circle" size={18} className="shrink-0" />
              <span>
                Tes berhasil — terbaca <span className="font-mono font-semibold">{cameraTestResult.code}</span>{' '}
                pada {formatTime(cameraTestResult.at)}
              </span>
            </div>
          )}
        </section>

        {/* ==================== USB / BLUETOOTH (WEDGE) ==================== */}
        <section>
          <div className="flex items-center gap-2 mb-3">
            <Icon name="barcode_scanner" size={20} className="text-text-sub dark:text-gray-400" />
            <h3 className="font-semibold text-text-main dark:text-white">Scanner USB / Bluetooth</h3>
          </div>

          <div className="p-4 rounded-lg border border-border-light dark:border-border-dark bg-background-light dark:bg-background-dark">
            <div className="flex items-center justify-between gap-3">
              <div className="flex items-center gap-3">
                <StatusDot tone={wedgeActive ? 'green' : lastWedge ? 'amber' : 'gray'} />
                <div>
                  <p className="text-sm font-medium text-text-main dark:text-white">
                    {wedgeActive
                      ? 'Aktif — scanner terdeteksi'
                      : lastWedge
                        ? 'Terhubung (idle)'
                        : 'Menunggu input scanner'}
                  </p>
                  <p className="text-xs text-text-sub dark:text-gray-400">
                    {lastWedge
                      ? `Pindai terakhir ${formatTime(lastWedge.at)}`
                      : 'Arahkan scanner ke barcode mana pun untuk menguji koneksi'}
                  </p>
                </div>
              </div>
              <Icon
                name="keyboard"
                size={22}
                className={wedgeActive ? 'text-green-500' : 'text-text-sub dark:text-gray-500'}
              />
            </div>

            {wedgeScans.length > 0 && (
              <div className="mt-3 pt-3 border-t border-border-light dark:border-border-dark">
                <p className="text-xs font-medium text-text-sub dark:text-gray-400 mb-1.5">
                  Pindai terbaru
                </p>
                <ul className="space-y-1">
                  {wedgeScans.map((s, i) => (
                    <li key={`${s.at}-${i}`} className="flex items-center justify-between text-sm">
                      <span className="font-mono text-text-main dark:text-gray-200">{s.code}</span>
                      <span className="text-xs text-text-sub dark:text-gray-400">{formatTime(s.at)}</span>
                    </li>
                  ))}
                </ul>
              </div>
            )}
          </div>

          <p className="text-xs text-text-sub dark:text-gray-400 mt-2 flex items-start gap-1.5">
            <Icon name="info" size={14} className="shrink-0 mt-0.5" />
            <span>
              Scanner USB/Bluetooth umumnya bekerja sebagai keyboard, sehingga browser tidak bisa
              menampilkannya satu per satu — statusnya terdeteksi dari aktivitas pemindaian.
            </span>
          </p>
        </section>

        {/* ========================= WEBHID (HID-POS) ===================== */}
        <section>
          <div className="flex items-center justify-between gap-3 mb-3">
            <div className="flex items-center gap-2">
              <Icon name="usb" size={20} className="text-text-sub dark:text-gray-400" />
              <h3 className="font-semibold text-text-main dark:text-white">
                Perangkat HID
                <span className="ml-1 text-xs font-normal text-text-sub dark:text-gray-400">
                  (mode HID-POS)
                </span>
              </h3>
            </div>
            {hidSupported && (
              <Button
                variant="outline"
                size="sm"
                leftIcon={<Icon name="add" size={16} />}
                onClick={connectHid}
              >
                Hubungkan
              </Button>
            )}
          </div>

          {!hidSupported ? (
            <p className="text-sm text-text-sub dark:text-gray-400 p-4 rounded-lg bg-background-light dark:bg-background-dark">
              Browser ini tidak mendukung WebHID (hanya Chrome/Edge di desktop). Untuk scanner mode
              keyboard, gunakan bagian di atas.
            </p>
          ) : hidDevices.length === 0 ? (
            <p className="text-sm text-text-sub dark:text-gray-400 p-4 rounded-lg bg-background-light dark:bg-background-dark">
              Belum ada perangkat HID terhubung. Klik "Hubungkan" untuk memilih scanner yang berjalan
              di mode HID-POS.
            </p>
          ) : (
            <ul className="space-y-2">
              {hidDevices.map((d, i) => (
                <li
                  key={`${d.vendorId}-${d.productId}-${i}`}
                  className="flex items-center gap-3 p-3 rounded-lg border border-border-light dark:border-border-dark bg-background-light dark:bg-background-dark"
                >
                  <StatusDot tone="green" />
                  <Icon name="usb" size={20} className="text-text-sub dark:text-gray-400" />
                  <div className="flex-1 min-w-0">
                    <p className="text-sm font-medium text-text-main dark:text-white truncate">
                      {d.productName || 'Perangkat HID'}
                    </p>
                    <p className="text-xs text-text-sub dark:text-gray-400 font-mono">
                      VID {(d.vendorId ?? 0).toString(16).padStart(4, '0')} · PID{' '}
                      {(d.productId ?? 0).toString(16).padStart(4, '0')}
                    </p>
                  </div>
                  <span className="text-xs font-medium text-green-600 dark:text-green-400">Terhubung</span>
                </li>
              ))}
            </ul>
          )}
        </section>
      </div>

      {/* Camera test scanner */}
      <BarcodeScannerModal
        isOpen={isTestScanOpen}
        onClose={() => setIsTestScanOpen(false)}
        onScan={code => setCameraTestResult({ code, at: Date.now() })}
      />
    </div>
  );
}

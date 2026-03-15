import { useEffect, useRef, useState } from 'react';
import { BrowserMultiFormatReader } from '@zxing/browser';
import type { IScannerControls } from '@zxing/browser';
import { NotFoundException } from '@zxing/library';
import { Button, Icon } from '@/components/ui';

interface Props {
  isOpen: boolean;
  onClose: () => void;
  onScan: (barcode: string) => void;
}

/**
 * Camera-based barcode scanner modal.
 * Uses @zxing/browser which supports Code128, EAN-13, QR, and more.
 */
export function BarcodeScannerModal({ isOpen, onClose, onScan }: Props) {
  const videoRef = useRef<HTMLVideoElement>(null);
  const controlsRef = useRef<IScannerControls | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [cameras, setCameras] = useState<MediaDeviceInfo[]>([]);
  const [activeCameraId, setActiveCameraId] = useState<string | undefined>();
  const [lastScanned, setLastScanned] = useState('');
  const lastScannedRef = useRef('');
  const lastScannedTimeRef = useRef(0);

  useEffect(() => {
    if (!isOpen) return;

    let active = true;

    const start = async () => {
      setError(null);
      setLastScanned('');
      lastScannedRef.current = '';

      try {
        const devices = await BrowserMultiFormatReader.listVideoInputDevices();
        if (!active) return;

        if (devices.length === 0) {
          setError('Kamera tidak ditemukan');
          return;
        }

        setCameras(devices);
        // Prefer back camera on mobile
        const back = devices.find(d =>
          d.label.toLowerCase().includes('back') ||
          d.label.toLowerCase().includes('rear') ||
          d.label.toLowerCase().includes('environment')
        );
        const selectedId = activeCameraId ?? back?.deviceId ?? devices[0].deviceId;
        setActiveCameraId(selectedId);

        const reader = new BrowserMultiFormatReader();
        const controls = await reader.decodeFromVideoDevice(selectedId, videoRef.current!, (result, err) => {
          if (!active) return;
          if (result) {
            const barcode = result.getText();
            const now = Date.now();
            // Debounce: ignore same barcode scanned within 2s
            if (barcode === lastScannedRef.current && now - lastScannedTimeRef.current < 2000) {
              return;
            }
            lastScannedRef.current = barcode;
            lastScannedTimeRef.current = now;
            setLastScanned(barcode);
            onScan(barcode);
            onClose();
          }
          if (err && !(err instanceof NotFoundException)) {
            // NotFoundException is normal (no barcode in frame), ignore
          }
        });

        if (active) {
          controlsRef.current = controls;
        } else {
          controls.stop();
        }
      } catch (e) {
        if (!active) return;
        const msg = e instanceof Error ? e.message : String(e);
        if (msg.toLowerCase().includes('permission') || msg.toLowerCase().includes('notallowed')) {
          setError('Akses kamera ditolak. Izinkan akses kamera di browser Anda.');
        } else {
          setError('Gagal mengakses kamera: ' + msg);
        }
      }
    };

    start();

    return () => {
      active = false;
      controlsRef.current?.stop();
      controlsRef.current = null;
    };
  }, [isOpen, activeCameraId]); // eslint-disable-line react-hooks/exhaustive-deps

  const handleCameraSwitch = (deviceId: string) => {
    controlsRef.current?.stop();
    controlsRef.current = null;
    setActiveCameraId(deviceId);
  };

  if (!isOpen) return null;

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/70"
      onClick={onClose}
    >
      <div
        className="bg-white dark:bg-card-dark rounded-2xl shadow-2xl w-full max-w-sm overflow-hidden"
        onClick={e => e.stopPropagation()}
      >
        {/* Header */}
        <div className="flex items-center justify-between px-4 py-3 border-b border-border-light dark:border-border-dark">
          <div className="flex items-center gap-2">
            <Icon name="qr_code_scanner" size={20} className="text-primary" />
            <h3 className="font-bold text-text-main dark:text-white">Scan Barcode</h3>
          </div>
          <button
            onClick={onClose}
            className="size-8 flex items-center justify-center rounded-lg hover:bg-background-light dark:hover:bg-background-dark"
          >
            <Icon name="close" size={20} className="text-text-sub" />
          </button>
        </div>

        {/* Camera feed */}
        <div className="relative bg-black aspect-square">
          <video
            ref={videoRef}
            className="w-full h-full object-cover"
            autoPlay
            playsInline
            muted
          />

          {/* Scan overlay */}
          {!error && (
            <div className="absolute inset-0 flex items-center justify-center pointer-events-none">
              <div className="relative w-48 h-48">
                {/* Corner brackets */}
                <div className="absolute top-0 left-0 w-8 h-8 border-t-4 border-l-4 border-primary rounded-tl-md" />
                <div className="absolute top-0 right-0 w-8 h-8 border-t-4 border-r-4 border-primary rounded-tr-md" />
                <div className="absolute bottom-0 left-0 w-8 h-8 border-b-4 border-l-4 border-primary rounded-bl-md" />
                <div className="absolute bottom-0 right-0 w-8 h-8 border-b-4 border-r-4 border-primary rounded-br-md" />
                {/* Scan line animation */}
                <div className="absolute left-2 right-2 h-0.5 bg-primary/80 animate-scan-line" />
              </div>
            </div>
          )}

          {/* Error overlay */}
          {error && (
            <div className="absolute inset-0 flex flex-col items-center justify-center p-6 bg-black/60">
              <Icon name="camera_off" size={40} className="text-white/60 mb-3" />
              <p className="text-white text-sm text-center">{error}</p>
            </div>
          )}

          {/* Last scanned feedback */}
          {lastScanned && (
            <div className="absolute bottom-2 left-2 right-2 bg-green-600 text-white text-xs font-bold px-3 py-1.5 rounded-lg text-center">
              ✓ {lastScanned}
            </div>
          )}
        </div>

        {/* Camera selector (if multiple cameras) */}
        {cameras.length > 1 && (
          <div className="px-4 py-2 border-t border-border-light dark:border-border-dark">
            <div className="flex gap-2 overflow-x-auto">
              {cameras.map(cam => (
                <button
                  key={cam.deviceId}
                  onClick={() => handleCameraSwitch(cam.deviceId)}
                  className={`px-3 py-1.5 rounded-lg text-xs font-medium whitespace-nowrap transition-colors ${
                    activeCameraId === cam.deviceId
                      ? 'bg-primary text-white'
                      : 'bg-background-light dark:bg-background-dark text-text-sub hover:text-primary'
                  }`}
                >
                  {cam.label || `Kamera ${cameras.indexOf(cam) + 1}`}
                </button>
              ))}
            </div>
          </div>
        )}

        {/* Footer */}
        <div className="px-4 py-3 border-t border-border-light dark:border-border-dark">
          <p className="text-xs text-text-sub dark:text-gray-400 text-center">
            Arahkan kamera ke barcode produk
          </p>
          <Button variant="outline" className="w-full mt-2" onClick={onClose}>
            Tutup
          </Button>
        </div>
      </div>
    </div>
  );
}

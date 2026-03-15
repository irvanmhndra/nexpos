import { useEffect, useRef } from 'react';

interface Options {
  onScan: (barcode: string) => void;
  enabled?: boolean;
  minLength?: number;
}

/**
 * Detects USB/Bluetooth barcode scanner input.
 *
 * Scanners act as keyboard emulators: they type every character very fast
 * (< 50ms between strokes) then send Enter. This hook distinguishes scanner
 * input from regular typing by tracking keystroke timing.
 *
 * Ignored when focus is inside an input/textarea/select element.
 */
export function useBarcodeScanner({ onScan, enabled = true, minLength = 3 }: Options) {
  const bufferRef = useRef('');
  const lastKeyTimeRef = useRef(0);
  // Scanners typically complete a full barcode in < 100ms total.
  // We allow up to 80ms between individual keystrokes.
  const THRESHOLD_MS = 80;

  useEffect(() => {
    if (!enabled) return;

    const handleKeyDown = (e: KeyboardEvent) => {
      // Don't intercept when user is typing in a form field
      const tag = (e.target as HTMLElement).tagName;
      if (tag === 'INPUT' || tag === 'TEXTAREA' || tag === 'SELECT') return;

      const now = Date.now();

      if (e.key === 'Enter') {
        const barcode = bufferRef.current.trim();
        if (barcode.length >= minLength) {
          onScan(barcode);
        }
        bufferRef.current = '';
        lastKeyTimeRef.current = 0;
        return;
      }

      // If gap since last key is too large → human typing, reset buffer
      if (lastKeyTimeRef.current > 0 && now - lastKeyTimeRef.current > THRESHOLD_MS) {
        bufferRef.current = '';
      }

      if (e.key.length === 1) {
        bufferRef.current += e.key;
        lastKeyTimeRef.current = now;
      }
    };

    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [enabled, onScan, minLength]);
}

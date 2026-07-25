import { API_BASE_URL } from '@/utils/constants';

const MAX_SIDE = 1000; // longest side (px) — cukup untuk grid POS & detail retina
const QUALITY = 0.8;
const MAX_INPUT_BYTES = 10 * 1024 * 1024;

/**
 * Compress an image client-side (resize + WebP) then upload it to the API, which
 * stores it in R2. Returns the public CDN URL. Re-encoding via canvas also
 * strips EXIF metadata.
 */
export async function uploadImage(file: File): Promise<string> {
  if (file.size > MAX_INPUT_BYTES) throw new Error('Ukuran file maksimal 10MB');
  const blob = await compress(file);
  const ext = blob.type === 'image/png' ? 'png' : blob.type === 'image/webp' ? 'webp' : 'jpg';

  const form = new FormData();
  form.append('file', blob, `image.${ext}`);
  const token = localStorage.getItem('nexpos_access_token');
  const res = await fetch(`${API_BASE_URL}/uploads`, {
    method: 'POST',
    headers: token ? { Authorization: `Bearer ${token}` } : {},
    body: form,
  });
  if (!res.ok) {
    const body = await res.json().catch(() => null);
    throw new Error(body?.error?.message ?? body?.message ?? 'Gagal mengunggah gambar');
  }
  const body = await res.json();
  return body.data.url as string;
}

/** Resize (longest side ≤ MAX_SIDE) + re-encode to WebP (JPEG fallback). */
async function compress(file: File): Promise<Blob> {
  const img = await loadImage(file);
  const scale = Math.min(1, MAX_SIDE / Math.max(img.width, img.height));
  const canvas = document.createElement('canvas');
  canvas.width = Math.round(img.width * scale);
  canvas.height = Math.round(img.height * scale);
  const ctx = canvas.getContext('2d');
  if (!ctx) return file;
  ctx.drawImage(img, 0, 0, canvas.width, canvas.height);
  const webp = await toBlob(canvas, 'image/webp', QUALITY);
  if (webp) return webp;
  const jpeg = await toBlob(canvas, 'image/jpeg', QUALITY);
  return jpeg ?? file;
}

function loadImage(file: File): Promise<HTMLImageElement> {
  return new Promise((resolve, reject) => {
    const url = URL.createObjectURL(file);
    const img = new Image();
    img.onload = () => {
      URL.revokeObjectURL(url);
      resolve(img);
    };
    img.onerror = () => {
      URL.revokeObjectURL(url);
      reject(new Error('Format gambar tidak didukung (coba JPG/PNG)'));
    };
    img.src = url;
  });
}

function toBlob(canvas: HTMLCanvasElement, type: string, quality: number): Promise<Blob | null> {
  return new Promise((resolve) => canvas.toBlob(resolve, type, quality));
}

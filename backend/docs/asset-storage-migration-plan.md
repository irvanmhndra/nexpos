# Asset Storage Migration Plan — base64-in-Postgres → Cloudflare R2

> Status: **DRAFT / for review** (belum dieksekusi). Disusun setelah brainstorming opsi
> object storage. Keputusan & alasan ada di bawah.

## 1. Decision summary

| | |
|---|---|
| **Target storage** | **Cloudflare R2** (S3-compatible object storage) |
| **Delivery** | Public bucket via custom domain **cdn.nexpos.irvanmahendra.com** (Cloudflare CDN cache, PoP Jakarta/SG) |
| **Transform** | **Resize/compress di sisi client** sebelum upload. Cloudflare Image Resizing = opsional, ditunda. |
| **Upload flow** | **Presigned PUT** — client upload langsung ke R2, API hanya menerbitkan URL bertanda tangan + menyimpan URL hasil |
| **DB** | Tambah kolom `image_url`; `image_data` (base64) dipertahankan sementara untuk backward-compat, di-drop di akhir |

**Kenapa R2** (mengacu jawaban brainstorming): fully managed, sudah pakai Cloudflare,
**egress gratis** (cocok POS read-heavy + aman walau volume belum diketahui), free tier 10GB
menutup fase awal ~$0, S3-compatible (pakai `aws-sdk-go-v2`, tidak lock-in).

## 2. Scope

- **In scope:** product image — `products.image_data` (TEXT, base64 data URI). Satu-satunya
  aset yang di-upload saat ini (entry point: `ProductFormPage` → `ImageUpload`).
- **Out of scope (sekarang):** `companies.logo_url` sudah berpola URL tapi belum diwire ke UI —
  saat dibangun nanti, ikut pola R2 yang sama. Avatar user di-generate dari inisial (bukan upload).

## 3. Target architecture

```
[Browser]
  1. pilih file → resize+compress (canvas → WebP ~1000px, q0.8)
  2. POST /uploads/presign  → { upload_url, public_url, key }     (ke nexpos-api)
  3. PUT <upload_url> (blob)  ───────────────────────────────────► [Cloudflare R2 bucket]
  4. simpan product dengan image_url = public_url                  (ke nexpos-api)

[POS / list load gambar]
  <img src="https://cdn.nexpos.irvanmahendra.com/products/{company}/{uuid}.webp">
                              │
                              ▼
                 [Cloudflare CDN cache] ──(miss)──► [R2]
```

Object key (tenant-scoped, non-guessable, cache-friendly):
```
products/{company_id}/{uuid}.webp
```
`uuid` baru setiap upload ⇒ otomatis cache-busting saat gambar diganti.

## 4. One-time setup (manual, sebelum coding)

1. Buat **R2 bucket** (mis. `nexpos-assets`) di dashboard Cloudflare.
2. Buat **R2 API token** (Access Key ID + Secret) dengan akses ke bucket itu.
3. Hubungkan **custom domain** publik untuk bucket (mis. `cdn.nexpos.irvanmahendra.com`) → otomatis lewat CDN
   Cloudflare; set **Cache rules** (cache everything, TTL panjang — aman karena key immutable).
4. Catat: `R2_ACCOUNT_ID`, `R2_ACCESS_KEY_ID`, `R2_SECRET_ACCESS_KEY`, `R2_BUCKET`,
   `R2_PUBLIC_BASE_URL` (= `https://cdn.nexpos.irvanmahendra.com`), `R2_ENDPOINT`
   (`https://<account_id>.r2.cloudflarestorage.com`).
5. Tambahkan semua di atas ke env VPS (`~/nexpos/.env` + docker-compose env) dan, jika backfill
   dijalankan via CI, ke GitHub secrets. **Jangan commit secrets.**

## 5. Backend changes (nexpos-api)

1. **Config** — baca env R2 (`internal/config`).
2. **Storage client** — paket baru `internal/storage` membungkus `aws-sdk-go-v2/service/s3`
   diarahkan ke endpoint R2; sediakan `PresignPut(ctx, key, contentType, expiry)` dan
   `Delete(ctx, key)`.
3. **Endpoint presign** — `POST /uploads/presign`
   - body: `{ "content_type": "image/webp", "kind": "product" }`
   - validasi: harus terautentikasi; `content_type` ∈ {webp, jpeg, png}; `kind` whitelist.
   - generate key `products/{company_id}/{uuid}.{ext}` (company_id dari JWT → **isolasi tenant**).
   - response: `{ "upload_url": <presigned PUT, exp 5 mnt>, "public_url": <R2_PUBLIC_BASE_URL/key>, "key": <key> }`.
4. **Migration** `0000XX_add_image_url_to_products`:
   ```sql
   ALTER TABLE products ADD COLUMN image_url TEXT;
   ```
   (additive, nullable — backward-compatible; `image_data` belum disentuh).
5. **Model/DTO** — tambah `image_url` di `Product`, `ProductResponse`, `CreateProductRequest`,
   `UpdateProductRequest`. API mengembalikan `image_url` bila ada, fallback `image_data`.
6. **Orphan cleanup** — saat update (gambar berubah) atau delete product, panggil
   `storage.Delete(oldKey)`. Simpel & sinkron dulu; bisa dibuat async nanti.

## 6. Frontend changes (nexpos-web)

1. **`ImageUpload.tsx`** — ganti alur `readAsDataURL`:
   - **Compress**: gambar → `<canvas>`, cap sisi terpanjang **1000px**, `canvas.toBlob('image/webp', 0.8)` (fallback `image/jpeg` bila WebP gagal).
   - **Upload**: `POST /uploads/presign` → `PUT upload_url` (blob) → `onChange(public_url)`.
   - State loading + error; pertahankan guard input 5MB.
2. **`productsService` / `ProductFormPage`** — kirim `image_url` (bukan base64) saat create/update.
3. **Display** — `<img src>` jalan untuk URL maupun data URI, jadi sisi tampil **tidak perlu diubah**
   selama transisi (komponen yang sekarang baca `image_data` cukup baca `image_url ?? image_data`).

## 7. Data migration / backfill

Skrip one-off Go (`cmd/backfill-images`) — **idempoten**:

```
untuk setiap product WHERE image_data IS NOT NULL AND image_url IS NULL:
    decode base64 image_data
    (opsional) re-encode/resize ke webp
    key = products/{company_id}/{uuid}.webp
    storage.Put(key, bytes)
    UPDATE products SET image_url = <public_url> WHERE id = ...
    (jangan null-kan image_data dulu — biar bisa rollback)
```

Dijalankan manual di VPS setelah Phase 1+2 stabil. Bisa di-batch + resume.

## 8. Rollout sequence (backward-compatible, bertahap)

| Phase | Aksi | Risiko |
|---|---|---|
| **0** | Provision R2 + custom domain + creds (bagian §4) | nol (di luar app) |
| **1** | Backend: config + storage client + endpoint presign + migration `image_url` (additive) + dual-read di response | rendah — belum ada yang nulis image_url |
| **2** | Frontend: ImageUpload compress+upload→R2, simpan image_url; display baca `image_url ?? image_data` | rendah — produk lama tetap pakai image_data |
| **3** | Jalankan backfill base64→R2 di VPS | sedang — tervalidasi, idempoten, image_data masih ada |
| **4** | Setelah verifikasi penuh: stop nulis image_data; migration **drop** `image_data` | rendah bila §3 sukses |

## 9. Security & multi-tenant

- Presign: **PUT only**, key spesifik, content-type spesifik, expiry pendek (±5 mnt).
- Key di-prefix `company_id` dari JWT ⇒ tenant tak bisa menulis ke namespace tenant lain.
- **Batas ukuran**: presigned PUT biasa tak menjamin content-length; gunakan **presigned POST
  dengan `content-length-range`** atau validasi ukuran setelah upload (cek HEAD object) lalu tolak/hapus bila lewat batas.
- **Public-read via CDN + key uuid non-guessable** dipilih demi simpel + cache. Katalog produk
  bukan data sensitif; bila nanti perlu privat, ganti ke signed URL / akses lewat Worker.
- Validasi MIME sebenarnya (magic bytes), bukan sekadar `content_type` dari client.

## 10. Cost & monitoring

- Fase awal diperkirakan **~$0** (dalam free tier R2: 10GB + 1M write + 10M read/bln; egress gratis).
- Pantau: ukuran bucket, Class A/B ops di dashboard R2. Set alert bila mendekati batas free tier.

## 11. Rollback

Karena semua additive: bila frontend bermasalah, revert frontend ke base64 — `image_url` di-abaikan,
`image_data` masih sumber kebenaran. Kolom `image_data` baru di-drop di Phase 4 setelah yakin.

## 12. Resolved decisions (terkunci)

1. **CDN domain**: `cdn.nexpos.irvanmahendra.com`.
2. **Format**: **WebP** utama, **otomatis fallback ke JPEG** bila encode WebP gagal. Ditangani di sisi client — tidak perlu konfigurasi.
3. **Ukuran tersimpan**: cap sisi terpanjang **1000px**, quality **0.8**. Satu ukuran dulu (cukup untuk grid POS & detail di layar retina); varian thumbnail via Cloudflare Image Resizing nanti bila perlu.
4. **Backfill**: `image_data` dibiarkan sampai **Phase 4** (rollback-safe).
5. **Batas ukuran upload**: enforce via **presigned POST `content-length-range`**.
```

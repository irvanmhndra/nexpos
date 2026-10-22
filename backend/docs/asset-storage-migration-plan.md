# Asset Storage Migration Plan — base64-in-Postgres → Cloudflare R2

> Status: **Phase 1 (backend) done — refactored to server-proxied** — config R2,
> `internal/storage` (`Upload`/`Delete`), endpoint **`POST /uploads`** (magic-byte validated,
> immutable cache), migration `image_url`, dual-read. Build + vet ✓.
> **Phase 2 (frontend) done** — `uploadImage` (compress→WebP 1000px→`POST /uploads`), `ImageUpload`
> refactored, `ProductFormPage` kirim `image_url`, semua display site dual-read (`image_url ?? image_data`).
> Build ✓. **Phase 3 backfill command written** — `cmd/backfill-images` (idempotent, resumable,
> baked into the API image; run `docker exec nexpos_api /app/backfill-images [-dry-run|-limit N]`).
> **Phase 0** (provisioning R2 + `R2_*` env on VPS) + running the backfill + **Phase 4** (drop `image_data`) sisa.
>
> ⚠️ **REVISI 2026-07-14** — setelah R2 dieksekusi beneran di [[project_undangin]], ada 3 koreksi
> yang **override** detail di bawah. Baca **§0** dulu.

## 0. Revisi dari eksekusi nyata (Undangin) — override yang di bawah

Tiga pelajaran dari nge-ship R2 di Undangin yang mengoreksi plan ini:

1. **CDN domain wajib 1-level: `nexpos-cdn.irvanmahendra.com`** (bukan `cdn.nexpos.…`). R2 custom
   domain selalu proxied; Universal SSL gratis Cloudflare cuma cover `*.irvanmahendra.com` (1 level),
   jadi hostname 2-level bikin **"not covered by a certificate"** / gagal SSL (kecuali bayar ACM).
   *(Semua `cdn.nexpos.…` di doc ini sudah diganti ke `nexpos-cdn.…`.)*

2. **Cache Rule itu WAJIB, bukan opsional.** Tanpa Cache Rule eksplisit, R2 custom domain balikin
   `cf-cache-status: DYNAMIC` (nggak ke-edge-cache). Buat POS read-heavy ini penting — bikin Cache Rule
   di `nexpos-cdn`: hostname match → **Eligible for cache**, Edge TTL = **honor cache-control** (objek
   immutable, aman TTL panjang).

3. **Upload flow: ganti presigned PUT → server-proxied.** Setelah dikerjain, proxied lebih pas untuk
   gambar kecil: (a) server bisa **validasi magic-byte** sebelum simpan (§9 minta ini; presigned nggak
   bisa karena byte langsung ke R2), (b) menghindari **footgun checksum** aws-sdk-go-v2 di presigned PUT
   ke R2 (bakal kena di Phase 2), (c) lebih simpel, tanpa CORS. Bandwidth negligible (upload produk occasional).
   **Konsekuensi:** Phase 1 (`/uploads/presign`) di-refactor ke **`POST /uploads`** — terima multipart →
   validasi magic-byte → `PutObject` (set `Cache-Control: immutable` + `RequestChecksumCalculation=WhenRequired`)
   → balikin `public_url`. Template siap: `undangin-api/internal/handler/upload.go` + `internal/storage/r2.go`.

Sisa plan (§1–§12) tetap valid — tinggal disesuaikan dengan 3 koreksi ini. (Bucket name saran ikut
konvensi `<app>-prod` → `nexpos-prod`, bukan `nexpos-assets`, biar seragam antar app.)

## 1. Decision summary

| | |
|---|---|
| **Target storage** | **Cloudflare R2** (S3-compatible object storage) |
| **Delivery** | Public bucket via custom domain **nexpos-cdn.irvanmahendra.com** (Cloudflare CDN cache, PoP Jakarta/SG) |
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
  <img src="https://nexpos-cdn.irvanmahendra.com/products/{company}/{uuid}.webp">
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
3. Hubungkan **custom domain** publik untuk bucket (mis. `nexpos-cdn.irvanmahendra.com`) → otomatis lewat CDN
   Cloudflare; set **Cache rules** (cache everything, TTL panjang — aman karena key immutable).
4. Catat: `R2_ACCOUNT_ID`, `R2_ACCESS_KEY_ID`, `R2_SECRET_ACCESS_KEY`, `R2_BUCKET`,
   `R2_PUBLIC_BASE_URL` (= `https://nexpos-cdn.irvanmahendra.com`), `R2_ENDPOINT`
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
- **Batas ukuran**: dipilih **presigned PUT** (didukung pasti oleh R2). Ukuran ditahan via
  kompresi sisi client (~1000px WebP → file kecil) + guard 5MB. Enforcement ketat di server
  (presigned POST `content-length-range` atau HEAD-check pasca-upload) **ditunda** karena
  dukungan presigned-POST di R2 belum terverifikasi.
- Client **wajib** mengirim header `Content-Type` yang sama persis dengan yang diminta saat
  presign (karena ikut ditandatangani), kalau tidak signature ditolak.
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

1. **CDN domain**: `nexpos-cdn.irvanmahendra.com`.
2. **Format**: **WebP** utama, **otomatis fallback ke JPEG** bila encode WebP gagal. Ditangani di sisi client — tidak perlu konfigurasi.
3. **Ukuran tersimpan**: cap sisi terpanjang **1000px**, quality **0.8**. Satu ukuran dulu (cukup untuk grid POS & detail di layar retina); varian thumbnail via Cloudflare Image Resizing nanti bila perlu.
4. **Backfill**: `image_data` dibiarkan sampai **Phase 4** (rollback-safe).
5. **Batas ukuran upload**: pakai **presigned PUT** (reliable di R2); ukuran ditahan via kompresi client + guard 5MB. Enforcement server-side ketat ditunda (lihat §9). *(Berubah dari rencana awal presigned POST karena dukungan R2 belum terverifikasi.)*
```

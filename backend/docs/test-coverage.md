# Nexpos — Test Coverage Ledger

Living record of what has been **exercised end-to-end** (not just typechecked/shipped),
so future test runs skip what's already covered and focus on the gaps.

**Before testing:** read this file first. Only re-test an item if its code changed
since the "last verified" date, or it's marked ❌/👁️.
**After testing:** update the rows (date + status) and move covered items up.

Covers `nexpos-api` (Go, Echo+Postgres) + `nexpos-web` (React SPA).

> A full E2E audit was done **2026-09-14** (API via curl + UI via browser) against a
> freshly-seeded local DB. Three bugs were found and fixed (see bottom).

## Legend
- ✅ **E2E** — action performed and result verified (curl and/or browser)
- 👁️ **render-only** — page loads with real data, action not exercised
- ❌ **not tested**
- 🚫 **blocked** — needs an external dependency

## Test environment
- API: `POSTGRES_PORT=… POSTGRES_PASSWORD=… JWT_SECRET=… go run ./cmd/api` (:8080).
  `config.Load` reads **env vars, not `.env`** — pass POSTGRES_PASSWORD/PORT explicitly
  when running by hand (the `.env` password is only used via `make` which exports it).
- **Local DB caveat:** a native postgres already owns host `:5432`, so the dev
  `docker compose` DB was run on host **:5434** (`docker run … -p 5434:5432 postgres:18-alpine`)
  and everything pointed at `POSTGRES_PORT=5434` to avoid the clash. Migrations:
  `migrate -path migrations -database postgres://pos_user:…@localhost:5434/pos_db?sslmode=disable up`.
- **Seeder:** `go run ./cmd/seed [-reset]` — demo shoe store (see below). Was built
  during this audit; didn't exist before.
- Web: `pnpm dev` (:5173), `.env.local` has `VITE_ENABLE_MSW=false` + API base :8080.
- **Most read endpoints require an explicit `?branch_id=` query param** (multi-branch);
  order create/POS take the branch from the session's default branch instead.
- Browser pane often hidden → drive React via `javascript_tool`; product cards are
  `<button>` — dispatch full mousedown/mouseup/click, not just a bare `.click()`.

**Seeder demo data:** company DEMO (tax 11%, `auto_complete_counter_orders=true`),
users owner@/admin@/kasir@demo.test (pw `password123`), branch MAIN + user_branches,
8 products / 11 variants (2 low-stock) + stock, 3 customers, 2 promotions
(DISKON20 %, POTONG50K fixed), 2 suppliers, 3 expense categories. **No orders/shifts**
are seeded — create those through the API/UI.

---

## API (curl) — ✅ verified 2026-09-14
| Flow | Status | Notes |
|---|---|---|
| Login | ✅ | returns token + company + default_branch |
| Products list / search (name+SKU, trim) | ✅ | "sepatu" & "sepatu " & "RUN" all filter correctly (old bug stays fixed) |
| Product lookup by barcode (`?code=`) | ✅ | |
| Inventory list + stats (`?branch_id=`) | ✅ | value/low-stock correct with branch param |
| Inventory adjust (IN/OUT/ADJUST) | ✅ | +50 applied, movement logged |
| Reports summary / sales-trend | ✅ | summary revenue now completed-only (bug #1 fixed) |
| Order create → confirm → add payment → complete | ✅ | tax 11% + DISKON20 auto-applied; stock deducted; movement OUT logged |
| Order void | ✅ | status voided + stock restored |
| Order cancel (draft) | ✅ | |
| Refund payment (partial) | ✅ | payment → partially_refunded, refunded_amount set |
| Shift open / current / close | ✅ | close computes expected/actual/difference |
| Purchase order create → receive | ✅ | received → stock increased |
| Expense create | ✅ | |
| CRUD create: customer / product+variant / promotion / supplier / category / user / branch | ✅ | (categories are at `/product-categories`) |

## Web UI (browser) — ✅ verified 2026-09-14
| Screen | Status | Notes |
|---|---|---|
| Login | ✅ | |
| Dashboard | ✅ | revenue Rp364k after bug #1 fix (was Rp1.09jt) |
| POS: add to cart → Bayar → Cash → Konfirmasi | ✅ | order auto-completes, stock deducts, receipt shown |
| Products list + search | ✅ | |
| Inventory (branch context, low-stock, value) | ✅ | |
| Reports page | ✅ | |
| Orders management page | 👁️ | list renders; per-order actions from UI not clicked |

## Not yet tested (future focus)
- POS: split payment, QRIS/Transfer methods, customer attach, promo manual apply, qr scanner.
- Order management UI: confirm/void/cancel/refund from the Orders screen (done via API only).
- Purchase order / shift / expense / supplier / customer / promotion / product **edit + delete** (creates done; updates/deletes not).
- Branch switcher UI (switching the active branch and re-scoping lists).
- Reports: number correctness for top-products / category-revenue / payment-methods / hourly-sales; date-range + branch filters.
- Stock opname, daily settlements (UI + API).
- Users/roles/permissions management UI; company settings UI.
- Auth: token refresh, logout.

## Accepted behaviors (NOT bugs)
- `auto_complete_counter_orders` gates whether a paid counter order auto-completes
  (and thus deducts stock / counts as revenue). Seeder sets it **true**.
- Stock deducts at order **completion**, not at payment/confirm, in the same
  transaction as the completion.
- Selling more than the recorded stock completes the sale and floors the stock at 0
  (logged as a warning); the next stock opname reconciles it.
- Read endpoints need `?branch_id=`; without it they scope to branch 0 → empty.

## Bugs found & fixed (2026-09-14 audit)
1. **Reports summary counted voided + cancelled orders as revenue** — `GetSummary`
   summed `grand_total`/discount/tax and `total_orders` over ALL statuses (every other
   report query filters `status='completed'`). Dashboard showed Rp1.092.000 instead of
   Rp364.000. Fixed: revenue/discount/tax/total_orders now `FILTER (WHERE status='completed')`.
2. **Seeder had `auto_complete_counter_orders=false`** → POS sales stayed as *paid drafts*:
   no stock deduction, not counted in reports. Fixed the seeder default to `true` (correct
   for a counter POS); POS sales now complete + deduct.
3. **`DateRangePicker` nested a `<button>` inside a `<button>`** (invalid HTML / React
   hydration warning) — the clear (×) control. Changed to `<span role="button">`.

## Enhancements (post-audit)
- **Promotion lifecycle status** — added a server-derived `status`
  (`inactive|scheduled|active|expired`) on the promotion read model, computed from
  `is_active` + `start_at`/`end_at` vs server time (no cron, no stored status). List
  supports `?status=` filtering (server-time via `NOW()`). UI shows a 4-state badge
  (Berjalan/Terjadwal/Kadaluarsa/Nonaktif) + status filter. Verified E2E 2026-09-14.
  Rationale: expiry was previously only enforced at order-apply time, so an expired
  promo still displayed "Aktif" in the list.

- **Orphaned R2 images cleaned up** (2026-09-16, backport of the fnb fix) — replacing or
  clearing a product image, or deleting the product, now best-effort deletes the old R2
  object via `storage.KeyFromURL` after a successful update/delete (never fails the
  operation). Company logo has no service update path, so it's not affected. Verified:
  `KeyFromURL` round-trip unit test; build + product tests green.

- **Order and stock flows are transactional** (2026-09-26) — order create/update/
  confirm/pay/complete/cancel/void/refund, inventory adjust, PO receive, and stock-opname
  complete now run in one DB transaction each, locking the order/payment/PO/opname row
  and every stock row they change (`FOR UPDATE`, stock locked in variant order).
  Stock deduction on completion is no longer best-effort: if it fails the completion
  rolls back. Before this, `tests/integration/concurrency_test.go` showed on `main`:
  8 parallel sales of one variant deducted 2 units instead of 8, 8 parallel +5
  adjustments added 15 instead of 40, and a create whose payment insert failed left
  the order behind. All five concurrency tests pass on the fix; the double-complete
  and double-refund cases did not reproduce on `main` in that run (narrower race) and
  stay as regression guards.

## Still open (feature, not a test gap)
- Payment gateway (Midtrans) — not built. Same direction as fnb (QRIS, per-tenant keys).

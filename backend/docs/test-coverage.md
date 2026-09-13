# Nexpos — Test Coverage Ledger

Living record of what has been **exercised end-to-end** (not just typechecked/shipped),
so future test runs skip what's already covered and focus on the gaps.

**Before testing:** read this file first. Only re-test an item if its code changed
since the "last verified" date, or it's marked ❌/👁️.
**After testing:** update the rows (date + status) and move covered items up.

Covers `nexpos-api` (Go, Echo+Postgres) + `nexpos-web` (React SPA, served under /mybiz).

> ⚠️ **No systematic E2E test pass has been done on nexpos yet.** The 2026-09 work
> was a production-incident fix + targeted bug fixes (below), verified by code
> reasoning and shipped — **not** click-tested by the assistant. Treat every feature
> table below as **needs testing** until proven otherwise.

## Legend
- ✅ **E2E** — action performed and result verified (curl and/or browser)
- 👁️ **render-only** — page loads with real data, action not exercised
- ❌ **not tested** (this cycle)
- 🚫 **blocked** — needs an external dependency

## Test environment
- API: `SERVER_PORT=8080`, Postgres `pos_user` / `pos_db` @ :5432.
- Web: `pnpm dev` (or npm) → :5173.
- ⚠️ **No `cmd/seed` exists** — there is no demo-data seeder (unlike fnb). Testing
  needs either data entered by hand or a copy of prod-shaped data. *Building a seeder
  is the first prerequisite for a proper E2E pass* (see gaps).
- Prod is **live** — don't run destructive tests against it.
- Browser pane often hidden → drive React via `javascript_tool` and verify via DOM /
  API fetch / the API request log.

---

## Feature surface (all ❌ = not E2E-tested this cycle)

### Auth & org
| Flow | Status | How |
|---|---|---|
| Login / logout / token refresh / register | ❌ | |
| Users CRUD + status | ❌ | |
| Branches CRUD | ❌ | (branch edit modal fix shipped — see below) |
| Company settings get/update | ❌ | (user-profile save fix shipped — see below) |

### Catalog
| Flow | Status | How |
|---|---|---|
| Product categories CRUD | ❌ | |
| Products CRUD + variants + image upload | ❌ | |
| Product search (name / SKU / barcode) | ❌ | fix shipped (see below), not re-tested E2E |
| Product lookup (barcode scan) | ❌ | |

### POS / Orders
| Flow | Status | How |
|---|---|---|
| POS: build cart → preview → create order | ❌ | |
| Order confirm / add payment / complete | ❌ | |
| Order cancel / void / refund payment | ❌ | |
| Receipt generation | ❌ | |
| Customers CRUD (+ attach to order) | ❌ | |
| Promotions CRUD + applied at checkout | ❌ | promo edit-modal fix shipped (see below) |

### Inventory & procurement
| Flow | Status | How |
|---|---|---|
| Inventory list + stats + adjust stock + min-stock | ❌ | |
| Stock movements list + stats | ❌ | |
| Suppliers CRUD | ❌ | |
| Purchase orders create → receive | ❌ | |

### Ops & finance
| Flow | Status | How |
|---|---|---|
| Shifts open / current / close | ❌ | |
| Expense categories CRUD | ❌ | |
| Expenses CRUD + summary | ❌ | |
| Daily settlements | ❌ | |
| Reports: summary / sales-trend / top-products / category-revenue / payment-methods / hourly | ❌ | |
| Dashboard | ❌ | row-click→/orders fix shipped (see below) |

---

## Shipped fixes (2026-09, code-verified + deployed; NOT formally E2E-tested)
Re-verify these first if a full pass is done — they're the most recently touched.
- **Prod incident**: login CORS + 502 — root cause Postgres down after reboot (missing
  `restart: unless-stopped`). Fixed via restart policy + ship-compose-from-repo. (Infra.)
- **Product search**: "sepatu" not filtering — uncorrelated `EXISTS` subquery returned
  global TRUE + no `TrimSpace`. Fixed to correlated `pv.product_id = products.id` across
  name/SKU/barcode + trim the query param.
- **Promo edit modal opened blank** (fresh insert instead of the edited row) — stale
  `useState(initialData)` on an always-mounted modal. Fixed with `key={editing?.id}`.
  Same fix applied to the **branch** edit modal.
- **Sellability facades removed**: deleted the mock Integrations tab and the fake
  Company-Profile save; wired the user-profile save to the real `useUpdateUser` API.
- **Dashboard** recent-orders row click → navigates to `/orders`.
- **Fonts**: FOUC/FOUT fix (fonts in `<head>` + `display=block` for Material Symbols).
- Infra sweep: GHCR non-expiring auth, govulncheck gate, postgres:18 parent-dir volume mount.

## Prerequisites before a real E2E pass
1. **Build a `cmd/seed`** (mirror fnb's) — company, users, branches, categories,
   products+variants, customers, a few orders across states, promotions, inventory,
   suppliers/POs, an open shift. Without it there's no data to test against locally.
2. Then run the pass area-by-area (POS checkout is the highest-value / highest-risk).

## Still open (feature, not a test gap)
- Payment gateway (discussed, Midtrans) — not built. Same direction as fnb
  (QRIS, per-tenant keys).

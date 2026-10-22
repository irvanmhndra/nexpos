# nexpos

Multi-tenant POS SaaS. One repo, two sides:

| Folder | What | Deploys to |
|---|---|---|
| `backend/` | **Rust API** (axum, sqlx): the rewrite. It takes over route group by route group behind nginx; see `backend-go/docs/rust-rewrite-plan.md` | not deployed yet |
| `backend-go/` | Go API (Echo v5, sqlx, Postgres), formerly `nexpos-api`. It serves production until the cutover and owns the migrations | `ghcr.io/irvanmhndra/nexpos-api` → `~/nexpos` on the VPS → `https://api.nexpos.irvanmahendra.com` |
| `frontend/` | React web app (Vite, TanStack Query), formerly `nexpos-web` | static `dist/` → `/var/www/nexpos-web/` on the VPS |

The Rust API must answer exactly like the Go one. The Go integration suite checks this: `just conformance <url> '<tests>'` runs it against a running server.

History of both old repos is preserved (`git log -- backend-go/…` and `git blame` work as before). The Go module path is still `github.com/irvanmhndra/nexpos-api`.

## Run locally

```bash
just dev-db     # Postgres (needs backend-go/.env, see backend-go/.env.example)
just dev-go     # Go API on :8080 (applies migrations)
just dev-api    # Rust API on :8090
just dev-web    # web on :5173
```

`just --list` shows the rest (tests, lint, migrations). See `backend-go/README.md` and `frontend/README.md` for details.

## CI/CD

- `.github/workflows/deploy.yml` runs on push to `main` and builds only the parts that changed:
  - **Go API** (`backend-go/`): lint, test, govulncheck and gosec, then image, compose, migrate and swap.
  - **Rust API** (`backend/`): fmt, clippy and tests; conformance (the Go suite against the Rust server); Docker image build. Not deployed yet.
  - **Frontend:** build and deploy. It waits for the Go API and is skipped if any Go job failed.
- `.github/workflows/pr-check.yml` runs the same checks on pull requests.
- Secrets: `VPS_HOST`, `VPS_USER`, `VPS_SSH_KEY`.

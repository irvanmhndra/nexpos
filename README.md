# nexpos

Multi-tenant POS SaaS. One repo, two sides:

| Folder | What | Deploys to |
|---|---|---|
| `backend/` | Go API (Echo v5, sqlx, Postgres) — formerly `nexpos-api` | `ghcr.io/irvanmhndra/nexpos-api` → `~/nexpos` on the VPS → `https://api.nexpos.irvanmahendra.com` |
| `frontend/` | React web app (Vite, TanStack Query) — formerly `nexpos-web` | static `dist/` → `/var/www/nexpos-web/` on the VPS |

History of both old repos is preserved (`git log -- backend/…`, `git blame` work as before). The Go module path is still `github.com/irvanmhndra/nexpos-api`.

## Run locally

```bash
just dev-db     # Postgres (needs backend/.env, see backend/.env.example)
just dev-api    # API on :8080
just dev-web    # web on :5173
```

`just --list` shows the rest (tests, lint, migrations). See `backend/README.md` and `frontend/README.md` for details.

## CI/CD

- `.github/workflows/deploy.yml` runs on push to `main`. Only the side that changed is built: backend lint, test, govulncheck and gosec, then image, compose, migrate and swap. The frontend deploy waits for the backend and is skipped if any backend job failed.
- `.github/workflows/pr-check.yml` runs the same checks on pull requests, per changed side.
- Secrets: `VPS_HOST`, `VPS_USER`, `VPS_SSH_KEY`.

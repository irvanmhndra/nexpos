# Nexpos Web

Frontend for Nexpos — a multi-tenant Point of Sale (POS) web application.

## Tech Stack

- **React 18** + TypeScript + Vite
- **Tailwind CSS** — Styling
- **TanStack Query (React Query v5)** — Server state management
- **React Router v6** — Routing
- **Axios** — HTTP client
- **MSW (Mock Service Worker)** — API mocking for development

## Features

- **POS Terminal** — Cart management, promo code application, server-side order preview, payment processing
- **Products** — Product catalog with multi-variant support
- **Inventory** — Stock levels, adjustments (in/out/adjustment), movement history
- **Orders** — Order history and details
- **Promotions** — Promo code management (fixed/percentage discounts)
- **Customers** — Customer profiles
- **Branches** — Branch management
- **Categories** — Product category management
- **Reports** — Sales summary, trends, top products, category/payment breakdowns
- **Settings** — Company settings, user management

## Getting Started

### Prerequisites

- Node.js 18+
- npm

### Install dependencies

```bash
npm install
```

### Configure environment

Create a `.env.local` file (this overrides `.env` locally and is git-ignored):

```env
VITE_API_BASE_URL=http://localhost:8080/api/v1
VITE_APP_NAME=Nexpos
VITE_APP_ENV=development
VITE_ENABLE_MSW=true
```

Set `VITE_ENABLE_MSW=true` to use mock data without a backend.

### Start development server

```bash
npm run dev
```

App runs at `http://localhost:5173`.

## Available Scripts

```bash
npm run dev        # Start dev server with HMR
npm run build      # Production build
npm run preview    # Preview production build locally
npm run lint       # Run ESLint
```

## Project Structure

```
src/
├── components/
│   ├── layout/          # Sidebar, Header, Layout wrapper
│   └── ui/              # Reusable UI components (Button, Modal, Badge, etc.)
├── hooks/               # TanStack Query hooks (useProducts, useOrders, useInventory, ...)
├── mocks/
│   ├── data.ts          # Mock data (products, orders, customers, stocks, movements, ...)
│   └── handlers.ts      # MSW request handlers
├── pages/
│   ├── auth/            # Login page
│   ├── dashboard/       # Dashboard
│   ├── pos/             # POS terminal
│   ├── products/        # Product catalog
│   ├── inventory/       # Stock overview + movement history
│   ├── orders/          # Order list + details
│   ├── promotions/      # Promo code management
│   ├── customers/       # Customer management
│   ├── branches/        # Branch management
│   ├── categories/      # Product category management
│   ├── reports/         # Sales reports
│   └── settings/        # Company settings, user management
├── services/            # Axios API call functions
├── types/               # TypeScript interfaces and types
└── utils/
    └── constants.ts     # Routes, status labels, badge color maps
```

## Environment Files

| File | Purpose |
|------|---------|
| `.env` | Base defaults (committed to git) |
| `.env.local` | Local overrides (git-ignored) — use for development |
| `.env.production` | Production build overrides |

**Note:** Vite applies env files in order of precedence: `.env` → `.env.local` (dev) / `.env.production` (build).

## API Mocking (MSW)

With `VITE_ENABLE_MSW=true`, MSW intercepts all HTTP requests and returns mock data — no backend needed for frontend development.

Mock data covers: products, variants, stocks, stock movements, orders, customers, promotions, branches, categories, users, and reports.

To use the real backend, set `VITE_ENABLE_MSW=false` and ensure `VITE_API_BASE_URL` points to the running API.

## Design System

- **Primary color:** `#137fec` (Blue)
- **Font family:** Manrope
- **Icons:** Material Symbols Outlined

## Backend

See [nexpos-api README](../pos-core-api/README.md) for backend setup and API documentation.

## License

Private — All Rights Reserved

# Plan Konversi HTML/Tailwind ke React - Nexpos Dashboard

## Executive Summary

Dokumen ini berisi rencana lengkap untuk mengkonversi 12 mockup HTML/Tailwind aplikasi Nexpos Store Dashboard menjadi aplikasi React modern dengan arsitektur yang scalable, maintainable, dan mengikuti best practices.

---

## 1. Analisis Mockup Existing

### 1.1 Halaman yang Teridentifikasi
Berdasarkan analisis direktori mockups, terdapat 12 halaman:

1. **Dashboard Utama** - Halaman overview dengan statistik dan grafik
2. **Halaman Login/Pendaftaran** - Authentication pages
3. **Manajemen Produk/Layanan** - Product listing dengan variant management
4. **Manajemen Pesanan/Penjualan** - Order management
5. **Manajemen Pelanggan** - Customer database
6. **Manajemen Cabang** - Branch management
7. **Laporan & Analitik** - Reports and analytics
8. **Halaman Pengaturan** - Settings hub
9. **Pengaturan: Umum** - General settings
10. **Pengaturan: Pengguna & Hak Akses** - User management
11. **Pengaturan: Notifikasi** - Notification preferences
12. **Pengaturan: Integrasi** - Third-party integrations

### 1.2 Komponen UI yang Teridentifikasi

#### Layout Components
- **Sidebar Navigation** (2 variants: dengan logo besar, dan compact)
- **Top Header/Navbar** dengan search, notifications, user profile
- **Breadcrumb Navigation**
- **Settings Sidebar** untuk halaman pengaturan

#### Data Display Components
- **Stats Card** - Kartu statistik dengan icon, value, dan trend
- **Data Table** dengan sorting, pagination, dan row actions
- **Expandable Table Row** untuk product variants
- **Chart Component** (SVG line chart)
- **Badge/Status Indicator** dengan berbagai warna

#### Form Components
- **Text Input** dengan berbagai styles
- **Select/Dropdown** dengan icon
- **Textarea**
- **Checkbox**
- **Search Input** dengan icon
- **File Upload** (untuk logo perusahaan)
- **Button** (primary, secondary, icon-only)

#### Utility Components
- **Avatar** (image dan initial-based)
- **Icon** (Material Symbols)
- **Pagination**
- **Filter Button** dengan dropdown
- **Empty State**
- **Loading State**

### 1.3 Theme & Design System

**Color Palette:**
- Primary: `#137fec` (Blue)
- Background Light: `#f6f7f8`
- Background Dark: `#101922`
- Card Light: `#ffffff`
- Card Dark: `#1a2634` / `#1A202C` / `#1e2732` (variations)
- Text Main: `#111418`
- Text Sub: `#617589`
- Border Light: `#dbe0e6` / `#e5e7eb` / `#f0f2f4`
- Border Dark: `#2d3b4a` / `#2a3441` / `#404b5a`

**Typography:**
- Font Family: Manrope (sans-serif)
- Font Weights: 400, 500, 600, 700, 800

**Icons:**
- Material Symbols Outlined (Google Fonts)

**Border Radius:**
- Default: 0.25rem
- lg: 0.5rem
- xl: 0.75rem
- full: 9999px

---

## 2. Arsitektur React yang Direkomendasikan

### 2.1 Technology Stack

#### Core
- **React 18+** - Latest stable version
- **TypeScript** - Type safety dan better DX
- **Vite** - Fast build tool dan dev server

#### Routing
- **React Router v6** - Client-side routing
- **Protected Routes** untuk authentication

#### State Management
- **Zustand** - Lightweight global state (recommended)
  - Stores: auth, user, theme, sidebar
- **React Query (TanStack Query)** - Server state management
  - Data fetching, caching, synchronization
- **React Hook Form** - Form state management
- **Zod** - Schema validation

#### Styling
- **Tailwind CSS v3+** - Utility-first CSS
- **clsx / cn utility** - Conditional classnames
- **tailwind-merge** - Merge Tailwind classes properly

#### UI Components (Optional, tapi recommended)
- **Radix UI** - Unstyled, accessible components
  - Dialog/Modal
  - Dropdown Menu
  - Select
  - Tabs
  - Tooltip
- **Headless UI** (alternative) - by Tailwind team

#### Charts & Visualization
- **Recharts** atau **Chart.js** - Untuk grafik penjualan

#### Utilities
- **date-fns** atau **dayjs** - Date manipulation
- **axios** - HTTP client
- **react-hot-toast** - Notifications
- **lucide-react** - Icon library (optional, karena sudah pakai Material Symbols)

#### Development Tools
- **ESLint** - Code linting
- **Prettier** - Code formatting
- **Husky** - Git hooks
- **lint-staged** - Pre-commit linting

### 2.2 Folder Structure

```
mybiz/
├── public/
│   ├── favicon.ico
│   └── logo.png
├── src/
│   ├── assets/              # Static assets
│   │   ├── images/
│   │   └── icons/
│   │
│   ├── components/          # Reusable UI components
│   │   ├── ui/              # Base UI components
│   │   │   ├── Button.tsx
│   │   │   ├── Input.tsx
│   │   │   ├── Select.tsx
│   │   │   ├── Badge.tsx
│   │   │   ├── Card.tsx
│   │   │   ├── Avatar.tsx
│   │   │   ├── Icon.tsx
│   │   │   └── index.ts
│   │   │
│   │   ├── layout/          # Layout components
│   │   │   ├── MainLayout.tsx
│   │   │   ├── AuthLayout.tsx
│   │   │   ├── Sidebar.tsx
│   │   │   ├── Header.tsx
│   │   │   ├── SettingsSidebar.tsx
│   │   │   └── index.ts
│   │   │
│   │   ├── common/          # Common shared components
│   │   │   ├── DataTable/
│   │   │   │   ├── DataTable.tsx
│   │   │   │   ├── DataTablePagination.tsx
│   │   │   │   └── index.ts
│   │   │   ├── StatsCard.tsx
│   │   │   ├── SearchBar.tsx
│   │   │   ├── FilterButton.tsx
│   │   │   ├── BranchSelector.tsx
│   │   │   ├── DateRangePicker.tsx
│   │   │   ├── EmptyState.tsx
│   │   │   └── LoadingSpinner.tsx
│   │   │
│   │   └── features/        # Feature-specific components
│   │       ├── auth/
│   │       │   ├── LoginForm.tsx
│   │       │   ├── RegisterForm.tsx
│   │       │   └── SocialLoginButton.tsx
│   │       ├── dashboard/
│   │       │   ├── SalesChart.tsx
│   │       │   ├── RecentOrders.tsx
│   │       │   └── QuickStats.tsx
│   │       ├── products/
│   │       │   ├── ProductTable.tsx
│   │       │   ├── ProductRow.tsx
│   │       │   ├── ProductVariants.tsx
│   │       │   ├── ProductForm.tsx
│   │       │   └── ProductFilters.tsx
│   │       ├── orders/
│   │       │   ├── OrderTable.tsx
│   │       │   ├── OrderDetails.tsx
│   │       │   └── OrderStatusBadge.tsx
│   │       ├── customers/
│   │       │   ├── CustomerTable.tsx
│   │       │   ├── CustomerCard.tsx
│   │       │   └── CustomerForm.tsx
│   │       └── settings/
│   │           ├── CompanyProfileForm.tsx
│   │           ├── UserManagement.tsx
│   │           ├── NotificationSettings.tsx
│   │           └── IntegrationsList.tsx
│   │
│   ├── pages/               # Page components (routes)
│   │   ├── auth/
│   │   │   ├── LoginPage.tsx
│   │   │   └── RegisterPage.tsx
│   │   ├── dashboard/
│   │   │   └── DashboardPage.tsx
│   │   ├── products/
│   │   │   ├── ProductsPage.tsx
│   │   │   └── ProductDetailPage.tsx
│   │   ├── orders/
│   │   │   ├── OrdersPage.tsx
│   │   │   └── OrderDetailPage.tsx
│   │   ├── customers/
│   │   │   ├── CustomersPage.tsx
│   │   │   └── CustomerDetailPage.tsx
│   │   ├── branches/
│   │   │   └── BranchesPage.tsx
│   │   ├── reports/
│   │   │   └── ReportsPage.tsx
│   │   ├── settings/
│   │   │   ├── SettingsLayout.tsx
│   │   │   ├── GeneralSettingsPage.tsx
│   │   │   ├── UsersSettingsPage.tsx
│   │   │   ├── NotificationsSettingsPage.tsx
│   │   │   └── IntegrationsSettingsPage.tsx
│   │   └── NotFoundPage.tsx
│   │
│   ├── hooks/               # Custom React hooks
│   │   ├── useAuth.ts
│   │   ├── useTheme.ts
│   │   ├── usePagination.ts
│   │   ├── useDebounce.ts
│   │   ├── useMediaQuery.ts
│   │   └── useLocalStorage.ts
│   │
│   ├── store/               # Zustand stores
│   │   ├── authStore.ts
│   │   ├── themeStore.ts
│   │   ├── sidebarStore.ts
│   │   └── userStore.ts
│   │
│   ├── services/            # API services
│   │   ├── api.ts           # Axios instance
│   │   ├── authService.ts
│   │   ├── productService.ts
│   │   ├── orderService.ts
│   │   ├── customerService.ts
│   │   └── settingsService.ts
│   │
│   ├── types/               # TypeScript types
│   │   ├── auth.types.ts
│   │   ├── product.types.ts
│   │   ├── order.types.ts
│   │   ├── customer.types.ts
│   │   ├── api.types.ts
│   │   └── index.ts
│   │
│   ├── utils/               # Utility functions
│   │   ├── cn.ts            # classNames utility
│   │   ├── formatters.ts    # Format currency, date, etc.
│   │   ├── validators.ts    # Validation helpers
│   │   └── constants.ts     # App constants
│   │
│   ├── config/              # Configuration files
│   │   ├── theme.ts         # Theme configuration
│   │   └── routes.ts        # Route paths
│   │
│   ├── App.tsx              # Root component
│   ├── main.tsx             # Entry point
│   ├── router.tsx           # Route configuration
│   └── index.css            # Global styles & Tailwind imports
│
├── .env.example
├── .env.local
├── .eslintrc.json
├── .prettierrc
├── tailwind.config.js
├── tsconfig.json
├── vite.config.ts
├── package.json
└── README.md
```

### 2.3 Key Architectural Decisions

#### Component Design Patterns

1. **Composition over Inheritance**
   - Gunakan component composition
   - Props drilling minimal dengan context/store

2. **Atomic Design Methodology**
   - Atoms: Button, Input, Icon, Badge
   - Molecules: SearchBar, StatsCard, Avatar dengan label
   - Organisms: DataTable, Sidebar, Header
   - Templates: MainLayout, AuthLayout
   - Pages: DashboardPage, ProductsPage, dll

3. **Container/Presentational Pattern**
   - Smart components (containers) handle logic
   - Dumb components (presentational) handle UI
   - Contoh: `ProductsPage` (smart) → `ProductTable` (presentational)

4. **Custom Hooks for Logic Reuse**
   - `useAuth()` - Authentication logic
   - `useProducts()` - Product data fetching
   - `usePagination()` - Pagination logic
   - `useDebounce()` - Debounced search

#### State Management Strategy

1. **Client State** (Zustand)
   - UI state (sidebar open/closed, theme)
   - User session
   - App preferences

2. **Server State** (React Query)
   - API data fetching & caching
   - Mutations (create, update, delete)
   - Optimistic updates
   - Cache invalidation

3. **Form State** (React Hook Form)
   - Form values
   - Validation states
   - Error handling

#### Routing Strategy

```typescript
// Protected routes wrapper
<Route element={<ProtectedRoute />}>
  <Route element={<MainLayout />}>
    <Route path="/" element={<DashboardPage />} />
    <Route path="/products" element={<ProductsPage />} />
    <Route path="/orders" element={<OrdersPage />} />
    {/* etc */}
  </Route>
</Route>

// Public routes
<Route element={<AuthLayout />}>
  <Route path="/login" element={<LoginPage />} />
  <Route path="/register" element={<RegisterPage />} />
</Route>
```

---

## 3. Implementation Plan

### Phase 1: Project Setup & Foundation (Estimated: 2-3 days)

#### 1.1 Initialize Project
- [ ] Setup Vite + React + TypeScript project
- [ ] Install dependencies (React Router, Zustand, React Query, etc.)
- [ ] Configure Tailwind CSS with custom theme
- [ ] Setup ESLint + Prettier
- [ ] Configure path aliases (@components, @hooks, etc.)

#### 1.2 Theme Configuration
- [ ] Create `tailwind.config.js` with custom colors
- [ ] Setup dark mode configuration
- [ ] Import Material Symbols font
- [ ] Create `cn()` utility for className merging

#### 1.3 Base Utilities & Types
- [ ] Setup Axios instance with interceptors
- [ ] Create base TypeScript types
- [ ] Setup utility functions (formatters, validators)
- [ ] Create constants file

### Phase 2: Core Components (Estimated: 3-4 days)

#### 2.1 UI Components (Atoms)
- [ ] Button component dengan variants
- [ ] Input component (text, email, password, search)
- [ ] Select/Dropdown component
- [ ] Textarea component
- [ ] Checkbox component
- [ ] Badge/Status component
- [ ] Avatar component
- [ ] Icon wrapper component
- [ ] Card component

#### 2.2 Layout Components
- [ ] MainLayout dengan sidebar & header
- [ ] AuthLayout untuk login/register
- [ ] Sidebar component dengan navigation
- [ ] Header/Navbar component
- [ ] SettingsSidebar component
- [ ] Breadcrumb component

#### 2.3 Common Components
- [ ] StatsCard component
- [ ] DataTable component (reusable)
- [ ] Pagination component
- [ ] SearchBar component
- [ ] FilterButton dengan dropdown
- [ ] BranchSelector dropdown
- [ ] DateRangePicker component
- [ ] EmptyState component
- [ ] LoadingSpinner/Skeleton

### Phase 3: Authentication & State Management (Estimated: 2-3 days)

#### 3.1 Auth Implementation
- [ ] Setup Zustand auth store
- [ ] Create auth service (login, register, logout)
- [ ] Implement useAuth hook
- [ ] Create ProtectedRoute wrapper
- [ ] Setup token management (localStorage)

#### 3.2 React Query Setup
- [ ] Configure QueryClient
- [ ] Setup query hooks (useProducts, useOrders, etc.)
- [ ] Implement mutation hooks
- [ ] Setup cache invalidation strategies

#### 3.3 Theme & UI State
- [ ] Dark mode store & toggle
- [ ] Sidebar state management
- [ ] User preferences store

### Phase 4: Authentication Pages (Estimated: 1-2 days)

#### 4.1 Login Page
- [ ] Convert login HTML to React
- [ ] Implement LoginForm component
- [ ] Add form validation dengan Zod
- [ ] Social login button component
- [ ] Handle authentication flow

#### 4.2 Register Page
- [ ] Convert register HTML to React
- [ ] Implement RegisterForm component
- [ ] Add validation & error handling

### Phase 5: Dashboard Page (Estimated: 2-3 days)

#### 5.1 Dashboard Layout
- [ ] Convert dashboard HTML structure
- [ ] Implement responsive layout

#### 5.2 Dashboard Components
- [ ] QuickStats component (4 stat cards)
- [ ] SalesChart component (dengan Recharts)
- [ ] RecentOrders table
- [ ] Branch selector integration
- [ ] Date filter integration

#### 5.3 Data Integration
- [ ] Connect to API endpoints (mock jika belum ada)
- [ ] Implement data fetching dengan React Query
- [ ] Add loading states

### Phase 6: Product Management (Estimated: 3-4 days)

#### 6.1 Products Page
- [ ] Convert products HTML to React
- [ ] ProductTable component
- [ ] ProductRow component dengan expand/collapse
- [ ] ProductVariants nested table

#### 6.2 Product Features
- [ ] Search & filter functionality
- [ ] Branch filter
- [ ] Stock indicators dengan progress bar
- [ ] Status badges
- [ ] Pagination

#### 6.3 Product Forms
- [ ] Add Product modal/page
- [ ] Edit Product functionality
- [ ] Add Variant form
- [ ] Form validation
- [ ] Image upload component

#### 6.4 Data Management
- [ ] Product CRUD operations
- [ ] Optimistic updates
- [ ] Error handling

### Phase 7: Order Management (Estimated: 2-3 days)

#### 7.1 Orders Page
- [ ] Convert orders HTML to React
- [ ] OrderTable component
- [ ] Order status badges
- [ ] Order filters

#### 7.2 Order Details
- [ ] Order detail view/modal
- [ ] Order status update
- [ ] Order actions

#### 7.3 Integration
- [ ] Connect to orders API
- [ ] Real-time status updates (optional)

### Phase 8: Customer Management (Estimated: 2 days)

#### 8.1 Customers Page
- [ ] Convert customers HTML to React
- [ ] CustomerTable component
- [ ] Customer stats cards
- [ ] Search & filter

#### 8.2 Customer Features
- [ ] Customer detail view
- [ ] Add/Edit customer form
- [ ] Customer status management

### Phase 9: Branch Management (Estimated: 1-2 days)

#### 9.1 Branches Page
- [ ] Convert branch management HTML
- [ ] Branch list/grid view
- [ ] Add/Edit branch functionality

### Phase 10: Reports & Analytics (Estimated: 2-3 days)

#### 10.1 Reports Page
- [ ] Convert reports HTML to React
- [ ] Chart components
- [ ] Report filters
- [ ] Export functionality

### Phase 11: Settings Pages (Estimated: 2-3 days)

#### 11.1 Settings Layout
- [ ] Settings sidebar navigation
- [ ] Settings layout wrapper

#### 11.2 Settings Pages
- [ ] General Settings (company profile)
- [ ] User Management
- [ ] Notification Settings
- [ ] Integrations Settings

#### 11.3 Settings Features
- [ ] Logo upload functionality
- [ ] Form handling untuk semua settings
- [ ] Save/Cancel functionality

### Phase 12: Polish & Optimization (Estimated: 2-3 days)

#### 12.1 Performance Optimization
- [ ] Code splitting dengan React.lazy
- [ ] Image optimization
- [ ] Bundle size analysis
- [ ] React Query cache optimization

#### 12.2 Accessibility
- [ ] Keyboard navigation
- [ ] ARIA labels
- [ ] Focus management
- [ ] Screen reader support

#### 12.3 Responsive Design
- [ ] Test semua breakpoints
- [ ] Mobile navigation
- [ ] Touch interactions
- [ ] Responsive tables (horizontal scroll)

#### 12.4 Error Handling
- [ ] Error boundaries
- [ ] API error handling
- [ ] User-friendly error messages
- [ ] Toast notifications

#### 12.5 Testing (Optional)
- [ ] Unit tests untuk utilities
- [ ] Component tests dengan React Testing Library
- [ ] E2E tests dengan Playwright/Cypress

---

## 4. Component Examples

### 4.1 Example: StatsCard Component

```typescript
// components/common/StatsCard.tsx
interface StatsCardProps {
  title: string;
  value: string | number;
  icon: string;
  trend?: {
    value: string;
    isPositive: boolean;
    label: string;
  };
}

export function StatsCard({ title, value, icon, trend }: StatsCardProps) {
  return (
    <div className="flex flex-col gap-3 rounded-xl bg-card-light dark:bg-card-dark p-5 border border-border-light dark:border-border-dark shadow-sm">
      <div className="flex items-center justify-between">
        <p className="text-text-sub dark:text-gray-400 text-sm font-medium">
          {title}
        </p>
        <span className="material-symbols-outlined text-primary bg-primary/10 p-1 rounded-md text-[20px]">
          {icon}
        </span>
      </div>
      <div className="flex flex-col gap-1">
        <p className="text-text-main dark:text-white text-2xl font-bold">
          {value}
        </p>
        {trend && (
          <div className={cn(
            "flex items-center gap-1 text-sm font-medium",
            trend.isPositive ? "text-green-600" : "text-red-600"
          )}>
            <span className="material-symbols-outlined text-[16px]">
              {trend.isPositive ? "trending_up" : "trending_down"}
            </span>
            <span>{trend.value}</span>
            <span className="text-text-sub dark:text-gray-500 font-normal ml-1">
              {trend.label}
            </span>
          </div>
        )}
      </div>
    </div>
  );
}
```

### 4.2 Example: DataTable Component

```typescript
// components/common/DataTable/DataTable.tsx
interface Column<T> {
  header: string;
  accessor: keyof T | ((row: T) => React.ReactNode);
  className?: string;
}

interface DataTableProps<T> {
  data: T[];
  columns: Column<T>[];
  onRowClick?: (row: T) => void;
  loading?: boolean;
  emptyMessage?: string;
}

export function DataTable<T extends { id: string | number }>({
  data,
  columns,
  onRowClick,
  loading,
  emptyMessage = "Tidak ada data"
}: DataTableProps<T>) {
  if (loading) return <LoadingSpinner />;
  if (data.length === 0) return <EmptyState message={emptyMessage} />;

  return (
    <div className="overflow-x-auto">
      <table className="w-full text-left text-sm">
        <thead className="bg-background-light dark:bg-background-dark border-y border-border-light dark:border-border-dark">
          <tr>
            {columns.map((column, idx) => (
              <th key={idx} className="px-6 py-3 font-medium text-text-sub dark:text-gray-400">
                {column.header}
              </th>
            ))}
          </tr>
        </thead>
        <tbody className="divide-y divide-border-light dark:divide-border-dark">
          {data.map((row) => (
            <tr
              key={row.id}
              onClick={() => onRowClick?.(row)}
              className="hover:bg-background-light dark:hover:bg-background-dark/50 transition-colors cursor-pointer"
            >
              {columns.map((column, idx) => (
                <td key={idx} className={cn("px-6 py-4", column.className)}>
                  {typeof column.accessor === 'function'
                    ? column.accessor(row)
                    : row[column.accessor]}
                </td>
              ))}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
```

### 4.3 Example: useAuth Hook

```typescript
// hooks/useAuth.ts
import { useAuthStore } from '@/store/authStore';
import { authService } from '@/services/authService';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { useNavigate } from 'react-router-dom';

export function useAuth() {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const { user, setUser, clearUser } = useAuthStore();

  const loginMutation = useMutation({
    mutationFn: authService.login,
    onSuccess: (data) => {
      setUser(data.user);
      localStorage.setItem('token', data.token);
      navigate('/');
    },
  });

  const logoutMutation = useMutation({
    mutationFn: authService.logout,
    onSuccess: () => {
      clearUser();
      localStorage.removeItem('token');
      queryClient.clear();
      navigate('/login');
    },
  });

  return {
    user,
    isAuthenticated: !!user,
    login: loginMutation.mutate,
    logout: logoutMutation.mutate,
    isLoggingIn: loginMutation.isPending,
  };
}
```

---

## 5. Best Practices & Guidelines

### 5.1 Code Style

- **TypeScript strict mode** - Enable untuk type safety
- **Named exports** - Lebih baik untuk tree-shaking
- **Consistent naming** - camelCase untuk variables/functions, PascalCase untuk components
- **File naming** - PascalCase untuk components (`Button.tsx`), camelCase untuk utilities (`formatters.ts`)

### 5.2 Component Guidelines

1. **Keep components small** - Single responsibility principle
2. **Extract reusable logic** ke custom hooks
3. **Use TypeScript interfaces** untuk props
4. **Avoid prop drilling** - Use context atau Zustand untuk deep data
5. **Memoization when needed** - React.memo, useMemo, useCallback untuk optimization
6. **Error boundaries** - Wrap risky components

### 5.3 State Management Rules

1. **Server state** → React Query
2. **Form state** → React Hook Form
3. **UI state** → Zustand (global) atau useState (local)
4. **Derived state** → useMemo atau computed values

### 5.4 Performance Tips

1. **Code splitting** - Lazy load routes
2. **Virtual scrolling** - Untuk long lists (react-virtual)
3. **Debounce search** - useDebounce hook
4. **Optimize images** - Use proper formats, lazy loading
5. **React Query cache** - Set appropriate stale times

### 5.5 Accessibility

1. **Semantic HTML** - Use proper HTML tags
2. **Keyboard navigation** - Tab order, Enter/Space for actions
3. **ARIA labels** - For screen readers
4. **Focus management** - Visible focus indicators
5. **Color contrast** - WCAG AA compliance

---

## 6. API Integration Strategy

### 6.1 Mock Data Phase
Gunakan mock data terlebih dahulu:
- Buat mock data di `src/mocks/`
- Gunakan `json-server` atau MSW (Mock Service Worker)
- Mock endpoints untuk semua CRUD operations

### 6.2 Real API Integration
Ketika backend ready:
- Update Axios base URL
- Implement proper error handling
- Add request/response interceptors
- Handle authentication tokens
- Implement retry logic untuk failed requests

### 6.3 API Service Pattern

```typescript
// services/productService.ts
import { api } from './api';
import { Product, ProductCreateInput } from '@/types/product.types';

export const productService = {
  getAll: async (params?: { search?: string; branch?: string }) => {
    const { data } = await api.get<Product[]>('/products', { params });
    return data;
  },

  getById: async (id: string) => {
    const { data } = await api.get<Product>(`/products/${id}`);
    return data;
  },

  create: async (input: ProductCreateInput) => {
    const { data } = await api.post<Product>('/products', input);
    return data;
  },

  update: async (id: string, input: Partial<ProductCreateInput>) => {
    const { data } = await api.patch<Product>(`/products/${id}`, input);
    return data;
  },

  delete: async (id: string) => {
    await api.delete(`/products/${id}`);
  },
};
```

---

## 7. Testing Strategy (Optional)

### 7.1 Unit Tests
- Test utility functions
- Test custom hooks dengan `@testing-library/react-hooks`
- Test pure components

### 7.2 Integration Tests
- Test components dengan user interactions
- Use `@testing-library/react`
- Mock API calls

### 7.3 E2E Tests
- Critical user flows (login, create order, etc.)
- Use Playwright atau Cypress
- Run in CI/CD pipeline

---

## 8. Deployment Considerations

### 8.1 Build Optimization
- [ ] Production build dengan Vite
- [ ] Analyze bundle size
- [ ] Enable gzip compression
- [ ] CDN untuk static assets

### 8.2 Environment Variables
```
VITE_API_BASE_URL=https://api.nexpos.com
VITE_APP_NAME=Nexpos
VITE_GOOGLE_CLIENT_ID=your-google-client-id
```

### 8.3 Deployment Platforms
- Vercel (recommended untuk React apps)
- Netlify
- AWS Amplify
- Traditional hosting dengan Nginx

---

## 9. Documentation Requirements

- [ ] README.md dengan setup instructions
- [ ] Component documentation (Storybook optional)
- [ ] API integration guide
- [ ] Deployment guide
- [ ] Contributing guidelines

---

## 10. Maintenance & Scalability

### 10.1 Code Quality
- Regular dependency updates
- Security audits (`npm audit`)
- Performance monitoring
- Error tracking (Sentry)

### 10.2 Feature Additions
- Feature flags untuk gradual rollout
- A/B testing setup
- Analytics integration (GA4, Mixpanel)

### 10.3 Team Collaboration
- Git workflow (feature branches)
- PR review guidelines
- Code review checklist
- Commit message conventions

---

## 11. Success Metrics

- [ ] 100% feature parity dengan mockups
- [ ] < 3s initial page load
- [ ] Lighthouse score > 90
- [ ] Zero critical accessibility issues
- [ ] Mobile responsive (all breakpoints)
- [ ] Dark mode fully functional
- [ ] TypeScript strict mode dengan no errors
- [ ] < 500kb initial bundle size

---

## Kesimpulan

Plan ini dirancang untuk mengkonversi mockup HTML/Tailwind ke aplikasi React yang production-ready dengan arsitektur yang scalable dan maintainable. Estimasi total waktu development: **4-6 minggu** (tergantung kompleksitas backend integration dan testing requirements).

**Next Steps:**
1. Review dan approval plan ini
2. Setup project dan dependencies
3. Mulai dengan Phase 1: Foundation
4. Iterative development mengikuti phases
5. Regular review dan adjustment

**Questions to Clarify:**
- Apakah akan ada backend API atau perlu mock data dulu?
- Apakah perlu multi-language support (i18n)?
- Apakah perlu real-time features (WebSocket)?
- Testing requirements (unit, integration, e2e)?
- Deployment target platform?

import { createBrowserRouter } from 'react-router-dom';
import { MainLayout, AuthLayout, POSLayout } from './components/layout';
import { ProtectedRoute } from './components/ProtectedRoute';

export const router = createBrowserRouter([
  // Auth routes
  {
    element: <AuthLayout />,
    children: [
      {
        path: '/login',
        lazy: async () => {
          const { LoginPage } = await import('./pages/auth/LoginPage');
          return { Component: LoginPage };
        },
      },
      {
        path: '/register',
        lazy: async () => {
          const { RegisterPage } = await import('./pages/auth/RegisterPage');
          return { Component: RegisterPage };
        },
      },
    ],
  },

  // POS route — full-height layout (no content padding)
  {
    element: (
      <ProtectedRoute>
        <POSLayout />
      </ProtectedRoute>
    ),
    children: [
      {
        path: '/pos',
        lazy: async () => {
          const { POSPage } = await import('./pages/pos/POSPage');
          return { Component: POSPage };
        },
      },
    ],
  },

  // Main app routes
  {
    element: (
      <ProtectedRoute>
        <MainLayout />
      </ProtectedRoute>
    ),
    children: [
      {
        path: '/',
        lazy: async () => {
          const { DashboardPage } = await import('./pages/dashboard/DashboardPage');
          return { Component: DashboardPage };
        },
      },
      {
        path: '/products',
        lazy: async () => {
          const { ProductsPage } = await import('./pages/products/ProductsPage');
          return { Component: ProductsPage };
        },
      },
      {
        path: '/products/new',
        lazy: async () => {
          const { ProductFormPage } = await import('./pages/products/ProductFormPage');
          return { Component: ProductFormPage };
        },
      },
      {
        path: '/products/:id/edit',
        lazy: async () => {
          const { ProductFormPage } = await import('./pages/products/ProductFormPage');
          return { Component: ProductFormPage };
        },
      },
      {
        path: '/categories',
        lazy: async () => {
          const { ProductCategoriesPage } = await import('./pages/categories/ProductCategoriesPage');
          return { Component: ProductCategoriesPage };
        },
      },
      {
        path: '/inventory',
        lazy: async () => {
          const { StockOverviewPage } = await import('./pages/inventory/StockOverviewPage');
          return { Component: StockOverviewPage };
        },
      },
      {
        path: '/inventory/movements',
        lazy: async () => {
          const { StockMovementPage } = await import('./pages/inventory/StockMovementPage');
          return { Component: StockMovementPage };
        },
      },
      {
        path: '/stock-opnames',
        lazy: async () => {
          const { StockOpnamesPage } = await import('./pages/inventory/StockOpnamesPage');
          return { Component: StockOpnamesPage };
        },
      },
      {
        path: '/stock-opnames/new',
        lazy: async () => {
          const { StockOpnameNewPage } = await import('./pages/inventory/StockOpnameNewPage');
          return { Component: StockOpnameNewPage };
        },
      },
      {
        path: '/stock-opnames/:id',
        lazy: async () => {
          const { StockOpnameDetailPage } = await import('./pages/inventory/StockOpnameDetailPage');
          return { Component: StockOpnameDetailPage };
        },
      },
      {
        path: '/orders',
        lazy: async () => {
          const { OrdersPage } = await import('./pages/orders/OrdersPage');
          return { Component: OrdersPage };
        },
      },
      {
        path: '/customers',
        lazy: async () => {
          const { CustomersPage } = await import('./pages/customers/CustomersPage');
          return { Component: CustomersPage };
        },
      },
      {
        path: '/branches',
        lazy: async () => {
          const { BranchesPage } = await import('./pages/branches/BranchesPage');
          return { Component: BranchesPage };
        },
      },
      {
        path: '/promotions',
        lazy: async () => {
          const { PromotionsPage } = await import('./pages/promotions/PromotionsPage');
          return { Component: PromotionsPage };
        },
      },
      {
        path: '/reports',
        lazy: async () => {
          const { ReportsPage } = await import('./pages/reports/ReportsPage');
          return { Component: ReportsPage };
        },
      },
      {
        path: '/settings',
        lazy: async () => {
          const { SettingsPage } = await import('./pages/settings/SettingsPage');
          return { Component: SettingsPage };
        },
      },
      {
        path: '/suppliers',
        lazy: async () => {
          const { SuppliersPage } = await import('./pages/suppliers/SuppliersPage');
          return { Component: SuppliersPage };
        },
      },
      {
        path: '/purchase-orders',
        lazy: async () => {
          const { PurchaseOrdersPage } = await import('./pages/purchase-orders/PurchaseOrdersPage');
          return { Component: PurchaseOrdersPage };
        },
      },
      {
        path: '/purchase-orders/new',
        lazy: async () => {
          const { PurchaseOrderFormPage } = await import('./pages/purchase-orders/PurchaseOrderFormPage');
          return { Component: PurchaseOrderFormPage };
        },
      },
      {
        path: '/purchase-orders/:id',
        lazy: async () => {
          const { PurchaseOrderDetailPage } = await import('./pages/purchase-orders/PurchaseOrderDetailPage');
          return { Component: PurchaseOrderDetailPage };
        },
      },
      {
        path: '/shifts',
        lazy: async () => {
          const { ShiftsPage } = await import('./pages/shifts/ShiftsPage');
          return { Component: ShiftsPage };
        },
      },
      {
        path: '/expenses',
        lazy: async () => {
          const { ExpensesPage } = await import('./pages/expenses/ExpensesPage');
          return { Component: ExpensesPage };
        },
      },
      {
        path: '/expense-categories',
        lazy: async () => {
          const { ExpenseCategoriesPage } = await import('./pages/expenses/ExpenseCategoriesPage');
          return { Component: ExpenseCategoriesPage };
        },
      },
      {
        path: '/daily-settlements',
        lazy: async () => {
          const { DailySettlementsPage } = await import('./pages/daily-settlements/DailySettlementsPage');
          return { Component: DailySettlementsPage };
        },
      },
      {
        path: '/daily-settlements/new',
        lazy: async () => {
          const { DailySettlementNewPage } = await import('./pages/daily-settlements/DailySettlementNewPage');
          return { Component: DailySettlementNewPage };
        },
      },
      {
        path: '/daily-settlements/:id',
        lazy: async () => {
          const { DailySettlementDetailPage } = await import('./pages/daily-settlements/DailySettlementDetailPage');
          return { Component: DailySettlementDetailPage };
        },
      },
    ],
  },
]);

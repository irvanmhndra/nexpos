import { createBrowserRouter } from 'react-router-dom';
import { MainLayout, AuthLayout, POSLayout } from './components/layout';
import { DashboardPage } from './pages/dashboard/DashboardPage';
import { ProductsPage } from './pages/products/ProductsPage';
import { ProductFormPage } from './pages/products/ProductFormPage';
import { ProductCategoriesPage } from './pages/categories/ProductCategoriesPage';
import { OrdersPage } from './pages/orders/OrdersPage';
import { CustomersPage } from './pages/customers/CustomersPage';
import { BranchesPage } from './pages/branches/BranchesPage';
import { PromotionsPage } from './pages/promotions/PromotionsPage';
import { ReportsPage } from './pages/reports/ReportsPage';
import { SettingsPage } from './pages/settings/SettingsPage';
import { LoginPage } from './pages/auth/LoginPage';
import { RegisterPage } from './pages/auth/RegisterPage';
import { StockOverviewPage } from './pages/inventory/StockOverviewPage';
import { StockMovementPage } from './pages/inventory/StockMovementPage';
import { POSPage } from './pages/pos/POSPage';
import { ProtectedRoute } from './components/ProtectedRoute';
import { SuppliersPage } from './pages/suppliers/SuppliersPage';
import { PurchaseOrdersPage } from './pages/purchase-orders/PurchaseOrdersPage';
import { PurchaseOrderFormPage } from './pages/purchase-orders/PurchaseOrderFormPage';
import { PurchaseOrderDetailPage } from './pages/purchase-orders/PurchaseOrderDetailPage';
import { ShiftsPage } from './pages/shifts/ShiftsPage';
import { ExpensesPage } from './pages/expenses/ExpensesPage';
import { ExpenseCategoriesPage } from './pages/expenses/ExpenseCategoriesPage';

export const router = createBrowserRouter([
  // Auth routes
  {
    element: <AuthLayout />,
    children: [
      {
        path: '/login',
        element: <LoginPage />,
      },
      {
        path: '/register',
        element: <RegisterPage />,
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
        element: <POSPage />,
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
        element: <DashboardPage />,
      },
      {
        path: '/products',
        element: <ProductsPage />,
      },
      {
        path: '/products/new',
        element: <ProductFormPage />,
      },
      {
        path: '/products/:id/edit',
        element: <ProductFormPage />,
      },
      {
        path: '/categories',
        element: <ProductCategoriesPage />,
      },
      {
        path: '/inventory',
        element: <StockOverviewPage />,
      },
      {
        path: '/inventory/movements',
        element: <StockMovementPage />,
      },
      {
        path: '/orders',
        element: <OrdersPage />,
      },
      {
        path: '/customers',
        element: <CustomersPage />,
      },
      {
        path: '/branches',
        element: <BranchesPage />,
      },
      {
        path: '/promotions',
        element: <PromotionsPage />,
      },
      {
        path: '/reports',
        element: <ReportsPage />,
      },
      {
        path: '/settings',
        element: <SettingsPage />,
      },
      // Suppliers & Purchase Orders
      {
        path: '/suppliers',
        element: <SuppliersPage />,
      },
      {
        path: '/purchase-orders',
        element: <PurchaseOrdersPage />,
      },
      {
        path: '/purchase-orders/new',
        element: <PurchaseOrderFormPage />,
      },
      {
        path: '/purchase-orders/:id',
        element: <PurchaseOrderDetailPage />,
      },
      // Shifts
      {
        path: '/shifts',
        element: <ShiftsPage />,
      },
      // Expenses
      {
        path: '/expenses',
        element: <ExpensesPage />,
      },
      {
        path: '/expense-categories',
        element: <ExpenseCategoriesPage />,
      },
    ],
  },
]);

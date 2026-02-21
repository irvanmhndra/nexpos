import { createBrowserRouter } from 'react-router-dom';
import { MainLayout, AuthLayout } from './components/layout';
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
import { ProtectedRoute } from './components/ProtectedRoute';

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
    ],
  },
]);

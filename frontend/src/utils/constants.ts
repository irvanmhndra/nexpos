/**
 * Application-wide constants
 */

// API Configuration
export const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api/v1';
export const API_TIMEOUT = 30000; // 30 seconds

// Pagination
export const DEFAULT_PAGE_SIZE = 10;
export const PAGE_SIZE_OPTIONS = [10, 25, 50, 100];

// Date Formats
export const DATE_FORMAT = 'dd MMM yyyy';
export const DATETIME_FORMAT = 'dd MMM yyyy HH:mm';
export const TIME_FORMAT = 'HH:mm';

// Order Status
export const ORDER_STATUS = {
  PENDING: 'pending',
  PROCESSING: 'processing',
  COMPLETED: 'completed',
  CANCELLED: 'cancelled',
} as const;

export const ORDER_STATUS_LABELS: Record<string, string> = {
  pending: 'Menunggu',
  processing: 'Proses',
  completed: 'Selesai',
  cancelled: 'Dibatalkan',
};

export const ORDER_STATUS_COLORS: Record<string, string> = {
  pending: 'yellow',
  processing: 'blue',
  completed: 'green',
  cancelled: 'red',
};

// Product Status
export const PRODUCT_STATUS = {
  ACTIVE: 'active',
  INACTIVE: 'inactive',
  OUT_OF_STOCK: 'out_of_stock',
} as const;

export const PRODUCT_STATUS_LABELS: Record<string, string> = {
  active: 'Aktif',
  inactive: 'Non-aktif',
  out_of_stock: 'Habis',
};

// Customer Status
export const CUSTOMER_STATUS = {
  ACTIVE: 'active',
  INACTIVE: 'inactive',
  VIP: 'vip',
} as const;

export const CUSTOMER_STATUS_LABELS: Record<string, string> = {
  active: 'Aktif',
  inactive: 'Inaktif',
  vip: 'VIP',
};

// Stock Status Thresholds
export const STOCK_STATUS = {
  SAFE: 'safe',
  LIMITED: 'limited',
  LOW: 'low',
  OUT: 'out',
} as const;

export const STOCK_THRESHOLDS = {
  LOW: 10,
  LIMITED: 30,
};

// Local Storage Keys
export const STORAGE_KEYS = {
  AUTH_TOKEN: 'nexpos_auth_token',
  USER: 'nexpos_user',
  THEME: 'nexpos_theme',
  SIDEBAR_COLLAPSED: 'nexpos_sidebar_collapsed',
  SELECTED_BRANCH: 'nexpos_selected_branch',
} as const;

// Inventory Status
export const INVENTORY_STATUS = {
  NORMAL: 'normal',
  LOW: 'low',
  OUT_OF_STOCK: 'out_of_stock',
} as const;

export const INVENTORY_STATUS_LABELS: Record<string, string> = {
  normal: 'Normal',
  low: 'Menipis',
  out_of_stock: 'Habis',
};

export const INVENTORY_STATUS_COLORS: Record<string, string> = {
  normal: 'green',
  low: 'yellow',
  out_of_stock: 'red',
};

// Movement Type
export const MOVEMENT_TYPE = {
  IN: 'IN',
  OUT: 'OUT',
  ADJUST: 'ADJUST',
} as const;

export const MOVEMENT_TYPE_LABELS: Record<string, string> = {
  IN: 'Stok Masuk',
  OUT: 'Stok Keluar',
  ADJUST: 'Penyesuaian',
};

export const MOVEMENT_TYPE_COLORS: Record<string, string> = {
  IN: 'green',
  OUT: 'red',
  ADJUST: 'blue',
};

// Routes
export const ROUTES = {
  HOME: '/',
  LOGIN: '/login',
  REGISTER: '/register',
  DASHBOARD: '/dashboard',
  PRODUCTS: '/products',
  INVENTORY: '/inventory',
  INVENTORY_MOVEMENTS: '/inventory/movements',
  ORDERS: '/orders',
  CUSTOMERS: '/customers',
  BRANCHES: '/branches',
  REPORTS: '/reports',
  SETTINGS: '/settings',
  SETTINGS_GENERAL: '/settings/general',
  SETTINGS_USERS: '/settings/users',
  SETTINGS_NOTIFICATIONS: '/settings/notifications',
  SETTINGS_INTEGRATIONS: '/settings/integrations',
  POS: '/pos',
} as const;

// File Upload
export const MAX_FILE_SIZE = 2 * 1024 * 1024; // 2MB
export const ACCEPTED_IMAGE_TYPES = ['image/jpeg', 'image/jpg', 'image/png', 'image/webp'];

// Chart Colors
export const CHART_COLORS = {
  primary: '#137fec',
  success: '#10b981',
  warning: '#f59e0b',
  danger: '#ef4444',
  info: '#3b82f6',
};

// Breakpoints (matching Tailwind)
export const BREAKPOINTS = {
  sm: 640,
  md: 768,
  lg: 1024,
  xl: 1280,
  '2xl': 1536,
} as const;

import { faker } from '@faker-js/faker';

// Note: locale setting removed for compatibility with faker v10+
// The default locale works fine for generating mock data

/**
 * Mock data generators
 */

// Generate mock user
export function generateUser(overrides?: Partial<User>): User {
  return {
    id: faker.string.uuid(),
    name: faker.person.fullName(),
    email: faker.internet.email(),
    phone: `+62 8${faker.string.numeric(2)} ${faker.string.numeric(4)} ${faker.string.numeric(4)}`,
    role: faker.helpers.arrayElement(['admin', 'manager', 'staff']),
    avatar: faker.image.avatar(),
    createdAt: faker.date.past().toISOString(),
    ...overrides,
  };
}

// Generate mock product variant
export function generateProductVariant(productId: string, baseSku: string, index: number, isDefault: boolean = false): ProductVariant {
  const price = faker.number.int({ min: 50000, max: 10000000 });
  // ~20% chance of having a sale price
  const hasSale = faker.datatype.boolean({ probability: 0.2 });
  return {
    id: faker.string.uuid(),
    productId,
    name: isDefault ? 'Default' : `${faker.color.human()} - ${faker.helpers.arrayElement(['S', 'M', 'L', 'XL'])}`,
    sku: isDefault ? baseSku : `${baseSku}-${index}`,
    price,
    cost: Math.floor(price * faker.number.float({ min: 0.4, max: 0.7 })),
    stock: faker.number.int({ min: 0, max: 50 }),
    attributes: isDefault ? {} : {
      color: faker.color.human(),
      size: faker.helpers.arrayElement(['S', 'M', 'L', 'XL']),
    },
    isActive: true,
    salePrice: hasSale ? Math.floor(price * faker.number.float({ min: 0.5, max: 0.85 })) : undefined,
    saleStart: hasSale ? faker.date.recent({ days: 7 }).toISOString() : undefined,
    saleEnd: hasSale ? faker.date.soon({ days: 30 }).toISOString() : undefined,
  };
}

// Generate mock product
export function generateProduct(overrides?: Partial<Product>): Product {
  const id = faker.string.uuid();
  const name = faker.commerce.productName();
  const category = faker.helpers.arrayElement([
    'Elektronik',
    'Aksesoris',
    'Pakaian',
    'Makanan',
  ]);
  const baseSku = `SKU-${faker.string.alphanumeric(6).toUpperCase()}`;

  // Randomly decide if product has multiple variants or just default
  const hasMultipleVariants = faker.datatype.boolean();
  const variantCount = hasMultipleVariants ? faker.number.int({ min: 2, max: 5 }) : 1;

  const variants: ProductVariant[] = [];
  for (let i = 0; i < variantCount; i++) {
    variants.push(generateProductVariant(id, baseSku, i + 1, variantCount === 1));
  }

  return {
    id,
    name,
    category,
    image: faker.image.url(),
    description: faker.commerce.productDescription(),
    isActive: true,
    variants,
    createdAt: faker.date.past().toISOString(),
    updatedAt: faker.date.recent().toISOString(),
    ...overrides,
  };
}

// Generate mock order
export function generateOrder(overrides?: Partial<Order>): Order {
  const items = Array.from({ length: faker.number.int({ min: 1, max: 5 }) }, () => ({
    id: faker.string.uuid(),
    productId: faker.string.uuid(),
    productName: faker.commerce.productName(),
    quantity: faker.number.int({ min: 1, max: 5 }),
    price: faker.number.int({ min: 50000, max: 1000000 }),
  }));

  const subtotal = items.reduce((sum, item) => sum + item.price * item.quantity, 0);

  return {
    id: faker.string.uuid(),
    orderNumber: `#ORD-${faker.string.numeric(4)}`,
    customerId: faker.string.uuid(),
    customerName: faker.person.fullName(),
    items,
    subtotal,
    tax: subtotal * 0.1,
    total: subtotal * 1.1,
    status: faker.helpers.arrayElement(['pending', 'processing', 'completed', 'cancelled']),
    paymentMethod: faker.helpers.arrayElement(['cash', 'transfer', 'card']),
    createdAt: faker.date.past().toISOString(),
    updatedAt: faker.date.recent().toISOString(),
    ...overrides,
  };
}

// Generate mock customer
export function generateCustomer(overrides?: Partial<Customer>): Customer {
  const totalPurchase = faker.number.int({ min: 0, max: 50000000 });
  const orderCount = faker.number.int({ min: 0, max: 50 });

  return {
    id: faker.string.uuid(),
    customerId: `#CUST-${faker.string.numeric(3)}`,
    name: faker.person.fullName(),
    email: faker.internet.email(),
    phone: `+62 8${faker.string.numeric(2)} ${faker.string.numeric(4)} ${faker.string.numeric(4)}`,
    avatar: faker.image.avatar(),
    totalPurchase,
    orderCount,
    status: totalPurchase > 10000000 ? 'vip' : faker.helpers.arrayElement(['active', 'inactive']),
    lastOrderDate: faker.date.recent().toISOString(),
    createdAt: faker.date.past().toISOString(),
    ...overrides,
  };
}

// Generate mock branch
let branchIdCounter = 1;
export function generateBranch(overrides?: Partial<Branch>): Branch {
  const now = new Date().toISOString();
  return {
    id: branchIdCounter++,
    code: `BR-${String(branchIdCounter).padStart(3, '0')}`,
    name: `Cabang ${faker.location.city()}`,
    address: faker.location.streetAddress(),
    phone: `+62 21 ${faker.string.numeric(4)} ${faker.string.numeric(4)}`,
    is_active: faker.helpers.arrayElement([true, true, true, false]),
    created_at: faker.date.past().toISOString(),
    updated_at: now,
    ...overrides,
  };
}

// Generate dashboard stats
export function generateDashboardStats(): DashboardStats {
  return {
    totalSales: faker.number.int({ min: 100000000, max: 500000000 }),
    salesGrowth: faker.number.float({ min: -10, max: 20, fractionDigits: 1 }),
    newOrders: faker.number.int({ min: 50, max: 200 }),
    ordersGrowth: faker.number.float({ min: -10, max: 20, fractionDigits: 1 }),
    newCustomers: faker.number.int({ min: 10, max: 100 }),
    customersGrowth: faker.number.float({ min: -10, max: 20, fractionDigits: 1 }),
    netIncome: faker.number.int({ min: 30000000, max: 150000000 }),
    incomeGrowth: faker.number.float({ min: -10, max: 20, fractionDigits: 1 }),
  };
}

// Generate sales trend data for the last 30 days
export function generateSalesTrend(): SalesTrendData[] {
  const data: SalesTrendData[] = [];
  const today = new Date();

  for (let i = 29; i >= 0; i--) {
    const date = new Date(today);
    date.setDate(date.getDate() - i);

    data.push({
      date: date.toISOString().split('T')[0],
      sales: faker.number.int({ min: 2000000, max: 15000000 }),
      orders: faker.number.int({ min: 10, max: 50 }),
    });
  }

  return data;
}

// Type definitions
export interface User {
  id: string;
  name: string;
  email: string;
  phone: string;
  role: string;
  avatar: string;
  createdAt: string;
}

export interface Product {
  id: string;
  name: string;
  category: string;
  image: string;
  description: string;
  isActive: boolean;
  variants: ProductVariant[];
  createdAt: string;
  updatedAt: string;
}

export interface ProductVariant {
  id: string;
  productId: string;
  name: string;
  sku: string;
  price: number;
  cost: number;
  stock: number;
  attributes: Record<string, string>;
  isActive: boolean;
  salePrice?: number;
  saleStart?: string;
  saleEnd?: string;
}

export interface Order {
  id: string;
  orderNumber: string;
  customerId: string;
  customerName: string;
  customerEmail?: string;
  customerPhone?: string;
  items: OrderItem[];
  subtotal: number;
  tax: number;
  total: number;
  status: string;
  paymentMethod: string;
  shippingAddress?: string;
  notes?: string;
  createdAt: string;
  updatedAt: string;
}

export interface OrderItem {
  id: string;
  productId: string;
  productName: string;
  quantity: number;
  price: number;
}

export interface Customer {
  id: string;
  customerId: string;
  name: string;
  email: string;
  phone: string;
  avatar: string;
  totalPurchase: number;
  orderCount: number;
  status: string;
  lastOrderDate: string;
  createdAt: string;
}

export interface Branch {
  id: number;
  code: string;
  name: string;
  address: string | null;
  phone: string | null;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface Promotion {
  id: number;
  code: string;
  name: string;
  description: string | null;
  type: 'discount' | 'bundle' | 'conditional';
  discount_type: 'percentage' | 'fixed' | null;
  discount_value: number | null;
  min_purchase: number | null;
  max_discount: number | null;
  start_at: string;
  end_at: string | null;
  priority: number;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface DashboardStats {
  totalSales: number;
  salesGrowth: number;
  newOrders: number;
  ordersGrowth: number;
  newCustomers: number;
  customersGrowth: number;
  netIncome: number;
  incomeGrowth: number;
}

export interface SalesTrendData {
  date: string;
  sales: number;
  orders: number;
}

// Generate initial mock data
export const mockUsers = [
  {
    id: 'admin-001',
    name: 'Admin User',
    email: 'admin@nexpos.com',
    phone: '+62 812 3456 7890',
    role: 'admin',
    avatar: '',
    createdAt: new Date().toISOString(),
  },
  ...Array.from({ length: 9 }, () => generateUser()),
];
export const mockProducts = Array.from({ length: 50 }, () => generateProduct());
export const mockOrders = Array.from({ length: 100 }, () => generateOrder());
export const mockCustomers = Array.from({ length: 200 }, () => generateCustomer());
export const mockBranches: Branch[] = [
  { id: 1, code: 'JKT-001', name: 'Cabang Pusat Jakarta', address: 'Jl. Jendral Sudirman Kav. 52-53, SCBD', phone: '021-555-0192', is_active: true, created_at: '2024-01-15T08:00:00Z', updated_at: '2024-01-15T08:00:00Z' },
  { id: 2, code: 'BDG-004', name: 'Cabang Bandung Indah', address: 'Jl. Braga No. 99, Sumur Bandung', phone: '022-423-8888', is_active: true, created_at: '2024-02-10T08:00:00Z', updated_at: '2024-02-10T08:00:00Z' },
  { id: 3, code: 'SBY-002', name: 'Outlet Surabaya Timur', address: 'Jl. Kertajaya Indah No. 45', phone: '031-594-7721', is_active: false, created_at: '2024-03-05T08:00:00Z', updated_at: '2024-03-05T08:00:00Z' },
  { id: 4, code: 'MDN-005', name: 'Cabang Medan Fair', address: 'Jl. Gatot Subroto No. 30', phone: '061-452-1109', is_active: false, created_at: '2024-03-20T08:00:00Z', updated_at: '2024-03-20T08:00:00Z' },
  { id: 5, code: 'DPS-008', name: 'Cabang Bali Sanur', address: 'Jl. Danau Tamblingan No. 12', phone: '0361-288-555', is_active: true, created_at: '2024-04-01T08:00:00Z', updated_at: '2024-04-01T08:00:00Z' },
  { id: 6, code: 'JKT-003', name: 'Cabang Jakarta Barat', address: 'Jl. Tanjung Duren Raya No. 7', phone: '021-566-1234', is_active: true, created_at: '2024-04-15T08:00:00Z', updated_at: '2024-04-15T08:00:00Z' },
  { id: 7, code: 'BKS-006', name: 'Cabang Bekasi Timur', address: 'Jl. Cut Meutia No. 55, Margahayu', phone: '021-880-2211', is_active: true, created_at: '2024-05-01T08:00:00Z', updated_at: '2024-05-01T08:00:00Z' },
  { id: 8, code: 'YGY-007', name: 'Cabang Yogyakarta', address: 'Jl. Malioboro No. 132', phone: '0274-587-890', is_active: false, created_at: '2024-05-20T08:00:00Z', updated_at: '2024-05-20T08:00:00Z' },
  { id: 9, code: 'DPS-009', name: 'Outlet Denpasar', address: 'Jl. Teuku Umar No. 88', phone: '0361-723-456', is_active: true, created_at: '2024-06-01T08:00:00Z', updated_at: '2024-06-01T08:00:00Z' },
  { id: 10, code: 'SMG-010', name: 'Cabang Semarang', address: 'Jl. Pemuda No. 150', phone: '024-354-7788', is_active: true, created_at: '2024-06-15T08:00:00Z', updated_at: '2024-06-15T08:00:00Z' },
  { id: 11, code: 'DPK-011', name: 'Cabang Depok', address: 'Jl. Margonda Raya No. 204', phone: '021-776-5500', is_active: false, created_at: '2024-07-01T08:00:00Z', updated_at: '2024-07-01T08:00:00Z' },
  { id: 12, code: 'JKT-012', name: 'Cabang Jakarta Utara', address: 'Jl. RE Martadinata No. 9', phone: '021-691-3300', is_active: true, created_at: '2024-07-15T08:00:00Z', updated_at: '2024-07-15T08:00:00Z' },
];
export const mockPromotions: Promotion[] = [
  { id: 1, code: 'DISCOUNT50', name: 'Diskon 50%', description: 'Diskon 50% untuk semua produk', type: 'discount', discount_type: 'percentage', discount_value: 50, min_purchase: 100000, max_discount: 50000, start_at: '2026-02-01T00:00:00Z', end_at: '2026-02-28T23:59:59Z', priority: 10, is_active: true, created_at: '2026-01-15T08:00:00Z', updated_at: '2026-01-15T08:00:00Z' },
  { id: 2, code: 'NEWYEAR2026', name: 'Tahun Baru 2026', description: 'Potongan Rp 20.000 untuk pembelian min. Rp 150.000', type: 'discount', discount_type: 'fixed', discount_value: 20000, min_purchase: 150000, max_discount: null, start_at: '2026-01-01T00:00:00Z', end_at: '2026-01-31T23:59:59Z', priority: 8, is_active: false, created_at: '2025-12-20T08:00:00Z', updated_at: '2026-02-01T08:00:00Z' },
  { id: 3, code: 'FLASHSALE', name: 'Flash Sale Weekend', description: 'Diskon hingga 30% setiap weekend', type: 'discount', discount_type: 'percentage', discount_value: 30, min_purchase: 50000, max_discount: 100000, start_at: '2026-02-15T00:00:00Z', end_at: null, priority: 5, is_active: true, created_at: '2026-02-10T08:00:00Z', updated_at: '2026-02-10T08:00:00Z' },
  { id: 4, code: 'BUNDLE3FOR2', name: 'Beli 3 Gratis 1', description: 'Paket bundling produk pilihan', type: 'bundle', discount_type: null, discount_value: null, min_purchase: null, max_discount: null, start_at: '2026-02-01T00:00:00Z', end_at: '2026-03-31T23:59:59Z', priority: 7, is_active: true, created_at: '2026-01-25T08:00:00Z', updated_at: '2026-01-25T08:00:00Z' },
  { id: 5, code: 'FREESHIP', name: 'Gratis Ongkir', description: 'Bebas ongkir untuk pembelian di atas Rp 200.000', type: 'conditional', discount_type: null, discount_value: null, min_purchase: 200000, max_discount: null, start_at: '2026-02-01T00:00:00Z', end_at: null, priority: 3, is_active: true, created_at: '2026-01-20T08:00:00Z', updated_at: '2026-01-20T08:00:00Z' },
];

export const mockDashboardStats = generateDashboardStats();
export const mockSalesTrend = generateSalesTrend();

// ============== Inventory / Stock Movement ==============

// Min stock per variant (keyed by variant id)
export const variantMinStock: Record<string, number> = {};
for (const product of mockProducts) {
  for (const variant of product.variants) {
    variantMinStock[variant.id] = faker.number.int({ min: 3, max: 15 });
  }
}

export interface MockStockMovement {
  id: string;
  product_id: string;
  variant_id: string;
  product_name: string;
  variant_name: string;
  sku: string;
  type: 'IN' | 'OUT' | 'ADJUST';
  quantity: number;
  stock_before: number;
  stock_after: number;
  reference: string | null;
  notes: string | null;
  created_at: string;
  created_by: string;
}

export function generateStockMovement(product: Product, variant: ProductVariant): MockStockMovement {
  const type = faker.helpers.arrayElement(['IN', 'OUT', 'ADJUST'] as const);
  const quantity = faker.number.int({ min: 1, max: 20 });
  const stockBefore = faker.number.int({ min: 0, max: 100 });
  const stockAfter = type === 'IN'
    ? stockBefore + quantity
    : type === 'OUT'
      ? Math.max(0, stockBefore - quantity)
      : stockBefore + faker.number.int({ min: -10, max: 10 });

  const references: Record<string, string | null> = {
    IN: `PO-${faker.string.numeric(5)}`,
    OUT: `SO-${faker.string.numeric(5)}`,
    ADJUST: null,
  };

  const notesByType: Record<string, string> = {
    IN: faker.helpers.arrayElement(['Pembelian dari supplier', 'Retur dari pelanggan', 'Transfer masuk dari cabang lain']),
    OUT: faker.helpers.arrayElement(['Penjualan', 'Barang rusak', 'Transfer keluar ke cabang lain']),
    ADJUST: faker.helpers.arrayElement(['Stock opname', 'Koreksi stok', 'Penyesuaian sistem']),
  };

  return {
    id: faker.string.uuid(),
    product_id: product.id,
    variant_id: variant.id,
    product_name: product.name,
    variant_name: variant.name,
    sku: variant.sku,
    type,
    quantity,
    stock_before: stockBefore,
    stock_after: stockAfter,
    reference: references[type],
    notes: notesByType[type],
    created_at: faker.date.recent({ days: 60 }).toISOString(),
    created_by: faker.helpers.arrayElement(mockUsers).name,
  };
}

export const mockStockMovements: MockStockMovement[] = Array.from({ length: 100 }, () => {
  const product = faker.helpers.arrayElement(mockProducts);
  const variant = faker.helpers.arrayElement(product.variants);
  return generateStockMovement(product, variant);
}).sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime());


// ============== Stock Opname ==============

export interface MockStockOpnameItem {
  id: number;
  stock_opname_id: number;
  product_variant_id: number;
  sku: string;
  product_name: string;
  variant_name: string;
  system_stock: number;
  counted_stock: number | null;
  variance_qty: number;
  unit_cost: number;
  variance_value: number;
  notes: string | null;
  counted_at: string | null;
}

export interface MockStockOpname {
  id: number;
  branch_id: number;
  opname_number: string;
  status: 'in_progress' | 'completed' | 'cancelled';
  notes: string | null;
  total_items: number;
  counted_items: number;
  total_variance_qty: number;
  total_variance_value: number;
  started_by: number | null;
  completed_by: number | null;
  started_at: string;
  completed_at: string | null;
  cancelled_at: string | null;
  items?: MockStockOpnameItem[];
  created_at: string;
  updated_at: string;
}

export const mockStockOpnames: MockStockOpname[] = [
  {
    id: 1,
    branch_id: 1,
    opname_number: 'OPN-20260512-0001',
    status: 'completed',
    notes: 'Stock opname bulanan',
    total_items: 24,
    counted_items: 24,
    total_variance_qty: -8,
    total_variance_value: -45000,
    started_by: 1,
    completed_by: 1,
    started_at: '2026-05-12T08:00:00Z',
    completed_at: '2026-05-12T15:30:00Z',
    cancelled_at: null,
    created_at: '2026-05-12T08:00:00Z',
    updated_at: '2026-05-12T15:30:00Z',
  },
  {
    id: 2,
    branch_id: 1,
    opname_number: 'OPN-20260514-0001',
    status: 'in_progress',
    notes: null,
    total_items: 32,
    counted_items: 12,
    total_variance_qty: 0,
    total_variance_value: 0,
    started_by: 1,
    completed_by: null,
    started_at: '2026-05-14T09:15:00Z',
    completed_at: null,
    cancelled_at: null,
    created_at: '2026-05-14T09:15:00Z',
    updated_at: '2026-05-14T09:15:00Z',
  },
];

// Pre-populate items for opname id=2 so the count session has something to show
const opnameItemsByOpnameId: Record<number, MockStockOpnameItem[]> = { 1: [], 2: [] };
let nextOpnameItemId = 1;

(function seedOpnameItems() {
  const variantsSeen = new Set<string>();
  const flat: { product: Product; variant: ProductVariant }[] = [];
  for (const p of mockProducts) {
    for (const v of p.variants) {
      if (variantsSeen.has(v.id)) continue;
      variantsSeen.add(v.id);
      flat.push({ product: p, variant: v });
    }
  }
  const seed = flat.slice(0, 32);
  seed.forEach((entry, idx) => {
    const counted = idx < 12 ? entry.variant.stock + faker.number.int({ min: -3, max: 2 }) : null;
    const variance = counted != null ? counted - entry.variant.stock : 0;
    const unitCost = entry.variant.cost;
    opnameItemsByOpnameId[2].push({
      id: nextOpnameItemId++,
      stock_opname_id: 2,
      product_variant_id: parseInt(entry.variant.id),
      sku: entry.variant.sku,
      product_name: entry.product.name,
      variant_name: entry.variant.name,
      system_stock: entry.variant.stock,
      counted_stock: counted,
      variance_qty: variance,
      unit_cost: unitCost,
      variance_value: variance * unitCost,
      notes: null,
      counted_at: counted != null ? faker.date.recent({ days: 1 }).toISOString() : null,
    });
  });
})();

export const opnameItems = opnameItemsByOpnameId;
export function nextOpnameItemIdValue() {
  return nextOpnameItemId++;
}

// ============== Daily Settlement ==============

export type SettlementPaymentMethod =
  | 'cash'
  | 'debit_card'
  | 'credit_card'
  | 'e_wallet'
  | 'bank_transfer'
  | 'qris';

export const SETTLEMENT_METHODS: SettlementPaymentMethod[] = [
  'cash',
  'debit_card',
  'credit_card',
  'e_wallet',
  'bank_transfer',
  'qris',
];

export interface MockDailySettlementItem {
  id: number;
  payment_method: SettlementPaymentMethod;
  expected_amount: number;
  actual_amount: number | null;
  variance_amount: number;
  notes: string | null;
}

export interface MockDailySettlement {
  id: number;
  branch_id: number;
  settlement_date: string;
  status: 'draft' | 'finalized';
  total_sales: number;
  total_refunds: number;
  total_expenses: number;
  total_expected: number;
  total_actual: number;
  total_variance: number;
  notes: string | null;
  recorded_by: number | null;
  finalized_by: number | null;
  finalized_at: string | null;
  items?: MockDailySettlementItem[];
  created_at: string;
  updated_at: string;
}

export const mockDailySettlements: MockDailySettlement[] = [];
export const dailySettlementItems: Record<number, MockDailySettlementItem[]> = {};
let nextSettlementItemId = 1;

export function nextSettlementItemIdValue() {
  return nextSettlementItemId++;
}

// Company Settings mock
export const mockCompanySettings = {
  id: 1,
  company_id: 1,
  tax_enabled: false,
  tax_rate: 11,
  tax_inclusive: true,
  rounding_enabled: false,
  rounding_amount: 100,
  auto_complete_counter_orders: true,
  require_customer_for_delivery: true,
  receipt_header: null as string | null,
  receipt_footer: null as string | null,
  show_tax_on_receipt: true,
  offline_mode_enabled: false,
  max_offline_days: 7,
  created_at: new Date().toISOString(),
  updated_at: new Date().toISOString(),
};

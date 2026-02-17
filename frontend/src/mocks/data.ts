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
export function generateBranch(overrides?: Partial<Branch>): Branch {
  const regions = ['DKI Jakarta', 'Jawa Barat', 'Jawa Timur', 'Sumatera Utara', 'Bali'];
  const region = faker.helpers.arrayElement(regions);
  return {
    id: faker.string.uuid(),
    name: `Cabang ${faker.location.city()}`,
    code: faker.string.alphanumeric(3).toUpperCase(),
    address: faker.location.streetAddress(),
    city: faker.location.city(),
    region,
    phone: `+62 21 ${faker.string.numeric(4)} ${faker.string.numeric(4)}`,
    manager: faker.person.fullName(),
    status: faker.helpers.arrayElement(['active', 'active', 'active', 'inactive', 'maintenance']),
    createdAt: faker.date.past().toISOString(),
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
  id: string;
  name: string;
  code: string;
  address: string;
  city: string;
  region: string;
  phone: string;
  manager: string;
  status: string;
  createdAt: string;
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
    email: 'admin@aplikabizz.com',
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
  generateBranch({ name: 'Cabang Pusat Jakarta', code: 'JKT-001', city: 'Jakarta Selatan', region: 'DKI Jakarta', address: 'Jl. Jendral Sudirman Kav. 52-53, SCBD', phone: '021-555-0192', status: 'active' }),
  generateBranch({ name: 'Cabang Bandung Indah', code: 'BDG-004', city: 'Kota Bandung', region: 'Jawa Barat', address: 'Jl. Braga No. 99, Sumur Bandung', phone: '022-423-8888', status: 'active' }),
  generateBranch({ name: 'Outlet Surabaya Timur', code: 'SBY-002', city: 'Surabaya', region: 'Jawa Timur', address: 'Jl. Kertajaya Indah No. 45', phone: '031-594-7721', status: 'inactive' }),
  generateBranch({ name: 'Cabang Medan Fair', code: 'MDN-005', city: 'Medan', region: 'Sumatera Utara', address: 'Jl. Gatot Subroto No. 30', phone: '061-452-1109', status: 'maintenance' }),
  generateBranch({ name: 'Cabang Bali Sanur', code: 'DPS-008', city: 'Sanur', region: 'Bali', address: 'Jl. Danau Tamblingan No. 12', phone: '0361-288-555', status: 'active' }),
  generateBranch({ name: 'Cabang Jakarta Barat', code: 'JKT-003', city: 'Jakarta Barat', region: 'DKI Jakarta', address: 'Jl. Tanjung Duren Raya No. 7', phone: '021-566-1234', status: 'active' }),
  generateBranch({ name: 'Cabang Bekasi Timur', code: 'BKS-006', city: 'Bekasi', region: 'Jawa Barat', address: 'Jl. Cut Meutia No. 55, Margahayu', phone: '021-880-2211', status: 'active' }),
  generateBranch({ name: 'Cabang Yogyakarta', code: 'YGY-007', city: 'Yogyakarta', region: 'Jawa Timur', address: 'Jl. Malioboro No. 132', phone: '0274-587-890', status: 'inactive' }),
  generateBranch({ name: 'Outlet Denpasar', code: 'DPS-009', city: 'Denpasar', region: 'Bali', address: 'Jl. Teuku Umar No. 88', phone: '0361-723-456', status: 'active' }),
  generateBranch({ name: 'Cabang Semarang', code: 'SMG-010', city: 'Semarang', region: 'Jawa Timur', address: 'Jl. Pemuda No. 150', phone: '024-354-7788', status: 'active' }),
  generateBranch({ name: 'Cabang Depok', code: 'DPK-011', city: 'Depok', region: 'Jawa Barat', address: 'Jl. Margonda Raya No. 204', phone: '021-776-5500', status: 'maintenance' }),
  generateBranch({ name: 'Cabang Jakarta Utara', code: 'JKT-012', city: 'Jakarta Utara', region: 'DKI Jakarta', address: 'Jl. RE Martadinata No. 9', phone: '021-691-3300', status: 'active' }),
];
export const mockDashboardStats = generateDashboardStats();
export const mockSalesTrend = generateSalesTrend();

// Products are already generated with variants in generateProduct function

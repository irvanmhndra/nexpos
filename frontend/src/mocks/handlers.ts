import { http, HttpResponse } from 'msw';

interface RawOrderItem {
  price: number;
  quantity: number;
  product_variant_id: number;
  discount_amount?: number;
}
interface RawOrderBody {
  items: RawOrderItem[];
  customer_id?: string | null;
  promo_code?: string;
  fulfillment_type?: string;
  payments?: { method: string; amount: number }[];
}
import {
  mockUsers,
  mockProducts,
  mockOrders,
  mockCustomers,
  mockBranches,
  mockPromotions,
  mockDashboardStats,
  mockSalesTrend,
  mockStockMovements,
  variantMinStock,
  mockCompanySettings,
  mockStockOpnames,
  opnameItems,
  nextOpnameItemIdValue,
} from './data';
import type { MockStockOpname, MockStockOpnameItem } from './data';

// Return sale price if active, otherwise normal price
function getEffectivePrice(v: { price: number; salePrice?: number; saleStart?: string; saleEnd?: string }): number {
  const now = new Date();
  if (
    v.salePrice != null &&
    v.salePrice < v.price &&
    (!v.saleStart || new Date(v.saleStart) <= now) &&
    (!v.saleEnd || new Date(v.saleEnd) >= now)
  ) {
    return v.salePrice;
  }
  return v.price;
}

// MSW handlers match the actual axios baseURL
const API_BASE_URL = 'http://localhost:8080/api/v1';

export const handlers = [
  // Auth endpoints
  http.post(`${API_BASE_URL}/auth/login`, async ({ request }) => {
    const body = await request.json() as { email: string; password: string };
    const user = mockUsers.find(u => u.email === body.email);

    if (user && body.password === 'password') {
      return HttpResponse.json({
        user,
        token: 'mock-jwt-token',
      });
    }

    return HttpResponse.json({ error: 'Invalid credentials' }, { status: 401 });
  }),

  http.post(`${API_BASE_URL}/auth/register`, async ({ request }) => {
    const body = await request.json() as { name: string; email: string; password: string };

    // Check if email already exists
    const existingUser = mockUsers.find(u => u.email === body.email);
    if (existingUser) {
      return HttpResponse.json({ error: 'Email sudah terdaftar' }, { status: 400 });
    }

    // Create new user
    const newUser: User = {
      id: crypto.randomUUID(),
      name: body.name,
      email: body.email,
      phone: '',
      role: 'user',
      avatar: '',
      createdAt: new Date().toISOString(),
    };

    mockUsers.push(newUser);

    return HttpResponse.json({
      user: newUser,
      token: 'mock-jwt-token-' + newUser.id,
    }, { status: 201 });
  }),

  http.post(`${API_BASE_URL}/auth/logout`, () => {
    return HttpResponse.json({ message: 'Logged out successfully' });
  }),

  // Dashboard endpoints
  http.get(`${API_BASE_URL}/dashboard/stats`, () => {
    return HttpResponse.json(mockDashboardStats);
  }),

  http.get(`${API_BASE_URL}/dashboard/recent-orders`, () => {
    return HttpResponse.json(mockOrders.slice(0, 5));
  }),

  http.get(`${API_BASE_URL}/dashboard/sales-trend`, () => {
    return HttpResponse.json(mockSalesTrend);
  }),

  // Products endpoints
  http.get(`${API_BASE_URL}/products`, ({ request }) => {
    const url = new URL(request.url);
    const search = url.searchParams.get('search') || '';
    const page = parseInt(url.searchParams.get('page') || '1');
    const limit = parseInt(url.searchParams.get('limit') || '10');

    let filtered = mockProducts;

    if (search) {
      filtered = filtered.filter(
        p =>
          p.name.toLowerCase().includes(search.toLowerCase()) ||
          p.variants?.some(v => v.sku.toLowerCase().includes(search.toLowerCase()))
      );
    }

    const start = (page - 1) * limit;
    const end = start + limit;
    const paginated = filtered.slice(start, end);

    return HttpResponse.json({
      data: paginated,
      total: filtered.length,
      page,
      limit,
      totalPages: Math.ceil(filtered.length / limit),
    });
  }),

  http.get(`${API_BASE_URL}/products/:id`, ({ params }) => {
    const product = mockProducts.find(p => p.id === params.id);

    if (!product) {
      return HttpResponse.json({ error: 'Product not found' }, { status: 404 });
    }

    return HttpResponse.json(product);
  }),

  http.post(`${API_BASE_URL}/products`, async ({ request }) => {
    const body = await request.json() as Partial<Product>;
    const newProduct = {
      ...body,
      id: crypto.randomUUID(),
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
    };

    mockProducts.push(newProduct as Product);

    return HttpResponse.json(newProduct, { status: 201 });
  }),

  // Orders endpoints
  http.get(`${API_BASE_URL}/orders`, ({ request }) => {
    const url = new URL(request.url);
    const page = parseInt(url.searchParams.get('page') || '1');
    const limit = parseInt(url.searchParams.get('limit') || '10');
    const status = url.searchParams.get('status');

    let filtered = mockOrders;

    if (status) {
      filtered = filtered.filter(o => o.status === status);
    }

    const start = (page - 1) * limit;
    const end = start + limit;
    const paginated = filtered.slice(start, end);

    return HttpResponse.json({
      data: paginated,
      total: filtered.length,
      page,
      limit,
      totalPages: Math.ceil(filtered.length / limit),
    });
  }),

  http.get(`${API_BASE_URL}/orders/:id`, ({ params }) => {
    const order = mockOrders.find(o => o.id === params.id);

    if (!order) {
      return HttpResponse.json({ error: 'Order not found' }, { status: 404 });
    }

    return HttpResponse.json(order);
  }),

  http.post(`${API_BASE_URL}/orders/preview`, async ({ request }) => {
    const body = await request.json() as {
      items: Array<{ product_variant_id: number; quantity: number; discount_amount?: number }>;
      promo_code?: string;
    };

    const settings = mockCompanySettings;
    let totalAmount = 0;
    let itemDiscount = 0;
    let totalTax = 0;

    for (const item of body.items) {
      let variant = null;
      for (const product of mockProducts) {
        for (const v of product.variants) {
          if (parseInt(v.id) === item.product_variant_id) {
            variant = v;
            break;
          }
        }
        if (variant) break;
      }
      if (!variant) {
        return HttpResponse.json({ success: false, message: 'Produk tidak ditemukan' }, { status: 404 });
      }

      const qty = item.quantity;
      const discount = item.discount_amount ?? 0;
      const unitPrice = getEffectivePrice(variant);
      const subtotalItem = (unitPrice * qty) - discount;

      let taxAmount = 0;
      if (settings.tax_enabled) {
        if (settings.tax_inclusive) {
          taxAmount = subtotalItem * settings.tax_rate / (100 + settings.tax_rate);
        } else {
          taxAmount = subtotalItem * settings.tax_rate / 100;
        }
      }

      totalAmount += unitPrice * qty;
      itemDiscount += discount;
      totalTax += taxAmount;
    }

    const subtotal = totalAmount - itemDiscount;
    let promoDiscount = 0;
    let appliedPromotion = null;

    if (body.promo_code) {
      const now = new Date();
      const promo = mockPromotions.find(p =>
        p.code.toUpperCase() === body.promo_code!.toUpperCase() &&
        p.is_active &&
        p.discount_type != null &&
        p.discount_value != null &&
        new Date(p.start_at) <= now &&
        (!p.end_at || new Date(p.end_at) >= now) &&
        (!p.min_purchase || subtotal >= p.min_purchase)
      );
      if (promo && promo.discount_type && promo.discount_value != null) {
        if (promo.discount_type === 'percentage') {
          promoDiscount = subtotal * promo.discount_value / 100;
          if (promo.max_discount) promoDiscount = Math.min(promoDiscount, promo.max_discount);
        } else {
          promoDiscount = Math.min(promo.discount_value, subtotal);
        }
        appliedPromotion = { id: promo.id, code: promo.code, name: promo.name, discount_amount: promoDiscount };
      }
    }

    const grandTotal = subtotal - promoDiscount + totalTax;

    return HttpResponse.json({
      success: true,
      data: {
        subtotal,
        item_discount: itemDiscount,
        promo_discount: promoDiscount,
        tax: totalTax,
        grand_total: grandTotal,
        applied_promotion: appliedPromotion,
      },
    });
  }),

  http.post(`${API_BASE_URL}/orders`, async ({ request }) => {
    const body = await request.json() as RawOrderBody;
    const newOrder = {
      ...body,
      id: crypto.randomUUID(),
      orderNumber: `ORD-${Math.floor(Math.random() * 100000)}`,
      customerId: `CUST-${Math.floor(Math.random() * 1000)}`,
      subtotal: body.items.reduce((sum: number, item: RawOrderItem) => sum + item.price * item.quantity, 0),
      tax: body.items.reduce((sum: number, item: RawOrderItem) => sum + item.price * item.quantity, 0) * 0.1,
      total: body.items.reduce((sum: number, item: RawOrderItem) => sum + item.price * item.quantity, 0) * 1.1,
      items: body.items.map((item: RawOrderItem) => ({
        ...item,
        id: crypto.randomUUID(),
      })),
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
    };

    mockOrders.unshift(newOrder as unknown as Order);

    return HttpResponse.json(newOrder, { status: 201 });
  }),

  http.put(`${API_BASE_URL}/orders/:id`, async ({ params, request }) => {
    const body = await request.json() as RawOrderBody;
    const orderIndex = mockOrders.findIndex(o => o.id === params.id);

    if (orderIndex === -1) {
      return HttpResponse.json({ error: 'Order not found' }, { status: 404 });
    }

    const updatedOrder = {
      ...mockOrders[orderIndex],
      ...body,
      id: mockOrders[orderIndex].id,
      orderNumber: mockOrders[orderIndex].orderNumber,
      customerId: mockOrders[orderIndex].customerId,
      subtotal: body.items.reduce((sum: number, item: RawOrderItem) => sum + item.price * item.quantity, 0),
      tax: body.items.reduce((sum: number, item: RawOrderItem) => sum + item.price * item.quantity, 0) * 0.1,
      total: body.items.reduce((sum: number, item: RawOrderItem) => sum + item.price * item.quantity, 0) * 1.1,
      items: body.items.map((item: RawOrderItem, index: number) => ({
        ...item,
        id: mockOrders[orderIndex].items[index]?.id || crypto.randomUUID(),
      })),
      updatedAt: new Date().toISOString(),
    };

    mockOrders[orderIndex] = updatedOrder as unknown as Order;

    return HttpResponse.json(updatedOrder);
  }),

  http.patch(`${API_BASE_URL}/orders/:id/status`, async ({ params, request }) => {
    const body = await request.json() as { status: string };
    const orderIndex = mockOrders.findIndex(o => o.id === params.id);

    if (orderIndex === -1) {
      return HttpResponse.json({ error: 'Order not found' }, { status: 404 });
    }

    mockOrders[orderIndex].status = body.status;
    mockOrders[orderIndex].updatedAt = new Date().toISOString();

    return HttpResponse.json(mockOrders[orderIndex]);
  }),

  http.delete(`${API_BASE_URL}/orders/:id`, ({ params }) => {
    const orderIndex = mockOrders.findIndex(o => o.id === params.id);

    if (orderIndex === -1) {
      return HttpResponse.json({ error: 'Order not found' }, { status: 404 });
    }

    mockOrders.splice(orderIndex, 1);

    return HttpResponse.json({ message: 'Order deleted successfully' });
  }),

  // Customers endpoints
  http.get(`${API_BASE_URL}/customers`, ({ request }) => {
    const url = new URL(request.url);
    const search = url.searchParams.get('search') || '';
    const page = parseInt(url.searchParams.get('page') || '1');
    const limit = parseInt(url.searchParams.get('limit') || '10');

    let filtered = mockCustomers;

    if (search) {
      filtered = filtered.filter(
        c =>
          c.name.toLowerCase().includes(search.toLowerCase()) ||
          c.email.toLowerCase().includes(search.toLowerCase())
      );
    }

    const start = (page - 1) * limit;
    const end = start + limit;
    const paginated = filtered.slice(start, end);

    return HttpResponse.json({
      data: paginated,
      total: filtered.length,
      page,
      limit,
      totalPages: Math.ceil(filtered.length / limit),
    });
  }),

  http.get(`${API_BASE_URL}/customers/:id`, ({ params }) => {
    const customer = mockCustomers.find(c => c.id === params.id);

    if (!customer) {
      return HttpResponse.json({ error: 'Customer not found' }, { status: 404 });
    }

    return HttpResponse.json(customer);
  }),

  // Branches endpoints
  http.get(`${API_BASE_URL}/branches`, ({ request }) => {
    const url = new URL(request.url);
    const search = url.searchParams.get('search') || '';
    const page = parseInt(url.searchParams.get('page') || '1');
    const perPage = parseInt(url.searchParams.get('per_page') || '10');
    const isActiveParam = url.searchParams.get('is_active');

    let filtered = [...mockBranches];

    if (search) {
      const q = search.toLowerCase();
      filtered = filtered.filter(
        b =>
          b.name.toLowerCase().includes(q) ||
          b.code.toLowerCase().includes(q) ||
          (b.address || '').toLowerCase().includes(q) ||
          (b.phone || '').includes(q)
      );
    }

    if (isActiveParam !== null) {
      const isActive = isActiveParam === 'true';
      filtered = filtered.filter(b => b.is_active === isActive);
    }

    const totalRecords = filtered.length;
    const totalPages = Math.ceil(totalRecords / perPage) || 1;
    const start = (page - 1) * perPage;
    const paginated = filtered.slice(start, start + perPage);

    return HttpResponse.json({
      success: true,
      message: 'Branch list retrieved',
      data: paginated,
      meta: {
        pagination: {
          total_records: totalRecords,
          total_pages: totalPages,
          current_page: page,
          per_page: perPage,
          next_page: page < totalPages ? page + 1 : null,
          prev_page: page > 1 ? page - 1 : null,
        },
        server_time: new Date().toISOString(),
      },
    });
  }),

  http.get(`${API_BASE_URL}/branches/:id`, ({ params }) => {
    const id = parseInt(params.id as string);
    const branch = mockBranches.find(b => b.id === id);
    if (!branch) {
      return HttpResponse.json({ success: false, message: 'Branch not found', error_code: 'NOT_FOUND' }, { status: 404 });
    }
    return HttpResponse.json({ success: true, message: 'Branch retrieved successfully', data: branch });
  }),

  http.post(`${API_BASE_URL}/branches`, async ({ request }) => {
    const body = await request.json() as Partial<Branch>;
    const nextId = mockBranches.length > 0 ? Math.max(...mockBranches.map(b => b.id)) + 1 : 1;
    const now = new Date().toISOString();
    const newBranch: Branch = {
      id: nextId,
      code: body.code || `BR-${String(nextId).padStart(3, '0')}`,
      name: body.name || '',
      address: body.address ?? null,
      phone: body.phone ?? null,
      is_active: body.is_active !== undefined ? body.is_active : true,
      created_at: now,
      updated_at: now,
    };
    mockBranches.push(newBranch);
    return HttpResponse.json({ success: true, message: 'Branch created successfully', data: newBranch }, { status: 201 });
  }),

  http.put(`${API_BASE_URL}/branches/:id`, async ({ params, request }) => {
    const id = parseInt(params.id as string);
    const body = await request.json() as Partial<Branch>;
    const index = mockBranches.findIndex(b => b.id === id);
    if (index === -1) {
      return HttpResponse.json({ success: false, message: 'Branch not found', error_code: 'NOT_FOUND' }, { status: 404 });
    }
    mockBranches[index] = { ...mockBranches[index], ...body, id, updated_at: new Date().toISOString() };
    return HttpResponse.json({ success: true, message: 'Branch updated successfully', data: mockBranches[index] });
  }),

  http.delete(`${API_BASE_URL}/branches/:id`, ({ params }) => {
    const id = parseInt(params.id as string);
    const index = mockBranches.findIndex(b => b.id === id);
    if (index === -1) {
      return HttpResponse.json({ success: false, message: 'Branch not found', error_code: 'NOT_FOUND' }, { status: 404 });
    }
    mockBranches.splice(index, 1);
    return HttpResponse.json({ success: true, message: 'Branch deleted successfully', data: null });
  }),

  // Promotions endpoints
  http.get(`${API_BASE_URL}/promotions`, ({ request }) => {
    const url = new URL(request.url);
    const search = url.searchParams.get('search') || '';
    const page = parseInt(url.searchParams.get('page') || '1');
    const perPage = parseInt(url.searchParams.get('per_page') || '10');
    const isActiveParam = url.searchParams.get('is_active');
    const typeParam = url.searchParams.get('type');

    let filtered = [...mockPromotions];

    if (search) {
      const q = search.toLowerCase();
      filtered = filtered.filter(
        p =>
          p.name.toLowerCase().includes(q) ||
          p.code.toLowerCase().includes(q) ||
          (p.description || '').toLowerCase().includes(q)
      );
    }

    if (isActiveParam !== null) {
      const isActive = isActiveParam === 'true';
      filtered = filtered.filter(p => p.is_active === isActive);
    }

    if (typeParam) {
      filtered = filtered.filter(p => p.type === typeParam);
    }

    const totalRecords = filtered.length;
    const totalPages = Math.ceil(totalRecords / perPage) || 1;
    const start = (page - 1) * perPage;
    const paginated = filtered.slice(start, start + perPage);

    return HttpResponse.json({
      success: true,
      message: 'Promotion list retrieved',
      data: paginated,
      meta: {
        pagination: {
          total_records: totalRecords,
          total_pages: totalPages,
          current_page: page,
          per_page: perPage,
          next_page: page < totalPages ? page + 1 : null,
          prev_page: page > 1 ? page - 1 : null,
        },
        server_time: new Date().toISOString(),
      },
    });
  }),

  http.get(`${API_BASE_URL}/promotions/:id`, ({ params }) => {
    const id = parseInt(params.id as string);
    const promo = mockPromotions.find(p => p.id === id);
    if (!promo) {
      return HttpResponse.json({ success: false, message: 'Promotion not found', error_code: 'NOT_FOUND' }, { status: 404 });
    }
    return HttpResponse.json({ success: true, message: 'Promotion retrieved successfully', data: promo });
  }),

  http.post(`${API_BASE_URL}/promotions`, async ({ request }) => {
    const body = await request.json() as Partial<Promotion>;
    const nextId = mockPromotions.length > 0 ? Math.max(...mockPromotions.map(p => p.id)) + 1 : 1;
    const now = new Date().toISOString();
    const newPromo: Promotion = {
      id: nextId,
      code: body.code || `PROMO${nextId}`,
      name: body.name || '',
      description: body.description ?? null,
      type: body.type || 'discount',
      discount_type: body.discount_type ?? null,
      discount_value: body.discount_value ?? null,
      min_purchase: body.min_purchase ?? null,
      max_discount: body.max_discount ?? null,
      start_at: body.start_at || now,
      end_at: body.end_at ?? null,
      priority: body.priority ?? 0,
      is_active: body.is_active !== undefined ? body.is_active : true,
      created_at: now,
      updated_at: now,
    };
    mockPromotions.push(newPromo);
    return HttpResponse.json({ success: true, message: 'Promotion created successfully', data: newPromo }, { status: 201 });
  }),

  http.put(`${API_BASE_URL}/promotions/:id`, async ({ params, request }) => {
    const id = parseInt(params.id as string);
    const body = await request.json() as Partial<Promotion>;
    const index = mockPromotions.findIndex(p => p.id === id);
    if (index === -1) {
      return HttpResponse.json({ success: false, message: 'Promotion not found', error_code: 'NOT_FOUND' }, { status: 404 });
    }
    mockPromotions[index] = { ...mockPromotions[index], ...body, id, updated_at: new Date().toISOString() };
    return HttpResponse.json({ success: true, message: 'Promotion updated successfully', data: mockPromotions[index] });
  }),

  http.delete(`${API_BASE_URL}/promotions/:id`, ({ params }) => {
    const id = parseInt(params.id as string);
    const index = mockPromotions.findIndex(p => p.id === id);
    if (index === -1) {
      return HttpResponse.json({ success: false, message: 'Promotion not found', error_code: 'NOT_FOUND' }, { status: 404 });
    }
    mockPromotions.splice(index, 1);
    return HttpResponse.json({ success: true, message: 'Promotion deleted successfully', data: null });
  }),

  // ============== Inventory endpoints ==============

  // Stats must come BEFORE /:id style routes
  http.get(`${API_BASE_URL}/inventory/stats`, () => {
    let totalSku = 0;
    let lowStock = 0;
    let outOfStock = 0;
    let totalStockValue = 0;

    for (const product of mockProducts) {
      for (const variant of product.variants) {
        totalSku++;
        const minStock = variantMinStock[variant.id] ?? 5;
        if (variant.stock === 0) {
          outOfStock++;
        } else if (variant.stock <= minStock) {
          lowStock++;
        }
        totalStockValue += variant.stock * variant.cost;
      }
    }

    return HttpResponse.json({
      success: true,
      message: 'Inventory stats retrieved',
      data: { total_sku: totalSku, low_stock: lowStock, out_of_stock: outOfStock, total_stock_value: totalStockValue },
    });
  }),

  http.get(`${API_BASE_URL}/inventory/movements/stats`, () => {
    const now = new Date();
    const startOfMonth = new Date(now.getFullYear(), now.getMonth(), 1);
    const thisMonth = mockStockMovements.filter(m => new Date(m.created_at) >= startOfMonth);

    return HttpResponse.json({
      success: true,
      message: 'Movement stats retrieved',
      data: {
        total_movements: thisMonth.length,
        stock_in: thisMonth.filter(m => m.type === 'IN').length,
        stock_out: thisMonth.filter(m => m.type === 'OUT').length,
        adjustments: thisMonth.filter(m => m.type === 'ADJUST').length,
      },
    });
  }),

  http.get(`${API_BASE_URL}/inventory/movements`, ({ request }) => {
    const url = new URL(request.url);
    const search = url.searchParams.get('search') || '';
    const type = url.searchParams.get('type') || '';
    const startDate = url.searchParams.get('start_date') || '';
    const endDate = url.searchParams.get('end_date') || '';
    const page = parseInt(url.searchParams.get('page') || '1');
    const perPage = parseInt(url.searchParams.get('per_page') || '10');

    let filtered = [...mockStockMovements];

    if (search) {
      const q = search.toLowerCase();
      filtered = filtered.filter(
        m => m.product_name.toLowerCase().includes(q) || m.sku.toLowerCase().includes(q)
      );
    }
    if (type) {
      filtered = filtered.filter(m => m.type === type);
    }
    if (startDate) {
      filtered = filtered.filter(m => m.created_at >= startDate);
    }
    if (endDate) {
      filtered = filtered.filter(m => m.created_at <= endDate + 'T23:59:59Z');
    }

    const totalRecords = filtered.length;
    const totalPages = Math.ceil(totalRecords / perPage) || 1;
    const start = (page - 1) * perPage;
    const paginated = filtered.slice(start, start + perPage);

    return HttpResponse.json({
      success: true,
      message: 'Stock movements retrieved',
      data: paginated,
      meta: {
        pagination: {
          total_records: totalRecords,
          total_pages: totalPages,
          current_page: page,
          per_page: perPage,
          next_page: page < totalPages ? page + 1 : null,
          prev_page: page > 1 ? page - 1 : null,
        },
      },
    });
  }),

  http.get(`${API_BASE_URL}/inventory`, ({ request }) => {
    const url = new URL(request.url);
    const search = url.searchParams.get('search') || '';
    const category = url.searchParams.get('category') || '';
    const status = url.searchParams.get('status') || '';
    const page = parseInt(url.searchParams.get('page') || '1');
    const perPage = parseInt(url.searchParams.get('per_page') || '10');

    // Flatten products → inventory items
    type InventoryRow = {
      product_id: number;
      product_variant_id: number;
      product_name: string;
      variant_name: string;
      sku: string;
      category_name: string;
      image_data: string | null;
      branch_id: number;
      branch_name: string;
      current_stock: number;
      min_stock: number;
      stock_value: number;
      status: string;
      has_variants: boolean;
    };

    const items: InventoryRow[] = [];
    for (const product of mockProducts) {
      const hasVariants = product.variants.length > 1;
      for (const variant of product.variants) {
        const minStock = variantMinStock[variant.id] ?? 5;
        let st = 'normal';
        if (variant.stock === 0) st = 'out_of_stock';
        else if (variant.stock <= minStock) st = 'low';

        items.push({
          product_id: parseInt(product.id) || 0,
          product_variant_id: parseInt(variant.id) || 0,
          product_name: product.name,
          variant_name: variant.name,
          sku: variant.sku,
          category_name: product.category,
          image_data: product.image || null,
          branch_id: 1,
          branch_name: 'Cabang Utama',
          current_stock: variant.stock,
          min_stock: minStock,
          stock_value: variant.stock * variant.cost,
          status: st,
          has_variants: hasVariants,
        });
      }
    }

    let filtered = items;

    if (search) {
      const q = search.toLowerCase();
      filtered = filtered.filter(
        i => i.product_name.toLowerCase().includes(q) || i.sku.toLowerCase().includes(q)
      );
    }
    if (category) {
      filtered = filtered.filter(i => i.category_name === category);
    }
    if (status) {
      filtered = filtered.filter(i => i.status === status);
    }

    const totalRecords = filtered.length;
    const totalPages = Math.ceil(totalRecords / perPage) || 1;
    const start = (page - 1) * perPage;
    const paginated = filtered.slice(start, start + perPage);

    return HttpResponse.json({
      success: true,
      message: 'Inventory retrieved',
      data: paginated,
      meta: {
        pagination: {
          total_records: totalRecords,
          total_pages: totalPages,
          current_page: page,
          per_page: perPage,
          next_page: page < totalPages ? page + 1 : null,
          prev_page: page > 1 ? page - 1 : null,
        },
      },
    });
  }),

  http.post(`${API_BASE_URL}/inventory/adjust`, async ({ request }) => {
    const body = await request.json() as { variant_id: number; type: string; quantity: number; note?: string };

    // Find the variant in mockProducts and update stock
    let targetProduct: Product | undefined;
    let targetVariant: ProductVariant | undefined;

    for (const product of mockProducts) {
      const variant = product.variants.find(v => parseInt(v.id) === body.variant_id);
      if (variant) {
        targetProduct = product;
        targetVariant = variant;
        break;
      }
    }

    if (!targetProduct || !targetVariant) {
      return HttpResponse.json({ success: false, message: 'Variant not found' }, { status: 404 });
    }

    const stockBefore = targetVariant.stock;
    let stockAfter = stockBefore;

    if (body.type === 'IN') {
      stockAfter = stockBefore + body.quantity;
    } else if (body.type === 'OUT') {
      stockAfter = Math.max(0, stockBefore - body.quantity);
    } else {
      stockAfter = body.quantity; // adjustment sets exact value
    }

    targetVariant.stock = stockAfter;

    const movement: MockStockMovement = {
      id: crypto.randomUUID(),
      product_id: targetProduct.id,
      variant_id: targetVariant.id,
      product_name: targetProduct.name,
      variant_name: targetVariant.name,
      sku: targetVariant.sku,
      type: body.type as 'IN' | 'OUT' | 'ADJUST',
      quantity: body.quantity,
      stock_before: stockBefore,
      stock_after: stockAfter,
      reference: null,
      notes: body.note || null,
      created_at: new Date().toISOString(),
      created_by: 'Admin User',
    };

    mockStockMovements.unshift(movement);

    return HttpResponse.json({ success: true, message: 'Stock adjusted successfully', data: movement }, { status: 201 });
  }),

  http.put(`${API_BASE_URL}/inventory/:variantId/min-stock`, async ({ params, request }) => {
    const variantId = params.variantId as string;
    const body = await request.json() as { min_stock: number };
    variantMinStock[variantId] = body.min_stock;
    return HttpResponse.json({ success: true, message: 'Min stock updated successfully', data: null });
  }),

  // Company settings
  http.get(`${API_BASE_URL}/company-settings`, () => {
    return HttpResponse.json({ success: true, message: 'Company settings retrieved successfully', data: mockCompanySettings });
  }),

  http.put(`${API_BASE_URL}/company-settings`, async ({ request }) => {
    const body = await request.json() as Record<string, unknown>;
    Object.assign(mockCompanySettings, body);
    mockCompanySettings.updated_at = new Date().toISOString();
    return HttpResponse.json({ success: true, message: 'Company settings updated successfully', data: mockCompanySettings });
  }),

  // ============== Stock Opname ==============
  http.get(`${API_BASE_URL}/stock-opnames`, ({ request }) => {
    const url = new URL(request.url);
    const page = parseInt(url.searchParams.get('page') || '1');
    const perPage = parseInt(url.searchParams.get('per_page') || '10');
    const status = url.searchParams.get('status');
    const branchID = url.searchParams.get('branch_id');

    let filtered = [...mockStockOpnames].sort(
      (a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime(),
    );
    if (status) filtered = filtered.filter(o => o.status === status);
    if (branchID) filtered = filtered.filter(o => o.branch_id === parseInt(branchID));

    const total = filtered.length;
    const totalPages = Math.ceil(total / perPage) || 1;
    const start = (page - 1) * perPage;
    const paged = filtered.slice(start, start + perPage);

    return HttpResponse.json({
      success: true,
      message: 'Stock opnames retrieved',
      data: paged,
      meta: {
        pagination: {
          total_records: total,
          total_pages: totalPages,
          current_page: page,
          per_page: perPage,
          next_page: page < totalPages ? page + 1 : null,
          prev_page: page > 1 ? page - 1 : null,
        },
      },
    });
  }),

  http.get(`${API_BASE_URL}/stock-opnames/:id`, ({ params }) => {
    const id = parseInt(params.id as string);
    const opname = mockStockOpnames.find(o => o.id === id);
    if (!opname) {
      return HttpResponse.json({ success: false, message: 'Stock opname not found' }, { status: 404 });
    }
    const items = opnameItems[id] || [];
    return HttpResponse.json({
      success: true,
      message: 'Stock opname retrieved successfully',
      data: { ...opname, items, total_items: items.length, counted_items: items.filter(i => i.counted_stock !== null).length },
    });
  }),

  http.post(`${API_BASE_URL}/stock-opnames`, async ({ request }) => {
    const body = (await request.json()) as { branch_id: number; notes?: string; category_id?: number };

    // Build snapshot from current products
    const newId = (mockStockOpnames[mockStockOpnames.length - 1]?.id || 0) + 1;
    const now = new Date().toISOString();
    const variantsSeen = new Set<string>();
    const items: MockStockOpnameItem[] = [];
    for (const p of mockProducts) {
      for (const v of p.variants) {
        if (variantsSeen.has(v.id)) continue;
        variantsSeen.add(v.id);
        items.push({
          id: nextOpnameItemIdValue(),
          stock_opname_id: newId,
          product_variant_id: parseInt(v.id),
          sku: v.sku,
          product_name: p.name,
          variant_name: v.name,
          system_stock: v.stock,
          counted_stock: null,
          variance_qty: 0,
          unit_cost: v.cost,
          variance_value: 0,
          notes: null,
          counted_at: null,
        });
      }
    }
    opnameItems[newId] = items;

    const ymd = now.slice(0, 10).replace(/-/g, '');
    const opname: MockStockOpname = {
      id: newId,
      branch_id: body.branch_id,
      opname_number: `OPN-${ymd}-${String(newId).padStart(4, '0')}`,
      status: 'in_progress',
      notes: body.notes ?? null,
      total_items: items.length,
      counted_items: 0,
      total_variance_qty: 0,
      total_variance_value: 0,
      started_by: 1,
      completed_by: null,
      started_at: now,
      completed_at: null,
      cancelled_at: null,
      created_at: now,
      updated_at: now,
    };
    mockStockOpnames.push(opname);
    return HttpResponse.json(
      { success: true, message: 'Stock opname created successfully', data: { ...opname, items } },
      { status: 201 },
    );
  }),

  http.patch(`${API_BASE_URL}/stock-opnames/:id/items/:itemId`, async ({ params, request }) => {
    const opnameId = parseInt(params.id as string);
    const itemId = parseInt(params.itemId as string);
    const body = (await request.json()) as { counted_stock: number; notes?: string };

    const opname = mockStockOpnames.find(o => o.id === opnameId);
    if (!opname || opname.status !== 'in_progress') {
      return HttpResponse.json({ success: false, message: 'Opname not editable' }, { status: 400 });
    }
    const item = (opnameItems[opnameId] || []).find(i => i.id === itemId);
    if (!item) {
      return HttpResponse.json({ success: false, message: 'Item not found' }, { status: 404 });
    }
    item.counted_stock = body.counted_stock;
    item.variance_qty = body.counted_stock - item.system_stock;
    item.variance_value = item.variance_qty * item.unit_cost;
    item.notes = body.notes ?? null;
    item.counted_at = new Date().toISOString();
    return HttpResponse.json({ success: true, message: 'Item updated', data: item });
  }),

  http.patch(`${API_BASE_URL}/stock-opnames/:id/items`, async ({ params, request }) => {
    const opnameId = parseInt(params.id as string);
    const body = (await request.json()) as { items: { item_id: number; counted_stock: number; notes?: string }[] };
    const opname = mockStockOpnames.find(o => o.id === opnameId);
    if (!opname || opname.status !== 'in_progress') {
      return HttpResponse.json({ success: false, message: 'Opname not editable' }, { status: 400 });
    }
    const items = opnameItems[opnameId] || [];
    const now = new Date().toISOString();
    for (const line of body.items) {
      const it = items.find(i => i.id === line.item_id);
      if (!it) continue;
      it.counted_stock = line.counted_stock;
      it.variance_qty = line.counted_stock - it.system_stock;
      it.variance_value = it.variance_qty * it.unit_cost;
      it.notes = line.notes ?? null;
      it.counted_at = now;
    }
    return HttpResponse.json({
      success: true,
      message: 'Items updated',
      data: { ...opname, items, counted_items: items.filter(i => i.counted_stock !== null).length },
    });
  }),

  http.post(`${API_BASE_URL}/stock-opnames/:id/complete`, async ({ params, request }) => {
    const id = parseInt(params.id as string);
    const body = (await request.json().catch(() => ({}))) as { notes?: string };
    const opname = mockStockOpnames.find(o => o.id === id);
    if (!opname || opname.status !== 'in_progress') {
      return HttpResponse.json({ success: false, message: 'Opname not in progress' }, { status: 400 });
    }
    const items = opnameItems[id] || [];

    // Apply adjustments to product stock for items with a count
    let totalQty = 0;
    let totalValue = 0;
    for (const item of items) {
      if (item.counted_stock === null) continue;
      const product = mockProducts.find(p => p.variants.some(v => parseInt(v.id) === item.product_variant_id));
      const variant = product?.variants.find(v => parseInt(v.id) === item.product_variant_id);
      if (variant) {
        const before = variant.stock;
        variant.stock = item.counted_stock;
        if (variant.stock !== before) {
          mockStockMovements.unshift({
            id: crypto.randomUUID(),
            product_id: product!.id,
            variant_id: variant.id,
            product_name: product!.name,
            variant_name: variant.name,
            sku: variant.sku,
            type: 'ADJUST',
            quantity: Math.abs(variant.stock - before),
            stock_before: before,
            stock_after: variant.stock,
            reference: opname.opname_number,
            notes: 'Stock opname',
            created_at: new Date().toISOString(),
            created_by: 'Admin User',
          });
        }
      }
      totalQty += item.variance_qty;
      totalValue += item.variance_value;
    }
    opname.status = 'completed';
    opname.completed_at = new Date().toISOString();
    opname.completed_by = 1;
    opname.total_variance_qty = totalQty;
    opname.total_variance_value = totalValue;
    if (body?.notes) opname.notes = body.notes;
    return HttpResponse.json({
      success: true,
      message: 'Stock opname completed',
      data: { ...opname, items },
    });
  }),

  http.post(`${API_BASE_URL}/stock-opnames/:id/cancel`, ({ params }) => {
    const id = parseInt(params.id as string);
    const opname = mockStockOpnames.find(o => o.id === id);
    if (!opname || opname.status !== 'in_progress') {
      return HttpResponse.json({ success: false, message: 'Opname not in progress' }, { status: 400 });
    }
    opname.status = 'cancelled';
    opname.cancelled_at = new Date().toISOString();
    return HttpResponse.json({
      success: true,
      message: 'Stock opname cancelled',
      data: { ...opname, items: opnameItems[id] || [] },
    });
  }),
];

// Import types from data
import type { Product, ProductVariant, Order, User, Branch, Promotion, MockStockMovement } from './data';

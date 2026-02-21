import { http, HttpResponse } from 'msw';
import {
  mockUsers,
  mockProducts,
  mockOrders,
  mockCustomers,
  mockBranches,
  mockPromotions,
  mockDashboardStats,
  mockSalesTrend,
} from './data';

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

  http.post(`${API_BASE_URL}/orders`, async ({ request }) => {
    const body = await request.json() as any;
    const newOrder = {
      ...body,
      id: crypto.randomUUID(),
      orderNumber: `ORD-${Math.floor(Math.random() * 100000)}`,
      customerId: `CUST-${Math.floor(Math.random() * 1000)}`,
      subtotal: body.items.reduce((sum: number, item: any) => sum + item.price * item.quantity, 0),
      tax: body.items.reduce((sum: number, item: any) => sum + item.price * item.quantity, 0) * 0.1,
      total: body.items.reduce((sum: number, item: any) => sum + item.price * item.quantity, 0) * 1.1,
      items: body.items.map((item: any) => ({
        ...item,
        id: crypto.randomUUID(),
      })),
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
    };

    mockOrders.unshift(newOrder as Order);

    return HttpResponse.json(newOrder, { status: 201 });
  }),

  http.put(`${API_BASE_URL}/orders/:id`, async ({ params, request }) => {
    const body = await request.json() as any;
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
      subtotal: body.items.reduce((sum: number, item: any) => sum + item.price * item.quantity, 0),
      tax: body.items.reduce((sum: number, item: any) => sum + item.price * item.quantity, 0) * 0.1,
      total: body.items.reduce((sum: number, item: any) => sum + item.price * item.quantity, 0) * 1.1,
      items: body.items.map((item: any, index: number) => ({
        ...item,
        id: mockOrders[orderIndex].items[index]?.id || crypto.randomUUID(),
      })),
      updatedAt: new Date().toISOString(),
    };

    mockOrders[orderIndex] = updatedOrder as Order;

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
];

// Import types from data
import type { Product, Order, User, Branch, Promotion } from './data';

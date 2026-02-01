# API Documentation - Aplikabizz Dashboard

## Overview

RESTful API untuk Aplikabizz Dashboard. API ini menggunakan JWT-based authentication dan mengikuti standar format JSON response "Best of 2026" untuk Go/Echo Framework.

## Base URL

```
Development: http://localhost:8080/api
Production: https://api.aplikabizz.com/api
```

## Response Format & Conventions

### 1. Naming Convention
API menggunakan **snake_case** untuk semua request parameters dan response fields.

### 2. Standard Response Structure

**Success Response (List):**
```json
{
  "success": true,
  "message": "Data retrieved successfully",
  "data": [ ... ],
  "meta": {
    "pagination": {
      "total_records": 100,
      "total_pages": 10,
      "current_page": 1,
      "per_page": 10,
      "count": 10,
      "next_page": 2,
      "prev_page": null
    },
    "summary": { ... }, // Optional
    "server_time": "2026-01-25T13:15:00Z"
  }
}
```

**Success Response (Single):**
```json
{
  "success": true,
  "message": "Data retrieved successfully",
  "data": { ... },
  "meta": {
    "server_time": "2026-01-25T13:15:00Z"
  }
}
```

**Error Response (Validation - 422):**
```json
{
  "success": false,
  "message": "Validation failed",
  "error_code": "VALIDATION_ERROR",
  "errors": [
    {
      "field": "price",
      "message": "Price must be positive"
    }
  ],
  "meta": {
    "server_time": "2026-01-25T13:17:00Z"
  }
}
```

**Error Response (Business Logic - 400):**
```json
{
  "success": false,
  "message": "Transaction failed: Insufficient stock",
  "error_code": "INSUFFICIENT_STOCK",
  "data": {
    "product_id": "PRD-001",
    "available": 0
  },
  "meta": {
    "server_time": "2026-01-25T13:18:00Z"
  }
}
```

---

## API Endpoints

## 1. Authentication

### POST /auth/login

**Request Body:**
```json
{
  "email": "admin@aplikabizz.com",
  "password": "password123"
}
```

**Response (200):**
```json
{
  "success": true,
  "message": "Login successful",
  "data": {
    "user": {
      "id": "uuid",
      "name": "Admin User",
      "email": "admin@aplikabizz.com",
      "role": "admin",
      "avatar": "https://...",
      "branch_id": "uuid"
    },
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  },
  "meta": {
    "server_time": "2026-01-25T13:00:00Z"
  }
}
```

### POST /auth/logout
**Response (200):**
```json
{
  "success": true,
  "message": "Logged out successfully",
  "data": null,
  "meta": {
    "server_time": "2026-01-25T13:00:00Z"
  }
}
```

---

## 2. Dashboard

### GET /dashboard/stats
**Response (200):**
```json
{
  "success": true,
  "message": "Dashboard stats retrieved",
  "data": {
    "total_sales": 250000000,
    "sales_growth": 12.5,
    "new_orders": 125,
    "orders_growth": 8.3,
    "new_customers": 45,
    "net_income": 75000000
  },
  "meta": {
    "server_time": "2026-01-25T13:00:00Z"
  }
}
```

### GET /dashboard/recent-orders
**Response (200):**
```json
{
  "success": true,
  "message": "Recent orders retrieved",
  "data": [
    {
      "id": "uuid",
      "order_number": "#ORD-0123",
      "customer_name": "John Doe",
      "total": 1500000,
      "status": "pending",
      "created_at": "2024-01-15T10:30:00Z"
    }
  ],
  "meta": {
    "server_time": "2026-01-25T13:00:00Z"
  }
}
```

### GET /dashboard/sales-trend
**Response (200):**
```json
{
  "success": true,
  "message": "Sales trend data retrieved",
  "data": [
    { "date": "2024-01-01", "sales": 5000000, "orders": 25 }
  ],
  "meta": {
    "server_time": "2026-01-25T13:00:00Z"
  }
}
```

---

## 3. Products

### GET /products

**Query Parameters:**
- `page`: int (default: 1)
- `limit`: int (default: 10)
- `search`: string
- `category`: string
- `status`: string
- `sort_by`: string
- `order`: `asc` | `desc`

**Response (200):**
```json
{
  "success": true,
  "message": "Product list retrieved",
  "data": [
    {
      "id": "uuid",
      "sku": "SKU-ABC123",
      "name": "Laptop Gaming ROG",
      "price": 15000000,
      "stock": 25,
      "status": "active"
    }
  ],
  "meta": {
    "pagination": {
      "total_records": 50,
      "total_pages": 5,
      "current_page": 1,
      "per_page": 10,
      "count": 10,
      "next_page": 2,
      "prev_page": null
    },
    "summary": {
      "total_active_products": 45,
      "out_of_stock_products": 5
    },
    "server_time": "2026-01-25T13:00:00Z"
  }
}
```

### GET /products/:id
**Response (200):**
```json
{
  "success": true,
  "message": "Product detail retrieved",
  "data": {
    "id": "uuid",
    "sku": "SKU-ABC123",
    "name": "Laptop Gaming ROG",
    "category": "Elektronik",
    "price": 15000000,
    "stock": 25,
    "status": "active",
    "variants": []
  },
  "meta": {
    "server_time": "2026-01-25T13:00:00Z"
  }
}
```

### POST /products
**Response (201):**
```json
{
  "success": true,
  "message": "Product created successfully",
  "data": {
    "id": "uuid",
    "sku": "SKU-XYZ789",
    "name": "Mouse Gaming"
  },
  "meta": {
    "server_time": "2026-01-25T13:00:00Z"
  }
}
```

**Error Response (422 - Validation):**
```json
{
  "success": false,
  "message": "Validation failed",
  "error_code": "VALIDATION_ERROR",
  "errors": [
    { "field": "sku", "message": "SKU already exists" },
    { "field": "price", "message": "Price must be greater than 0" }
  ],
  "meta": {
    "server_time": "2026-01-25T13:00:00Z"
  }
}
```

---

## 4. Orders

### GET /orders
**Response (200):**
```json
{
  "success": true,
  "message": "Order list retrieved",
  "data": [
    {
      "id": "uuid",
      "order_number": "#ORD-0123",
      "customer_name": "John Doe",
      "total": 16550000,
      "status": "pending"
    }
  ],
  "meta": {
    "pagination": {
      "total_records": 100,
      "total_pages": 10,
      "current_page": 1,
      "per_page": 10,
      "count": 10,
      "next_page": 2,
      "prev_page": null
    },
    "summary": {
      "pending_orders": 5,
      "todays_revenue": 50000000
    },
    "server_time": "2026-01-25T13:00:00Z"
  }
}
```

### POST /orders
**Response (201):**
```json
{
  "success": true,
  "message": "Order created successfully",
  "data": {
    "id": "uuid",
    "order_number": "#ORD-0124",
    "total": 1100000,
    "status": "pending"
  },
  "meta": {
    "server_time": "2026-01-25T13:00:00Z"
  }
}
```

**Error Response (400 - Stock Empty):**
```json
{
  "success": false,
  "message": "Order failed: Insufficient stock",
  "error_code": "INSUFFICIENT_STOCK",
  "data": {
    "product_id": "PRD-002",
    "requested_qty": 5,
    "available_stock": 2
  },
  "meta": {
    "server_time": "2026-01-25T13:00:00Z"
  }
}
```

---

## 5. Customers

### GET /customers
**Response (200):**
```json
{
  "success": true,
  "message": "Customer list retrieved",
  "data": [
    {
      "id": "uuid",
      "customer_id": "#CUST-001",
      "name": "John Doe",
      "status": "vip"
    }
  ],
  "meta": {
    "pagination": {
      "total_records": 200,
      "total_pages": 20,
      "current_page": 1,
      "per_page": 10,
      "count": 10,
      "next_page": 2,
      "prev_page": null
    },
    "server_time": "2026-01-25T13:00:00Z"
  }
}
```

---

## 6. Branches

### GET /branches
**Response (200):**
```json
{
  "success": true,
  "message": "Branch list retrieved",
  "data": [
    {
      "id": "uuid",
      "code": "JKT",
      "name": "Cabang Jakarta Pusat",
      "status": "active"
    }
  ],
  "meta": {
    "server_time": "2026-01-25T13:00:00Z"
  }
}
```

---

## 7. Reports

### GET /reports/sales
**Response (200):**
```json
{
  "success": true,
  "message": "Sales report generated",
  "data": {
    "summary": {
      "total_sales": 150000000,
      "total_orders": 320
    },
    "breakdown": [...]
  },
  "meta": {
    "server_time": "2026-01-25T13:00:00Z"
  }
}
```

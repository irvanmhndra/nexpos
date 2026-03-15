# API Documentation — Nexpos

## Overview

RESTful API for Nexpos — a multi-tenant Point of Sale system. All endpoints use JSON. Protected routes require JWT authentication.

## Base URL

```
Development: http://localhost:8080/api/v1
Production:  https://api.nexpos.irvanmahendra.com/api/v1
```

---

## Response Format

All responses follow this structure:

**Success (single item):**
```json
{
  "success": true,
  "message": "Data retrieved successfully",
  "data": { ... }
}
```

**Success (list with pagination):**
```json
{
  "success": true,
  "message": "Data retrieved successfully",
  "data": [ ... ],
  "meta": {
    "page": 1,
    "limit": 10,
    "total": 100,
    "total_pages": 10
  }
}
```

**Error (validation — 422):**
```json
{
  "success": false,
  "message": "Validation failed",
  "errors": {
    "name": "name is required"
  }
}
```

**Error (business logic — 400/404/409):**
```json
{
  "success": false,
  "message": "Insufficient stock for variant SKU-001"
}
```

---

## Authentication

All protected routes require the header:
```
Authorization: Bearer <access_token>
```

---

## 1. Auth

### POST /auth/login

**Request:**
```json
{
  "email": "admin@example.com",
  "password": "password123"
}
```

**Response (200):**
```json
{
  "success": true,
  "message": "Login successful",
  "data": {
    "access_token": "eyJhbGci...",
    "refresh_token": "eyJhbGci...",
    "user": {
      "id": 1,
      "name": "Admin",
      "email": "admin@example.com",
      "role": "admin",
      "company_id": 1,
      "branch_id": 1
    }
  }
}
```

### POST /auth/register

**Request:**
```json
{
  "name": "John Doe",
  "email": "john@example.com",
  "password": "password123",
  "company_name": "My Store"
}
```

**Response (201):** Same structure as login.

### POST /auth/refresh

**Request:**
```json
{ "refresh_token": "eyJhbGci..." }
```

**Response (200):**
```json
{
  "success": true,
  "message": "Token refreshed",
  "data": { "access_token": "eyJhbGci..." }
}
```

### POST /auth/logout

**Response (200):**
```json
{ "success": true, "message": "Logged out successfully", "data": null }
```

---

## 2. Users

### GET /users

**Query parameters:** `page`, `limit`, `search`, `role`, `status`

**Response (200):**
```json
{
  "success": true,
  "message": "Users retrieved",
  "data": [
    {
      "id": 1,
      "name": "John Doe",
      "email": "john@example.com",
      "role": "cashier",
      "status": "active",
      "branch_id": 1,
      "created_at": "2026-01-01T00:00:00Z"
    }
  ],
  "meta": { "page": 1, "limit": 10, "total": 5, "total_pages": 1 }
}
```

### POST /users

**Request:**
```json
{
  "name": "Jane Doe",
  "email": "jane@example.com",
  "password": "password123",
  "role": "cashier",
  "branch_id": 1
}
```

### GET /users/:id / PUT /users/:id / DELETE /users/:id

Standard single-item CRUD. PUT accepts same fields as POST (all optional).

### PATCH /users/:id/status

**Request:**
```json
{ "status": "inactive" }
```

---

## 3. Branches

### GET /branches

**Query parameters:** `page`, `limit`, `search`

**Response (200):**
```json
{
  "success": true,
  "message": "Branches retrieved",
  "data": [
    {
      "id": 1,
      "name": "Main Branch",
      "address": "Jl. Sudirman No. 1",
      "phone": "021-1234567",
      "is_active": true,
      "created_at": "2026-01-01T00:00:00Z"
    }
  ],
  "meta": { "page": 1, "limit": 10, "total": 2, "total_pages": 1 }
}
```

### POST /branches

```json
{
  "name": "Branch 2",
  "address": "Jl. Thamrin No. 5",
  "phone": "021-9876543"
}
```

---

## 4. Product Categories

### GET /product-categories

**Query parameters:** `page`, `limit`, `search`

**Response (200):**
```json
{
  "success": true,
  "message": "Categories retrieved",
  "data": [
    { "id": 1, "name": "Beverages", "created_at": "2026-01-01T00:00:00Z" }
  ],
  "meta": { "page": 1, "limit": 10, "total": 3, "total_pages": 1 }
}
```

### GET /product-categories/all

Returns all categories as a flat array (no pagination). Useful for dropdowns.

### POST /product-categories

```json
{ "name": "Snacks" }
```

---

## 5. Products

### GET /products

**Query parameters:** `page`, `limit`, `search`, `category`, `status`

**Response (200):**
```json
{
  "success": true,
  "message": "Products retrieved",
  "data": [
    {
      "id": 1,
      "name": "Teh Botol",
      "image_data": "data:image/jpeg;base64,...",
      "product_category_id": 1,
      "category_name": "Beverages",
      "is_active": true,
      "variants": [
        {
          "id": 1,
          "name": "330ml",
          "sku": "TEH-330",
          "price": 5000,
          "standard_cost": 3000,
          "is_active": true
        }
      ]
    }
  ],
  "meta": { "page": 1, "limit": 10, "total": 20, "total_pages": 2 }
}
```

### POST /products

```json
{
  "name": "Kopi Susu",
  "product_category_id": 1,
  "image_data": "data:image/jpeg;base64,...",
  "variants": [
    { "name": "Default", "sku": "KPS-001", "price": 15000, "standard_cost": 8000 }
  ]
}
```

**Response (201):** Returns the created product with all variants.

### GET /products/:id

Returns product with full variant list.

### PUT /products/:id

Accepts same body as POST. Can update product fields and its variants.

---

## 6. Customers

### GET /customers

**Query parameters:** `page`, `limit`, `search`

**Response (200):**
```json
{
  "success": true,
  "message": "Customers retrieved",
  "data": [
    {
      "id": 1,
      "name": "Budi Santoso",
      "phone": "08123456789",
      "email": "budi@example.com",
      "address": "Jakarta",
      "created_at": "2026-01-01T00:00:00Z"
    }
  ],
  "meta": { "page": 1, "limit": 10, "total": 50, "total_pages": 5 }
}
```

### POST /customers

```json
{
  "name": "Ani Wijaya",
  "phone": "08987654321",
  "email": "ani@example.com",
  "address": "Bandung"
}
```

---

## 7. Orders

### GET /orders

**Query parameters:** `page`, `limit`, `search`, `status`, `date_from`, `date_to`

**Response (200):**
```json
{
  "success": true,
  "message": "Orders retrieved",
  "data": [
    {
      "id": 1,
      "order_number": "ORD-20260101-0001",
      "status": "completed",
      "subtotal": 50000,
      "promo_discount": 5000,
      "grand_total": 45000,
      "customer_name": "Budi",
      "cashier_name": "Admin",
      "branch_name": "Main Branch",
      "created_at": "2026-01-01T10:00:00Z"
    }
  ],
  "meta": { "page": 1, "limit": 10, "total": 100, "total_pages": 10 }
}
```

### POST /orders

Creates an order and immediately deducts stock.

**Request:**
```json
{
  "branch_id": 1,
  "customer_id": 1,
  "promo_code": "DISC10",
  "items": [
    { "product_variant_id": 1, "quantity": 2, "discount_amount": 0 }
  ]
}
```

**Response (201):** Full order object (see GET /orders/:id shape).

**Error (400 — insufficient stock):**
```json
{
  "success": false,
  "message": "Insufficient stock for variant TEH-330 (requested: 5, available: 2)"
}
```

### POST /orders/preview

Calculates order totals (with promo validation) **without creating the order**. Used by POS before checkout to show accurate pricing.

**Request:** Same as POST /orders.

**Response (200):**
```json
{
  "success": true,
  "message": "Order preview calculated",
  "data": {
    "subtotal": 50000,
    "item_discount": 0,
    "promo_discount": 5000,
    "tax": 0,
    "grand_total": 45000,
    "applied_promotion": {
      "id": 1,
      "code": "DISC10",
      "name": "Diskon 10%",
      "discount_amount": 5000
    }
  }
}
```

If promo code is invalid or inactive, `applied_promotion` is `null` and `promo_discount` is `0`.

### GET /orders/:id

**Response (200):**
```json
{
  "success": true,
  "message": "Order retrieved",
  "data": {
    "id": 1,
    "order_number": "ORD-20260101-0001",
    "status": "completed",
    "subtotal": 50000,
    "item_discount": 0,
    "promo_discount": 5000,
    "tax": 0,
    "grand_total": 45000,
    "promo_code": "DISC10",
    "customer_id": 1,
    "customer_name": "Budi",
    "branch_id": 1,
    "branch_name": "Main Branch",
    "items": [
      {
        "id": 1,
        "product_variant_id": 1,
        "product_name": "Teh Botol",
        "variant_name": "330ml",
        "sku": "TEH-330",
        "quantity": 2,
        "unit_price": 5000,
        "discount_amount": 0,
        "subtotal": 10000
      }
    ],
    "payments": [
      {
        "id": 1,
        "method": "cash",
        "amount": 50000,
        "change_amount": 5000,
        "paid_at": "2026-01-01T10:05:00Z"
      }
    ],
    "created_at": "2026-01-01T10:00:00Z",
    "completed_at": "2026-01-01T10:05:00Z"
  }
}
```

### POST /orders/:id/confirm

No request body. Transitions order from `draft` to `confirmed`.

### POST /orders/:id/payments

```json
{
  "method": "cash",
  "amount": 50000
}
```

`method` values: `cash`, `transfer`, `qris`, `card`, `other`

### POST /orders/:id/complete

No request body. Transitions order to `completed`.

### POST /orders/:id/cancel

```json
{ "reason": "Customer changed mind" }
```

### POST /orders/:id/void

```json
{ "reason": "Incorrect order" }
```

### POST /orders/:id/refund

```json
{ "payment_id": 1, "amount": 5000, "reason": "Returned item" }
```

---

## 8. Promotions

### GET /promotions

**Query parameters:** `page`, `limit`, `search`, `status`

**Response (200):**
```json
{
  "success": true,
  "message": "Promotions retrieved",
  "data": [
    {
      "id": 1,
      "code": "DISC10",
      "name": "Diskon 10%",
      "type": "percentage",
      "value": 10,
      "min_purchase": 20000,
      "max_discount": 50000,
      "start_date": "2026-01-01T00:00:00Z",
      "end_date": "2026-12-31T23:59:59Z",
      "is_active": true,
      "usage_count": 15,
      "usage_limit": 100
    }
  ],
  "meta": { "page": 1, "limit": 10, "total": 3, "total_pages": 1 }
}
```

### POST /promotions

```json
{
  "code": "FLAT5K",
  "name": "Potongan 5000",
  "type": "fixed",
  "value": 5000,
  "min_purchase": 25000,
  "max_discount": null,
  "start_date": "2026-03-01T00:00:00Z",
  "end_date": "2026-03-31T23:59:59Z",
  "usage_limit": 200,
  "is_active": true
}
```

`type` values: `fixed` | `percentage`

---

## 9. Inventory

### GET /inventory

**Query parameters:** `page`, `limit`, `search`, `category`, `status`, `branch_id`

`status` values: `normal` | `low` | `out_of_stock`

**Response (200):**
```json
{
  "success": true,
  "message": "Inventory retrieved",
  "data": [
    {
      "product_variant_id": 1,
      "product_id": 1,
      "product_name": "Teh Botol",
      "variant_name": "330ml",
      "sku": "TEH-330",
      "category_name": "Beverages",
      "image_data": "data:image/jpeg;base64,...",
      "branch_id": 1,
      "branch_name": "Main Branch",
      "current_stock": 50,
      "min_stock": 10,
      "stock_value": 150000,
      "has_variants": false
    }
  ],
  "meta": { "page": 1, "limit": 20, "total": 35, "total_pages": 2 }
}
```

**Stock status logic:**
- `out_of_stock`: `current_stock === 0`
- `low`: `current_stock > 0 && current_stock <= min_stock`
- `normal`: `current_stock > min_stock`

### GET /inventory/stats

**Query parameters:** `branch_id`

**Response (200):**
```json
{
  "success": true,
  "message": "Inventory stats retrieved",
  "data": {
    "total_sku": 35,
    "out_of_stock": 3,
    "low_stock": 5,
    "total_stock_value": 12500000
  }
}
```

### POST /inventory/adjust

Creates a stock movement and updates the variant's stock quantity.

**Request:**
```json
{
  "product_variant_id": 1,
  "branch_id": 1,
  "type": "in",
  "quantity": 50,
  "notes": "Restok dari supplier"
}
```

`type` values: `in` | `out` | `adjustment`

- `in`: adds quantity
- `out`: subtracts quantity
- `adjustment`: sets stock to the given quantity

**Response (200):**
```json
{
  "success": true,
  "message": "Stock adjusted",
  "data": {
    "id": 10,
    "product_variant_id": 1,
    "branch_id": 1,
    "type": "in",
    "quantity": 50,
    "stock_before": 10,
    "stock_after": 60,
    "notes": "Restok dari supplier",
    "created_at": "2026-03-15T09:00:00Z"
  }
}
```

### PUT /inventory/:variantId/min-stock

Updates the minimum stock threshold for a variant at a branch.

**Request:**
```json
{
  "branch_id": 1,
  "min_quantity": 15
}
```

### GET /inventory/movements

**Query parameters:** `page`, `limit`, `search`, `type`, `date_from`, `date_to`, `branch_id`

**Response (200):**
```json
{
  "success": true,
  "message": "Stock movements retrieved",
  "data": [
    {
      "id": 10,
      "product_variant_id": 1,
      "product_name": "Teh Botol",
      "variant_name": "330ml",
      "sku": "TEH-330",
      "branch_id": 1,
      "branch_name": "Main Branch",
      "type": "in",
      "quantity": 50,
      "stock_before": 10,
      "stock_after": 60,
      "reference": null,
      "notes": "Restok dari supplier",
      "created_by": "Admin",
      "created_at": "2026-03-15T09:00:00Z"
    }
  ],
  "meta": { "page": 1, "limit": 20, "total": 120, "total_pages": 6 }
}
```

### GET /inventory/movements/stats

**Query parameters:** `branch_id`, `date_from`, `date_to`

**Response (200):**
```json
{
  "success": true,
  "message": "Movement stats retrieved",
  "data": {
    "total_movements": 120,
    "total_in": 800,
    "total_out": 350,
    "total_adjustment": 25,
    "net_change": 450
  }
}
```

---

## 10. Reports

All report endpoints support `branch_id`, `date_from`, `date_to` query parameters.

### GET /reports/summary

**Response (200):**
```json
{
  "success": true,
  "message": "Summary retrieved",
  "data": {
    "total_revenue": 15000000,
    "total_orders": 320,
    "average_order_value": 46875,
    "total_items_sold": 850,
    "revenue_growth": 12.5,
    "orders_growth": 8.3
  }
}
```

### GET /reports/sales-trend

**Response (200):**
```json
{
  "success": true,
  "message": "Sales trend retrieved",
  "data": [
    { "date": "2026-03-01", "revenue": 500000, "orders": 12 },
    { "date": "2026-03-02", "revenue": 750000, "orders": 18 }
  ]
}
```

### GET /reports/top-products

**Response (200):**
```json
{
  "success": true,
  "message": "Top products retrieved",
  "data": [
    {
      "product_id": 1,
      "product_name": "Teh Botol",
      "variant_name": "330ml",
      "sku": "TEH-330",
      "quantity_sold": 200,
      "revenue": 1000000
    }
  ]
}
```

### GET /reports/category-revenue

**Response (200):**
```json
{
  "success": true,
  "message": "Category revenue retrieved",
  "data": [
    { "category_name": "Beverages", "revenue": 5000000, "percentage": 33.3 }
  ]
}
```

### GET /reports/payment-methods

**Response (200):**
```json
{
  "success": true,
  "message": "Payment methods retrieved",
  "data": [
    { "method": "cash", "count": 150, "amount": 7500000, "percentage": 50.0 }
  ]
}
```

### GET /reports/hourly-sales

**Response (200):**
```json
{
  "success": true,
  "message": "Hourly sales retrieved",
  "data": [
    { "hour": 9, "orders": 15, "revenue": 750000 },
    { "hour": 10, "orders": 25, "revenue": 1250000 }
  ]
}
```

---

## 11. Company Settings

### GET /company-settings

**Response (200):**
```json
{
  "success": true,
  "message": "Company settings retrieved",
  "data": {
    "company_name": "Toko Saya",
    "address": "Jl. Sudirman No. 1, Jakarta",
    "phone": "021-1234567",
    "email": "toko@example.com",
    "tax_rate": 0,
    "currency": "IDR",
    "logo_data": "data:image/jpeg;base64,...",
    "receipt_footer": "Terima kasih telah berbelanja!"
  }
}
```

### PUT /company-settings

Accepts any subset of the fields above. Only provided fields are updated.

**Request:**
```json
{
  "company_name": "Toko Maju Jaya",
  "receipt_footer": "Selamat datang kembali!"
}
```

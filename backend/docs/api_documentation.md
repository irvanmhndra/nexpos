# API Documentation — Nexpos

## Overview

RESTful API for **Nexpos** — a multi-tenant POS SaaS built with Go + Echo v5 + PostgreSQL.
The API uses JWT-based authentication and follows the consistent response format below.

## Base URL

```
Development: http://localhost:8080/api/v1
Production:  https://api.nexpos.irvanmahendra.com/api/v1
```

Health endpoints (outside the `/api/v1` group):

```
GET /health
GET /health/live
GET /health/ready
```

## Conventions

- **Naming**: all request/response fields use `snake_case`.
- **IDs**: integer (`int64`). Orders use the string `order_no` as the user-facing identifier.
- **Timestamps**: RFC 3339 (`2026-05-14T08:00:00Z`).
- **Money**: decimal numbers (rupiah), not strings.
- **Multi-tenant**: scoped per company; users are limited to their own company via JWT claims.

## Authentication

Most endpoints are in the `protected` group and require `Authorization: Bearer <access_token>`. Public endpoints are only `/auth/login`, `/auth/register`, `/auth/refresh`, `/auth/logout`, and the health checks.

## Standard Response Format

**Success — single resource:**
```json
{
  "success": true,
  "message": "Customer retrieved successfully",
  "data": { ... }
}
```

**Success — paginated list:**
```json
{
  "success": true,
  "message": "Customers retrieved",
  "data": [ ... ],
  "meta": {
    "pagination": {
      "total_records": 200,
      "total_pages": 10,
      "current_page": 1,
      "per_page": 20,
      "next_page": 2,
      "prev_page": null
    }
  }
}
```

**Error:**
```json
{
  "success": false,
  "message": "Branch not found",
  "error_code": "NOT_FOUND"
}
```

**Validation error (422):**
```json
{
  "success": false,
  "message": "Validation failed",
  "errors": {
    "email": "email is required",
    "password": "password must be at least 8 characters"
  }
}
```

---

# Endpoints

## 1. Authentication

### POST `/auth/login`
Body:
```json
{ "email": "admin@example.com", "password": "secret123" }
```
Response (200):
```json
{
  "success": true,
  "message": "Login successful",
  "data": {
    "user": {
      "id": 1,
      "name": "Admin User",
      "email": "admin@example.com",
      "role": "owner",
      "company": { "id": 1, "code": "ACME", "name": "Acme Retail" },
      "default_branch": { "id": 1, "code": "JKT-001", "name": "Cabang Pusat Jakarta" }
    },
    "access_token": "<jwt>",
    "refresh_token": "<jwt>",
    "expires_in": 3600
  }
}
```

### POST `/auth/register`
Body:
```json
{ "name": "Owner Baru", "email": "owner@toko.com", "password": "secret123" }
```
Creates a new user + company. Response mirrors login.

### POST `/auth/refresh`
Body: `{ "refresh_token": "<jwt>" }`
Response: `{ access_token, refresh_token, expires_in }`.

### POST `/auth/logout`
Header: `Authorization: Bearer <access_token>`. Revokes the session.

---

## 2. Users

All endpoints require auth.

| Method | Path | Description |
| --- | --- | --- |
| `POST` | `/users` | Create user |
| `GET` | `/users` | List users (paginated, query: `page`, `limit`) |
| `GET` | `/users/:id` | Get user |
| `PUT` | `/users/:id` | Update user |
| `PATCH` | `/users/:id/status` | Update status (`active` / `inactive`) |
| `DELETE` | `/users/:id` | Soft delete |

**Create user body:**
```json
{ "email": "kasir@toko.com", "password": "secret123", "name": "Kasir 1", "role_id": 3 }
```

**User response:**
```json
{
  "id": 5,
  "company_id": 1,
  "role_id": 3,
  "email": "kasir@toko.com",
  "name": "Kasir 1",
  "status": "active",
  "role": { "id": 3, "code": "cashier", "name": "Kasir" },
  "created_at": "2026-05-14T08:00:00Z",
  "updated_at": "2026-05-14T08:00:00Z"
}
```

---

## 3. Branches

| Method | Path | Description |
| --- | --- | --- |
| `POST` | `/branches` | Create |
| `GET` | `/branches` | List (query: `page`, `per_page`, `search`, `is_active`) |
| `GET` | `/branches/:id` | Get |
| `PUT` | `/branches/:id` | Update |
| `DELETE` | `/branches/:id` | Delete |

**Create branch body:**
```json
{
  "code": "JKT-001",
  "name": "Cabang Pusat Jakarta",
  "address": "Jl. Sudirman Kav. 52-53",
  "phone": "021-555-0192",
  "is_active": true
}
```

---

## 4. Customers

| Method | Path | Description |
| --- | --- | --- |
| `POST` | `/customers` | Create |
| `GET` | `/customers` | List (query: `page`, `per_page`, `search`, `is_member`) |
| `GET` | `/customers/:id` | Get |
| `PUT` | `/customers/:id` | Update |
| `DELETE` | `/customers/:id` | Delete |

**Create customer body:**
```json
{
  "code": "C-0001",
  "name": "Budi Santoso",
  "phone": "+62 812 0000 0000",
  "email": "budi@example.com",
  "is_member": true
}
```

---

## 5. Product Categories

| Method | Path | Description |
| --- | --- | --- |
| `POST` | `/product-categories` | Create |
| `GET` | `/product-categories` | List (paginated) |
| `GET` | `/product-categories/all` | List all (no pagination, for selectors) |
| `GET` | `/product-categories/:id` | Get |
| `PUT` | `/product-categories/:id` | Update |
| `DELETE` | `/product-categories/:id` | Delete |

**Create body:**
```json
{ "code": "BEV", "name": "Minuman", "parent_id": null, "sort_order": 1, "is_active": true }
```

---

## 6. Products

| Method | Path | Description |
| --- | --- | --- |
| `POST` | `/products` | Create (with variants) |
| `GET` | `/products` | List (query: `page`, `per_page`, `search`, `category_id`, `is_active`) |
| `GET` | `/products/:id` | Get |
| `PUT` | `/products/:id` | Update (variants merged by id) |
| `DELETE` | `/products/:id` | Delete |

**Create product body:**
```json
{
  "name": "Kopi Susu Gula Aren",
  "category_id": 2,
  "description": "Signature kami",
  "image_data": "data:image/png;base64,...",
  "is_active": true,
  "variants": [
    {
      "sku": "KS-REG",
      "name": "Regular",
      "attributes": { "size": "regular" },
      "price": 22000,
      "standard_cost": 9000,
      "last_purchase_cost": 9500,
      "is_default": true,
      "is_active": true
    },
    {
      "sku": "KS-LRG",
      "name": "Large",
      "price": 28000,
      "standard_cost": 11000,
      "last_purchase_cost": 11500,
      "is_default": false
    }
  ]
}
```

**Sale price** (harga coret) bisa diset per-variant lewat field `sale_price`, `sale_start`, `sale_end`. Saat aktif, kasir akan menampilkan harga coret.

---

## 7. Orders

Order lifecycle: `draft → confirmed → completed/cancelled/voided`, with a separate `payment_status` (`unpaid → partial → paid`).

| Method | Path | Description |
| --- | --- | --- |
| `POST` | `/orders/preview` | Compute total + promo discount (does not persist) |
| `POST` | `/orders` | Create order |
| `GET` | `/orders` | List |
| `GET` | `/orders/:id` | Get |
| `PUT` | `/orders/:id` | Update items / customer |
| `POST` | `/orders/:id/confirm` | Confirm (locks order for payment) |
| `POST` | `/orders/:id/payments` | Add payment |
| `POST` | `/orders/:id/complete` | Complete the order (auto-deducts stock) |
| `POST` | `/orders/:id/cancel` | Cancel (before payment) |
| `POST` | `/orders/:id/void` | Void (after completed, for corrections) |
| `POST` | `/orders/:id/refund` | Refund a payment |
| `GET` | `/orders/:id/receipt` | Get receipt (requires the MongoDB receipt store) |

**List query:** `page`, `per_page`, `search`, `status`, `payment_status`, `fulfillment_type`, `fulfillment_status`, `customer_id`, `payment_method`, `date_from`, `date_to`.

**Create order body:**
```json
{
  "customer_id": 12,
  "fulfillment_type": "counter",
  "items": [
    { "product_variant_id": 7, "quantity": 2, "discount_amount": 0 },
    { "product_variant_id": 11, "quantity": 1 }
  ],
  "payments": [{ "method": "cash", "amount": 60000 }],
  "promo_code": "WELCOME10",
  "notes": "Take-away",
  "offline_id": "8b3e1f8b-...-uuid"
}
```

**Payment methods:** `cash`, `debit_card`, `credit_card`, `e_wallet`, `bank_transfer`, `qris`.

**Order response (excerpt):**
```json
{
  "id": 101,
  "order_no": "ORD-20260514-0123",
  "status": "completed",
  "payment_status": "paid",
  "fulfillment_type": "counter",
  "total_amount": 72000,
  "total_discount": 7000,
  "total_tax": 0,
  "grand_total": 65000,
  "paid_amount": 65000,
  "balance_due": 0,
  "applied_promotion": { "id": 1, "code": "WELCOME10", "name": "Welcome 10%", "discount_amount": 7000 },
  "items": [ { "id": 1, "product_variant_id": 7, "sku": "KS-REG", "product_name": "Kopi Susu", "variant_name": "Regular", "unit_price": 22000, "quantity": 2, "subtotal": 44000 } ],
  "payments": [ { "id": 1, "method": "cash", "amount": 65000, "status": "completed", "paid_at": "2026-05-14T10:15:00Z" } ],
  "created_at": "2026-05-14T10:14:30Z",
  "completed_at": "2026-05-14T10:15:00Z"
}
```

**Refund body:**
```json
{ "payment_id": 1, "amount": 22000, "refund_reason": "Salah varian" }
```

---

## 8. Reports

| Method | Path | Description |
| --- | --- | --- |
| `GET` | `/reports/summary` | Ringkasan periode |
| `GET` | `/reports/sales-trend` | Trend per hari |
| `GET` | `/reports/top-products` | Top produk (query: `limit`) |
| `GET` | `/reports/category-revenue` | Revenue per kategori |
| `GET` | `/reports/payment-methods` | Breakdown metode bayar |
| `GET` | `/reports/hourly-sales` | Penjualan per jam |

All endpoints require `date_from` and `date_to` query params (`YYYY-MM-DD`).

**Summary response:**
```json
{
  "total_revenue": 15000000,
  "total_orders": 320,
  "total_items_sold": 980,
  "avg_order_value": 46875,
  "new_customers": 18,
  "completed_orders": 305,
  "cancelled_orders": 8,
  "pending_orders": 7,
  "total_discount": 750000,
  "total_tax": 0,
  "gross_profit": 5400000,
  "gross_profit_margin": 36.0
}
```

---

## 9. Promotions

| Method | Path | Description |
| --- | --- | --- |
| `POST` | `/promotions` | Create |
| `GET` | `/promotions` | List (query: `page`, `per_page`, `search`, `is_active`, `type`) |
| `GET` | `/promotions/:id` | Get |
| `PUT` | `/promotions/:id` | Update |
| `DELETE` | `/promotions/:id` | Delete |

**Create promotion body:**
```json
{
  "code": "WELCOME10",
  "name": "Welcome 10%",
  "type": "discount",
  "discount_type": "percentage",
  "discount_value": 10,
  "min_purchase": 50000,
  "max_discount": 25000,
  "start_at": "2026-05-01T00:00:00Z",
  "end_at": "2026-05-31T23:59:59Z",
  "priority": 10,
  "is_active": true
}
```

`type`: `discount` | `bundle` | `conditional`. Active promotions are applied automatically on preview/create when `promo_code` matches.

---

## 10. Inventory & Stock Movements

| Method | Path | Description |
| --- | --- | --- |
| `GET` | `/inventory` | List stok (query: `page`, `per_page`, `search`, `category`, `status`, `branch_id`) |
| `GET` | `/inventory/stats` | Stats (total SKU, low/out, total value) |
| `POST` | `/inventory/adjust` | Adjust manual (IN/OUT/ADJUST) |
| `PUT` | `/inventory/:variantId/min-stock` | Update min stock |
| `GET` | `/inventory/movements` | List stock movements |
| `GET` | `/inventory/movements/stats` | Movement stats (this month) |

`status`: `normal` | `low` | `out_of_stock`.

**Adjust body:**
```json
{
  "variant_id": 7,
  "branch_id": 1,
  "type": "ADJUST",
  "quantity": 50,
  "unit_cost": 9000,
  "note": "Koreksi opname manual"
}
```

`type` semantics:
- `IN`: increase stock by `quantity`
- `OUT`: decrease stock by `quantity` (fails on insufficient stock)
- `ADJUST`: **set** stock to `quantity` (not a delta)

---

## 11. Company Settings

| Method | Path | Description |
| --- | --- | --- |
| `GET` | `/company-settings` | Get settings |
| `PUT` | `/company-settings` | Update (partial; null fields ignored) |

**Response:**
```json
{
  "id": 1,
  "company_id": 1,
  "tax_enabled": false,
  "tax_rate": 11,
  "tax_inclusive": true,
  "rounding_enabled": false,
  "rounding_amount": 100,
  "auto_complete_counter_orders": true,
  "require_customer_for_delivery": true,
  "receipt_header": null,
  "receipt_footer": null,
  "show_tax_on_receipt": true,
  "offline_mode_enabled": false,
  "max_offline_days": 7
}
```

---

## 12. Suppliers

| Method | Path | Description |
| --- | --- | --- |
| `POST` | `/suppliers` | Create |
| `GET` | `/suppliers` | List (query: `page`, `per_page`, `search`, `is_active`) |
| `GET` | `/suppliers/:id` | Get |
| `PUT` | `/suppliers/:id` | Update |
| `DELETE` | `/suppliers/:id` | Delete |

**Create body:**
```json
{
  "code": "SUP-001",
  "name": "PT Sumber Sejahtera",
  "contact_name": "Pak Andi",
  "phone": "021-1234-5678",
  "email": "sales@sumber.co.id",
  "address": "Jl. Industri No. 12",
  "notes": null,
  "is_active": true
}
```

---

## 13. Purchase Orders

| Method | Path | Description |
| --- | --- | --- |
| `POST` | `/purchase-orders` | Create draft PO |
| `GET` | `/purchase-orders` | List (query: `page`, `per_page`, `status`, `supplier_id`) |
| `GET` | `/purchase-orders/:id` | Get with items |
| `POST` | `/purchase-orders/:id/receive` | Receive goods (creates an IN stock movement) |

Status: `draft` | `ordered` | `partial` | `received` | `cancelled`.

**Create body:**
```json
{
  "branch_id": 1,
  "supplier_id": 3,
  "notes": "Order rutin mingguan",
  "items": [
    { "product_variant_id": 7, "sku": "KS-REG", "variant_name": "Kopi Susu Regular", "quantity": 100, "unit_cost": 9000 },
    { "product_variant_id": 8, "sku": "KS-LRG", "variant_name": "Kopi Susu Large", "quantity": 50, "unit_cost": 11000 }
  ]
}
```

**Receive body:**
```json
{
  "items": [
    { "item_id": 1, "received_quantity": 100 },
    { "item_id": 2, "received_quantity": 30 }
  ]
}
```
Status auto-transitions: `ordered → partial → received` based on total received qty.

---

## 14. Shifts

Cash drawer reconciliation per kasir per branch.

| Method | Path | Description |
| --- | --- | --- |
| `POST` | `/shifts` | Open shift |
| `GET` | `/shifts` | List (query: `page`, `per_page`, `branch_id`, `cashier_id`, `status`, `date_from`, `date_to`) |
| `GET` | `/shifts/current` | Get my open shift |
| `GET` | `/shifts/:id` | Get shift detail |
| `POST` | `/shifts/:id/close` | Close shift (computes cash difference) |

**Open body:**
```json
{ "branch_id": 1, "opening_float": 500000 }
```

**Close body:**
```json
{ "actual_cash": 1825000, "notes": "Kelebihan 5rb di laci" }
```

On close, the system computes `expected_cash = opening_float + total_cash_payments_during_shift` and `cash_difference = actual_cash - expected_cash`.

---

## 15. Expenses

### Categories
| Method | Path | Description |
| --- | --- | --- |
| `POST` | `/expense-categories` | Create |
| `GET` | `/expense-categories` | List |
| `PUT` | `/expense-categories/:id` | Update |
| `DELETE` | `/expense-categories/:id` | Delete |

### Expenses
| Method | Path | Description |
| --- | --- | --- |
| `POST` | `/expenses` | Create |
| `GET` | `/expenses` | List (query: `page`, `per_page`, `branch_id`, `category_id`, `date_from`, `date_to`) |
| `GET` | `/expenses/summary` | Summary by category (query: `branch_id`, `date_from`, `date_to`) |
| `GET` | `/expenses/:id` | Get |
| `PUT` | `/expenses/:id` | Update |
| `DELETE` | `/expenses/:id` | Delete |

**Create body:**
```json
{
  "branch_id": 1,
  "category_id": 4,
  "amount": 150000,
  "description": "Beli kantong plastik",
  "reference_no": "NOTA-2304",
  "expense_date": "2026-05-14",
  "notes": null
}
```

---

## 16. Stock Opname

A session for counting physical stock and reconciling it with system stock. Each session auto-snapshots every active variant in the branch. On completion, the system generates an `ADJUST` stock movement for each variance.

Status: `in_progress` | `completed` | `cancelled`.

| Method | Path | Description |
| --- | --- | --- |
| `POST` | `/stock-opnames` | Create (auto-snapshot) |
| `GET` | `/stock-opnames` | List (query: `page`, `per_page`, `status`, `branch_id`) |
| `GET` | `/stock-opnames/:id` | Get with items |
| `PATCH` | `/stock-opnames/:id/items/:itemId` | Update count for one item |
| `PATCH` | `/stock-opnames/:id/items` | Bulk update counts |
| `POST` | `/stock-opnames/:id/complete` | Complete the session and apply adjustments |
| `POST` | `/stock-opnames/:id/cancel` | Cancel the session |

**Create body:**
```json
{ "branch_id": 1, "notes": "Opname bulanan", "category_id": null }
```

`category_id` is optional — if provided, the snapshot is restricted to variants in that category.

**Update single item body:**
```json
{ "counted_stock": 47, "notes": "OK" }
```

**Bulk update body:**
```json
{
  "items": [
    { "item_id": 12, "counted_stock": 47 },
    { "item_id": 13, "counted_stock": 30, "notes": "Some items expired" }
  ]
}
```

**Opname response (excerpt):**
```json
{
  "id": 2,
  "branch_id": 1,
  "opname_number": "OPN-20260514-0001",
  "status": "in_progress",
  "total_items": 32,
  "counted_items": 12,
  "total_variance_qty": 0,
  "total_variance_value": 0,
  "started_at": "2026-05-14T09:15:00Z",
  "items": [
    {
      "id": 12,
      "product_variant_id": 7,
      "sku": "KS-REG",
      "product_name": "Kopi Susu",
      "variant_name": "Regular",
      "system_stock": 50,
      "counted_stock": 47,
      "variance_qty": -3,
      "unit_cost": 9000,
      "variance_value": -27000,
      "counted_at": "2026-05-14T09:30:00Z"
    }
  ]
}
```

On `POST /complete`, for each item where `counted_stock != null`:
- If `counted_stock != current_system_stock`, the system creates a `StockMovement` with `type=ADJUST` (`reference_type=stock_opname`, `reference_id=opname.id`) and upserts the `stocks` table.
- Uncounted items are left untouched.

---

## Error Codes

| Code | HTTP | Meaning |
| --- | --- | --- |
| `VALIDATION_ERROR` | 422 | Body/query failed validation (field-level detail in `errors`) |
| `UNAUTHORIZED` | 401 | Token invalid / expired |
| `FORBIDDEN` | 403 | Resource does not belong to the user's company, or insufficient permissions |
| `NOT_FOUND` | 404 | Resource not found |
| `BAD_REQUEST` | 400 | Business rule violation (e.g. insufficient stock, opname not in_progress) |
| `INTERNAL_ERROR` | 500 | Unexpected server error |

---

## Pagination Conventions

Default `per_page=20`, max `100`. The first page is `1`.
List endpoints return `meta.pagination` (see _Standard Response Format_).
`/users` uses `limit` instead of `per_page` for backward compatibility.

## Multi-tenant Scoping

All protected endpoints are automatically scoped to the user's `company_id` from the JWT. Cross-company access attempts return `403 FORBIDDEN`.
Endpoints that accept `branch_id` in the body or query validate that the branch belongs to the same company.

## Receipts (Optional)

If MongoDB is configured (`MONGO_URI` env set), `GET /orders/:id/receipt` is active and returns the receipt document created when the order completed. Without MongoDB, the endpoint is not registered.

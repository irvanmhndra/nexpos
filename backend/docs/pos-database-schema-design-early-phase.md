# POS Database Schema Design Documentation (Final – Early Phase, Refactorable)

## Overview

This document defines the **final database schema for early-phase research and development**
of a **company-based, multi-branch Point of Sale (POS) system**.

This version intentionally prioritizes:
- Development speed
- Debuggability
- Observability

✅ **Resolved**  
The early-phase raw token storage has been replaced: migration `000036_hash_session_tokens`
stores only SHA-256 digests (see `user_sessions` and the Exit Plan below).

---

## Core Design Principles

1. Company-based tenancy (strict isolation)
2. Branch-scoped operations
3. Variant-based inventory (no stock on product)
4. Ledger-based stock movement
5. Immutable transaction tables
6. Snapshot-by-table (not per column)
7. Promotions never mutate master data
8. OAuth 2.0 with multi-session support (early-phase mode)

---

## Naming Conventions

### Table Naming
- plural nouns
- snake_case
- domain explicit

---

### Foreign Key Naming

**Default**
```
<parent_table_singular>_id
```

Examples:
```
user_id
product_variant_id
promotion_id
```

**Intentional business exception**
```
cashier_id  → users.id
```

---

### Snapshot Strategy

- All `order_*` tables are **immutable snapshots**
- Column names do NOT use `_snapshot`
- Reporting MUST NOT join master tables

---

## Authentication & Authorization (OAuth 2.0 + RBAC)

### users
```
users
- id (PK)
- company_id
- email
- password_hash
- status
- created_at
```

---

### oauth_clients
```
oauth_clients
- id (PK)
- client_id
- name
- type              -- public | confidential
```

---

### user_sessions
```
user_sessions
- id (PK)
- user_id
- company_id
- branch_id (nullable)
- terminal_id (nullable)

- access_token_hash            -- hex SHA-256 of the access token
- access_token_expires_at      -- now() + 2 hours

- refresh_token_hash           -- hex SHA-256 of the refresh token
- refresh_token_expires_at     -- refresh token expiry

- is_revoked
- device_info
- ip_address
- last_used_at
- created_at
```

Tokens are random 256-bit values returned to the client once
(`pkg/sessiontoken`); the table only holds their digests, so a leaked backup or
replica cannot be replayed. Refreshing rotates the session atomically
(`UserSessionRepository.Rotate`): the old row is revoked with a conditional
update and a refresh token can be exchanged only once.

---

### roles / permissions
```
roles
permissions
role_permissions
user_roles
user_branches
```

Authorization hierarchy:
```
User → Role → Permission → Branch
```

---

## Products & Categories

### product_categories
```
product_categories
- id (PK)
- company_id
- parent_id (nullable)
- code
- name
- sort_order
- is_active
```

---

### products
```
products
- id (PK)
- company_id
- product_category_id
- name
- description
- is_active
```

---

### product_variants
```
product_variants
- id (PK)
- product_id
- sku
- attributes (JSON)
- price
- standard_cost
- last_purchase_cost
- is_default
- is_active
```

---

## Inventory

### stocks
```
stocks
- id (PK)
- product_variant_id
- branch_id
- quantity
- updated_at
```

---

### stock_movements (source of truth)
```
stock_movements
- id (PK)
- product_variant_id
- branch_id
- type              -- IN | OUT | ADJUST | TRANSFER
- quantity
- unit_cost         -- only for IN
- reference_type
- reference_id
- note
- created_at
```

---

## Customers

```
customers
- id (PK)
- company_id
- code
- name
- phone
- email
- is_member
- created_at
```

---

## Orders (Immutable Snapshot)

### orders
```
orders
- id (UUID, PK)
- company_id
- branch_id
- order_no
- customer_id (nullable)
- cashier_id         -- FK → users.id (intentional exception)
- status
- total_amount
- total_discount
- total_tax
- grand_total
- applied_promotions (JSON)
- is_offline
- created_at
```

---

### order_items
```
order_items
- id (PK)
- order_id
- product_id
- product_variant_id
- sku
- product_name
- variant_attributes (JSON)
- unit_price
- unit_cost
- quantity
- discount_amount
- tax_amount
- subtotal
- cogs_amount
```

---

### payments
```
payments
- id (PK)
- order_id
- method
- amount
- paid_at
- reference_no
```

---

## Promotions & Bundles

### promotions
```
promotions
- id (PK)
- company_id
- code
- name
- type              -- discount | bundle | conditional
- start_at
- end_at
- priority
- is_active
```

---

### promotion_rules
```
promotion_rules
- id (PK)
- promotion_id
- rule_type
- operator
- value (JSON)
```

---

### promotion_rewards
```
promotion_rewards
- id (PK)
- promotion_id
- reward_type
- value (JSON)
```

---

## POS Terminals & Invoice Numbering

### pos_terminals
```
pos_terminals
- id (PK)
- branch_id
- code
```

Invoice format:
```
INV-{BRANCH_CODE}-{YYYYMMDD}-{SEQUENCE}
```

---

## Exit Plan — done (migration 000036)

```
- access_token       → access_token_hash      ✅ renamed, existing rows hashed in place
- refresh_token      → refresh_token_hash     ✅ renamed, existing rows hashed in place
- DROP raw token columns                      ✅ no raw column remains
```

Existing sessions keep working across the upgrade because the migration and
the API hash with the same function (`encode(sha256(convert_to(t, 'UTF8')), 'hex')`
= `sessiontoken.Hash`). Rolling back 000036 revokes every session, since
digests cannot be turned back into tokens.

---

## Final Notes

This schema is **intentionally optimized for early-phase development**.

It is:
- fast to implement
- easy to debug
- safe to refactor later

⚠️ **DO NOT SHIP TO PRODUCTION WITHOUT TOKEN REFACTOR**

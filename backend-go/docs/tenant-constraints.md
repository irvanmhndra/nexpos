# Tenant constraints

Migration `000037_tenant_composite_foreign_keys` adds composite foreign keys
so a row can only reference a row of the **same company** (for example an
order's customer, branch, and cashier). Each parent table has
`UNIQUE (company_id, id)` and each reference is
`FOREIGN KEY (company_id, <column>) REFERENCES <parent> (company_id, id)`.

The foreign keys are created `NOT VALID`: Postgres enforces them for every
new insert and update, but did not check rows that existed before the
migration. That keeps the migration from failing on old data during an
unattended deploy.

## Validating existing rows (once, after deploying)

1. Look for rows that already reference another company's data. Every count
   should be 0:

```sql
SELECT 'user_sessions.branch_id' AS reference, count(*) AS cross_company_rows
  FROM user_sessions ch JOIN branches pa ON pa.id = ch.branch_id
 WHERE pa.company_id <> ch.company_id
UNION ALL
SELECT 'user_sessions.user_id' AS reference, count(*) AS cross_company_rows
  FROM user_sessions ch JOIN users pa ON pa.id = ch.user_id
 WHERE pa.company_id <> ch.company_id
UNION ALL
SELECT 'product_categories.parent_id' AS reference, count(*) AS cross_company_rows
  FROM product_categories ch JOIN product_categories pa ON pa.id = ch.parent_id
 WHERE pa.company_id <> ch.company_id
UNION ALL
SELECT 'products.product_category_id' AS reference, count(*) AS cross_company_rows
  FROM products ch JOIN product_categories pa ON pa.id = ch.product_category_id
 WHERE pa.company_id <> ch.company_id
UNION ALL
SELECT 'orders.branch_id' AS reference, count(*) AS cross_company_rows
  FROM orders ch JOIN branches pa ON pa.id = ch.branch_id
 WHERE pa.company_id <> ch.company_id
UNION ALL
SELECT 'orders.cashier_id' AS reference, count(*) AS cross_company_rows
  FROM orders ch JOIN users pa ON pa.id = ch.cashier_id
 WHERE pa.company_id <> ch.company_id
UNION ALL
SELECT 'orders.customer_id' AS reference, count(*) AS cross_company_rows
  FROM orders ch JOIN customers pa ON pa.id = ch.customer_id
 WHERE pa.company_id <> ch.company_id
UNION ALL
SELECT 'purchase_orders.branch_id' AS reference, count(*) AS cross_company_rows
  FROM purchase_orders ch JOIN branches pa ON pa.id = ch.branch_id
 WHERE pa.company_id <> ch.company_id
UNION ALL
SELECT 'purchase_orders.supplier_id' AS reference, count(*) AS cross_company_rows
  FROM purchase_orders ch JOIN suppliers pa ON pa.id = ch.supplier_id
 WHERE pa.company_id <> ch.company_id
UNION ALL
SELECT 'shifts.branch_id' AS reference, count(*) AS cross_company_rows
  FROM shifts ch JOIN branches pa ON pa.id = ch.branch_id
 WHERE pa.company_id <> ch.company_id
UNION ALL
SELECT 'shifts.cashier_id' AS reference, count(*) AS cross_company_rows
  FROM shifts ch JOIN users pa ON pa.id = ch.cashier_id
 WHERE pa.company_id <> ch.company_id
UNION ALL
SELECT 'expenses.branch_id' AS reference, count(*) AS cross_company_rows
  FROM expenses ch JOIN branches pa ON pa.id = ch.branch_id
 WHERE pa.company_id <> ch.company_id
UNION ALL
SELECT 'expenses.category_id' AS reference, count(*) AS cross_company_rows
  FROM expenses ch JOIN expense_categories pa ON pa.id = ch.category_id
 WHERE pa.company_id <> ch.company_id
UNION ALL
SELECT 'expenses.recorded_by' AS reference, count(*) AS cross_company_rows
  FROM expenses ch JOIN users pa ON pa.id = ch.recorded_by
 WHERE pa.company_id <> ch.company_id
UNION ALL
SELECT 'stock_opnames.branch_id' AS reference, count(*) AS cross_company_rows
  FROM stock_opnames ch JOIN branches pa ON pa.id = ch.branch_id
 WHERE pa.company_id <> ch.company_id
UNION ALL
SELECT 'stock_opnames.started_by' AS reference, count(*) AS cross_company_rows
  FROM stock_opnames ch JOIN users pa ON pa.id = ch.started_by
 WHERE pa.company_id <> ch.company_id
UNION ALL
SELECT 'stock_opnames.completed_by' AS reference, count(*) AS cross_company_rows
  FROM stock_opnames ch JOIN users pa ON pa.id = ch.completed_by
 WHERE pa.company_id <> ch.company_id
UNION ALL
SELECT 'daily_settlements.branch_id' AS reference, count(*) AS cross_company_rows
  FROM daily_settlements ch JOIN branches pa ON pa.id = ch.branch_id
 WHERE pa.company_id <> ch.company_id
UNION ALL
SELECT 'daily_settlements.recorded_by' AS reference, count(*) AS cross_company_rows
  FROM daily_settlements ch JOIN users pa ON pa.id = ch.recorded_by
 WHERE pa.company_id <> ch.company_id
UNION ALL
SELECT 'daily_settlements.finalized_by' AS reference, count(*) AS cross_company_rows
  FROM daily_settlements ch JOIN users pa ON pa.id = ch.finalized_by
 WHERE pa.company_id <> ch.company_id
UNION ALL
SELECT 'receipts.order_id' AS reference, count(*) AS cross_company_rows
  FROM receipts ch JOIN orders pa ON pa.id = ch.order_id
 WHERE pa.company_id <> ch.company_id;
```

2. If a count is not 0, inspect and fix those rows first (they are data
   that crossed tenants before this protection existed).

3. Mark the constraints as validated (this scans the tables but takes only a
   light lock, so it can run while the app is serving):

```sql
ALTER TABLE user_sessions VALIDATE CONSTRAINT fk_user_sessions_branch_id_same_company;
ALTER TABLE user_sessions VALIDATE CONSTRAINT fk_user_sessions_user_id_same_company;
ALTER TABLE product_categories VALIDATE CONSTRAINT fk_product_categories_parent_id_same_company;
ALTER TABLE products VALIDATE CONSTRAINT fk_products_product_category_id_same_company;
ALTER TABLE orders VALIDATE CONSTRAINT fk_orders_branch_id_same_company;
ALTER TABLE orders VALIDATE CONSTRAINT fk_orders_cashier_id_same_company;
ALTER TABLE orders VALIDATE CONSTRAINT fk_orders_customer_id_same_company;
ALTER TABLE purchase_orders VALIDATE CONSTRAINT fk_purchase_orders_branch_id_same_company;
ALTER TABLE purchase_orders VALIDATE CONSTRAINT fk_purchase_orders_supplier_id_same_company;
ALTER TABLE shifts VALIDATE CONSTRAINT fk_shifts_branch_id_same_company;
ALTER TABLE shifts VALIDATE CONSTRAINT fk_shifts_cashier_id_same_company;
ALTER TABLE expenses VALIDATE CONSTRAINT fk_expenses_branch_id_same_company;
ALTER TABLE expenses VALIDATE CONSTRAINT fk_expenses_category_id_same_company;
ALTER TABLE expenses VALIDATE CONSTRAINT fk_expenses_recorded_by_same_company;
ALTER TABLE stock_opnames VALIDATE CONSTRAINT fk_stock_opnames_branch_id_same_company;
ALTER TABLE stock_opnames VALIDATE CONSTRAINT fk_stock_opnames_started_by_same_company;
ALTER TABLE stock_opnames VALIDATE CONSTRAINT fk_stock_opnames_completed_by_same_company;
ALTER TABLE daily_settlements VALIDATE CONSTRAINT fk_daily_settlements_branch_id_same_company;
ALTER TABLE daily_settlements VALIDATE CONSTRAINT fk_daily_settlements_recorded_by_same_company;
ALTER TABLE daily_settlements VALIDATE CONSTRAINT fk_daily_settlements_finalized_by_same_company;
ALTER TABLE receipts VALIDATE CONSTRAINT fk_receipts_order_id_same_company;
```

`users.role_id` has no composite key on purpose: system roles are shared
across companies (`roles.company_id` is NULL).

Tables without a `company_id` column (order items, payments, product
variants, stocks, stock movements, PO items) inherit their tenant from their
parent row; the services scope those lookups (see
`ProductVariantRepository.GetByIDForCompany`).

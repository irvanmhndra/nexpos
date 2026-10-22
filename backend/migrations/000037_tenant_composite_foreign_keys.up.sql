-- Tenant-safe foreign keys: a row may only reference a row of the same
-- company. The existing single-column FKs only check that the referenced row
-- exists, so an order could point at another company's customer or branch
-- if the application ever let a foreign ID through.
--
-- Each parent gets UNIQUE (company_id, id) as the target, and each reference
-- gets a composite FK (company_id, <col>) -> parent (company_id, id). A NULL
-- reference is not checked (MATCH SIMPLE), as before.
--
-- The constraints are NOT VALID: Postgres enforces them for every insert and
-- update from now on but does not scan existing rows, so this migration cannot
-- fail on legacy data and leave the schema dirty. Check and validate existing
-- rows afterwards with the queries in docs/tenant-constraints.md.

ALTER TABLE branches ADD CONSTRAINT uq_branches_company_id_id UNIQUE (company_id, id);
ALTER TABLE users ADD CONSTRAINT uq_users_company_id_id UNIQUE (company_id, id);
ALTER TABLE customers ADD CONSTRAINT uq_customers_company_id_id UNIQUE (company_id, id);
ALTER TABLE product_categories ADD CONSTRAINT uq_product_categories_company_id_id UNIQUE (company_id, id);
ALTER TABLE suppliers ADD CONSTRAINT uq_suppliers_company_id_id UNIQUE (company_id, id);
ALTER TABLE expense_categories ADD CONSTRAINT uq_expense_categories_company_id_id UNIQUE (company_id, id);
ALTER TABLE orders ADD CONSTRAINT uq_orders_company_id_id UNIQUE (company_id, id);

ALTER TABLE user_sessions ADD CONSTRAINT fk_user_sessions_branch_id_same_company
    FOREIGN KEY (company_id, branch_id) REFERENCES branches (company_id, id) NOT VALID;
ALTER TABLE user_sessions ADD CONSTRAINT fk_user_sessions_user_id_same_company
    FOREIGN KEY (company_id, user_id) REFERENCES users (company_id, id) NOT VALID;
ALTER TABLE product_categories ADD CONSTRAINT fk_product_categories_parent_id_same_company
    FOREIGN KEY (company_id, parent_id) REFERENCES product_categories (company_id, id) NOT VALID;
ALTER TABLE products ADD CONSTRAINT fk_products_product_category_id_same_company
    FOREIGN KEY (company_id, product_category_id) REFERENCES product_categories (company_id, id) NOT VALID;
ALTER TABLE orders ADD CONSTRAINT fk_orders_branch_id_same_company
    FOREIGN KEY (company_id, branch_id) REFERENCES branches (company_id, id) NOT VALID;
ALTER TABLE orders ADD CONSTRAINT fk_orders_cashier_id_same_company
    FOREIGN KEY (company_id, cashier_id) REFERENCES users (company_id, id) NOT VALID;
ALTER TABLE orders ADD CONSTRAINT fk_orders_customer_id_same_company
    FOREIGN KEY (company_id, customer_id) REFERENCES customers (company_id, id) NOT VALID;
ALTER TABLE purchase_orders ADD CONSTRAINT fk_purchase_orders_branch_id_same_company
    FOREIGN KEY (company_id, branch_id) REFERENCES branches (company_id, id) NOT VALID;
ALTER TABLE purchase_orders ADD CONSTRAINT fk_purchase_orders_supplier_id_same_company
    FOREIGN KEY (company_id, supplier_id) REFERENCES suppliers (company_id, id) NOT VALID;
ALTER TABLE shifts ADD CONSTRAINT fk_shifts_branch_id_same_company
    FOREIGN KEY (company_id, branch_id) REFERENCES branches (company_id, id) NOT VALID;
ALTER TABLE shifts ADD CONSTRAINT fk_shifts_cashier_id_same_company
    FOREIGN KEY (company_id, cashier_id) REFERENCES users (company_id, id) NOT VALID;
ALTER TABLE expenses ADD CONSTRAINT fk_expenses_branch_id_same_company
    FOREIGN KEY (company_id, branch_id) REFERENCES branches (company_id, id) NOT VALID;
ALTER TABLE expenses ADD CONSTRAINT fk_expenses_category_id_same_company
    FOREIGN KEY (company_id, category_id) REFERENCES expense_categories (company_id, id) NOT VALID;
ALTER TABLE expenses ADD CONSTRAINT fk_expenses_recorded_by_same_company
    FOREIGN KEY (company_id, recorded_by) REFERENCES users (company_id, id) NOT VALID;
ALTER TABLE stock_opnames ADD CONSTRAINT fk_stock_opnames_branch_id_same_company
    FOREIGN KEY (company_id, branch_id) REFERENCES branches (company_id, id) NOT VALID;
ALTER TABLE stock_opnames ADD CONSTRAINT fk_stock_opnames_started_by_same_company
    FOREIGN KEY (company_id, started_by) REFERENCES users (company_id, id) NOT VALID;
ALTER TABLE stock_opnames ADD CONSTRAINT fk_stock_opnames_completed_by_same_company
    FOREIGN KEY (company_id, completed_by) REFERENCES users (company_id, id) NOT VALID;
ALTER TABLE daily_settlements ADD CONSTRAINT fk_daily_settlements_branch_id_same_company
    FOREIGN KEY (company_id, branch_id) REFERENCES branches (company_id, id) NOT VALID;
ALTER TABLE daily_settlements ADD CONSTRAINT fk_daily_settlements_recorded_by_same_company
    FOREIGN KEY (company_id, recorded_by) REFERENCES users (company_id, id) NOT VALID;
ALTER TABLE daily_settlements ADD CONSTRAINT fk_daily_settlements_finalized_by_same_company
    FOREIGN KEY (company_id, finalized_by) REFERENCES users (company_id, id) NOT VALID;
ALTER TABLE receipts ADD CONSTRAINT fk_receipts_order_id_same_company
    FOREIGN KEY (company_id, order_id) REFERENCES orders (company_id, id) NOT VALID;

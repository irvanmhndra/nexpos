-- Remove check constraints
ALTER TABLE orders DROP CONSTRAINT IF EXISTS chk_orders_status;
ALTER TABLE orders DROP CONSTRAINT IF EXISTS chk_orders_payment_status;
ALTER TABLE orders DROP CONSTRAINT IF EXISTS chk_orders_fulfillment_type;
ALTER TABLE orders DROP CONSTRAINT IF EXISTS chk_orders_fulfillment_status;

-- Remove indexes
DROP INDEX IF EXISTS idx_orders_payment_status;
DROP INDEX IF EXISTS idx_orders_fulfillment_type;
DROP INDEX IF EXISTS idx_orders_offline_id;

-- Remove columns
ALTER TABLE orders DROP COLUMN IF EXISTS payment_status;
ALTER TABLE orders DROP COLUMN IF EXISTS fulfillment_type;
ALTER TABLE orders DROP COLUMN IF EXISTS fulfillment_status;
ALTER TABLE orders DROP COLUMN IF EXISTS shipping_address;
ALTER TABLE orders DROP COLUMN IF EXISTS confirmed_at;
ALTER TABLE orders DROP COLUMN IF EXISTS paid_at;
ALTER TABLE orders DROP COLUMN IF EXISTS completed_at;
ALTER TABLE orders DROP COLUMN IF EXISTS cancelled_at;
ALTER TABLE orders DROP COLUMN IF EXISTS voided_at;
ALTER TABLE orders DROP COLUMN IF EXISTS cancel_reason;
ALTER TABLE orders DROP COLUMN IF EXISTS void_reason;
ALTER TABLE orders DROP COLUMN IF EXISTS offline_id;
ALTER TABLE orders DROP COLUMN IF EXISTS synced_at;

-- Revert default status
ALTER TABLE orders ALTER COLUMN status SET DEFAULT 'pending';

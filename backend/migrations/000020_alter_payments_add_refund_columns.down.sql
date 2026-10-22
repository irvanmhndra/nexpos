-- Remove check constraints
ALTER TABLE payments DROP CONSTRAINT IF EXISTS chk_payments_status;
ALTER TABLE payments DROP CONSTRAINT IF EXISTS chk_payments_method;
ALTER TABLE payments DROP CONSTRAINT IF EXISTS chk_payments_refund_amount;

-- Remove indexes
DROP INDEX IF EXISTS idx_payments_status;

-- Remove columns
ALTER TABLE payments DROP COLUMN IF EXISTS status;
ALTER TABLE payments DROP COLUMN IF EXISTS refunded_amount;
ALTER TABLE payments DROP COLUMN IF EXISTS refunded_at;
ALTER TABLE payments DROP COLUMN IF EXISTS refund_reason;

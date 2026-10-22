-- Add refund support columns to payments table

-- Payment status: tracks individual payment lifecycle
ALTER TABLE payments ADD COLUMN status VARCHAR(20) NOT NULL DEFAULT 'completed';

-- Refund tracking
ALTER TABLE payments ADD COLUMN refunded_amount DECIMAL(15, 2) NOT NULL DEFAULT 0;
ALTER TABLE payments ADD COLUMN refunded_at TIMESTAMPTZ NULL;
ALTER TABLE payments ADD COLUMN refund_reason TEXT NULL;

-- Add check constraints for valid values
ALTER TABLE payments ADD CONSTRAINT chk_payments_status
    CHECK (status IN ('pending', 'completed', 'refunded', 'partially_refunded', 'failed'));

ALTER TABLE payments ADD CONSTRAINT chk_payments_method
    CHECK (method IN ('cash', 'debit_card', 'credit_card', 'e_wallet', 'bank_transfer', 'qris'));

-- Ensure refunded_amount doesn't exceed original amount
ALTER TABLE payments ADD CONSTRAINT chk_payments_refund_amount
    CHECK (refunded_amount >= 0 AND refunded_amount <= amount);

-- Index for status queries
CREATE INDEX idx_payments_status ON payments(status);

COMMENT ON COLUMN payments.status IS 'Payment lifecycle: pending → completed | failed, completed → refunded | partially_refunded';
COMMENT ON COLUMN payments.refunded_amount IS 'Amount refunded from this payment (can be partial)';

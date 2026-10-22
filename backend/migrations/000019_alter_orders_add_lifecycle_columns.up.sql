-- Add new lifecycle columns to orders table for proper state management

-- Payment status: tracks payment lifecycle separately from order status
ALTER TABLE orders ADD COLUMN payment_status VARCHAR(20) NOT NULL DEFAULT 'unpaid';

-- Fulfillment type: determines if order needs shipping
ALTER TABLE orders ADD COLUMN fulfillment_type VARCHAR(20) NOT NULL DEFAULT 'counter';

-- Fulfillment status: only relevant for delivery orders
ALTER TABLE orders ADD COLUMN fulfillment_status VARCHAR(20) NULL;

-- Shipping address: only for delivery orders
ALTER TABLE orders ADD COLUMN shipping_address TEXT NULL;

-- Lifecycle timestamps
ALTER TABLE orders ADD COLUMN confirmed_at TIMESTAMPTZ NULL;
ALTER TABLE orders ADD COLUMN paid_at TIMESTAMPTZ NULL;
ALTER TABLE orders ADD COLUMN completed_at TIMESTAMPTZ NULL;
ALTER TABLE orders ADD COLUMN cancelled_at TIMESTAMPTZ NULL;
ALTER TABLE orders ADD COLUMN voided_at TIMESTAMPTZ NULL;

-- Void/Cancel reason for audit
ALTER TABLE orders ADD COLUMN cancel_reason TEXT NULL;
ALTER TABLE orders ADD COLUMN void_reason TEXT NULL;

-- Offline sync support
ALTER TABLE orders ADD COLUMN offline_id UUID NULL;
ALTER TABLE orders ADD COLUMN synced_at TIMESTAMPTZ NULL;

-- Update default status from 'pending' to 'draft'
ALTER TABLE orders ALTER COLUMN status SET DEFAULT 'draft';

-- Add indexes for new columns
CREATE INDEX idx_orders_payment_status ON orders(company_id, payment_status);
CREATE INDEX idx_orders_fulfillment_type ON orders(company_id, fulfillment_type);
CREATE INDEX idx_orders_offline_id ON orders(offline_id) WHERE offline_id IS NOT NULL;

-- Add check constraints for valid status values
ALTER TABLE orders ADD CONSTRAINT chk_orders_status
    CHECK (status IN ('draft', 'confirmed', 'completed', 'cancelled', 'voided'));

ALTER TABLE orders ADD CONSTRAINT chk_orders_payment_status
    CHECK (payment_status IN ('unpaid', 'partial', 'paid', 'refunded'));

ALTER TABLE orders ADD CONSTRAINT chk_orders_fulfillment_type
    CHECK (fulfillment_type IN ('counter', 'delivery'));

ALTER TABLE orders ADD CONSTRAINT chk_orders_fulfillment_status
    CHECK (fulfillment_status IS NULL OR fulfillment_status IN ('pending', 'processing', 'shipped', 'delivered'));

COMMENT ON COLUMN orders.status IS 'Order lifecycle: draft → confirmed → completed | cancelled | voided';
COMMENT ON COLUMN orders.payment_status IS 'Payment lifecycle: unpaid → partial → paid | refunded';
COMMENT ON COLUMN orders.fulfillment_type IS 'counter = in-store pickup, delivery = needs shipping';
COMMENT ON COLUMN orders.fulfillment_status IS 'Only for delivery: pending → processing → shipped → delivered';

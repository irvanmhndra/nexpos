ALTER TABLE product_variants
DROP COLUMN IF EXISTS sale_price,
DROP COLUMN IF EXISTS sale_start,
DROP COLUMN IF EXISTS sale_end;

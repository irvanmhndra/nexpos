DROP INDEX IF EXISTS idx_product_variants_barcode;
ALTER TABLE product_variants DROP COLUMN IF EXISTS barcode;

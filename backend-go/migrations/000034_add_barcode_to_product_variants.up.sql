-- Manufacturer/printed barcode (EAN/UPC), distinct from the internal SKU.
-- Nullable: not every product is sold by barcode. Uniqueness is enforced
-- per-company in the service layer (variants carry tenant via their product).
ALTER TABLE product_variants ADD COLUMN barcode VARCHAR(100);

CREATE INDEX idx_product_variants_barcode
    ON product_variants (barcode)
    WHERE barcode IS NOT NULL;

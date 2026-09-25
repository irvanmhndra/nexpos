package postgres

import (
	"context"
	"database/sql"

	"github.com/irvanmhndra/nexpos-api/internal/model"
	"github.com/jmoiron/sqlx"
)

type ProductVariantRepository struct {
	db *sqlx.DB
}

func NewProductVariantRepository(db *sqlx.DB) *ProductVariantRepository {
	return &ProductVariantRepository{db: db}
}

func (r *ProductVariantRepository) Create(ctx context.Context, variant *model.ProductVariant) error {
	query := `
		INSERT INTO product_variants (product_id, sku, name, attributes, price, standard_cost, last_purchase_cost, is_default, is_active, sale_price, sale_start, sale_end, barcode)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id, created_at, updated_at
	`
	return conn(ctx, r.db).QueryRowContext(ctx, query,
		variant.ProductID,
		variant.SKU,
		variant.Name,
		variant.Attributes,
		variant.Price,
		variant.StandardCost,
		variant.LastPurchaseCost,
		variant.IsDefault,
		variant.IsActive,
		variant.SalePrice,
		variant.SaleStart,
		variant.SaleEnd,
		variant.Barcode,
	).Scan(&variant.ID, &variant.CreatedAt, &variant.UpdatedAt)
}

func (r *ProductVariantRepository) GetByID(ctx context.Context, id int64) (*model.ProductVariant, error) {
	var variant model.ProductVariant
	query := `SELECT pv.*, p.name AS product_name FROM product_variants pv JOIN products p ON p.id = pv.product_id WHERE pv.id = $1`
	err := conn(ctx, r.db).GetContext(ctx, &variant, query, id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &variant, nil
}

func (r *ProductVariantRepository) GetByProductID(ctx context.Context, productID int64) ([]*model.ProductVariant, error) {
	var variants []*model.ProductVariant
	query := `SELECT * FROM product_variants WHERE product_id = $1 ORDER BY is_default DESC, name ASC`
	if err := conn(ctx, r.db).SelectContext(ctx, &variants, query, productID); err != nil {
		return nil, err
	}
	return variants, nil
}

func (r *ProductVariantRepository) GetBySKU(ctx context.Context, sku string) (*model.ProductVariant, error) {
	var variant model.ProductVariant
	query := `SELECT * FROM product_variants WHERE sku = $1`
	err := conn(ctx, r.db).GetContext(ctx, &variant, query, sku)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &variant, nil
}

func (r *ProductVariantRepository) Update(ctx context.Context, variant *model.ProductVariant) error {
	query := `
		UPDATE product_variants
		SET sku = $1, name = $2, attributes = $3, price = $4, standard_cost = $5, last_purchase_cost = $6, is_default = $7, is_active = $8, sale_price = $9, sale_start = $10, sale_end = $11, barcode = $12, updated_at = NOW()
		WHERE id = $13
	`
	result, err := conn(ctx, r.db).ExecContext(ctx, query,
		variant.SKU,
		variant.Name,
		variant.Attributes,
		variant.Price,
		variant.StandardCost,
		variant.LastPurchaseCost,
		variant.IsDefault,
		variant.IsActive,
		variant.SalePrice,
		variant.SaleStart,
		variant.SaleEnd,
		variant.Barcode,
		variant.ID,
	)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *ProductVariantRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM product_variants WHERE id = $1`
	result, err := conn(ctx, r.db).ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *ProductVariantRepository) DeleteByProductID(ctx context.Context, productID int64) error {
	query := `DELETE FROM product_variants WHERE product_id = $1`
	_, err := conn(ctx, r.db).ExecContext(ctx, query, productID)
	return err
}

func (r *ProductVariantRepository) SKUExists(ctx context.Context, sku string, excludeID int64) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM product_variants WHERE sku = $1 AND id != $2)`
	err := conn(ctx, r.db).GetContext(ctx, &exists, query, sku, excludeID)
	return exists, err
}

func (r *ProductVariantRepository) SKUExistsInOtherProduct(ctx context.Context, sku string, productID int64) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM product_variants WHERE sku = $1 AND product_id != $2)`
	err := conn(ctx, r.db).GetContext(ctx, &exists, query, sku, productID)
	return exists, err
}

// FindByCodeInCompany resolves a scanned code to a variant within a company,
// matching the barcode first and falling back to the SKU. Active variants and
// barcode matches are preferred. Returns nil when nothing matches.
func (r *ProductVariantRepository) FindByCodeInCompany(ctx context.Context, companyID int64, code string) (*model.ProductVariant, error) {
	var variant model.ProductVariant
	query := `
		SELECT pv.* FROM product_variants pv
		JOIN products p ON p.id = pv.product_id
		WHERE p.company_id = $1 AND (pv.barcode = $2 OR pv.sku = $2)
		ORDER BY (pv.barcode = $2) DESC, pv.is_active DESC
		LIMIT 1
	`
	err := conn(ctx, r.db).GetContext(ctx, &variant, query, companyID, code)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &variant, nil
}

// BarcodeExistsInOtherProduct reports whether a barcode is already used by a
// variant of a *different* product in the same company (uniqueness is scoped
// per-tenant). Pass productID 0 on create to check against every product.
func (r *ProductVariantRepository) BarcodeExistsInOtherProduct(ctx context.Context, companyID int64, barcode string, productID int64) (bool, error) {
	var exists bool
	query := `
		SELECT EXISTS(
			SELECT 1 FROM product_variants pv
			JOIN products p ON p.id = pv.product_id
			WHERE p.company_id = $1 AND pv.barcode = $2 AND pv.product_id != $3
		)
	`
	err := conn(ctx, r.db).GetContext(ctx, &exists, query, companyID, barcode, productID)
	return exists, err
}

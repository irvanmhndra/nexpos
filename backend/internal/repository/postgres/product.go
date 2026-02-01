package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/irvanmhndra/pos-core-api/internal/model"
	"github.com/jmoiron/sqlx"
)

type ProductRepository struct {
	db *sqlx.DB
}

func NewProductRepository(db *sqlx.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

func (r *ProductRepository) Create(ctx context.Context, product *model.Product) error {
	query := `
		INSERT INTO products (company_id, product_category_id, name, description, image_data, is_active)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at
	`
	return r.db.QueryRowContext(ctx, query,
		product.CompanyID,
		product.ProductCategoryID,
		product.Name,
		product.Description,
		product.ImageData,
		product.IsActive,
	).Scan(&product.ID, &product.CreatedAt, &product.UpdatedAt)
}

func (r *ProductRepository) GetByID(ctx context.Context, companyID, id int64) (*model.Product, error) {
	var product model.Product
	query := `SELECT * FROM products WHERE id = $1 AND company_id = $2`
	err := r.db.GetContext(ctx, &product, query, id, companyID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &product, nil
}

func (r *ProductRepository) List(ctx context.Context, companyID int64, search string, categoryID *int64, isActive *bool, limit, offset int) ([]*model.Product, int, error) {
	var products []*model.Product
	var total int

	// Build WHERE clause
	whereClause := "WHERE company_id = $1"
	args := []interface{}{companyID}
	argIndex := 2

	if search != "" {
		whereClause += fmt.Sprintf(" AND name ILIKE $%d", argIndex)
		args = append(args, "%"+search+"%")
		argIndex++
	}

	if categoryID != nil {
		whereClause += fmt.Sprintf(" AND product_category_id = $%d", argIndex)
		args = append(args, *categoryID)
		argIndex++
	}

	if isActive != nil {
		whereClause += fmt.Sprintf(" AND is_active = $%d", argIndex)
		args = append(args, *isActive)
		argIndex++
	}

	// Count total
	countQuery := "SELECT COUNT(*) FROM products " + whereClause
	if err := r.db.GetContext(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, err
	}

	// Get list
	listQuery := fmt.Sprintf(`
		SELECT * FROM products %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIndex, argIndex+1)
	args = append(args, limit, offset)

	if err := r.db.SelectContext(ctx, &products, listQuery, args...); err != nil {
		return nil, 0, err
	}

	return products, total, nil
}

func (r *ProductRepository) Update(ctx context.Context, product *model.Product) error {
	query := `
		UPDATE products
		SET product_category_id = $1, name = $2, description = $3, image_data = $4, is_active = $5, updated_at = NOW()
		WHERE id = $6 AND company_id = $7
	`
	result, err := r.db.ExecContext(ctx, query,
		product.ProductCategoryID,
		product.Name,
		product.Description,
		product.ImageData,
		product.IsActive,
		product.ID,
		product.CompanyID,
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

func (r *ProductRepository) Delete(ctx context.Context, companyID, id int64) error {
	query := `DELETE FROM products WHERE id = $1 AND company_id = $2`
	result, err := r.db.ExecContext(ctx, query, id, companyID)
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

package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/irvanmhndra/nexpos-api/internal/model"
	"github.com/jmoiron/sqlx"
)

type ProductCategoryRepository struct {
	db *sqlx.DB
}

func NewProductCategoryRepository(db *sqlx.DB) *ProductCategoryRepository {
	return &ProductCategoryRepository{db: db}
}

func (r *ProductCategoryRepository) Create(ctx context.Context, category *model.ProductCategory) error {
	query := `
		INSERT INTO product_categories (company_id, parent_id, code, name, sort_order, is_active)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at
	`
	return r.db.QueryRowContext(ctx, query,
		category.CompanyID,
		category.ParentID,
		category.Code,
		category.Name,
		category.SortOrder,
		category.IsActive,
	).Scan(&category.ID, &category.CreatedAt, &category.UpdatedAt)
}

func (r *ProductCategoryRepository) GetByID(ctx context.Context, companyID, id int64) (*model.ProductCategory, error) {
	var category model.ProductCategory
	query := `SELECT * FROM product_categories WHERE id = $1 AND company_id = $2`
	err := r.db.GetContext(ctx, &category, query, id, companyID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &category, nil
}

func (r *ProductCategoryRepository) GetByCode(ctx context.Context, companyID int64, code string) (*model.ProductCategory, error) {
	var category model.ProductCategory
	query := `SELECT * FROM product_categories WHERE company_id = $1 AND code = $2`
	err := r.db.GetContext(ctx, &category, query, companyID, code)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &category, nil
}

func (r *ProductCategoryRepository) List(ctx context.Context, companyID int64, search string, isActive *bool, limit, offset int) ([]*model.ProductCategory, int, error) {
	var categories []*model.ProductCategory
	var total int

	// Build WHERE clause
	whereClause := "WHERE company_id = $1"
	args := []interface{}{companyID}
	argIndex := 2

	if search != "" {
		whereClause += fmt.Sprintf(" AND (name ILIKE $%d OR code ILIKE $%d)", argIndex, argIndex)
		args = append(args, "%"+search+"%")
		argIndex++
	}

	if isActive != nil {
		whereClause += fmt.Sprintf(" AND is_active = $%d", argIndex)
		args = append(args, *isActive)
		argIndex++
	}

	// Count total
	countQuery := "SELECT COUNT(*) FROM product_categories " + whereClause
	if err := r.db.GetContext(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, err
	}

	// Get list
	listQuery := fmt.Sprintf(`
		SELECT * FROM product_categories %s
		ORDER BY sort_order ASC, name ASC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIndex, argIndex+1)
	args = append(args, limit, offset)

	if err := r.db.SelectContext(ctx, &categories, listQuery, args...); err != nil {
		return nil, 0, err
	}

	return categories, total, nil
}

func (r *ProductCategoryRepository) ListAll(ctx context.Context, companyID int64) ([]*model.ProductCategory, error) {
	var categories []*model.ProductCategory
	query := `
		SELECT * FROM product_categories
		WHERE company_id = $1 AND is_active = true
		ORDER BY sort_order ASC, name ASC
	`
	if err := r.db.SelectContext(ctx, &categories, query, companyID); err != nil {
		return nil, err
	}
	return categories, nil
}

func (r *ProductCategoryRepository) Update(ctx context.Context, category *model.ProductCategory) error {
	query := `
		UPDATE product_categories
		SET code = $1, name = $2, parent_id = $3, sort_order = $4, is_active = $5, updated_at = NOW()
		WHERE id = $6 AND company_id = $7
	`
	result, err := r.db.ExecContext(ctx, query,
		category.Code,
		category.Name,
		category.ParentID,
		category.SortOrder,
		category.IsActive,
		category.ID,
		category.CompanyID,
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

func (r *ProductCategoryRepository) Delete(ctx context.Context, companyID, id int64) error {
	query := `DELETE FROM product_categories WHERE id = $1 AND company_id = $2`
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

func (r *ProductCategoryRepository) CodeExists(ctx context.Context, companyID int64, code string, excludeID int64) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM product_categories WHERE company_id = $1 AND code = $2 AND id != $3)`
	err := r.db.GetContext(ctx, &exists, query, companyID, code, excludeID)
	return exists, err
}

func (r *ProductCategoryRepository) HasChildren(ctx context.Context, companyID, id int64) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM product_categories WHERE company_id = $1 AND parent_id = $2)`
	err := r.db.GetContext(ctx, &exists, query, companyID, id)
	return exists, err
}

func (r *ProductCategoryRepository) HasProducts(ctx context.Context, companyID, id int64) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM products WHERE company_id = $1 AND product_category_id = $2)`
	err := r.db.GetContext(ctx, &exists, query, companyID, id)
	return exists, err
}

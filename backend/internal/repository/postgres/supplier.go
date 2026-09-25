package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/irvanmhndra/nexpos-api/internal/model"
	"github.com/jmoiron/sqlx"
)

type SupplierRepository struct {
	db *sqlx.DB
}

func NewSupplierRepository(db *sqlx.DB) *SupplierRepository {
	return &SupplierRepository{db: db}
}

func (r *SupplierRepository) Create(ctx context.Context, supplier *model.Supplier) error {
	query := `
		INSERT INTO suppliers (company_id, name, code, contact_name, phone, email, address, notes, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at, updated_at
	`
	return conn(ctx, r.db).QueryRowContext(ctx, query,
		supplier.CompanyID,
		supplier.Name,
		supplier.Code,
		supplier.ContactName,
		supplier.Phone,
		supplier.Email,
		supplier.Address,
		supplier.Notes,
		supplier.IsActive,
	).Scan(&supplier.ID, &supplier.CreatedAt, &supplier.UpdatedAt)
}

func (r *SupplierRepository) GetByID(ctx context.Context, companyID, id int64) (*model.Supplier, error) {
	var supplier model.Supplier
	query := `SELECT * FROM suppliers WHERE id = $1 AND company_id = $2`
	err := conn(ctx, r.db).GetContext(ctx, &supplier, query, id, companyID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &supplier, nil
}

func (r *SupplierRepository) GetByCode(ctx context.Context, companyID int64, code string) (*model.Supplier, error) {
	var supplier model.Supplier
	query := `SELECT * FROM suppliers WHERE company_id = $1 AND code = $2`
	err := conn(ctx, r.db).GetContext(ctx, &supplier, query, companyID, code)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &supplier, nil
}

func (r *SupplierRepository) List(ctx context.Context, companyID int64, search string, isActive *bool, limit, offset int) ([]*model.Supplier, int, error) {
	var suppliers []*model.Supplier
	var total int

	whereClause := "WHERE company_id = $1"
	args := []interface{}{companyID}
	argIndex := 2

	if search != "" {
		whereClause += fmt.Sprintf(" AND (name ILIKE $%d OR code ILIKE $%d OR contact_name ILIKE $%d)", argIndex, argIndex, argIndex)
		args = append(args, "%"+search+"%")
		argIndex++
	}

	if isActive != nil {
		whereClause += fmt.Sprintf(" AND is_active = $%d", argIndex)
		args = append(args, *isActive)
		argIndex++
	}

	countQuery := "SELECT COUNT(*) FROM suppliers " + whereClause
	if err := conn(ctx, r.db).GetContext(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, err
	}

	listQuery := fmt.Sprintf(`
		SELECT * FROM suppliers %s
		ORDER BY name ASC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIndex, argIndex+1)
	args = append(args, limit, offset)

	if err := conn(ctx, r.db).SelectContext(ctx, &suppliers, listQuery, args...); err != nil {
		return nil, 0, err
	}

	return suppliers, total, nil
}

func (r *SupplierRepository) Update(ctx context.Context, supplier *model.Supplier) error {
	query := `
		UPDATE suppliers
		SET name = $1, code = $2, contact_name = $3, phone = $4, email = $5, address = $6, notes = $7, is_active = $8, updated_at = NOW()
		WHERE id = $9 AND company_id = $10
	`
	result, err := conn(ctx, r.db).ExecContext(ctx, query,
		supplier.Name,
		supplier.Code,
		supplier.ContactName,
		supplier.Phone,
		supplier.Email,
		supplier.Address,
		supplier.Notes,
		supplier.IsActive,
		supplier.ID,
		supplier.CompanyID,
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

func (r *SupplierRepository) Delete(ctx context.Context, companyID, id int64) error {
	query := `DELETE FROM suppliers WHERE id = $1 AND company_id = $2`
	result, err := conn(ctx, r.db).ExecContext(ctx, query, id, companyID)
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

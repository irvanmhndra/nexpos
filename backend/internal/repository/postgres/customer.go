package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/irvanmhndra/nexpos-api/internal/model"
	"github.com/jmoiron/sqlx"
)

type CustomerRepository struct {
	db *sqlx.DB
}

func NewCustomerRepository(db *sqlx.DB) *CustomerRepository {
	return &CustomerRepository{db: db}
}

func (r *CustomerRepository) Create(ctx context.Context, customer *model.Customer) error {
	query := `
		INSERT INTO customers (company_id, code, name, phone, email, is_member)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at
	`
	return r.db.QueryRowContext(ctx, query,
		customer.CompanyID,
		customer.Code,
		customer.Name,
		customer.Phone,
		customer.Email,
		customer.IsMember,
	).Scan(&customer.ID, &customer.CreatedAt, &customer.UpdatedAt)
}

func (r *CustomerRepository) GetByID(ctx context.Context, companyID, id int64) (*model.Customer, error) {
	var customer model.Customer
	query := `SELECT * FROM customers WHERE id = $1 AND company_id = $2`
	err := r.db.GetContext(ctx, &customer, query, id, companyID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &customer, nil
}

func (r *CustomerRepository) GetByCode(ctx context.Context, companyID int64, code string) (*model.Customer, error) {
	var customer model.Customer
	query := `SELECT * FROM customers WHERE company_id = $1 AND code = $2`
	err := r.db.GetContext(ctx, &customer, query, companyID, code)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &customer, nil
}

func (r *CustomerRepository) List(ctx context.Context, companyID int64, search string, isMember *bool, limit, offset int) ([]*model.Customer, int, error) {
	var customers []*model.Customer
	var total int

	// Build WHERE clause
	whereClause := "WHERE company_id = $1"
	args := []interface{}{companyID}
	argIndex := 2

	if search != "" {
		whereClause += fmt.Sprintf(" AND (name ILIKE $%d OR code ILIKE $%d OR phone ILIKE $%d OR email ILIKE $%d)", argIndex, argIndex, argIndex, argIndex)
		args = append(args, "%"+search+"%")
		argIndex++
	}

	if isMember != nil {
		whereClause += fmt.Sprintf(" AND is_member = $%d", argIndex)
		args = append(args, *isMember)
		argIndex++
	}

	// Count total
	countQuery := "SELECT COUNT(*) FROM customers " + whereClause
	if err := r.db.GetContext(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, err
	}

	// Get list
	listQuery := fmt.Sprintf(`
		SELECT * FROM customers %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIndex, argIndex+1)
	args = append(args, limit, offset)

	if err := r.db.SelectContext(ctx, &customers, listQuery, args...); err != nil {
		return nil, 0, err
	}

	return customers, total, nil
}

func (r *CustomerRepository) Update(ctx context.Context, customer *model.Customer) error {
	query := `
		UPDATE customers
		SET code = $1, name = $2, phone = $3, email = $4, is_member = $5, updated_at = NOW()
		WHERE id = $6 AND company_id = $7
	`
	result, err := r.db.ExecContext(ctx, query,
		customer.Code,
		customer.Name,
		customer.Phone,
		customer.Email,
		customer.IsMember,
		customer.ID,
		customer.CompanyID,
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

func (r *CustomerRepository) Delete(ctx context.Context, companyID, id int64) error {
	query := `DELETE FROM customers WHERE id = $1 AND company_id = $2`
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

func (r *CustomerRepository) CodeExists(ctx context.Context, companyID int64, code string, excludeID int64) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM customers WHERE company_id = $1 AND code = $2 AND id != $3)`
	err := r.db.GetContext(ctx, &exists, query, companyID, code, excludeID)
	return exists, err
}

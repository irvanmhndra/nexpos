package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/irvanmhndra/nexpos-api/internal/model"
	"github.com/irvanmhndra/nexpos-api/internal/repository"
	"github.com/jmoiron/sqlx"
)

type ExpenseCategoryRepository struct {
	db *sqlx.DB
}

func NewExpenseCategoryRepository(db *sqlx.DB) *ExpenseCategoryRepository {
	return &ExpenseCategoryRepository{db: db}
}

func (r *ExpenseCategoryRepository) Create(ctx context.Context, category *model.ExpenseCategory) error {
	query := `
		INSERT INTO expense_categories (company_id, name, description)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at
	`
	return conn(ctx, r.db).QueryRowContext(ctx, query,
		category.CompanyID,
		category.Name,
		category.Description,
	).Scan(&category.ID, &category.CreatedAt, &category.UpdatedAt)
}

func (r *ExpenseCategoryRepository) GetByID(ctx context.Context, companyID, id int64) (*model.ExpenseCategory, error) {
	var category model.ExpenseCategory
	query := `SELECT * FROM expense_categories WHERE id = $1 AND company_id = $2`
	err := conn(ctx, r.db).GetContext(ctx, &category, query, id, companyID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &category, nil
}

func (r *ExpenseCategoryRepository) List(ctx context.Context, companyID int64, search string, limit, offset int) ([]*model.ExpenseCategory, int, error) {
	var categories []*model.ExpenseCategory
	var total int

	whereClause := "WHERE company_id = $1"
	args := []interface{}{companyID}
	argIndex := 2

	if search != "" {
		whereClause += fmt.Sprintf(" AND name ILIKE $%d", argIndex)
		args = append(args, "%"+search+"%")
		argIndex++
	}

	countQuery := "SELECT COUNT(*) FROM expense_categories " + whereClause
	if err := conn(ctx, r.db).GetContext(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, err
	}

	listQuery := fmt.Sprintf(`
		SELECT * FROM expense_categories %s
		ORDER BY name ASC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIndex, argIndex+1)
	args = append(args, limit, offset)

	if err := conn(ctx, r.db).SelectContext(ctx, &categories, listQuery, args...); err != nil {
		return nil, 0, err
	}

	return categories, total, nil
}

func (r *ExpenseCategoryRepository) Update(ctx context.Context, category *model.ExpenseCategory) error {
	query := `
		UPDATE expense_categories
		SET name = $1, description = $2, updated_at = NOW()
		WHERE id = $3 AND company_id = $4
	`
	result, err := conn(ctx, r.db).ExecContext(ctx, query,
		category.Name,
		category.Description,
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

func (r *ExpenseCategoryRepository) Delete(ctx context.Context, companyID, id int64) error {
	query := `DELETE FROM expense_categories WHERE id = $1 AND company_id = $2`
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

// ExpenseRepository

type ExpenseRepository struct {
	db *sqlx.DB
}

func NewExpenseRepository(db *sqlx.DB) *ExpenseRepository {
	return &ExpenseRepository{db: db}
}

func (r *ExpenseRepository) Create(ctx context.Context, expense *model.Expense) error {
	query := `
		INSERT INTO expenses (company_id, branch_id, category_id, amount, description, reference_no, expense_date, recorded_by, notes)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at, updated_at
	`
	return conn(ctx, r.db).QueryRowContext(ctx, query,
		expense.CompanyID,
		expense.BranchID,
		expense.CategoryID,
		expense.Amount,
		expense.Description,
		expense.ReferenceNo,
		expense.ExpenseDate,
		expense.RecordedBy,
		expense.Notes,
	).Scan(&expense.ID, &expense.CreatedAt, &expense.UpdatedAt)
}

func (r *ExpenseRepository) GetByID(ctx context.Context, companyID, id int64) (*model.Expense, error) {
	var expense model.Expense
	query := `SELECT * FROM expenses WHERE id = $1 AND company_id = $2`
	err := conn(ctx, r.db).GetContext(ctx, &expense, query, id, companyID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &expense, nil
}

func (r *ExpenseRepository) List(ctx context.Context, companyID int64, params *repository.ExpenseListParams) ([]*model.Expense, int, error) {
	var expenses []*model.Expense
	var total int

	whereClause := "WHERE company_id = $1"
	args := []interface{}{companyID}
	argIndex := 2

	if params.BranchID != nil {
		whereClause += fmt.Sprintf(" AND branch_id = $%d", argIndex)
		args = append(args, *params.BranchID)
		argIndex++
	}

	if params.CategoryID != nil {
		whereClause += fmt.Sprintf(" AND category_id = $%d", argIndex)
		args = append(args, *params.CategoryID)
		argIndex++
	}

	if params.DateFrom != "" {
		whereClause += fmt.Sprintf(" AND expense_date >= $%d", argIndex)
		args = append(args, params.DateFrom)
		argIndex++
	}

	if params.DateTo != "" {
		whereClause += fmt.Sprintf(" AND expense_date <= $%d", argIndex)
		args = append(args, params.DateTo)
		argIndex++
	}

	countQuery := "SELECT COUNT(*) FROM expenses " + whereClause
	if err := conn(ctx, r.db).GetContext(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, err
	}

	listQuery := fmt.Sprintf(`
		SELECT * FROM expenses %s
		ORDER BY expense_date DESC, created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIndex, argIndex+1)
	args = append(args, params.Limit, params.Offset)

	if err := conn(ctx, r.db).SelectContext(ctx, &expenses, listQuery, args...); err != nil {
		return nil, 0, err
	}

	return expenses, total, nil
}

func (r *ExpenseRepository) Update(ctx context.Context, expense *model.Expense) error {
	query := `
		UPDATE expenses
		SET branch_id = $1, category_id = $2, amount = $3, description = $4, reference_no = $5, expense_date = $6, notes = $7, updated_at = NOW()
		WHERE id = $8 AND company_id = $9
	`
	result, err := conn(ctx, r.db).ExecContext(ctx, query,
		expense.BranchID,
		expense.CategoryID,
		expense.Amount,
		expense.Description,
		expense.ReferenceNo,
		expense.ExpenseDate,
		expense.Notes,
		expense.ID,
		expense.CompanyID,
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

func (r *ExpenseRepository) Delete(ctx context.Context, companyID, id int64) error {
	query := `DELETE FROM expenses WHERE id = $1 AND company_id = $2`
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

func (r *ExpenseRepository) GetTotalByDateRange(ctx context.Context, companyID int64, branchID *int64, from, to string) (float64, error) {
	var total float64

	whereClause := "WHERE company_id = $1"
	args := []interface{}{companyID}
	argIndex := 2

	if branchID != nil {
		whereClause += fmt.Sprintf(" AND branch_id = $%d", argIndex)
		args = append(args, *branchID)
		argIndex++
	}

	if from != "" {
		whereClause += fmt.Sprintf(" AND expense_date >= $%d", argIndex)
		args = append(args, from)
		argIndex++
	}

	if to != "" {
		whereClause += fmt.Sprintf(" AND expense_date <= $%d", argIndex)
		args = append(args, to)
	}

	query := "SELECT COALESCE(SUM(amount), 0) FROM expenses " + whereClause
	if err := conn(ctx, r.db).GetContext(ctx, &total, query, args...); err != nil {
		return 0, err
	}

	return total, nil
}

package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/irvanmhndra/nexpos-api/internal/model"
	"github.com/irvanmhndra/nexpos-api/internal/repository"
	"github.com/jmoiron/sqlx"
)

type ShiftRepository struct {
	db *sqlx.DB
}

func NewShiftRepository(db *sqlx.DB) *ShiftRepository {
	return &ShiftRepository{db: db}
}

func (r *ShiftRepository) Create(ctx context.Context, shift *model.Shift) error {
	query := `
		INSERT INTO shifts (company_id, branch_id, cashier_id, status, opening_float, closing_float, expected_cash, opened_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at
	`
	return conn(ctx, r.db).QueryRowContext(ctx, query,
		shift.CompanyID,
		shift.BranchID,
		shift.CashierID,
		shift.Status,
		shift.OpeningFloat,
		shift.ClosingFloat,
		shift.ExpectedCash,
		shift.OpenedAt,
	).Scan(&shift.ID, &shift.CreatedAt, &shift.UpdatedAt)
}

func (r *ShiftRepository) GetByID(ctx context.Context, companyID, id int64) (*model.Shift, error) {
	var shift model.Shift
	query := `SELECT * FROM shifts WHERE id = $1 AND company_id = $2`
	err := conn(ctx, r.db).GetContext(ctx, &shift, query, id, companyID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &shift, nil
}

func (r *ShiftRepository) GetOpenShift(ctx context.Context, companyID, branchID, cashierID int64) (*model.Shift, error) {
	var shift model.Shift
	query := `SELECT * FROM shifts WHERE company_id = $1 AND branch_id = $2 AND cashier_id = $3 AND status = 'open' ORDER BY opened_at DESC LIMIT 1`
	err := conn(ctx, r.db).GetContext(ctx, &shift, query, companyID, branchID, cashierID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &shift, nil
}

func (r *ShiftRepository) List(ctx context.Context, companyID int64, params *repository.ShiftListParams) ([]*model.Shift, int, error) {
	var shifts []*model.Shift
	var total int

	whereClause := "WHERE company_id = $1"
	args := []interface{}{companyID}
	argIndex := 2

	if params.BranchID != nil {
		whereClause += fmt.Sprintf(" AND branch_id = $%d", argIndex)
		args = append(args, *params.BranchID)
		argIndex++
	}

	if params.CashierID != nil {
		whereClause += fmt.Sprintf(" AND cashier_id = $%d", argIndex)
		args = append(args, *params.CashierID)
		argIndex++
	}

	if params.Status != "" {
		whereClause += fmt.Sprintf(" AND status = $%d", argIndex)
		args = append(args, params.Status)
		argIndex++
	}

	if params.DateFrom != "" {
		whereClause += fmt.Sprintf(" AND opened_at >= $%d", argIndex)
		args = append(args, params.DateFrom)
		argIndex++
	}

	if params.DateTo != "" {
		whereClause += fmt.Sprintf(" AND opened_at < $%d", argIndex)
		args = append(args, params.DateTo)
		argIndex++
	}

	countQuery := "SELECT COUNT(*) FROM shifts " + whereClause
	if err := conn(ctx, r.db).GetContext(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, err
	}

	listQuery := fmt.Sprintf(`
		SELECT * FROM shifts %s
		ORDER BY opened_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIndex, argIndex+1)
	args = append(args, params.Limit, params.Offset)

	if err := conn(ctx, r.db).SelectContext(ctx, &shifts, listQuery, args...); err != nil {
		return nil, 0, err
	}

	return shifts, total, nil
}

func (r *ShiftRepository) Update(ctx context.Context, shift *model.Shift) error {
	query := `
		UPDATE shifts
		SET status = $1, closing_float = $2, expected_cash = $3, actual_cash = $4, cash_difference = $5, notes = $6, closed_at = $7, updated_at = NOW()
		WHERE id = $8 AND company_id = $9
	`
	result, err := conn(ctx, r.db).ExecContext(ctx, query,
		shift.Status,
		shift.ClosingFloat,
		shift.ExpectedCash,
		shift.ActualCash,
		shift.CashDifference,
		shift.Notes,
		shift.ClosedAt,
		shift.ID,
		shift.CompanyID,
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

package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/irvanmhndra/nexpos-api/internal/model"
	"github.com/irvanmhndra/nexpos-api/internal/repository"
	"github.com/jmoiron/sqlx"
)

type DailySettlementRepository struct {
	db *sqlx.DB
}

func NewDailySettlementRepository(db *sqlx.DB) *DailySettlementRepository {
	return &DailySettlementRepository{db: db}
}

func (r *DailySettlementRepository) Create(ctx context.Context, s *model.DailySettlement) error {
	query := `
		INSERT INTO daily_settlements (
			company_id, branch_id, settlement_date, status,
			total_sales, total_refunds, total_expenses, notes, recorded_by
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at, updated_at
	`
	return r.db.QueryRowContext(ctx, query,
		s.CompanyID,
		s.BranchID,
		s.SettlementDate,
		s.Status,
		s.TotalSales,
		s.TotalRefunds,
		s.TotalExpenses,
		s.Notes,
		s.RecordedBy,
	).Scan(&s.ID, &s.CreatedAt, &s.UpdatedAt)
}

func (r *DailySettlementRepository) GetByID(ctx context.Context, companyID, id int64) (*model.DailySettlement, error) {
	var s model.DailySettlement
	query := `SELECT * FROM daily_settlements WHERE id = $1 AND company_id = $2`
	err := r.db.GetContext(ctx, &s, query, id, companyID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *DailySettlementRepository) GetByBranchAndDate(ctx context.Context, companyID, branchID int64, date string) (*model.DailySettlement, error) {
	var s model.DailySettlement
	query := `SELECT * FROM daily_settlements WHERE company_id = $1 AND branch_id = $2 AND settlement_date = $3`
	err := r.db.GetContext(ctx, &s, query, companyID, branchID, date)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *DailySettlementRepository) List(ctx context.Context, companyID int64, params *repository.DailySettlementListParams) ([]*model.DailySettlement, int, error) {
	var settlements []*model.DailySettlement
	var total int

	where := "WHERE company_id = $1"
	args := []interface{}{companyID}
	argIdx := 2

	if params.Status != "" {
		where += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, params.Status)
		argIdx++
	}
	if params.BranchID != nil {
		where += fmt.Sprintf(" AND branch_id = $%d", argIdx)
		args = append(args, *params.BranchID)
		argIdx++
	}
	if params.DateFrom != "" {
		where += fmt.Sprintf(" AND settlement_date >= $%d", argIdx)
		args = append(args, params.DateFrom)
		argIdx++
	}
	if params.DateTo != "" {
		where += fmt.Sprintf(" AND settlement_date <= $%d", argIdx)
		args = append(args, params.DateTo)
		argIdx++
	}

	countQuery := "SELECT COUNT(*) FROM daily_settlements " + where
	if err := r.db.GetContext(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, err
	}

	listQuery := fmt.Sprintf(`
		SELECT * FROM daily_settlements %s
		ORDER BY settlement_date DESC, id DESC
		LIMIT $%d OFFSET $%d
	`, where, argIdx, argIdx+1)
	args = append(args, params.Limit, params.Offset)

	if err := r.db.SelectContext(ctx, &settlements, listQuery, args...); err != nil {
		return nil, 0, err
	}
	return settlements, total, nil
}

func (r *DailySettlementRepository) Update(ctx context.Context, s *model.DailySettlement) error {
	query := `
		UPDATE daily_settlements
		SET status = $1,
		    total_sales = $2,
		    total_refunds = $3,
		    total_expenses = $4,
		    notes = $5,
		    finalized_by = $6,
		    finalized_at = $7,
		    updated_at = NOW()
		WHERE id = $8 AND company_id = $9
	`
	result, err := r.db.ExecContext(ctx, query,
		s.Status,
		s.TotalSales,
		s.TotalRefunds,
		s.TotalExpenses,
		s.Notes,
		s.FinalizedBy,
		s.FinalizedAt,
		s.ID,
		s.CompanyID,
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

func (r *DailySettlementRepository) CreateItem(ctx context.Context, item *model.DailySettlementItem) error {
	query := `
		INSERT INTO daily_settlement_items (
			daily_settlement_id, payment_method, expected_amount, actual_amount, variance_amount, notes
		)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at
	`
	return r.db.QueryRowContext(ctx, query,
		item.DailySettlementID,
		item.PaymentMethod,
		item.ExpectedAmount,
		item.ActualAmount,
		item.VarianceAmount,
		item.Notes,
	).Scan(&item.ID, &item.CreatedAt, &item.UpdatedAt)
}

func (r *DailySettlementRepository) GetItems(ctx context.Context, settlementID int64) ([]*model.DailySettlementItem, error) {
	var items []*model.DailySettlementItem
	query := `SELECT * FROM daily_settlement_items WHERE daily_settlement_id = $1 ORDER BY id`
	if err := r.db.SelectContext(ctx, &items, query, settlementID); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *DailySettlementRepository) GetItem(ctx context.Context, settlementID, itemID int64) (*model.DailySettlementItem, error) {
	var item model.DailySettlementItem
	query := `SELECT * FROM daily_settlement_items WHERE id = $1 AND daily_settlement_id = $2`
	err := r.db.GetContext(ctx, &item, query, itemID, settlementID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *DailySettlementRepository) UpdateItem(ctx context.Context, item *model.DailySettlementItem) error {
	query := `
		UPDATE daily_settlement_items
		SET actual_amount = $1,
		    variance_amount = $2,
		    notes = $3,
		    updated_at = NOW()
		WHERE id = $4
	`
	_, err := r.db.ExecContext(ctx, query,
		item.ActualAmount,
		item.VarianceAmount,
		item.Notes,
		item.ID,
	)
	return err
}

// GetPaymentBreakdown aggregates gross sales and refunds per payment method
// for a given company + branch + date. The date is matched against payments.paid_at
// (sales) and payments.refunded_at (refunds) so a refund issued the next day
// shows up on the refund date, not the original sale date.
func (r *DailySettlementRepository) GetPaymentBreakdown(ctx context.Context, companyID, branchID int64, date string) ([]*repository.PaymentMethodTotals, error) {
	query := `
		SELECT
			method,
			COALESCE(SUM(gross_sales), 0)  AS gross_sales,
			COALESCE(SUM(refunds), 0)      AS refunds
		FROM (
			SELECT
				p.method,
				p.amount AS gross_sales,
				0::numeric AS refunds
			FROM payments p
			JOIN orders o ON o.id = p.order_id
			WHERE o.company_id = $1
			  AND o.branch_id = $2
			  AND p.status != 'failed'
			  AND p.paid_at::date = $3::date

			UNION ALL

			SELECT
				p.method,
				0::numeric AS gross_sales,
				p.refunded_amount AS refunds
			FROM payments p
			JOIN orders o ON o.id = p.order_id
			WHERE o.company_id = $1
			  AND o.branch_id = $2
			  AND p.refunded_at IS NOT NULL
			  AND p.refunded_at::date = $3::date
		) t
		GROUP BY method
		ORDER BY method
	`
	var totals []*repository.PaymentMethodTotals
	if err := r.db.SelectContext(ctx, &totals, query, companyID, branchID, date); err != nil {
		return nil, err
	}
	return totals, nil
}

func (r *DailySettlementRepository) GetExpensesTotal(ctx context.Context, companyID, branchID int64, date string) (float64, error) {
	var total float64
	query := `
		SELECT COALESCE(SUM(amount), 0)
		FROM expenses
		WHERE company_id = $1
		  AND branch_id = $2
		  AND expense_date = $3
	`
	if err := r.db.GetContext(ctx, &total, query, companyID, branchID, date); err != nil {
		return 0, err
	}
	return total, nil
}

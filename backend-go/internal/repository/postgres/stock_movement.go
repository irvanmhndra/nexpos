package postgres

import (
	"context"
	"time"

	"github.com/irvanmhndra/nexpos-api/internal/model"
	"github.com/irvanmhndra/nexpos-api/internal/repository"
	"github.com/jmoiron/sqlx"
)

type StockMovementRepository struct {
	db *sqlx.DB
}

func NewStockMovementRepository(db *sqlx.DB) *StockMovementRepository {
	return &StockMovementRepository{db: db}
}

func (r *StockMovementRepository) Create(ctx context.Context, m *model.StockMovement) error {
	query := `
		INSERT INTO stock_movements (product_variant_id, branch_id, type, quantity, stock_before, stock_after, unit_cost, reference_type, reference_id, note, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, created_at
	`
	return conn(ctx, r.db).QueryRowContext(ctx, query,
		m.ProductVariantID,
		m.BranchID,
		m.Type,
		m.Quantity,
		m.StockBefore,
		m.StockAfter,
		m.UnitCost,
		m.ReferenceType,
		m.ReferenceID,
		m.Note,
		m.CreatedBy,
	).Scan(&m.ID, &m.CreatedAt)
}

func (r *StockMovementRepository) List(ctx context.Context, companyID, branchID int64, movType, search, startDate, endDate string, limit, offset int) ([]*repository.MovementRow, int, error) {
	args := []interface{}{companyID}
	where := `p.company_id = $1`
	argIdx := 2

	if branchID > 0 {
		where += ` AND sm.branch_id = $` + itoa(argIdx)
		args = append(args, branchID)
		argIdx++
	}

	if movType != "" {
		where += ` AND sm.type = $` + itoa(argIdx)
		args = append(args, movType)
		argIdx++
	}

	if search != "" {
		where += ` AND (p.name ILIKE $` + itoa(argIdx) + ` OR pv.sku ILIKE $` + itoa(argIdx) + `)`
		args = append(args, "%"+search+"%")
		argIdx++
	}

	if startDate != "" {
		where += ` AND sm.created_at >= $` + itoa(argIdx)
		args = append(args, startDate)
		argIdx++
	}

	if endDate != "" {
		where += ` AND sm.created_at < $` + itoa(argIdx)
		args = append(args, endDate)
		argIdx++
	}

	countQuery := `
		SELECT COUNT(*)
		FROM stock_movements sm
		JOIN product_variants pv ON pv.id = sm.product_variant_id
		JOIN products p ON p.id = pv.product_id
		WHERE ` + where

	var total int
	if err := conn(ctx, r.db).GetContext(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, err
	}

	query := `
		SELECT
			sm.id,
			sm.product_variant_id,
			p.name              AS product_name,
			pv.name             AS variant_name,
			pv.sku,
			sm.branch_id,
			b.name              AS branch_name,
			sm.type,
			sm.quantity,
			sm.stock_before,
			sm.stock_after,
			sm.unit_cost,
			sm.reference_type,
			sm.reference_id,
			sm.note,
			sm.created_by,
			COALESCE(u.name, '') AS created_by_name,
			sm.created_at
		FROM stock_movements sm
		JOIN product_variants pv ON pv.id = sm.product_variant_id
		JOIN products p ON p.id = pv.product_id
		JOIN branches b ON b.id = sm.branch_id
		LEFT JOIN users u ON u.id = sm.created_by
		WHERE ` + where + `
		ORDER BY sm.created_at DESC
		LIMIT $` + itoa(argIdx) + ` OFFSET $` + itoa(argIdx+1)

	args = append(args, limit, offset)

	rows, err := conn(ctx, r.db).QueryxContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = rows.Close() }()

	var result []*repository.MovementRow
	for rows.Next() {
		var row repository.MovementRow
		if err := rows.StructScan(&row); err != nil {
			return nil, 0, err
		}
		result = append(result, &row)
	}
	return result, total, rows.Err()
}

func (r *StockMovementRepository) GetMonthlyStats(ctx context.Context, companyID, branchID int64) (*repository.MovementStats, error) {
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)

	args := []interface{}{companyID, startOfMonth}
	where := `p.company_id = $1 AND sm.created_at >= $2`
	if branchID > 0 {
		where += ` AND sm.branch_id = $3`
		args = append(args, branchID)
	}

	query := `
		SELECT
			COUNT(*)                                              AS total_movements,
			COUNT(*) FILTER (WHERE sm.type = 'IN')               AS stock_in,
			COUNT(*) FILTER (WHERE sm.type = 'OUT')              AS stock_out,
			COUNT(*) FILTER (WHERE sm.type = 'ADJUST')           AS adjustments
		FROM stock_movements sm
		JOIN product_variants pv ON pv.id = sm.product_variant_id
		JOIN products p ON p.id = pv.product_id
		WHERE ` + where

	var stats repository.MovementStats
	if err := conn(ctx, r.db).GetContext(ctx, &stats, query, args...); err != nil {
		return nil, err
	}
	return &stats, nil
}

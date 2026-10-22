package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/irvanmhndra/nexpos-api/internal/model"
	"github.com/irvanmhndra/nexpos-api/internal/repository"
	"github.com/jmoiron/sqlx"
)

type StockOpnameRepository struct {
	db *sqlx.DB
}

func NewStockOpnameRepository(db *sqlx.DB) *StockOpnameRepository {
	return &StockOpnameRepository{db: db}
}

func (r *StockOpnameRepository) Create(ctx context.Context, opname *model.StockOpname) error {
	query := `
		INSERT INTO stock_opnames (company_id, branch_id, opname_number, status, notes, started_by, started_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at
	`
	return conn(ctx, r.db).QueryRowContext(ctx, query,
		opname.CompanyID,
		opname.BranchID,
		opname.OpnameNumber,
		opname.Status,
		opname.Notes,
		opname.StartedBy,
		opname.StartedAt,
	).Scan(&opname.ID, &opname.CreatedAt, &opname.UpdatedAt)
}

func (r *StockOpnameRepository) GetByID(ctx context.Context, companyID, id int64) (*model.StockOpname, error) {
	var opname model.StockOpname
	query := `SELECT * FROM stock_opnames WHERE id = $1 AND company_id = $2`
	err := conn(ctx, r.db).GetContext(ctx, &opname, query, id, companyID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &opname, nil
}

func (r *StockOpnameRepository) GetByIDForUpdate(ctx context.Context, companyID, id int64) (*model.StockOpname, error) {
	if err := requireTx(ctx); err != nil {
		return nil, err
	}
	var opname model.StockOpname
	query := `SELECT * FROM stock_opnames WHERE id = $1 AND company_id = $2 FOR UPDATE`
	err := conn(ctx, r.db).GetContext(ctx, &opname, query, id, companyID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &opname, nil
}

func (r *StockOpnameRepository) GenerateOpnameNumber(ctx context.Context, companyID int64) (string, error) {
	now := time.Now()
	prefix := fmt.Sprintf("OPN-%d%02d%02d-", now.Year(), now.Month(), now.Day())

	var count int
	query := `SELECT COUNT(*) FROM stock_opnames WHERE company_id = $1 AND opname_number LIKE $2`
	if err := conn(ctx, r.db).GetContext(ctx, &count, query, companyID, prefix+"%"); err != nil {
		return "", err
	}
	return fmt.Sprintf("%s%04d", prefix, count+1), nil
}

func (r *StockOpnameRepository) List(ctx context.Context, companyID int64, params *repository.StockOpnameListParams) ([]*model.StockOpname, int, error) {
	var opnames []*model.StockOpname
	var total int

	whereClause := "WHERE company_id = $1"
	args := []interface{}{companyID}
	argIndex := 2

	if params.Status != "" {
		whereClause += fmt.Sprintf(" AND status = $%d", argIndex)
		args = append(args, params.Status)
		argIndex++
	}

	if params.BranchID != nil {
		whereClause += fmt.Sprintf(" AND branch_id = $%d", argIndex)
		args = append(args, *params.BranchID)
		argIndex++
	}

	countQuery := "SELECT COUNT(*) FROM stock_opnames " + whereClause
	if err := conn(ctx, r.db).GetContext(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, err
	}

	listQuery := fmt.Sprintf(`
		SELECT * FROM stock_opnames %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIndex, argIndex+1)
	args = append(args, params.Limit, params.Offset)

	if err := conn(ctx, r.db).SelectContext(ctx, &opnames, listQuery, args...); err != nil {
		return nil, 0, err
	}
	return opnames, total, nil
}

func (r *StockOpnameRepository) Update(ctx context.Context, opname *model.StockOpname) error {
	query := `
		UPDATE stock_opnames
		SET status = $1,
		    notes = $2,
		    total_variance_qty = $3,
		    total_variance_value = $4,
		    completed_by = $5,
		    completed_at = $6,
		    cancelled_at = $7,
		    updated_at = NOW()
		WHERE id = $8 AND company_id = $9
	`
	result, err := conn(ctx, r.db).ExecContext(ctx, query,
		opname.Status,
		opname.Notes,
		opname.TotalVarianceQty,
		opname.TotalVarianceValue,
		opname.CompletedBy,
		opname.CompletedAt,
		opname.CancelledAt,
		opname.ID,
		opname.CompanyID,
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

// SnapshotItems inserts one stock_opname_items row for each active variant of the
// given branch, capturing current system stock and unit cost at snapshot time.
// Returns the number of items snapshotted.
func (r *StockOpnameRepository) SnapshotItems(ctx context.Context, opnameID, companyID, branchID int64, categoryID *int64) (int, error) {
	args := []interface{}{opnameID, companyID, branchID}
	categoryFilter := ""
	if categoryID != nil {
		categoryFilter = " AND p.product_category_id = $4"
		args = append(args, *categoryID)
	}

	query := `
		INSERT INTO stock_opname_items (
			stock_opname_id, product_variant_id, sku, product_name, variant_name,
			system_stock, unit_cost
		)
		SELECT
			$1,
			pv.id,
			pv.sku,
			p.name,
			pv.name,
			COALESCE(s.quantity, 0),
			COALESCE(NULLIF(pv.last_purchase_cost, 0), pv.standard_cost)
		FROM product_variants pv
		JOIN products p ON p.id = pv.product_id
		LEFT JOIN stocks s ON s.product_variant_id = pv.id AND s.branch_id = $3
		WHERE p.company_id = $2
		  AND p.is_active = true
		  AND pv.is_active = true` + categoryFilter

	result, err := conn(ctx, r.db).ExecContext(ctx, query, args...)
	if err != nil {
		return 0, err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}
	return int(n), nil
}

func (r *StockOpnameRepository) GetItems(ctx context.Context, opnameID int64) ([]*model.StockOpnameItem, error) {
	var items []*model.StockOpnameItem
	query := `SELECT * FROM stock_opname_items WHERE stock_opname_id = $1 ORDER BY product_name, variant_name, id`
	if err := conn(ctx, r.db).SelectContext(ctx, &items, query, opnameID); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *StockOpnameRepository) GetItem(ctx context.Context, opnameID, itemID int64) (*model.StockOpnameItem, error) {
	var item model.StockOpnameItem
	query := `SELECT * FROM stock_opname_items WHERE id = $1 AND stock_opname_id = $2`
	err := conn(ctx, r.db).GetContext(ctx, &item, query, itemID, opnameID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *StockOpnameRepository) UpdateItem(ctx context.Context, item *model.StockOpnameItem) error {
	query := `
		UPDATE stock_opname_items
		SET counted_stock = $1,
		    variance_qty = $2,
		    variance_value = $3,
		    notes = $4,
		    counted_by = $5,
		    counted_at = $6,
		    updated_at = NOW()
		WHERE id = $7
	`
	_, err := conn(ctx, r.db).ExecContext(ctx, query,
		item.CountedStock,
		item.VarianceQty,
		item.VarianceValue,
		item.Notes,
		item.CountedBy,
		item.CountedAt,
		item.ID,
	)
	return err
}

func (r *StockOpnameRepository) GetItemStats(ctx context.Context, opnameID int64) (int, int, error) {
	var stats struct {
		Total   int `db:"total"`
		Counted int `db:"counted"`
	}
	query := `
		SELECT
			COUNT(*) AS total,
			COUNT(*) FILTER (WHERE counted_stock IS NOT NULL) AS counted
		FROM stock_opname_items
		WHERE stock_opname_id = $1
	`
	if err := conn(ctx, r.db).GetContext(ctx, &stats, query, opnameID); err != nil {
		return 0, 0, err
	}
	return stats.Total, stats.Counted, nil
}

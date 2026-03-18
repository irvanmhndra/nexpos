package postgres

import (
	"context"
	"database/sql"
	"strconv"

	"github.com/irvanmhndra/nexpos-api/internal/model"
	"github.com/irvanmhndra/nexpos-api/internal/repository"
	"github.com/jmoiron/sqlx"
)

type StockRepository struct {
	db *sqlx.DB
}

func NewStockRepository(db *sqlx.DB) *StockRepository {
	return &StockRepository{db: db}
}

func (r *StockRepository) GetByVariantAndBranch(ctx context.Context, variantID, branchID int64) (*model.Stock, error) {
	var stock model.Stock
	query := `SELECT * FROM stocks WHERE product_variant_id = $1 AND branch_id = $2`
	err := r.db.GetContext(ctx, &stock, query, variantID, branchID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &stock, nil
}

func (r *StockRepository) Upsert(ctx context.Context, stock *model.Stock) error {
	query := `
		INSERT INTO stocks (product_variant_id, branch_id, quantity, min_quantity, updated_at)
		VALUES ($1, $2, $3, $4, NOW())
		ON CONFLICT (product_variant_id, branch_id)
		DO UPDATE SET quantity = $3, updated_at = NOW()
		RETURNING id, updated_at
	`
	return r.db.QueryRowContext(ctx, query,
		stock.ProductVariantID,
		stock.BranchID,
		stock.Quantity,
		stock.MinQuantity,
	).Scan(&stock.ID, &stock.UpdatedAt)
}

func (r *StockRepository) UpdateMinQuantity(ctx context.Context, variantID, branchID int64, minQuantity int) error {
	query := `
		INSERT INTO stocks (product_variant_id, branch_id, quantity, min_quantity, updated_at)
		VALUES ($1, $2, 0, $3, NOW())
		ON CONFLICT (product_variant_id, branch_id)
		DO UPDATE SET min_quantity = $3, updated_at = NOW()
	`
	_, err := r.db.ExecContext(ctx, query, variantID, branchID, minQuantity)
	return err
}

func (r *StockRepository) ListInventory(ctx context.Context, companyID, branchID int64, search, category, status string, limit, offset int) ([]*repository.InventoryRow, int, error) {
	args := []interface{}{companyID}
	where := `p.company_id = $1 AND pv.is_active = true AND p.is_active = true`
	argIdx := 2

	// Use branchID directly in JOIN conditions (not in WHERE) so LEFT JOIN preserves
	// products without stock records (they show with quantity = 0).
	var stockJoin, branchJoin string
	if branchID > 0 {
		stockJoin = `LEFT JOIN stocks s ON s.product_variant_id = pv.id AND s.branch_id = $` + itoa(argIdx)
		branchJoin = `LEFT JOIN branches b ON b.id = $` + itoa(argIdx)
		args = append(args, branchID)
		argIdx++
	} else {
		// No branch filter: show all variants, no stock data
		stockJoin = `LEFT JOIN stocks s ON false`
		branchJoin = `LEFT JOIN branches b ON false`
	}

	if search != "" {
		where += ` AND (p.name ILIKE $` + itoa(argIdx) + ` OR pv.sku ILIKE $` + itoa(argIdx) + `)`
		args = append(args, "%"+search+"%")
		argIdx++
	}

	if category != "" {
		where += ` AND pc.name = $` + itoa(argIdx)
		args = append(args, category)
		argIdx++
	}

	var statusFilter string
	switch status {
	case "out_of_stock":
		statusFilter = ` AND COALESCE(s.quantity, 0) <= 0`
	case "low":
		statusFilter = ` AND COALESCE(s.quantity, 0) > 0 AND COALESCE(s.quantity, 0) <= COALESCE(s.min_quantity, 0)`
	case "normal":
		statusFilter = ` AND COALESCE(s.quantity, 0) > COALESCE(s.min_quantity, 0)`
	}

	countQuery := `
		SELECT COUNT(*)
		FROM product_variants pv
		JOIN products p ON p.id = pv.product_id
		LEFT JOIN product_categories pc ON pc.id = p.product_category_id
		` + stockJoin + `
		` + branchJoin + `
		WHERE ` + where + statusFilter

	var total int
	if err := r.db.GetContext(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, err
	}

	query := `
		SELECT
			pv.id                            AS product_variant_id,
			p.id                             AS product_id,
			p.name                           AS product_name,
			pv.name                          AS variant_name,
			pv.sku,
			COALESCE(pc.name, '')            AS category_name,
			p.image_data,
			COALESCE(b.id, 0)                AS branch_id,
			COALESCE(b.name, '')             AS branch_name,
			COALESCE(s.quantity, 0)          AS current_stock,
			COALESCE(s.min_quantity, 0)      AS min_stock,
			GREATEST(COALESCE(s.quantity, 0), 0) * pv.standard_cost AS stock_value,
			(SELECT COUNT(*) FROM product_variants pv2 WHERE pv2.product_id = p.id AND pv2.is_active = true) > 1 AS has_variants
		FROM product_variants pv
		JOIN products p ON p.id = pv.product_id
		LEFT JOIN product_categories pc ON pc.id = p.product_category_id
		` + stockJoin + `
		` + branchJoin + `
		WHERE ` + where + statusFilter + `
		ORDER BY p.name, pv.name
		LIMIT $` + itoa(argIdx) + ` OFFSET $` + itoa(argIdx+1)

	args = append(args, limit, offset)

	rows, err := r.db.QueryxContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var result []*repository.InventoryRow
	for rows.Next() {
		var row repository.InventoryRow
		if err := rows.StructScan(&row); err != nil {
			return nil, 0, err
		}
		result = append(result, &row)
	}
	return result, total, rows.Err()
}

func (r *StockRepository) GetInventoryStats(ctx context.Context, companyID, branchID int64) (*repository.InventoryStats, error) {
	args := []interface{}{companyID}
	where := `p.company_id = $1 AND pv.is_active = true AND p.is_active = true`

	stockJoin := `LEFT JOIN stocks s ON false`
	if branchID > 0 {
		stockJoin = `LEFT JOIN stocks s ON s.product_variant_id = pv.id AND s.branch_id = $2`
		args = append(args, branchID)
	}

	query := `
		SELECT
			COUNT(*)                                                         AS total_sku,
			COUNT(*) FILTER (WHERE COALESCE(s.quantity,0) <= 0)             AS out_of_stock,
			COUNT(*) FILTER (WHERE COALESCE(s.quantity,0) > 0
				AND COALESCE(s.quantity,0) <= COALESCE(s.min_quantity,0))   AS low_stock,
			COALESCE(SUM(GREATEST(COALESCE(s.quantity,0), 0) * pv.standard_cost), 0) AS total_stock_value
		FROM product_variants pv
		JOIN products p ON p.id = pv.product_id
		` + stockJoin + `
		WHERE ` + where

	var stats repository.InventoryStats
	if err := r.db.GetContext(ctx, &stats, query, args...); err != nil {
		return nil, err
	}
	return &stats, nil
}

func itoa(i int) string {
	return strconv.Itoa(i)
}

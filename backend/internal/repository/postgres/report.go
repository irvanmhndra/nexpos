package postgres

import (
	"context"

	"github.com/irvanmhndra/nexpos-api/internal/repository"
	"github.com/jmoiron/sqlx"
)

type ReportRepository struct {
	db *sqlx.DB
}

func NewReportRepository(db *sqlx.DB) *ReportRepository {
	return &ReportRepository{db: db}
}

func (r *ReportRepository) GetSummary(ctx context.Context, companyID int64, dateFrom, dateTo string) (*repository.ReportSummary, error) {
	summary := &repository.ReportSummary{}

	// Get order stats
	orderQuery := `
		SELECT
			COALESCE(SUM(grand_total), 0) as total_revenue,
			COUNT(*) as total_orders,
			COALESCE(SUM(total_discount), 0) as total_discount,
			COALESCE(SUM(total_tax), 0) as total_tax,
			COUNT(*) FILTER (WHERE status = 'completed') as completed_orders,
			COUNT(*) FILTER (WHERE status = 'cancelled') as cancelled_orders,
			COUNT(*) FILTER (WHERE status = 'pending') as pending_orders,
			COUNT(DISTINCT customer_id) as total_customers
		FROM orders
		WHERE company_id = $1
		  AND created_at >= $2::date
		  AND created_at < ($3::date + interval '1 day')
	`
	err := r.db.QueryRowContext(ctx, orderQuery, companyID, dateFrom, dateTo).Scan(
		&summary.TotalRevenue,
		&summary.TotalOrders,
		&summary.TotalDiscount,
		&summary.TotalTax,
		&summary.CompletedOrders,
		&summary.CancelledOrders,
		&summary.PendingOrders,
		&summary.TotalCustomers,
	)
	if err != nil {
		return nil, err
	}

	// Get items sold and COGS
	itemsQuery := `
		SELECT
			COALESCE(SUM(oi.quantity), 0) as total_items,
			COALESCE(SUM(oi.cogs_amount), 0) as total_cogs
		FROM order_items oi
		JOIN orders o ON o.id = oi.order_id
		WHERE o.company_id = $1
		  AND o.created_at >= $2::date
		  AND o.created_at < ($3::date + interval '1 day')
		  AND o.status = 'completed'
	`
	err = r.db.QueryRowContext(ctx, itemsQuery, companyID, dateFrom, dateTo).Scan(
		&summary.TotalItemsSold,
		&summary.TotalCOGS,
	)
	if err != nil {
		return nil, err
	}

	return summary, nil
}

func (r *ReportRepository) GetSalesTrend(ctx context.Context, companyID int64, dateFrom, dateTo string) ([]*repository.SalesTrendItem, error) {
	query := `
		SELECT
			DATE(created_at) as date,
			COALESCE(SUM(grand_total), 0) as sales,
			COUNT(*) as orders,
			COALESCE(SUM((SELECT SUM(quantity) FROM order_items WHERE order_id = orders.id)), 0) as items
		FROM orders
		WHERE company_id = $1
		  AND created_at >= $2::date
		  AND created_at < ($3::date + interval '1 day')
		  AND status = 'completed'
		GROUP BY DATE(created_at)
		ORDER BY date ASC
	`

	rows, err := r.db.QueryContext(ctx, query, companyID, dateFrom, dateTo)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var items []*repository.SalesTrendItem
	for rows.Next() {
		item := &repository.SalesTrendItem{}
		if err := rows.Scan(&item.Date, &item.Sales, &item.Orders, &item.Items); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, nil
}

func (r *ReportRepository) GetTopProducts(ctx context.Context, companyID int64, dateFrom, dateTo string, limit int) ([]*repository.TopProductItem, error) {
	query := `
		SELECT
			COALESCE(oi.product_id, 0) as product_id,
			oi.product_name,
			SUM(oi.quantity) as total_sold,
			SUM(oi.subtotal) as total_amount
		FROM order_items oi
		JOIN orders o ON o.id = oi.order_id
		WHERE o.company_id = $1
		  AND o.created_at >= $2::date
		  AND o.created_at < ($3::date + interval '1 day')
		  AND o.status = 'completed'
		GROUP BY oi.product_id, oi.product_name
		ORDER BY total_sold DESC
		LIMIT $4
	`

	rows, err := r.db.QueryContext(ctx, query, companyID, dateFrom, dateTo, limit)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var items []*repository.TopProductItem
	for rows.Next() {
		item := &repository.TopProductItem{}
		if err := rows.Scan(&item.ProductID, &item.ProductName, &item.TotalSold, &item.TotalAmount); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, nil
}

func (r *ReportRepository) GetCategoryRevenue(ctx context.Context, companyID int64, dateFrom, dateTo string) ([]*repository.CategoryRevenueItem, error) {
	query := `
		SELECT
			COALESCE(p.product_category_id, 0) as category_id,
			COALESCE(pc.name, 'Tanpa Kategori') as category_name,
			SUM(oi.subtotal) as total_amount,
			COUNT(DISTINCT o.id) as order_count
		FROM order_items oi
		JOIN orders o ON o.id = oi.order_id
		LEFT JOIN products p ON p.id = oi.product_id
		LEFT JOIN product_categories pc ON pc.id = p.product_category_id
		WHERE o.company_id = $1
		  AND o.created_at >= $2::date
		  AND o.created_at < ($3::date + interval '1 day')
		  AND o.status = 'completed'
		GROUP BY p.product_category_id, pc.name
		ORDER BY total_amount DESC
	`

	rows, err := r.db.QueryContext(ctx, query, companyID, dateFrom, dateTo)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var items []*repository.CategoryRevenueItem
	for rows.Next() {
		item := &repository.CategoryRevenueItem{}
		if err := rows.Scan(&item.CategoryID, &item.CategoryName, &item.TotalAmount, &item.OrderCount); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, nil
}

func (r *ReportRepository) GetPaymentMethods(ctx context.Context, companyID int64, dateFrom, dateTo string) ([]*repository.PaymentMethodItem, error) {
	query := `
		SELECT
			p.method,
			SUM(p.amount) as total_amount,
			COUNT(*) as count
		FROM payments p
		JOIN orders o ON o.id = p.order_id
		WHERE o.company_id = $1
		  AND o.created_at >= $2::date
		  AND o.created_at < ($3::date + interval '1 day')
		  AND o.status = 'completed'
		GROUP BY p.method
		ORDER BY total_amount DESC
	`

	rows, err := r.db.QueryContext(ctx, query, companyID, dateFrom, dateTo)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var items []*repository.PaymentMethodItem
	for rows.Next() {
		item := &repository.PaymentMethodItem{}
		if err := rows.Scan(&item.Method, &item.TotalAmount, &item.Count); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, nil
}

func (r *ReportRepository) GetHourlySales(ctx context.Context, companyID int64, dateFrom, dateTo string) ([]*repository.HourlySalesItem, error) {
	query := `
		SELECT
			EXTRACT(HOUR FROM created_at)::int as hour,
			SUM(grand_total) as total_amount,
			COUNT(*) as order_count
		FROM orders
		WHERE company_id = $1
		  AND created_at >= $2::date
		  AND created_at < ($3::date + interval '1 day')
		  AND status = 'completed'
		GROUP BY EXTRACT(HOUR FROM created_at)
		ORDER BY hour ASC
	`

	rows, err := r.db.QueryContext(ctx, query, companyID, dateFrom, dateTo)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var items []*repository.HourlySalesItem
	for rows.Next() {
		item := &repository.HourlySalesItem{}
		if err := rows.Scan(&item.Hour, &item.TotalAmount, &item.OrderCount); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, nil
}

func (r *ReportRepository) GetNewCustomersCount(ctx context.Context, companyID int64, dateFrom, dateTo string) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM customers
		WHERE company_id = $1
		  AND created_at >= $2::date
		  AND created_at < ($3::date + interval '1 day')
	`

	var count int
	err := r.db.QueryRowContext(ctx, query, companyID, dateFrom, dateTo).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

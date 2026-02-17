package dto

import "time"

// ============== Request DTOs ==============

type ReportRequest struct {
	DateFrom string `query:"date_from" validate:"required"`
	DateTo   string `query:"date_to" validate:"required"`
}

// ============== Response DTOs ==============

type ReportSummaryResponse struct {
	TotalRevenue      float64 `json:"total_revenue"`
	TotalOrders       int     `json:"total_orders"`
	TotalItemsSold    int     `json:"total_items_sold"`
	AvgOrderValue     float64 `json:"avg_order_value"`
	TotalCustomers    int     `json:"total_customers"`
	NewCustomers      int     `json:"new_customers"`
	CompletedOrders   int     `json:"completed_orders"`
	CancelledOrders   int     `json:"cancelled_orders"`
	PendingOrders     int     `json:"pending_orders"`
	TotalDiscount     float64 `json:"total_discount"`
	TotalTax          float64 `json:"total_tax"`
	GrossProfit       float64 `json:"gross_profit"`
	GrossProfitMargin float64 `json:"gross_profit_margin"`
}

type SalesTrendItem struct {
	Date   string  `json:"date"`
	Sales  float64 `json:"sales"`
	Orders int     `json:"orders"`
	Items  int     `json:"items"`
}

type SalesTrendResponse struct {
	Data []*SalesTrendItem `json:"data"`
}

type TopProductItem struct {
	ProductID   int64   `json:"product_id"`
	ProductName string  `json:"product_name"`
	TotalSold   int     `json:"total_sold"`
	TotalAmount float64 `json:"total_amount"`
}

type TopProductsResponse struct {
	Data []*TopProductItem `json:"data"`
}

type CategoryRevenueItem struct {
	CategoryID   int64   `json:"category_id"`
	CategoryName string  `json:"category_name"`
	TotalAmount  float64 `json:"total_amount"`
	OrderCount   int     `json:"order_count"`
}

type CategoryRevenueResponse struct {
	Data []*CategoryRevenueItem `json:"data"`
}

type PaymentMethodItem struct {
	Method      string  `json:"method"`
	TotalAmount float64 `json:"total_amount"`
	Count       int     `json:"count"`
}

type PaymentMethodResponse struct {
	Data []*PaymentMethodItem `json:"data"`
}

type HourlySalesItem struct {
	Hour        int     `json:"hour"`
	TotalAmount float64 `json:"total_amount"`
	OrderCount  int     `json:"order_count"`
}

type HourlySalesResponse struct {
	Data []*HourlySalesItem `json:"data"`
}

// Comparison with previous period
type ReportComparisonResponse struct {
	Current          *ReportSummaryResponse `json:"current"`
	Previous         *ReportSummaryResponse `json:"previous"`
	RevenueChange    float64                `json:"revenue_change"`
	OrdersChange     float64                `json:"orders_change"`
	AvgOrderChange   float64                `json:"avg_order_change"`
	ItemsSoldChange  float64                `json:"items_sold_change"`
	CustomerChange   float64                `json:"customer_change"`
}

// Date range helper
type DateRange struct {
	From time.Time
	To   time.Time
}

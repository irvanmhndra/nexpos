package router

import (
	"github.com/irvanmhndra/nexpos-api/internal/handler"
	"github.com/labstack/echo/v5"
)

type Handlers struct {
	Health          *handler.HealthHandler
	Auth            *handler.AuthHandler
	User            *handler.UserHandler
	Branch          *handler.BranchHandler
	Customer        *handler.CustomerHandler
	ProductCategory *handler.ProductCategoryHandler
	Product         *handler.ProductHandler
	Order           *handler.OrderHandler
	Report          *handler.ReportHandler
	Promotion       *handler.PromotionHandler
	Inventory       *handler.InventoryHandler
	CompanySettings *handler.CompanySettingsHandler
	Supplier        *handler.SupplierHandler
	PurchaseOrder   *handler.PurchaseOrderHandler
	Shift           *handler.ShiftHandler
	ExpenseCategory *handler.ExpenseCategoryHandler
	Expense         *handler.ExpenseHandler
	StockOpname     *handler.StockOpnameHandler
	DailySettlement *handler.DailySettlementHandler
	Receipt         *handler.ReceiptHandler
}

func Setup(e *echo.Echo, h *Handlers, authMW echo.MiddlewareFunc) {
	// Health check endpoints (public)
	e.GET("/health", h.Health.Check)
	e.GET("/health/live", h.Health.Liveness)
	e.GET("/health/ready", h.Health.Readiness)

	// API group
	api := e.Group("/api")
	v1 := api.Group("/v1")

	// Auth routes (public)
	auth := v1.Group("/auth")
	auth.POST("/login", h.Auth.Login)
	auth.POST("/register", h.Auth.Register)
	auth.POST("/refresh", h.Auth.RefreshToken)
	auth.POST("/logout", h.Auth.Logout)

	// Protected routes
	protected := v1.Group("", authMW)

	// User routes
	users := protected.Group("/users")
	users.POST("", h.User.Create)
	users.GET("", h.User.List)
	users.GET("/:id", h.User.Get)
	users.PUT("/:id", h.User.Update)
	users.PATCH("/:id/status", h.User.UpdateStatus)
	users.DELETE("/:id", h.User.Delete)

	// Branch routes
	branches := protected.Group("/branches")
	branches.POST("", h.Branch.Create)
	branches.GET("", h.Branch.List)
	branches.GET("/:id", h.Branch.Get)
	branches.PUT("/:id", h.Branch.Update)
	branches.DELETE("/:id", h.Branch.Delete)

	// Customer routes
	customers := protected.Group("/customers")
	customers.POST("", h.Customer.Create)
	customers.GET("", h.Customer.List)
	customers.GET("/:id", h.Customer.Get)
	customers.PUT("/:id", h.Customer.Update)
	customers.DELETE("/:id", h.Customer.Delete)

	// Product Category routes
	categories := protected.Group("/product-categories")
	categories.POST("", h.ProductCategory.Create)
	categories.GET("", h.ProductCategory.List)
	categories.GET("/all", h.ProductCategory.ListAll)
	categories.GET("/:id", h.ProductCategory.Get)
	categories.PUT("/:id", h.ProductCategory.Update)
	categories.DELETE("/:id", h.ProductCategory.Delete)

	// Product routes
	products := protected.Group("/products")
	products.POST("", h.Product.Create)
	products.GET("", h.Product.List)
	products.GET("/:id", h.Product.Get)
	products.PUT("/:id", h.Product.Update)
	products.DELETE("/:id", h.Product.Delete)

	// Order routes
	orders := protected.Group("/orders")
	orders.POST("", h.Order.Create)
	orders.GET("", h.Order.List)
	orders.POST("/preview", h.Order.Preview)
	orders.GET("/:id", h.Order.Get)
	orders.PUT("/:id", h.Order.Update)
	orders.POST("/:id/confirm", h.Order.Confirm)
	orders.POST("/:id/payments", h.Order.AddPayment)
	orders.POST("/:id/complete", h.Order.Complete)
	orders.POST("/:id/cancel", h.Order.Cancel)
	orders.POST("/:id/void", h.Order.Void)
	orders.POST("/:id/refund", h.Order.RefundPayment)
	orders.GET("/:id/receipt", h.Receipt.Get)

	// Report routes
	reports := protected.Group("/reports")
	reports.GET("/summary", h.Report.GetSummary)
	reports.GET("/sales-trend", h.Report.GetSalesTrend)
	reports.GET("/top-products", h.Report.GetTopProducts)
	reports.GET("/category-revenue", h.Report.GetCategoryRevenue)
	reports.GET("/payment-methods", h.Report.GetPaymentMethods)
	reports.GET("/hourly-sales", h.Report.GetHourlySales)

	// Promotion routes
	promotions := protected.Group("/promotions")
	promotions.POST("", h.Promotion.Create)
	promotions.GET("", h.Promotion.List)
	promotions.GET("/:id", h.Promotion.Get)
	promotions.PUT("/:id", h.Promotion.Update)
	promotions.DELETE("/:id", h.Promotion.Delete)

	// Inventory routes
	inventory := protected.Group("/inventory")
	inventory.GET("", h.Inventory.ListInventory)
	inventory.GET("/stats", h.Inventory.GetInventoryStats)
	inventory.POST("/adjust", h.Inventory.AdjustStock)
	inventory.PUT("/:variantId/min-stock", h.Inventory.UpdateMinStock)

	// Stock movement routes
	movements := protected.Group("/inventory/movements")
	movements.GET("", h.Inventory.ListMovements)
	movements.GET("/stats", h.Inventory.GetMovementStats)

	// Company settings routes
	protected.GET("/company-settings", h.CompanySettings.Get)
	protected.PUT("/company-settings", h.CompanySettings.Update)

	// Supplier routes
	suppliers := protected.Group("/suppliers")
	suppliers.POST("", h.Supplier.Create)
	suppliers.GET("", h.Supplier.List)
	suppliers.GET("/:id", h.Supplier.Get)
	suppliers.PUT("/:id", h.Supplier.Update)
	suppliers.DELETE("/:id", h.Supplier.Delete)

	// Purchase order routes
	purchaseOrders := protected.Group("/purchase-orders")
	purchaseOrders.POST("", h.PurchaseOrder.Create)
	purchaseOrders.GET("", h.PurchaseOrder.List)
	purchaseOrders.GET("/:id", h.PurchaseOrder.Get)
	purchaseOrders.POST("/:id/receive", h.PurchaseOrder.Receive)

	// Shift routes
	shifts := protected.Group("/shifts")
	shifts.POST("", h.Shift.Open)
	shifts.GET("", h.Shift.List)
	shifts.GET("/current", h.Shift.GetCurrent)
	shifts.GET("/:id", h.Shift.Get)
	shifts.POST("/:id/close", h.Shift.Close)

	// Expense category routes
	expenseCategories := protected.Group("/expense-categories")
	expenseCategories.POST("", h.ExpenseCategory.Create)
	expenseCategories.GET("", h.ExpenseCategory.List)
	expenseCategories.PUT("/:id", h.ExpenseCategory.Update)
	expenseCategories.DELETE("/:id", h.ExpenseCategory.Delete)

	// Expense routes
	expenses := protected.Group("/expenses")
	expenses.POST("", h.Expense.Create)
	expenses.GET("", h.Expense.List)
	expenses.GET("/summary", h.Expense.Summary)
	expenses.GET("/:id", h.Expense.Get)
	expenses.PUT("/:id", h.Expense.Update)
	expenses.DELETE("/:id", h.Expense.Delete)

	// Stock opname routes
	opnames := protected.Group("/stock-opnames")
	opnames.POST("", h.StockOpname.Create)
	opnames.GET("", h.StockOpname.List)
	opnames.GET("/:id", h.StockOpname.Get)
	opnames.PATCH("/:id/items/:itemId", h.StockOpname.UpdateItem)
	opnames.PATCH("/:id/items", h.StockOpname.BulkUpdateItems)
	opnames.POST("/:id/complete", h.StockOpname.Complete)
	opnames.POST("/:id/cancel", h.StockOpname.Cancel)

	// Daily settlement routes
	settlements := protected.Group("/daily-settlements")
	settlements.GET("/report", h.DailySettlement.Report)
	settlements.POST("", h.DailySettlement.Create)
	settlements.GET("", h.DailySettlement.List)
	settlements.GET("/:id", h.DailySettlement.Get)
	settlements.PATCH("/:id/items/:itemId", h.DailySettlement.UpdateItem)
	settlements.PATCH("/:id/items", h.DailySettlement.BulkUpdateItems)
	settlements.POST("/:id/finalize", h.DailySettlement.Finalize)
}

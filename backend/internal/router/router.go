package router

import (
	"github.com/irvanmhndra/pos-core-api/internal/handler"
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
}

func Setup(e *echo.Echo, h *Handlers) {
	// Health check endpoints
	e.GET("/", h.Health.Check)
	e.GET("/health", h.Health.Check)
	e.GET("/health/live", h.Health.Liveness)
	e.GET("/health/ready", h.Health.Readiness)

	// API group
	api := e.Group("/api")
	v1 := api.Group("/v1")

	// Auth routes
	auth := v1.Group("/auth")
	auth.POST("/login", h.Auth.Login)
	auth.POST("/register", h.Auth.Register)
	auth.POST("/refresh", h.Auth.RefreshToken)
	auth.POST("/logout", h.Auth.Logout)

	// User routes
	users := v1.Group("/users")
	users.POST("", h.User.Create)
	users.GET("", h.User.List)
	users.GET("/:id", h.User.Get)
	users.PUT("/:id", h.User.Update)
	users.PATCH("/:id/status", h.User.UpdateStatus)
	users.DELETE("/:id", h.User.Delete)

	// Branch routes
	branches := v1.Group("/branches")
	branches.POST("", h.Branch.Create)
	branches.GET("", h.Branch.List)
	branches.GET("/:id", h.Branch.Get)
	branches.PUT("/:id", h.Branch.Update)
	branches.DELETE("/:id", h.Branch.Delete)

	// Customer routes
	customers := v1.Group("/customers")
	customers.POST("", h.Customer.Create)
	customers.GET("", h.Customer.List)
	customers.GET("/:id", h.Customer.Get)
	customers.PUT("/:id", h.Customer.Update)
	customers.DELETE("/:id", h.Customer.Delete)

	// Product Category routes
	categories := v1.Group("/product-categories")
	categories.POST("", h.ProductCategory.Create)
	categories.GET("", h.ProductCategory.List)
	categories.GET("/all", h.ProductCategory.ListAll)
	categories.GET("/:id", h.ProductCategory.Get)
	categories.PUT("/:id", h.ProductCategory.Update)
	categories.DELETE("/:id", h.ProductCategory.Delete)

	// Product routes
	products := v1.Group("/products")
	products.POST("", h.Product.Create)
	products.GET("", h.Product.List)
	products.GET("/:id", h.Product.Get)
	products.PUT("/:id", h.Product.Update)
	products.DELETE("/:id", h.Product.Delete)

	// Order routes
	orders := v1.Group("/orders")
	orders.POST("", h.Order.Create)
	orders.GET("", h.Order.List)
	orders.GET("/:id", h.Order.Get)
	orders.PUT("/:id", h.Order.Update)
	orders.POST("/:id/confirm", h.Order.Confirm)
	orders.POST("/:id/payments", h.Order.AddPayment)
	orders.POST("/:id/complete", h.Order.Complete)
	orders.POST("/:id/cancel", h.Order.Cancel)
	orders.POST("/:id/void", h.Order.Void)
	orders.POST("/:id/refund", h.Order.RefundPayment)

	// Report routes
	reports := v1.Group("/reports")
	reports.GET("/summary", h.Report.GetSummary)
	reports.GET("/sales-trend", h.Report.GetSalesTrend)
	reports.GET("/top-products", h.Report.GetTopProducts)
	reports.GET("/category-revenue", h.Report.GetCategoryRevenue)
	reports.GET("/payment-methods", h.Report.GetPaymentMethods)
	reports.GET("/hourly-sales", h.Report.GetHourlySales)

	// Promotion routes
	promotions := v1.Group("/promotions")
	promotions.POST("", h.Promotion.Create)
	promotions.GET("", h.Promotion.List)
	promotions.GET("/:id", h.Promotion.Get)
	promotions.PUT("/:id", h.Promotion.Update)
	promotions.DELETE("/:id", h.Promotion.Delete)
}

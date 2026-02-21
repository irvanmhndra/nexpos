package app

import (
	"github.com/irvanmhndra/pos-core-api/internal/handler"
	"github.com/irvanmhndra/pos-core-api/pkg/validator"
	"github.com/jmoiron/sqlx"
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

func initHandlers(db *sqlx.DB, services *Services, v *validator.CustomValidator) *Handlers {
	return &Handlers{
		Health:          handler.NewHealthHandler(db),
		Auth:            handler.NewAuthHandler(services.Auth, v),
		User:            handler.NewUserHandler(services.User, v),
		Branch:          handler.NewBranchHandler(services.Branch, v),
		Customer:        handler.NewCustomerHandler(services.Customer, v),
		ProductCategory: handler.NewProductCategoryHandler(services.ProductCategory, v),
		Product:         handler.NewProductHandler(services.Product, v),
		Order:           handler.NewOrderHandler(services.Order, v),
		Report:          handler.NewReportHandler(services.Report, v),
		Promotion:       handler.NewPromotionHandler(services.Promotion, v),
	}
}

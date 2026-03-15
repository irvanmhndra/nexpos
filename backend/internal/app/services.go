package app

import (
	"github.com/irvanmhndra/nexpos-api/config"
	"github.com/irvanmhndra/nexpos-api/internal/service"
)

type Services struct {
	Auth            *service.AuthService
	User            *service.UserService
	Branch          *service.BranchService
	Customer        *service.CustomerService
	ProductCategory *service.ProductCategoryService
	Product         *service.ProductService
	Order           *service.OrderService
	Report          *service.ReportService
	Promotion       *service.PromotionService
	Inventory       *service.InventoryService
	CompanySettings *service.CompanySettingsService
}

func initServices(repos *Repositories, cfg *config.Config) *Services {
	return &Services{
		Auth: service.NewAuthService(
			repos.User,
			repos.Company,
			repos.Branch,
			repos.UserSession,
			repos.Role,
			repos.UserBranch,
			cfg.JWT,
		),
		User:            service.NewUserService(repos.User),
		Branch:          service.NewBranchService(repos.Branch),
		Customer:        service.NewCustomerService(repos.Customer),
		ProductCategory: service.NewProductCategoryService(repos.ProductCategory),
		Product:         service.NewProductService(repos.Product, repos.ProductVariant, repos.ProductCategory),
		Order: service.NewOrderService(
			repos.Order,
			repos.OrderItem,
			repos.Payment,
			repos.ProductVariant,
			repos.Customer,
			repos.User,
			repos.CompanySettings,
			repos.Stock,
			repos.StockMovement,
			repos.Promotion,
		),
		Report:          service.NewReportService(repos.Report),
		Promotion:       service.NewPromotionService(repos.Promotion),
		Inventory:       service.NewInventoryService(repos.Stock, repos.StockMovement, repos.ProductVariant, repos.Branch),
		CompanySettings: service.NewCompanySettingsService(repos.CompanySettings),
	}
}

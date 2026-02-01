package app

import (
	"github.com/irvanmhndra/pos-core-api/config"
	"github.com/irvanmhndra/pos-core-api/internal/service"
)

type Services struct {
	Auth            *service.AuthService
	User            *service.UserService
	Customer        *service.CustomerService
	ProductCategory *service.ProductCategoryService
	Product         *service.ProductService
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
		Customer:        service.NewCustomerService(repos.Customer),
		ProductCategory: service.NewProductCategoryService(repos.ProductCategory),
		Product:         service.NewProductService(repos.Product, repos.ProductVariant, repos.ProductCategory),
	}
}

package app

import (
	"github.com/irvanmhndra/pos-core-api/internal/repository"
	"github.com/irvanmhndra/pos-core-api/internal/repository/postgres"
	"github.com/jmoiron/sqlx"
)

type Repositories struct {
	User            repository.UserRepository
	Company         repository.CompanyRepository
	Branch          repository.BranchRepository
	UserSession     repository.UserSessionRepository
	Role            repository.RoleRepository
	Permission      repository.PermissionRepository
	RolePermission  repository.RolePermissionRepository
	UserBranch      repository.UserBranchRepository
	Customer        repository.CustomerRepository
	ProductCategory repository.ProductCategoryRepository
	Product         repository.ProductRepository
	ProductVariant  repository.ProductVariantRepository
}

func initRepositories(db *sqlx.DB) *Repositories {
	return &Repositories{
		User:            postgres.NewUserRepository(db),
		Company:         postgres.NewCompanyRepository(db),
		Branch:          postgres.NewBranchRepository(db),
		UserSession:     postgres.NewUserSessionRepository(db),
		Role:            postgres.NewRoleRepository(db),
		Permission:      postgres.NewPermissionRepository(db),
		RolePermission:  postgres.NewRolePermissionRepository(db),
		UserBranch:      postgres.NewUserBranchRepository(db),
		Customer:        postgres.NewCustomerRepository(db),
		ProductCategory: postgres.NewProductCategoryRepository(db),
		Product:         postgres.NewProductRepository(db),
		ProductVariant:  postgres.NewProductVariantRepository(db),
	}
}

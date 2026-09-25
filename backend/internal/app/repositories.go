package app

import (
	"github.com/irvanmhndra/nexpos-api/internal/repository"
	"github.com/irvanmhndra/nexpos-api/internal/repository/postgres"
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
	Order           repository.OrderRepository
	OrderItem       repository.OrderItemRepository
	Payment         repository.PaymentRepository
	Report          repository.ReportRepository
	CompanySettings repository.CompanySettingsRepository
	Promotion       repository.PromotionRepository
	Stock           repository.StockRepository
	StockMovement   repository.StockMovementRepository
	Supplier        repository.SupplierRepository
	PurchaseOrder   repository.PurchaseOrderRepository
	Shift           repository.ShiftRepository
	ExpenseCategory repository.ExpenseCategoryRepository
	Expense         repository.ExpenseRepository
	StockOpname     repository.StockOpnameRepository
	DailySettlement repository.DailySettlementRepository
	Receipt         repository.ReceiptRepository

	// Tx runs service operations atomically across the repositories above.
	Tx repository.Transactor
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
		Order:           postgres.NewOrderRepository(db),
		OrderItem:       postgres.NewOrderItemRepository(db),
		Payment:         postgres.NewPaymentRepository(db),
		Report:          postgres.NewReportRepository(db),
		CompanySettings: postgres.NewCompanySettingsRepository(db),
		Promotion:       postgres.NewPromotionRepository(db),
		Stock:           postgres.NewStockRepository(db),
		StockMovement:   postgres.NewStockMovementRepository(db),
		Supplier:        postgres.NewSupplierRepository(db),
		PurchaseOrder:   postgres.NewPurchaseOrderRepository(db),
		Shift:           postgres.NewShiftRepository(db),
		ExpenseCategory: postgres.NewExpenseCategoryRepository(db),
		Expense:         postgres.NewExpenseRepository(db),
		StockOpname:     postgres.NewStockOpnameRepository(db),
		DailySettlement: postgres.NewDailySettlementRepository(db),
		Receipt:         postgres.NewReceiptRepository(db),
		Tx:              postgres.NewTxManager(db),
	}
}

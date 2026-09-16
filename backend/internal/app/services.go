package app

import (
	"github.com/irvanmhndra/nexpos-api/config"
	"github.com/irvanmhndra/nexpos-api/internal/service"
	"github.com/irvanmhndra/nexpos-api/internal/storage"
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
	PurchaseOrder   *service.PurchaseOrderService
	Shift           *service.ShiftService
	Expense         *service.ExpenseService
	StockOpname     *service.StockOpnameService
	DailySettlement *service.DailySettlementService
	Receipt         *service.ReceiptService
}

func initServices(repos *Repositories, cfg *config.Config, store *storage.Client) *Services {
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
		Product:         service.NewProductService(repos.Product, repos.ProductVariant, repos.ProductCategory, store),
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
		PurchaseOrder: service.NewPurchaseOrderService(
			repos.Supplier,
			repos.PurchaseOrder,
			repos.Stock,
			repos.StockMovement,
			repos.ProductVariant,
		),
		Shift:   service.NewShiftService(repos.Shift, repos.Payment),
		Expense: service.NewExpenseService(repos.Expense, repos.ExpenseCategory),
		StockOpname: service.NewStockOpnameService(
			repos.StockOpname,
			repos.Stock,
			repos.StockMovement,
			repos.Branch,
		),
		DailySettlement: service.NewDailySettlementService(repos.DailySettlement, repos.Branch),
		Receipt:         service.NewReceiptService(repos.Receipt),
	}
}

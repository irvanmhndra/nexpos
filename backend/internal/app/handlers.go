package app

import (
	"github.com/irvanmhndra/nexpos-api/internal/handler"
	"github.com/irvanmhndra/nexpos-api/internal/storage"
	"github.com/irvanmhndra/nexpos-api/pkg/validator"
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
	Upload          *handler.UploadHandler
}

func initHandlers(db *sqlx.DB, services *Services, v *validator.CustomValidator, store *storage.Client) *Handlers {
	return &Handlers{
		Health:          handler.NewHealthHandler(db),
		Auth:            handler.NewAuthHandler(services.Auth, v),
		User:            handler.NewUserHandler(services.User, v),
		Branch:          handler.NewBranchHandler(services.Branch, v),
		Customer:        handler.NewCustomerHandler(services.Customer, v),
		ProductCategory: handler.NewProductCategoryHandler(services.ProductCategory, v),
		Product:         handler.NewProductHandler(services.Product, v),
		Order:           handler.NewOrderHandler(services.Order, v, services.Receipt),
		Report:          handler.NewReportHandler(services.Report, v),
		Promotion:       handler.NewPromotionHandler(services.Promotion, v),
		Inventory:       handler.NewInventoryHandler(services.Inventory, v),
		CompanySettings: handler.NewCompanySettingsHandler(services.CompanySettings, v),
		Supplier:        handler.NewSupplierHandler(services.PurchaseOrder, v),
		PurchaseOrder:   handler.NewPurchaseOrderHandler(services.PurchaseOrder, v),
		Shift:           handler.NewShiftHandler(services.Shift, v),
		ExpenseCategory: handler.NewExpenseCategoryHandler(services.Expense, v),
		Expense:         handler.NewExpenseHandler(services.Expense, v),
		StockOpname:     handler.NewStockOpnameHandler(services.StockOpname, v),
		DailySettlement: handler.NewDailySettlementHandler(services.DailySettlement, v),
		Receipt:         handler.NewReceiptHandler(services.Receipt),
		Upload:          handler.NewUploadHandler(store),
	}
}

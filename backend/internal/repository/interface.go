package repository

import (
	"context"
	"time"

	"github.com/irvanmhndra/nexpos-api/internal/model"
)

type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	GetByID(ctx context.Context, companyID, id int64) (*model.User, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	List(ctx context.Context, companyID int64, limit, offset int) ([]*model.User, int, error)
	Update(ctx context.Context, user *model.User) error
	UpdateStatus(ctx context.Context, companyID, id int64, status model.UserStatus) error
	Delete(ctx context.Context, companyID, id int64) error
	EmailExists(ctx context.Context, email string, excludeID int64) (bool, error)
}

type CompanyRepository interface {
	Create(ctx context.Context, company *model.Company) error
	GetByID(ctx context.Context, id int64) (*model.Company, error)
	GetByCode(ctx context.Context, code string) (*model.Company, error)
	Update(ctx context.Context, company *model.Company) error
}

type BranchRepository interface {
	Create(ctx context.Context, branch *model.Branch) error
	GetByID(ctx context.Context, id int64) (*model.Branch, error)
	ListByCompanyID(ctx context.Context, companyID int64) ([]*model.Branch, error)
	List(ctx context.Context, companyID int64, search string, isActive *bool, limit, offset int) ([]*model.Branch, int, error)
	Update(ctx context.Context, branch *model.Branch) error
	Delete(ctx context.Context, id int64) error
	CodeExists(ctx context.Context, companyID int64, code string, excludeID int64) (bool, error)
}

type UserSessionRepository interface {
	Create(ctx context.Context, session *model.UserSession) error
	GetByAccessToken(ctx context.Context, accessToken string) (*model.UserSession, error)
	GetByRefreshToken(ctx context.Context, refreshToken string) (*model.UserSession, error)
	UpdateLastUsed(ctx context.Context, id int64) error
	Revoke(ctx context.Context, id int64) error
	RevokeAllByUserID(ctx context.Context, userID int64) error
}

type RoleRepository interface {
	GetByID(ctx context.Context, id int64) (*model.Role, error)
	GetByCode(ctx context.Context, code string, companyID *int64) (*model.Role, error)
	ListByCompanyID(ctx context.Context, companyID int64) ([]*model.Role, error)
	ListSystemRoles(ctx context.Context) ([]*model.Role, error)
	Create(ctx context.Context, role *model.Role) error
	Update(ctx context.Context, role *model.Role) error
	Delete(ctx context.Context, id int64) error
	GetWithPermissions(ctx context.Context, id int64) (*model.Role, error)
}

type PermissionRepository interface {
	GetAll(ctx context.Context) ([]*model.Permission, error)
	GetByModule(ctx context.Context, module string) ([]*model.Permission, error)
	GetByRoleID(ctx context.Context, roleID int64) ([]*model.Permission, error)
}

type RolePermissionRepository interface {
	SetPermissions(ctx context.Context, roleID int64, permissionIDs []int64) error
	GetPermissionIDs(ctx context.Context, roleID int64) ([]int64, error)
}

type UserBranchRepository interface {
	Create(ctx context.Context, ub *model.UserBranch) error
	Delete(ctx context.Context, userID, branchID int64) error
	SetBranches(ctx context.Context, userID int64, branchIDs []int64, defaultBranchID int64) error
	GetByUserID(ctx context.Context, userID int64) ([]*model.UserBranch, error)
	GetDefaultBranch(ctx context.Context, userID int64) (*model.Branch, error)
}

type CustomerRepository interface {
	Create(ctx context.Context, customer *model.Customer) error
	GetByID(ctx context.Context, companyID, id int64) (*model.Customer, error)
	GetByCode(ctx context.Context, companyID int64, code string) (*model.Customer, error)
	List(ctx context.Context, companyID int64, search string, isMember *bool, limit, offset int) ([]*model.Customer, int, error)
	Update(ctx context.Context, customer *model.Customer) error
	Delete(ctx context.Context, companyID, id int64) error
	CodeExists(ctx context.Context, companyID int64, code string, excludeID int64) (bool, error)
}

type ProductCategoryRepository interface {
	Create(ctx context.Context, category *model.ProductCategory) error
	GetByID(ctx context.Context, companyID, id int64) (*model.ProductCategory, error)
	GetByCode(ctx context.Context, companyID int64, code string) (*model.ProductCategory, error)
	List(ctx context.Context, companyID int64, search string, isActive *bool, limit, offset int) ([]*model.ProductCategory, int, error)
	ListAll(ctx context.Context, companyID int64) ([]*model.ProductCategory, error)
	Update(ctx context.Context, category *model.ProductCategory) error
	Delete(ctx context.Context, companyID, id int64) error
	CodeExists(ctx context.Context, companyID int64, code string, excludeID int64) (bool, error)
	HasChildren(ctx context.Context, companyID, id int64) (bool, error)
	HasProducts(ctx context.Context, companyID, id int64) (bool, error)
}

type ProductRepository interface {
	Create(ctx context.Context, product *model.Product) error
	GetByID(ctx context.Context, companyID, id int64) (*model.Product, error)
	List(ctx context.Context, companyID int64, search string, categoryID *int64, isActive *bool, limit, offset int) ([]*model.Product, int, error)
	Update(ctx context.Context, product *model.Product) error
	Delete(ctx context.Context, companyID, id int64) error
}

type ProductVariantRepository interface {
	Create(ctx context.Context, variant *model.ProductVariant) error
	GetByID(ctx context.Context, id int64) (*model.ProductVariant, error)
	GetByProductID(ctx context.Context, productID int64) ([]*model.ProductVariant, error)
	GetBySKU(ctx context.Context, sku string) (*model.ProductVariant, error)
	Update(ctx context.Context, variant *model.ProductVariant) error
	Delete(ctx context.Context, id int64) error
	DeleteByProductID(ctx context.Context, productID int64) error
	SKUExists(ctx context.Context, sku string, excludeID int64) (bool, error)
	SKUExistsInOtherProduct(ctx context.Context, sku string, productID int64) (bool, error)
}

type OrderRepository interface {
	Create(ctx context.Context, order *model.Order) error
	GetByID(ctx context.Context, companyID, id int64) (*model.Order, error)
	GetByOrderNo(ctx context.Context, companyID int64, orderNo string) (*model.Order, error)
	GetByOfflineID(ctx context.Context, companyID int64, offlineID string) (*model.Order, error)
	List(ctx context.Context, companyID int64, params *OrderListParams) ([]*model.Order, int, error)
	Update(ctx context.Context, order *model.Order) error
	Delete(ctx context.Context, companyID, id int64) error
	GenerateOrderNo(ctx context.Context, companyID, branchID int64) (string, error)
}

// OrderListParams for advanced filtering
type OrderListParams struct {
	Search            string
	Status            *string
	PaymentStatus     *string
	FulfillmentType   *string
	FulfillmentStatus *string
	CustomerID        *int64
	DateFrom          *string
	DateTo            *string
	Limit             int
	Offset            int
}

type OrderItemRepository interface {
	Create(ctx context.Context, item *model.OrderItem) error
	GetByOrderID(ctx context.Context, orderID int64) ([]*model.OrderItem, error)
	DeleteByOrderID(ctx context.Context, orderID int64) error
}

type PaymentRepository interface {
	Create(ctx context.Context, payment *model.Payment) error
	GetByID(ctx context.Context, id int64) (*model.Payment, error)
	GetByOrderID(ctx context.Context, orderID int64) ([]*model.Payment, error)
	Update(ctx context.Context, payment *model.Payment) error
	DeleteByOrderID(ctx context.Context, orderID int64) error
	GetTotalPaidByOrderID(ctx context.Context, orderID int64) (float64, error)
	GetTotalRefundedByOrderID(ctx context.Context, orderID int64) (float64, error)
}

type CompanySettingsRepository interface {
	GetByCompanyID(ctx context.Context, companyID int64) (*model.CompanySettings, error)
	Upsert(ctx context.Context, settings *model.CompanySettings) error
}

type ReportRepository interface {
	GetSummary(ctx context.Context, companyID int64, dateFrom, dateTo string) (*ReportSummary, error)
	GetSalesTrend(ctx context.Context, companyID int64, dateFrom, dateTo string) ([]*SalesTrendItem, error)
	GetTopProducts(ctx context.Context, companyID int64, dateFrom, dateTo string, limit int) ([]*TopProductItem, error)
	GetCategoryRevenue(ctx context.Context, companyID int64, dateFrom, dateTo string) ([]*CategoryRevenueItem, error)
	GetPaymentMethods(ctx context.Context, companyID int64, dateFrom, dateTo string) ([]*PaymentMethodItem, error)
	GetHourlySales(ctx context.Context, companyID int64, dateFrom, dateTo string) ([]*HourlySalesItem, error)
	GetNewCustomersCount(ctx context.Context, companyID int64, dateFrom, dateTo string) (int, error)
}

// Report data structures
type ReportSummary struct {
	TotalRevenue    float64
	TotalOrders     int
	TotalItemsSold  int
	TotalCustomers  int
	CompletedOrders int
	CancelledOrders int
	PendingOrders   int
	TotalDiscount   float64
	TotalTax        float64
	TotalCOGS       float64
}

type SalesTrendItem struct {
	Date   string
	Sales  float64
	Orders int
	Items  int
}

type TopProductItem struct {
	ProductID   int64
	ProductName string
	TotalSold   int
	TotalAmount float64
}

type CategoryRevenueItem struct {
	CategoryID   int64
	CategoryName string
	TotalAmount  float64
	OrderCount   int
}

type PaymentMethodItem struct {
	Method      string
	TotalAmount float64
	Count       int
}

type HourlySalesItem struct {
	Hour        int
	TotalAmount float64
	OrderCount  int
}

type PromotionRepository interface {
	Create(ctx context.Context, promotion *model.Promotion) error
	GetByID(ctx context.Context, companyID, id int64) (*model.Promotion, error)
	List(ctx context.Context, companyID int64, search string, isActive *bool, promoType string, limit, offset int) ([]*model.Promotion, int, error)
	Update(ctx context.Context, promotion *model.Promotion) error
	Delete(ctx context.Context, companyID, id int64) error
	CodeExists(ctx context.Context, companyID int64, code string, excludeID int64) (bool, error)
	GetActivePromotions(ctx context.Context, companyID int64, now time.Time) ([]*model.Promotion, error)
	GetByCode(ctx context.Context, companyID int64, code string) (*model.Promotion, error)
}

type StockRepository interface {
	GetByVariantAndBranch(ctx context.Context, variantID, branchID int64) (*model.Stock, error)
	Upsert(ctx context.Context, stock *model.Stock) error
	UpdateMinQuantity(ctx context.Context, variantID, branchID int64, minQuantity int) error
	ListInventory(ctx context.Context, companyID, branchID int64, search, category, status string, limit, offset int) ([]*InventoryRow, int, error)
	GetInventoryStats(ctx context.Context, companyID, branchID int64) (*InventoryStats, error)
}

type StockMovementRepository interface {
	Create(ctx context.Context, m *model.StockMovement) error
	List(ctx context.Context, companyID, branchID int64, movType, search, startDate, endDate string, limit, offset int) ([]*MovementRow, int, error)
	GetMonthlyStats(ctx context.Context, companyID, branchID int64) (*MovementStats, error)
}

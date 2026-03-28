package service

import (
	"context"

	"github.com/irvanmhndra/nexpos-api/internal/dto"
	"github.com/irvanmhndra/nexpos-api/internal/model"
)

// AuthServiceInterface defines the contract for authentication operations
type AuthServiceInterface interface {
	Login(ctx context.Context, req dto.LoginRequest, ipAddress, userAgent string) (*dto.LoginResponse, error)
	Register(ctx context.Context, req dto.RegisterRequest, ipAddress, userAgent string) (*dto.RegisterResponse, error)
	RefreshToken(ctx context.Context, refreshToken string) (*dto.RefreshTokenResponse, error)
	Logout(ctx context.Context, accessToken string) error
	ValidateAccessToken(ctx context.Context, accessToken string) (*model.UserSession, error)
}

// UserServiceInterface defines the contract for user operations
type UserServiceInterface interface {
	Create(ctx context.Context, companyID int64, req dto.CreateUserRequest) (*dto.UserResponse, error)
	GetByID(ctx context.Context, companyID, id int64) (*dto.UserResponse, error)
	List(ctx context.Context, companyID int64, req dto.ListUsersRequest) (*UserListResult, error)
	Update(ctx context.Context, companyID, id int64, req dto.UpdateUserRequest) (*dto.UserResponse, error)
	UpdateStatus(ctx context.Context, companyID, id int64, req dto.UpdateUserStatusRequest) error
	Delete(ctx context.Context, companyID, id int64) error
}

// CustomerServiceInterface defines the contract for customer operations
type CustomerServiceInterface interface {
	Create(ctx context.Context, companyID int64, req dto.CreateCustomerRequest) (*dto.CustomerResponse, error)
	GetByID(ctx context.Context, companyID, id int64) (*dto.CustomerResponse, error)
	List(ctx context.Context, companyID int64, req dto.ListCustomerRequest) (*dto.CustomerListResponse, error)
	Update(ctx context.Context, companyID, id int64, req dto.UpdateCustomerRequest) (*dto.CustomerResponse, error)
	Delete(ctx context.Context, companyID, id int64) error
}

// ProductCategoryServiceInterface defines the contract for product category operations
type ProductCategoryServiceInterface interface {
	Create(ctx context.Context, companyID int64, req dto.CreateProductCategoryRequest) (*dto.ProductCategoryResponse, error)
	GetByID(ctx context.Context, companyID, id int64) (*dto.ProductCategoryResponse, error)
	List(ctx context.Context, companyID int64, req dto.ListProductCategoryRequest) (*dto.ProductCategoryListResponse, error)
	ListAll(ctx context.Context, companyID int64) ([]*dto.ProductCategoryResponse, error)
	Update(ctx context.Context, companyID, id int64, req dto.UpdateProductCategoryRequest) (*dto.ProductCategoryResponse, error)
	Delete(ctx context.Context, companyID, id int64) error
}

// ProductServiceInterface defines the contract for product operations
type ProductServiceInterface interface {
	Create(ctx context.Context, companyID int64, req dto.CreateProductRequest) (*dto.ProductResponse, error)
	GetByID(ctx context.Context, companyID, id int64) (*dto.ProductResponse, error)
	List(ctx context.Context, companyID int64, req dto.ListProductRequest) (*dto.ProductListResponse, error)
	Update(ctx context.Context, companyID, id int64, req dto.UpdateProductRequest) (*dto.ProductResponse, error)
	Delete(ctx context.Context, companyID, id int64) error
}

// OrderServiceInterface defines the contract for order operations
type OrderServiceInterface interface {
	Preview(ctx context.Context, companyID, branchID int64, req dto.PreviewOrderRequest) (*dto.PreviewOrderResponse, error)
	Create(ctx context.Context, companyID, branchID, cashierID int64, req dto.CreateOrderRequest) (*dto.OrderResponse, error)
	GetByID(ctx context.Context, companyID, id int64) (*dto.OrderResponse, error)
	List(ctx context.Context, companyID int64, req dto.ListOrderRequest) (*dto.OrderListResponse, error)
	UpdateOrder(ctx context.Context, companyID, id int64, req dto.UpdateOrderRequest) (*dto.OrderResponse, error)
	ConfirmOrder(ctx context.Context, companyID, id int64) (*dto.OrderResponse, error)
	AddPayment(ctx context.Context, companyID, id int64, req dto.AddPaymentRequest) (*dto.OrderResponse, error)
	CompleteOrder(ctx context.Context, companyID, id int64, req dto.CompleteOrderRequest) (*dto.OrderResponse, error)
	CancelOrder(ctx context.Context, companyID, id int64, req dto.CancelOrderRequest) (*dto.OrderResponse, error)
	VoidOrder(ctx context.Context, companyID, id int64, req dto.VoidOrderRequest) (*dto.OrderResponse, error)
	RefundPayment(ctx context.Context, companyID, orderID int64, req dto.RefundPaymentRequest) (*dto.OrderResponse, error)
}

// BranchServiceInterface defines the contract for branch operations
type BranchServiceInterface interface {
	Create(ctx context.Context, companyID int64, req dto.CreateBranchRequest) (*dto.BranchResponse, error)
	GetByID(ctx context.Context, companyID, id int64) (*dto.BranchResponse, error)
	List(ctx context.Context, companyID int64, req dto.ListBranchRequest) (*dto.BranchListResponse, error)
	Update(ctx context.Context, companyID, id int64, req dto.UpdateBranchRequest) (*dto.BranchResponse, error)
	Delete(ctx context.Context, companyID, id int64) error
}

// PromotionServiceInterface defines the contract for promotion operations
type PromotionServiceInterface interface {
	Create(ctx context.Context, companyID int64, req dto.CreatePromotionRequest) (*dto.PromotionResponse, error)
	GetByID(ctx context.Context, companyID, id int64) (*dto.PromotionResponse, error)
	List(ctx context.Context, companyID int64, req dto.ListPromotionRequest) (*dto.PromotionListResponse, error)
	Update(ctx context.Context, companyID, id int64, req dto.UpdatePromotionRequest) (*dto.PromotionResponse, error)
	Delete(ctx context.Context, companyID, id int64) error
}

// ReportServiceInterface defines the contract for reporting operations
type ReportServiceInterface interface {
	GetSummary(ctx context.Context, companyID int64, dateFrom, dateTo string) (*dto.ReportSummaryResponse, error)
	GetSalesTrend(ctx context.Context, companyID int64, dateFrom, dateTo string) (*dto.SalesTrendResponse, error)
	GetTopProducts(ctx context.Context, companyID int64, dateFrom, dateTo string, limit int) (*dto.TopProductsResponse, error)
	GetCategoryRevenue(ctx context.Context, companyID int64, dateFrom, dateTo string) (*dto.CategoryRevenueResponse, error)
	GetPaymentMethods(ctx context.Context, companyID int64, dateFrom, dateTo string) (*dto.PaymentMethodResponse, error)
	GetHourlySales(ctx context.Context, companyID int64, dateFrom, dateTo string) (*dto.HourlySalesResponse, error)
}

// CompanySettingsServiceInterface defines the contract for company settings operations
type CompanySettingsServiceInterface interface {
	Get(ctx context.Context, companyID int64) (*dto.CompanySettingsResponse, error)
	Update(ctx context.Context, companyID int64, req dto.UpdateCompanySettingsRequest) (*dto.CompanySettingsResponse, error)
}

// InventoryServiceInterface defines the contract for inventory operations
type InventoryServiceInterface interface {
	AdjustStock(ctx context.Context, companyID int64, createdBy *int64, req dto.AdjustStockRequest) error
	GetInventoryStats(ctx context.Context, companyID, branchID int64) (*dto.InventoryStatsResponse, error)
	GetMovementStats(ctx context.Context, companyID, branchID int64) (*dto.MovementStatsResponse, error)
	ListInventory(ctx context.Context, companyID int64, req dto.ListInventoryRequest) (*dto.InventoryListResponse, error)
	ListMovements(ctx context.Context, companyID int64, req dto.ListMovementsRequest) (*dto.MovementListResponse, error)
	UpdateMinStock(ctx context.Context, companyID, variantID, branchID int64, req dto.UpdateMinStockRequest) error
}

// PurchaseOrderServiceInterface defines the contract for purchase order and supplier operations
type PurchaseOrderServiceInterface interface {
	CreateSupplier(ctx context.Context, companyID int64, req dto.CreateSupplierRequest) (*dto.SupplierResponse, error)
	GetSupplier(ctx context.Context, companyID, id int64) (*dto.SupplierResponse, error)
	ListSuppliers(ctx context.Context, companyID int64, req dto.ListSupplierRequest) (*dto.SupplierListResponse, error)
	UpdateSupplier(ctx context.Context, companyID, id int64, req dto.UpdateSupplierRequest) (*dto.SupplierResponse, error)
	DeleteSupplier(ctx context.Context, companyID, id int64) error
	CreatePO(ctx context.Context, companyID, branchID, userID int64, req dto.CreatePurchaseOrderRequest) (*dto.PurchaseOrderResponse, error)
	GetPO(ctx context.Context, companyID, id int64) (*dto.PurchaseOrderResponse, error)
	ListPOs(ctx context.Context, companyID int64, req dto.ListPurchaseOrderRequest) (*dto.PurchaseOrderListResponse, error)
	ReceivePO(ctx context.Context, companyID, poID int64, req dto.ReceivePurchaseOrderRequest) (*dto.PurchaseOrderResponse, error)
}

// ShiftServiceInterface defines the contract for shift management operations
type ShiftServiceInterface interface {
	OpenShift(ctx context.Context, companyID, branchID, cashierID int64, req dto.OpenShiftRequest) (*dto.ShiftResponse, error)
	GetCurrentShift(ctx context.Context, companyID, branchID, cashierID int64) (*dto.ShiftResponse, error)
	CloseShift(ctx context.Context, companyID, shiftID int64, req dto.CloseShiftRequest) (*dto.ShiftResponse, error)
	GetShift(ctx context.Context, companyID, id int64) (*dto.ShiftResponse, error)
	ListShifts(ctx context.Context, companyID int64, req dto.ListShiftRequest) (*dto.ShiftListResponse, error)
}

// ExpenseServiceInterface defines the contract for expense tracking operations
type ExpenseServiceInterface interface {
	CreateCategory(ctx context.Context, companyID int64, req dto.CreateExpenseCategoryRequest) (*dto.ExpenseCategoryResponse, error)
	ListCategories(ctx context.Context, companyID int64, req dto.ListExpenseCategoryRequest) (*dto.ExpenseCategoryListResponse, error)
	UpdateCategory(ctx context.Context, companyID, id int64, req dto.UpdateExpenseCategoryRequest) (*dto.ExpenseCategoryResponse, error)
	DeleteCategory(ctx context.Context, companyID, id int64) error
	CreateExpense(ctx context.Context, companyID int64, recordedBy *int64, req dto.CreateExpenseRequest) (*dto.ExpenseResponse, error)
	GetExpense(ctx context.Context, companyID, id int64) (*dto.ExpenseResponse, error)
	ListExpenses(ctx context.Context, companyID int64, req dto.ListExpenseRequest) (*dto.ExpenseListResponse, error)
	UpdateExpense(ctx context.Context, companyID, id int64, req dto.UpdateExpenseRequest) (*dto.ExpenseResponse, error)
	DeleteExpense(ctx context.Context, companyID, id int64) error
	GetExpenseSummary(ctx context.Context, companyID int64, req dto.ExpenseSummaryRequest) (*dto.ExpenseSummaryResponse, error)
}

// Ensure concrete types implement interfaces
var _ AuthServiceInterface = (*AuthService)(nil)
var _ UserServiceInterface = (*UserService)(nil)
var _ CustomerServiceInterface = (*CustomerService)(nil)
var _ ProductCategoryServiceInterface = (*ProductCategoryService)(nil)
var _ ProductServiceInterface = (*ProductService)(nil)
var _ OrderServiceInterface = (*OrderService)(nil)
var _ BranchServiceInterface = (*BranchService)(nil)
var _ PromotionServiceInterface = (*PromotionService)(nil)
var _ ReportServiceInterface = (*ReportService)(nil)
var _ CompanySettingsServiceInterface = (*CompanySettingsService)(nil)
var _ InventoryServiceInterface = (*InventoryService)(nil)
var _ PurchaseOrderServiceInterface = (*PurchaseOrderService)(nil)
var _ ShiftServiceInterface = (*ShiftService)(nil)
var _ ExpenseServiceInterface = (*ExpenseService)(nil)

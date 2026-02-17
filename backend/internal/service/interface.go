package service

import (
	"context"

	"github.com/irvanmhndra/pos-core-api/internal/dto"
	"github.com/irvanmhndra/pos-core-api/internal/model"
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

// Ensure concrete types implement interfaces
var _ AuthServiceInterface = (*AuthService)(nil)
var _ UserServiceInterface = (*UserService)(nil)
var _ CustomerServiceInterface = (*CustomerService)(nil)
var _ ProductCategoryServiceInterface = (*ProductCategoryService)(nil)
var _ ProductServiceInterface = (*ProductService)(nil)
var _ OrderServiceInterface = (*OrderService)(nil)

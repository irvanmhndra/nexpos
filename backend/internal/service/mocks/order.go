package mocks

import (
	"context"

	"github.com/irvanmhndra/nexpos-api/internal/dto"
	"github.com/stretchr/testify/mock"
)

// MockOrderService is a mock implementation of service.OrderServiceInterface
type MockOrderService struct {
	mock.Mock
}

func (m *MockOrderService) Create(ctx context.Context, companyID, branchID, cashierID int64, req dto.CreateOrderRequest) (*dto.OrderResponse, error) {
	args := m.Called(ctx, companyID, branchID, cashierID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.OrderResponse), args.Error(1)
}

func (m *MockOrderService) GetByID(ctx context.Context, companyID, id int64) (*dto.OrderResponse, error) {
	args := m.Called(ctx, companyID, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.OrderResponse), args.Error(1)
}

func (m *MockOrderService) List(ctx context.Context, companyID int64, req dto.ListOrderRequest) (*dto.OrderListResponse, error) {
	args := m.Called(ctx, companyID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.OrderListResponse), args.Error(1)
}

func (m *MockOrderService) UpdateOrder(ctx context.Context, companyID, id int64, req dto.UpdateOrderRequest) (*dto.OrderResponse, error) {
	args := m.Called(ctx, companyID, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.OrderResponse), args.Error(1)
}

func (m *MockOrderService) ConfirmOrder(ctx context.Context, companyID, id int64) (*dto.OrderResponse, error) {
	args := m.Called(ctx, companyID, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.OrderResponse), args.Error(1)
}

func (m *MockOrderService) AddPayment(ctx context.Context, companyID, id int64, req dto.AddPaymentRequest) (*dto.OrderResponse, error) {
	args := m.Called(ctx, companyID, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.OrderResponse), args.Error(1)
}

func (m *MockOrderService) CompleteOrder(ctx context.Context, companyID, id int64, req dto.CompleteOrderRequest) (*dto.OrderResponse, error) {
	args := m.Called(ctx, companyID, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.OrderResponse), args.Error(1)
}

func (m *MockOrderService) CancelOrder(ctx context.Context, companyID, id int64, req dto.CancelOrderRequest) (*dto.OrderResponse, error) {
	args := m.Called(ctx, companyID, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.OrderResponse), args.Error(1)
}

func (m *MockOrderService) VoidOrder(ctx context.Context, companyID, id int64, req dto.VoidOrderRequest) (*dto.OrderResponse, error) {
	args := m.Called(ctx, companyID, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.OrderResponse), args.Error(1)
}

func (m *MockOrderService) RefundPayment(ctx context.Context, companyID, orderID int64, req dto.RefundPaymentRequest) (*dto.OrderResponse, error) {
	args := m.Called(ctx, companyID, orderID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.OrderResponse), args.Error(1)
}

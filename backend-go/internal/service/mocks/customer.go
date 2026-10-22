package mocks

import (
	"context"

	"github.com/irvanmhndra/nexpos-api/internal/dto"
	"github.com/stretchr/testify/mock"
)

// MockCustomerService is a mock implementation of service.CustomerServiceInterface
type MockCustomerService struct {
	mock.Mock
}

func (m *MockCustomerService) Create(ctx context.Context, companyID int64, req dto.CreateCustomerRequest) (*dto.CustomerResponse, error) {
	args := m.Called(ctx, companyID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.CustomerResponse), args.Error(1)
}

func (m *MockCustomerService) GetByID(ctx context.Context, companyID, id int64) (*dto.CustomerResponse, error) {
	args := m.Called(ctx, companyID, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.CustomerResponse), args.Error(1)
}

func (m *MockCustomerService) List(ctx context.Context, companyID int64, req dto.ListCustomerRequest) (*dto.CustomerListResponse, error) {
	args := m.Called(ctx, companyID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.CustomerListResponse), args.Error(1)
}

func (m *MockCustomerService) Update(ctx context.Context, companyID, id int64, req dto.UpdateCustomerRequest) (*dto.CustomerResponse, error) {
	args := m.Called(ctx, companyID, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.CustomerResponse), args.Error(1)
}

func (m *MockCustomerService) Delete(ctx context.Context, companyID, id int64) error {
	args := m.Called(ctx, companyID, id)
	return args.Error(0)
}

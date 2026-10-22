package mocks

import (
	"context"

	"github.com/irvanmhndra/nexpos-api/internal/dto"
	"github.com/stretchr/testify/mock"
)

// MockProductService is a mock implementation of service.ProductServiceInterface
type MockProductService struct {
	mock.Mock
}

func (m *MockProductService) Create(ctx context.Context, companyID int64, req dto.CreateProductRequest) (*dto.ProductResponse, error) {
	args := m.Called(ctx, companyID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.ProductResponse), args.Error(1)
}

func (m *MockProductService) GetByID(ctx context.Context, companyID, id int64) (*dto.ProductResponse, error) {
	args := m.Called(ctx, companyID, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.ProductResponse), args.Error(1)
}

func (m *MockProductService) List(ctx context.Context, companyID int64, req dto.ListProductRequest) (*dto.ProductListResponse, error) {
	args := m.Called(ctx, companyID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.ProductListResponse), args.Error(1)
}

func (m *MockProductService) Update(ctx context.Context, companyID, id int64, req dto.UpdateProductRequest) (*dto.ProductResponse, error) {
	args := m.Called(ctx, companyID, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.ProductResponse), args.Error(1)
}

func (m *MockProductService) Delete(ctx context.Context, companyID, id int64) error {
	args := m.Called(ctx, companyID, id)
	return args.Error(0)
}

func (m *MockProductService) LookupByCode(ctx context.Context, companyID int64, code string) (*dto.ProductLookupResponse, error) {
	args := m.Called(ctx, companyID, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.ProductLookupResponse), args.Error(1)
}

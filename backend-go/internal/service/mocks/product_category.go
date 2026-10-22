package mocks

import (
	"context"

	"github.com/irvanmhndra/nexpos-api/internal/dto"
	"github.com/stretchr/testify/mock"
)

// MockProductCategoryService is a mock implementation of service.ProductCategoryServiceInterface
type MockProductCategoryService struct {
	mock.Mock
}

func (m *MockProductCategoryService) Create(ctx context.Context, companyID int64, req dto.CreateProductCategoryRequest) (*dto.ProductCategoryResponse, error) {
	args := m.Called(ctx, companyID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.ProductCategoryResponse), args.Error(1)
}

func (m *MockProductCategoryService) GetByID(ctx context.Context, companyID, id int64) (*dto.ProductCategoryResponse, error) {
	args := m.Called(ctx, companyID, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.ProductCategoryResponse), args.Error(1)
}

func (m *MockProductCategoryService) List(ctx context.Context, companyID int64, req dto.ListProductCategoryRequest) (*dto.ProductCategoryListResponse, error) {
	args := m.Called(ctx, companyID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.ProductCategoryListResponse), args.Error(1)
}

func (m *MockProductCategoryService) ListAll(ctx context.Context, companyID int64) ([]*dto.ProductCategoryResponse, error) {
	args := m.Called(ctx, companyID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*dto.ProductCategoryResponse), args.Error(1)
}

func (m *MockProductCategoryService) Update(ctx context.Context, companyID, id int64, req dto.UpdateProductCategoryRequest) (*dto.ProductCategoryResponse, error) {
	args := m.Called(ctx, companyID, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.ProductCategoryResponse), args.Error(1)
}

func (m *MockProductCategoryService) Delete(ctx context.Context, companyID, id int64) error {
	args := m.Called(ctx, companyID, id)
	return args.Error(0)
}

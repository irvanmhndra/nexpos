package mocks

import (
	"context"

	"github.com/irvanmhndra/nexpos-api/internal/dto"
	"github.com/stretchr/testify/mock"
)

// MockPromotionService is a mock implementation of service.PromotionServiceInterface
type MockPromotionService struct {
	mock.Mock
}

func (m *MockPromotionService) Create(ctx context.Context, companyID int64, req dto.CreatePromotionRequest) (*dto.PromotionResponse, error) {
	args := m.Called(ctx, companyID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.PromotionResponse), args.Error(1)
}

func (m *MockPromotionService) GetByID(ctx context.Context, companyID, id int64) (*dto.PromotionResponse, error) {
	args := m.Called(ctx, companyID, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.PromotionResponse), args.Error(1)
}

func (m *MockPromotionService) List(ctx context.Context, companyID int64, req dto.ListPromotionRequest) (*dto.PromotionListResponse, error) {
	args := m.Called(ctx, companyID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.PromotionListResponse), args.Error(1)
}

func (m *MockPromotionService) Update(ctx context.Context, companyID, id int64, req dto.UpdatePromotionRequest) (*dto.PromotionResponse, error) {
	args := m.Called(ctx, companyID, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.PromotionResponse), args.Error(1)
}

func (m *MockPromotionService) Delete(ctx context.Context, companyID, id int64) error {
	args := m.Called(ctx, companyID, id)
	return args.Error(0)
}

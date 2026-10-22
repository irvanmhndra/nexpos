package mocks

import (
	"context"

	"github.com/irvanmhndra/nexpos-api/internal/dto"
	"github.com/stretchr/testify/mock"
)

// MockInventoryService is a mock implementation of service.InventoryServiceInterface
type MockInventoryService struct {
	mock.Mock
}

func (m *MockInventoryService) AdjustStock(ctx context.Context, companyID int64, createdBy *int64, req dto.AdjustStockRequest) error {
	args := m.Called(ctx, companyID, createdBy, req)
	return args.Error(0)
}

func (m *MockInventoryService) GetInventoryStats(ctx context.Context, companyID, branchID int64) (*dto.InventoryStatsResponse, error) {
	args := m.Called(ctx, companyID, branchID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.InventoryStatsResponse), args.Error(1)
}

func (m *MockInventoryService) GetMovementStats(ctx context.Context, companyID, branchID int64) (*dto.MovementStatsResponse, error) {
	args := m.Called(ctx, companyID, branchID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.MovementStatsResponse), args.Error(1)
}

func (m *MockInventoryService) ListInventory(ctx context.Context, companyID int64, req dto.ListInventoryRequest) (*dto.InventoryListResponse, error) {
	args := m.Called(ctx, companyID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.InventoryListResponse), args.Error(1)
}

func (m *MockInventoryService) ListMovements(ctx context.Context, companyID int64, req dto.ListMovementsRequest) (*dto.MovementListResponse, error) {
	args := m.Called(ctx, companyID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.MovementListResponse), args.Error(1)
}

func (m *MockInventoryService) UpdateMinStock(ctx context.Context, companyID, variantID, branchID int64, req dto.UpdateMinStockRequest) error {
	args := m.Called(ctx, companyID, variantID, branchID, req)
	return args.Error(0)
}

package mocks

import (
	"context"

	"github.com/irvanmhndra/nexpos-api/internal/model"
	"github.com/irvanmhndra/nexpos-api/internal/repository"
	"github.com/stretchr/testify/mock"
)

// MockStockRepository is a mock implementation of repository.StockRepository.
type MockStockRepository struct {
	mock.Mock
}

func NewMockStockRepository(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockStockRepository {
	m := &MockStockRepository{}
	m.Test(t)
	t.Cleanup(func() { m.AssertExpectations(t) })
	return m
}

func (m *MockStockRepository) EXPECT() *MockStockRepositoryExpectation {
	return &MockStockRepositoryExpectation{mock: m}
}

func (m *MockStockRepository) LockForUpdate(ctx context.Context, variantID, branchID int64) (*model.Stock, error) {
	ret := m.Called(ctx, variantID, branchID)
	var r0 *model.Stock
	if ret.Get(0) != nil {
		r0 = ret.Get(0).(*model.Stock)
	}
	return r0, ret.Error(1)
}

func (m *MockStockRepository) GetByVariantAndBranch(ctx context.Context, variantID, branchID int64) (*model.Stock, error) {
	ret := m.Called(ctx, variantID, branchID)
	var r0 *model.Stock
	if ret.Get(0) != nil {
		r0 = ret.Get(0).(*model.Stock)
	}
	return r0, ret.Error(1)
}

func (m *MockStockRepository) Upsert(ctx context.Context, stock *model.Stock) error {
	ret := m.Called(ctx, stock)
	return ret.Error(0)
}

func (m *MockStockRepository) UpdateMinQuantity(ctx context.Context, variantID, branchID int64, minQuantity int) error {
	ret := m.Called(ctx, variantID, branchID, minQuantity)
	return ret.Error(0)
}

func (m *MockStockRepository) ListInventory(ctx context.Context, companyID, branchID int64, search, category, status string, limit, offset int) ([]*repository.InventoryRow, int, error) {
	ret := m.Called(ctx, companyID, branchID, search, category, status, limit, offset)
	var r0 []*repository.InventoryRow
	if ret.Get(0) != nil {
		r0 = ret.Get(0).([]*repository.InventoryRow)
	}
	return r0, ret.Int(1), ret.Error(2)
}

func (m *MockStockRepository) GetInventoryStats(ctx context.Context, companyID, branchID int64) (*repository.InventoryStats, error) {
	ret := m.Called(ctx, companyID, branchID)
	var r0 *repository.InventoryStats
	if ret.Get(0) != nil {
		r0 = ret.Get(0).(*repository.InventoryStats)
	}
	return r0, ret.Error(1)
}

// MockStockRepositoryExpectation is the expectation builder for MockStockRepository.
type MockStockRepositoryExpectation struct {
	mock *MockStockRepository
}

func (e *MockStockRepositoryExpectation) GetByVariantAndBranch(ctx context.Context, variantID, branchID int64) *mock.Call {
	return e.mock.On("GetByVariantAndBranch", ctx, variantID, branchID)
}

func (e *MockStockRepositoryExpectation) Upsert(ctx context.Context, stock interface{}) *mock.Call {
	return e.mock.On("Upsert", ctx, stock)
}

func (e *MockStockRepositoryExpectation) UpdateMinQuantity(ctx context.Context, variantID, branchID int64, minQuantity int) *mock.Call {
	return e.mock.On("UpdateMinQuantity", ctx, variantID, branchID, minQuantity)
}

func (e *MockStockRepositoryExpectation) ListInventory(ctx context.Context, companyID, branchID int64, search, category, status string, limit, offset int) *mock.Call {
	return e.mock.On("ListInventory", ctx, companyID, branchID, search, category, status, limit, offset)
}

func (e *MockStockRepositoryExpectation) GetInventoryStats(ctx context.Context, companyID, branchID int64) *mock.Call {
	return e.mock.On("GetInventoryStats", ctx, companyID, branchID)
}

func (e *MockStockRepositoryExpectation) LockForUpdate(ctx context.Context, variantID, branchID int64) *mock.Call {
	return e.mock.On("LockForUpdate", ctx, variantID, branchID)
}

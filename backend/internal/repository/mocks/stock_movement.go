package mocks

import (
	"context"

	"github.com/irvanmhndra/nexpos-api/internal/model"
	"github.com/irvanmhndra/nexpos-api/internal/repository"
	"github.com/stretchr/testify/mock"
)

// MockStockMovementRepository is a mock implementation of repository.StockMovementRepository.
type MockStockMovementRepository struct {
	mock.Mock
}

func NewMockStockMovementRepository(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockStockMovementRepository {
	m := &MockStockMovementRepository{}
	m.Test(t)
	t.Cleanup(func() { m.AssertExpectations(t) })
	return m
}

func (m *MockStockMovementRepository) EXPECT() *MockStockMovementRepositoryExpectation {
	return &MockStockMovementRepositoryExpectation{mock: m}
}

func (m *MockStockMovementRepository) Create(ctx context.Context, mv *model.StockMovement) error {
	ret := m.Called(ctx, mv)
	return ret.Error(0)
}

func (m *MockStockMovementRepository) List(ctx context.Context, companyID, branchID int64, movType, search, startDate, endDate string, limit, offset int) ([]*repository.MovementRow, int, error) {
	ret := m.Called(ctx, companyID, branchID, movType, search, startDate, endDate, limit, offset)
	var r0 []*repository.MovementRow
	if ret.Get(0) != nil {
		r0 = ret.Get(0).([]*repository.MovementRow)
	}
	return r0, ret.Int(1), ret.Error(2)
}

func (m *MockStockMovementRepository) GetMonthlyStats(ctx context.Context, companyID, branchID int64) (*repository.MovementStats, error) {
	ret := m.Called(ctx, companyID, branchID)
	var r0 *repository.MovementStats
	if ret.Get(0) != nil {
		r0 = ret.Get(0).(*repository.MovementStats)
	}
	return r0, ret.Error(1)
}

// MockStockMovementRepositoryExpectation is the expectation builder for MockStockMovementRepository.
type MockStockMovementRepositoryExpectation struct {
	mock *MockStockMovementRepository
}

func (e *MockStockMovementRepositoryExpectation) Create(ctx context.Context, mv interface{}) *mock.Call {
	return e.mock.On("Create", ctx, mv)
}

func (e *MockStockMovementRepositoryExpectation) List(ctx context.Context, companyID, branchID int64, movType, search, startDate, endDate string, limit, offset int) *mock.Call {
	return e.mock.On("List", ctx, companyID, branchID, movType, search, startDate, endDate, limit, offset)
}

func (e *MockStockMovementRepositoryExpectation) GetMonthlyStats(ctx context.Context, companyID, branchID int64) *mock.Call {
	return e.mock.On("GetMonthlyStats", ctx, companyID, branchID)
}

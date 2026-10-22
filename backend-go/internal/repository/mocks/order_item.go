package mocks

import (
	"context"

	"github.com/irvanmhndra/nexpos-api/internal/model"
	"github.com/stretchr/testify/mock"
)

// MockOrderItemRepository is a mock implementation of repository.OrderItemRepository.
type MockOrderItemRepository struct {
	mock.Mock
}

func NewMockOrderItemRepository(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockOrderItemRepository {
	m := &MockOrderItemRepository{}
	m.Test(t)
	t.Cleanup(func() { m.AssertExpectations(t) })
	return m
}

func (m *MockOrderItemRepository) EXPECT() *MockOrderItemRepositoryExpectation {
	return &MockOrderItemRepositoryExpectation{mock: m}
}

func (m *MockOrderItemRepository) Create(ctx context.Context, item *model.OrderItem) error {
	ret := m.Called(ctx, item)
	return ret.Error(0)
}

func (m *MockOrderItemRepository) GetByOrderID(ctx context.Context, orderID int64) ([]*model.OrderItem, error) {
	ret := m.Called(ctx, orderID)
	var r0 []*model.OrderItem
	if ret.Get(0) != nil {
		r0 = ret.Get(0).([]*model.OrderItem)
	}
	return r0, ret.Error(1)
}

func (m *MockOrderItemRepository) DeleteByOrderID(ctx context.Context, orderID int64) error {
	ret := m.Called(ctx, orderID)
	return ret.Error(0)
}

// MockOrderItemRepositoryExpectation is the expectation builder for MockOrderItemRepository.
type MockOrderItemRepositoryExpectation struct {
	mock *MockOrderItemRepository
}

func (e *MockOrderItemRepositoryExpectation) Create(ctx context.Context, item interface{}) *mock.Call {
	return e.mock.On("Create", ctx, item)
}

func (e *MockOrderItemRepositoryExpectation) GetByOrderID(ctx context.Context, orderID int64) *mock.Call {
	return e.mock.On("GetByOrderID", ctx, orderID)
}

func (e *MockOrderItemRepositoryExpectation) DeleteByOrderID(ctx context.Context, orderID int64) *mock.Call {
	return e.mock.On("DeleteByOrderID", ctx, orderID)
}

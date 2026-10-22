package mocks

import (
	"context"

	"github.com/irvanmhndra/nexpos-api/internal/model"
	"github.com/irvanmhndra/nexpos-api/internal/repository"
	"github.com/stretchr/testify/mock"
)

// MockOrderRepository is a mock implementation of repository.OrderRepository.
type MockOrderRepository struct {
	mock.Mock
}

func NewMockOrderRepository(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockOrderRepository {
	m := &MockOrderRepository{}
	m.Test(t)
	t.Cleanup(func() { m.AssertExpectations(t) })
	return m
}

func (m *MockOrderRepository) EXPECT() *MockOrderRepositoryExpectation {
	return &MockOrderRepositoryExpectation{mock: m}
}

func (m *MockOrderRepository) GetByIDForUpdate(ctx context.Context, companyID, id int64) (*model.Order, error) {
	ret := m.Called(ctx, companyID, id)
	var r0 *model.Order
	if ret.Get(0) != nil {
		r0 = ret.Get(0).(*model.Order)
	}
	return r0, ret.Error(1)
}

func (m *MockOrderRepository) Create(ctx context.Context, order *model.Order) error {
	ret := m.Called(ctx, order)
	return ret.Error(0)
}

func (m *MockOrderRepository) GetByID(ctx context.Context, companyID, id int64) (*model.Order, error) {
	ret := m.Called(ctx, companyID, id)
	var r0 *model.Order
	if ret.Get(0) != nil {
		r0 = ret.Get(0).(*model.Order)
	}
	return r0, ret.Error(1)
}

func (m *MockOrderRepository) GetByOrderNo(ctx context.Context, companyID int64, orderNo string) (*model.Order, error) {
	ret := m.Called(ctx, companyID, orderNo)
	var r0 *model.Order
	if ret.Get(0) != nil {
		r0 = ret.Get(0).(*model.Order)
	}
	return r0, ret.Error(1)
}

func (m *MockOrderRepository) GetByOfflineID(ctx context.Context, companyID int64, offlineID string) (*model.Order, error) {
	ret := m.Called(ctx, companyID, offlineID)
	var r0 *model.Order
	if ret.Get(0) != nil {
		r0 = ret.Get(0).(*model.Order)
	}
	return r0, ret.Error(1)
}

func (m *MockOrderRepository) List(ctx context.Context, companyID int64, params *repository.OrderListParams) ([]*model.Order, int, error) {
	ret := m.Called(ctx, companyID, params)
	var r0 []*model.Order
	if ret.Get(0) != nil {
		r0 = ret.Get(0).([]*model.Order)
	}
	return r0, ret.Int(1), ret.Error(2)
}

func (m *MockOrderRepository) Update(ctx context.Context, order *model.Order) error {
	ret := m.Called(ctx, order)
	return ret.Error(0)
}

func (m *MockOrderRepository) Delete(ctx context.Context, companyID, id int64) error {
	ret := m.Called(ctx, companyID, id)
	return ret.Error(0)
}

func (m *MockOrderRepository) GenerateOrderNo(ctx context.Context, companyID, branchID int64) (string, error) {
	ret := m.Called(ctx, companyID, branchID)
	return ret.String(0), ret.Error(1)
}

// MockOrderRepositoryExpectation is the expectation builder for MockOrderRepository.
type MockOrderRepositoryExpectation struct {
	mock *MockOrderRepository
}

func (e *MockOrderRepositoryExpectation) Create(ctx context.Context, order interface{}) *mock.Call {
	return e.mock.On("Create", ctx, order)
}

func (e *MockOrderRepositoryExpectation) GetByID(ctx context.Context, companyID, id int64) *mock.Call {
	return e.mock.On("GetByID", ctx, companyID, id)
}

func (e *MockOrderRepositoryExpectation) GetByOrderNo(ctx context.Context, companyID int64, orderNo string) *mock.Call {
	return e.mock.On("GetByOrderNo", ctx, companyID, orderNo)
}

func (e *MockOrderRepositoryExpectation) GetByOfflineID(ctx context.Context, companyID int64, offlineID string) *mock.Call {
	return e.mock.On("GetByOfflineID", ctx, companyID, offlineID)
}

func (e *MockOrderRepositoryExpectation) List(ctx context.Context, companyID int64, params interface{}) *mock.Call {
	return e.mock.On("List", ctx, companyID, params)
}

func (e *MockOrderRepositoryExpectation) Update(ctx context.Context, order interface{}) *mock.Call {
	return e.mock.On("Update", ctx, order)
}

func (e *MockOrderRepositoryExpectation) Delete(ctx context.Context, companyID, id int64) *mock.Call {
	return e.mock.On("Delete", ctx, companyID, id)
}

func (e *MockOrderRepositoryExpectation) GenerateOrderNo(ctx context.Context, companyID, branchID int64) *mock.Call {
	return e.mock.On("GenerateOrderNo", ctx, companyID, branchID)
}

func (e *MockOrderRepositoryExpectation) GetByIDForUpdate(ctx context.Context, companyID, id int64) *mock.Call {
	return e.mock.On("GetByIDForUpdate", ctx, companyID, id)
}

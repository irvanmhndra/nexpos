package mocks

import (
	"context"
	"time"

	"github.com/irvanmhndra/nexpos-api/internal/model"
	"github.com/stretchr/testify/mock"
)

// MockPaymentRepository is a mock implementation of repository.PaymentRepository.
type MockPaymentRepository struct {
	mock.Mock
}

func NewMockPaymentRepository(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockPaymentRepository {
	m := &MockPaymentRepository{}
	m.Test(t)
	t.Cleanup(func() { m.AssertExpectations(t) })
	return m
}

func (m *MockPaymentRepository) EXPECT() *MockPaymentRepositoryExpectation {
	return &MockPaymentRepositoryExpectation{mock: m}
}

func (m *MockPaymentRepository) GetByIDForUpdate(ctx context.Context, id int64) (*model.Payment, error) {
	ret := m.Called(ctx, id)
	var r0 *model.Payment
	if ret.Get(0) != nil {
		r0 = ret.Get(0).(*model.Payment)
	}
	return r0, ret.Error(1)
}

func (m *MockPaymentRepository) Create(ctx context.Context, payment *model.Payment) error {
	ret := m.Called(ctx, payment)
	return ret.Error(0)
}

func (m *MockPaymentRepository) GetByID(ctx context.Context, id int64) (*model.Payment, error) {
	ret := m.Called(ctx, id)
	var r0 *model.Payment
	if ret.Get(0) != nil {
		r0 = ret.Get(0).(*model.Payment)
	}
	return r0, ret.Error(1)
}

func (m *MockPaymentRepository) GetByOrderID(ctx context.Context, orderID int64) ([]*model.Payment, error) {
	ret := m.Called(ctx, orderID)
	var r0 []*model.Payment
	if ret.Get(0) != nil {
		r0 = ret.Get(0).([]*model.Payment)
	}
	return r0, ret.Error(1)
}

func (m *MockPaymentRepository) Update(ctx context.Context, payment *model.Payment) error {
	ret := m.Called(ctx, payment)
	return ret.Error(0)
}

func (m *MockPaymentRepository) DeleteByOrderID(ctx context.Context, orderID int64) error {
	ret := m.Called(ctx, orderID)
	return ret.Error(0)
}

func (m *MockPaymentRepository) GetTotalPaidByOrderID(ctx context.Context, orderID int64) (float64, error) {
	ret := m.Called(ctx, orderID)
	return ret.Get(0).(float64), ret.Error(1)
}

func (m *MockPaymentRepository) GetTotalRefundedByOrderID(ctx context.Context, orderID int64) (float64, error) {
	ret := m.Called(ctx, orderID)
	return ret.Get(0).(float64), ret.Error(1)
}

func (m *MockPaymentRepository) GetCashTotalByPeriod(ctx context.Context, branchID int64, from, to time.Time) (float64, error) {
	ret := m.Called(ctx, branchID, from, to)
	return ret.Get(0).(float64), ret.Error(1)
}

// MockPaymentRepositoryExpectation is the expectation builder for MockPaymentRepository.
type MockPaymentRepositoryExpectation struct {
	mock *MockPaymentRepository
}

func (e *MockPaymentRepositoryExpectation) Create(ctx context.Context, payment interface{}) *mock.Call {
	return e.mock.On("Create", ctx, payment)
}

func (e *MockPaymentRepositoryExpectation) GetByID(ctx context.Context, id int64) *mock.Call {
	return e.mock.On("GetByID", ctx, id)
}

func (e *MockPaymentRepositoryExpectation) GetByOrderID(ctx context.Context, orderID int64) *mock.Call {
	return e.mock.On("GetByOrderID", ctx, orderID)
}

func (e *MockPaymentRepositoryExpectation) Update(ctx context.Context, payment interface{}) *mock.Call {
	return e.mock.On("Update", ctx, payment)
}

func (e *MockPaymentRepositoryExpectation) DeleteByOrderID(ctx context.Context, orderID int64) *mock.Call {
	return e.mock.On("DeleteByOrderID", ctx, orderID)
}

func (e *MockPaymentRepositoryExpectation) GetTotalPaidByOrderID(ctx context.Context, orderID int64) *mock.Call {
	return e.mock.On("GetTotalPaidByOrderID", ctx, orderID)
}

func (e *MockPaymentRepositoryExpectation) GetTotalRefundedByOrderID(ctx context.Context, orderID int64) *mock.Call {
	return e.mock.On("GetTotalRefundedByOrderID", ctx, orderID)
}

func (e *MockPaymentRepositoryExpectation) GetCashTotalByPeriod(ctx context.Context, branchID int64, from, to time.Time) *mock.Call {
	return e.mock.On("GetCashTotalByPeriod", ctx, branchID, from, to)
}

func (e *MockPaymentRepositoryExpectation) GetByIDForUpdate(ctx context.Context, id int64) *mock.Call {
	return e.mock.On("GetByIDForUpdate", ctx, id)
}

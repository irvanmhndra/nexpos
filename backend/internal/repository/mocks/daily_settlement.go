package mocks

import (
	"context"

	"github.com/irvanmhndra/nexpos-api/internal/model"
	"github.com/irvanmhndra/nexpos-api/internal/repository"
	"github.com/stretchr/testify/mock"
)

type MockDailySettlementRepository struct {
	mock.Mock
}

func NewMockDailySettlementRepository(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockDailySettlementRepository {
	m := &MockDailySettlementRepository{}
	m.Test(t)
	t.Cleanup(func() { m.AssertExpectations(t) })
	return m
}

func (m *MockDailySettlementRepository) EXPECT() *MockDailySettlementRepositoryExpectation {
	return &MockDailySettlementRepositoryExpectation{mock: m}
}

func (m *MockDailySettlementRepository) Create(ctx context.Context, s *model.DailySettlement) error {
	return m.Called(ctx, s).Error(0)
}

func (m *MockDailySettlementRepository) GetByID(ctx context.Context, companyID, id int64) (*model.DailySettlement, error) {
	ret := m.Called(ctx, companyID, id)
	var r0 *model.DailySettlement
	if ret.Get(0) != nil {
		r0 = ret.Get(0).(*model.DailySettlement)
	}
	return r0, ret.Error(1)
}

func (m *MockDailySettlementRepository) GetByBranchAndDate(ctx context.Context, companyID, branchID int64, date string) (*model.DailySettlement, error) {
	ret := m.Called(ctx, companyID, branchID, date)
	var r0 *model.DailySettlement
	if ret.Get(0) != nil {
		r0 = ret.Get(0).(*model.DailySettlement)
	}
	return r0, ret.Error(1)
}

func (m *MockDailySettlementRepository) List(ctx context.Context, companyID int64, params *repository.DailySettlementListParams) ([]*model.DailySettlement, int, error) {
	ret := m.Called(ctx, companyID, params)
	var r0 []*model.DailySettlement
	if ret.Get(0) != nil {
		r0 = ret.Get(0).([]*model.DailySettlement)
	}
	return r0, ret.Int(1), ret.Error(2)
}

func (m *MockDailySettlementRepository) Update(ctx context.Context, s *model.DailySettlement) error {
	return m.Called(ctx, s).Error(0)
}

func (m *MockDailySettlementRepository) CreateItem(ctx context.Context, item *model.DailySettlementItem) error {
	return m.Called(ctx, item).Error(0)
}

func (m *MockDailySettlementRepository) GetItems(ctx context.Context, settlementID int64) ([]*model.DailySettlementItem, error) {
	ret := m.Called(ctx, settlementID)
	var r0 []*model.DailySettlementItem
	if ret.Get(0) != nil {
		r0 = ret.Get(0).([]*model.DailySettlementItem)
	}
	return r0, ret.Error(1)
}

func (m *MockDailySettlementRepository) GetItem(ctx context.Context, settlementID, itemID int64) (*model.DailySettlementItem, error) {
	ret := m.Called(ctx, settlementID, itemID)
	var r0 *model.DailySettlementItem
	if ret.Get(0) != nil {
		r0 = ret.Get(0).(*model.DailySettlementItem)
	}
	return r0, ret.Error(1)
}

func (m *MockDailySettlementRepository) UpdateItem(ctx context.Context, item *model.DailySettlementItem) error {
	return m.Called(ctx, item).Error(0)
}

func (m *MockDailySettlementRepository) GetPaymentBreakdown(ctx context.Context, companyID, branchID int64, date string) ([]*repository.PaymentMethodTotals, error) {
	ret := m.Called(ctx, companyID, branchID, date)
	var r0 []*repository.PaymentMethodTotals
	if ret.Get(0) != nil {
		r0 = ret.Get(0).([]*repository.PaymentMethodTotals)
	}
	return r0, ret.Error(1)
}

func (m *MockDailySettlementRepository) GetExpensesTotal(ctx context.Context, companyID, branchID int64, date string) (float64, error) {
	ret := m.Called(ctx, companyID, branchID, date)
	return ret.Get(0).(float64), ret.Error(1)
}

type MockDailySettlementRepositoryExpectation struct {
	mock *MockDailySettlementRepository
}

func (e *MockDailySettlementRepositoryExpectation) Create(ctx context.Context, s interface{}) *mock.Call {
	return e.mock.On("Create", ctx, s)
}

func (e *MockDailySettlementRepositoryExpectation) GetByID(ctx context.Context, companyID, id int64) *mock.Call {
	return e.mock.On("GetByID", ctx, companyID, id)
}

func (e *MockDailySettlementRepositoryExpectation) GetByBranchAndDate(ctx context.Context, companyID, branchID int64, date string) *mock.Call {
	return e.mock.On("GetByBranchAndDate", ctx, companyID, branchID, date)
}

func (e *MockDailySettlementRepositoryExpectation) List(ctx context.Context, companyID int64, params interface{}) *mock.Call {
	return e.mock.On("List", ctx, companyID, params)
}

func (e *MockDailySettlementRepositoryExpectation) Update(ctx context.Context, s interface{}) *mock.Call {
	return e.mock.On("Update", ctx, s)
}

func (e *MockDailySettlementRepositoryExpectation) CreateItem(ctx context.Context, item interface{}) *mock.Call {
	return e.mock.On("CreateItem", ctx, item)
}

func (e *MockDailySettlementRepositoryExpectation) GetItems(ctx context.Context, settlementID int64) *mock.Call {
	return e.mock.On("GetItems", ctx, settlementID)
}

func (e *MockDailySettlementRepositoryExpectation) GetItem(ctx context.Context, settlementID, itemID int64) *mock.Call {
	return e.mock.On("GetItem", ctx, settlementID, itemID)
}

func (e *MockDailySettlementRepositoryExpectation) UpdateItem(ctx context.Context, item interface{}) *mock.Call {
	return e.mock.On("UpdateItem", ctx, item)
}

func (e *MockDailySettlementRepositoryExpectation) GetPaymentBreakdown(ctx context.Context, companyID, branchID int64, date string) *mock.Call {
	return e.mock.On("GetPaymentBreakdown", ctx, companyID, branchID, date)
}

func (e *MockDailySettlementRepositoryExpectation) GetExpensesTotal(ctx context.Context, companyID, branchID int64, date string) *mock.Call {
	return e.mock.On("GetExpensesTotal", ctx, companyID, branchID, date)
}

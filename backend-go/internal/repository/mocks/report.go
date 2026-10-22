package mocks

import (
	"context"

	"github.com/irvanmhndra/nexpos-api/internal/repository"
	"github.com/stretchr/testify/mock"
)

// MockReportRepository is a mock implementation of repository.ReportRepository.
type MockReportRepository struct {
	mock.Mock
}

func NewMockReportRepository(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockReportRepository {
	m := &MockReportRepository{}
	m.Test(t)
	t.Cleanup(func() { m.AssertExpectations(t) })
	return m
}

func (m *MockReportRepository) EXPECT() *MockReportRepositoryExpectation {
	return &MockReportRepositoryExpectation{mock: m}
}

func (m *MockReportRepository) GetSummary(ctx context.Context, companyID int64, dateFrom, dateTo string) (*repository.ReportSummary, error) {
	ret := m.Called(ctx, companyID, dateFrom, dateTo)
	var r0 *repository.ReportSummary
	if ret.Get(0) != nil {
		r0 = ret.Get(0).(*repository.ReportSummary)
	}
	return r0, ret.Error(1)
}

func (m *MockReportRepository) GetSalesTrend(ctx context.Context, companyID int64, dateFrom, dateTo string) ([]*repository.SalesTrendItem, error) {
	ret := m.Called(ctx, companyID, dateFrom, dateTo)
	var r0 []*repository.SalesTrendItem
	if ret.Get(0) != nil {
		r0 = ret.Get(0).([]*repository.SalesTrendItem)
	}
	return r0, ret.Error(1)
}

func (m *MockReportRepository) GetTopProducts(ctx context.Context, companyID int64, dateFrom, dateTo string, limit int) ([]*repository.TopProductItem, error) {
	ret := m.Called(ctx, companyID, dateFrom, dateTo, limit)
	var r0 []*repository.TopProductItem
	if ret.Get(0) != nil {
		r0 = ret.Get(0).([]*repository.TopProductItem)
	}
	return r0, ret.Error(1)
}

func (m *MockReportRepository) GetCategoryRevenue(ctx context.Context, companyID int64, dateFrom, dateTo string) ([]*repository.CategoryRevenueItem, error) {
	ret := m.Called(ctx, companyID, dateFrom, dateTo)
	var r0 []*repository.CategoryRevenueItem
	if ret.Get(0) != nil {
		r0 = ret.Get(0).([]*repository.CategoryRevenueItem)
	}
	return r0, ret.Error(1)
}

func (m *MockReportRepository) GetPaymentMethods(ctx context.Context, companyID int64, dateFrom, dateTo string) ([]*repository.PaymentMethodItem, error) {
	ret := m.Called(ctx, companyID, dateFrom, dateTo)
	var r0 []*repository.PaymentMethodItem
	if ret.Get(0) != nil {
		r0 = ret.Get(0).([]*repository.PaymentMethodItem)
	}
	return r0, ret.Error(1)
}

func (m *MockReportRepository) GetHourlySales(ctx context.Context, companyID int64, dateFrom, dateTo string) ([]*repository.HourlySalesItem, error) {
	ret := m.Called(ctx, companyID, dateFrom, dateTo)
	var r0 []*repository.HourlySalesItem
	if ret.Get(0) != nil {
		r0 = ret.Get(0).([]*repository.HourlySalesItem)
	}
	return r0, ret.Error(1)
}

func (m *MockReportRepository) GetNewCustomersCount(ctx context.Context, companyID int64, dateFrom, dateTo string) (int, error) {
	ret := m.Called(ctx, companyID, dateFrom, dateTo)
	return ret.Int(0), ret.Error(1)
}

// MockReportRepositoryExpectation is the expectation builder for MockReportRepository.
type MockReportRepositoryExpectation struct {
	mock *MockReportRepository
}

func (e *MockReportRepositoryExpectation) GetSummary(ctx context.Context, companyID int64, dateFrom, dateTo string) *mock.Call {
	return e.mock.On("GetSummary", ctx, companyID, dateFrom, dateTo)
}

func (e *MockReportRepositoryExpectation) GetSalesTrend(ctx context.Context, companyID int64, dateFrom, dateTo string) *mock.Call {
	return e.mock.On("GetSalesTrend", ctx, companyID, dateFrom, dateTo)
}

func (e *MockReportRepositoryExpectation) GetTopProducts(ctx context.Context, companyID int64, dateFrom, dateTo string, limit int) *mock.Call {
	return e.mock.On("GetTopProducts", ctx, companyID, dateFrom, dateTo, limit)
}

func (e *MockReportRepositoryExpectation) GetCategoryRevenue(ctx context.Context, companyID int64, dateFrom, dateTo string) *mock.Call {
	return e.mock.On("GetCategoryRevenue", ctx, companyID, dateFrom, dateTo)
}

func (e *MockReportRepositoryExpectation) GetPaymentMethods(ctx context.Context, companyID int64, dateFrom, dateTo string) *mock.Call {
	return e.mock.On("GetPaymentMethods", ctx, companyID, dateFrom, dateTo)
}

func (e *MockReportRepositoryExpectation) GetHourlySales(ctx context.Context, companyID int64, dateFrom, dateTo string) *mock.Call {
	return e.mock.On("GetHourlySales", ctx, companyID, dateFrom, dateTo)
}

func (e *MockReportRepositoryExpectation) GetNewCustomersCount(ctx context.Context, companyID int64, dateFrom, dateTo string) *mock.Call {
	return e.mock.On("GetNewCustomersCount", ctx, companyID, dateFrom, dateTo)
}

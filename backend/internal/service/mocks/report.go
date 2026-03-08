package mocks

import (
	"context"

	"github.com/irvanmhndra/nexpos-api/internal/dto"
	"github.com/stretchr/testify/mock"
)

// MockReportService is a mock implementation of service.ReportServiceInterface
type MockReportService struct {
	mock.Mock
}

func (m *MockReportService) GetSummary(ctx context.Context, companyID int64, req dto.ReportSummaryRequest) (*dto.ReportSummaryResponse, error) {
	args := m.Called(ctx, companyID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.ReportSummaryResponse), args.Error(1)
}

func (m *MockReportService) GetSalesTrend(ctx context.Context, companyID int64, req dto.SalesTrendRequest) ([]*dto.SalesTrendResponse, error) {
	args := m.Called(ctx, companyID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*dto.SalesTrendResponse), args.Error(1)
}

func (m *MockReportService) GetTopProducts(ctx context.Context, companyID int64, req dto.TopProductsRequest) ([]*dto.TopProductResponse, error) {
	args := m.Called(ctx, companyID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*dto.TopProductResponse), args.Error(1)
}

func (m *MockReportService) GetCategoryRevenue(ctx context.Context, companyID int64, req dto.CategoryRevenueRequest) ([]*dto.CategoryRevenueResponse, error) {
	args := m.Called(ctx, companyID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*dto.CategoryRevenueResponse), args.Error(1)
}

func (m *MockReportService) GetPaymentMethods(ctx context.Context, companyID int64, req dto.PaymentMethodsRequest) ([]*dto.PaymentMethodResponse, error) {
	args := m.Called(ctx, companyID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*dto.PaymentMethodResponse), args.Error(1)
}

func (m *MockReportService) GetHourlySales(ctx context.Context, companyID int64, req dto.HourlySalesRequest) ([]*dto.HourlySalesResponse, error) {
	args := m.Called(ctx, companyID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*dto.HourlySalesResponse), args.Error(1)
}

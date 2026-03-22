package service

import (
	"context"
	"errors"
	"testing"

	"github.com/irvanmhndra/nexpos-api/internal/repository"
	repoMocks "github.com/irvanmhndra/nexpos-api/internal/repository/mocks"
	"github.com/irvanmhndra/nexpos-api/pkg/apperror"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupReportTest(t *testing.T) (*ReportService, *repoMocks.MockReportRepository) {
	t.Helper()
	mockRepo := repoMocks.NewMockReportRepository(t)
	svc := NewReportService(mockRepo)
	return svc, mockRepo
}

func TestNewReportService(t *testing.T) {
	mockRepo := repoMocks.NewMockReportRepository(t)
	svc := NewReportService(mockRepo)
	assert.NotNil(t, svc)
}

// GetSummary
func TestReportService_GetSummary_Success(t *testing.T) {
	svc, mockRepo := setupReportTest(t)
	ctx := context.Background()

	summary := &repository.ReportSummary{
		TotalRevenue:    10000,
		TotalOrders:     100,
		TotalItemsSold:  500,
		TotalCustomers:  50,
		CompletedOrders: 90,
		CancelledOrders: 5,
		PendingOrders:   5,
		TotalDiscount:   500,
		TotalTax:        1000,
		TotalCOGS:       6000,
	}

	mockRepo.EXPECT().GetSummary(ctx, int64(1), "2026-01-01", "2026-01-31").Return(summary, nil).Once()
	mockRepo.EXPECT().GetNewCustomersCount(ctx, int64(1), "2026-01-01", "2026-01-31").Return(10, nil).Once()

	result, err := svc.GetSummary(ctx, 1, "2026-01-01", "2026-01-31")
	require.NoError(t, err)
	assert.Equal(t, 10000.0, result.TotalRevenue)
	assert.Equal(t, 100, result.TotalOrders)
	assert.Equal(t, 100.0, result.AvgOrderValue)
	assert.Equal(t, 4000.0, result.GrossProfit)
	assert.Equal(t, 40.0, result.GrossProfitMargin)
	assert.Equal(t, 10, result.NewCustomers)
}

func TestReportService_GetSummary_ZeroOrders(t *testing.T) {
	svc, mockRepo := setupReportTest(t)
	ctx := context.Background()

	summary := &repository.ReportSummary{TotalOrders: 0, TotalRevenue: 0}
	mockRepo.EXPECT().GetSummary(ctx, int64(1), "2026-01-01", "2026-01-31").Return(summary, nil).Once()
	mockRepo.EXPECT().GetNewCustomersCount(ctx, int64(1), "2026-01-01", "2026-01-31").Return(0, nil).Once()

	result, err := svc.GetSummary(ctx, 1, "2026-01-01", "2026-01-31")
	require.NoError(t, err)
	assert.Equal(t, 0.0, result.AvgOrderValue)
	assert.Equal(t, 0.0, result.GrossProfitMargin)
}

func TestReportService_GetSummary_RepoError(t *testing.T) {
	svc, mockRepo := setupReportTest(t)
	ctx := context.Background()

	mockRepo.EXPECT().GetSummary(ctx, int64(1), "2026-01-01", "2026-01-31").Return(nil, errors.New("db error")).Once()

	result, err := svc.GetSummary(ctx, 1, "2026-01-01", "2026-01-31")
	require.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, apperror.IsInternalError(err))
}

func TestReportService_GetSummary_NewCustomersError_NonCritical(t *testing.T) {
	svc, mockRepo := setupReportTest(t)
	ctx := context.Background()

	summary := &repository.ReportSummary{TotalRevenue: 1000, TotalOrders: 10}
	mockRepo.EXPECT().GetSummary(ctx, int64(1), "2026-01-01", "2026-01-31").Return(summary, nil).Once()
	// NewCustomers error is non-critical, should still return 0
	mockRepo.EXPECT().GetNewCustomersCount(ctx, int64(1), "2026-01-01", "2026-01-31").Return(0, errors.New("db error")).Once()

	result, err := svc.GetSummary(ctx, 1, "2026-01-01", "2026-01-31")
	require.NoError(t, err)
	assert.Equal(t, 0, result.NewCustomers)
}

// GetSalesTrend
func TestReportService_GetSalesTrend_Success(t *testing.T) {
	svc, mockRepo := setupReportTest(t)
	ctx := context.Background()

	items := []*repository.SalesTrendItem{
		{Date: "2026-01-01", Sales: 1000, Orders: 10, Items: 50},
		{Date: "2026-01-02", Sales: 2000, Orders: 20, Items: 100},
	}
	mockRepo.EXPECT().GetSalesTrend(ctx, int64(1), "2026-01-01", "2026-01-31").Return(items, nil).Once()

	result, err := svc.GetSalesTrend(ctx, 1, "2026-01-01", "2026-01-31")
	require.NoError(t, err)
	assert.Len(t, result.Data, 2)
	assert.Equal(t, "2026-01-01", result.Data[0].Date)
}

func TestReportService_GetSalesTrend_RepoError(t *testing.T) {
	svc, mockRepo := setupReportTest(t)
	ctx := context.Background()

	mockRepo.EXPECT().GetSalesTrend(ctx, int64(1), "2026-01-01", "2026-01-31").Return(nil, errors.New("db error")).Once()

	result, err := svc.GetSalesTrend(ctx, 1, "2026-01-01", "2026-01-31")
	require.Error(t, err)
	assert.Nil(t, result)
}

// GetTopProducts
func TestReportService_GetTopProducts_Success(t *testing.T) {
	svc, mockRepo := setupReportTest(t)
	ctx := context.Background()

	items := []*repository.TopProductItem{
		{ProductID: 1, ProductName: "Product A", TotalSold: 100, TotalAmount: 5000},
	}
	mockRepo.EXPECT().GetTopProducts(ctx, int64(1), "2026-01-01", "2026-01-31", 10).Return(items, nil).Once()

	result, err := svc.GetTopProducts(ctx, 1, "2026-01-01", "2026-01-31", 10)
	require.NoError(t, err)
	assert.Len(t, result.Data, 1)
}

func TestReportService_GetTopProducts_DefaultLimit(t *testing.T) {
	svc, mockRepo := setupReportTest(t)
	ctx := context.Background()

	mockRepo.EXPECT().GetTopProducts(ctx, int64(1), "2026-01-01", "2026-01-31", 10).Return([]*repository.TopProductItem{}, nil).Once()

	result, err := svc.GetTopProducts(ctx, 1, "2026-01-01", "2026-01-31", 0) // 0 → 10
	require.NoError(t, err)
	assert.Len(t, result.Data, 0)
}

func TestReportService_GetTopProducts_RepoError(t *testing.T) {
	svc, mockRepo := setupReportTest(t)
	ctx := context.Background()

	mockRepo.EXPECT().GetTopProducts(ctx, int64(1), "2026-01-01", "2026-01-31", 10).Return(nil, errors.New("db error")).Once()

	result, err := svc.GetTopProducts(ctx, 1, "2026-01-01", "2026-01-31", 10)
	require.Error(t, err)
	assert.Nil(t, result)
}

// GetCategoryRevenue
func TestReportService_GetCategoryRevenue_Success(t *testing.T) {
	svc, mockRepo := setupReportTest(t)
	ctx := context.Background()

	items := []*repository.CategoryRevenueItem{{CategoryID: 1, CategoryName: "Food", TotalAmount: 5000, OrderCount: 50}}
	mockRepo.EXPECT().GetCategoryRevenue(ctx, int64(1), "2026-01-01", "2026-01-31").Return(items, nil).Once()

	result, err := svc.GetCategoryRevenue(ctx, 1, "2026-01-01", "2026-01-31")
	require.NoError(t, err)
	assert.Len(t, result.Data, 1)
}

func TestReportService_GetCategoryRevenue_RepoError(t *testing.T) {
	svc, mockRepo := setupReportTest(t)
	ctx := context.Background()

	mockRepo.EXPECT().GetCategoryRevenue(ctx, int64(1), "2026-01-01", "2026-01-31").Return(nil, errors.New("db error")).Once()

	result, err := svc.GetCategoryRevenue(ctx, 1, "2026-01-01", "2026-01-31")
	require.Error(t, err)
	assert.Nil(t, result)
}

// GetPaymentMethods
func TestReportService_GetPaymentMethods_Success(t *testing.T) {
	svc, mockRepo := setupReportTest(t)
	ctx := context.Background()

	items := []*repository.PaymentMethodItem{{Method: "cash", TotalAmount: 5000, Count: 50}}
	mockRepo.EXPECT().GetPaymentMethods(ctx, int64(1), "2026-01-01", "2026-01-31").Return(items, nil).Once()

	result, err := svc.GetPaymentMethods(ctx, 1, "2026-01-01", "2026-01-31")
	require.NoError(t, err)
	assert.Len(t, result.Data, 1)
}

func TestReportService_GetPaymentMethods_RepoError(t *testing.T) {
	svc, mockRepo := setupReportTest(t)
	ctx := context.Background()

	mockRepo.EXPECT().GetPaymentMethods(ctx, int64(1), "2026-01-01", "2026-01-31").Return(nil, errors.New("db error")).Once()

	result, err := svc.GetPaymentMethods(ctx, 1, "2026-01-01", "2026-01-31")
	require.Error(t, err)
	assert.Nil(t, result)
}

// GetHourlySales
func TestReportService_GetHourlySales_Success(t *testing.T) {
	svc, mockRepo := setupReportTest(t)
	ctx := context.Background()

	items := []*repository.HourlySalesItem{{Hour: 12, TotalAmount: 3000, OrderCount: 30}}
	mockRepo.EXPECT().GetHourlySales(ctx, int64(1), "2026-01-01", "2026-01-31").Return(items, nil).Once()

	result, err := svc.GetHourlySales(ctx, 1, "2026-01-01", "2026-01-31")
	require.NoError(t, err)
	assert.Len(t, result.Data, 1)
	assert.Equal(t, 12, result.Data[0].Hour)
}

func TestReportService_GetHourlySales_RepoError(t *testing.T) {
	svc, mockRepo := setupReportTest(t)
	ctx := context.Background()

	mockRepo.EXPECT().GetHourlySales(ctx, int64(1), "2026-01-01", "2026-01-31").Return(nil, errors.New("db error")).Once()

	result, err := svc.GetHourlySales(ctx, 1, "2026-01-01", "2026-01-31")
	require.Error(t, err)
	assert.Nil(t, result)
}

package service

import (
	"context"

	"github.com/irvanmhndra/pos-core-api/internal/dto"
	"github.com/irvanmhndra/pos-core-api/internal/repository"
	"github.com/irvanmhndra/pos-core-api/pkg/apperror"
)

type ReportService struct {
	reportRepo repository.ReportRepository
}

func NewReportService(reportRepo repository.ReportRepository) *ReportService {
	return &ReportService{
		reportRepo: reportRepo,
	}
}

func (s *ReportService) GetSummary(ctx context.Context, companyID int64, dateFrom, dateTo string) (*dto.ReportSummaryResponse, error) {
	summary, err := s.reportRepo.GetSummary(ctx, companyID, dateFrom, dateTo)
	if err != nil {
		return nil, apperror.InternalError(err)
	}

	// Calculate derived metrics
	avgOrderValue := float64(0)
	if summary.TotalOrders > 0 {
		avgOrderValue = summary.TotalRevenue / float64(summary.TotalOrders)
	}

	grossProfit := summary.TotalRevenue - summary.TotalCOGS
	grossProfitMargin := float64(0)
	if summary.TotalRevenue > 0 {
		grossProfitMargin = (grossProfit / summary.TotalRevenue) * 100
	}

	// Get new customers count
	newCustomers, err := s.reportRepo.GetNewCustomersCount(ctx, companyID, dateFrom, dateTo)
	if err != nil {
		newCustomers = 0 // Non-critical, continue
	}

	return &dto.ReportSummaryResponse{
		TotalRevenue:      summary.TotalRevenue,
		TotalOrders:       summary.TotalOrders,
		TotalItemsSold:    summary.TotalItemsSold,
		AvgOrderValue:     avgOrderValue,
		TotalCustomers:    summary.TotalCustomers,
		NewCustomers:      newCustomers,
		CompletedOrders:   summary.CompletedOrders,
		CancelledOrders:   summary.CancelledOrders,
		PendingOrders:     summary.PendingOrders,
		TotalDiscount:     summary.TotalDiscount,
		TotalTax:          summary.TotalTax,
		GrossProfit:       grossProfit,
		GrossProfitMargin: grossProfitMargin,
	}, nil
}

func (s *ReportService) GetSalesTrend(ctx context.Context, companyID int64, dateFrom, dateTo string) (*dto.SalesTrendResponse, error) {
	items, err := s.reportRepo.GetSalesTrend(ctx, companyID, dateFrom, dateTo)
	if err != nil {
		return nil, apperror.InternalError(err)
	}

	response := &dto.SalesTrendResponse{
		Data: make([]*dto.SalesTrendItem, 0, len(items)),
	}

	for _, item := range items {
		response.Data = append(response.Data, &dto.SalesTrendItem{
			Date:   item.Date,
			Sales:  item.Sales,
			Orders: item.Orders,
			Items:  item.Items,
		})
	}

	return response, nil
}

func (s *ReportService) GetTopProducts(ctx context.Context, companyID int64, dateFrom, dateTo string, limit int) (*dto.TopProductsResponse, error) {
	if limit <= 0 {
		limit = 10
	}

	items, err := s.reportRepo.GetTopProducts(ctx, companyID, dateFrom, dateTo, limit)
	if err != nil {
		return nil, apperror.InternalError(err)
	}

	response := &dto.TopProductsResponse{
		Data: make([]*dto.TopProductItem, 0, len(items)),
	}

	for _, item := range items {
		response.Data = append(response.Data, &dto.TopProductItem{
			ProductID:   item.ProductID,
			ProductName: item.ProductName,
			TotalSold:   item.TotalSold,
			TotalAmount: item.TotalAmount,
		})
	}

	return response, nil
}

func (s *ReportService) GetCategoryRevenue(ctx context.Context, companyID int64, dateFrom, dateTo string) (*dto.CategoryRevenueResponse, error) {
	items, err := s.reportRepo.GetCategoryRevenue(ctx, companyID, dateFrom, dateTo)
	if err != nil {
		return nil, apperror.InternalError(err)
	}

	response := &dto.CategoryRevenueResponse{
		Data: make([]*dto.CategoryRevenueItem, 0, len(items)),
	}

	for _, item := range items {
		response.Data = append(response.Data, &dto.CategoryRevenueItem{
			CategoryID:   item.CategoryID,
			CategoryName: item.CategoryName,
			TotalAmount:  item.TotalAmount,
			OrderCount:   item.OrderCount,
		})
	}

	return response, nil
}

func (s *ReportService) GetPaymentMethods(ctx context.Context, companyID int64, dateFrom, dateTo string) (*dto.PaymentMethodResponse, error) {
	items, err := s.reportRepo.GetPaymentMethods(ctx, companyID, dateFrom, dateTo)
	if err != nil {
		return nil, apperror.InternalError(err)
	}

	response := &dto.PaymentMethodResponse{
		Data: make([]*dto.PaymentMethodItem, 0, len(items)),
	}

	for _, item := range items {
		response.Data = append(response.Data, &dto.PaymentMethodItem{
			Method:      item.Method,
			TotalAmount: item.TotalAmount,
			Count:       item.Count,
		})
	}

	return response, nil
}

func (s *ReportService) GetHourlySales(ctx context.Context, companyID int64, dateFrom, dateTo string) (*dto.HourlySalesResponse, error) {
	items, err := s.reportRepo.GetHourlySales(ctx, companyID, dateFrom, dateTo)
	if err != nil {
		return nil, apperror.InternalError(err)
	}

	response := &dto.HourlySalesResponse{
		Data: make([]*dto.HourlySalesItem, 0, len(items)),
	}

	for _, item := range items {
		response.Data = append(response.Data, &dto.HourlySalesItem{
			Hour:        item.Hour,
			TotalAmount: item.TotalAmount,
			OrderCount:  item.OrderCount,
		})
	}

	return response, nil
}

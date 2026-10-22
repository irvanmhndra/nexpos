package service

import (
	"context"

	"github.com/irvanmhndra/nexpos-api/internal/dto"
	"github.com/irvanmhndra/nexpos-api/internal/model"
	"github.com/irvanmhndra/nexpos-api/internal/repository"
	"github.com/irvanmhndra/nexpos-api/pkg/apperror"
)

type CompanySettingsService struct {
	settingsRepo repository.CompanySettingsRepository
}

func NewCompanySettingsService(settingsRepo repository.CompanySettingsRepository) *CompanySettingsService {
	return &CompanySettingsService{settingsRepo: settingsRepo}
}

func (s *CompanySettingsService) Get(ctx context.Context, companyID int64) (*dto.CompanySettingsResponse, error) {
	settings, err := s.settingsRepo.GetByCompanyID(ctx, companyID)
	if err != nil {
		return nil, apperror.InternalError(err)
	}
	return s.toResponse(settings), nil
}

func (s *CompanySettingsService) Update(ctx context.Context, companyID int64, req dto.UpdateCompanySettingsRequest) (*dto.CompanySettingsResponse, error) {
	// Fetch current (returns defaults if not found)
	settings, err := s.settingsRepo.GetByCompanyID(ctx, companyID)
	if err != nil {
		return nil, apperror.InternalError(err)
	}

	// Apply partial updates
	if req.TaxEnabled != nil {
		settings.TaxEnabled = *req.TaxEnabled
	}
	if req.TaxRate != nil {
		settings.TaxRate = *req.TaxRate
	}
	if req.TaxInclusive != nil {
		settings.TaxInclusive = *req.TaxInclusive
	}
	if req.RoundingEnabled != nil {
		settings.RoundingEnabled = *req.RoundingEnabled
	}
	if req.RoundingAmount != nil {
		settings.RoundingAmount = *req.RoundingAmount
	}
	if req.AutoCompleteCounterOrders != nil {
		settings.AutoCompleteCounterOrders = *req.AutoCompleteCounterOrders
	}
	if req.RequireCustomerForDelivery != nil {
		settings.RequireCustomerForDelivery = *req.RequireCustomerForDelivery
	}
	if req.ReceiptHeader != nil {
		settings.ReceiptHeader = req.ReceiptHeader
	}
	if req.ReceiptFooter != nil {
		settings.ReceiptFooter = req.ReceiptFooter
	}
	if req.ShowTaxOnReceipt != nil {
		settings.ShowTaxOnReceipt = *req.ShowTaxOnReceipt
	}
	if req.OfflineModeEnabled != nil {
		settings.OfflineModeEnabled = *req.OfflineModeEnabled
	}
	if req.MaxOfflineDays != nil {
		settings.MaxOfflineDays = *req.MaxOfflineDays
	}

	settings.CompanyID = companyID

	if err := s.settingsRepo.Upsert(ctx, settings); err != nil {
		return nil, apperror.InternalError(err)
	}

	return s.toResponse(settings), nil
}

func (s *CompanySettingsService) toResponse(m *model.CompanySettings) *dto.CompanySettingsResponse {
	return &dto.CompanySettingsResponse{
		ID:                         m.ID,
		CompanyID:                  m.CompanyID,
		TaxEnabled:                 m.TaxEnabled,
		TaxRate:                    m.TaxRate,
		TaxInclusive:               m.TaxInclusive,
		RoundingEnabled:            m.RoundingEnabled,
		RoundingAmount:             m.RoundingAmount,
		AutoCompleteCounterOrders:  m.AutoCompleteCounterOrders,
		RequireCustomerForDelivery: m.RequireCustomerForDelivery,
		ReceiptHeader:              m.ReceiptHeader,
		ReceiptFooter:              m.ReceiptFooter,
		ShowTaxOnReceipt:           m.ShowTaxOnReceipt,
		OfflineModeEnabled:         m.OfflineModeEnabled,
		MaxOfflineDays:             m.MaxOfflineDays,
		CreatedAt:                  m.CreatedAt,
		UpdatedAt:                  m.UpdatedAt,
	}
}

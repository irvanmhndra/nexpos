package mocks

import (
	"context"

	"github.com/irvanmhndra/nexpos-api/internal/dto"
	"github.com/stretchr/testify/mock"
)

// MockCompanySettingsService is a mock implementation of service.CompanySettingsServiceInterface
type MockCompanySettingsService struct {
	mock.Mock
}

func (m *MockCompanySettingsService) Get(ctx context.Context, companyID int64) (*dto.CompanySettingsResponse, error) {
	args := m.Called(ctx, companyID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.CompanySettingsResponse), args.Error(1)
}

func (m *MockCompanySettingsService) Update(ctx context.Context, companyID int64, req dto.UpdateCompanySettingsRequest) (*dto.CompanySettingsResponse, error) {
	args := m.Called(ctx, companyID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.CompanySettingsResponse), args.Error(1)
}

package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/irvanmhndra/nexpos-api/internal/dto"
	"github.com/irvanmhndra/nexpos-api/internal/model"
	repoMocks "github.com/irvanmhndra/nexpos-api/internal/repository/mocks"
	"github.com/irvanmhndra/nexpos-api/pkg/apperror"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func setupCompanySettingsTest(t *testing.T) (*CompanySettingsService, *repoMocks.MockCompanySettingsRepository) {
	t.Helper()
	mockRepo := repoMocks.NewMockCompanySettingsRepository(t)
	svc := NewCompanySettingsService(mockRepo)
	return svc, mockRepo
}

func createTestSettings(companyID int64) *model.CompanySettings {
	now := time.Now()
	return &model.CompanySettings{
		ID:              1,
		CompanyID:       companyID,
		TaxEnabled:      true,
		TaxRate:         10.0,
		TaxInclusive:    false,
		RoundingEnabled: false,
		RoundingAmount:  0,
		MaxOfflineDays:  7,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
}

func TestNewCompanySettingsService(t *testing.T) {
	mockRepo := repoMocks.NewMockCompanySettingsRepository(t)
	svc := NewCompanySettingsService(mockRepo)
	assert.NotNil(t, svc)
}

// Get
func TestCompanySettingsService_Get_Success(t *testing.T) {
	svc, mockRepo := setupCompanySettingsTest(t)
	ctx := context.Background()
	settings := createTestSettings(1)

	mockRepo.EXPECT().GetByCompanyID(ctx, int64(1)).Return(settings, nil).Once()

	result, err := svc.Get(ctx, 1)
	require.NoError(t, err)
	assert.Equal(t, true, result.TaxEnabled)
	assert.Equal(t, 10.0, result.TaxRate)
}

func TestCompanySettingsService_Get_RepoError(t *testing.T) {
	svc, mockRepo := setupCompanySettingsTest(t)
	ctx := context.Background()

	mockRepo.EXPECT().GetByCompanyID(ctx, int64(1)).Return(nil, errors.New("db error")).Once()

	result, err := svc.Get(ctx, 1)
	require.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, apperror.IsInternalError(err))
}

// Update
func TestCompanySettingsService_Update_Success(t *testing.T) {
	svc, mockRepo := setupCompanySettingsTest(t)
	ctx := context.Background()
	settings := createTestSettings(1)

	taxRate := 12.0
	req := dto.UpdateCompanySettingsRequest{TaxRate: &taxRate}

	mockRepo.EXPECT().GetByCompanyID(ctx, int64(1)).Return(settings, nil).Once()
	mockRepo.EXPECT().Upsert(ctx, mock.AnythingOfType("*model.CompanySettings")).Return(nil).Once()

	result, err := svc.Update(ctx, 1, req)
	require.NoError(t, err)
	assert.Equal(t, 12.0, result.TaxRate)
}

func TestCompanySettingsService_Update_PartialUpdate(t *testing.T) {
	svc, mockRepo := setupCompanySettingsTest(t)
	ctx := context.Background()
	settings := createTestSettings(1)

	enabled := false
	req := dto.UpdateCompanySettingsRequest{TaxEnabled: &enabled}

	mockRepo.EXPECT().GetByCompanyID(ctx, int64(1)).Return(settings, nil).Once()
	mockRepo.EXPECT().Upsert(ctx, mock.AnythingOfType("*model.CompanySettings")).Return(nil).Once()

	result, err := svc.Update(ctx, 1, req)
	require.NoError(t, err)
	assert.Equal(t, false, result.TaxEnabled)
	assert.Equal(t, 10.0, result.TaxRate) // unchanged
}

func TestCompanySettingsService_Update_GetError(t *testing.T) {
	svc, mockRepo := setupCompanySettingsTest(t)
	ctx := context.Background()

	req := dto.UpdateCompanySettingsRequest{}
	mockRepo.EXPECT().GetByCompanyID(ctx, int64(1)).Return(nil, errors.New("db error")).Once()

	result, err := svc.Update(ctx, 1, req)
	require.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, apperror.IsInternalError(err))
}

func TestCompanySettingsService_Update_UpsertError(t *testing.T) {
	svc, mockRepo := setupCompanySettingsTest(t)
	ctx := context.Background()
	settings := createTestSettings(1)

	req := dto.UpdateCompanySettingsRequest{}
	mockRepo.EXPECT().GetByCompanyID(ctx, int64(1)).Return(settings, nil).Once()
	mockRepo.EXPECT().Upsert(ctx, mock.AnythingOfType("*model.CompanySettings")).Return(errors.New("db error")).Once()

	result, err := svc.Update(ctx, 1, req)
	require.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, apperror.IsInternalError(err))
}

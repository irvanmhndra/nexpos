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

func setupPromotionTest(t *testing.T) (*PromotionService, *repoMocks.MockPromotionRepository) {
	t.Helper()
	mockRepo := repoMocks.NewMockPromotionRepository(t)
	svc := NewPromotionService(mockRepo)
	return svc, mockRepo
}

func createTestPromotion(id, companyID int64, code string) *model.Promotion {
	now := time.Now()
	discountType := "percentage"
	discountValue := 10.0
	return &model.Promotion{
		ID:            id,
		CompanyID:     companyID,
		Code:          code,
		Name:          "Test Promo " + code,
		Type:          "discount",
		DiscountType:  &discountType,
		DiscountValue: &discountValue,
		StartAt:       now,
		IsActive:      true,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

func TestNewPromotionService(t *testing.T) {
	mockRepo := repoMocks.NewMockPromotionRepository(t)
	svc := NewPromotionService(mockRepo)
	assert.NotNil(t, svc)
}

// Create
func TestPromotionService_Create_Success(t *testing.T) {
	svc, mockRepo := setupPromotionTest(t)
	ctx := context.Background()

	req := dto.CreatePromotionRequest{
		Code: "PROMO1", Name: "Summer Sale", Type: "discount",
		StartAt: time.Now().Format(time.RFC3339), IsActive: true,
	}

	mockRepo.EXPECT().CodeExists(ctx, int64(1), "PROMO1", int64(0)).Return(false, nil).Once()
	mockRepo.EXPECT().Create(ctx, mock.AnythingOfType("*model.Promotion")).Return(nil).
		Run(func(args mock.Arguments) {
			p := args.Get(1).(*model.Promotion)
			p.ID = 1
			p.CreatedAt = time.Now()
			p.UpdatedAt = time.Now()
		}).Once()

	result, err := svc.Create(ctx, 1, req)
	require.NoError(t, err)
	assert.Equal(t, "PROMO1", result.Code)
}

func TestPromotionService_Create_CodeExists(t *testing.T) {
	svc, mockRepo := setupPromotionTest(t)
	ctx := context.Background()

	req := dto.CreatePromotionRequest{Code: "PROMO1", Name: "Sale", Type: "discount", StartAt: time.Now().Format(time.RFC3339)}
	mockRepo.EXPECT().CodeExists(ctx, int64(1), "PROMO1", int64(0)).Return(true, nil).Once()

	result, err := svc.Create(ctx, 1, req)
	require.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, apperror.IsBadRequest(err))
}

func TestPromotionService_Create_InvalidStartAt(t *testing.T) {
	svc, mockRepo := setupPromotionTest(t)
	ctx := context.Background()

	req := dto.CreatePromotionRequest{Code: "PROMO1", Name: "Sale", Type: "discount", StartAt: "not-a-date"}
	mockRepo.EXPECT().CodeExists(ctx, int64(1), "PROMO1", int64(0)).Return(false, nil).Once()

	result, err := svc.Create(ctx, 1, req)
	require.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, apperror.IsBadRequest(err))
	assert.Contains(t, err.Error(), "Invalid start_at")
}

func TestPromotionService_Create_InvalidEndAt(t *testing.T) {
	svc, mockRepo := setupPromotionTest(t)
	ctx := context.Background()

	badEnd := "not-a-date"
	req := dto.CreatePromotionRequest{Code: "PROMO1", Name: "Sale", Type: "discount", StartAt: time.Now().Format(time.RFC3339), EndAt: &badEnd}
	mockRepo.EXPECT().CodeExists(ctx, int64(1), "PROMO1", int64(0)).Return(false, nil).Once()

	result, err := svc.Create(ctx, 1, req)
	require.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, apperror.IsBadRequest(err))
	assert.Contains(t, err.Error(), "Invalid end_at")
}

// GetByID
func TestPromotionService_GetByID_Success(t *testing.T) {
	svc, mockRepo := setupPromotionTest(t)
	ctx := context.Background()
	expected := createTestPromotion(1, 1, "PROMO1")

	mockRepo.EXPECT().GetByID(ctx, int64(1), int64(1)).Return(expected, nil).Once()

	result, err := svc.GetByID(ctx, 1, 1)
	require.NoError(t, err)
	assert.Equal(t, "PROMO1", result.Code)
}

func TestPromotionService_GetByID_NotFound(t *testing.T) {
	svc, mockRepo := setupPromotionTest(t)
	ctx := context.Background()

	mockRepo.EXPECT().GetByID(ctx, int64(1), int64(999)).Return(nil, nil).Once()

	result, err := svc.GetByID(ctx, 1, 999)
	require.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, apperror.IsNotFound(err))
}

// List
func TestPromotionService_List_Success(t *testing.T) {
	svc, mockRepo := setupPromotionTest(t)
	ctx := context.Background()
	promos := []*model.Promotion{createTestPromotion(1, 1, "P1")}

	req := dto.ListPromotionRequest{Page: 1, PerPage: 20}
	mockRepo.EXPECT().List(ctx, int64(1), "", (*bool)(nil), "", 20, 0).Return(promos, 1, nil).Once()

	result, err := svc.List(ctx, 1, req)
	require.NoError(t, err)
	assert.Len(t, result.Promotions, 1)
}

func TestPromotionService_List_RepoError(t *testing.T) {
	svc, mockRepo := setupPromotionTest(t)
	ctx := context.Background()

	req := dto.ListPromotionRequest{Page: 1, PerPage: 20}
	mockRepo.EXPECT().List(ctx, int64(1), "", (*bool)(nil), "", 20, 0).Return(nil, 0, errors.New("db error")).Once()

	result, err := svc.List(ctx, 1, req)
	require.Error(t, err)
	assert.Nil(t, result)
}

// Update
func TestPromotionService_Update_Success(t *testing.T) {
	svc, mockRepo := setupPromotionTest(t)
	ctx := context.Background()
	existing := createTestPromotion(1, 1, "PROMO1")

	req := dto.UpdatePromotionRequest{
		Code: "PROMO2", Name: "Updated", Type: "discount",
		StartAt: time.Now().Format(time.RFC3339), IsActive: true,
	}
	mockRepo.EXPECT().GetByID(ctx, int64(1), int64(1)).Return(existing, nil).Once()
	mockRepo.EXPECT().CodeExists(ctx, int64(1), "PROMO2", int64(1)).Return(false, nil).Once()
	mockRepo.EXPECT().Update(ctx, mock.AnythingOfType("*model.Promotion")).Return(nil).Once()

	result, err := svc.Update(ctx, 1, 1, req)
	require.NoError(t, err)
	assert.Equal(t, "PROMO2", result.Code)
}

func TestPromotionService_Update_NotFound(t *testing.T) {
	svc, mockRepo := setupPromotionTest(t)
	ctx := context.Background()

	req := dto.UpdatePromotionRequest{Code: "PROMO2", Name: "Updated", Type: "discount", StartAt: time.Now().Format(time.RFC3339)}
	mockRepo.EXPECT().GetByID(ctx, int64(1), int64(999)).Return(nil, nil).Once()

	result, err := svc.Update(ctx, 1, 999, req)
	require.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, apperror.IsNotFound(err))
}

func TestPromotionService_Update_CodeExists(t *testing.T) {
	svc, mockRepo := setupPromotionTest(t)
	ctx := context.Background()
	existing := createTestPromotion(1, 1, "PROMO1")

	req := dto.UpdatePromotionRequest{Code: "TAKEN", Name: "Updated", Type: "discount", StartAt: time.Now().Format(time.RFC3339)}
	mockRepo.EXPECT().GetByID(ctx, int64(1), int64(1)).Return(existing, nil).Once()
	mockRepo.EXPECT().CodeExists(ctx, int64(1), "TAKEN", int64(1)).Return(true, nil).Once()

	result, err := svc.Update(ctx, 1, 1, req)
	require.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, apperror.IsBadRequest(err))
}

// Delete
func TestPromotionService_Delete_Success(t *testing.T) {
	svc, mockRepo := setupPromotionTest(t)
	ctx := context.Background()
	existing := createTestPromotion(1, 1, "PROMO1")

	mockRepo.EXPECT().GetByID(ctx, int64(1), int64(1)).Return(existing, nil).Once()
	mockRepo.EXPECT().Delete(ctx, int64(1), int64(1)).Return(nil).Once()

	err := svc.Delete(ctx, 1, 1)
	require.NoError(t, err)
}

func TestPromotionService_Delete_NotFound(t *testing.T) {
	svc, mockRepo := setupPromotionTest(t)
	ctx := context.Background()

	mockRepo.EXPECT().GetByID(ctx, int64(1), int64(999)).Return(nil, nil).Once()

	err := svc.Delete(ctx, 1, 999)
	require.Error(t, err)
	assert.True(t, apperror.IsNotFound(err))
}

func TestPromotionService_Delete_RepoError(t *testing.T) {
	svc, mockRepo := setupPromotionTest(t)
	ctx := context.Background()
	existing := createTestPromotion(1, 1, "PROMO1")

	mockRepo.EXPECT().GetByID(ctx, int64(1), int64(1)).Return(existing, nil).Once()
	mockRepo.EXPECT().Delete(ctx, int64(1), int64(1)).Return(errors.New("db error")).Once()

	err := svc.Delete(ctx, 1, 1)
	require.Error(t, err)
	assert.True(t, apperror.IsInternalError(err))
}

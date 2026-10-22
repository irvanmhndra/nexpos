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

func setupProductCategoryTest(t *testing.T) (*ProductCategoryService, *repoMocks.MockProductCategoryRepository) {
	t.Helper()
	mockRepo := repoMocks.NewMockProductCategoryRepository(t)
	svc := NewProductCategoryService(mockRepo)
	return svc, mockRepo
}

func createTestCategory(id, companyID int64, code string) *model.ProductCategory {
	now := time.Now()
	return &model.ProductCategory{
		ID:        id,
		CompanyID: companyID,
		Code:      code,
		Name:      "Category " + code,
		SortOrder: 0,
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func TestNewProductCategoryService(t *testing.T) {
	mockRepo := repoMocks.NewMockProductCategoryRepository(t)
	svc := NewProductCategoryService(mockRepo)
	assert.NotNil(t, svc)
}

// Create
func TestProductCategoryService_Create_Success(t *testing.T) {
	svc, mockRepo := setupProductCategoryTest(t)
	ctx := context.Background()

	req := dto.CreateProductCategoryRequest{Code: "CAT1", Name: "Food"}
	mockRepo.EXPECT().CodeExists(ctx, int64(1), "CAT1", int64(0)).Return(false, nil).Once()
	mockRepo.EXPECT().Create(ctx, mock.AnythingOfType("*model.ProductCategory")).Return(nil).
		Run(func(args mock.Arguments) {
			c := args.Get(1).(*model.ProductCategory)
			c.ID = 1
			c.CreatedAt = time.Now()
			c.UpdatedAt = time.Now()
		}).Once()
	// toResponse does NOT call GetByID since ParentID is nil

	result, err := svc.Create(ctx, 1, req)
	require.NoError(t, err)
	assert.Equal(t, "CAT1", result.Code)
	assert.True(t, result.IsActive) // default true
}

func TestProductCategoryService_Create_CodeExists(t *testing.T) {
	svc, mockRepo := setupProductCategoryTest(t)
	ctx := context.Background()

	req := dto.CreateProductCategoryRequest{Code: "CAT1", Name: "Food"}
	mockRepo.EXPECT().CodeExists(ctx, int64(1), "CAT1", int64(0)).Return(true, nil).Once()

	result, err := svc.Create(ctx, 1, req)
	require.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, apperror.IsBadRequest(err))
}

func TestProductCategoryService_Create_WithParent_Success(t *testing.T) {
	svc, mockRepo := setupProductCategoryTest(t)
	ctx := context.Background()
	parentID := int64(1)
	parent := createTestCategory(1, 1, "PARENT")

	req := dto.CreateProductCategoryRequest{Code: "CHILD", Name: "Child", ParentID: &parentID}
	mockRepo.EXPECT().CodeExists(ctx, int64(1), "CHILD", int64(0)).Return(false, nil).Once()
	mockRepo.EXPECT().GetByID(ctx, int64(1), int64(1)).Return(parent, nil).Once() // validate parent
	mockRepo.EXPECT().Create(ctx, mock.AnythingOfType("*model.ProductCategory")).Return(nil).
		Run(func(args mock.Arguments) {
			c := args.Get(1).(*model.ProductCategory)
			c.ID = 2
			c.ParentID = &parentID
			c.CreatedAt = time.Now()
			c.UpdatedAt = time.Now()
		}).Once()
	// toResponse calls GetByID for parent name
	mockRepo.EXPECT().GetByID(ctx, int64(1), int64(1)).Return(parent, nil).Once()

	result, err := svc.Create(ctx, 1, req)
	require.NoError(t, err)
	assert.NotNil(t, result.ParentID)
	assert.NotNil(t, result.ParentName)
}

func TestProductCategoryService_Create_ParentNotFound(t *testing.T) {
	svc, mockRepo := setupProductCategoryTest(t)
	ctx := context.Background()
	parentID := int64(999)

	req := dto.CreateProductCategoryRequest{Code: "CHILD", Name: "Child", ParentID: &parentID}
	mockRepo.EXPECT().CodeExists(ctx, int64(1), "CHILD", int64(0)).Return(false, nil).Once()
	mockRepo.EXPECT().GetByID(ctx, int64(1), int64(999)).Return(nil, nil).Once()

	result, err := svc.Create(ctx, 1, req)
	require.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, apperror.IsBadRequest(err))
	assert.Contains(t, err.Error(), "Parent category not found")
}

// GetByID
func TestProductCategoryService_GetByID_Success(t *testing.T) {
	svc, mockRepo := setupProductCategoryTest(t)
	ctx := context.Background()
	expected := createTestCategory(1, 1, "CAT1")

	mockRepo.EXPECT().GetByID(ctx, int64(1), int64(1)).Return(expected, nil).Once()

	result, err := svc.GetByID(ctx, 1, 1)
	require.NoError(t, err)
	assert.Equal(t, "CAT1", result.Code)
}

func TestProductCategoryService_GetByID_NotFound(t *testing.T) {
	svc, mockRepo := setupProductCategoryTest(t)
	ctx := context.Background()

	mockRepo.EXPECT().GetByID(ctx, int64(1), int64(999)).Return(nil, nil).Once()

	result, err := svc.GetByID(ctx, 1, 999)
	require.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, apperror.IsNotFound(err))
}

// List
func TestProductCategoryService_List_Success(t *testing.T) {
	svc, mockRepo := setupProductCategoryTest(t)
	ctx := context.Background()
	cats := []*model.ProductCategory{createTestCategory(1, 1, "C1")}

	req := dto.ListProductCategoryRequest{Page: 1, PerPage: 20}
	mockRepo.EXPECT().List(ctx, int64(1), "", (*bool)(nil), 20, 0).Return(cats, 1, nil).Once()

	result, err := svc.List(ctx, 1, req)
	require.NoError(t, err)
	assert.Len(t, result.Categories, 1)
}

func TestProductCategoryService_List_RepoError(t *testing.T) {
	svc, mockRepo := setupProductCategoryTest(t)
	ctx := context.Background()

	req := dto.ListProductCategoryRequest{Page: 1, PerPage: 20}
	mockRepo.EXPECT().List(ctx, int64(1), "", (*bool)(nil), 20, 0).Return(nil, 0, errors.New("db error")).Once()

	result, err := svc.List(ctx, 1, req)
	require.Error(t, err)
	assert.Nil(t, result)
}

// ListAll
func TestProductCategoryService_ListAll_Success(t *testing.T) {
	svc, mockRepo := setupProductCategoryTest(t)
	ctx := context.Background()
	cats := []*model.ProductCategory{createTestCategory(1, 1, "C1"), createTestCategory(2, 1, "C2")}

	mockRepo.EXPECT().ListAll(ctx, int64(1)).Return(cats, nil).Once()

	result, err := svc.ListAll(ctx, 1)
	require.NoError(t, err)
	assert.Len(t, result, 2)
}

func TestProductCategoryService_ListAll_RepoError(t *testing.T) {
	svc, mockRepo := setupProductCategoryTest(t)
	ctx := context.Background()

	mockRepo.EXPECT().ListAll(ctx, int64(1)).Return(nil, errors.New("db error")).Once()

	result, err := svc.ListAll(ctx, 1)
	require.Error(t, err)
	assert.Nil(t, result)
}

// Update
func TestProductCategoryService_Update_Success(t *testing.T) {
	svc, mockRepo := setupProductCategoryTest(t)
	ctx := context.Background()
	existing := createTestCategory(1, 1, "CAT1")

	req := dto.UpdateProductCategoryRequest{Code: "CAT2", Name: "Updated"}
	mockRepo.EXPECT().GetByID(ctx, int64(1), int64(1)).Return(existing, nil).Once() // exists check
	mockRepo.EXPECT().CodeExists(ctx, int64(1), "CAT2", int64(1)).Return(false, nil).Once()
	mockRepo.EXPECT().Update(ctx, mock.AnythingOfType("*model.ProductCategory")).Return(nil).Once()
	// After update, GetByID is called again
	updated := createTestCategory(1, 1, "CAT2")
	updated.Name = "Updated"
	mockRepo.EXPECT().GetByID(ctx, int64(1), int64(1)).Return(updated, nil).Once()

	result, err := svc.Update(ctx, 1, 1, req)
	require.NoError(t, err)
	assert.NotNil(t, result)
}

func TestProductCategoryService_Update_NotFound(t *testing.T) {
	svc, mockRepo := setupProductCategoryTest(t)
	ctx := context.Background()

	req := dto.UpdateProductCategoryRequest{Code: "CAT2", Name: "Updated"}
	mockRepo.EXPECT().GetByID(ctx, int64(1), int64(999)).Return(nil, nil).Once()

	result, err := svc.Update(ctx, 1, 999, req)
	require.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, apperror.IsNotFound(err))
}

func TestProductCategoryService_Update_SelfAsParent(t *testing.T) {
	svc, mockRepo := setupProductCategoryTest(t)
	ctx := context.Background()
	existing := createTestCategory(1, 1, "CAT1")
	selfID := int64(1)

	req := dto.UpdateProductCategoryRequest{Code: "CAT1", Name: "Cat", ParentID: &selfID}
	mockRepo.EXPECT().GetByID(ctx, int64(1), int64(1)).Return(existing, nil).Once()
	mockRepo.EXPECT().CodeExists(ctx, int64(1), "CAT1", int64(1)).Return(false, nil).Once()

	result, err := svc.Update(ctx, 1, 1, req)
	require.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, apperror.IsBadRequest(err))
	assert.Contains(t, err.Error(), "cannot be its own parent")
}

func TestProductCategoryService_Update_CircularRef(t *testing.T) {
	svc, mockRepo := setupProductCategoryTest(t)
	ctx := context.Background()
	// Cat 1 is parent, Cat 2 has parent=1. Now try to set Cat 1's parent to Cat 2 → circular
	existing := createTestCategory(1, 1, "CAT1")
	parentID := int64(2)
	childID := int64(1)
	parent := createTestCategory(2, 1, "CAT2")
	parent.ParentID = &childID // parent's parent is this category → circular

	req := dto.UpdateProductCategoryRequest{Code: "CAT1", Name: "Cat", ParentID: &parentID}
	mockRepo.EXPECT().GetByID(ctx, int64(1), int64(1)).Return(existing, nil).Once() // exists check
	mockRepo.EXPECT().CodeExists(ctx, int64(1), "CAT1", int64(1)).Return(false, nil).Once()
	mockRepo.EXPECT().GetByID(ctx, int64(1), int64(2)).Return(parent, nil).Once() // parent check

	result, err := svc.Update(ctx, 1, 1, req)
	require.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, apperror.IsBadRequest(err))
	assert.Contains(t, err.Error(), "Circular reference")
}

// Delete
func TestProductCategoryService_Delete_Success(t *testing.T) {
	svc, mockRepo := setupProductCategoryTest(t)
	ctx := context.Background()
	existing := createTestCategory(1, 1, "CAT1")

	mockRepo.EXPECT().GetByID(ctx, int64(1), int64(1)).Return(existing, nil).Once()
	mockRepo.EXPECT().HasChildren(ctx, int64(1), int64(1)).Return(false, nil).Once()
	mockRepo.EXPECT().HasProducts(ctx, int64(1), int64(1)).Return(false, nil).Once()
	mockRepo.EXPECT().Delete(ctx, int64(1), int64(1)).Return(nil).Once()

	err := svc.Delete(ctx, 1, 1)
	require.NoError(t, err)
}

func TestProductCategoryService_Delete_NotFound(t *testing.T) {
	svc, mockRepo := setupProductCategoryTest(t)
	ctx := context.Background()

	mockRepo.EXPECT().GetByID(ctx, int64(1), int64(999)).Return(nil, nil).Once()

	err := svc.Delete(ctx, 1, 999)
	require.Error(t, err)
	assert.True(t, apperror.IsNotFound(err))
}

func TestProductCategoryService_Delete_HasChildren(t *testing.T) {
	svc, mockRepo := setupProductCategoryTest(t)
	ctx := context.Background()
	existing := createTestCategory(1, 1, "CAT1")

	mockRepo.EXPECT().GetByID(ctx, int64(1), int64(1)).Return(existing, nil).Once()
	mockRepo.EXPECT().HasChildren(ctx, int64(1), int64(1)).Return(true, nil).Once()

	err := svc.Delete(ctx, 1, 1)
	require.Error(t, err)
	assert.True(t, apperror.IsBadRequest(err))
	assert.Contains(t, err.Error(), "subcategories")
}

func TestProductCategoryService_Delete_HasProducts(t *testing.T) {
	svc, mockRepo := setupProductCategoryTest(t)
	ctx := context.Background()
	existing := createTestCategory(1, 1, "CAT1")

	mockRepo.EXPECT().GetByID(ctx, int64(1), int64(1)).Return(existing, nil).Once()
	mockRepo.EXPECT().HasChildren(ctx, int64(1), int64(1)).Return(false, nil).Once()
	mockRepo.EXPECT().HasProducts(ctx, int64(1), int64(1)).Return(true, nil).Once()

	err := svc.Delete(ctx, 1, 1)
	require.Error(t, err)
	assert.True(t, apperror.IsBadRequest(err))
	assert.Contains(t, err.Error(), "products")
}

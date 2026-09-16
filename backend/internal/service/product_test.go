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

func setupProductTest(t *testing.T) (*ProductService, *repoMocks.MockProductRepository, *repoMocks.MockProductVariantRepository, *repoMocks.MockProductCategoryRepository) {
	t.Helper()
	productRepo := repoMocks.NewMockProductRepository(t)
	variantRepo := repoMocks.NewMockProductVariantRepository(t)
	categoryRepo := repoMocks.NewMockProductCategoryRepository(t)
	svc := NewProductService(productRepo, variantRepo, categoryRepo, nil)
	return svc, productRepo, variantRepo, categoryRepo
}

func createTestProduct(id, companyID int64) *model.Product {
	now := time.Now()
	return &model.Product{
		ID:        id,
		CompanyID: companyID,
		Name:      "Test Product",
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func createTestVariant(id, productID int64, sku string) *model.ProductVariant {
	now := time.Now()
	return &model.ProductVariant{
		ID:        id,
		ProductID: productID,
		SKU:       sku,
		Name:      "Variant " + sku,
		Price:     10000,
		IsDefault: id == 1,
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func TestNewProductService(t *testing.T) {
	productRepo := repoMocks.NewMockProductRepository(t)
	variantRepo := repoMocks.NewMockProductVariantRepository(t)
	categoryRepo := repoMocks.NewMockProductCategoryRepository(t)
	svc := NewProductService(productRepo, variantRepo, categoryRepo, nil)
	assert.NotNil(t, svc)
}

// Create
func TestProductService_Create_Success(t *testing.T) {
	svc, productRepo, variantRepo, categoryRepo := setupProductTest(t)
	ctx := context.Background()
	companyID := int64(1)
	catID := int64(1)

	req := dto.CreateProductRequest{
		Name:       "New Product",
		CategoryID: &catID,
		Variants:   []dto.ProductVariantInput{{SKU: "SKU001", Name: "Default", Price: 10000}},
	}

	category := createTestCategory(1, 1, "CAT1")
	categoryRepo.EXPECT().GetByID(ctx, companyID, catID).Return(category, nil).Once()
	variantRepo.EXPECT().SKUExists(ctx, "SKU001", int64(0)).Return(false, nil).Once()
	productRepo.EXPECT().Create(ctx, mock.AnythingOfType("*model.Product")).Return(nil).
		Run(func(args mock.Arguments) {
			p := args.Get(1).(*model.Product)
			p.ID = 1
		}).Once()
	variantRepo.EXPECT().Create(ctx, mock.AnythingOfType("*model.ProductVariant")).Return(nil).Once()

	// GetByID is called at end (via s.GetByID)
	product := createTestProduct(1, companyID)
	product.Name = "New Product"
	product.ProductCategoryID = &catID
	productRepo.EXPECT().GetByID(ctx, companyID, int64(1)).Return(product, nil).Once()
	variants := []*model.ProductVariant{createTestVariant(1, 1, "SKU001")}
	variantRepo.EXPECT().GetByProductID(ctx, int64(1)).Return(variants, nil).Once()
	categoryRepo.EXPECT().GetByID(ctx, companyID, catID).Return(category, nil).Once() // toResponse

	result, err := svc.Create(ctx, companyID, req)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "New Product", result.Name)
}

func TestProductService_Create_NoVariants(t *testing.T) {
	svc, _, _, _ := setupProductTest(t)
	ctx := context.Background()

	req := dto.CreateProductRequest{Name: "No Variants", Variants: []dto.ProductVariantInput{}}

	result, err := svc.Create(ctx, 1, req)
	require.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, apperror.IsBadRequest(err))
	assert.Contains(t, err.Error(), "at least one variant")
}

func TestProductService_Create_CategoryNotFound(t *testing.T) {
	svc, _, _, categoryRepo := setupProductTest(t)
	ctx := context.Background()
	catID := int64(999)

	req := dto.CreateProductRequest{
		Name:       "Product",
		CategoryID: &catID,
		Variants:   []dto.ProductVariantInput{{SKU: "SKU001", Name: "V1", Price: 100}},
	}

	categoryRepo.EXPECT().GetByID(ctx, int64(1), int64(999)).Return(nil, nil).Once()

	result, err := svc.Create(ctx, 1, req)
	require.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, apperror.IsBadRequest(err))
	assert.Contains(t, err.Error(), "Category not found")
}

func TestProductService_Create_SKUExists(t *testing.T) {
	svc, _, variantRepo, _ := setupProductTest(t)
	ctx := context.Background()

	req := dto.CreateProductRequest{
		Name:     "Product",
		Variants: []dto.ProductVariantInput{{SKU: "EXISTING", Name: "V1", Price: 100}},
	}

	variantRepo.EXPECT().SKUExists(ctx, "EXISTING", int64(0)).Return(true, nil).Once()

	result, err := svc.Create(ctx, 1, req)
	require.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, apperror.IsBadRequest(err))
	assert.Contains(t, err.Error(), "SKU already exists")
}

// GetByID
func TestProductService_GetByID_Success(t *testing.T) {
	svc, productRepo, variantRepo, _ := setupProductTest(t)
	ctx := context.Background()

	product := createTestProduct(1, 1)
	variants := []*model.ProductVariant{createTestVariant(1, 1, "SKU001")}

	productRepo.EXPECT().GetByID(ctx, int64(1), int64(1)).Return(product, nil).Once()
	variantRepo.EXPECT().GetByProductID(ctx, int64(1)).Return(variants, nil).Once()

	result, err := svc.GetByID(ctx, 1, 1)
	require.NoError(t, err)
	assert.Equal(t, "Test Product", result.Name)
	assert.Len(t, result.Variants, 1)
}

func TestProductService_GetByID_NotFound(t *testing.T) {
	svc, productRepo, _, _ := setupProductTest(t)
	ctx := context.Background()

	productRepo.EXPECT().GetByID(ctx, int64(1), int64(999)).Return(nil, nil).Once()

	result, err := svc.GetByID(ctx, 1, 999)
	require.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, apperror.IsNotFound(err))
}

// List
func TestProductService_List_Success(t *testing.T) {
	svc, productRepo, variantRepo, _ := setupProductTest(t)
	ctx := context.Background()

	products := []*model.Product{createTestProduct(1, 1)}
	req := dto.ListProductRequest{Page: 1, PerPage: 20}

	productRepo.EXPECT().List(ctx, int64(1), "", (*int64)(nil), (*bool)(nil), 20, 0).Return(products, 1, nil).Once()
	variantRepo.EXPECT().GetByProductID(ctx, int64(1)).Return([]*model.ProductVariant{createTestVariant(1, 1, "SKU1")}, nil).Once()

	result, err := svc.List(ctx, 1, req)
	require.NoError(t, err)
	assert.Len(t, result.Products, 1)
}

func TestProductService_List_RepoError(t *testing.T) {
	svc, productRepo, _, _ := setupProductTest(t)
	ctx := context.Background()

	req := dto.ListProductRequest{Page: 1, PerPage: 20}
	productRepo.EXPECT().List(ctx, int64(1), "", (*int64)(nil), (*bool)(nil), 20, 0).Return(nil, 0, errors.New("db error")).Once()

	result, err := svc.List(ctx, 1, req)
	require.Error(t, err)
	assert.Nil(t, result)
}

// Delete
func TestProductService_Delete_Success(t *testing.T) {
	svc, productRepo, variantRepo, _ := setupProductTest(t)
	ctx := context.Background()

	product := createTestProduct(1, 1)
	productRepo.EXPECT().GetByID(ctx, int64(1), int64(1)).Return(product, nil).Once()
	variantRepo.EXPECT().DeleteByProductID(ctx, int64(1)).Return(nil).Once()
	productRepo.EXPECT().Delete(ctx, int64(1), int64(1)).Return(nil).Once()

	err := svc.Delete(ctx, 1, 1)
	require.NoError(t, err)
}

func TestProductService_Delete_NotFound(t *testing.T) {
	svc, productRepo, _, _ := setupProductTest(t)
	ctx := context.Background()

	productRepo.EXPECT().GetByID(ctx, int64(1), int64(999)).Return(nil, nil).Once()

	err := svc.Delete(ctx, 1, 999)
	require.Error(t, err)
	assert.True(t, apperror.IsNotFound(err))
}

func TestProductService_Delete_RepoError(t *testing.T) {
	svc, productRepo, variantRepo, _ := setupProductTest(t)
	ctx := context.Background()

	product := createTestProduct(1, 1)
	productRepo.EXPECT().GetByID(ctx, int64(1), int64(1)).Return(product, nil).Once()
	variantRepo.EXPECT().DeleteByProductID(ctx, int64(1)).Return(nil).Once()
	productRepo.EXPECT().Delete(ctx, int64(1), int64(1)).Return(errors.New("db error")).Once()

	err := svc.Delete(ctx, 1, 1)
	require.Error(t, err)
	assert.True(t, apperror.IsInternalError(err))
}

// LookupByCode (barcode scan resolution)
func TestProductService_LookupByCode_MatchBarcode(t *testing.T) {
	svc, productRepo, variantRepo, _ := setupProductTest(t)
	ctx := context.Background()
	companyID := int64(1)
	code := "8991234567890"

	variant := createTestVariant(1, 1, "SKU001")
	variant.Barcode = &code
	variantRepo.EXPECT().FindByCodeInCompany(ctx, companyID, code).Return(variant, nil).Once()
	productRepo.EXPECT().GetByID(ctx, companyID, int64(1)).Return(createTestProduct(1, companyID), nil).Once()
	variantRepo.EXPECT().GetByProductID(ctx, int64(1)).Return([]*model.ProductVariant{variant}, nil).Once()

	res, err := svc.LookupByCode(ctx, companyID, code)
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, int64(1), res.MatchedVariantID)
	assert.Equal(t, "barcode", res.MatchedBy)
	assert.Equal(t, int64(1), res.Product.ID)
}

func TestProductService_LookupByCode_MatchSKUFallback(t *testing.T) {
	svc, productRepo, variantRepo, _ := setupProductTest(t)
	ctx := context.Background()
	companyID := int64(1)
	code := "SKU001"

	variant := createTestVariant(2, 1, "SKU001") // no barcode → matched on SKU
	variantRepo.EXPECT().FindByCodeInCompany(ctx, companyID, code).Return(variant, nil).Once()
	productRepo.EXPECT().GetByID(ctx, companyID, int64(1)).Return(createTestProduct(1, companyID), nil).Once()
	variantRepo.EXPECT().GetByProductID(ctx, int64(1)).Return([]*model.ProductVariant{variant}, nil).Once()

	res, err := svc.LookupByCode(ctx, companyID, code)
	require.NoError(t, err)
	assert.Equal(t, "sku", res.MatchedBy)
	assert.Equal(t, int64(2), res.MatchedVariantID)
}

func TestProductService_LookupByCode_NotFound(t *testing.T) {
	svc, _, variantRepo, _ := setupProductTest(t)
	ctx := context.Background()
	companyID := int64(1)
	code := "0000000000000"

	variantRepo.EXPECT().FindByCodeInCompany(ctx, companyID, code).Return(nil, nil).Once()

	res, err := svc.LookupByCode(ctx, companyID, code)
	require.Error(t, err)
	assert.Nil(t, res)
	assert.True(t, apperror.IsNotFound(err))
}

func TestProductService_LookupByCode_EmptyCode(t *testing.T) {
	svc, _, _, _ := setupProductTest(t)
	res, err := svc.LookupByCode(context.Background(), 1, "   ")
	require.Error(t, err)
	assert.Nil(t, res)
	assert.True(t, apperror.IsBadRequest(err))
}

// Create rejects a barcode already registered to another product
func TestProductService_Create_DuplicateBarcode_Rejected(t *testing.T) {
	svc, productRepo, variantRepo, _ := setupProductTest(t)
	ctx := context.Background()
	companyID := int64(1)
	barcode := "8991234567890"

	req := dto.CreateProductRequest{
		Name:     "New Product",
		Variants: []dto.ProductVariantInput{{SKU: "SKU001", Barcode: &barcode, Name: "Default", Price: 10000}},
	}

	variantRepo.EXPECT().SKUExists(ctx, "SKU001", int64(0)).Return(false, nil).Once()
	variantRepo.EXPECT().BarcodeExistsInOtherProduct(ctx, companyID, barcode, int64(0)).Return(true, nil).Once()

	res, err := svc.Create(ctx, companyID, req)
	require.Error(t, err)
	assert.Nil(t, res)
	assert.True(t, apperror.IsBadRequest(err))
	productRepo.AssertNotCalled(t, "Create")
}

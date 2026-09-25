package service

import (
	"context"
	"testing"

	"github.com/irvanmhndra/nexpos-api/internal/dto"
	"github.com/irvanmhndra/nexpos-api/internal/model"
	"github.com/irvanmhndra/nexpos-api/internal/repository"
	repoMocks "github.com/irvanmhndra/nexpos-api/internal/repository/mocks"
	"github.com/irvanmhndra/nexpos-api/pkg/apperror"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// =======================
// Test Helpers
// =======================

type inventoryTestSetup struct {
	svc          *InventoryService
	stockRepo    *repoMocks.MockStockRepository
	movementRepo *repoMocks.MockStockMovementRepository
	variantRepo  *repoMocks.MockProductVariantRepository
	branchRepo   *repoMocks.MockBranchRepository
}

func setupInventoryTest(t *testing.T) *inventoryTestSetup {
	t.Helper()
	s := &inventoryTestSetup{
		stockRepo:    repoMocks.NewMockStockRepository(t),
		movementRepo: repoMocks.NewMockStockMovementRepository(t),
		variantRepo:  repoMocks.NewMockProductVariantRepository(t),
		branchRepo:   repoMocks.NewMockBranchRepository(t),
	}
	s.svc = NewInventoryService(s.stockRepo, s.movementRepo, s.variantRepo, s.branchRepo, repoMocks.Transactor{})
	return s
}

func testBranch(id, companyID int64) *model.Branch {
	return &model.Branch{ID: id, CompanyID: companyID, Code: "BR01", Name: "Main Branch", IsActive: true}
}

func testVariantBasic(id int64) *model.ProductVariant {
	return &model.ProductVariant{ID: id, SKU: "SKU-001", Name: "Default", ProductID: 10}
}

// =======================
// AdjustStock Tests
// =======================

func TestInventoryService_AdjustStock_In_Success(t *testing.T) {
	s := setupInventoryTest(t)
	ctx := context.Background()
	companyID, variantID, branchID := int64(1), int64(5), int64(2)

	req := dto.AdjustStockRequest{
		VariantID: variantID,
		BranchID:  branchID,
		Type:      model.StockMovementIn,
		Quantity:  10,
	}

	s.variantRepo.EXPECT().GetByID(ctx, variantID).Return(testVariantBasic(variantID), nil).Once()
	s.branchRepo.EXPECT().GetByID(ctx, branchID).Return(testBranch(branchID, companyID), nil).Once()
	s.stockRepo.EXPECT().LockForUpdate(ctx, variantID, branchID).
		Return(&model.Stock{Quantity: 5, MinQuantity: 2}, nil).Once()
	s.movementRepo.EXPECT().Create(ctx, mock.MatchedBy(func(mv *model.StockMovement) bool {
		return mv.Type == model.StockMovementIn &&
			mv.StockBefore == 5 &&
			mv.StockAfter == 15 &&
			mv.Quantity == 10
	})).Return(nil).Once()
	s.stockRepo.EXPECT().Upsert(ctx, mock.MatchedBy(func(st *model.Stock) bool {
		return st.Quantity == 15
	})).Return(nil).Once()

	err := s.svc.AdjustStock(ctx, companyID, nil, req)

	require.NoError(t, err)
}

func TestInventoryService_AdjustStock_Out_Success(t *testing.T) {
	s := setupInventoryTest(t)
	ctx := context.Background()
	companyID, variantID, branchID := int64(1), int64(5), int64(2)

	req := dto.AdjustStockRequest{
		VariantID: variantID,
		BranchID:  branchID,
		Type:      model.StockMovementOut,
		Quantity:  3,
	}

	s.variantRepo.EXPECT().GetByID(ctx, variantID).Return(testVariantBasic(variantID), nil).Once()
	s.branchRepo.EXPECT().GetByID(ctx, branchID).Return(testBranch(branchID, companyID), nil).Once()
	s.stockRepo.EXPECT().LockForUpdate(ctx, variantID, branchID).
		Return(&model.Stock{Quantity: 10}, nil).Once()
	s.movementRepo.EXPECT().Create(ctx, mock.MatchedBy(func(mv *model.StockMovement) bool {
		return mv.StockBefore == 10 && mv.StockAfter == 7
	})).Return(nil).Once()
	s.stockRepo.EXPECT().Upsert(ctx, mock.MatchedBy(func(st *model.Stock) bool {
		return st.Quantity == 7
	})).Return(nil).Once()

	err := s.svc.AdjustStock(ctx, companyID, nil, req)

	require.NoError(t, err)
}

func TestInventoryService_AdjustStock_Adjust_Success(t *testing.T) {
	s := setupInventoryTest(t)
	ctx := context.Background()
	companyID, variantID, branchID := int64(1), int64(5), int64(2)

	req := dto.AdjustStockRequest{
		VariantID: variantID,
		BranchID:  branchID,
		Type:      model.StockMovementAdjust,
		Quantity:  25,
	}

	s.variantRepo.EXPECT().GetByID(ctx, variantID).Return(testVariantBasic(variantID), nil).Once()
	s.branchRepo.EXPECT().GetByID(ctx, branchID).Return(testBranch(branchID, companyID), nil).Once()
	s.stockRepo.EXPECT().LockForUpdate(ctx, variantID, branchID).
		Return(&model.Stock{Quantity: 10}, nil).Once()
	s.movementRepo.EXPECT().Create(ctx, mock.MatchedBy(func(mv *model.StockMovement) bool {
		return mv.StockAfter == 25
	})).Return(nil).Once()
	s.stockRepo.EXPECT().Upsert(ctx, mock.MatchedBy(func(st *model.Stock) bool {
		return st.Quantity == 25
	})).Return(nil).Once()

	err := s.svc.AdjustStock(ctx, companyID, nil, req)

	require.NoError(t, err)
}

func TestInventoryService_AdjustStock_NoExistingStock(t *testing.T) {
	// When no stock record exists, treat as 0 quantity
	s := setupInventoryTest(t)
	ctx := context.Background()
	companyID, variantID, branchID := int64(1), int64(5), int64(2)

	req := dto.AdjustStockRequest{
		VariantID: variantID,
		BranchID:  branchID,
		Type:      model.StockMovementIn,
		Quantity:  5,
	}

	s.variantRepo.EXPECT().GetByID(ctx, variantID).Return(testVariantBasic(variantID), nil).Once()
	s.branchRepo.EXPECT().GetByID(ctx, branchID).Return(testBranch(branchID, companyID), nil).Once()
	s.stockRepo.EXPECT().LockForUpdate(ctx, variantID, branchID).
		Return(nil, nil).Once() // no stock record yet
	s.movementRepo.EXPECT().Create(ctx, mock.MatchedBy(func(mv *model.StockMovement) bool {
		return mv.StockBefore == 0 && mv.StockAfter == 5
	})).Return(nil).Once()
	s.stockRepo.EXPECT().Upsert(ctx, mock.MatchedBy(func(st *model.Stock) bool {
		return st.Quantity == 5
	})).Return(nil).Once()

	err := s.svc.AdjustStock(ctx, companyID, nil, req)

	require.NoError(t, err)
}

func TestInventoryService_AdjustStock_InsufficientStock(t *testing.T) {
	s := setupInventoryTest(t)
	ctx := context.Background()
	companyID, variantID, branchID := int64(1), int64(5), int64(2)

	req := dto.AdjustStockRequest{
		VariantID: variantID,
		BranchID:  branchID,
		Type:      model.StockMovementOut,
		Quantity:  20,
	}

	s.variantRepo.EXPECT().GetByID(ctx, variantID).Return(testVariantBasic(variantID), nil).Once()
	s.branchRepo.EXPECT().GetByID(ctx, branchID).Return(testBranch(branchID, companyID), nil).Once()
	s.stockRepo.EXPECT().LockForUpdate(ctx, variantID, branchID).
		Return(&model.Stock{Quantity: 5}, nil).Once()

	err := s.svc.AdjustStock(ctx, companyID, nil, req)

	require.Error(t, err)
	assert.True(t, apperror.IsBadRequest(err))
	assert.Contains(t, err.Error(), "Insufficient stock")
}

func TestInventoryService_AdjustStock_VariantNotFound(t *testing.T) {
	s := setupInventoryTest(t)
	ctx := context.Background()

	req := dto.AdjustStockRequest{VariantID: 99, BranchID: 2, Type: model.StockMovementIn, Quantity: 1}

	s.variantRepo.EXPECT().GetByID(ctx, int64(99)).Return(nil, nil).Once()

	err := s.svc.AdjustStock(ctx, 1, nil, req)

	require.Error(t, err)
	assert.True(t, apperror.IsNotFound(err))
}

func TestInventoryService_AdjustStock_BranchNotFound(t *testing.T) {
	s := setupInventoryTest(t)
	ctx := context.Background()
	variantID, branchID := int64(5), int64(99)

	req := dto.AdjustStockRequest{VariantID: variantID, BranchID: branchID, Type: model.StockMovementIn, Quantity: 1}

	s.variantRepo.EXPECT().GetByID(ctx, variantID).Return(testVariantBasic(variantID), nil).Once()
	s.branchRepo.EXPECT().GetByID(ctx, branchID).Return(nil, nil).Once()

	err := s.svc.AdjustStock(ctx, 1, nil, req)

	require.Error(t, err)
	assert.True(t, apperror.IsNotFound(err))
}

func TestInventoryService_AdjustStock_BranchWrongCompany(t *testing.T) {
	s := setupInventoryTest(t)
	ctx := context.Background()
	variantID, branchID := int64(5), int64(2)

	req := dto.AdjustStockRequest{VariantID: variantID, BranchID: branchID, Type: model.StockMovementIn, Quantity: 1}

	s.variantRepo.EXPECT().GetByID(ctx, variantID).Return(testVariantBasic(variantID), nil).Once()
	s.branchRepo.EXPECT().GetByID(ctx, branchID).Return(testBranch(branchID, 999), nil).Once() // different company

	err := s.svc.AdjustStock(ctx, 1, nil, req)

	require.Error(t, err)
	var appErr *apperror.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, "FORBIDDEN", appErr.Code)
}

func TestInventoryService_AdjustStock_InvalidType(t *testing.T) {
	s := setupInventoryTest(t)
	ctx := context.Background()
	variantID, branchID := int64(5), int64(2)

	req := dto.AdjustStockRequest{VariantID: variantID, BranchID: branchID, Type: "INVALID", Quantity: 1}

	s.variantRepo.EXPECT().GetByID(ctx, variantID).Return(testVariantBasic(variantID), nil).Once()
	s.branchRepo.EXPECT().GetByID(ctx, branchID).Return(testBranch(branchID, 1), nil).Once()
	s.stockRepo.EXPECT().LockForUpdate(ctx, variantID, branchID).Return(nil, nil).Once()

	err := s.svc.AdjustStock(ctx, 1, nil, req)

	require.Error(t, err)
	assert.True(t, apperror.IsBadRequest(err))
}

// =======================
// UpdateMinStock Tests
// =======================

func TestInventoryService_UpdateMinStock_Success(t *testing.T) {
	s := setupInventoryTest(t)
	ctx := context.Background()
	companyID, variantID, branchID := int64(1), int64(5), int64(2)

	req := dto.UpdateMinStockRequest{MinQuantity: 10}

	s.variantRepo.EXPECT().GetByID(ctx, variantID).Return(testVariantBasic(variantID), nil).Once()
	s.branchRepo.EXPECT().GetByID(ctx, branchID).Return(testBranch(branchID, companyID), nil).Once()
	s.stockRepo.EXPECT().UpdateMinQuantity(ctx, variantID, branchID, 10).Return(nil).Once()

	err := s.svc.UpdateMinStock(ctx, companyID, variantID, branchID, req)

	require.NoError(t, err)
}

func TestInventoryService_UpdateMinStock_VariantNotFound(t *testing.T) {
	s := setupInventoryTest(t)
	ctx := context.Background()

	s.variantRepo.EXPECT().GetByID(ctx, int64(99)).Return(nil, nil).Once()

	err := s.svc.UpdateMinStock(ctx, 1, 99, 2, dto.UpdateMinStockRequest{MinQuantity: 5})

	require.Error(t, err)
	assert.True(t, apperror.IsNotFound(err))
}

func TestInventoryService_UpdateMinStock_BranchWrongCompany(t *testing.T) {
	s := setupInventoryTest(t)
	ctx := context.Background()
	variantID, branchID := int64(5), int64(2)

	s.variantRepo.EXPECT().GetByID(ctx, variantID).Return(testVariantBasic(variantID), nil).Once()
	s.branchRepo.EXPECT().GetByID(ctx, branchID).Return(testBranch(branchID, 999), nil).Once()

	err := s.svc.UpdateMinStock(ctx, 1, variantID, branchID, dto.UpdateMinStockRequest{MinQuantity: 5})

	require.Error(t, err)
	var appErr *apperror.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, "FORBIDDEN", appErr.Code)
}

// =======================
// ListInventory Tests
// =======================

func TestInventoryService_ListInventory_Success(t *testing.T) {
	s := setupInventoryTest(t)
	ctx := context.Background()
	companyID := int64(1)

	rows := []*repository.InventoryRow{
		{ProductVariantID: 1, ProductName: "Product A", VariantName: "Default", SKU: "SKU-001", CurrentStock: 10, MinStock: 2},
	}

	s.stockRepo.EXPECT().
		ListInventory(ctx, companyID, int64(0), "", "", "", 20, 0).
		Return(rows, 1, nil).Once()

	resp, err := s.svc.ListInventory(ctx, companyID, dto.ListInventoryRequest{Page: 1, PerPage: 20})

	require.NoError(t, err)
	assert.Len(t, resp.Items, 1)
	assert.Equal(t, 1, resp.Pagination.TotalRecords)
	assert.Equal(t, dto.StockStatusNormal, resp.Items[0].Status)
}

func TestInventoryService_ListInventory_DefaultPagination(t *testing.T) {
	s := setupInventoryTest(t)
	ctx := context.Background()

	s.stockRepo.EXPECT().
		ListInventory(ctx, int64(1), int64(0), "", "", "", 20, 0).
		Return([]*repository.InventoryRow{}, 0, nil).Once()

	resp, err := s.svc.ListInventory(ctx, 1, dto.ListInventoryRequest{})

	require.NoError(t, err)
	assert.Equal(t, 1, resp.Pagination.CurrentPage)
	assert.Equal(t, 20, resp.Pagination.PerPage)
}

func TestInventoryService_ListInventory_LowStockStatus(t *testing.T) {
	s := setupInventoryTest(t)
	ctx := context.Background()

	rows := []*repository.InventoryRow{
		{ProductVariantID: 1, CurrentStock: 2, MinStock: 5}, // current <= min → low
		{ProductVariantID: 2, CurrentStock: 0, MinStock: 5}, // out of stock
		{ProductVariantID: 3, CurrentStock: 10, MinStock: 5}, // normal
	}

	s.stockRepo.EXPECT().
		ListInventory(ctx, int64(1), int64(0), "", "", "", 20, 0).
		Return(rows, 3, nil).Once()

	resp, err := s.svc.ListInventory(ctx, 1, dto.ListInventoryRequest{Page: 1, PerPage: 20})

	require.NoError(t, err)
	assert.Equal(t, dto.StockStatusLow, resp.Items[0].Status)
	assert.Equal(t, dto.StockStatusOutOfStock, resp.Items[1].Status)
	assert.Equal(t, dto.StockStatusNormal, resp.Items[2].Status)
}

// =======================
// GetInventoryStats Tests
// =======================

func TestInventoryService_GetInventoryStats_Success(t *testing.T) {
	s := setupInventoryTest(t)
	ctx := context.Background()
	companyID, branchID := int64(1), int64(2)

	stats := &repository.InventoryStats{
		TotalSKU:        50,
		LowStock:        5,
		OutOfStock:      2,
		TotalStockValue: 5_000_000,
	}

	s.stockRepo.EXPECT().GetInventoryStats(ctx, companyID, branchID).Return(stats, nil).Once()

	resp, err := s.svc.GetInventoryStats(ctx, companyID, branchID)

	require.NoError(t, err)
	assert.Equal(t, 50, resp.TotalSKU)
	assert.Equal(t, 5, resp.LowStock)
	assert.Equal(t, 2, resp.OutOfStock)
	assert.Equal(t, float64(5_000_000), resp.TotalStockValue)
}

func TestInventoryService_GetInventoryStats_RepoError(t *testing.T) {
	s := setupInventoryTest(t)
	ctx := context.Background()

	s.stockRepo.EXPECT().GetInventoryStats(ctx, int64(1), int64(2)).
		Return(nil, assert.AnError).Once()

	_, err := s.svc.GetInventoryStats(ctx, 1, 2)

	require.Error(t, err)
	assert.True(t, apperror.IsInternalError(err))
}

// =======================
// ListMovements Tests
// =======================

func TestInventoryService_ListMovements_Success(t *testing.T) {
	s := setupInventoryTest(t)
	ctx := context.Background()
	companyID := int64(1)

	rows := []*repository.MovementRow{
		{ID: 1, ProductName: "Product A", VariantName: "Default", SKU: "SKU-001",
			Type: model.StockMovementIn, Quantity: 5, StockBefore: 10, StockAfter: 15},
	}

	s.movementRepo.EXPECT().
		List(ctx, companyID, int64(0), "", "", "", "", 20, 0).
		Return(rows, 1, nil).Once()

	resp, err := s.svc.ListMovements(ctx, companyID, dto.ListMovementsRequest{Page: 1, PerPage: 20})

	require.NoError(t, err)
	assert.Len(t, resp.Movements, 1)
	assert.Equal(t, 1, resp.Pagination.TotalRecords)
}

func TestInventoryService_ListMovements_DefaultPagination(t *testing.T) {
	s := setupInventoryTest(t)
	ctx := context.Background()

	s.movementRepo.EXPECT().
		List(ctx, int64(1), int64(0), "", "", "", "", 20, 0).
		Return([]*repository.MovementRow{}, 0, nil).Once()

	resp, err := s.svc.ListMovements(ctx, 1, dto.ListMovementsRequest{})

	require.NoError(t, err)
	assert.Equal(t, 1, resp.Pagination.CurrentPage)
	assert.Equal(t, 20, resp.Pagination.PerPage)
}

// =======================
// GetMovementStats Tests
// =======================

func TestInventoryService_GetMovementStats_Success(t *testing.T) {
	s := setupInventoryTest(t)
	ctx := context.Background()
	companyID, branchID := int64(1), int64(2)

	stats := &repository.MovementStats{
		TotalMovements: 100,
		StockIn:        60,
		StockOut:       30,
		Adjustments:    10,
	}

	s.movementRepo.EXPECT().GetMonthlyStats(ctx, companyID, branchID).Return(stats, nil).Once()

	resp, err := s.svc.GetMovementStats(ctx, companyID, branchID)

	require.NoError(t, err)
	assert.Equal(t, 100, resp.TotalMovements)
	assert.Equal(t, 60, resp.StockIn)
	assert.Equal(t, 30, resp.StockOut)
	assert.Equal(t, 10, resp.Adjustments)
}

func TestInventoryService_GetMovementStats_RepoError(t *testing.T) {
	s := setupInventoryTest(t)
	ctx := context.Background()

	s.movementRepo.EXPECT().GetMonthlyStats(ctx, int64(1), int64(2)).
		Return(nil, assert.AnError).Once()

	_, err := s.svc.GetMovementStats(ctx, 1, 2)

	require.Error(t, err)
	assert.True(t, apperror.IsInternalError(err))
}

package service

import (
	"context"
	"testing"

	"github.com/irvanmhndra/nexpos-api/internal/dto"
	"github.com/irvanmhndra/nexpos-api/internal/model"
	repoMocks "github.com/irvanmhndra/nexpos-api/internal/repository/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type opnameTestSetup struct {
	svc          *StockOpnameService
	opnameRepo   *repoMocks.MockStockOpnameRepository
	stockRepo    *repoMocks.MockStockRepository
	movementRepo *repoMocks.MockStockMovementRepository
	branchRepo   *repoMocks.MockBranchRepository
}

func setupOpnameTest(t *testing.T) *opnameTestSetup {
	t.Helper()
	s := &opnameTestSetup{
		opnameRepo:   repoMocks.NewMockStockOpnameRepository(t),
		stockRepo:    repoMocks.NewMockStockRepository(t),
		movementRepo: repoMocks.NewMockStockMovementRepository(t),
		branchRepo:   repoMocks.NewMockBranchRepository(t),
	}
	s.svc = NewStockOpnameService(s.opnameRepo, s.stockRepo, s.movementRepo, s.branchRepo)
	return s
}

func intPtr(v int) *int {
	return &v
}

func TestStockOpnameService_Create_Success(t *testing.T) {
	s := setupOpnameTest(t)
	ctx := context.Background()
	companyID, userID, branchID := int64(1), int64(7), int64(2)

	req := dto.CreateStockOpnameRequest{BranchID: branchID}

	s.branchRepo.EXPECT().GetByID(ctx, branchID).
		Return(&model.Branch{ID: branchID, CompanyID: companyID, Name: "BR"}, nil).Once()
	s.opnameRepo.EXPECT().GenerateOpnameNumber(ctx, companyID).Return("OPN-20260514-0001", nil).Once()
	s.opnameRepo.EXPECT().Create(ctx, mock.MatchedBy(func(op *model.StockOpname) bool {
		return op.CompanyID == companyID && op.BranchID == branchID &&
			op.Status == model.StockOpnameStatusInProgress && op.OpnameNumber == "OPN-20260514-0001"
	})).Run(func(args mock.Arguments) {
		op := args.Get(1).(*model.StockOpname)
		op.ID = 100
	}).Return(nil).Once()
	s.opnameRepo.EXPECT().SnapshotItems(ctx, int64(100), companyID, branchID, (*int64)(nil)).
		Return(5, nil).Once()

	// GetByID branch (called via Create at the end)
	s.opnameRepo.EXPECT().GetByID(ctx, companyID, int64(100)).
		Return(&model.StockOpname{ID: 100, CompanyID: companyID, BranchID: branchID,
			Status: model.StockOpnameStatusInProgress, OpnameNumber: "OPN-20260514-0001"}, nil).Once()
	s.opnameRepo.EXPECT().GetItems(ctx, int64(100)).Return([]*model.StockOpnameItem{}, nil).Once()
	s.opnameRepo.EXPECT().GetItemStats(ctx, int64(100)).Return(5, 0, nil).Once()

	resp, err := s.svc.Create(ctx, companyID, userID, req)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, int64(100), resp.ID)
	assert.Equal(t, model.StockOpnameStatusInProgress, resp.Status)
	assert.Equal(t, 5, resp.TotalItems)
	assert.Equal(t, 0, resp.CountedItems)
}

func TestStockOpnameService_Create_BranchNotFound(t *testing.T) {
	s := setupOpnameTest(t)
	ctx := context.Background()

	s.branchRepo.EXPECT().GetByID(ctx, int64(99)).Return(nil, nil).Once()

	_, err := s.svc.Create(ctx, 1, 1, dto.CreateStockOpnameRequest{BranchID: 99})
	require.Error(t, err)
}

func TestStockOpnameService_Create_BranchForbidden(t *testing.T) {
	s := setupOpnameTest(t)
	ctx := context.Background()

	s.branchRepo.EXPECT().GetByID(ctx, int64(2)).
		Return(&model.Branch{ID: 2, CompanyID: 999}, nil).Once()

	_, err := s.svc.Create(ctx, 1, 1, dto.CreateStockOpnameRequest{BranchID: 2})
	require.Error(t, err)
}

func TestStockOpnameService_Create_NoProducts(t *testing.T) {
	s := setupOpnameTest(t)
	ctx := context.Background()
	companyID, branchID := int64(1), int64(2)

	s.branchRepo.EXPECT().GetByID(ctx, branchID).
		Return(&model.Branch{ID: branchID, CompanyID: companyID}, nil).Once()
	s.opnameRepo.EXPECT().GenerateOpnameNumber(ctx, companyID).Return("OPN-1", nil).Once()
	s.opnameRepo.EXPECT().Create(ctx, mock.Anything).Run(func(args mock.Arguments) {
		args.Get(1).(*model.StockOpname).ID = 50
	}).Return(nil).Once()
	s.opnameRepo.EXPECT().SnapshotItems(ctx, int64(50), companyID, branchID, (*int64)(nil)).
		Return(0, nil).Once()

	_, err := s.svc.Create(ctx, companyID, 1, dto.CreateStockOpnameRequest{BranchID: branchID})
	require.Error(t, err)
}

func TestStockOpnameService_UpdateItem_Success(t *testing.T) {
	s := setupOpnameTest(t)
	ctx := context.Background()
	companyID, userID := int64(1), int64(3)

	s.opnameRepo.EXPECT().GetByID(ctx, companyID, int64(10)).
		Return(&model.StockOpname{ID: 10, CompanyID: companyID, Status: model.StockOpnameStatusInProgress}, nil).Once()
	s.opnameRepo.EXPECT().GetItem(ctx, int64(10), int64(20)).
		Return(&model.StockOpnameItem{
			ID: 20, StockOpnameID: 10, ProductVariantID: 5,
			SystemStock: 10, UnitCost: 1000,
		}, nil).Once()
	s.opnameRepo.EXPECT().UpdateItem(ctx, mock.MatchedBy(func(it *model.StockOpnameItem) bool {
		return it.ID == 20 && it.CountedStock != nil && *it.CountedStock == 8 &&
			it.VarianceQty == -2 && it.VarianceValue == -2000
	})).Return(nil).Once()

	resp, err := s.svc.UpdateItem(ctx, companyID, 10, 20, userID, dto.UpdateOpnameItemRequest{CountedStock: 8})
	require.NoError(t, err)
	assert.Equal(t, -2, resp.VarianceQty)
	assert.Equal(t, float64(-2000), resp.VarianceValue)
}

func TestStockOpnameService_UpdateItem_NotInProgress(t *testing.T) {
	s := setupOpnameTest(t)
	ctx := context.Background()

	s.opnameRepo.EXPECT().GetByID(ctx, int64(1), int64(10)).
		Return(&model.StockOpname{ID: 10, CompanyID: 1, Status: model.StockOpnameStatusCompleted}, nil).Once()

	_, err := s.svc.UpdateItem(ctx, 1, 10, 20, 1, dto.UpdateOpnameItemRequest{CountedStock: 5})
	require.Error(t, err)
}

func TestStockOpnameService_Complete_Success(t *testing.T) {
	s := setupOpnameTest(t)
	ctx := context.Background()
	companyID, userID, branchID := int64(1), int64(7), int64(2)

	s.opnameRepo.EXPECT().GetByID(ctx, companyID, int64(10)).
		Return(&model.StockOpname{
			ID: 10, CompanyID: companyID, BranchID: branchID,
			Status: model.StockOpnameStatusInProgress,
		}, nil).Once()

	items := []*model.StockOpnameItem{
		{ID: 1, StockOpnameID: 10, ProductVariantID: 100, SystemStock: 10, CountedStock: intPtr(8),
			VarianceQty: -2, UnitCost: 1000, VarianceValue: -2000},
		{ID: 2, StockOpnameID: 10, ProductVariantID: 101, SystemStock: 5, CountedStock: intPtr(5),
			VarianceQty: 0, UnitCost: 500, VarianceValue: 0},
		// Item 3 has no count — should be skipped from adjustment
		{ID: 3, StockOpnameID: 10, ProductVariantID: 102, SystemStock: 3, CountedStock: nil,
			VarianceQty: 0, UnitCost: 100, VarianceValue: 0},
	}
	s.opnameRepo.EXPECT().GetItems(ctx, int64(10)).Return(items, nil).Once()

	// Item 1: counted (8) != current system (10), adjustment expected
	s.stockRepo.EXPECT().GetByVariantAndBranch(ctx, int64(100), branchID).
		Return(&model.Stock{ProductVariantID: 100, BranchID: branchID, Quantity: 10}, nil).Once()
	s.movementRepo.EXPECT().Create(ctx, mock.MatchedBy(func(mv *model.StockMovement) bool {
		return mv.ProductVariantID == 100 && mv.Type == model.StockMovementAdjust &&
			mv.StockBefore == 10 && mv.StockAfter == 8 && mv.Quantity == 2
	})).Return(nil).Once()
	s.stockRepo.EXPECT().Upsert(ctx, mock.MatchedBy(func(st *model.Stock) bool {
		return st.ProductVariantID == 100 && st.Quantity == 8
	})).Return(nil).Once()

	// Item 2: counted == current, no adjustment
	s.stockRepo.EXPECT().GetByVariantAndBranch(ctx, int64(101), branchID).
		Return(&model.Stock{ProductVariantID: 101, BranchID: branchID, Quantity: 5}, nil).Once()

	// Final status update
	s.opnameRepo.EXPECT().Update(ctx, mock.MatchedBy(func(op *model.StockOpname) bool {
		return op.ID == 10 && op.Status == model.StockOpnameStatusCompleted &&
			op.TotalVarianceQty == -2 && op.TotalVarianceValue == -2000
	})).Return(nil).Once()

	// GetByID for the return value
	s.opnameRepo.EXPECT().GetByID(ctx, companyID, int64(10)).
		Return(&model.StockOpname{ID: 10, CompanyID: companyID, BranchID: branchID,
			Status: model.StockOpnameStatusCompleted, TotalVarianceQty: -2, TotalVarianceValue: -2000}, nil).Once()
	s.opnameRepo.EXPECT().GetItems(ctx, int64(10)).Return(items, nil).Once()
	s.opnameRepo.EXPECT().GetItemStats(ctx, int64(10)).Return(3, 2, nil).Once()

	resp, err := s.svc.Complete(ctx, companyID, 10, userID, dto.CompleteStockOpnameRequest{})
	require.NoError(t, err)
	assert.Equal(t, model.StockOpnameStatusCompleted, resp.Status)
	assert.Equal(t, -2, resp.TotalVarianceQty)
	assert.Equal(t, float64(-2000), resp.TotalVarianceValue)
}

func TestStockOpnameService_Complete_NotInProgress(t *testing.T) {
	s := setupOpnameTest(t)
	ctx := context.Background()

	s.opnameRepo.EXPECT().GetByID(ctx, int64(1), int64(10)).
		Return(&model.StockOpname{ID: 10, CompanyID: 1, Status: model.StockOpnameStatusCancelled}, nil).Once()

	_, err := s.svc.Complete(ctx, 1, 10, 1, dto.CompleteStockOpnameRequest{})
	require.Error(t, err)
}

func TestStockOpnameService_Cancel_Success(t *testing.T) {
	s := setupOpnameTest(t)
	ctx := context.Background()

	s.opnameRepo.EXPECT().GetByID(ctx, int64(1), int64(10)).
		Return(&model.StockOpname{ID: 10, CompanyID: 1, Status: model.StockOpnameStatusInProgress}, nil).Once()
	s.opnameRepo.EXPECT().Update(ctx, mock.MatchedBy(func(op *model.StockOpname) bool {
		return op.Status == model.StockOpnameStatusCancelled && op.CancelledAt != nil
	})).Return(nil).Once()

	// GetByID for response
	s.opnameRepo.EXPECT().GetByID(ctx, int64(1), int64(10)).
		Return(&model.StockOpname{ID: 10, CompanyID: 1, Status: model.StockOpnameStatusCancelled}, nil).Once()
	s.opnameRepo.EXPECT().GetItems(ctx, int64(10)).Return(nil, nil).Once()
	s.opnameRepo.EXPECT().GetItemStats(ctx, int64(10)).Return(0, 0, nil).Once()

	resp, err := s.svc.Cancel(ctx, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, model.StockOpnameStatusCancelled, resp.Status)
}

func TestStockOpnameService_BulkUpdateItems_Success(t *testing.T) {
	s := setupOpnameTest(t)
	ctx := context.Background()
	companyID, userID := int64(1), int64(3)

	s.opnameRepo.EXPECT().GetByID(ctx, companyID, int64(10)).
		Return(&model.StockOpname{ID: 10, CompanyID: companyID, Status: model.StockOpnameStatusInProgress}, nil).Once()

	s.opnameRepo.EXPECT().GetItem(ctx, int64(10), int64(1)).
		Return(&model.StockOpnameItem{ID: 1, StockOpnameID: 10, SystemStock: 10, UnitCost: 100}, nil).Once()
	s.opnameRepo.EXPECT().UpdateItem(ctx, mock.MatchedBy(func(it *model.StockOpnameItem) bool {
		return it.ID == 1 && *it.CountedStock == 12 && it.VarianceQty == 2 && it.VarianceValue == 200
	})).Return(nil).Once()

	s.opnameRepo.EXPECT().GetItem(ctx, int64(10), int64(2)).
		Return(&model.StockOpnameItem{ID: 2, StockOpnameID: 10, SystemStock: 5, UnitCost: 50}, nil).Once()
	s.opnameRepo.EXPECT().UpdateItem(ctx, mock.MatchedBy(func(it *model.StockOpnameItem) bool {
		return it.ID == 2 && *it.CountedStock == 3 && it.VarianceQty == -2 && it.VarianceValue == -100
	})).Return(nil).Once()

	// GetByID for the response
	s.opnameRepo.EXPECT().GetByID(ctx, companyID, int64(10)).
		Return(&model.StockOpname{ID: 10, CompanyID: companyID, Status: model.StockOpnameStatusInProgress}, nil).Once()
	s.opnameRepo.EXPECT().GetItems(ctx, int64(10)).Return(nil, nil).Once()
	s.opnameRepo.EXPECT().GetItemStats(ctx, int64(10)).Return(2, 2, nil).Once()

	req := dto.BulkUpdateOpnameItemsRequest{
		Items: []dto.BulkUpdateOpnameItem{
			{ItemID: 1, CountedStock: 12},
			{ItemID: 2, CountedStock: 3},
		},
	}
	resp, err := s.svc.BulkUpdateItems(ctx, companyID, 10, userID, req)
	require.NoError(t, err)
	assert.Equal(t, 2, resp.CountedItems)
}

func TestStockOpnameService_List_Success(t *testing.T) {
	s := setupOpnameTest(t)
	ctx := context.Background()
	companyID := int64(1)

	opnames := []*model.StockOpname{
		{ID: 1, CompanyID: companyID, Status: model.StockOpnameStatusInProgress, OpnameNumber: "OPN-1"},
		{ID: 2, CompanyID: companyID, Status: model.StockOpnameStatusCompleted, OpnameNumber: "OPN-2"},
	}
	s.opnameRepo.EXPECT().List(ctx, companyID, mock.Anything).Return(opnames, 2, nil).Once()
	s.opnameRepo.EXPECT().GetItemStats(ctx, int64(1)).Return(10, 5, nil).Once()
	s.opnameRepo.EXPECT().GetItemStats(ctx, int64(2)).Return(10, 10, nil).Once()

	resp, err := s.svc.List(ctx, companyID, dto.ListStockOpnameRequest{})
	require.NoError(t, err)
	assert.Len(t, resp.Opnames, 2)
	assert.Equal(t, 5, resp.Opnames[0].CountedItems)
	assert.Equal(t, 10, resp.Opnames[1].CountedItems)
}

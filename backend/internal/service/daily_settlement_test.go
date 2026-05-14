package service

import (
	"context"
	"testing"

	"github.com/irvanmhndra/nexpos-api/internal/dto"
	"github.com/irvanmhndra/nexpos-api/internal/model"
	"github.com/irvanmhndra/nexpos-api/internal/repository"
	repoMocks "github.com/irvanmhndra/nexpos-api/internal/repository/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type settlementTestSetup struct {
	svc        *DailySettlementService
	repo       *repoMocks.MockDailySettlementRepository
	branchRepo *repoMocks.MockBranchRepository
}

func setupSettlementTest(t *testing.T) *settlementTestSetup {
	t.Helper()
	s := &settlementTestSetup{
		repo:       repoMocks.NewMockDailySettlementRepository(t),
		branchRepo: repoMocks.NewMockBranchRepository(t),
	}
	s.svc = NewDailySettlementService(s.repo, s.branchRepo)
	return s
}

func floatPtr(v float64) *float64 {
	return &v
}

func TestDailySettlementService_Report_Success(t *testing.T) {
	s := setupSettlementTest(t)
	ctx := context.Background()
	companyID, branchID, date := int64(1), int64(2), "2026-05-14"

	s.branchRepo.EXPECT().GetByID(ctx, branchID).
		Return(&model.Branch{ID: branchID, CompanyID: companyID}, nil).Once()
	s.repo.EXPECT().GetPaymentBreakdown(ctx, companyID, branchID, date).
		Return([]*repository.PaymentMethodTotals{
			{Method: model.PaymentMethodCash, GrossSales: 1_500_000, Refunds: 50_000},
			{Method: model.PaymentMethodQRIS, GrossSales: 800_000, Refunds: 0},
		}, nil).Once()
	s.repo.EXPECT().GetExpensesTotal(ctx, companyID, branchID, date).
		Return(100_000.0, nil).Once()

	resp, err := s.svc.Report(ctx, companyID, dto.DailySettlementReportRequest{BranchID: branchID, Date: date})

	require.NoError(t, err)
	assert.Equal(t, 2_300_000.0, resp.TotalSales)         // 1.5M + 0.8M
	assert.Equal(t, 50_000.0, resp.TotalRefunds)
	assert.Equal(t, 100_000.0, resp.TotalExpenses)
	// expected_cash = 1.5M - 50k - 100k = 1.35M, qris = 800k, others = 0
	assert.Equal(t, 2_150_000.0, resp.TotalExpected)
	assert.Len(t, resp.ByMethod, 6) // all canonical methods always present

	var cash, qris *dto.SettlementMethodBreakdown
	for _, m := range resp.ByMethod {
		switch m.PaymentMethod {
		case model.PaymentMethodCash:
			cash = m
		case model.PaymentMethodQRIS:
			qris = m
		}
	}
	require.NotNil(t, cash)
	require.NotNil(t, qris)
	assert.Equal(t, 1_350_000.0, cash.ExpectedAmount)
	assert.Equal(t, 100_000.0, cash.ExpensesOut)
	assert.Equal(t, 800_000.0, qris.ExpectedAmount)
	assert.Equal(t, 0.0, qris.ExpensesOut)
}

func TestDailySettlementService_Report_MissingArgs(t *testing.T) {
	s := setupSettlementTest(t)
	_, err := s.svc.Report(context.Background(), 1, dto.DailySettlementReportRequest{BranchID: 0, Date: "2026-05-14"})
	require.Error(t, err)
}

func TestDailySettlementService_Create_Success(t *testing.T) {
	s := setupSettlementTest(t)
	ctx := context.Background()
	companyID, userID, branchID, date := int64(1), int64(7), int64(2), "2026-05-14"

	s.branchRepo.EXPECT().GetByID(ctx, branchID).
		Return(&model.Branch{ID: branchID, CompanyID: companyID}, nil).Once()
	s.repo.EXPECT().GetByBranchAndDate(ctx, companyID, branchID, date).Return(nil, nil).Once()
	s.repo.EXPECT().GetPaymentBreakdown(ctx, companyID, branchID, date).
		Return([]*repository.PaymentMethodTotals{
			{Method: model.PaymentMethodCash, GrossSales: 1_000_000, Refunds: 0},
		}, nil).Once()
	s.repo.EXPECT().GetExpensesTotal(ctx, companyID, branchID, date).Return(0.0, nil).Once()
	s.repo.EXPECT().Create(ctx, mock.MatchedBy(func(st *model.DailySettlement) bool {
		return st.CompanyID == companyID && st.BranchID == branchID &&
			st.SettlementDate == date && st.Status == model.DailySettlementStatusDraft &&
			st.TotalSales == 1_000_000
	})).Run(func(args mock.Arguments) {
		args.Get(1).(*model.DailySettlement).ID = 50
	}).Return(nil).Once()

	// CreateItem called 6 times (one per canonical method)
	s.repo.EXPECT().CreateItem(ctx, mock.Anything).Return(nil).Times(6)

	// GetByID for response
	s.repo.EXPECT().GetByID(ctx, companyID, int64(50)).
		Return(&model.DailySettlement{ID: 50, Status: model.DailySettlementStatusDraft}, nil).Once()
	s.repo.EXPECT().GetItems(ctx, int64(50)).Return([]*model.DailySettlementItem{}, nil).Once()

	resp, err := s.svc.Create(ctx, companyID, userID, dto.CreateDailySettlementRequest{
		BranchID:       branchID,
		SettlementDate: date,
	})

	require.NoError(t, err)
	assert.Equal(t, int64(50), resp.ID)
}

func TestDailySettlementService_Create_AlreadyExists(t *testing.T) {
	s := setupSettlementTest(t)
	ctx := context.Background()

	s.branchRepo.EXPECT().GetByID(ctx, int64(2)).
		Return(&model.Branch{ID: 2, CompanyID: 1}, nil).Once()
	s.repo.EXPECT().GetByBranchAndDate(ctx, int64(1), int64(2), "2026-05-14").
		Return(&model.DailySettlement{ID: 10}, nil).Once()

	_, err := s.svc.Create(ctx, 1, 1, dto.CreateDailySettlementRequest{BranchID: 2, SettlementDate: "2026-05-14"})
	require.Error(t, err)
}

func TestDailySettlementService_Create_InvalidDate(t *testing.T) {
	s := setupSettlementTest(t)
	ctx := context.Background()

	s.branchRepo.EXPECT().GetByID(ctx, int64(2)).
		Return(&model.Branch{ID: 2, CompanyID: 1}, nil).Once()

	_, err := s.svc.Create(ctx, 1, 1, dto.CreateDailySettlementRequest{BranchID: 2, SettlementDate: "14-05-2026"})
	require.Error(t, err)
}

func TestDailySettlementService_UpdateItem_Success(t *testing.T) {
	s := setupSettlementTest(t)
	ctx := context.Background()

	s.repo.EXPECT().GetByID(ctx, int64(1), int64(50)).
		Return(&model.DailySettlement{ID: 50, CompanyID: 1, Status: model.DailySettlementStatusDraft}, nil).Once()
	s.repo.EXPECT().GetItem(ctx, int64(50), int64(100)).
		Return(&model.DailySettlementItem{
			ID: 100, DailySettlementID: 50,
			PaymentMethod: model.PaymentMethodCash, ExpectedAmount: 1_000_000,
		}, nil).Once()
	s.repo.EXPECT().UpdateItem(ctx, mock.MatchedBy(func(it *model.DailySettlementItem) bool {
		return it.ID == 100 && it.ActualAmount != nil && *it.ActualAmount == 990_000 &&
			it.VarianceAmount == -10_000
	})).Return(nil).Once()

	resp, err := s.svc.UpdateItem(ctx, 1, 50, 100, dto.UpdateSettlementItemRequest{ActualAmount: 990_000})
	require.NoError(t, err)
	assert.Equal(t, -10_000.0, resp.VarianceAmount)
	require.NotNil(t, resp.ActualAmount)
	assert.Equal(t, 990_000.0, *resp.ActualAmount)
}

func TestDailySettlementService_UpdateItem_Finalized(t *testing.T) {
	s := setupSettlementTest(t)
	ctx := context.Background()

	s.repo.EXPECT().GetByID(ctx, int64(1), int64(50)).
		Return(&model.DailySettlement{ID: 50, CompanyID: 1, Status: model.DailySettlementStatusFinalized}, nil).Once()

	_, err := s.svc.UpdateItem(ctx, 1, 50, 100, dto.UpdateSettlementItemRequest{ActualAmount: 100})
	require.Error(t, err)
}

func TestDailySettlementService_Finalize_Success(t *testing.T) {
	s := setupSettlementTest(t)
	ctx := context.Background()

	s.repo.EXPECT().GetByID(ctx, int64(1), int64(50)).
		Return(&model.DailySettlement{ID: 50, CompanyID: 1, Status: model.DailySettlementStatusDraft}, nil).Once()
	s.repo.EXPECT().Update(ctx, mock.MatchedBy(func(st *model.DailySettlement) bool {
		return st.Status == model.DailySettlementStatusFinalized && st.FinalizedAt != nil
	})).Return(nil).Once()

	// GetByID for response
	s.repo.EXPECT().GetByID(ctx, int64(1), int64(50)).
		Return(&model.DailySettlement{ID: 50, CompanyID: 1, Status: model.DailySettlementStatusFinalized}, nil).Once()
	s.repo.EXPECT().GetItems(ctx, int64(50)).Return(nil, nil).Once()

	resp, err := s.svc.Finalize(ctx, 1, 50, 3, dto.FinalizeDailySettlementRequest{})
	require.NoError(t, err)
	assert.Equal(t, model.DailySettlementStatusFinalized, resp.Status)
}

func TestDailySettlementService_Finalize_NotDraft(t *testing.T) {
	s := setupSettlementTest(t)
	ctx := context.Background()

	s.repo.EXPECT().GetByID(ctx, int64(1), int64(50)).
		Return(&model.DailySettlement{ID: 50, CompanyID: 1, Status: model.DailySettlementStatusFinalized}, nil).Once()

	_, err := s.svc.Finalize(ctx, 1, 50, 3, dto.FinalizeDailySettlementRequest{})
	require.Error(t, err)
}

func TestDailySettlementService_BulkUpdate_Success(t *testing.T) {
	s := setupSettlementTest(t)
	ctx := context.Background()

	s.repo.EXPECT().GetByID(ctx, int64(1), int64(50)).
		Return(&model.DailySettlement{ID: 50, CompanyID: 1, Status: model.DailySettlementStatusDraft}, nil).Once()

	s.repo.EXPECT().GetItem(ctx, int64(50), int64(100)).
		Return(&model.DailySettlementItem{ID: 100, ExpectedAmount: 1_000_000}, nil).Once()
	s.repo.EXPECT().UpdateItem(ctx, mock.MatchedBy(func(it *model.DailySettlementItem) bool {
		return it.ID == 100 && *it.ActualAmount == 1_000_000 && it.VarianceAmount == 0
	})).Return(nil).Once()

	s.repo.EXPECT().GetItem(ctx, int64(50), int64(101)).
		Return(&model.DailySettlementItem{ID: 101, ExpectedAmount: 500_000}, nil).Once()
	s.repo.EXPECT().UpdateItem(ctx, mock.MatchedBy(func(it *model.DailySettlementItem) bool {
		return it.ID == 101 && *it.ActualAmount == 480_000 && it.VarianceAmount == -20_000
	})).Return(nil).Once()

	// GetByID for response
	s.repo.EXPECT().GetByID(ctx, int64(1), int64(50)).
		Return(&model.DailySettlement{ID: 50, CompanyID: 1, Status: model.DailySettlementStatusDraft}, nil).Once()
	s.repo.EXPECT().GetItems(ctx, int64(50)).Return(nil, nil).Once()

	resp, err := s.svc.BulkUpdateItems(ctx, 1, 50, dto.BulkUpdateSettlementItemsRequest{
		Items: []dto.BulkUpdateSettlementItem{
			{ItemID: 100, ActualAmount: 1_000_000},
			{ItemID: 101, ActualAmount: 480_000},
		},
	})
	require.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestDailySettlementService_List_Success(t *testing.T) {
	s := setupSettlementTest(t)
	ctx := context.Background()

	settlements := []*model.DailySettlement{
		{ID: 1, CompanyID: 1, BranchID: 2, SettlementDate: "2026-05-14", Status: model.DailySettlementStatusDraft},
		{ID: 2, CompanyID: 1, BranchID: 2, SettlementDate: "2026-05-13", Status: model.DailySettlementStatusFinalized},
	}
	s.repo.EXPECT().List(ctx, int64(1), mock.Anything).Return(settlements, 2, nil).Once()
	s.repo.EXPECT().GetItems(ctx, int64(1)).Return([]*model.DailySettlementItem{
		{ID: 10, ExpectedAmount: 500_000, ActualAmount: floatPtr(490_000), VarianceAmount: -10_000},
	}, nil).Once()
	s.repo.EXPECT().GetItems(ctx, int64(2)).Return([]*model.DailySettlementItem{}, nil).Once()

	resp, err := s.svc.List(ctx, 1, dto.ListDailySettlementRequest{})
	require.NoError(t, err)
	assert.Len(t, resp.Settlements, 2)
	assert.Equal(t, 500_000.0, resp.Settlements[0].TotalExpected)
	assert.Equal(t, 490_000.0, resp.Settlements[0].TotalActual)
	assert.Equal(t, -10_000.0, resp.Settlements[0].TotalVariance)
}

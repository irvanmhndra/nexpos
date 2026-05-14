package service

import (
	"context"
	"time"

	"github.com/irvanmhndra/nexpos-api/internal/dto"
	"github.com/irvanmhndra/nexpos-api/internal/model"
	"github.com/irvanmhndra/nexpos-api/internal/repository"
	"github.com/irvanmhndra/nexpos-api/pkg/apperror"
)

type DailySettlementService struct {
	repo       repository.DailySettlementRepository
	branchRepo repository.BranchRepository
}

func NewDailySettlementService(
	repo repository.DailySettlementRepository,
	branchRepo repository.BranchRepository,
) *DailySettlementService {
	return &DailySettlementService{repo: repo, branchRepo: branchRepo}
}

// allMethods is the canonical set of payment methods we always show on a
// settlement, even if a method had zero activity that day. Keeps the UI
// consistent and prevents missing-method confusion.
var allMethods = []string{
	model.PaymentMethodCash,
	model.PaymentMethodDebitCard,
	model.PaymentMethodCreditCard,
	model.PaymentMethodEWallet,
	model.PaymentMethodBankTransfer,
	model.PaymentMethodQRIS,
}

func (s *DailySettlementService) Report(ctx context.Context, companyID int64, req dto.DailySettlementReportRequest) (*dto.DailySettlementReportResponse, error) {
	if req.BranchID == 0 {
		return nil, apperror.BadRequest("branch_id is required")
	}
	if req.Date == "" {
		return nil, apperror.BadRequest("date is required")
	}
	if err := s.validateBranch(ctx, companyID, req.BranchID); err != nil {
		return nil, err
	}

	breakdown, err := s.computeBreakdown(ctx, companyID, req.BranchID, req.Date)
	if err != nil {
		return nil, err
	}

	return &dto.DailySettlementReportResponse{
		BranchID:       req.BranchID,
		SettlementDate: req.Date,
		TotalSales:     breakdown.totalSales,
		TotalRefunds:   breakdown.totalRefunds,
		TotalExpenses:  breakdown.totalExpenses,
		TotalExpected:  breakdown.totalExpected,
		ByMethod:       breakdown.byMethod,
	}, nil
}

func (s *DailySettlementService) Create(ctx context.Context, companyID, userID int64, req dto.CreateDailySettlementRequest) (*dto.DailySettlementResponse, error) {
	if err := s.validateBranch(ctx, companyID, req.BranchID); err != nil {
		return nil, err
	}
	if _, err := time.Parse("2006-01-02", req.SettlementDate); err != nil {
		return nil, apperror.BadRequest("settlement_date must be YYYY-MM-DD")
	}

	existing, err := s.repo.GetByBranchAndDate(ctx, companyID, req.BranchID, req.SettlementDate)
	if err != nil {
		return nil, apperror.InternalError(err)
	}
	if existing != nil {
		return nil, apperror.BadRequest("Settlement already exists for this branch and date")
	}

	breakdown, err := s.computeBreakdown(ctx, companyID, req.BranchID, req.SettlementDate)
	if err != nil {
		return nil, err
	}

	settlement := &model.DailySettlement{
		CompanyID:      companyID,
		BranchID:       req.BranchID,
		SettlementDate: req.SettlementDate,
		Status:         model.DailySettlementStatusDraft,
		TotalSales:     breakdown.totalSales,
		TotalRefunds:   breakdown.totalRefunds,
		TotalExpenses:  breakdown.totalExpenses,
		Notes:          req.Notes,
		RecordedBy:     &userID,
	}
	if err := s.repo.Create(ctx, settlement); err != nil {
		return nil, apperror.InternalError(err)
	}

	for _, bm := range breakdown.byMethod {
		item := &model.DailySettlementItem{
			DailySettlementID: settlement.ID,
			PaymentMethod:     bm.PaymentMethod,
			ExpectedAmount:    bm.ExpectedAmount,
			VarianceAmount:    -bm.ExpectedAmount, // actual=0 until user enters
		}
		if err := s.repo.CreateItem(ctx, item); err != nil {
			return nil, apperror.InternalError(err)
		}
	}

	return s.GetByID(ctx, companyID, settlement.ID)
}

func (s *DailySettlementService) GetByID(ctx context.Context, companyID, id int64) (*dto.DailySettlementResponse, error) {
	settlement, err := s.repo.GetByID(ctx, companyID, id)
	if err != nil {
		return nil, apperror.InternalError(err)
	}
	if settlement == nil {
		return nil, apperror.NotFound("Daily settlement not found")
	}
	items, err := s.repo.GetItems(ctx, settlement.ID)
	if err != nil {
		return nil, apperror.InternalError(err)
	}
	settlement.Items = items
	return toSettlementResponse(settlement), nil
}

func (s *DailySettlementService) List(ctx context.Context, companyID int64, req dto.ListDailySettlementRequest) (*dto.DailySettlementListResponse, error) {
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PerPage < 1 {
		req.PerPage = 20
	}
	if req.PerPage > 100 {
		req.PerPage = 100
	}
	offset := (req.Page - 1) * req.PerPage

	params := &repository.DailySettlementListParams{
		Status:   req.Status,
		BranchID: req.BranchID,
		DateFrom: req.DateFrom,
		DateTo:   req.DateTo,
		Limit:    req.PerPage,
		Offset:   offset,
	}
	settlements, total, err := s.repo.List(ctx, companyID, params)
	if err != nil {
		return nil, apperror.InternalError(err)
	}

	responses := make([]*dto.DailySettlementResponse, len(settlements))
	for i, st := range settlements {
		items, err := s.repo.GetItems(ctx, st.ID)
		if err != nil {
			return nil, apperror.InternalError(err)
		}
		st.Items = items
		responses[i] = toSettlementResponse(st)
	}

	return &dto.DailySettlementListResponse{
		Settlements: responses,
		Pagination:  buildPagination(total, req.Page, req.PerPage),
	}, nil
}

func (s *DailySettlementService) UpdateItem(ctx context.Context, companyID, settlementID, itemID int64, req dto.UpdateSettlementItemRequest) (*dto.DailySettlementItemResponse, error) {
	settlement, err := s.repo.GetByID(ctx, companyID, settlementID)
	if err != nil {
		return nil, apperror.InternalError(err)
	}
	if settlement == nil {
		return nil, apperror.NotFound("Daily settlement not found")
	}
	if settlement.Status != model.DailySettlementStatusDraft {
		return nil, apperror.BadRequest("Cannot edit a finalized settlement")
	}

	item, err := s.repo.GetItem(ctx, settlementID, itemID)
	if err != nil {
		return nil, apperror.InternalError(err)
	}
	if item == nil {
		return nil, apperror.NotFound("Settlement item not found")
	}

	actual := req.ActualAmount
	item.ActualAmount = &actual
	item.VarianceAmount = actual - item.ExpectedAmount
	item.Notes = req.Notes

	if err := s.repo.UpdateItem(ctx, item); err != nil {
		return nil, apperror.InternalError(err)
	}
	return toSettlementItemResponse(item), nil
}

func (s *DailySettlementService) BulkUpdateItems(ctx context.Context, companyID, settlementID int64, req dto.BulkUpdateSettlementItemsRequest) (*dto.DailySettlementResponse, error) {
	settlement, err := s.repo.GetByID(ctx, companyID, settlementID)
	if err != nil {
		return nil, apperror.InternalError(err)
	}
	if settlement == nil {
		return nil, apperror.NotFound("Daily settlement not found")
	}
	if settlement.Status != model.DailySettlementStatusDraft {
		return nil, apperror.BadRequest("Cannot edit a finalized settlement")
	}

	for _, line := range req.Items {
		item, err := s.repo.GetItem(ctx, settlementID, line.ItemID)
		if err != nil {
			return nil, apperror.InternalError(err)
		}
		if item == nil {
			return nil, apperror.BadRequest("Settlement item not found")
		}
		actual := line.ActualAmount
		item.ActualAmount = &actual
		item.VarianceAmount = actual - item.ExpectedAmount
		item.Notes = line.Notes
		if err := s.repo.UpdateItem(ctx, item); err != nil {
			return nil, apperror.InternalError(err)
		}
	}
	return s.GetByID(ctx, companyID, settlementID)
}

func (s *DailySettlementService) Finalize(ctx context.Context, companyID, settlementID, userID int64, req dto.FinalizeDailySettlementRequest) (*dto.DailySettlementResponse, error) {
	settlement, err := s.repo.GetByID(ctx, companyID, settlementID)
	if err != nil {
		return nil, apperror.InternalError(err)
	}
	if settlement == nil {
		return nil, apperror.NotFound("Daily settlement not found")
	}
	if settlement.Status != model.DailySettlementStatusDraft {
		return nil, apperror.BadRequest("Settlement is not in draft status")
	}

	now := time.Now()
	settlement.Status = model.DailySettlementStatusFinalized
	settlement.FinalizedBy = &userID
	settlement.FinalizedAt = &now
	if req.Notes != nil {
		settlement.Notes = req.Notes
	}

	if err := s.repo.Update(ctx, settlement); err != nil {
		return nil, apperror.InternalError(err)
	}
	return s.GetByID(ctx, companyID, settlement.ID)
}

// ============== Helpers ==============

func (s *DailySettlementService) validateBranch(ctx context.Context, companyID, branchID int64) error {
	branch, err := s.branchRepo.GetByID(ctx, branchID)
	if err != nil {
		return apperror.InternalError(err)
	}
	if branch == nil {
		return apperror.NotFound("Branch not found")
	}
	if branch.CompanyID != companyID {
		return apperror.Forbidden("Branch does not belong to your company")
	}
	return nil
}

type breakdownResult struct {
	byMethod       []*dto.SettlementMethodBreakdown
	totalSales     float64
	totalRefunds   float64
	totalExpenses  float64
	totalExpected  float64
}

func (s *DailySettlementService) computeBreakdown(ctx context.Context, companyID, branchID int64, date string) (*breakdownResult, error) {
	totals, err := s.repo.GetPaymentBreakdown(ctx, companyID, branchID, date)
	if err != nil {
		return nil, apperror.InternalError(err)
	}
	expenses, err := s.repo.GetExpensesTotal(ctx, companyID, branchID, date)
	if err != nil {
		return nil, apperror.InternalError(err)
	}

	byMethod := make(map[string]*repository.PaymentMethodTotals, len(totals))
	for _, t := range totals {
		byMethod[t.Method] = t
	}

	result := &breakdownResult{totalExpenses: expenses}
	result.byMethod = make([]*dto.SettlementMethodBreakdown, 0, len(allMethods))
	for _, method := range allMethods {
		gross := 0.0
		refunds := 0.0
		if t, ok := byMethod[method]; ok {
			gross = t.GrossSales
			refunds = t.Refunds
		}
		expensesOut := 0.0
		if method == model.PaymentMethodCash {
			expensesOut = expenses
		}
		expected := gross - refunds - expensesOut

		result.byMethod = append(result.byMethod, &dto.SettlementMethodBreakdown{
			PaymentMethod:  method,
			GrossSales:     gross,
			Refunds:        refunds,
			ExpensesOut:    expensesOut,
			ExpectedAmount: expected,
		})
		result.totalSales += gross
		result.totalRefunds += refunds
		result.totalExpected += expected
	}
	return result, nil
}

func toSettlementResponse(s *model.DailySettlement) *dto.DailySettlementResponse {
	resp := &dto.DailySettlementResponse{
		ID:             s.ID,
		BranchID:       s.BranchID,
		SettlementDate: s.SettlementDate,
		Status:         s.Status,
		TotalSales:     s.TotalSales,
		TotalRefunds:   s.TotalRefunds,
		TotalExpenses:  s.TotalExpenses,
		Notes:          s.Notes,
		RecordedBy:     s.RecordedBy,
		FinalizedBy:    s.FinalizedBy,
		FinalizedAt:    s.FinalizedAt,
		CreatedAt:      s.CreatedAt,
		UpdatedAt:      s.UpdatedAt,
	}

	if s.Items != nil {
		resp.Items = make([]*dto.DailySettlementItemResponse, len(s.Items))
		for i, it := range s.Items {
			resp.Items[i] = toSettlementItemResponse(it)
			resp.TotalExpected += it.ExpectedAmount
			if it.ActualAmount != nil {
				resp.TotalActual += *it.ActualAmount
			}
			resp.TotalVariance += it.VarianceAmount
		}
	}
	return resp
}

func toSettlementItemResponse(it *model.DailySettlementItem) *dto.DailySettlementItemResponse {
	return &dto.DailySettlementItemResponse{
		ID:             it.ID,
		PaymentMethod:  it.PaymentMethod,
		ExpectedAmount: it.ExpectedAmount,
		ActualAmount:   it.ActualAmount,
		VarianceAmount: it.VarianceAmount,
		Notes:          it.Notes,
	}
}

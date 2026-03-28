package service

import (
	"context"
	"time"

	"github.com/irvanmhndra/nexpos-api/internal/dto"
	"github.com/irvanmhndra/nexpos-api/internal/model"
	"github.com/irvanmhndra/nexpos-api/internal/repository"
	"github.com/irvanmhndra/nexpos-api/pkg/apperror"
)

type ShiftService struct {
	shiftRepo   repository.ShiftRepository
	paymentRepo repository.PaymentRepository
}

func NewShiftService(
	shiftRepo repository.ShiftRepository,
	paymentRepo repository.PaymentRepository,
) *ShiftService {
	return &ShiftService{
		shiftRepo:   shiftRepo,
		paymentRepo: paymentRepo,
	}
}

func (s *ShiftService) OpenShift(ctx context.Context, companyID, branchID, cashierID int64, req dto.OpenShiftRequest) (*dto.ShiftResponse, error) {
	existing, err := s.shiftRepo.GetOpenShift(ctx, companyID, branchID, cashierID)
	if err != nil {
		return nil, apperror.InternalError(err)
	}
	if existing != nil {
		return nil, apperror.BadRequest("There is already an open shift for this cashier at this branch")
	}

	now := time.Now()
	shift := &model.Shift{
		CompanyID:    companyID,
		BranchID:     branchID,
		CashierID:    cashierID,
		Status:       model.ShiftStatusOpen,
		OpeningFloat: req.OpeningFloat,
		OpenedAt:     now,
	}

	if err := s.shiftRepo.Create(ctx, shift); err != nil {
		return nil, apperror.InternalError(err)
	}

	return toShiftResponse(shift), nil
}

func (s *ShiftService) GetCurrentShift(ctx context.Context, companyID, branchID, cashierID int64) (*dto.ShiftResponse, error) {
	shift, err := s.shiftRepo.GetOpenShift(ctx, companyID, branchID, cashierID)
	if err != nil {
		return nil, apperror.InternalError(err)
	}
	if shift == nil {
		return nil, apperror.NotFound("No open shift found")
	}
	return toShiftResponse(shift), nil
}

func (s *ShiftService) CloseShift(ctx context.Context, companyID, shiftID int64, req dto.CloseShiftRequest) (*dto.ShiftResponse, error) {
	shift, err := s.shiftRepo.GetByID(ctx, companyID, shiftID)
	if err != nil {
		return nil, apperror.InternalError(err)
	}
	if shift == nil {
		return nil, apperror.NotFound("Shift not found")
	}
	if shift.Status != model.ShiftStatusOpen {
		return nil, apperror.BadRequest("Shift is not open")
	}

	now := time.Now()

	cashTotal, err := s.paymentRepo.GetCashTotalByPeriod(ctx, shift.BranchID, shift.OpenedAt, now)
	if err != nil {
		return nil, apperror.InternalError(err)
	}

	expectedCash := shift.OpeningFloat + cashTotal
	actualCash := req.ActualCash
	cashDifference := actualCash - expectedCash

	shift.Status = model.ShiftStatusClosed
	shift.ClosingFloat = actualCash
	shift.ExpectedCash = expectedCash
	shift.ActualCash = &actualCash
	shift.CashDifference = &cashDifference
	shift.Notes = req.Notes
	shift.ClosedAt = &now

	if err := s.shiftRepo.Update(ctx, shift); err != nil {
		return nil, apperror.InternalError(err)
	}

	shift, _ = s.shiftRepo.GetByID(ctx, companyID, shiftID)
	return toShiftResponse(shift), nil
}

func (s *ShiftService) GetShift(ctx context.Context, companyID, id int64) (*dto.ShiftResponse, error) {
	shift, err := s.shiftRepo.GetByID(ctx, companyID, id)
	if err != nil {
		return nil, apperror.InternalError(err)
	}
	if shift == nil {
		return nil, apperror.NotFound("Shift not found")
	}
	return toShiftResponse(shift), nil
}

func (s *ShiftService) ListShifts(ctx context.Context, companyID int64, req dto.ListShiftRequest) (*dto.ShiftListResponse, error) {
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

	params := &repository.ShiftListParams{
		BranchID:  req.BranchID,
		CashierID: req.CashierID,
		Status:    req.Status,
		DateFrom:  req.DateFrom,
		DateTo:    req.DateTo,
		Limit:     req.PerPage,
		Offset:    offset,
	}

	shifts, total, err := s.shiftRepo.List(ctx, companyID, params)
	if err != nil {
		return nil, apperror.InternalError(err)
	}

	responses := make([]*dto.ShiftResponse, len(shifts))
	for i, sh := range shifts {
		responses[i] = toShiftResponse(sh)
	}

	return &dto.ShiftListResponse{
		Shifts:     responses,
		Pagination: buildPagination(total, req.Page, req.PerPage),
	}, nil
}

func toShiftResponse(s *model.Shift) *dto.ShiftResponse {
	return &dto.ShiftResponse{
		ID:             s.ID,
		BranchID:       s.BranchID,
		CashierID:      s.CashierID,
		Status:         s.Status,
		OpeningFloat:   s.OpeningFloat,
		ClosingFloat:   s.ClosingFloat,
		ExpectedCash:   s.ExpectedCash,
		ActualCash:     s.ActualCash,
		CashDifference: s.CashDifference,
		Notes:          s.Notes,
		OpenedAt:       s.OpenedAt,
		ClosedAt:       s.ClosedAt,
		CreatedAt:      s.CreatedAt,
		UpdatedAt:      s.UpdatedAt,
	}
}

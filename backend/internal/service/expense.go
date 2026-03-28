package service

import (
	"context"
	"fmt"

	"github.com/irvanmhndra/nexpos-api/internal/dto"
	"github.com/irvanmhndra/nexpos-api/internal/model"
	"github.com/irvanmhndra/nexpos-api/internal/repository"
	"github.com/irvanmhndra/nexpos-api/pkg/apperror"
)

type ExpenseService struct {
	expenseRepo  repository.ExpenseRepository
	categoryRepo repository.ExpenseCategoryRepository
}

func NewExpenseService(
	expenseRepo repository.ExpenseRepository,
	categoryRepo repository.ExpenseCategoryRepository,
) *ExpenseService {
	return &ExpenseService{
		expenseRepo:  expenseRepo,
		categoryRepo: categoryRepo,
	}
}

// ============== Category Methods ==============

func (s *ExpenseService) CreateCategory(ctx context.Context, companyID int64, req dto.CreateExpenseCategoryRequest) (*dto.ExpenseCategoryResponse, error) {
	category := &model.ExpenseCategory{
		CompanyID:   companyID,
		Name:        req.Name,
		Description: req.Description,
	}

	if err := s.categoryRepo.Create(ctx, category); err != nil {
		return nil, apperror.InternalError(err)
	}

	return toExpenseCategoryResponse(category), nil
}

func (s *ExpenseService) ListCategories(ctx context.Context, companyID int64, req dto.ListExpenseCategoryRequest) (*dto.ExpenseCategoryListResponse, error) {
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

	categories, total, err := s.categoryRepo.List(ctx, companyID, req.Search, req.PerPage, offset)
	if err != nil {
		return nil, apperror.InternalError(err)
	}

	responses := make([]*dto.ExpenseCategoryResponse, len(categories))
	for i, cat := range categories {
		responses[i] = toExpenseCategoryResponse(cat)
	}

	return &dto.ExpenseCategoryListResponse{
		Categories: responses,
		Pagination: buildPagination(total, req.Page, req.PerPage),
	}, nil
}

func (s *ExpenseService) UpdateCategory(ctx context.Context, companyID, id int64, req dto.UpdateExpenseCategoryRequest) (*dto.ExpenseCategoryResponse, error) {
	category, err := s.categoryRepo.GetByID(ctx, companyID, id)
	if err != nil {
		return nil, apperror.InternalError(err)
	}
	if category == nil {
		return nil, apperror.NotFound("Expense category not found")
	}

	category.Name = req.Name
	category.Description = req.Description

	if err := s.categoryRepo.Update(ctx, category); err != nil {
		return nil, apperror.InternalError(err)
	}

	category, _ = s.categoryRepo.GetByID(ctx, companyID, id)
	return toExpenseCategoryResponse(category), nil
}

func (s *ExpenseService) DeleteCategory(ctx context.Context, companyID, id int64) error {
	category, err := s.categoryRepo.GetByID(ctx, companyID, id)
	if err != nil {
		return apperror.InternalError(err)
	}
	if category == nil {
		return apperror.NotFound("Expense category not found")
	}

	if err := s.categoryRepo.Delete(ctx, companyID, id); err != nil {
		return apperror.InternalError(err)
	}
	return nil
}

// ============== Expense Methods ==============

func (s *ExpenseService) CreateExpense(ctx context.Context, companyID int64, recordedBy *int64, req dto.CreateExpenseRequest) (*dto.ExpenseResponse, error) {
	if req.CategoryID != nil {
		cat, err := s.categoryRepo.GetByID(ctx, companyID, *req.CategoryID)
		if err != nil {
			return nil, apperror.InternalError(err)
		}
		if cat == nil {
			return nil, apperror.NotFound("Expense category not found")
		}
	}

	expense := &model.Expense{
		CompanyID:   companyID,
		BranchID:    req.BranchID,
		CategoryID:  req.CategoryID,
		Amount:      req.Amount,
		Description: req.Description,
		ReferenceNo: req.ReferenceNo,
		ExpenseDate: req.ExpenseDate,
		RecordedBy:  recordedBy,
		Notes:       req.Notes,
	}

	if err := s.expenseRepo.Create(ctx, expense); err != nil {
		return nil, apperror.InternalError(err)
	}

	return toExpenseResponse(expense), nil
}

func (s *ExpenseService) GetExpense(ctx context.Context, companyID, id int64) (*dto.ExpenseResponse, error) {
	expense, err := s.expenseRepo.GetByID(ctx, companyID, id)
	if err != nil {
		return nil, apperror.InternalError(err)
	}
	if expense == nil {
		return nil, apperror.NotFound("Expense not found")
	}
	return toExpenseResponse(expense), nil
}

func (s *ExpenseService) ListExpenses(ctx context.Context, companyID int64, req dto.ListExpenseRequest) (*dto.ExpenseListResponse, error) {
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

	params := &repository.ExpenseListParams{
		BranchID:   req.BranchID,
		CategoryID: req.CategoryID,
		DateFrom:   req.DateFrom,
		DateTo:     req.DateTo,
		Limit:      req.PerPage,
		Offset:     offset,
	}

	expenses, total, err := s.expenseRepo.List(ctx, companyID, params)
	if err != nil {
		return nil, apperror.InternalError(err)
	}

	responses := make([]*dto.ExpenseResponse, len(expenses))
	for i, exp := range expenses {
		responses[i] = toExpenseResponse(exp)
	}

	return &dto.ExpenseListResponse{
		Expenses:   responses,
		Pagination: buildPagination(total, req.Page, req.PerPage),
	}, nil
}

func (s *ExpenseService) UpdateExpense(ctx context.Context, companyID, id int64, req dto.UpdateExpenseRequest) (*dto.ExpenseResponse, error) {
	expense, err := s.expenseRepo.GetByID(ctx, companyID, id)
	if err != nil {
		return nil, apperror.InternalError(err)
	}
	if expense == nil {
		return nil, apperror.NotFound("Expense not found")
	}

	if req.CategoryID != nil {
		cat, err := s.categoryRepo.GetByID(ctx, companyID, *req.CategoryID)
		if err != nil {
			return nil, apperror.InternalError(err)
		}
		if cat == nil {
			return nil, apperror.NotFound("Expense category not found")
		}
	}

	expense.BranchID = req.BranchID
	expense.CategoryID = req.CategoryID
	expense.Amount = req.Amount
	expense.Description = req.Description
	expense.ReferenceNo = req.ReferenceNo
	expense.ExpenseDate = req.ExpenseDate
	expense.Notes = req.Notes

	if err := s.expenseRepo.Update(ctx, expense); err != nil {
		return nil, apperror.InternalError(err)
	}

	expense, _ = s.expenseRepo.GetByID(ctx, companyID, id)
	return toExpenseResponse(expense), nil
}

func (s *ExpenseService) DeleteExpense(ctx context.Context, companyID, id int64) error {
	expense, err := s.expenseRepo.GetByID(ctx, companyID, id)
	if err != nil {
		return apperror.InternalError(err)
	}
	if expense == nil {
		return apperror.NotFound("Expense not found")
	}

	if err := s.expenseRepo.Delete(ctx, companyID, id); err != nil {
		return apperror.InternalError(err)
	}
	return nil
}

func (s *ExpenseService) GetExpenseSummary(ctx context.Context, companyID int64, req dto.ExpenseSummaryRequest) (*dto.ExpenseSummaryResponse, error) {
	total, err := s.expenseRepo.GetTotalByDateRange(ctx, companyID, req.BranchID, req.DateFrom, req.DateTo)
	if err != nil {
		return nil, apperror.InternalError(err)
	}

	// Get all categories for this company to aggregate
	categories, _, err := s.categoryRepo.List(ctx, companyID, "", 1000, 0)
	if err != nil {
		return nil, apperror.InternalError(err)
	}

	byCategory := make([]*dto.ExpenseByCategoryResponse, 0)

	// Get total for uncategorized expenses
	uncategorizedParams := &repository.ExpenseListParams{
		BranchID: req.BranchID,
		DateFrom: req.DateFrom,
		DateTo:   req.DateTo,
		Limit:    1,
		Offset:   0,
	}

	// Summarize by category using existing list + aggregation
	for _, cat := range categories {
		catID := cat.ID
		params := &repository.ExpenseListParams{
			BranchID:   req.BranchID,
			CategoryID: &catID,
			DateFrom:   req.DateFrom,
			DateTo:     req.DateTo,
			Limit:      10000,
			Offset:     0,
		}
		catExpenses, catTotal, err := s.expenseRepo.List(ctx, companyID, params)
		if err != nil {
			return nil, apperror.InternalError(err)
		}
		if catTotal > 0 {
			catSum := 0.0
			for _, e := range catExpenses {
				catSum += e.Amount
			}
			catIDPtr := catID
			byCategory = append(byCategory, &dto.ExpenseByCategoryResponse{
				CategoryID:   &catIDPtr,
				CategoryName: cat.Name,
				TotalAmount:  catSum,
				Count:        catTotal,
			})
		}
	}

	// Count uncategorized
	_ = uncategorizedParams
	uncatExpenses, uncatTotal, err := s.expenseRepo.List(ctx, companyID, &repository.ExpenseListParams{
		BranchID:   req.BranchID,
		CategoryID: nil,
		DateFrom:   req.DateFrom,
		DateTo:     req.DateTo,
		Limit:      10000,
		Offset:     0,
	})
	if err != nil {
		return nil, apperror.InternalError(err)
	}

	// Only add uncategorized if there's a way to filter nil category_id separately
	// Here we rely on the list returning all (including uncategorized) when CategoryID is nil
	// So we need a dedicated query; we'll compute it differently
	_ = uncatExpenses
	_ = uncatTotal

	// Simple summary: just return total per category that we already have
	// For uncategorized, compute as total - sum of all categorized
	categorizedTotal := 0.0
	categorizedCount := 0
	for _, item := range byCategory {
		categorizedTotal += item.TotalAmount
		categorizedCount += item.Count
	}

	// Get all expenses to get uncategorized count
	allParams := &repository.ExpenseListParams{
		BranchID: req.BranchID,
		DateFrom: req.DateFrom,
		DateTo:   req.DateTo,
		Limit:    10000,
		Offset:   0,
	}
	allExpenses, allCount, err := s.expenseRepo.List(ctx, companyID, allParams)
	if err != nil {
		return nil, apperror.InternalError(err)
	}

	uncategorizedSum := 0.0
	uncategorizedCount := 0
	for _, e := range allExpenses {
		if e.CategoryID == nil {
			uncategorizedSum += e.Amount
			uncategorizedCount++
		}
	}
	_ = allCount

	if uncategorizedCount > 0 {
		byCategory = append(byCategory, &dto.ExpenseByCategoryResponse{
			CategoryID:   nil,
			CategoryName: fmt.Sprintf("Uncategorized (%d)", uncategorizedCount),
			TotalAmount:  uncategorizedSum,
			Count:        uncategorizedCount,
		})
	}

	return &dto.ExpenseSummaryResponse{
		TotalAmount: total,
		DateFrom:    req.DateFrom,
		DateTo:      req.DateTo,
		ByCategory:  byCategory,
	}, nil
}

func toExpenseCategoryResponse(c *model.ExpenseCategory) *dto.ExpenseCategoryResponse {
	return &dto.ExpenseCategoryResponse{
		ID:          c.ID,
		Name:        c.Name,
		Description: c.Description,
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   c.UpdatedAt,
	}
}

func toExpenseResponse(e *model.Expense) *dto.ExpenseResponse {
	return &dto.ExpenseResponse{
		ID:          e.ID,
		BranchID:    e.BranchID,
		CategoryID:  e.CategoryID,
		Amount:      e.Amount,
		Description: e.Description,
		ReferenceNo: e.ReferenceNo,
		ExpenseDate: e.ExpenseDate,
		RecordedBy:  e.RecordedBy,
		Notes:       e.Notes,
		CreatedAt:   e.CreatedAt,
		UpdatedAt:   e.UpdatedAt,
	}
}

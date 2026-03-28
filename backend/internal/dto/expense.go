package dto

import (
	"time"

	"github.com/irvanmhndra/nexpos-api/pkg/httputil"
)

// ============== Requests ==============

type CreateExpenseCategoryRequest struct {
	Name        string  `json:"name"        validate:"required,min=1,max=100"`
	Description *string `json:"description"`
}

type UpdateExpenseCategoryRequest struct {
	Name        string  `json:"name"        validate:"required,min=1,max=100"`
	Description *string `json:"description"`
}

type ListExpenseCategoryRequest struct {
	Page    int    `query:"page"`
	PerPage int    `query:"per_page"`
	Search  string `query:"search"`
}

type CreateExpenseRequest struct {
	BranchID    *int64  `json:"branch_id"`
	CategoryID  *int64  `json:"category_id"`
	Amount      float64 `json:"amount"       validate:"required,min=0.01"`
	Description string  `json:"description"  validate:"required,min=1,max=500"`
	ReferenceNo *string `json:"reference_no" validate:"omitempty,max=100"`
	ExpenseDate string  `json:"expense_date" validate:"required"`
	Notes       *string `json:"notes"`
}

type UpdateExpenseRequest struct {
	BranchID    *int64  `json:"branch_id"`
	CategoryID  *int64  `json:"category_id"`
	Amount      float64 `json:"amount"       validate:"required,min=0.01"`
	Description string  `json:"description"  validate:"required,min=1,max=500"`
	ReferenceNo *string `json:"reference_no" validate:"omitempty,max=100"`
	ExpenseDate string  `json:"expense_date" validate:"required"`
	Notes       *string `json:"notes"`
}

type ListExpenseRequest struct {
	Page       int    `query:"page"`
	PerPage    int    `query:"per_page"`
	BranchID   *int64 `query:"branch_id"`
	CategoryID *int64 `query:"category_id"`
	DateFrom   string `query:"date_from"`
	DateTo     string `query:"date_to"`
}

type ExpenseSummaryRequest struct {
	BranchID *int64 `query:"branch_id"`
	DateFrom string `query:"date_from"`
	DateTo   string `query:"date_to"`
}

// ============== Responses ==============

type ExpenseCategoryResponse struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ExpenseCategoryListResponse struct {
	Categories []*ExpenseCategoryResponse `json:"categories"`
	Pagination *httputil.Pagination       `json:"pagination"`
}

type ExpenseResponse struct {
	ID          int64     `json:"id"`
	BranchID    *int64    `json:"branch_id"`
	CategoryID  *int64    `json:"category_id"`
	Amount      float64   `json:"amount"`
	Description string    `json:"description"`
	ReferenceNo *string   `json:"reference_no"`
	ExpenseDate string    `json:"expense_date"`
	RecordedBy  *int64    `json:"recorded_by"`
	Notes       *string   `json:"notes"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ExpenseListResponse struct {
	Expenses   []*ExpenseResponse   `json:"expenses"`
	Pagination *httputil.Pagination `json:"pagination"`
}

type ExpenseSummaryResponse struct {
	TotalAmount float64                      `json:"total_amount"`
	DateFrom    string                       `json:"date_from"`
	DateTo      string                       `json:"date_to"`
	ByCategory  []*ExpenseByCategoryResponse `json:"by_category"`
}

type ExpenseByCategoryResponse struct {
	CategoryID   *int64  `json:"category_id"`
	CategoryName string  `json:"category_name"`
	TotalAmount  float64 `json:"total_amount"`
	Count        int     `json:"count"`
}

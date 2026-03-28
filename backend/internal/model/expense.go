package model

import "time"

type ExpenseCategory struct {
	ID          int64     `db:"id"          json:"id"`
	CompanyID   int64     `db:"company_id"  json:"company_id"`
	Name        string    `db:"name"        json:"name"`
	Description *string   `db:"description" json:"description"`
	CreatedAt   time.Time `db:"created_at"  json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"  json:"updated_at"`
}

type Expense struct {
	ID          int64     `db:"id"           json:"id"`
	CompanyID   int64     `db:"company_id"   json:"company_id"`
	BranchID    *int64    `db:"branch_id"    json:"branch_id"`
	CategoryID  *int64    `db:"category_id"  json:"category_id"`
	Amount      float64   `db:"amount"       json:"amount"`
	Description string    `db:"description"  json:"description"`
	ReferenceNo *string   `db:"reference_no" json:"reference_no"`
	ExpenseDate string    `db:"expense_date" json:"expense_date"`
	RecordedBy  *int64    `db:"recorded_by"  json:"recorded_by"`
	Notes       *string   `db:"notes"        json:"notes"`
	CreatedAt   time.Time `db:"created_at"   json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"   json:"updated_at"`
}

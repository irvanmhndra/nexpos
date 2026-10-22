package model

import "time"

type UserBranch struct {
	ID        int64     `db:"id"`
	UserID    int64     `db:"user_id"`
	BranchID  int64     `db:"branch_id"`
	IsDefault bool      `db:"is_default"`
	CreatedAt time.Time `db:"created_at"`
}

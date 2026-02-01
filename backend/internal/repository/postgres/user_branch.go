package postgres

import (
	"context"
	"database/sql"

	"github.com/irvanmhndra/pos-core-api/internal/model"
	"github.com/jmoiron/sqlx"
)

type UserBranchRepository struct {
	db *sqlx.DB
}

func NewUserBranchRepository(db *sqlx.DB) *UserBranchRepository {
	return &UserBranchRepository{db: db}
}

func (r *UserBranchRepository) Create(ctx context.Context, ub *model.UserBranch) error {
	return r.db.QueryRowContext(ctx, `
		INSERT INTO user_branches (user_id, branch_id, is_default)
		VALUES ($1, $2, $3)
		RETURNING id, created_at
	`, ub.UserID, ub.BranchID, ub.IsDefault).Scan(&ub.ID, &ub.CreatedAt)
}

func (r *UserBranchRepository) Delete(ctx context.Context, userID, branchID int64) error {
	_, err := r.db.ExecContext(ctx, `
		DELETE FROM user_branches WHERE user_id = $1 AND branch_id = $2
	`, userID, branchID)
	return err
}

func (r *UserBranchRepository) SetBranches(ctx context.Context, userID int64, branchIDs []int64, defaultBranchID int64) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Delete existing branches
	_, err = tx.ExecContext(ctx, `DELETE FROM user_branches WHERE user_id = $1`, userID)
	if err != nil {
		return err
	}

	// Insert new branches
	for _, branchID := range branchIDs {
		isDefault := branchID == defaultBranchID
		_, err = tx.ExecContext(ctx, `
			INSERT INTO user_branches (user_id, branch_id, is_default) VALUES ($1, $2, $3)
		`, userID, branchID, isDefault)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *UserBranchRepository) GetByUserID(ctx context.Context, userID int64) ([]*model.UserBranch, error) {
	var userBranches []*model.UserBranch
	err := r.db.SelectContext(ctx, &userBranches, `
		SELECT id, user_id, branch_id, is_default, created_at
		FROM user_branches WHERE user_id = $1
		ORDER BY is_default DESC, created_at
	`, userID)
	if err != nil {
		return nil, err
	}
	return userBranches, nil
}

func (r *UserBranchRepository) GetDefaultBranch(ctx context.Context, userID int64) (*model.Branch, error) {
	var branch model.Branch
	err := r.db.GetContext(ctx, &branch, `
		SELECT b.id, b.company_id, b.code, b.name, b.address, b.phone, b.is_active, b.created_at, b.updated_at
		FROM branches b
		JOIN user_branches ub ON ub.branch_id = b.id
		WHERE ub.user_id = $1 AND ub.is_default = true
		LIMIT 1
	`, userID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &branch, nil
}

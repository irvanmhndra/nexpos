package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/irvanmhndra/pos-core-api/internal/model"
	"github.com/irvanmhndra/pos-core-api/internal/repository"
	"github.com/jmoiron/sqlx"
)

type branchRepository struct {
	db *sqlx.DB
}

func NewBranchRepository(db *sqlx.DB) repository.BranchRepository {
	return &branchRepository{db: db}
}

func (r *branchRepository) Create(ctx context.Context, branch *model.Branch) error {
	query := `
		INSERT INTO branches (company_id, code, name, address, phone, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`
	return r.db.QueryRowxContext(ctx, query,
		branch.CompanyID,
		branch.Code,
		branch.Name,
		branch.Address,
		branch.Phone,
		branch.IsActive,
	).Scan(&branch.ID, &branch.CreatedAt, &branch.UpdatedAt)
}

func (r *branchRepository) GetByID(ctx context.Context, id int64) (*model.Branch, error) {
	var branch model.Branch
	query := `SELECT * FROM branches WHERE id = $1`
	err := r.db.GetContext(ctx, &branch, query, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &branch, nil
}

func (r *branchRepository) ListByCompanyID(ctx context.Context, companyID int64) ([]*model.Branch, error) {
	var branches []*model.Branch
	query := `SELECT * FROM branches WHERE company_id = $1 ORDER BY name`
	err := r.db.SelectContext(ctx, &branches, query, companyID)
	if err != nil {
		return nil, err
	}
	return branches, nil
}

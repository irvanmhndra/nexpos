package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/irvanmhndra/nexpos-api/internal/model"
	"github.com/irvanmhndra/nexpos-api/internal/repository"
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
	return conn(ctx, r.db).QueryRowxContext(ctx, query,
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
	err := conn(ctx, r.db).GetContext(ctx, &branch, query, id)
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
	err := conn(ctx, r.db).SelectContext(ctx, &branches, query, companyID)
	if err != nil {
		return nil, err
	}
	return branches, nil
}

func (r *branchRepository) List(ctx context.Context, companyID int64, search string, isActive *bool, limit, offset int) ([]*model.Branch, int, error) {
	args := []any{companyID}
	conditions := []string{"company_id = $1"}
	argIdx := 2

	if search != "" {
		conditions = append(conditions, fmt.Sprintf(
			"(name ILIKE $%d OR code ILIKE $%d OR address ILIKE $%d OR phone ILIKE $%d)",
			argIdx, argIdx+1, argIdx+2, argIdx+3,
		))
		like := "%" + search + "%"
		args = append(args, like, like, like, like)
		argIdx += 4
	}

	if isActive != nil {
		conditions = append(conditions, fmt.Sprintf("is_active = $%d", argIdx))
		args = append(args, *isActive)
		argIdx++
	}

	where := strings.Join(conditions, " AND ")

	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM branches WHERE %s", where)
	if err := conn(ctx, r.db).GetContext(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, err
	}

	listArgs := append(args, limit, offset)
	listQuery := fmt.Sprintf(
		"SELECT * FROM branches WHERE %s ORDER BY name LIMIT $%d OFFSET $%d",
		where, argIdx, argIdx+1,
	)

	var branches []*model.Branch
	if err := conn(ctx, r.db).SelectContext(ctx, &branches, listQuery, listArgs...); err != nil {
		return nil, 0, err
	}

	return branches, total, nil
}

func (r *branchRepository) Update(ctx context.Context, branch *model.Branch) error {
	query := `
		UPDATE branches
		SET code = $1, name = $2, address = $3, phone = $4, is_active = $5, updated_at = NOW()
		WHERE id = $6
		RETURNING updated_at
	`
	return conn(ctx, r.db).QueryRowxContext(ctx, query,
		branch.Code,
		branch.Name,
		branch.Address,
		branch.Phone,
		branch.IsActive,
		branch.ID,
	).Scan(&branch.UpdatedAt)
}

func (r *branchRepository) Delete(ctx context.Context, id int64) error {
	_, err := conn(ctx, r.db).ExecContext(ctx, `DELETE FROM branches WHERE id = $1`, id)
	return err
}

func (r *branchRepository) CodeExists(ctx context.Context, companyID int64, code string, excludeID int64) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM branches WHERE company_id = $1 AND code = $2 AND id != $3`
	err := conn(ctx, r.db).GetContext(ctx, &count, query, companyID, code, excludeID)
	return count > 0, err
}

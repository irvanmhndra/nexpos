package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/irvanmhndra/nexpos-api/internal/model"
	"github.com/irvanmhndra/nexpos-api/internal/repository"
	"github.com/jmoiron/sqlx"
)

type companyRepository struct {
	db *sqlx.DB
}

func NewCompanyRepository(db *sqlx.DB) repository.CompanyRepository {
	return &companyRepository{db: db}
}

func (r *companyRepository) Create(ctx context.Context, company *model.Company) error {
	query := `
		INSERT INTO companies (code, name, address, phone, email, tax_id, logo_url, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`
	return r.db.QueryRowxContext(ctx, query,
		company.Code,
		company.Name,
		company.Address,
		company.Phone,
		company.Email,
		company.TaxID,
		company.LogoURL,
		company.IsActive,
	).Scan(&company.ID, &company.CreatedAt, &company.UpdatedAt)
}

func (r *companyRepository) GetByID(ctx context.Context, id int64) (*model.Company, error) {
	var company model.Company
	query := `SELECT * FROM companies WHERE id = $1`
	err := r.db.GetContext(ctx, &company, query, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &company, nil
}

func (r *companyRepository) GetByCode(ctx context.Context, code string) (*model.Company, error) {
	var company model.Company
	query := `SELECT * FROM companies WHERE code = $1`
	err := r.db.GetContext(ctx, &company, query, code)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &company, nil
}

func (r *companyRepository) Update(ctx context.Context, company *model.Company) error {
	query := `
		UPDATE companies
		SET name = $1, address = $2, phone = $3, email = $4, tax_id = $5, logo_url = $6, is_active = $7, updated_at = NOW()
		WHERE id = $8
	`
	_, err := r.db.ExecContext(ctx, query,
		company.Name,
		company.Address,
		company.Phone,
		company.Email,
		company.TaxID,
		company.LogoURL,
		company.IsActive,
		company.ID,
	)
	return err
}

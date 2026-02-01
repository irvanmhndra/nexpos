package postgres

import (
	"context"
	"database/sql"

	"github.com/irvanmhndra/pos-core-api/internal/model"
	"github.com/jmoiron/sqlx"
)

type RoleRepository struct {
	db *sqlx.DB
}

func NewRoleRepository(db *sqlx.DB) *RoleRepository {
	return &RoleRepository{db: db}
}

func (r *RoleRepository) GetByID(ctx context.Context, id int64) (*model.Role, error) {
	var role model.Role
	err := r.db.GetContext(ctx, &role, `
		SELECT id, company_id, code, name, description, is_system, is_active, created_at, updated_at
		FROM roles WHERE id = $1
	`, id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *RoleRepository) GetByCode(ctx context.Context, code string, companyID *int64) (*model.Role, error) {
	var role model.Role
	var err error

	if companyID != nil {
		// Check company-specific role first, then fall back to system role
		err = r.db.GetContext(ctx, &role, `
			SELECT id, company_id, code, name, description, is_system, is_active, created_at, updated_at
			FROM roles
			WHERE code = $1 AND (company_id = $2 OR company_id IS NULL)
			ORDER BY company_id NULLS LAST
			LIMIT 1
		`, code, *companyID)
	} else {
		// Only system roles
		err = r.db.GetContext(ctx, &role, `
			SELECT id, company_id, code, name, description, is_system, is_active, created_at, updated_at
			FROM roles WHERE code = $1 AND company_id IS NULL
		`, code)
	}

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *RoleRepository) ListByCompanyID(ctx context.Context, companyID int64) ([]*model.Role, error) {
	var roles []*model.Role
	err := r.db.SelectContext(ctx, &roles, `
		SELECT id, company_id, code, name, description, is_system, is_active, created_at, updated_at
		FROM roles
		WHERE (company_id = $1 OR company_id IS NULL) AND is_active = true
		ORDER BY is_system DESC, name
	`, companyID)
	if err != nil {
		return nil, err
	}
	return roles, nil
}

func (r *RoleRepository) ListSystemRoles(ctx context.Context) ([]*model.Role, error) {
	var roles []*model.Role
	err := r.db.SelectContext(ctx, &roles, `
		SELECT id, company_id, code, name, description, is_system, is_active, created_at, updated_at
		FROM roles WHERE company_id IS NULL AND is_active = true
		ORDER BY name
	`)
	if err != nil {
		return nil, err
	}
	return roles, nil
}

func (r *RoleRepository) Create(ctx context.Context, role *model.Role) error {
	return r.db.QueryRowContext(ctx, `
		INSERT INTO roles (company_id, code, name, description, is_system, is_active)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at
	`, role.CompanyID, role.Code, role.Name, role.Description, role.IsSystem, role.IsActive).
		Scan(&role.ID, &role.CreatedAt, &role.UpdatedAt)
}

func (r *RoleRepository) Update(ctx context.Context, role *model.Role) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE roles SET name = $1, description = $2, is_active = $3, updated_at = NOW()
		WHERE id = $4
	`, role.Name, role.Description, role.IsActive, role.ID)
	return err
}

func (r *RoleRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM roles WHERE id = $1 AND is_system = false`, id)
	return err
}

func (r *RoleRepository) GetWithPermissions(ctx context.Context, id int64) (*model.Role, error) {
	role, err := r.GetByID(ctx, id)
	if err != nil || role == nil {
		return role, err
	}

	var permissions []model.Permission
	err = r.db.SelectContext(ctx, &permissions, `
		SELECT p.id, p.code, p.name, p.module, p.description, p.created_at
		FROM permissions p
		JOIN role_permissions rp ON rp.permission_id = p.id
		WHERE rp.role_id = $1
		ORDER BY p.module, p.code
	`, id)
	if err != nil {
		return nil, err
	}

	role.Permissions = permissions
	return role, nil
}

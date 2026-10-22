package postgres

import (
	"context"

	"github.com/irvanmhndra/nexpos-api/internal/model"
	"github.com/jmoiron/sqlx"
)

type PermissionRepository struct {
	db *sqlx.DB
}

func NewPermissionRepository(db *sqlx.DB) *PermissionRepository {
	return &PermissionRepository{db: db}
}

func (r *PermissionRepository) GetAll(ctx context.Context) ([]*model.Permission, error) {
	var permissions []*model.Permission
	err := conn(ctx, r.db).SelectContext(ctx, &permissions, `
		SELECT id, code, name, module, description, created_at
		FROM permissions
		ORDER BY module, code
	`)
	if err != nil {
		return nil, err
	}
	return permissions, nil
}

func (r *PermissionRepository) GetByModule(ctx context.Context, module string) ([]*model.Permission, error) {
	var permissions []*model.Permission
	err := conn(ctx, r.db).SelectContext(ctx, &permissions, `
		SELECT id, code, name, module, description, created_at
		FROM permissions WHERE module = $1
		ORDER BY code
	`, module)
	if err != nil {
		return nil, err
	}
	return permissions, nil
}

func (r *PermissionRepository) GetByRoleID(ctx context.Context, roleID int64) ([]*model.Permission, error) {
	var permissions []*model.Permission
	err := conn(ctx, r.db).SelectContext(ctx, &permissions, `
		SELECT p.id, p.code, p.name, p.module, p.description, p.created_at
		FROM permissions p
		JOIN role_permissions rp ON rp.permission_id = p.id
		WHERE rp.role_id = $1
		ORDER BY p.module, p.code
	`, roleID)
	if err != nil {
		return nil, err
	}
	return permissions, nil
}

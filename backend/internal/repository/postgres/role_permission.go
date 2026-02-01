package postgres

import (
	"context"

	"github.com/jmoiron/sqlx"
)

type RolePermissionRepository struct {
	db *sqlx.DB
}

func NewRolePermissionRepository(db *sqlx.DB) *RolePermissionRepository {
	return &RolePermissionRepository{db: db}
}

func (r *RolePermissionRepository) SetPermissions(ctx context.Context, roleID int64, permissionIDs []int64) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Delete existing permissions
	_, err = tx.ExecContext(ctx, `DELETE FROM role_permissions WHERE role_id = $1`, roleID)
	if err != nil {
		return err
	}

	// Insert new permissions
	if len(permissionIDs) > 0 {
		for _, permID := range permissionIDs {
			_, err = tx.ExecContext(ctx, `
				INSERT INTO role_permissions (role_id, permission_id) VALUES ($1, $2)
			`, roleID, permID)
			if err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}

func (r *RolePermissionRepository) GetPermissionIDs(ctx context.Context, roleID int64) ([]int64, error) {
	var ids []int64
	err := r.db.SelectContext(ctx, &ids, `
		SELECT permission_id FROM role_permissions WHERE role_id = $1
	`, roleID)
	if err != nil {
		return nil, err
	}
	return ids, nil
}

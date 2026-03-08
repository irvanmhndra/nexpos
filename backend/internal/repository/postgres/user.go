package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/irvanmhndra/nexpos-api/internal/model"
	"github.com/irvanmhndra/nexpos-api/internal/repository"
	"github.com/jmoiron/sqlx"
)

type userRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) repository.UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *model.User) error {
	query := `
		INSERT INTO users (company_id, role_id, email, password_hash, name, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`
	return r.db.QueryRowxContext(ctx, query,
		user.CompanyID,
		user.RoleID,
		user.Email,
		user.PasswordHash,
		user.Name,
		user.Status,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
}

func (r *userRepository) GetByID(ctx context.Context, companyID, id int64) (*model.User, error) {
	var user model.User
	query := `SELECT * FROM users WHERE id = $1 AND company_id = $2`
	err := r.db.GetContext(ctx, &user, query, id, companyID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User
	query := `SELECT * FROM users WHERE email = $1`
	err := r.db.GetContext(ctx, &user, query, email)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) List(ctx context.Context, companyID int64, limit, offset int) ([]*model.User, int, error) {
	var users []*model.User
	var total int

	countQuery := `SELECT COUNT(*) FROM users WHERE company_id = $1`
	if err := r.db.GetContext(ctx, &total, countQuery, companyID); err != nil {
		return nil, 0, err
	}

	query := `
		SELECT * FROM users
		WHERE company_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	if err := r.db.SelectContext(ctx, &users, query, companyID, limit, offset); err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

func (r *userRepository) Update(ctx context.Context, user *model.User) error {
	query := `
		UPDATE users
		SET email = $1, name = $2, role_id = $3, status = $4, updated_at = NOW()
		WHERE id = $5 AND company_id = $6
	`
	result, err := r.db.ExecContext(ctx, query,
		user.Email,
		user.Name,
		user.RoleID,
		user.Status,
		user.ID,
		user.CompanyID,
	)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *userRepository) UpdateStatus(ctx context.Context, companyID, id int64, status model.UserStatus) error {
	query := `UPDATE users SET status = $1, updated_at = NOW() WHERE id = $2 AND company_id = $3`
	result, err := r.db.ExecContext(ctx, query, status, id, companyID)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *userRepository) Delete(ctx context.Context, companyID, id int64) error {
	query := `DELETE FROM users WHERE id = $1 AND company_id = $2`
	result, err := r.db.ExecContext(ctx, query, id, companyID)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *userRepository) EmailExists(ctx context.Context, email string, excludeID int64) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1 AND id != $2)`
	err := r.db.GetContext(ctx, &exists, query, email, excludeID)
	return exists, err
}

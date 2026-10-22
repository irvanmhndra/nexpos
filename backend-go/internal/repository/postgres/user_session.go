package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/irvanmhndra/nexpos-api/internal/model"
	"github.com/irvanmhndra/nexpos-api/internal/repository"
	"github.com/jmoiron/sqlx"
)

type userSessionRepository struct {
	db *sqlx.DB
}

func NewUserSessionRepository(db *sqlx.DB) repository.UserSessionRepository {
	return &userSessionRepository{db: db}
}

func (r *userSessionRepository) Create(ctx context.Context, session *model.UserSession) error {
	query := `
		INSERT INTO user_sessions (
			user_id, company_id, branch_id,
			access_token_hash, access_token_expires_at,
			refresh_token_hash, refresh_token_expires_at,
			is_revoked, device_info, ip_address, user_agent, created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, NOW())
		RETURNING id, created_at
	`
	return r.db.QueryRowxContext(ctx, query,
		session.UserID,
		session.CompanyID,
		session.BranchID,
		session.AccessTokenHash,
		session.AccessTokenExpiresAt,
		session.RefreshTokenHash,
		session.RefreshTokenExpiresAt,
		session.IsRevoked,
		session.DeviceInfo,
		session.IPAddress,
		session.UserAgent,
	).Scan(&session.ID, &session.CreatedAt)
}

func (r *userSessionRepository) GetByAccessTokenHash(ctx context.Context, accessTokenHash string) (*model.UserSession, error) {
	var session model.UserSession
	query := `SELECT * FROM user_sessions WHERE access_token_hash = $1 AND NOT is_revoked`
	err := r.db.GetContext(ctx, &session, query, accessTokenHash)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *userSessionRepository) GetByRefreshTokenHash(ctx context.Context, refreshTokenHash string) (*model.UserSession, error) {
	var session model.UserSession
	query := `SELECT * FROM user_sessions WHERE refresh_token_hash = $1 AND NOT is_revoked`
	err := r.db.GetContext(ctx, &session, query, refreshTokenHash)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *userSessionRepository) Rotate(ctx context.Context, oldID int64, next *model.UserSession) (bool, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return false, err
	}
	defer func() { _ = tx.Rollback() }()

	// The conditional update is the claim: of two concurrent refreshes with
	// the same token, only one sees a row affected.
	res, err := tx.ExecContext(ctx, `UPDATE user_sessions SET is_revoked = true WHERE id = $1 AND NOT is_revoked`, oldID)
	if err != nil {
		return false, err
	}
	if n, err := res.RowsAffected(); err != nil || n == 0 {
		return false, err
	}

	err = tx.QueryRowxContext(ctx, `
		INSERT INTO user_sessions (
			user_id, company_id, branch_id,
			access_token_hash, access_token_expires_at,
			refresh_token_hash, refresh_token_expires_at,
			is_revoked, device_info, ip_address, user_agent, created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, false, $8, $9, $10, NOW())
		RETURNING id, created_at`,
		next.UserID, next.CompanyID, next.BranchID,
		next.AccessTokenHash, next.AccessTokenExpiresAt,
		next.RefreshTokenHash, next.RefreshTokenExpiresAt,
		next.DeviceInfo, next.IPAddress, next.UserAgent,
	).Scan(&next.ID, &next.CreatedAt)
	if err != nil {
		return false, err
	}
	return true, tx.Commit()
}

func (r *userSessionRepository) UpdateLastUsed(ctx context.Context, id int64) error {
	query := `UPDATE user_sessions SET last_used_at = NOW() WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *userSessionRepository) Revoke(ctx context.Context, id int64) error {
	query := `UPDATE user_sessions SET is_revoked = true WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *userSessionRepository) RevokeAllByUserID(ctx context.Context, userID int64) error {
	query := `UPDATE user_sessions SET is_revoked = true WHERE user_id = $1 AND NOT is_revoked`
	_, err := r.db.ExecContext(ctx, query, userID)
	return err
}

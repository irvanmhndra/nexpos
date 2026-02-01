package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/irvanmhndra/pos-core-api/internal/model"
	"github.com/irvanmhndra/pos-core-api/internal/repository"
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
			access_token, access_token_expires_at,
			refresh_token, refresh_token_expires_at,
			is_revoked, device_info, ip_address, user_agent, created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, NOW())
		RETURNING id, created_at
	`
	return r.db.QueryRowxContext(ctx, query,
		session.UserID,
		session.CompanyID,
		session.BranchID,
		session.AccessToken,
		session.AccessTokenExpiresAt,
		session.RefreshToken,
		session.RefreshTokenExpiresAt,
		session.IsRevoked,
		session.DeviceInfo,
		session.IPAddress,
		session.UserAgent,
	).Scan(&session.ID, &session.CreatedAt)
}

func (r *userSessionRepository) GetByAccessToken(ctx context.Context, accessToken string) (*model.UserSession, error) {
	var session model.UserSession
	query := `SELECT * FROM user_sessions WHERE access_token = $1 AND NOT is_revoked`
	err := r.db.GetContext(ctx, &session, query, accessToken)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *userSessionRepository) GetByRefreshToken(ctx context.Context, refreshToken string) (*model.UserSession, error) {
	var session model.UserSession
	query := `SELECT * FROM user_sessions WHERE refresh_token = $1 AND NOT is_revoked`
	err := r.db.GetContext(ctx, &session, query, refreshToken)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &session, nil
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

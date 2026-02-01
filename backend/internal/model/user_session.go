package model

import "time"

type UserSession struct {
	ID                    int64      `db:"id"`
	UserID                int64      `db:"user_id"`
	CompanyID             int64      `db:"company_id"`
	BranchID              *int64     `db:"branch_id"`
	AccessToken           string     `db:"access_token"`
	AccessTokenExpiresAt  time.Time  `db:"access_token_expires_at"`
	RefreshToken          string     `db:"refresh_token"`
	RefreshTokenExpiresAt time.Time  `db:"refresh_token_expires_at"`
	IsRevoked             bool       `db:"is_revoked"`
	DeviceInfo            *string    `db:"device_info"`
	IPAddress             *string    `db:"ip_address"`
	UserAgent             *string    `db:"user_agent"`
	LastUsedAt            *time.Time `db:"last_used_at"`
	CreatedAt             time.Time  `db:"created_at"`
}

func (s *UserSession) IsAccessTokenExpired() bool {
	return time.Now().After(s.AccessTokenExpiresAt)
}

func (s *UserSession) IsRefreshTokenExpired() bool {
	return time.Now().After(s.RefreshTokenExpiresAt)
}

package middleware

import (
	"net/http"
	"strings"

	"github.com/irvanmhndra/nexpos-api/internal/repository"
	"github.com/irvanmhndra/nexpos-api/pkg/httputil"
	"github.com/labstack/echo/v5"
)

func Auth(sessionRepo repository.UserSessionRepository) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				return c.JSON(http.StatusUnauthorized, httputil.Response{
					Success:   false,
					Message:   "unauthorized",
					ErrorCode: "UNAUTHORIZED",
				})
			}

			token := strings.TrimPrefix(authHeader, "Bearer ")
			ctx := c.Request().Context()

			session, err := sessionRepo.GetByAccessToken(ctx, token)
			if err != nil || session == nil {
				return c.JSON(http.StatusUnauthorized, httputil.Response{
					Success:   false,
					Message:   "invalid or expired token",
					ErrorCode: "UNAUTHORIZED",
				})
			}

			if session.IsAccessTokenExpired() {
				return c.JSON(http.StatusUnauthorized, httputil.Response{
					Success:   false,
					Message:   "token expired",
					ErrorCode: "TOKEN_EXPIRED",
				})
			}

			// Update last used (best-effort)
			_ = sessionRepo.UpdateLastUsed(ctx, session.ID)

			// Set auth context values
			c.Set("user_id", session.UserID)
			c.Set("company_id", session.CompanyID)
			if session.BranchID != nil {
				c.Set("branch_id", *session.BranchID)
			}

			return next(c)
		}
	}
}

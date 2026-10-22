package handler

import (
	"net/http"

	"github.com/irvanmhndra/nexpos-api/pkg/httputil"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v5"
)

type HealthHandler struct {
	db *sqlx.DB
}

func NewHealthHandler(db *sqlx.DB) *HealthHandler {
	return &HealthHandler{db: db}
}

func (h *HealthHandler) Check(c *echo.Context) error {
	status := "healthy"
	dbStatus := "connected"

	// Check database connectivity
	if err := h.db.Ping(); err != nil {
		dbStatus = "disconnected"
		status = "unhealthy"
	}

	data := map[string]any{
		"status": status,
		"services": map[string]string{
			"database": dbStatus,
		},
	}

	statusCode := http.StatusOK
	if status != "healthy" {
		statusCode = http.StatusServiceUnavailable
	}

	return httputil.Success(c, statusCode, "Health check completed", data)
}

func (h *HealthHandler) Liveness(c *echo.Context) error {
	return httputil.Success(c, http.StatusOK, "OK", map[string]string{"status": "alive"})
}

func (h *HealthHandler) Readiness(c *echo.Context) error {
	if err := h.db.Ping(); err != nil {
		return httputil.Success(c, http.StatusServiceUnavailable, "Not ready", map[string]string{"status": "not ready"})
	}
	return httputil.Success(c, http.StatusOK, "Ready", map[string]string{"status": "ready"})
}

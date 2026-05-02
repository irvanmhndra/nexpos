package handler

import (
	"net/http"
	"strings"

	"github.com/irvanmhndra/nexpos-api/internal/dto"
	"github.com/irvanmhndra/nexpos-api/internal/service"
	"github.com/irvanmhndra/nexpos-api/pkg/httputil"
	"github.com/irvanmhndra/nexpos-api/pkg/validator"
	"github.com/labstack/echo/v5"
)

type AuthHandler struct {
	authSvc   service.AuthServiceInterface
	validator *validator.CustomValidator
}

func NewAuthHandler(authSvc service.AuthServiceInterface, validator *validator.CustomValidator) *AuthHandler {
	return &AuthHandler{
		authSvc:   authSvc,
		validator: validator,
	}
}

func (h *AuthHandler) Login(c *echo.Context) error {
	var req dto.LoginRequest
	if err := c.Bind(&req); err != nil {
		return httputil.Error(c, httputil.BindError(err))
	}

	if fieldErrors := h.validator.ValidateAndFormat(&req); fieldErrors != nil {
		return httputil.ValidationError(c, fieldErrors)
	}

	ctx := c.Request().Context()
	ipAddress := c.RealIP()
	userAgent := c.Request().UserAgent()

	result, err := h.authSvc.Login(ctx, req, ipAddress, userAgent)
	if err != nil {
		return httputil.Error(c, err)
	}

	return httputil.Success(c, http.StatusOK, "Login successful", result)
}

func (h *AuthHandler) Register(c *echo.Context) error {
	var req dto.RegisterRequest
	if err := c.Bind(&req); err != nil {
		return httputil.Error(c, httputil.BindError(err))
	}

	if fieldErrors := h.validator.ValidateAndFormat(&req); fieldErrors != nil {
		return httputil.ValidationError(c, fieldErrors)
	}

	ctx := c.Request().Context()
	ipAddress := c.RealIP()
	userAgent := c.Request().UserAgent()

	result, err := h.authSvc.Register(ctx, req, ipAddress, userAgent)
	if err != nil {
		return httputil.Error(c, err)
	}

	return httputil.Success(c, http.StatusCreated, "Registration successful", result)
}

func (h *AuthHandler) RefreshToken(c *echo.Context) error {
	var req dto.RefreshTokenRequest
	if err := c.Bind(&req); err != nil {
		return httputil.Error(c, httputil.BindError(err))
	}

	if fieldErrors := h.validator.ValidateAndFormat(&req); fieldErrors != nil {
		return httputil.ValidationError(c, fieldErrors)
	}

	ctx := c.Request().Context()

	result, err := h.authSvc.RefreshToken(ctx, req.RefreshToken)
	if err != nil {
		return httputil.Error(c, err)
	}

	return httputil.Success(c, http.StatusOK, "Token refreshed successfully", result)
}

func (h *AuthHandler) Logout(c *echo.Context) error {
	ctx := c.Request().Context()

	// Extract token from Authorization header
	authHeader := c.Request().Header.Get("Authorization")
	accessToken := strings.TrimPrefix(authHeader, "Bearer ")

	if accessToken != "" && accessToken != authHeader {
		_ = h.authSvc.Logout(ctx, accessToken)
	}

	return httputil.Success(c, http.StatusOK, "Logged out successfully", nil)
}

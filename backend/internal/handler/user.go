package handler

import (
	"net/http"
	"strconv"

	"github.com/irvanmhndra/nexpos-api/internal/dto"
	"github.com/irvanmhndra/nexpos-api/internal/service"
	"github.com/irvanmhndra/nexpos-api/pkg/apperror"
	"github.com/irvanmhndra/nexpos-api/pkg/httputil"
	"github.com/irvanmhndra/nexpos-api/pkg/validator"
	"github.com/labstack/echo/v5"
)

type UserHandler struct {
	userSvc   *service.UserService
	validator *validator.CustomValidator
}

func NewUserHandler(userSvc *service.UserService, validator *validator.CustomValidator) *UserHandler {
	return &UserHandler{
		userSvc:   userSvc,
		validator: validator,
	}
}

func (h *UserHandler) Create(c *echo.Context) error {
	var req dto.CreateUserRequest
	if err := c.Bind(&req); err != nil {
		return httputil.Error(c, apperror.BadRequest("invalid request body"))
	}

	if fieldErrors := h.validator.ValidateAndFormat(&req); fieldErrors != nil {
		return httputil.ValidationError(c, fieldErrors)
	}

	ctx := c.Request().Context()
	companyID := getCompanyID(c)

	result, err := h.userSvc.Create(ctx, companyID, req)
	if err != nil {
		return httputil.Error(c, err)
	}

	return httputil.Success(c, http.StatusCreated, "User created successfully", result)
}

func (h *UserHandler) Get(c *echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return httputil.Error(c, apperror.BadRequest("invalid user id"))
	}

	ctx := c.Request().Context()
	companyID := getCompanyID(c)

	result, err := h.userSvc.GetByID(ctx, companyID, id)
	if err != nil {
		return httputil.Error(c, err)
	}

	return httputil.Success(c, http.StatusOK, "User retrieved successfully", result)
}

func (h *UserHandler) List(c *echo.Context) error {
	var req dto.ListUsersRequest
	if err := c.Bind(&req); err != nil {
		return httputil.Error(c, apperror.BadRequest("invalid query params"))
	}

	ctx := c.Request().Context()
	companyID := getCompanyID(c)

	result, err := h.userSvc.List(ctx, companyID, req)
	if err != nil {
		return httputil.Error(c, err)
	}

	return httputil.SuccessWithPagination(c, http.StatusOK, "User list retrieved", result.Users, result.Pagination, nil)
}

func (h *UserHandler) Update(c *echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return httputil.Error(c, apperror.BadRequest("invalid user id"))
	}

	var req dto.UpdateUserRequest
	if err := c.Bind(&req); err != nil {
		return httputil.Error(c, apperror.BadRequest("invalid request body"))
	}

	if fieldErrors := h.validator.ValidateAndFormat(&req); fieldErrors != nil {
		return httputil.ValidationError(c, fieldErrors)
	}

	ctx := c.Request().Context()
	companyID := getCompanyID(c)

	result, err := h.userSvc.Update(ctx, companyID, id, req)
	if err != nil {
		return httputil.Error(c, err)
	}

	return httputil.Success(c, http.StatusOK, "User updated successfully", result)
}

func (h *UserHandler) UpdateStatus(c *echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return httputil.Error(c, apperror.BadRequest("invalid user id"))
	}

	var req dto.UpdateUserStatusRequest
	if err := c.Bind(&req); err != nil {
		return httputil.Error(c, apperror.BadRequest("invalid request body"))
	}

	if fieldErrors := h.validator.ValidateAndFormat(&req); fieldErrors != nil {
		return httputil.ValidationError(c, fieldErrors)
	}

	ctx := c.Request().Context()
	companyID := getCompanyID(c)

	if err := h.userSvc.UpdateStatus(ctx, companyID, id, req); err != nil {
		return httputil.Error(c, err)
	}

	return httputil.Success(c, http.StatusOK, "User status updated successfully", nil)
}

func (h *UserHandler) Delete(c *echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return httputil.Error(c, apperror.BadRequest("invalid user id"))
	}

	ctx := c.Request().Context()
	companyID := getCompanyID(c)

	if err := h.userSvc.Delete(ctx, companyID, id); err != nil {
		return httputil.Error(c, err)
	}

	return httputil.Success(c, http.StatusOK, "User deleted successfully", nil)
}

// Context helpers
func getCompanyID(c *echo.Context) int64 {
	if v := c.Get("company_id"); v != nil {
		return v.(int64)
	}
	return 1 // Default for development
}

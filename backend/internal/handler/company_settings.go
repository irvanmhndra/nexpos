package handler

import (
	"net/http"

	"github.com/irvanmhndra/nexpos-api/internal/dto"
	"github.com/irvanmhndra/nexpos-api/internal/service"
	"github.com/irvanmhndra/nexpos-api/pkg/apperror"
	"github.com/irvanmhndra/nexpos-api/pkg/httputil"
	"github.com/irvanmhndra/nexpos-api/pkg/validator"
	"github.com/labstack/echo/v5"
)

type CompanySettingsHandler struct {
	settingsSvc *service.CompanySettingsService
	validator   *validator.CustomValidator
}

func NewCompanySettingsHandler(settingsSvc *service.CompanySettingsService, v *validator.CustomValidator) *CompanySettingsHandler {
	return &CompanySettingsHandler{settingsSvc: settingsSvc, validator: v}
}

func (h *CompanySettingsHandler) Get(c *echo.Context) error {
	ctx := c.Request().Context()
	companyID := getCompanyID(c)

	result, err := h.settingsSvc.Get(ctx, companyID)
	if err != nil {
		return httputil.Error(c, err)
	}

	return httputil.Success(c, http.StatusOK, "Company settings retrieved successfully", result)
}

func (h *CompanySettingsHandler) Update(c *echo.Context) error {
	var req dto.UpdateCompanySettingsRequest
	if err := c.Bind(&req); err != nil {
		return httputil.Error(c, apperror.BadRequest("invalid request body"))
	}

	if fieldErrors := h.validator.ValidateAndFormat(&req); fieldErrors != nil {
		return httputil.ValidationError(c, fieldErrors)
	}

	ctx := c.Request().Context()
	companyID := getCompanyID(c)

	result, err := h.settingsSvc.Update(ctx, companyID, req)
	if err != nil {
		return httputil.Error(c, err)
	}

	return httputil.Success(c, http.StatusOK, "Company settings updated successfully", result)
}

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

type PromotionHandler struct {
	promotionSvc service.PromotionServiceInterface
	validator    *validator.CustomValidator
}

func NewPromotionHandler(promotionSvc service.PromotionServiceInterface, validator *validator.CustomValidator) *PromotionHandler {
	return &PromotionHandler{
		promotionSvc: promotionSvc,
		validator:    validator,
	}
}

func (h *PromotionHandler) Create(c *echo.Context) error {
	var req dto.CreatePromotionRequest
	if err := c.Bind(&req); err != nil {
		return httputil.Error(c, httputil.BindError(err))
	}

	if fieldErrors := h.validator.ValidateAndFormat(&req); fieldErrors != nil {
		return httputil.ValidationError(c, fieldErrors)
	}

	ctx := c.Request().Context()
	companyID := getCompanyID(c)

	result, err := h.promotionSvc.Create(ctx, companyID, req)
	if err != nil {
		return httputil.Error(c, err)
	}

	return httputil.Success(c, http.StatusCreated, "Promotion created successfully", result)
}

func (h *PromotionHandler) Get(c *echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return httputil.Error(c, apperror.BadRequest("invalid promotion ID"))
	}

	ctx := c.Request().Context()
	companyID := getCompanyID(c)

	result, err := h.promotionSvc.GetByID(ctx, companyID, id)
	if err != nil {
		return httputil.Error(c, err)
	}

	return httputil.Success(c, http.StatusOK, "Promotion retrieved successfully", result)
}

func (h *PromotionHandler) List(c *echo.Context) error {
	var req dto.ListPromotionRequest

	if page := c.QueryParam("page"); page != "" {
		if p, err := strconv.Atoi(page); err == nil {
			req.Page = p
		}
	}
	if perPage := c.QueryParam("per_page"); perPage != "" {
		if pp, err := strconv.Atoi(perPage); err == nil {
			req.PerPage = pp
		}
	}
	req.Search = c.QueryParam("search")
	req.Type = c.QueryParam("type")
	req.Status = c.QueryParam("status")
	if isActive := c.QueryParam("is_active"); isActive != "" {
		if b, err := strconv.ParseBool(isActive); err == nil {
			req.IsActive = &b
		}
	}

	ctx := c.Request().Context()
	companyID := getCompanyID(c)

	result, err := h.promotionSvc.List(ctx, companyID, req)
	if err != nil {
		return httputil.Error(c, err)
	}

	return httputil.SuccessWithPagination(c, http.StatusOK, "Promotion list retrieved",
		result.Promotions,
		&httputil.Pagination{
			TotalRecords: result.Pagination.TotalRecords,
			TotalPages:   result.Pagination.TotalPages,
			CurrentPage:  result.Pagination.CurrentPage,
			PerPage:      result.Pagination.PerPage,
			NextPage:     result.Pagination.NextPage,
			PrevPage:     result.Pagination.PrevPage,
		},
		nil,
	)
}

func (h *PromotionHandler) Update(c *echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return httputil.Error(c, apperror.BadRequest("invalid promotion ID"))
	}

	var req dto.UpdatePromotionRequest
	if err := c.Bind(&req); err != nil {
		return httputil.Error(c, httputil.BindError(err))
	}

	if fieldErrors := h.validator.ValidateAndFormat(&req); fieldErrors != nil {
		return httputil.ValidationError(c, fieldErrors)
	}

	ctx := c.Request().Context()
	companyID := getCompanyID(c)

	result, err := h.promotionSvc.Update(ctx, companyID, id, req)
	if err != nil {
		return httputil.Error(c, err)
	}

	return httputil.Success(c, http.StatusOK, "Promotion updated successfully", result)
}

func (h *PromotionHandler) Delete(c *echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return httputil.Error(c, apperror.BadRequest("invalid promotion ID"))
	}

	ctx := c.Request().Context()
	companyID := getCompanyID(c)

	if err := h.promotionSvc.Delete(ctx, companyID, id); err != nil {
		return httputil.Error(c, err)
	}

	return httputil.Success(c, http.StatusOK, "Promotion deleted successfully", nil)
}

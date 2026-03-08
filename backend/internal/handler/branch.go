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

type BranchHandler struct {
	branchSvc *service.BranchService
	validator *validator.CustomValidator
}

func NewBranchHandler(branchSvc *service.BranchService, validator *validator.CustomValidator) *BranchHandler {
	return &BranchHandler{
		branchSvc: branchSvc,
		validator: validator,
	}
}

func (h *BranchHandler) Create(c *echo.Context) error {
	var req dto.CreateBranchRequest
	if err := c.Bind(&req); err != nil {
		return httputil.Error(c, apperror.BadRequest("invalid request body"))
	}

	if fieldErrors := h.validator.ValidateAndFormat(&req); fieldErrors != nil {
		return httputil.ValidationError(c, fieldErrors)
	}

	ctx := c.Request().Context()
	companyID := int64(1) // TODO: from auth context

	result, err := h.branchSvc.Create(ctx, companyID, req)
	if err != nil {
		return httputil.Error(c, err)
	}

	return httputil.Success(c, http.StatusCreated, "Branch created successfully", result)
}

func (h *BranchHandler) Get(c *echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return httputil.Error(c, apperror.BadRequest("invalid branch ID"))
	}

	ctx := c.Request().Context()
	companyID := int64(1) // TODO: from auth context

	result, err := h.branchSvc.GetByID(ctx, companyID, id)
	if err != nil {
		return httputil.Error(c, err)
	}

	return httputil.Success(c, http.StatusOK, "Branch retrieved successfully", result)
}

func (h *BranchHandler) List(c *echo.Context) error {
	var req dto.ListBranchRequest

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
	if isActive := c.QueryParam("is_active"); isActive != "" {
		if b, err := strconv.ParseBool(isActive); err == nil {
			req.IsActive = &b
		}
	}

	ctx := c.Request().Context()
	companyID := int64(1) // TODO: from auth context

	result, err := h.branchSvc.List(ctx, companyID, req)
	if err != nil {
		return httputil.Error(c, err)
	}

	return httputil.SuccessWithPagination(c, http.StatusOK, "Branch list retrieved",
		result.Branches,
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

func (h *BranchHandler) Update(c *echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return httputil.Error(c, apperror.BadRequest("invalid branch ID"))
	}

	var req dto.UpdateBranchRequest
	if err := c.Bind(&req); err != nil {
		return httputil.Error(c, apperror.BadRequest("invalid request body"))
	}

	if fieldErrors := h.validator.ValidateAndFormat(&req); fieldErrors != nil {
		return httputil.ValidationError(c, fieldErrors)
	}

	ctx := c.Request().Context()
	companyID := int64(1) // TODO: from auth context

	result, err := h.branchSvc.Update(ctx, companyID, id, req)
	if err != nil {
		return httputil.Error(c, err)
	}

	return httputil.Success(c, http.StatusOK, "Branch updated successfully", result)
}

func (h *BranchHandler) Delete(c *echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return httputil.Error(c, apperror.BadRequest("invalid branch ID"))
	}

	ctx := c.Request().Context()
	companyID := int64(1) // TODO: from auth context

	if err := h.branchSvc.Delete(ctx, companyID, id); err != nil {
		return httputil.Error(c, err)
	}

	return httputil.Success(c, http.StatusOK, "Branch deleted successfully", nil)
}

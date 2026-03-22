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

type ProductCategoryHandler struct {
	categorySvc service.ProductCategoryServiceInterface
	validator   *validator.CustomValidator
}

func NewProductCategoryHandler(categorySvc service.ProductCategoryServiceInterface, validator *validator.CustomValidator) *ProductCategoryHandler {
	return &ProductCategoryHandler{
		categorySvc: categorySvc,
		validator:   validator,
	}
}

func (h *ProductCategoryHandler) Create(c *echo.Context) error {
	var req dto.CreateProductCategoryRequest
	if err := c.Bind(&req); err != nil {
		return httputil.Error(c, apperror.BadRequest("invalid request body"))
	}

	if fieldErrors := h.validator.ValidateAndFormat(&req); fieldErrors != nil {
		return httputil.ValidationError(c, fieldErrors)
	}

	ctx := c.Request().Context()
	companyID := getCompanyID(c)

	result, err := h.categorySvc.Create(ctx, companyID, req)
	if err != nil {
		return httputil.Error(c, err)
	}

	return httputil.Success(c, http.StatusCreated, "Category created successfully", result)
}

func (h *ProductCategoryHandler) Get(c *echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return httputil.Error(c, apperror.BadRequest("invalid category ID"))
	}

	ctx := c.Request().Context()
	companyID := getCompanyID(c)

	result, err := h.categorySvc.GetByID(ctx, companyID, id)
	if err != nil {
		return httputil.Error(c, err)
	}

	return httputil.Success(c, http.StatusOK, "Category retrieved successfully", result)
}

func (h *ProductCategoryHandler) List(c *echo.Context) error {
	var req dto.ListProductCategoryRequest

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
	companyID := getCompanyID(c)

	result, err := h.categorySvc.List(ctx, companyID, req)
	if err != nil {
		return httputil.Error(c, err)
	}

	return httputil.SuccessWithPagination(c, http.StatusOK, "Category list retrieved",
		result.Categories,
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

func (h *ProductCategoryHandler) ListAll(c *echo.Context) error {
	ctx := c.Request().Context()
	companyID := getCompanyID(c)

	result, err := h.categorySvc.ListAll(ctx, companyID)
	if err != nil {
		return httputil.Error(c, err)
	}

	return httputil.Success(c, http.StatusOK, "All categories retrieved", result)
}

func (h *ProductCategoryHandler) Update(c *echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return httputil.Error(c, apperror.BadRequest("invalid category ID"))
	}

	var req dto.UpdateProductCategoryRequest
	if err := c.Bind(&req); err != nil {
		return httputil.Error(c, apperror.BadRequest("invalid request body"))
	}

	if fieldErrors := h.validator.ValidateAndFormat(&req); fieldErrors != nil {
		return httputil.ValidationError(c, fieldErrors)
	}

	ctx := c.Request().Context()
	companyID := getCompanyID(c)

	result, err := h.categorySvc.Update(ctx, companyID, id, req)
	if err != nil {
		return httputil.Error(c, err)
	}

	return httputil.Success(c, http.StatusOK, "Category updated successfully", result)
}

func (h *ProductCategoryHandler) Delete(c *echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return httputil.Error(c, apperror.BadRequest("invalid category ID"))
	}

	ctx := c.Request().Context()
	companyID := getCompanyID(c)

	if err := h.categorySvc.Delete(ctx, companyID, id); err != nil {
		return httputil.Error(c, err)
	}

	return httputil.Success(c, http.StatusOK, "Category deleted successfully", nil)
}

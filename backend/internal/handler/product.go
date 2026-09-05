package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/irvanmhndra/nexpos-api/internal/dto"
	"github.com/irvanmhndra/nexpos-api/internal/service"
	"github.com/irvanmhndra/nexpos-api/pkg/apperror"
	"github.com/irvanmhndra/nexpos-api/pkg/httputil"
	"github.com/irvanmhndra/nexpos-api/pkg/validator"
	"github.com/labstack/echo/v5"
)

type ProductHandler struct {
	productSvc service.ProductServiceInterface
	validator  *validator.CustomValidator
}

func NewProductHandler(productSvc service.ProductServiceInterface, validator *validator.CustomValidator) *ProductHandler {
	return &ProductHandler{
		productSvc: productSvc,
		validator:  validator,
	}
}

func (h *ProductHandler) Create(c *echo.Context) error {
	var req dto.CreateProductRequest
	if err := c.Bind(&req); err != nil {
		return httputil.Error(c, httputil.BindError(err))
	}

	if fieldErrors := h.validator.ValidateAndFormat(&req); fieldErrors != nil {
		return httputil.ValidationError(c, fieldErrors)
	}

	ctx := c.Request().Context()
	companyID := getCompanyID(c)

	result, err := h.productSvc.Create(ctx, companyID, req)
	if err != nil {
		return httputil.Error(c, err)
	}

	return httputil.Success(c, http.StatusCreated, "Product created successfully", result)
}

func (h *ProductHandler) Get(c *echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return httputil.Error(c, apperror.BadRequest("invalid product ID"))
	}

	ctx := c.Request().Context()
	companyID := getCompanyID(c)

	result, err := h.productSvc.GetByID(ctx, companyID, id)
	if err != nil {
		return httputil.Error(c, err)
	}

	return httputil.Success(c, http.StatusOK, "Product retrieved successfully", result)
}

func (h *ProductHandler) Lookup(c *echo.Context) error {
	code := c.QueryParam("code")
	if code == "" {
		return httputil.Error(c, apperror.BadRequest("code query parameter is required"))
	}

	ctx := c.Request().Context()
	companyID := getCompanyID(c)

	result, err := h.productSvc.LookupByCode(ctx, companyID, code)
	if err != nil {
		return httputil.Error(c, err)
	}

	return httputil.Success(c, http.StatusOK, "Product found", result)
}

func (h *ProductHandler) List(c *echo.Context) error {
	var req dto.ListProductRequest

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
	req.Search = strings.TrimSpace(c.QueryParam("search"))
	if categoryID := c.QueryParam("category_id"); categoryID != "" {
		if cid, err := strconv.ParseInt(categoryID, 10, 64); err == nil {
			req.CategoryID = &cid
		}
	}
	if isActive := c.QueryParam("is_active"); isActive != "" {
		if b, err := strconv.ParseBool(isActive); err == nil {
			req.IsActive = &b
		}
	}

	ctx := c.Request().Context()
	companyID := getCompanyID(c)

	result, err := h.productSvc.List(ctx, companyID, req)
	if err != nil {
		return httputil.Error(c, err)
	}

	return httputil.SuccessWithPagination(c, http.StatusOK, "Product list retrieved",
		result.Products,
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

func (h *ProductHandler) Update(c *echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return httputil.Error(c, apperror.BadRequest("invalid product ID"))
	}

	var req dto.UpdateProductRequest
	if err := c.Bind(&req); err != nil {
		return httputil.Error(c, httputil.BindError(err))
	}

	if fieldErrors := h.validator.ValidateAndFormat(&req); fieldErrors != nil {
		return httputil.ValidationError(c, fieldErrors)
	}

	ctx := c.Request().Context()
	companyID := getCompanyID(c)

	result, err := h.productSvc.Update(ctx, companyID, id, req)
	if err != nil {
		return httputil.Error(c, err)
	}

	return httputil.Success(c, http.StatusOK, "Product updated successfully", result)
}

func (h *ProductHandler) Delete(c *echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return httputil.Error(c, apperror.BadRequest("invalid product ID"))
	}

	ctx := c.Request().Context()
	companyID := getCompanyID(c)

	if err := h.productSvc.Delete(ctx, companyID, id); err != nil {
		return httputil.Error(c, err)
	}

	return httputil.Success(c, http.StatusOK, "Product deleted successfully", nil)
}

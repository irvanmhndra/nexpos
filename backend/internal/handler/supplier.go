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

type SupplierHandler struct {
	purchaseOrderSvc service.PurchaseOrderServiceInterface
	validator        *validator.CustomValidator
}

func NewSupplierHandler(purchaseOrderSvc service.PurchaseOrderServiceInterface, v *validator.CustomValidator) *SupplierHandler {
	return &SupplierHandler{
		purchaseOrderSvc: purchaseOrderSvc,
		validator:        v,
	}
}

func (h *SupplierHandler) Create(c *echo.Context) error {
	var req dto.CreateSupplierRequest
	if err := c.Bind(&req); err != nil {
		return httputil.Error(c, httputil.BindError(err))
	}

	if fieldErrors := h.validator.ValidateAndFormat(&req); fieldErrors != nil {
		return httputil.ValidationError(c, fieldErrors)
	}

	ctx := c.Request().Context()
	companyID := getCompanyID(c)

	result, err := h.purchaseOrderSvc.CreateSupplier(ctx, companyID, req)
	if err != nil {
		return httputil.Error(c, err)
	}

	return httputil.Success(c, http.StatusCreated, "Supplier created successfully", result)
}

func (h *SupplierHandler) Get(c *echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return httputil.Error(c, apperror.BadRequest("invalid supplier ID"))
	}

	ctx := c.Request().Context()
	companyID := getCompanyID(c)

	result, err := h.purchaseOrderSvc.GetSupplier(ctx, companyID, id)
	if err != nil {
		return httputil.Error(c, err)
	}

	return httputil.Success(c, http.StatusOK, "Supplier retrieved successfully", result)
}

func (h *SupplierHandler) List(c *echo.Context) error {
	var req dto.ListSupplierRequest

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

	result, err := h.purchaseOrderSvc.ListSuppliers(ctx, companyID, req)
	if err != nil {
		return httputil.Error(c, err)
	}

	return httputil.SuccessWithPagination(c, http.StatusOK, "Supplier list retrieved",
		result.Suppliers, result.Pagination, nil)
}

func (h *SupplierHandler) Update(c *echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return httputil.Error(c, apperror.BadRequest("invalid supplier ID"))
	}

	var req dto.UpdateSupplierRequest
	if err := c.Bind(&req); err != nil {
		return httputil.Error(c, httputil.BindError(err))
	}

	if fieldErrors := h.validator.ValidateAndFormat(&req); fieldErrors != nil {
		return httputil.ValidationError(c, fieldErrors)
	}

	ctx := c.Request().Context()
	companyID := getCompanyID(c)

	result, err := h.purchaseOrderSvc.UpdateSupplier(ctx, companyID, id, req)
	if err != nil {
		return httputil.Error(c, err)
	}

	return httputil.Success(c, http.StatusOK, "Supplier updated successfully", result)
}

func (h *SupplierHandler) Delete(c *echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return httputil.Error(c, apperror.BadRequest("invalid supplier ID"))
	}

	ctx := c.Request().Context()
	companyID := getCompanyID(c)

	if err := h.purchaseOrderSvc.DeleteSupplier(ctx, companyID, id); err != nil {
		return httputil.Error(c, err)
	}

	return httputil.Success(c, http.StatusOK, "Supplier deleted successfully", nil)
}

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

type StockOpnameHandler struct {
	svc       service.StockOpnameServiceInterface
	validator *validator.CustomValidator
}

func NewStockOpnameHandler(svc service.StockOpnameServiceInterface, v *validator.CustomValidator) *StockOpnameHandler {
	return &StockOpnameHandler{svc: svc, validator: v}
}

func (h *StockOpnameHandler) Create(c *echo.Context) error {
	var req dto.CreateStockOpnameRequest
	if err := c.Bind(&req); err != nil {
		return httputil.Error(c, httputil.BindError(err))
	}
	if fieldErrors := h.validator.ValidateAndFormat(&req); fieldErrors != nil {
		return httputil.ValidationError(c, fieldErrors)
	}

	ctx := c.Request().Context()
	result, err := h.svc.Create(ctx, getCompanyID(c), getUserID(c), req)
	if err != nil {
		return httputil.Error(c, err)
	}
	return httputil.Success(c, http.StatusCreated, "Stock opname created successfully", result)
}

func (h *StockOpnameHandler) Get(c *echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return httputil.Error(c, apperror.BadRequest("invalid stock opname ID"))
	}
	result, err := h.svc.GetByID(c.Request().Context(), getCompanyID(c), id)
	if err != nil {
		return httputil.Error(c, err)
	}
	return httputil.Success(c, http.StatusOK, "Stock opname retrieved successfully", result)
}

func (h *StockOpnameHandler) List(c *echo.Context) error {
	var req dto.ListStockOpnameRequest
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
	req.Status = c.QueryParam("status")
	if branchID := c.QueryParam("branch_id"); branchID != "" {
		if bid, err := strconv.ParseInt(branchID, 10, 64); err == nil {
			req.BranchID = &bid
		}
	}

	result, err := h.svc.List(c.Request().Context(), getCompanyID(c), req)
	if err != nil {
		return httputil.Error(c, err)
	}
	return httputil.SuccessWithPagination(c, http.StatusOK, "Stock opnames retrieved",
		result.Opnames, result.Pagination, nil)
}

func (h *StockOpnameHandler) UpdateItem(c *echo.Context) error {
	opnameID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return httputil.Error(c, apperror.BadRequest("invalid stock opname ID"))
	}
	itemID, err := strconv.ParseInt(c.Param("itemId"), 10, 64)
	if err != nil {
		return httputil.Error(c, apperror.BadRequest("invalid item ID"))
	}

	var req dto.UpdateOpnameItemRequest
	if err := c.Bind(&req); err != nil {
		return httputil.Error(c, httputil.BindError(err))
	}
	if fieldErrors := h.validator.ValidateAndFormat(&req); fieldErrors != nil {
		return httputil.ValidationError(c, fieldErrors)
	}

	result, err := h.svc.UpdateItem(c.Request().Context(), getCompanyID(c), opnameID, itemID, getUserID(c), req)
	if err != nil {
		return httputil.Error(c, err)
	}
	return httputil.Success(c, http.StatusOK, "Stock opname item updated successfully", result)
}

func (h *StockOpnameHandler) BulkUpdateItems(c *echo.Context) error {
	opnameID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return httputil.Error(c, apperror.BadRequest("invalid stock opname ID"))
	}

	var req dto.BulkUpdateOpnameItemsRequest
	if err := c.Bind(&req); err != nil {
		return httputil.Error(c, httputil.BindError(err))
	}
	if fieldErrors := h.validator.ValidateAndFormat(&req); fieldErrors != nil {
		return httputil.ValidationError(c, fieldErrors)
	}

	result, err := h.svc.BulkUpdateItems(c.Request().Context(), getCompanyID(c), opnameID, getUserID(c), req)
	if err != nil {
		return httputil.Error(c, err)
	}
	return httputil.Success(c, http.StatusOK, "Stock opname items updated successfully", result)
}

func (h *StockOpnameHandler) Complete(c *echo.Context) error {
	opnameID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return httputil.Error(c, apperror.BadRequest("invalid stock opname ID"))
	}

	var req dto.CompleteStockOpnameRequest
	if err := c.Bind(&req); err != nil {
		return httputil.Error(c, httputil.BindError(err))
	}

	result, err := h.svc.Complete(c.Request().Context(), getCompanyID(c), opnameID, getUserID(c), req)
	if err != nil {
		return httputil.Error(c, err)
	}
	return httputil.Success(c, http.StatusOK, "Stock opname completed successfully", result)
}

func (h *StockOpnameHandler) Cancel(c *echo.Context) error {
	opnameID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return httputil.Error(c, apperror.BadRequest("invalid stock opname ID"))
	}
	result, err := h.svc.Cancel(c.Request().Context(), getCompanyID(c), opnameID)
	if err != nil {
		return httputil.Error(c, err)
	}
	return httputil.Success(c, http.StatusOK, "Stock opname cancelled successfully", result)
}

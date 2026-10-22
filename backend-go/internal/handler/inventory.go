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

type InventoryHandler struct {
	inventorySvc service.InventoryServiceInterface
	validator    *validator.CustomValidator
}

func NewInventoryHandler(inventorySvc service.InventoryServiceInterface, v *validator.CustomValidator) *InventoryHandler {
	return &InventoryHandler{
		inventorySvc: inventorySvc,
		validator:    v,
	}
}

func (h *InventoryHandler) ListInventory(c *echo.Context) error {
	var req dto.ListInventoryRequest
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
	req.Category = c.QueryParam("category")
	req.Status = c.QueryParam("status")
	if branchID := c.QueryParam("branch_id"); branchID != "" {
		if b, err := strconv.ParseInt(branchID, 10, 64); err == nil {
			req.BranchID = b
		}
	}

	ctx := c.Request().Context()
	companyID := getCompanyID(c)

	result, err := h.inventorySvc.ListInventory(ctx, companyID, req)
	if err != nil {
		return httputil.Error(c, err)
	}

	return httputil.SuccessWithPagination(c, http.StatusOK, "Inventory list retrieved", result.Items, result.Pagination, nil)
}

func (h *InventoryHandler) GetInventoryStats(c *echo.Context) error {
	branchID := int64(0)
	if b := c.QueryParam("branch_id"); b != "" {
		if bid, err := strconv.ParseInt(b, 10, 64); err == nil {
			branchID = bid
		}
	}

	ctx := c.Request().Context()
	companyID := getCompanyID(c)

	result, err := h.inventorySvc.GetInventoryStats(ctx, companyID, branchID)
	if err != nil {
		return httputil.Error(c, err)
	}

	return httputil.Success(c, http.StatusOK, "Inventory stats retrieved", result)
}

func (h *InventoryHandler) AdjustStock(c *echo.Context) error {
	var req dto.AdjustStockRequest
	if err := c.Bind(&req); err != nil {
		return httputil.Error(c, httputil.BindError(err))
	}

	if fieldErrors := h.validator.ValidateAndFormat(&req); fieldErrors != nil {
		return httputil.ValidationError(c, fieldErrors)
	}

	ctx := c.Request().Context()
	companyID := getCompanyID(c)
	userID := getUserID(c)
	createdBy := &userID

	if err := h.inventorySvc.AdjustStock(ctx, companyID, createdBy, req); err != nil {
		return httputil.Error(c, err)
	}

	return httputil.Success(c, http.StatusOK, "Stock adjusted successfully", nil)
}

func (h *InventoryHandler) UpdateMinStock(c *echo.Context) error {
	variantID, err := strconv.ParseInt(c.Param("variantId"), 10, 64)
	if err != nil {
		return httputil.Error(c, apperror.BadRequest("invalid variant ID"))
	}

	branchID := int64(0)
	if b := c.QueryParam("branch_id"); b != "" {
		if bid, err := strconv.ParseInt(b, 10, 64); err == nil {
			branchID = bid
		}
	}
	if branchID == 0 {
		return httputil.Error(c, apperror.BadRequest("branch_id is required"))
	}

	var req dto.UpdateMinStockRequest
	if err := c.Bind(&req); err != nil {
		return httputil.Error(c, httputil.BindError(err))
	}

	if fieldErrors := h.validator.ValidateAndFormat(&req); fieldErrors != nil {
		return httputil.ValidationError(c, fieldErrors)
	}

	ctx := c.Request().Context()
	companyID := getCompanyID(c)

	if err := h.inventorySvc.UpdateMinStock(ctx, companyID, variantID, branchID, req); err != nil {
		return httputil.Error(c, err)
	}

	return httputil.Success(c, http.StatusOK, "Min stock updated successfully", nil)
}

func (h *InventoryHandler) ListMovements(c *echo.Context) error {
	var req dto.ListMovementsRequest
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
	req.StartDate = c.QueryParam("start_date")
	req.EndDate = c.QueryParam("end_date")
	if branchID := c.QueryParam("branch_id"); branchID != "" {
		if b, err := strconv.ParseInt(branchID, 10, 64); err == nil {
			req.BranchID = b
		}
	}

	ctx := c.Request().Context()
	companyID := getCompanyID(c)

	result, err := h.inventorySvc.ListMovements(ctx, companyID, req)
	if err != nil {
		return httputil.Error(c, err)
	}

	return httputil.SuccessWithPagination(c, http.StatusOK, "Stock movements retrieved", result.Movements, result.Pagination, nil)
}

func (h *InventoryHandler) GetMovementStats(c *echo.Context) error {
	branchID := int64(0)
	if b := c.QueryParam("branch_id"); b != "" {
		if bid, err := strconv.ParseInt(b, 10, 64); err == nil {
			branchID = bid
		}
	}

	ctx := c.Request().Context()
	companyID := getCompanyID(c)

	result, err := h.inventorySvc.GetMovementStats(ctx, companyID, branchID)
	if err != nil {
		return httputil.Error(c, err)
	}

	return httputil.Success(c, http.StatusOK, "Movement stats retrieved", result)
}

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

type DailySettlementHandler struct {
	svc       service.DailySettlementServiceInterface
	validator *validator.CustomValidator
}

func NewDailySettlementHandler(svc service.DailySettlementServiceInterface, v *validator.CustomValidator) *DailySettlementHandler {
	return &DailySettlementHandler{svc: svc, validator: v}
}

func (h *DailySettlementHandler) Report(c *echo.Context) error {
	var req dto.DailySettlementReportRequest
	if branchID := c.QueryParam("branch_id"); branchID != "" {
		if bid, err := strconv.ParseInt(branchID, 10, 64); err == nil {
			req.BranchID = bid
		}
	}
	req.Date = c.QueryParam("date")

	result, err := h.svc.Report(c.Request().Context(), getCompanyID(c), req)
	if err != nil {
		return httputil.Error(c, err)
	}
	return httputil.Success(c, http.StatusOK, "Daily settlement report retrieved", result)
}

func (h *DailySettlementHandler) Create(c *echo.Context) error {
	var req dto.CreateDailySettlementRequest
	if err := c.Bind(&req); err != nil {
		return httputil.Error(c, httputil.BindError(err))
	}
	if fieldErrors := h.validator.ValidateAndFormat(&req); fieldErrors != nil {
		return httputil.ValidationError(c, fieldErrors)
	}

	result, err := h.svc.Create(c.Request().Context(), getCompanyID(c), getUserID(c), req)
	if err != nil {
		return httputil.Error(c, err)
	}
	return httputil.Success(c, http.StatusCreated, "Daily settlement created successfully", result)
}

func (h *DailySettlementHandler) Get(c *echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return httputil.Error(c, apperror.BadRequest("invalid settlement ID"))
	}
	result, err := h.svc.GetByID(c.Request().Context(), getCompanyID(c), id)
	if err != nil {
		return httputil.Error(c, err)
	}
	return httputil.Success(c, http.StatusOK, "Daily settlement retrieved successfully", result)
}

func (h *DailySettlementHandler) List(c *echo.Context) error {
	var req dto.ListDailySettlementRequest
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
	req.DateFrom = c.QueryParam("date_from")
	req.DateTo = c.QueryParam("date_to")
	if branchID := c.QueryParam("branch_id"); branchID != "" {
		if bid, err := strconv.ParseInt(branchID, 10, 64); err == nil {
			req.BranchID = &bid
		}
	}

	result, err := h.svc.List(c.Request().Context(), getCompanyID(c), req)
	if err != nil {
		return httputil.Error(c, err)
	}
	return httputil.SuccessWithPagination(c, http.StatusOK, "Daily settlements retrieved",
		result.Settlements, result.Pagination, nil)
}

func (h *DailySettlementHandler) UpdateItem(c *echo.Context) error {
	settlementID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return httputil.Error(c, apperror.BadRequest("invalid settlement ID"))
	}
	itemID, err := strconv.ParseInt(c.Param("itemId"), 10, 64)
	if err != nil {
		return httputil.Error(c, apperror.BadRequest("invalid item ID"))
	}

	var req dto.UpdateSettlementItemRequest
	if err := c.Bind(&req); err != nil {
		return httputil.Error(c, httputil.BindError(err))
	}
	if fieldErrors := h.validator.ValidateAndFormat(&req); fieldErrors != nil {
		return httputil.ValidationError(c, fieldErrors)
	}

	result, err := h.svc.UpdateItem(c.Request().Context(), getCompanyID(c), settlementID, itemID, req)
	if err != nil {
		return httputil.Error(c, err)
	}
	return httputil.Success(c, http.StatusOK, "Settlement item updated successfully", result)
}

func (h *DailySettlementHandler) BulkUpdateItems(c *echo.Context) error {
	settlementID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return httputil.Error(c, apperror.BadRequest("invalid settlement ID"))
	}

	var req dto.BulkUpdateSettlementItemsRequest
	if err := c.Bind(&req); err != nil {
		return httputil.Error(c, httputil.BindError(err))
	}
	if fieldErrors := h.validator.ValidateAndFormat(&req); fieldErrors != nil {
		return httputil.ValidationError(c, fieldErrors)
	}

	result, err := h.svc.BulkUpdateItems(c.Request().Context(), getCompanyID(c), settlementID, req)
	if err != nil {
		return httputil.Error(c, err)
	}
	return httputil.Success(c, http.StatusOK, "Settlement items updated successfully", result)
}

func (h *DailySettlementHandler) Finalize(c *echo.Context) error {
	settlementID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return httputil.Error(c, apperror.BadRequest("invalid settlement ID"))
	}

	var req dto.FinalizeDailySettlementRequest
	if err := c.Bind(&req); err != nil {
		return httputil.Error(c, httputil.BindError(err))
	}

	result, err := h.svc.Finalize(c.Request().Context(), getCompanyID(c), settlementID, getUserID(c), req)
	if err != nil {
		return httputil.Error(c, err)
	}
	return httputil.Success(c, http.StatusOK, "Daily settlement finalized successfully", result)
}

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

type ShiftHandler struct {
	shiftSvc  service.ShiftServiceInterface
	validator *validator.CustomValidator
}

func NewShiftHandler(shiftSvc service.ShiftServiceInterface, v *validator.CustomValidator) *ShiftHandler {
	return &ShiftHandler{
		shiftSvc:  shiftSvc,
		validator: v,
	}
}

func (h *ShiftHandler) Open(c *echo.Context) error {
	var req dto.OpenShiftRequest
	if err := c.Bind(&req); err != nil {
		return httputil.Error(c, httputil.BindError(err))
	}

	if fieldErrors := h.validator.ValidateAndFormat(&req); fieldErrors != nil {
		return httputil.ValidationError(c, fieldErrors)
	}

	ctx := c.Request().Context()
	companyID := getCompanyID(c)
	cashierID := getUserID(c)
	branchID := req.BranchID

	result, err := h.shiftSvc.OpenShift(ctx, companyID, branchID, cashierID, req)
	if err != nil {
		return httputil.Error(c, err)
	}

	return httputil.Success(c, http.StatusCreated, "Shift opened successfully", result)
}

func (h *ShiftHandler) Close(c *echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return httputil.Error(c, apperror.BadRequest("invalid shift ID"))
	}

	var req dto.CloseShiftRequest
	if err := c.Bind(&req); err != nil {
		return httputil.Error(c, httputil.BindError(err))
	}

	if fieldErrors := h.validator.ValidateAndFormat(&req); fieldErrors != nil {
		return httputil.ValidationError(c, fieldErrors)
	}

	ctx := c.Request().Context()
	companyID := getCompanyID(c)

	result, err := h.shiftSvc.CloseShift(ctx, companyID, id, req)
	if err != nil {
		return httputil.Error(c, err)
	}

	return httputil.Success(c, http.StatusOK, "Shift closed successfully", result)
}

func (h *ShiftHandler) GetCurrent(c *echo.Context) error {
	ctx := c.Request().Context()
	companyID := getCompanyID(c)
	cashierID := getUserID(c)
	branchID := getBranchID(c)

	result, err := h.shiftSvc.GetCurrentShift(ctx, companyID, branchID, cashierID)
	if err != nil {
		return httputil.Error(c, err)
	}

	return httputil.Success(c, http.StatusOK, "Current shift retrieved successfully", result)
}

func (h *ShiftHandler) Get(c *echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return httputil.Error(c, apperror.BadRequest("invalid shift ID"))
	}

	ctx := c.Request().Context()
	companyID := getCompanyID(c)

	result, err := h.shiftSvc.GetShift(ctx, companyID, id)
	if err != nil {
		return httputil.Error(c, err)
	}

	return httputil.Success(c, http.StatusOK, "Shift retrieved successfully", result)
}

func (h *ShiftHandler) List(c *echo.Context) error {
	var req dto.ListShiftRequest

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
	if branchID := c.QueryParam("branch_id"); branchID != "" {
		if b, err := strconv.ParseInt(branchID, 10, 64); err == nil {
			req.BranchID = &b
		}
	}
	if cashierID := c.QueryParam("cashier_id"); cashierID != "" {
		if cid, err := strconv.ParseInt(cashierID, 10, 64); err == nil {
			req.CashierID = &cid
		}
	}
	req.Status = c.QueryParam("status")
	req.DateFrom = c.QueryParam("date_from")
	req.DateTo = c.QueryParam("date_to")

	ctx := c.Request().Context()
	companyID := getCompanyID(c)

	result, err := h.shiftSvc.ListShifts(ctx, companyID, req)
	if err != nil {
		return httputil.Error(c, err)
	}

	return httputil.SuccessWithPagination(c, http.StatusOK, "Shifts retrieved",
		result.Shifts, result.Pagination, nil)
}

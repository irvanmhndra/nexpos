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

type ReportHandler struct {
	reportSvc service.ReportServiceInterface
	validator *validator.CustomValidator
}

func NewReportHandler(reportSvc service.ReportServiceInterface, validator *validator.CustomValidator) *ReportHandler {
	return &ReportHandler{
		reportSvc: reportSvc,
		validator: validator,
	}
}

func (h *ReportHandler) GetSummary(c *echo.Context) error {
	var req dto.ReportRequest
	if err := (*c).Bind(&req); err != nil {
		return httputil.Error(c, apperror.BadRequest("invalid query parameters"))
	}

	if req.DateFrom == "" || req.DateTo == "" {
		return httputil.Error(c, apperror.BadRequest("date_from and date_to are required"))
	}

	ctx := (*c).Request().Context()
	companyID := getCompanyID(c)

	result, err := h.reportSvc.GetSummary(ctx, companyID, req.DateFrom, req.DateTo)
	if err != nil {
		return httputil.Error(c, err)
	}

	return httputil.Success(c, http.StatusOK, "Report summary retrieved successfully", result)
}

func (h *ReportHandler) GetSalesTrend(c *echo.Context) error {
	var req dto.ReportRequest
	if err := (*c).Bind(&req); err != nil {
		return httputil.Error(c, apperror.BadRequest("invalid query parameters"))
	}

	if req.DateFrom == "" || req.DateTo == "" {
		return httputil.Error(c, apperror.BadRequest("date_from and date_to are required"))
	}

	ctx := (*c).Request().Context()
	companyID := getCompanyID(c)

	result, err := h.reportSvc.GetSalesTrend(ctx, companyID, req.DateFrom, req.DateTo)
	if err != nil {
		return httputil.Error(c, err)
	}

	return httputil.Success(c, http.StatusOK, "Sales trend retrieved successfully", result)
}

func (h *ReportHandler) GetTopProducts(c *echo.Context) error {
	var req dto.ReportRequest
	if err := (*c).Bind(&req); err != nil {
		return httputil.Error(c, apperror.BadRequest("invalid query parameters"))
	}

	if req.DateFrom == "" || req.DateTo == "" {
		return httputil.Error(c, apperror.BadRequest("date_from and date_to are required"))
	}

	limitStr := (*c).QueryParam("limit")
	limit := 10
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	ctx := (*c).Request().Context()
	companyID := getCompanyID(c)

	result, err := h.reportSvc.GetTopProducts(ctx, companyID, req.DateFrom, req.DateTo, limit)
	if err != nil {
		return httputil.Error(c, err)
	}

	return httputil.Success(c, http.StatusOK, "Top products retrieved successfully", result)
}

func (h *ReportHandler) GetCategoryRevenue(c *echo.Context) error {
	var req dto.ReportRequest
	if err := (*c).Bind(&req); err != nil {
		return httputil.Error(c, apperror.BadRequest("invalid query parameters"))
	}

	if req.DateFrom == "" || req.DateTo == "" {
		return httputil.Error(c, apperror.BadRequest("date_from and date_to are required"))
	}

	ctx := (*c).Request().Context()
	companyID := getCompanyID(c)

	result, err := h.reportSvc.GetCategoryRevenue(ctx, companyID, req.DateFrom, req.DateTo)
	if err != nil {
		return httputil.Error(c, err)
	}

	return httputil.Success(c, http.StatusOK, "Category revenue retrieved successfully", result)
}

func (h *ReportHandler) GetPaymentMethods(c *echo.Context) error {
	var req dto.ReportRequest
	if err := (*c).Bind(&req); err != nil {
		return httputil.Error(c, apperror.BadRequest("invalid query parameters"))
	}

	if req.DateFrom == "" || req.DateTo == "" {
		return httputil.Error(c, apperror.BadRequest("date_from and date_to are required"))
	}

	ctx := (*c).Request().Context()
	companyID := getCompanyID(c)

	result, err := h.reportSvc.GetPaymentMethods(ctx, companyID, req.DateFrom, req.DateTo)
	if err != nil {
		return httputil.Error(c, err)
	}

	return httputil.Success(c, http.StatusOK, "Payment methods retrieved successfully", result)
}

func (h *ReportHandler) GetHourlySales(c *echo.Context) error {
	var req dto.ReportRequest
	if err := (*c).Bind(&req); err != nil {
		return httputil.Error(c, apperror.BadRequest("invalid query parameters"))
	}

	if req.DateFrom == "" || req.DateTo == "" {
		return httputil.Error(c, apperror.BadRequest("date_from and date_to are required"))
	}

	ctx := (*c).Request().Context()
	companyID := getCompanyID(c)

	result, err := h.reportSvc.GetHourlySales(ctx, companyID, req.DateFrom, req.DateTo)
	if err != nil {
		return httputil.Error(c, err)
	}

	return httputil.Success(c, http.StatusOK, "Hourly sales retrieved successfully", result)
}

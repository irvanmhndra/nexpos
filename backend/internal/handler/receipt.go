package handler

import (
	"net/http"
	"strconv"

	"github.com/irvanmhndra/nexpos-api/internal/service"
	"github.com/irvanmhndra/nexpos-api/pkg/apperror"
	"github.com/irvanmhndra/nexpos-api/pkg/httputil"
	"github.com/labstack/echo/v5"
)

type ReceiptHandler struct {
	receiptSvc service.ReceiptServiceInterface
}

func NewReceiptHandler(receiptSvc service.ReceiptServiceInterface) *ReceiptHandler {
	return &ReceiptHandler{receiptSvc: receiptSvc}
}

func (h *ReceiptHandler) Get(c *echo.Context) error {
	id, err := strconv.ParseInt((*c).Param("id"), 10, 64)
	if err != nil {
		return httputil.Error(c, apperror.BadRequest("invalid order ID"))
	}

	companyID := getCompanyID(c)
	ctx := (*c).Request().Context()

	rec, err := h.receiptSvc.GetByOrderID(ctx, companyID, id)
	if err != nil {
		return httputil.Error(c, err)
	}

	return httputil.Success(c, http.StatusOK, "Receipt retrieved successfully", rec)
}

package handler

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/irvanmhndra/nexpos-api/internal/dto"
	"github.com/irvanmhndra/nexpos-api/internal/service"
	"github.com/irvanmhndra/nexpos-api/pkg/apperror"
	"github.com/irvanmhndra/nexpos-api/pkg/httputil"
	"github.com/irvanmhndra/nexpos-api/pkg/validator"
	"github.com/labstack/echo/v5"
)

type OrderHandler struct {
	orderSvc   service.OrderServiceInterface
	receiptSvc service.ReceiptServiceInterface // nil if MongoDB not configured
	validator  *validator.CustomValidator
}

func NewOrderHandler(orderSvc service.OrderServiceInterface, validator *validator.CustomValidator, receiptSvc ...service.ReceiptServiceInterface) *OrderHandler {
	h := &OrderHandler{
		orderSvc:  orderSvc,
		validator: validator,
	}
	if len(receiptSvc) > 0 {
		h.receiptSvc = receiptSvc[0]
	}
	return h
}

func (h *OrderHandler) Preview(c *echo.Context) error {
	var req dto.PreviewOrderRequest
	if err := (*c).Bind(&req); err != nil {
		return httputil.Error(c, httputil.BindError(err))
	}

	if fieldErrors := h.validator.ValidateAndFormat(&req); fieldErrors != nil {
		return httputil.ValidationError(c, fieldErrors)
	}

	ctx := (*c).Request().Context()
	companyID := getCompanyID(c)
	branchID := getBranchID(c)

	result, err := h.orderSvc.Preview(ctx, companyID, branchID, req)
	if err != nil {
		return httputil.Error(c, err)
	}

	return httputil.Success(c, http.StatusOK, "Order preview calculated", result)
}

func (h *OrderHandler) Create(c *echo.Context) error {
	var req dto.CreateOrderRequest
	if err := (*c).Bind(&req); err != nil {
		return httputil.Error(c, httputil.BindError(err))
	}

	if fieldErrors := h.validator.ValidateAndFormat(&req); fieldErrors != nil {
		return httputil.ValidationError(c, fieldErrors)
	}

	ctx := (*c).Request().Context()
	companyID := getCompanyID(c)
	branchID := getBranchID(c)
	cashierID := getUserID(c)

	result, err := h.orderSvc.Create(ctx, companyID, branchID, cashierID, req)
	if err != nil {
		return httputil.Error(c, err)
	}

	return httputil.Success(c, http.StatusCreated, "Order created successfully", result)
}

func (h *OrderHandler) Get(c *echo.Context) error {
	id, err := strconv.ParseInt((*c).Param("id"), 10, 64)
	if err != nil {
		return httputil.Error(c, apperror.BadRequest("invalid order ID"))
	}

	ctx := (*c).Request().Context()
	companyID := getCompanyID(c)

	result, err := h.orderSvc.GetByID(ctx, companyID, id)
	if err != nil {
		return httputil.Error(c, err)
	}

	return httputil.Success(c, http.StatusOK, "Order retrieved successfully", result)
}

func (h *OrderHandler) List(c *echo.Context) error {
	var req dto.ListOrderRequest
	if err := (*c).Bind(&req); err != nil {
		return httputil.Error(c, apperror.BadRequest("invalid query parameters"))
	}

	ctx := (*c).Request().Context()
	companyID := getCompanyID(c)

	result, err := h.orderSvc.List(ctx, companyID, req)
	if err != nil {
		return httputil.Error(c, err)
	}

	pagination := &httputil.Pagination{
		TotalRecords: result.Pagination.TotalRecords,
		TotalPages:   result.Pagination.TotalPages,
		CurrentPage:  result.Pagination.CurrentPage,
		PerPage:      result.Pagination.PerPage,
		Count:        len(result.Orders),
		NextPage:     result.Pagination.NextPage,
		PrevPage:     result.Pagination.PrevPage,
	}

	return httputil.SuccessWithPagination(c, http.StatusOK, "Orders retrieved successfully", result.Orders, pagination, nil)
}

func (h *OrderHandler) Update(c *echo.Context) error {
	id, err := strconv.ParseInt((*c).Param("id"), 10, 64)
	if err != nil {
		return httputil.Error(c, apperror.BadRequest("invalid order ID"))
	}

	var req dto.UpdateOrderRequest
	if err := (*c).Bind(&req); err != nil {
		return httputil.Error(c, httputil.BindError(err))
	}

	if fieldErrors := h.validator.ValidateAndFormat(&req); fieldErrors != nil {
		return httputil.ValidationError(c, fieldErrors)
	}

	ctx := (*c).Request().Context()
	companyID := getCompanyID(c)

	result, err := h.orderSvc.UpdateOrder(ctx, companyID, id, req)
	if err != nil {
		return httputil.Error(c, err)
	}

	return httputil.Success(c, http.StatusOK, "Order updated successfully", result)
}

func (h *OrderHandler) Confirm(c *echo.Context) error {
	id, err := strconv.ParseInt((*c).Param("id"), 10, 64)
	if err != nil {
		return httputil.Error(c, apperror.BadRequest("invalid order ID"))
	}

	ctx := (*c).Request().Context()
	companyID := getCompanyID(c)

	result, err := h.orderSvc.ConfirmOrder(ctx, companyID, id)
	if err != nil {
		return httputil.Error(c, err)
	}

	return httputil.Success(c, http.StatusOK, "Order confirmed successfully", result)
}

func (h *OrderHandler) AddPayment(c *echo.Context) error {
	id, err := strconv.ParseInt((*c).Param("id"), 10, 64)
	if err != nil {
		return httputil.Error(c, apperror.BadRequest("invalid order ID"))
	}

	var req dto.AddPaymentRequest
	if err := (*c).Bind(&req); err != nil {
		return httputil.Error(c, httputil.BindError(err))
	}

	if fieldErrors := h.validator.ValidateAndFormat(&req); fieldErrors != nil {
		return httputil.ValidationError(c, fieldErrors)
	}

	ctx := (*c).Request().Context()
	companyID := getCompanyID(c)

	result, err := h.orderSvc.AddPayment(ctx, companyID, id, req)
	if err != nil {
		return httputil.Error(c, err)
	}

	return httputil.Success(c, http.StatusOK, "Payment added successfully", result)
}

func (h *OrderHandler) Complete(c *echo.Context) error {
	id, err := strconv.ParseInt((*c).Param("id"), 10, 64)
	if err != nil {
		return httputil.Error(c, apperror.BadRequest("invalid order ID"))
	}

	var req dto.CompleteOrderRequest
	if err := (*c).Bind(&req); err != nil {
		return httputil.Error(c, httputil.BindError(err))
	}

	ctx := (*c).Request().Context()
	companyID := getCompanyID(c)

	result, err := h.orderSvc.CompleteOrder(ctx, companyID, id, req)
	if err != nil {
		return httputil.Error(c, err)
	}

	// Best-effort: persist receipt snapshot to MongoDB (non-blocking failure)
	if h.receiptSvc != nil {
		if err := h.receiptSvc.CreateFromOrder(ctx, companyID, result); err != nil {
			slog.Error("failed to save receipt snapshot", "order_id", id, "error", err)
		}
	}

	return httputil.Success(c, http.StatusOK, "Order completed successfully", result)
}

func (h *OrderHandler) Cancel(c *echo.Context) error {
	id, err := strconv.ParseInt((*c).Param("id"), 10, 64)
	if err != nil {
		return httputil.Error(c, apperror.BadRequest("invalid order ID"))
	}

	var req dto.CancelOrderRequest
	if err := (*c).Bind(&req); err != nil {
		return httputil.Error(c, httputil.BindError(err))
	}

	if fieldErrors := h.validator.ValidateAndFormat(&req); fieldErrors != nil {
		return httputil.ValidationError(c, fieldErrors)
	}

	ctx := (*c).Request().Context()
	companyID := getCompanyID(c)

	result, err := h.orderSvc.CancelOrder(ctx, companyID, id, req)
	if err != nil {
		return httputil.Error(c, err)
	}

	return httputil.Success(c, http.StatusOK, "Order cancelled successfully", result)
}

func (h *OrderHandler) Void(c *echo.Context) error {
	id, err := strconv.ParseInt((*c).Param("id"), 10, 64)
	if err != nil {
		return httputil.Error(c, apperror.BadRequest("invalid order ID"))
	}

	var req dto.VoidOrderRequest
	if err := (*c).Bind(&req); err != nil {
		return httputil.Error(c, httputil.BindError(err))
	}

	if fieldErrors := h.validator.ValidateAndFormat(&req); fieldErrors != nil {
		return httputil.ValidationError(c, fieldErrors)
	}

	ctx := (*c).Request().Context()
	companyID := getCompanyID(c)

	result, err := h.orderSvc.VoidOrder(ctx, companyID, id, req)
	if err != nil {
		return httputil.Error(c, err)
	}

	return httputil.Success(c, http.StatusOK, "Order voided successfully", result)
}

func (h *OrderHandler) RefundPayment(c *echo.Context) error {
	id, err := strconv.ParseInt((*c).Param("id"), 10, 64)
	if err != nil {
		return httputil.Error(c, apperror.BadRequest("invalid order ID"))
	}

	var req dto.RefundPaymentRequest
	if err := (*c).Bind(&req); err != nil {
		return httputil.Error(c, httputil.BindError(err))
	}

	if fieldErrors := h.validator.ValidateAndFormat(&req); fieldErrors != nil {
		return httputil.ValidationError(c, fieldErrors)
	}

	ctx := (*c).Request().Context()
	companyID := getCompanyID(c)

	result, err := h.orderSvc.RefundPayment(ctx, companyID, id, req)
	if err != nil {
		return httputil.Error(c, err)
	}

	return httputil.Success(c, http.StatusOK, "Payment refunded successfully", result)
}

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

type PurchaseOrderHandler struct {
	purchaseOrderSvc service.PurchaseOrderServiceInterface
	validator        *validator.CustomValidator
}

func NewPurchaseOrderHandler(purchaseOrderSvc service.PurchaseOrderServiceInterface, v *validator.CustomValidator) *PurchaseOrderHandler {
	return &PurchaseOrderHandler{
		purchaseOrderSvc: purchaseOrderSvc,
		validator:        v,
	}
}

func (h *PurchaseOrderHandler) Create(c *echo.Context) error {
	var req dto.CreatePurchaseOrderRequest
	if err := c.Bind(&req); err != nil {
		return httputil.Error(c, apperror.BadRequest("invalid request body"))
	}

	if fieldErrors := h.validator.ValidateAndFormat(&req); fieldErrors != nil {
		return httputil.ValidationError(c, fieldErrors)
	}

	ctx := c.Request().Context()
	companyID := getCompanyID(c)
	branchID := getBranchID(c)
	userID := getUserID(c)

	result, err := h.purchaseOrderSvc.CreatePO(ctx, companyID, branchID, userID, req)
	if err != nil {
		return httputil.Error(c, err)
	}

	return httputil.Success(c, http.StatusCreated, "Purchase order created successfully", result)
}

func (h *PurchaseOrderHandler) Get(c *echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return httputil.Error(c, apperror.BadRequest("invalid purchase order ID"))
	}

	ctx := c.Request().Context()
	companyID := getCompanyID(c)

	result, err := h.purchaseOrderSvc.GetPO(ctx, companyID, id)
	if err != nil {
		return httputil.Error(c, err)
	}

	return httputil.Success(c, http.StatusOK, "Purchase order retrieved successfully", result)
}

func (h *PurchaseOrderHandler) List(c *echo.Context) error {
	var req dto.ListPurchaseOrderRequest

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
	if supplierID := c.QueryParam("supplier_id"); supplierID != "" {
		if sid, err := strconv.ParseInt(supplierID, 10, 64); err == nil {
			req.SupplierID = &sid
		}
	}

	ctx := c.Request().Context()
	companyID := getCompanyID(c)

	result, err := h.purchaseOrderSvc.ListPOs(ctx, companyID, req)
	if err != nil {
		return httputil.Error(c, err)
	}

	return httputil.SuccessWithPagination(c, http.StatusOK, "Purchase orders retrieved",
		result.PurchaseOrders, result.Pagination, nil)
}

func (h *PurchaseOrderHandler) Receive(c *echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return httputil.Error(c, apperror.BadRequest("invalid purchase order ID"))
	}

	var req dto.ReceivePurchaseOrderRequest
	if err := c.Bind(&req); err != nil {
		return httputil.Error(c, apperror.BadRequest("invalid request body"))
	}

	if fieldErrors := h.validator.ValidateAndFormat(&req); fieldErrors != nil {
		return httputil.ValidationError(c, fieldErrors)
	}

	ctx := c.Request().Context()
	companyID := getCompanyID(c)

	result, err := h.purchaseOrderSvc.ReceivePO(ctx, companyID, id, req)
	if err != nil {
		return httputil.Error(c, err)
	}

	return httputil.Success(c, http.StatusOK, "Purchase order received successfully", result)
}

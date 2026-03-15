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

type CustomerHandler struct {
	customerSvc *service.CustomerService
	validator   *validator.CustomValidator
}

func NewCustomerHandler(customerSvc *service.CustomerService, validator *validator.CustomValidator) *CustomerHandler {
	return &CustomerHandler{
		customerSvc: customerSvc,
		validator:   validator,
	}
}

func (h *CustomerHandler) Create(c *echo.Context) error {
	var req dto.CreateCustomerRequest
	if err := c.Bind(&req); err != nil {
		return httputil.Error(c, apperror.BadRequest("invalid request body"))
	}

	if fieldErrors := h.validator.ValidateAndFormat(&req); fieldErrors != nil {
		return httputil.ValidationError(c, fieldErrors)
	}

	ctx := c.Request().Context()
	companyID := getCompanyID(c)

	result, err := h.customerSvc.Create(ctx, companyID, req)
	if err != nil {
		return httputil.Error(c, err)
	}

	return httputil.Success(c, http.StatusCreated, "Customer created successfully", result)
}

func (h *CustomerHandler) Get(c *echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return httputil.Error(c, apperror.BadRequest("invalid customer ID"))
	}

	ctx := c.Request().Context()
	companyID := getCompanyID(c)

	result, err := h.customerSvc.GetByID(ctx, companyID, id)
	if err != nil {
		return httputil.Error(c, err)
	}

	return httputil.Success(c, http.StatusOK, "Customer retrieved successfully", result)
}

func (h *CustomerHandler) List(c *echo.Context) error {
	var req dto.ListCustomerRequest

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
	if isMember := c.QueryParam("is_member"); isMember != "" {
		if b, err := strconv.ParseBool(isMember); err == nil {
			req.IsMember = &b
		}
	}

	ctx := c.Request().Context()
	companyID := getCompanyID(c)

	result, err := h.customerSvc.List(ctx, companyID, req)
	if err != nil {
		return httputil.Error(c, err)
	}

	return httputil.SuccessWithPagination(c, http.StatusOK, "Customer list retrieved",
		result.Customers,
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

func (h *CustomerHandler) Update(c *echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return httputil.Error(c, apperror.BadRequest("invalid customer ID"))
	}

	var req dto.UpdateCustomerRequest
	if err := c.Bind(&req); err != nil {
		return httputil.Error(c, apperror.BadRequest("invalid request body"))
	}

	if fieldErrors := h.validator.ValidateAndFormat(&req); fieldErrors != nil {
		return httputil.ValidationError(c, fieldErrors)
	}

	ctx := c.Request().Context()
	companyID := getCompanyID(c)

	result, err := h.customerSvc.Update(ctx, companyID, id, req)
	if err != nil {
		return httputil.Error(c, err)
	}

	return httputil.Success(c, http.StatusOK, "Customer updated successfully", result)
}

func (h *CustomerHandler) Delete(c *echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return httputil.Error(c, apperror.BadRequest("invalid customer ID"))
	}

	ctx := c.Request().Context()
	companyID := getCompanyID(c)

	if err := h.customerSvc.Delete(ctx, companyID, id); err != nil {
		return httputil.Error(c, err)
	}

	return httputil.Success(c, http.StatusOK, "Customer deleted successfully", nil)
}

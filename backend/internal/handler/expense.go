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

type ExpenseCategoryHandler struct {
	expenseSvc service.ExpenseServiceInterface
	validator  *validator.CustomValidator
}

func NewExpenseCategoryHandler(expenseSvc service.ExpenseServiceInterface, v *validator.CustomValidator) *ExpenseCategoryHandler {
	return &ExpenseCategoryHandler{
		expenseSvc: expenseSvc,
		validator:  v,
	}
}

func (h *ExpenseCategoryHandler) Create(c *echo.Context) error {
	var req dto.CreateExpenseCategoryRequest
	if err := c.Bind(&req); err != nil {
		return httputil.Error(c, httputil.BindError(err))
	}

	if fieldErrors := h.validator.ValidateAndFormat(&req); fieldErrors != nil {
		return httputil.ValidationError(c, fieldErrors)
	}

	ctx := c.Request().Context()
	companyID := getCompanyID(c)

	result, err := h.expenseSvc.CreateCategory(ctx, companyID, req)
	if err != nil {
		return httputil.Error(c, err)
	}

	return httputil.Success(c, http.StatusCreated, "Expense category created successfully", result)
}

func (h *ExpenseCategoryHandler) List(c *echo.Context) error {
	var req dto.ListExpenseCategoryRequest

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

	ctx := c.Request().Context()
	companyID := getCompanyID(c)

	result, err := h.expenseSvc.ListCategories(ctx, companyID, req)
	if err != nil {
		return httputil.Error(c, err)
	}

	return httputil.SuccessWithPagination(c, http.StatusOK, "Expense categories retrieved",
		result.Categories, result.Pagination, nil)
}

func (h *ExpenseCategoryHandler) Update(c *echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return httputil.Error(c, apperror.BadRequest("invalid expense category ID"))
	}

	var req dto.UpdateExpenseCategoryRequest
	if err := c.Bind(&req); err != nil {
		return httputil.Error(c, httputil.BindError(err))
	}

	if fieldErrors := h.validator.ValidateAndFormat(&req); fieldErrors != nil {
		return httputil.ValidationError(c, fieldErrors)
	}

	ctx := c.Request().Context()
	companyID := getCompanyID(c)

	result, err := h.expenseSvc.UpdateCategory(ctx, companyID, id, req)
	if err != nil {
		return httputil.Error(c, err)
	}

	return httputil.Success(c, http.StatusOK, "Expense category updated successfully", result)
}

func (h *ExpenseCategoryHandler) Delete(c *echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return httputil.Error(c, apperror.BadRequest("invalid expense category ID"))
	}

	ctx := c.Request().Context()
	companyID := getCompanyID(c)

	if err := h.expenseSvc.DeleteCategory(ctx, companyID, id); err != nil {
		return httputil.Error(c, err)
	}

	return httputil.Success(c, http.StatusOK, "Expense category deleted successfully", nil)
}

// ============== ExpenseHandler ==============

type ExpenseHandler struct {
	expenseSvc service.ExpenseServiceInterface
	validator  *validator.CustomValidator
}

func NewExpenseHandler(expenseSvc service.ExpenseServiceInterface, v *validator.CustomValidator) *ExpenseHandler {
	return &ExpenseHandler{
		expenseSvc: expenseSvc,
		validator:  v,
	}
}

func (h *ExpenseHandler) Create(c *echo.Context) error {
	var req dto.CreateExpenseRequest
	if err := c.Bind(&req); err != nil {
		return httputil.Error(c, httputil.BindError(err))
	}

	if fieldErrors := h.validator.ValidateAndFormat(&req); fieldErrors != nil {
		return httputil.ValidationError(c, fieldErrors)
	}

	ctx := c.Request().Context()
	companyID := getCompanyID(c)
	userID := getUserID(c)
	recordedBy := &userID

	result, err := h.expenseSvc.CreateExpense(ctx, companyID, recordedBy, req)
	if err != nil {
		return httputil.Error(c, err)
	}

	return httputil.Success(c, http.StatusCreated, "Expense created successfully", result)
}

func (h *ExpenseHandler) Get(c *echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return httputil.Error(c, apperror.BadRequest("invalid expense ID"))
	}

	ctx := c.Request().Context()
	companyID := getCompanyID(c)

	result, err := h.expenseSvc.GetExpense(ctx, companyID, id)
	if err != nil {
		return httputil.Error(c, err)
	}

	return httputil.Success(c, http.StatusOK, "Expense retrieved successfully", result)
}

func (h *ExpenseHandler) List(c *echo.Context) error {
	var req dto.ListExpenseRequest

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
	if categoryID := c.QueryParam("category_id"); categoryID != "" {
		if cid, err := strconv.ParseInt(categoryID, 10, 64); err == nil {
			req.CategoryID = &cid
		}
	}
	req.DateFrom = c.QueryParam("date_from")
	req.DateTo = c.QueryParam("date_to")

	ctx := c.Request().Context()
	companyID := getCompanyID(c)

	result, err := h.expenseSvc.ListExpenses(ctx, companyID, req)
	if err != nil {
		return httputil.Error(c, err)
	}

	return httputil.SuccessWithPagination(c, http.StatusOK, "Expenses retrieved",
		result.Expenses, result.Pagination, nil)
}

func (h *ExpenseHandler) Update(c *echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return httputil.Error(c, apperror.BadRequest("invalid expense ID"))
	}

	var req dto.UpdateExpenseRequest
	if err := c.Bind(&req); err != nil {
		return httputil.Error(c, httputil.BindError(err))
	}

	if fieldErrors := h.validator.ValidateAndFormat(&req); fieldErrors != nil {
		return httputil.ValidationError(c, fieldErrors)
	}

	ctx := c.Request().Context()
	companyID := getCompanyID(c)

	result, err := h.expenseSvc.UpdateExpense(ctx, companyID, id, req)
	if err != nil {
		return httputil.Error(c, err)
	}

	return httputil.Success(c, http.StatusOK, "Expense updated successfully", result)
}

func (h *ExpenseHandler) Delete(c *echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return httputil.Error(c, apperror.BadRequest("invalid expense ID"))
	}

	ctx := c.Request().Context()
	companyID := getCompanyID(c)

	if err := h.expenseSvc.DeleteExpense(ctx, companyID, id); err != nil {
		return httputil.Error(c, err)
	}

	return httputil.Success(c, http.StatusOK, "Expense deleted successfully", nil)
}

func (h *ExpenseHandler) Summary(c *echo.Context) error {
	var req dto.ExpenseSummaryRequest

	if branchID := c.QueryParam("branch_id"); branchID != "" {
		if b, err := strconv.ParseInt(branchID, 10, 64); err == nil {
			req.BranchID = &b
		}
	}
	req.DateFrom = c.QueryParam("date_from")
	req.DateTo = c.QueryParam("date_to")

	ctx := c.Request().Context()
	companyID := getCompanyID(c)

	result, err := h.expenseSvc.GetExpenseSummary(ctx, companyID, req)
	if err != nil {
		return httputil.Error(c, err)
	}

	return httputil.Success(c, http.StatusOK, "Expense summary retrieved", result)
}

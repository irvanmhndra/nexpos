package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/irvanmhndra/nexpos-api/internal/dto"
	"github.com/irvanmhndra/nexpos-api/internal/service/mocks"
	"github.com/irvanmhndra/nexpos-api/pkg/apperror"
	"github.com/irvanmhndra/nexpos-api/pkg/validator"
	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func setupCustomerHandler(t *testing.T) (*CustomerHandler, *mocks.MockCustomerService) {
	t.Helper()
	v := validator.New()
	mockSvc := new(mocks.MockCustomerService)
	h := NewCustomerHandler(mockSvc, v)
	return h, mockSvc
}

func createCustomerContext(method, path, body string) (*echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("company_id", int64(1))
	return c, rec
}

// =======================
// Create Tests
// =======================

func TestCustomerHandler_Create_Success(t *testing.T) {
	h, mockSvc := setupCustomerHandler(t)

	expected := &dto.CustomerResponse{
		ID:        1,
		Code:      "CUST-001",
		Name:      "Test Customer",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mockSvc.On("Create", mock.Anything, int64(1), mock.AnythingOfType("dto.CreateCustomerRequest")).
		Return(expected, nil)

	body := `{"code": "CUST-001", "name": "Test Customer"}`
	c, rec := createCustomerContext(http.MethodPost, "/api/v1/customers", body)

	err := h.Create(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, true, resp["success"])
	assert.Equal(t, "Customer created successfully", resp["message"])
	assert.NotNil(t, resp["data"])

	mockSvc.AssertExpectations(t)
}

func TestCustomerHandler_Create_InvalidJSON(t *testing.T) {
	h, _ := setupCustomerHandler(t)

	c, rec := createCustomerContext(http.MethodPost, "/api/v1/customers", "{invalid}")

	err := h.Create(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestCustomerHandler_Create_ValidationError(t *testing.T) {
	h, _ := setupCustomerHandler(t)

	body := `{"code": "", "name": ""}`
	c, rec := createCustomerContext(http.MethodPost, "/api/v1/customers", body)

	err := h.Create(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, false, resp["success"])
	assert.Equal(t, "Validation failed", resp["message"])
}

func TestCustomerHandler_Create_ServiceConflict(t *testing.T) {
	h, mockSvc := setupCustomerHandler(t)

	mockSvc.On("Create", mock.Anything, int64(1), mock.AnythingOfType("dto.CreateCustomerRequest")).
		Return(nil, apperror.Conflict("code already exists"))

	body := `{"code": "CUST-001", "name": "Test Customer"}`
	c, rec := createCustomerContext(http.MethodPost, "/api/v1/customers", body)

	err := h.Create(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusConflict, rec.Code)

	mockSvc.AssertExpectations(t)
}

// =======================
// Get Tests
// =======================

func TestCustomerHandler_Get_Success(t *testing.T) {
	h, mockSvc := setupCustomerHandler(t)

	expected := &dto.CustomerResponse{ID: 1, Code: "CUST-001", Name: "Test Customer"}
	mockSvc.On("GetByID", mock.Anything, int64(1), int64(1)).Return(expected, nil)

	c, rec := createCustomerContext(http.MethodGet, "/api/v1/customers/1", "")
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "1"}})

	err := h.Get(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, true, resp["success"])
	assert.Equal(t, "Customer retrieved successfully", resp["message"])

	mockSvc.AssertExpectations(t)
}

func TestCustomerHandler_Get_InvalidID(t *testing.T) {
	h, _ := setupCustomerHandler(t)

	c, rec := createCustomerContext(http.MethodGet, "/api/v1/customers/abc", "")
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "abc"}})

	err := h.Get(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestCustomerHandler_Get_NotFound(t *testing.T) {
	h, mockSvc := setupCustomerHandler(t)

	mockSvc.On("GetByID", mock.Anything, int64(1), int64(99)).
		Return(nil, apperror.NotFound("Customer"))

	c, rec := createCustomerContext(http.MethodGet, "/api/v1/customers/99", "")
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "99"}})

	err := h.Get(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)

	mockSvc.AssertExpectations(t)
}

// =======================
// List Tests
// =======================

func TestCustomerHandler_List_Success(t *testing.T) {
	h, mockSvc := setupCustomerHandler(t)

	expected := &dto.CustomerListResponse{
		Customers: []*dto.CustomerResponse{
			{ID: 1, Code: "CUST-001", Name: "Customer 1"},
		},
		Pagination: &dto.PaginationMeta{
			TotalRecords: 1,
			TotalPages:   1,
			CurrentPage:  1,
			PerPage:      20,
		},
	}

	mockSvc.On("List", mock.Anything, int64(1), mock.AnythingOfType("dto.ListCustomerRequest")).
		Return(expected, nil)

	c, rec := createCustomerContext(http.MethodGet, "/api/v1/customers", "")

	err := h.List(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, true, resp["success"])
	assert.Equal(t, "Customer list retrieved", resp["message"])

	mockSvc.AssertExpectations(t)
}

func TestCustomerHandler_List_ServiceError(t *testing.T) {
	h, mockSvc := setupCustomerHandler(t)

	mockSvc.On("List", mock.Anything, int64(1), mock.AnythingOfType("dto.ListCustomerRequest")).
		Return(nil, apperror.InternalError(nil))

	c, rec := createCustomerContext(http.MethodGet, "/api/v1/customers", "")

	err := h.List(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	mockSvc.AssertExpectations(t)
}

// =======================
// Update Tests
// =======================

func TestCustomerHandler_Update_Success(t *testing.T) {
	h, mockSvc := setupCustomerHandler(t)

	expected := &dto.CustomerResponse{ID: 1, Code: "CUST-001", Name: "Updated"}
	mockSvc.On("Update", mock.Anything, int64(1), int64(1), mock.AnythingOfType("dto.UpdateCustomerRequest")).
		Return(expected, nil)

	body := `{"code": "CUST-001", "name": "Updated"}`
	c, rec := createCustomerContext(http.MethodPut, "/api/v1/customers/1", body)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "1"}})

	err := h.Update(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, "Customer updated successfully", resp["message"])

	mockSvc.AssertExpectations(t)
}

func TestCustomerHandler_Update_InvalidID(t *testing.T) {
	h, _ := setupCustomerHandler(t)

	body := `{"code": "CUST-001", "name": "Updated"}`
	c, rec := createCustomerContext(http.MethodPut, "/api/v1/customers/abc", body)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "abc"}})

	err := h.Update(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestCustomerHandler_Update_ValidationError(t *testing.T) {
	h, _ := setupCustomerHandler(t)

	body := `{"code": "", "name": ""}`
	c, rec := createCustomerContext(http.MethodPut, "/api/v1/customers/1", body)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "1"}})

	err := h.Update(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

func TestCustomerHandler_Update_NotFound(t *testing.T) {
	h, mockSvc := setupCustomerHandler(t)

	mockSvc.On("Update", mock.Anything, int64(1), int64(99), mock.AnythingOfType("dto.UpdateCustomerRequest")).
		Return(nil, apperror.NotFound("Customer"))

	body := `{"code": "CUST-001", "name": "Updated"}`
	c, rec := createCustomerContext(http.MethodPut, "/api/v1/customers/99", body)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "99"}})

	err := h.Update(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)

	mockSvc.AssertExpectations(t)
}

// =======================
// Delete Tests
// =======================

func TestCustomerHandler_Delete_Success(t *testing.T) {
	h, mockSvc := setupCustomerHandler(t)

	mockSvc.On("Delete", mock.Anything, int64(1), int64(1)).Return(nil)

	c, rec := createCustomerContext(http.MethodDelete, "/api/v1/customers/1", "")
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "1"}})

	err := h.Delete(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, "Customer deleted successfully", resp["message"])

	mockSvc.AssertExpectations(t)
}

func TestCustomerHandler_Delete_InvalidID(t *testing.T) {
	h, _ := setupCustomerHandler(t)

	c, rec := createCustomerContext(http.MethodDelete, "/api/v1/customers/abc", "")
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "abc"}})

	err := h.Delete(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestCustomerHandler_Delete_NotFound(t *testing.T) {
	h, mockSvc := setupCustomerHandler(t)

	mockSvc.On("Delete", mock.Anything, int64(1), int64(99)).
		Return(apperror.NotFound("Customer"))

	c, rec := createCustomerContext(http.MethodDelete, "/api/v1/customers/99", "")
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "99"}})

	err := h.Delete(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)

	mockSvc.AssertExpectations(t)
}

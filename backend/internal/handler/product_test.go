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

func setupProductHandler(t *testing.T) (*ProductHandler, *mocks.MockProductService) {
	v := validator.New()
	mockSvc := new(mocks.MockProductService)
	h := NewProductHandler(mockSvc, v)
	return h, mockSvc
}

func createProductContext(method, path, body string) (*echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("company_id", int64(1))
	return c, rec
}

func TestProductHandler_Create_Success(t *testing.T) {
	h, mockSvc := setupProductHandler(t)

	expected := &dto.ProductResponse{
		ID:       1,
		Name:     "Test Product",
		IsActive: true,
		Variants: []*dto.ProductVariantResponse{
			{ID: 1, SKU: "SKU-001", Name: "Default", Price: 10000},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mockSvc.On("Create", mock.Anything, int64(1), mock.AnythingOfType("dto.CreateProductRequest")).
		Return(expected, nil)

	body := `{"name": "Test Product", "variants": [{"sku": "SKU-001", "name": "Default", "price": 10000}]}`
	c, rec := createProductContext(http.MethodPost, "/api/v1/products", body)

	err := h.Create(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, "Product created successfully", resp["message"])

	mockSvc.AssertExpectations(t)
}

func TestProductHandler_Create_ValidationError(t *testing.T) {
	h, _ := setupProductHandler(t)

	body := `{"name": ""}`
	c, rec := createProductContext(http.MethodPost, "/api/v1/products", body)

	err := h.Create(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

func TestProductHandler_Create_InvalidJSON(t *testing.T) {
	h, _ := setupProductHandler(t)

	c, rec := createProductContext(http.MethodPost, "/api/v1/products", "{invalid json}")

	err := h.Create(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestProductHandler_Get_Success(t *testing.T) {
	h, mockSvc := setupProductHandler(t)

	expected := &dto.ProductResponse{ID: 1, Name: "Test Product"}
	mockSvc.On("GetByID", mock.Anything, int64(1), int64(1)).Return(expected, nil)

	c, rec := createProductContext(http.MethodGet, "/api/v1/products/1", "")
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "1"}})

	err := h.Get(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	mockSvc.AssertExpectations(t)
}

func TestProductHandler_Get_NotFound(t *testing.T) {
	h, mockSvc := setupProductHandler(t)

	mockSvc.On("GetByID", mock.Anything, int64(1), int64(99)).
		Return(nil, apperror.NotFound("Product not found"))

	c, rec := createProductContext(http.MethodGet, "/api/v1/products/99", "")
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "99"}})

	err := h.Get(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)

	mockSvc.AssertExpectations(t)
}

func TestProductHandler_Get_InvalidID(t *testing.T) {
	h, _ := setupProductHandler(t)

	c, rec := createProductContext(http.MethodGet, "/api/v1/products/abc", "")
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "abc"}})

	err := h.Get(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestProductHandler_List_Success(t *testing.T) {
	h, mockSvc := setupProductHandler(t)

	expected := &dto.ProductListResponse{
		Products: []*dto.ProductResponse{
			{ID: 1, Name: "Product 1"},
		},
		Pagination: &dto.PaginationMeta{
			TotalRecords: 1,
			TotalPages:   1,
			CurrentPage:  1,
			PerPage:      20,
		},
	}

	mockSvc.On("List", mock.Anything, int64(1), mock.AnythingOfType("dto.ListProductRequest")).
		Return(expected, nil)

	c, rec := createProductContext(http.MethodGet, "/api/v1/products", "")

	err := h.List(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	mockSvc.AssertExpectations(t)
}

func TestProductHandler_Update_Success(t *testing.T) {
	h, mockSvc := setupProductHandler(t)

	expected := &dto.ProductResponse{ID: 1, Name: "Updated"}
	mockSvc.On("Update", mock.Anything, int64(1), int64(1), mock.AnythingOfType("dto.UpdateProductRequest")).
		Return(expected, nil)

	body := `{"name": "Updated", "variants": [{"sku": "SKU-001", "name": "Default", "price": 15000}]}`
	c, rec := createProductContext(http.MethodPut, "/api/v1/products/1", body)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "1"}})

	err := h.Update(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	mockSvc.AssertExpectations(t)
}

func TestProductHandler_Update_InvalidID(t *testing.T) {
	h, _ := setupProductHandler(t)

	body := `{"name": "Updated", "variants": [{"sku": "SKU-001", "name": "Default", "price": 15000}]}`
	c, rec := createProductContext(http.MethodPut, "/api/v1/products/abc", body)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "abc"}})

	err := h.Update(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestProductHandler_Update_ValidationError(t *testing.T) {
	h, _ := setupProductHandler(t)

	body := `{"name": ""}`
	c, rec := createProductContext(http.MethodPut, "/api/v1/products/1", body)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "1"}})

	err := h.Update(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

func TestProductHandler_Delete_Success(t *testing.T) {
	h, mockSvc := setupProductHandler(t)

	mockSvc.On("Delete", mock.Anything, int64(1), int64(1)).Return(nil)

	c, rec := createProductContext(http.MethodDelete, "/api/v1/products/1", "")
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "1"}})

	err := h.Delete(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	mockSvc.AssertExpectations(t)
}

func TestProductHandler_Delete_NotFound(t *testing.T) {
	h, mockSvc := setupProductHandler(t)

	mockSvc.On("Delete", mock.Anything, int64(1), int64(99)).
		Return(apperror.NotFound("Product"))

	c, rec := createProductContext(http.MethodDelete, "/api/v1/products/99", "")
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "99"}})

	err := h.Delete(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)

	mockSvc.AssertExpectations(t)
}

func TestProductHandler_Delete_InvalidID(t *testing.T) {
	h, _ := setupProductHandler(t)

	c, rec := createProductContext(http.MethodDelete, "/api/v1/products/abc", "")
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "abc"}})

	err := h.Delete(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

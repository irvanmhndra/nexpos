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

func setupProductCategoryHandler(t *testing.T) (*ProductCategoryHandler, *mocks.MockProductCategoryService) {
	v := validator.New()
	mockSvc := new(mocks.MockProductCategoryService)
	h := NewProductCategoryHandler(mockSvc, v)
	return h, mockSvc
}

func createCategoryContext(method, path, body string) (*echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("company_id", int64(1))
	return c, rec
}

func TestProductCategoryHandler_Create_Success(t *testing.T) {
	h, mockSvc := setupProductCategoryHandler(t)

	expected := &dto.ProductCategoryResponse{
		ID:        1,
		Code:      "CAT-001",
		Name:      "Electronics",
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mockSvc.On("Create", mock.Anything, int64(1), mock.AnythingOfType("dto.CreateProductCategoryRequest")).
		Return(expected, nil)

	body := `{"code": "CAT-001", "name": "Electronics"}`
	c, rec := createCategoryContext(http.MethodPost, "/api/v1/product-categories", body)

	err := h.Create(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, "Category created successfully", resp["message"])

	mockSvc.AssertExpectations(t)
}

func TestProductCategoryHandler_Create_ValidationError(t *testing.T) {
	h, _ := setupProductCategoryHandler(t)

	body := `{"code": "", "name": ""}`
	c, rec := createCategoryContext(http.MethodPost, "/api/v1/product-categories", body)

	err := h.Create(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

func TestProductCategoryHandler_Create_InvalidJSON(t *testing.T) {
	h, _ := setupProductCategoryHandler(t)

	c, rec := createCategoryContext(http.MethodPost, "/api/v1/product-categories", "{invalid json}")

	err := h.Create(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestProductCategoryHandler_Get_Success(t *testing.T) {
	h, mockSvc := setupProductCategoryHandler(t)

	expected := &dto.ProductCategoryResponse{ID: 1, Code: "CAT-001", Name: "Electronics"}
	mockSvc.On("GetByID", mock.Anything, int64(1), int64(1)).Return(expected, nil)

	c, rec := createCategoryContext(http.MethodGet, "/api/v1/product-categories/1", "")
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "1"}})

	err := h.Get(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	mockSvc.AssertExpectations(t)
}

func TestProductCategoryHandler_Get_InvalidID(t *testing.T) {
	h, _ := setupProductCategoryHandler(t)

	c, rec := createCategoryContext(http.MethodGet, "/api/v1/product-categories/abc", "")
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "abc"}})

	err := h.Get(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestProductCategoryHandler_Get_NotFound(t *testing.T) {
	h, mockSvc := setupProductCategoryHandler(t)

	mockSvc.On("GetByID", mock.Anything, int64(1), int64(99)).
		Return(nil, apperror.NotFound("Category"))

	c, rec := createCategoryContext(http.MethodGet, "/api/v1/product-categories/99", "")
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "99"}})

	err := h.Get(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)

	mockSvc.AssertExpectations(t)
}

func TestProductCategoryHandler_List_Success(t *testing.T) {
	h, mockSvc := setupProductCategoryHandler(t)

	expected := &dto.ProductCategoryListResponse{
		Categories: []*dto.ProductCategoryResponse{
			{ID: 1, Code: "CAT-001", Name: "Electronics"},
		},
		Pagination: &dto.PaginationMeta{
			TotalRecords: 1,
			TotalPages:   1,
			CurrentPage:  1,
			PerPage:      20,
		},
	}

	mockSvc.On("List", mock.Anything, int64(1), mock.AnythingOfType("dto.ListProductCategoryRequest")).
		Return(expected, nil)

	c, rec := createCategoryContext(http.MethodGet, "/api/v1/product-categories", "")

	err := h.List(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	mockSvc.AssertExpectations(t)
}

func TestProductCategoryHandler_ListAll_Success(t *testing.T) {
	h, mockSvc := setupProductCategoryHandler(t)

	expected := []*dto.ProductCategoryResponse{
		{ID: 1, Code: "CAT-001", Name: "Electronics"},
	}

	mockSvc.On("ListAll", mock.Anything, int64(1)).Return(expected, nil)

	c, rec := createCategoryContext(http.MethodGet, "/api/v1/product-categories/all", "")

	err := h.ListAll(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	mockSvc.AssertExpectations(t)
}

func TestProductCategoryHandler_Update_Success(t *testing.T) {
	h, mockSvc := setupProductCategoryHandler(t)

	expected := &dto.ProductCategoryResponse{ID: 1, Code: "CAT-001", Name: "Updated"}
	mockSvc.On("Update", mock.Anything, int64(1), int64(1), mock.AnythingOfType("dto.UpdateProductCategoryRequest")).
		Return(expected, nil)

	body := `{"code": "CAT-001", "name": "Updated"}`
	c, rec := createCategoryContext(http.MethodPut, "/api/v1/product-categories/1", body)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "1"}})

	err := h.Update(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	mockSvc.AssertExpectations(t)
}

func TestProductCategoryHandler_Update_InvalidID(t *testing.T) {
	h, _ := setupProductCategoryHandler(t)

	body := `{"code": "CAT-001", "name": "Updated"}`
	c, rec := createCategoryContext(http.MethodPut, "/api/v1/product-categories/abc", body)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "abc"}})

	err := h.Update(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestProductCategoryHandler_Update_ValidationError(t *testing.T) {
	h, _ := setupProductCategoryHandler(t)

	body := `{"code": "", "name": ""}`
	c, rec := createCategoryContext(http.MethodPut, "/api/v1/product-categories/1", body)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "1"}})

	err := h.Update(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

func TestProductCategoryHandler_Delete_Success(t *testing.T) {
	h, mockSvc := setupProductCategoryHandler(t)

	mockSvc.On("Delete", mock.Anything, int64(1), int64(1)).Return(nil)

	c, rec := createCategoryContext(http.MethodDelete, "/api/v1/product-categories/1", "")
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "1"}})

	err := h.Delete(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	mockSvc.AssertExpectations(t)
}

func TestProductCategoryHandler_Delete_NotFound(t *testing.T) {
	h, mockSvc := setupProductCategoryHandler(t)

	mockSvc.On("Delete", mock.Anything, int64(1), int64(99)).
		Return(apperror.NotFound("Category"))

	c, rec := createCategoryContext(http.MethodDelete, "/api/v1/product-categories/99", "")
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "99"}})

	err := h.Delete(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)

	mockSvc.AssertExpectations(t)
}

func TestProductCategoryHandler_Delete_InvalidID(t *testing.T) {
	h, _ := setupProductCategoryHandler(t)

	c, rec := createCategoryContext(http.MethodDelete, "/api/v1/product-categories/abc", "")
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "abc"}})

	err := h.Delete(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

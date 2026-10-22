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

func setupPromotionHandler(t *testing.T) (*PromotionHandler, *mocks.MockPromotionService) {
	v := validator.New()
	mockSvc := new(mocks.MockPromotionService)
	h := NewPromotionHandler(mockSvc, v)
	return h, mockSvc
}

func createPromotionContext(method, path, body string) (*echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("company_id", int64(1))
	return c, rec
}

func TestPromotionHandler_Create_Success(t *testing.T) {
	h, mockSvc := setupPromotionHandler(t)

	expected := &dto.PromotionResponse{
		ID:        1,
		Code:      "PROMO-001",
		Name:      "Summer Sale",
		Type:      "discount",
		IsActive:  true,
		StartAt:   time.Now(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mockSvc.On("Create", mock.Anything, int64(1), mock.AnythingOfType("dto.CreatePromotionRequest")).
		Return(expected, nil)

	body := `{"code": "PROMO-001", "name": "Summer Sale", "type": "discount", "start_at": "2025-01-01T00:00:00Z", "is_active": true}`
	c, rec := createPromotionContext(http.MethodPost, "/api/v1/promotions", body)

	err := h.Create(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, "Promotion created successfully", resp["message"])

	mockSvc.AssertExpectations(t)
}

func TestPromotionHandler_Create_ValidationError(t *testing.T) {
	h, _ := setupPromotionHandler(t)

	body := `{"code": "", "name": ""}`
	c, rec := createPromotionContext(http.MethodPost, "/api/v1/promotions", body)

	err := h.Create(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

func TestPromotionHandler_Create_InvalidJSON(t *testing.T) {
	h, _ := setupPromotionHandler(t)

	c, rec := createPromotionContext(http.MethodPost, "/api/v1/promotions", "{invalid json}")

	err := h.Create(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestPromotionHandler_Get_Success(t *testing.T) {
	h, mockSvc := setupPromotionHandler(t)

	expected := &dto.PromotionResponse{ID: 1, Code: "PROMO-001", Name: "Summer Sale"}
	mockSvc.On("GetByID", mock.Anything, int64(1), int64(1)).Return(expected, nil)

	c, rec := createPromotionContext(http.MethodGet, "/api/v1/promotions/1", "")
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "1"}})

	err := h.Get(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	mockSvc.AssertExpectations(t)
}

func TestPromotionHandler_Get_NotFound(t *testing.T) {
	h, mockSvc := setupPromotionHandler(t)

	mockSvc.On("GetByID", mock.Anything, int64(1), int64(99)).
		Return(nil, apperror.NotFound("Promotion"))

	c, rec := createPromotionContext(http.MethodGet, "/api/v1/promotions/99", "")
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "99"}})

	err := h.Get(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)

	mockSvc.AssertExpectations(t)
}

func TestPromotionHandler_Get_InvalidID(t *testing.T) {
	h, _ := setupPromotionHandler(t)

	c, rec := createPromotionContext(http.MethodGet, "/api/v1/promotions/abc", "")
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "abc"}})

	err := h.Get(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestPromotionHandler_List_Success(t *testing.T) {
	h, mockSvc := setupPromotionHandler(t)

	expected := &dto.PromotionListResponse{
		Promotions: []*dto.PromotionResponse{
			{ID: 1, Code: "PROMO-001", Name: "Summer Sale"},
		},
		Pagination: &dto.PaginationMeta{
			TotalRecords: 1,
			TotalPages:   1,
			CurrentPage:  1,
			PerPage:      20,
		},
	}

	mockSvc.On("List", mock.Anything, int64(1), mock.AnythingOfType("dto.ListPromotionRequest")).
		Return(expected, nil)

	c, rec := createPromotionContext(http.MethodGet, "/api/v1/promotions", "")

	err := h.List(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	mockSvc.AssertExpectations(t)
}

func TestPromotionHandler_Update_Success(t *testing.T) {
	h, mockSvc := setupPromotionHandler(t)

	expected := &dto.PromotionResponse{ID: 1, Code: "PROMO-001", Name: "Updated Sale"}
	mockSvc.On("Update", mock.Anything, int64(1), int64(1), mock.AnythingOfType("dto.UpdatePromotionRequest")).
		Return(expected, nil)

	body := `{"code": "PROMO-001", "name": "Updated Sale", "type": "discount", "start_at": "2025-01-01T00:00:00Z"}`
	c, rec := createPromotionContext(http.MethodPut, "/api/v1/promotions/1", body)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "1"}})

	err := h.Update(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	mockSvc.AssertExpectations(t)
}

func TestPromotionHandler_Update_InvalidID(t *testing.T) {
	h, _ := setupPromotionHandler(t)

	body := `{"code": "PROMO-001", "name": "Updated", "type": "discount", "start_at": "2025-01-01T00:00:00Z"}`
	c, rec := createPromotionContext(http.MethodPut, "/api/v1/promotions/abc", body)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "abc"}})

	err := h.Update(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestPromotionHandler_Update_ValidationError(t *testing.T) {
	h, _ := setupPromotionHandler(t)

	body := `{"code": "", "name": ""}`
	c, rec := createPromotionContext(http.MethodPut, "/api/v1/promotions/1", body)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "1"}})

	err := h.Update(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

func TestPromotionHandler_Delete_Success(t *testing.T) {
	h, mockSvc := setupPromotionHandler(t)

	mockSvc.On("Delete", mock.Anything, int64(1), int64(1)).Return(nil)

	c, rec := createPromotionContext(http.MethodDelete, "/api/v1/promotions/1", "")
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "1"}})

	err := h.Delete(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	mockSvc.AssertExpectations(t)
}

func TestPromotionHandler_Delete_NotFound(t *testing.T) {
	h, mockSvc := setupPromotionHandler(t)

	mockSvc.On("Delete", mock.Anything, int64(1), int64(99)).
		Return(apperror.NotFound("Promotion"))

	c, rec := createPromotionContext(http.MethodDelete, "/api/v1/promotions/99", "")
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "99"}})

	err := h.Delete(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)

	mockSvc.AssertExpectations(t)
}

func TestPromotionHandler_Delete_InvalidID(t *testing.T) {
	h, _ := setupPromotionHandler(t)

	c, rec := createPromotionContext(http.MethodDelete, "/api/v1/promotions/abc", "")
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "abc"}})

	err := h.Delete(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

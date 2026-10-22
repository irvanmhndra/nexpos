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

func setupBranchHandler(t *testing.T) (*BranchHandler, *mocks.MockBranchService) {
	t.Helper()
	v := validator.New()
	mockSvc := new(mocks.MockBranchService)
	h := NewBranchHandler(mockSvc, v)
	return h, mockSvc
}

func createBranchContext(method, path, body string) (*echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("company_id", int64(1))
	return c, rec
}

func TestBranchHandler_Create_Success(t *testing.T) {
	h, mockSvc := setupBranchHandler(t)

	expected := &dto.BranchResponse{
		ID:        1,
		Code:      "BR-001",
		Name:      "Main Branch",
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mockSvc.On("Create", mock.Anything, int64(1), mock.AnythingOfType("dto.CreateBranchRequest")).
		Return(expected, nil)

	body := `{"code": "BR-001", "name": "Main Branch", "is_active": true}`
	c, rec := createBranchContext(http.MethodPost, "/api/v1/branches", body)

	err := h.Create(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, "Branch created successfully", resp["message"])

	mockSvc.AssertExpectations(t)
}

func TestBranchHandler_Create_InvalidJSON(t *testing.T) {
	h, _ := setupBranchHandler(t)

	c, rec := createBranchContext(http.MethodPost, "/api/v1/branches", "{invalid}")

	err := h.Create(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestBranchHandler_Create_ValidationError(t *testing.T) {
	h, _ := setupBranchHandler(t)

	body := `{"code": "", "name": ""}`
	c, rec := createBranchContext(http.MethodPost, "/api/v1/branches", body)

	err := h.Create(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

func TestBranchHandler_Get_Success(t *testing.T) {
	h, mockSvc := setupBranchHandler(t)

	expected := &dto.BranchResponse{ID: 1, Code: "BR-001", Name: "Main Branch"}
	mockSvc.On("GetByID", mock.Anything, int64(1), int64(1)).Return(expected, nil)

	c, rec := createBranchContext(http.MethodGet, "/api/v1/branches/1", "")
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "1"}})

	err := h.Get(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	mockSvc.AssertExpectations(t)
}

func TestBranchHandler_Get_InvalidID(t *testing.T) {
	h, _ := setupBranchHandler(t)

	c, rec := createBranchContext(http.MethodGet, "/api/v1/branches/abc", "")
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "abc"}})

	err := h.Get(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestBranchHandler_Get_NotFound(t *testing.T) {
	h, mockSvc := setupBranchHandler(t)

	mockSvc.On("GetByID", mock.Anything, int64(1), int64(99)).
		Return(nil, apperror.NotFound("Branch"))

	c, rec := createBranchContext(http.MethodGet, "/api/v1/branches/99", "")
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "99"}})

	err := h.Get(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)

	mockSvc.AssertExpectations(t)
}

func TestBranchHandler_List_Success(t *testing.T) {
	h, mockSvc := setupBranchHandler(t)

	expected := &dto.BranchListResponse{
		Branches: []*dto.BranchResponse{
			{ID: 1, Code: "BR-001", Name: "Main Branch"},
		},
		Pagination: &dto.PaginationMeta{
			TotalRecords: 1,
			TotalPages:   1,
			CurrentPage:  1,
			PerPage:      20,
		},
	}

	mockSvc.On("List", mock.Anything, int64(1), mock.AnythingOfType("dto.ListBranchRequest")).
		Return(expected, nil)

	c, rec := createBranchContext(http.MethodGet, "/api/v1/branches", "")

	err := h.List(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	mockSvc.AssertExpectations(t)
}

func TestBranchHandler_Update_Success(t *testing.T) {
	h, mockSvc := setupBranchHandler(t)

	expected := &dto.BranchResponse{ID: 1, Code: "BR-001", Name: "Updated"}
	mockSvc.On("Update", mock.Anything, int64(1), int64(1), mock.AnythingOfType("dto.UpdateBranchRequest")).
		Return(expected, nil)

	body := `{"code": "BR-001", "name": "Updated", "is_active": true}`
	c, rec := createBranchContext(http.MethodPut, "/api/v1/branches/1", body)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "1"}})

	err := h.Update(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	mockSvc.AssertExpectations(t)
}

func TestBranchHandler_Update_InvalidID(t *testing.T) {
	h, _ := setupBranchHandler(t)

	body := `{"code": "BR-001", "name": "Updated"}`
	c, rec := createBranchContext(http.MethodPut, "/api/v1/branches/abc", body)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "abc"}})

	err := h.Update(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestBranchHandler_Update_ValidationError(t *testing.T) {
	h, _ := setupBranchHandler(t)

	body := `{"code": "", "name": ""}`
	c, rec := createBranchContext(http.MethodPut, "/api/v1/branches/1", body)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "1"}})

	err := h.Update(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

func TestBranchHandler_Delete_Success(t *testing.T) {
	h, mockSvc := setupBranchHandler(t)

	mockSvc.On("Delete", mock.Anything, int64(1), int64(1)).Return(nil)

	c, rec := createBranchContext(http.MethodDelete, "/api/v1/branches/1", "")
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "1"}})

	err := h.Delete(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	mockSvc.AssertExpectations(t)
}

func TestBranchHandler_Delete_InvalidID(t *testing.T) {
	h, _ := setupBranchHandler(t)

	c, rec := createBranchContext(http.MethodDelete, "/api/v1/branches/abc", "")
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "abc"}})

	err := h.Delete(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestBranchHandler_Delete_NotFound(t *testing.T) {
	h, mockSvc := setupBranchHandler(t)

	mockSvc.On("Delete", mock.Anything, int64(1), int64(99)).
		Return(apperror.NotFound("Branch"))

	c, rec := createBranchContext(http.MethodDelete, "/api/v1/branches/99", "")
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "99"}})

	err := h.Delete(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)

	mockSvc.AssertExpectations(t)
}

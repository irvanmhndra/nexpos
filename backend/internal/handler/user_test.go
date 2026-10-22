package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/irvanmhndra/nexpos-api/internal/dto"
	"github.com/irvanmhndra/nexpos-api/internal/service"
	"github.com/irvanmhndra/nexpos-api/internal/service/mocks"
	"github.com/irvanmhndra/nexpos-api/pkg/apperror"
	"github.com/irvanmhndra/nexpos-api/pkg/httputil"
	"github.com/irvanmhndra/nexpos-api/pkg/validator"
	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func setupUserHandler(t *testing.T) (*UserHandler, *mocks.MockUserService) {
	t.Helper()
	v := validator.New()
	mockSvc := new(mocks.MockUserService)
	h := NewUserHandler(mockSvc, v)
	return h, mockSvc
}

func createUserContext(method, path, body string) (*echo.Context, *httptest.ResponseRecorder) {
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

func TestUserHandler_Create_Success(t *testing.T) {
	h, mockSvc := setupUserHandler(t)

	expected := &dto.UserResponse{
		ID:        1,
		Email:     "test@example.com",
		Name:      "Test User",
		Status:    "active",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mockSvc.On("Create", mock.Anything, int64(1), mock.AnythingOfType("dto.CreateUserRequest")).
		Return(expected, nil)

	body := `{"email": "test@example.com", "password": "password123", "name": "Test User"}`
	c, rec := createUserContext(http.MethodPost, "/api/v1/users", body)

	err := h.Create(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, "User created successfully", resp["message"])

	mockSvc.AssertExpectations(t)
}

func TestUserHandler_Create_InvalidJSON(t *testing.T) {
	h, _ := setupUserHandler(t)

	c, rec := createUserContext(http.MethodPost, "/api/v1/users", "{invalid}")

	err := h.Create(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestUserHandler_Create_ValidationError(t *testing.T) {
	h, _ := setupUserHandler(t)

	body := `{"email": "", "password": "", "name": ""}`
	c, rec := createUserContext(http.MethodPost, "/api/v1/users", body)

	err := h.Create(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

// =======================
// Get Tests
// =======================

func TestUserHandler_Get_Success(t *testing.T) {
	h, mockSvc := setupUserHandler(t)

	expected := &dto.UserResponse{ID: 1, Email: "test@example.com", Name: "Test User"}
	mockSvc.On("GetByID", mock.Anything, int64(1), int64(1)).Return(expected, nil)

	c, rec := createUserContext(http.MethodGet, "/api/v1/users/1", "")
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "1"}})

	err := h.Get(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	mockSvc.AssertExpectations(t)
}

func TestUserHandler_Get_InvalidID(t *testing.T) {
	h, _ := setupUserHandler(t)

	c, rec := createUserContext(http.MethodGet, "/api/v1/users/abc", "")
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "abc"}})

	err := h.Get(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestUserHandler_Get_NotFound(t *testing.T) {
	h, mockSvc := setupUserHandler(t)

	mockSvc.On("GetByID", mock.Anything, int64(1), int64(99)).
		Return(nil, apperror.NotFound("User"))

	c, rec := createUserContext(http.MethodGet, "/api/v1/users/99", "")
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "99"}})

	err := h.Get(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)

	mockSvc.AssertExpectations(t)
}

// =======================
// List Tests
// =======================

func TestUserHandler_List_Success(t *testing.T) {
	h, mockSvc := setupUserHandler(t)

	expected := &service.UserListResult{
		Users: []*dto.UserResponse{
			{ID: 1, Email: "test@example.com", Name: "Test User"},
		},
		Pagination: &httputil.Pagination{
			TotalRecords: 1,
			TotalPages:   1,
			CurrentPage:  1,
			PerPage:      20,
		},
	}

	mockSvc.On("List", mock.Anything, int64(1), mock.AnythingOfType("dto.ListUsersRequest")).
		Return(expected, nil)

	c, rec := createUserContext(http.MethodGet, "/api/v1/users", "")

	err := h.List(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	mockSvc.AssertExpectations(t)
}

// =======================
// Update Tests
// =======================

func TestUserHandler_Update_Success(t *testing.T) {
	h, mockSvc := setupUserHandler(t)

	expected := &dto.UserResponse{ID: 1, Email: "updated@example.com", Name: "Updated"}
	mockSvc.On("Update", mock.Anything, int64(1), int64(1), mock.AnythingOfType("dto.UpdateUserRequest")).
		Return(expected, nil)

	body := `{"email": "updated@example.com", "name": "Updated"}`
	c, rec := createUserContext(http.MethodPut, "/api/v1/users/1", body)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "1"}})

	err := h.Update(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	mockSvc.AssertExpectations(t)
}

func TestUserHandler_Update_InvalidID(t *testing.T) {
	h, _ := setupUserHandler(t)

	body := `{"email": "updated@example.com", "name": "Updated"}`
	c, rec := createUserContext(http.MethodPut, "/api/v1/users/abc", body)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "abc"}})

	err := h.Update(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// =======================
// UpdateStatus Tests
// =======================

func TestUserHandler_UpdateStatus_Success(t *testing.T) {
	h, mockSvc := setupUserHandler(t)

	mockSvc.On("UpdateStatus", mock.Anything, int64(1), int64(1), mock.AnythingOfType("dto.UpdateUserStatusRequest")).
		Return(nil)

	body := `{"status": "inactive"}`
	c, rec := createUserContext(http.MethodPatch, "/api/v1/users/1/status", body)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "1"}})

	err := h.UpdateStatus(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	mockSvc.AssertExpectations(t)
}

func TestUserHandler_UpdateStatus_InvalidID(t *testing.T) {
	h, _ := setupUserHandler(t)

	body := `{"status": "inactive"}`
	c, rec := createUserContext(http.MethodPatch, "/api/v1/users/abc/status", body)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "abc"}})

	err := h.UpdateStatus(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// =======================
// Delete Tests
// =======================

func TestUserHandler_Delete_Success(t *testing.T) {
	h, mockSvc := setupUserHandler(t)

	mockSvc.On("Delete", mock.Anything, int64(1), int64(1)).Return(nil)

	c, rec := createUserContext(http.MethodDelete, "/api/v1/users/1", "")
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "1"}})

	err := h.Delete(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	mockSvc.AssertExpectations(t)
}

func TestUserHandler_Delete_InvalidID(t *testing.T) {
	h, _ := setupUserHandler(t)

	c, rec := createUserContext(http.MethodDelete, "/api/v1/users/abc", "")
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "abc"}})

	err := h.Delete(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestUserHandler_Delete_NotFound(t *testing.T) {
	h, mockSvc := setupUserHandler(t)

	mockSvc.On("Delete", mock.Anything, int64(1), int64(99)).
		Return(apperror.NotFound("User"))

	c, rec := createUserContext(http.MethodDelete, "/api/v1/users/99", "")
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "99"}})

	err := h.Delete(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)

	mockSvc.AssertExpectations(t)
}

package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/irvanmhndra/nexpos-api/internal/dto"
	"github.com/irvanmhndra/nexpos-api/internal/model"
	repoMocks "github.com/irvanmhndra/nexpos-api/internal/repository/mocks"
	"github.com/irvanmhndra/nexpos-api/pkg/apperror"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func setupUserTest(t *testing.T) (*UserService, *repoMocks.MockUserRepository) {
	t.Helper()
	mockRepo := repoMocks.NewMockUserRepository(t)
	svc := NewUserService(mockRepo)
	return svc, mockRepo
}

func createTestUser(id, companyID int64) *model.User {
	now := time.Now()
	return &model.User{
		ID:           id,
		CompanyID:    companyID,
		Email:        "user@example.com",
		PasswordHash: "$2a$10$dummy",
		Name:         "Test User",
		Status:       model.UserStatusActive,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

func TestNewUserService(t *testing.T) {
	mockRepo := repoMocks.NewMockUserRepository(t)
	svc := NewUserService(mockRepo)
	assert.NotNil(t, svc)
}

// Create
func TestUserService_Create_Success(t *testing.T) {
	svc, mockRepo := setupUserTest(t)
	ctx := context.Background()

	req := dto.CreateUserRequest{Email: "new@example.com", Password: "password123", Name: "New User"}
	mockRepo.EXPECT().GetByEmail(ctx, "new@example.com").Return(nil, nil).Once()
	mockRepo.EXPECT().Create(ctx, mock.AnythingOfType("*model.User")).Return(nil).
		Run(func(args mock.Arguments) {
			u := args.Get(1).(*model.User)
			u.ID = 1
			u.CreatedAt = time.Now()
			u.UpdatedAt = time.Now()
		}).Once()

	result, err := svc.Create(ctx, 1, req)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "new@example.com", result.Email)
}

func TestUserService_Create_EmailExists(t *testing.T) {
	svc, mockRepo := setupUserTest(t)
	ctx := context.Background()

	existing := createTestUser(1, 1)
	req := dto.CreateUserRequest{Email: "user@example.com", Password: "password123", Name: "User"}
	mockRepo.EXPECT().GetByEmail(ctx, "user@example.com").Return(existing, nil).Once()

	result, err := svc.Create(ctx, 1, req)
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "email already exists")
}

func TestUserService_Create_GetByEmailError(t *testing.T) {
	svc, mockRepo := setupUserTest(t)
	ctx := context.Background()

	req := dto.CreateUserRequest{Email: "new@example.com", Password: "password123", Name: "User"}
	mockRepo.EXPECT().GetByEmail(ctx, "new@example.com").Return(nil, errors.New("db error")).Once()

	result, err := svc.Create(ctx, 1, req)
	require.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, apperror.IsInternalError(err))
}

// GetByID
func TestUserService_GetByID_Success(t *testing.T) {
	svc, mockRepo := setupUserTest(t)
	ctx := context.Background()
	expected := createTestUser(1, 1)

	mockRepo.EXPECT().GetByID(ctx, int64(1), int64(1)).Return(expected, nil).Once()

	result, err := svc.GetByID(ctx, 1, 1)
	require.NoError(t, err)
	assert.Equal(t, expected.Email, result.Email)
}

func TestUserService_GetByID_NotFound(t *testing.T) {
	svc, mockRepo := setupUserTest(t)
	ctx := context.Background()

	mockRepo.EXPECT().GetByID(ctx, int64(1), int64(999)).Return(nil, nil).Once()

	result, err := svc.GetByID(ctx, 1, 999)
	require.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, apperror.IsNotFound(err))
}

// List
func TestUserService_List_Success(t *testing.T) {
	svc, mockRepo := setupUserTest(t)
	ctx := context.Background()
	users := []*model.User{createTestUser(1, 1), createTestUser(2, 1)}

	// SetDefaults: page=0→1, limit=0→10; offset=(1-1)*10=0
	req := dto.ListUsersRequest{Page: 0, Limit: 0}
	mockRepo.EXPECT().List(ctx, int64(1), 10, 0).Return(users, 2, nil).Once()

	result, err := svc.List(ctx, 1, req)
	require.NoError(t, err)
	assert.Len(t, result.Users, 2)
	assert.Equal(t, 2, result.Pagination.TotalRecords)
}

func TestUserService_List_RepoError(t *testing.T) {
	svc, mockRepo := setupUserTest(t)
	ctx := context.Background()

	req := dto.ListUsersRequest{Page: 1, Limit: 10}
	mockRepo.EXPECT().List(ctx, int64(1), 10, 0).Return(nil, 0, errors.New("db error")).Once()

	result, err := svc.List(ctx, 1, req)
	require.Error(t, err)
	assert.Nil(t, result)
}

// Update
func TestUserService_Update_Success(t *testing.T) {
	svc, mockRepo := setupUserTest(t)
	ctx := context.Background()
	existing := createTestUser(1, 1)

	req := dto.UpdateUserRequest{Email: "updated@example.com", Name: "Updated"}
	mockRepo.EXPECT().GetByID(ctx, int64(1), int64(1)).Return(existing, nil).Once()
	mockRepo.EXPECT().EmailExists(ctx, "updated@example.com", int64(1)).Return(false, nil).Once()
	mockRepo.EXPECT().Update(ctx, mock.AnythingOfType("*model.User")).Return(nil).Once()

	result, err := svc.Update(ctx, 1, 1, req)
	require.NoError(t, err)
	assert.Equal(t, "updated@example.com", result.Email)
}

func TestUserService_Update_NotFound(t *testing.T) {
	svc, mockRepo := setupUserTest(t)
	ctx := context.Background()

	req := dto.UpdateUserRequest{Name: "Updated"}
	mockRepo.EXPECT().GetByID(ctx, int64(1), int64(999)).Return(nil, nil).Once()

	result, err := svc.Update(ctx, 1, 999, req)
	require.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, apperror.IsNotFound(err))
}

func TestUserService_Update_EmailExists(t *testing.T) {
	svc, mockRepo := setupUserTest(t)
	ctx := context.Background()
	existing := createTestUser(1, 1)

	req := dto.UpdateUserRequest{Email: "taken@example.com"}
	mockRepo.EXPECT().GetByID(ctx, int64(1), int64(1)).Return(existing, nil).Once()
	mockRepo.EXPECT().EmailExists(ctx, "taken@example.com", int64(1)).Return(true, nil).Once()

	result, err := svc.Update(ctx, 1, 1, req)
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "email already exists")
}

func TestUserService_Update_SameEmail_NoCheck(t *testing.T) {
	svc, mockRepo := setupUserTest(t)
	ctx := context.Background()
	existing := createTestUser(1, 1)

	// Email same as existing - no EmailExists check needed
	req := dto.UpdateUserRequest{Email: "user@example.com", Name: "Updated"}
	mockRepo.EXPECT().GetByID(ctx, int64(1), int64(1)).Return(existing, nil).Once()
	mockRepo.EXPECT().Update(ctx, mock.AnythingOfType("*model.User")).Return(nil).Once()

	result, err := svc.Update(ctx, 1, 1, req)
	require.NoError(t, err)
	assert.Equal(t, "Updated", result.Name)
}

// UpdateStatus
func TestUserService_UpdateStatus_Success(t *testing.T) {
	svc, mockRepo := setupUserTest(t)
	ctx := context.Background()
	existing := createTestUser(1, 1)

	req := dto.UpdateUserStatusRequest{Status: "inactive"}
	mockRepo.EXPECT().GetByID(ctx, int64(1), int64(1)).Return(existing, nil).Once()
	mockRepo.EXPECT().UpdateStatus(ctx, int64(1), int64(1), model.UserStatus("inactive")).Return(nil).Once()

	err := svc.UpdateStatus(ctx, 1, 1, req)
	require.NoError(t, err)
}

func TestUserService_UpdateStatus_NotFound(t *testing.T) {
	svc, mockRepo := setupUserTest(t)
	ctx := context.Background()

	req := dto.UpdateUserStatusRequest{Status: "inactive"}
	mockRepo.EXPECT().GetByID(ctx, int64(1), int64(999)).Return(nil, nil).Once()

	err := svc.UpdateStatus(ctx, 1, 999, req)
	require.Error(t, err)
	assert.True(t, apperror.IsNotFound(err))
}

// Delete
func TestUserService_Delete_Success(t *testing.T) {
	svc, mockRepo := setupUserTest(t)
	ctx := context.Background()
	existing := createTestUser(1, 1)

	mockRepo.EXPECT().GetByID(ctx, int64(1), int64(1)).Return(existing, nil).Once()
	mockRepo.EXPECT().Delete(ctx, int64(1), int64(1)).Return(nil).Once()

	err := svc.Delete(ctx, 1, 1)
	require.NoError(t, err)
}

func TestUserService_Delete_NotFound(t *testing.T) {
	svc, mockRepo := setupUserTest(t)
	ctx := context.Background()

	mockRepo.EXPECT().GetByID(ctx, int64(1), int64(999)).Return(nil, nil).Once()

	err := svc.Delete(ctx, 1, 999)
	require.Error(t, err)
	assert.True(t, apperror.IsNotFound(err))
}

func TestUserService_Delete_RepoError(t *testing.T) {
	svc, mockRepo := setupUserTest(t)
	ctx := context.Background()
	existing := createTestUser(1, 1)

	mockRepo.EXPECT().GetByID(ctx, int64(1), int64(1)).Return(existing, nil).Once()
	mockRepo.EXPECT().Delete(ctx, int64(1), int64(1)).Return(errors.New("db error")).Once()

	err := svc.Delete(ctx, 1, 1)
	require.Error(t, err)
	assert.True(t, apperror.IsInternalError(err))
}

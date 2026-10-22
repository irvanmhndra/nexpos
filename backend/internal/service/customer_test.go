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

func setupCustomerTest(t *testing.T) (*CustomerService, *repoMocks.MockCustomerRepository) {
	t.Helper()
	mockRepo := repoMocks.NewMockCustomerRepository(t)
	service := NewCustomerService(mockRepo)
	return service, mockRepo
}

func createTestCustomer(id, companyID int64, code string) *model.Customer {
	now := time.Now()
	phone := "+1234567890"
	email := "test@example.com"
	return &model.Customer{
		ID:        id,
		CompanyID: companyID,
		Code:      code,
		Name:      "Test Customer " + code,
		Phone:     &phone,
		Email:     &email,
		IsMember:  true,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// Constructor
func TestNewCustomerService(t *testing.T) {
	mockRepo := repoMocks.NewMockCustomerRepository(t)
	svc := NewCustomerService(mockRepo)
	assert.NotNil(t, svc)
}

// Create
func TestCustomerService_Create_Success(t *testing.T) {
	svc, mockRepo := setupCustomerTest(t)
	ctx := context.Background()
	companyID := int64(1)
	phone := "+1234567890"

	req := dto.CreateCustomerRequest{
		Code: "C001", Name: "John", Phone: &phone, IsMember: true,
	}

	mockRepo.EXPECT().CodeExists(ctx, companyID, "C001", int64(0)).Return(false, nil).Once()
	mockRepo.EXPECT().Create(ctx, mock.AnythingOfType("*model.Customer")).Return(nil).
		Run(func(args mock.Arguments) {
			c := args.Get(1).(*model.Customer)
			c.ID = 1
			c.CreatedAt = time.Now()
			c.UpdatedAt = time.Now()
		}).Once()

	result, err := svc.Create(ctx, companyID, req)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "C001", result.Code)
	assert.Equal(t, "John", result.Name)
}

func TestCustomerService_Create_CodeExists(t *testing.T) {
	svc, mockRepo := setupCustomerTest(t)
	ctx := context.Background()

	req := dto.CreateCustomerRequest{Code: "C001", Name: "John"}
	mockRepo.EXPECT().CodeExists(ctx, int64(1), "C001", int64(0)).Return(true, nil).Once()

	result, err := svc.Create(ctx, 1, req)
	require.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, apperror.IsBadRequest(err))
}

func TestCustomerService_Create_CodeExistsError(t *testing.T) {
	svc, mockRepo := setupCustomerTest(t)
	ctx := context.Background()

	req := dto.CreateCustomerRequest{Code: "C001", Name: "John"}
	mockRepo.EXPECT().CodeExists(ctx, int64(1), "C001", int64(0)).Return(false, errors.New("db error")).Once()

	result, err := svc.Create(ctx, 1, req)
	require.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, apperror.IsInternalError(err))
}

func TestCustomerService_Create_RepoError(t *testing.T) {
	svc, mockRepo := setupCustomerTest(t)
	ctx := context.Background()

	req := dto.CreateCustomerRequest{Code: "C001", Name: "John"}
	mockRepo.EXPECT().CodeExists(ctx, int64(1), "C001", int64(0)).Return(false, nil).Once()
	mockRepo.EXPECT().Create(ctx, mock.AnythingOfType("*model.Customer")).Return(errors.New("db error")).Once()

	result, err := svc.Create(ctx, 1, req)
	require.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, apperror.IsInternalError(err))
}

// GetByID
func TestCustomerService_GetByID_Success(t *testing.T) {
	svc, mockRepo := setupCustomerTest(t)
	ctx := context.Background()
	expected := createTestCustomer(1, 1, "C001")

	mockRepo.EXPECT().GetByID(ctx, int64(1), int64(1)).Return(expected, nil).Once()

	result, err := svc.GetByID(ctx, 1, 1)
	require.NoError(t, err)
	assert.Equal(t, expected.ID, result.ID)
	assert.Equal(t, expected.Code, result.Code)
}

func TestCustomerService_GetByID_NotFound(t *testing.T) {
	svc, mockRepo := setupCustomerTest(t)
	ctx := context.Background()

	mockRepo.EXPECT().GetByID(ctx, int64(1), int64(999)).Return(nil, nil).Once()

	result, err := svc.GetByID(ctx, 1, 999)
	require.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, apperror.IsNotFound(err))
}

func TestCustomerService_GetByID_RepoError(t *testing.T) {
	svc, mockRepo := setupCustomerTest(t)
	ctx := context.Background()

	mockRepo.EXPECT().GetByID(ctx, int64(1), int64(1)).Return(nil, errors.New("db error")).Once()

	result, err := svc.GetByID(ctx, 1, 1)
	require.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, apperror.IsInternalError(err))
}

// List
func TestCustomerService_List_Success(t *testing.T) {
	svc, mockRepo := setupCustomerTest(t)
	ctx := context.Background()
	customers := []*model.Customer{createTestCustomer(1, 1, "C001"), createTestCustomer(2, 1, "C002")}

	req := dto.ListCustomerRequest{Page: 1, PerPage: 20}
	mockRepo.EXPECT().List(ctx, int64(1), "", (*bool)(nil), 20, 0).Return(customers, 2, nil).Once()

	result, err := svc.List(ctx, 1, req)
	require.NoError(t, err)
	assert.Len(t, result.Customers, 2)
	assert.Equal(t, 2, result.Pagination.TotalRecords)
}

func TestCustomerService_List_Pagination(t *testing.T) {
	tests := []struct {
		name            string
		page, perPage   int
		expectedPerPage int
		expectedOffset  int
	}{
		{"defaults", 0, 0, 20, 0},
		{"page 2", 2, 10, 10, 10},
		{"max perPage", 1, 150, 100, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, mockRepo := setupCustomerTest(t)
			ctx := context.Background()
			req := dto.ListCustomerRequest{Page: tt.page, PerPage: tt.perPage}
			mockRepo.EXPECT().List(ctx, int64(1), "", (*bool)(nil), tt.expectedPerPage, tt.expectedOffset).Return([]*model.Customer{}, 0, nil).Once()

			result, err := svc.List(ctx, 1, req)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedPerPage, result.Pagination.PerPage)
		})
	}
}

func TestCustomerService_List_RepoError(t *testing.T) {
	svc, mockRepo := setupCustomerTest(t)
	ctx := context.Background()
	req := dto.ListCustomerRequest{Page: 1, PerPage: 20}
	mockRepo.EXPECT().List(ctx, int64(1), "", (*bool)(nil), 20, 0).Return(nil, 0, errors.New("db error")).Once()

	result, err := svc.List(ctx, 1, req)
	require.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, apperror.IsInternalError(err))
}

// Update
func TestCustomerService_Update_Success(t *testing.T) {
	svc, mockRepo := setupCustomerTest(t)
	ctx := context.Background()
	existing := createTestCustomer(1, 1, "C001")

	req := dto.UpdateCustomerRequest{Code: "C002", Name: "Updated"}
	mockRepo.EXPECT().GetByID(ctx, int64(1), int64(1)).Return(existing, nil).Once()
	mockRepo.EXPECT().CodeExists(ctx, int64(1), "C002", int64(1)).Return(false, nil).Once()
	mockRepo.EXPECT().Update(ctx, mock.AnythingOfType("*model.Customer")).Return(nil).Once()
	// Update calls GetByID again to return fresh data
	updated := createTestCustomer(1, 1, "C002")
	updated.Name = "Updated"
	mockRepo.EXPECT().GetByID(ctx, int64(1), int64(1)).Return(updated, nil).Once()

	result, err := svc.Update(ctx, 1, 1, req)
	require.NoError(t, err)
	assert.NotNil(t, result)
}

func TestCustomerService_Update_NotFound(t *testing.T) {
	svc, mockRepo := setupCustomerTest(t)
	ctx := context.Background()

	req := dto.UpdateCustomerRequest{Code: "C002", Name: "Updated"}
	mockRepo.EXPECT().GetByID(ctx, int64(1), int64(999)).Return(nil, nil).Once()

	result, err := svc.Update(ctx, 1, 999, req)
	require.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, apperror.IsNotFound(err))
}

func TestCustomerService_Update_CodeExists(t *testing.T) {
	svc, mockRepo := setupCustomerTest(t)
	ctx := context.Background()
	existing := createTestCustomer(1, 1, "C001")

	req := dto.UpdateCustomerRequest{Code: "C002", Name: "Updated"}
	mockRepo.EXPECT().GetByID(ctx, int64(1), int64(1)).Return(existing, nil).Once()
	mockRepo.EXPECT().CodeExists(ctx, int64(1), "C002", int64(1)).Return(true, nil).Once()

	result, err := svc.Update(ctx, 1, 1, req)
	require.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, apperror.IsBadRequest(err))
}

// Delete
func TestCustomerService_Delete_Success(t *testing.T) {
	svc, mockRepo := setupCustomerTest(t)
	ctx := context.Background()
	existing := createTestCustomer(1, 1, "C001")

	mockRepo.EXPECT().GetByID(ctx, int64(1), int64(1)).Return(existing, nil).Once()
	mockRepo.EXPECT().Delete(ctx, int64(1), int64(1)).Return(nil).Once()

	err := svc.Delete(ctx, 1, 1)
	require.NoError(t, err)
}

func TestCustomerService_Delete_NotFound(t *testing.T) {
	svc, mockRepo := setupCustomerTest(t)
	ctx := context.Background()

	mockRepo.EXPECT().GetByID(ctx, int64(1), int64(999)).Return(nil, nil).Once()

	err := svc.Delete(ctx, 1, 999)
	require.Error(t, err)
	assert.True(t, apperror.IsNotFound(err))
}

func TestCustomerService_Delete_RepoError(t *testing.T) {
	svc, mockRepo := setupCustomerTest(t)
	ctx := context.Background()
	existing := createTestCustomer(1, 1, "C001")

	mockRepo.EXPECT().GetByID(ctx, int64(1), int64(1)).Return(existing, nil).Once()
	mockRepo.EXPECT().Delete(ctx, int64(1), int64(1)).Return(errors.New("db error")).Once()

	err := svc.Delete(ctx, 1, 1)
	require.Error(t, err)
	assert.True(t, apperror.IsInternalError(err))
}

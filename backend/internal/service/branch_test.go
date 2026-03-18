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

// =======================
// Test Helpers
// =======================

func setupBranchTest(t *testing.T) (*BranchService, *repoMocks.MockBranchRepository) {
	t.Helper()
	mockRepo := repoMocks.NewMockBranchRepository(t)
	service := NewBranchService(mockRepo)
	return service, mockRepo
}

func strPtr(s string) *string { return &s }

func createTestBranch(id int64, companyID int64, code string) *model.Branch {
	now := time.Now()
	return &model.Branch{
		ID:        id,
		CompanyID: companyID,
		Code:      code,
		Name:      "Test Branch " + code,
		Address:   strPtr("123 Test Street"),
		Phone:     strPtr("+1234567890"),
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// =======================
// Constructor Tests
// =======================

func TestNewBranchService(t *testing.T) {
	mockRepo := repoMocks.NewMockBranchRepository(t)
	service := NewBranchService(mockRepo)

	assert.NotNil(t, service)
	assert.Equal(t, mockRepo, service.branchRepo)
}

// =======================
// Create Tests
// =======================

func TestBranchService_Create_Success(t *testing.T) {
	// Arrange
	service, mockRepo := setupBranchTest(t)
	ctx := context.Background()
	companyID := int64(1)

	req := dto.CreateBranchRequest{
		Code:     "BR001",
		Name:     "Main Branch",
		Address:  strPtr("123 Main St"),
		Phone:    strPtr("+1234567890"),
		IsActive: true,
	}

	mockRepo.EXPECT().
		CodeExists(ctx, companyID, req.Code, int64(0)).
		Return(false, nil).
		Once()

	mockRepo.EXPECT().
		Create(ctx, mock.MatchedBy(func(b *model.Branch) bool {
			return b.CompanyID == companyID &&
				b.Code == req.Code &&
				b.Name == req.Name &&
				*b.Address == *req.Address &&
				*b.Phone == *req.Phone &&
				b.IsActive == req.IsActive
		})).
		Return(nil).
		Run(func(args mock.Arguments) {
			// Simulate database setting the ID
			branch := args.Get(1).(*model.Branch)
			branch.ID = 1
			branch.CreatedAt = time.Now()
			branch.UpdatedAt = time.Now()
		}).
		Once()

	// Act
	result, err := service.Create(ctx, companyID, req)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, int64(1), result.ID)
	assert.Equal(t, req.Code, result.Code)
	assert.Equal(t, req.Name, result.Name)
	assert.Equal(t, req.Address, result.Address)
	assert.Equal(t, req.Phone, result.Phone)
	assert.Equal(t, req.IsActive, result.IsActive)
	mockRepo.AssertExpectations(t)
}

func TestBranchService_Create_CodeExists(t *testing.T) {
	// Arrange
	service, mockRepo := setupBranchTest(t)
	ctx := context.Background()
	companyID := int64(1)

	req := dto.CreateBranchRequest{
		Code:     "BR001",
		Name:     "Main Branch",
		Address:  strPtr("123 Main St"),
		Phone:    strPtr("+1234567890"),
		IsActive: true,
	}

	mockRepo.EXPECT().
		CodeExists(ctx, companyID, req.Code, int64(0)).
		Return(true, nil).
		Once()

	// Act
	result, err := service.Create(ctx, companyID, req)

	// Assert
	require.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, apperror.IsBadRequest(err))
	assert.Contains(t, err.Error(), "Branch code already exists")
	mockRepo.AssertExpectations(t)
}

func TestBranchService_Create_CodeExistsCheckError(t *testing.T) {
	// Arrange
	service, mockRepo := setupBranchTest(t)
	ctx := context.Background()
	companyID := int64(1)

	req := dto.CreateBranchRequest{
		Code:     "BR001",
		Name:     "Main Branch",
		Address:  strPtr("123 Main St"),
		Phone:    strPtr("+1234567890"),
		IsActive: true,
	}

	dbError := errors.New("database connection error")
	mockRepo.EXPECT().
		CodeExists(ctx, companyID, req.Code, int64(0)).
		Return(false, dbError).
		Once()

	// Act
	result, err := service.Create(ctx, companyID, req)

	// Assert
	require.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, apperror.IsInternalError(err))
	mockRepo.AssertExpectations(t)
}

func TestBranchService_Create_RepositoryError(t *testing.T) {
	// Arrange
	service, mockRepo := setupBranchTest(t)
	ctx := context.Background()
	companyID := int64(1)

	req := dto.CreateBranchRequest{
		Code:     "BR001",
		Name:     "Main Branch",
		Address:  strPtr("123 Main St"),
		Phone:    strPtr("+1234567890"),
		IsActive: true,
	}

	mockRepo.EXPECT().
		CodeExists(ctx, companyID, req.Code, int64(0)).
		Return(false, nil).
		Once()

	dbError := errors.New("database insert error")
	mockRepo.EXPECT().
		Create(ctx, mock.AnythingOfType("*model.Branch")).
		Return(dbError).
		Once()

	// Act
	result, err := service.Create(ctx, companyID, req)

	// Assert
	require.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, apperror.IsInternalError(err))
	mockRepo.AssertExpectations(t)
}

// =======================
// GetByID Tests
// =======================

func TestBranchService_GetByID_Success(t *testing.T) {
	// Arrange
	service, mockRepo := setupBranchTest(t)
	ctx := context.Background()
	companyID := int64(1)
	branchID := int64(10)

	expectedBranch := createTestBranch(branchID, companyID, "BR001")

	mockRepo.EXPECT().
		GetByID(ctx, branchID).
		Return(expectedBranch, nil).
		Once()

	// Act
	result, err := service.GetByID(ctx, companyID, branchID)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedBranch.ID, result.ID)
	assert.Equal(t, expectedBranch.Code, result.Code)
	assert.Equal(t, expectedBranch.Name, result.Name)
	assert.Equal(t, expectedBranch.Address, result.Address)
	assert.Equal(t, expectedBranch.Phone, result.Phone)
	assert.Equal(t, expectedBranch.IsActive, result.IsActive)
	mockRepo.AssertExpectations(t)
}

func TestBranchService_GetByID_NotFound(t *testing.T) {
	// Arrange
	service, mockRepo := setupBranchTest(t)
	ctx := context.Background()
	companyID := int64(1)
	branchID := int64(999)

	mockRepo.EXPECT().
		GetByID(ctx, branchID).
		Return(nil, nil).
		Once()

	// Act
	result, err := service.GetByID(ctx, companyID, branchID)

	// Assert
	require.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, apperror.IsNotFound(err))
	assert.Contains(t, err.Error(), "Branch not found")
	mockRepo.AssertExpectations(t)
}

func TestBranchService_GetByID_WrongCompany(t *testing.T) {
	// Arrange
	service, mockRepo := setupBranchTest(t)
	ctx := context.Background()
	requestedCompanyID := int64(1)
	branchID := int64(10)

	// Branch belongs to company 2, but we're requesting with company 1
	branchFromDifferentCompany := createTestBranch(branchID, int64(2), "BR001")

	mockRepo.EXPECT().
		GetByID(ctx, branchID).
		Return(branchFromDifferentCompany, nil).
		Once()

	// Act
	result, err := service.GetByID(ctx, requestedCompanyID, branchID)

	// Assert
	require.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, apperror.IsNotFound(err))
	assert.Contains(t, err.Error(), "Branch not found")
	mockRepo.AssertExpectations(t)
}

func TestBranchService_GetByID_RepositoryError(t *testing.T) {
	// Arrange
	service, mockRepo := setupBranchTest(t)
	ctx := context.Background()
	companyID := int64(1)
	branchID := int64(10)

	dbError := errors.New("database query error")
	mockRepo.EXPECT().
		GetByID(ctx, branchID).
		Return(nil, dbError).
		Once()

	// Act
	result, err := service.GetByID(ctx, companyID, branchID)

	// Assert
	require.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, apperror.IsInternalError(err))
	mockRepo.AssertExpectations(t)
}

// =======================
// List Tests
// =======================

func TestBranchService_List_Success(t *testing.T) {
	// Arrange
	service, mockRepo := setupBranchTest(t)
	ctx := context.Background()
	companyID := int64(1)

	req := dto.ListBranchRequest{
		Page:    1,
		PerPage: 20,
	}

	expectedBranches := []*model.Branch{
		createTestBranch(1, companyID, "BR001"),
		createTestBranch(2, companyID, "BR002"),
		createTestBranch(3, companyID, "BR003"),
	}
	totalRecords := 3

	mockRepo.EXPECT().
		List(ctx, companyID, "", (*bool)(nil), 20, 0).
		Return(expectedBranches, totalRecords, nil).
		Once()

	// Act
	result, err := service.List(ctx, companyID, req)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Branches, 3)
	assert.Equal(t, 3, result.Pagination.TotalRecords)
	assert.Equal(t, 1, result.Pagination.TotalPages)
	assert.Equal(t, 1, result.Pagination.CurrentPage)
	assert.Equal(t, 20, result.Pagination.PerPage)
	assert.Nil(t, result.Pagination.NextPage)
	assert.Nil(t, result.Pagination.PrevPage)
	mockRepo.AssertExpectations(t)
}

func TestBranchService_List_WithFilters(t *testing.T) {
	// Arrange
	service, mockRepo := setupBranchTest(t)
	ctx := context.Background()
	companyID := int64(1)

	isActive := true
	req := dto.ListBranchRequest{
		Page:     1,
		PerPage:  10,
		Search:   "Main",
		IsActive: &isActive,
	}

	expectedBranches := []*model.Branch{
		createTestBranch(1, companyID, "BR001"),
	}
	totalRecords := 1

	mockRepo.EXPECT().
		List(ctx, companyID, "Main", &isActive, 10, 0).
		Return(expectedBranches, totalRecords, nil).
		Once()

	// Act
	result, err := service.List(ctx, companyID, req)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Branches, 1)
	assert.Equal(t, "BR001", result.Branches[0].Code)
	mockRepo.AssertExpectations(t)
}

func TestBranchService_List_Pagination(t *testing.T) {
	tests := []struct {
		name           string
		page           int
		perPage        int
		totalRecords   int
		expectedPage   int
		expectedPerPage int
		expectedOffset int
		hasNext        bool
		hasPrev        bool
	}{
		{
			name:           "First page with results",
			page:           1,
			perPage:        10,
			totalRecords:   25,
			expectedPage:   1,
			expectedPerPage: 10,
			expectedOffset: 0,
			hasNext:        true,
			hasPrev:        false,
		},
		{
			name:           "Second page",
			page:           2,
			perPage:        10,
			totalRecords:   25,
			expectedPage:   2,
			expectedPerPage: 10,
			expectedOffset: 10,
			hasNext:        true,
			hasPrev:        true,
		},
		{
			name:           "Last page",
			page:           3,
			perPage:        10,
			totalRecords:   25,
			expectedPage:   3,
			expectedPerPage: 10,
			expectedOffset: 20,
			hasNext:        false,
			hasPrev:        true,
		},
		{
			name:           "Default page (invalid 0)",
			page:           0,
			perPage:        20,
			totalRecords:   10,
			expectedPage:   1,
			expectedPerPage: 20,
			expectedOffset: 0,
			hasNext:        false,
			hasPrev:        false,
		},
		{
			name:           "Default perPage (invalid 0)",
			page:           1,
			perPage:        0,
			totalRecords:   10,
			expectedPage:   1,
			expectedPerPage: 20,
			expectedOffset: 0,
			hasNext:        false,
			hasPrev:        false,
		},
		{
			name:           "Max perPage (over 100)",
			page:           1,
			perPage:        150,
			totalRecords:   200,
			expectedPage:   1,
			expectedPerPage: 100,
			expectedOffset: 0,
			hasNext:        true,
			hasPrev:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			service, mockRepo := setupBranchTest(t)
			ctx := context.Background()
			companyID := int64(1)

			req := dto.ListBranchRequest{
				Page:    tt.page,
				PerPage: tt.perPage,
			}

			mockRepo.EXPECT().
				List(ctx, companyID, "", (*bool)(nil), tt.expectedPerPage, tt.expectedOffset).
				Return([]*model.Branch{}, tt.totalRecords, nil).
				Once()

			// Act
			result, err := service.List(ctx, companyID, req)

			// Assert
			require.NoError(t, err)
			assert.Equal(t, tt.expectedPage, result.Pagination.CurrentPage)
			assert.Equal(t, tt.expectedPerPage, result.Pagination.PerPage)

			if tt.hasNext {
				assert.NotNil(t, result.Pagination.NextPage)
				assert.Equal(t, tt.expectedPage+1, *result.Pagination.NextPage)
			} else {
				assert.Nil(t, result.Pagination.NextPage)
			}

			if tt.hasPrev {
				assert.NotNil(t, result.Pagination.PrevPage)
				assert.Equal(t, tt.expectedPage-1, *result.Pagination.PrevPage)
			} else {
				assert.Nil(t, result.Pagination.PrevPage)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestBranchService_List_RepositoryError(t *testing.T) {
	// Arrange
	service, mockRepo := setupBranchTest(t)
	ctx := context.Background()
	companyID := int64(1)

	req := dto.ListBranchRequest{
		Page:    1,
		PerPage: 20,
	}

	dbError := errors.New("database query error")
	mockRepo.EXPECT().
		List(ctx, companyID, "", (*bool)(nil), 20, 0).
		Return(nil, 0, dbError).
		Once()

	// Act
	result, err := service.List(ctx, companyID, req)

	// Assert
	require.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, apperror.IsInternalError(err))
	mockRepo.AssertExpectations(t)
}

// =======================
// Update Tests
// =======================

func TestBranchService_Update_Success(t *testing.T) {
	// Arrange
	service, mockRepo := setupBranchTest(t)
	ctx := context.Background()
	companyID := int64(1)
	branchID := int64(10)

	existingBranch := createTestBranch(branchID, companyID, "BR001")

	req := dto.UpdateBranchRequest{
		Code:     "BR002",
		Name:     "Updated Branch",
		Address:  strPtr("456 New Street"),
		Phone:    strPtr("+9876543210"),
		IsActive: false,
	}

	mockRepo.EXPECT().
		GetByID(ctx, branchID).
		Return(existingBranch, nil).
		Once()

	mockRepo.EXPECT().
		CodeExists(ctx, companyID, req.Code, branchID).
		Return(false, nil).
		Once()

	mockRepo.EXPECT().
		Update(ctx, mock.MatchedBy(func(b *model.Branch) bool {
			return b.ID == branchID &&
				b.CompanyID == companyID &&
				b.Code == req.Code &&
				b.Name == req.Name &&
				*b.Address == *req.Address &&
				*b.Phone == *req.Phone &&
				b.IsActive == req.IsActive
		})).
		Return(nil).
		Once()

	// Act
	result, err := service.Update(ctx, companyID, branchID, req)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, branchID, result.ID)
	assert.Equal(t, req.Code, result.Code)
	assert.Equal(t, req.Name, result.Name)
	assert.Equal(t, req.Address, result.Address)
	assert.Equal(t, req.Phone, result.Phone)
	assert.Equal(t, req.IsActive, result.IsActive)
	mockRepo.AssertExpectations(t)
}

func TestBranchService_Update_NotFound(t *testing.T) {
	// Arrange
	service, mockRepo := setupBranchTest(t)
	ctx := context.Background()
	companyID := int64(1)
	branchID := int64(999)

	req := dto.UpdateBranchRequest{
		Code:     "BR002",
		Name:     "Updated Branch",
		Address:  strPtr("456 New Street"),
		Phone:    strPtr("+9876543210"),
		IsActive: true,
	}

	mockRepo.EXPECT().
		GetByID(ctx, branchID).
		Return(nil, nil).
		Once()

	// Act
	result, err := service.Update(ctx, companyID, branchID, req)

	// Assert
	require.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, apperror.IsNotFound(err))
	assert.Contains(t, err.Error(), "Branch not found")
	mockRepo.AssertExpectations(t)
}

func TestBranchService_Update_WrongCompany(t *testing.T) {
	// Arrange
	service, mockRepo := setupBranchTest(t)
	ctx := context.Background()
	requestedCompanyID := int64(1)
	branchID := int64(10)

	// Branch belongs to company 2
	branchFromDifferentCompany := createTestBranch(branchID, int64(2), "BR001")

	req := dto.UpdateBranchRequest{
		Code:     "BR002",
		Name:     "Updated Branch",
		Address:  strPtr("456 New Street"),
		Phone:    strPtr("+9876543210"),
		IsActive: true,
	}

	mockRepo.EXPECT().
		GetByID(ctx, branchID).
		Return(branchFromDifferentCompany, nil).
		Once()

	// Act
	result, err := service.Update(ctx, requestedCompanyID, branchID, req)

	// Assert
	require.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, apperror.IsNotFound(err))
	assert.Contains(t, err.Error(), "Branch not found")
	mockRepo.AssertExpectations(t)
}

func TestBranchService_Update_CodeExists(t *testing.T) {
	// Arrange
	service, mockRepo := setupBranchTest(t)
	ctx := context.Background()
	companyID := int64(1)
	branchID := int64(10)

	existingBranch := createTestBranch(branchID, companyID, "BR001")

	req := dto.UpdateBranchRequest{
		Code:     "BR002", // This code already exists for another branch
		Name:     "Updated Branch",
		Address:  strPtr("456 New Street"),
		Phone:    strPtr("+9876543210"),
		IsActive: true,
	}

	mockRepo.EXPECT().
		GetByID(ctx, branchID).
		Return(existingBranch, nil).
		Once()

	mockRepo.EXPECT().
		CodeExists(ctx, companyID, req.Code, branchID).
		Return(true, nil).
		Once()

	// Act
	result, err := service.Update(ctx, companyID, branchID, req)

	// Assert
	require.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, apperror.IsBadRequest(err))
	assert.Contains(t, err.Error(), "Branch code already exists")
	mockRepo.AssertExpectations(t)
}

func TestBranchService_Update_GetByIDError(t *testing.T) {
	// Arrange
	service, mockRepo := setupBranchTest(t)
	ctx := context.Background()
	companyID := int64(1)
	branchID := int64(10)

	req := dto.UpdateBranchRequest{
		Code:     "BR002",
		Name:     "Updated Branch",
		Address:  strPtr("456 New Street"),
		Phone:    strPtr("+9876543210"),
		IsActive: true,
	}

	dbError := errors.New("database query error")
	mockRepo.EXPECT().
		GetByID(ctx, branchID).
		Return(nil, dbError).
		Once()

	// Act
	result, err := service.Update(ctx, companyID, branchID, req)

	// Assert
	require.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, apperror.IsInternalError(err))
	mockRepo.AssertExpectations(t)
}

func TestBranchService_Update_CodeExistsCheckError(t *testing.T) {
	// Arrange
	service, mockRepo := setupBranchTest(t)
	ctx := context.Background()
	companyID := int64(1)
	branchID := int64(10)

	existingBranch := createTestBranch(branchID, companyID, "BR001")

	req := dto.UpdateBranchRequest{
		Code:     "BR002",
		Name:     "Updated Branch",
		Address:  strPtr("456 New Street"),
		Phone:    strPtr("+9876543210"),
		IsActive: true,
	}

	mockRepo.EXPECT().
		GetByID(ctx, branchID).
		Return(existingBranch, nil).
		Once()

	dbError := errors.New("database connection error")
	mockRepo.EXPECT().
		CodeExists(ctx, companyID, req.Code, branchID).
		Return(false, dbError).
		Once()

	// Act
	result, err := service.Update(ctx, companyID, branchID, req)

	// Assert
	require.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, apperror.IsInternalError(err))
	mockRepo.AssertExpectations(t)
}

func TestBranchService_Update_RepositoryError(t *testing.T) {
	// Arrange
	service, mockRepo := setupBranchTest(t)
	ctx := context.Background()
	companyID := int64(1)
	branchID := int64(10)

	existingBranch := createTestBranch(branchID, companyID, "BR001")

	req := dto.UpdateBranchRequest{
		Code:     "BR002",
		Name:     "Updated Branch",
		Address:  strPtr("456 New Street"),
		Phone:    strPtr("+9876543210"),
		IsActive: true,
	}

	mockRepo.EXPECT().
		GetByID(ctx, branchID).
		Return(existingBranch, nil).
		Once()

	mockRepo.EXPECT().
		CodeExists(ctx, companyID, req.Code, branchID).
		Return(false, nil).
		Once()

	dbError := errors.New("database update error")
	mockRepo.EXPECT().
		Update(ctx, mock.AnythingOfType("*model.Branch")).
		Return(dbError).
		Once()

	// Act
	result, err := service.Update(ctx, companyID, branchID, req)

	// Assert
	require.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, apperror.IsInternalError(err))
	mockRepo.AssertExpectations(t)
}

// =======================
// Delete Tests
// =======================

func TestBranchService_Delete_Success(t *testing.T) {
	// Arrange
	service, mockRepo := setupBranchTest(t)
	ctx := context.Background()
	companyID := int64(1)
	branchID := int64(10)

	existingBranch := createTestBranch(branchID, companyID, "BR001")

	mockRepo.EXPECT().
		GetByID(ctx, branchID).
		Return(existingBranch, nil).
		Once()

	mockRepo.EXPECT().
		Delete(ctx, branchID).
		Return(nil).
		Once()

	// Act
	err := service.Delete(ctx, companyID, branchID)

	// Assert
	require.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestBranchService_Delete_NotFound(t *testing.T) {
	// Arrange
	service, mockRepo := setupBranchTest(t)
	ctx := context.Background()
	companyID := int64(1)
	branchID := int64(999)

	mockRepo.EXPECT().
		GetByID(ctx, branchID).
		Return(nil, nil).
		Once()

	// Act
	err := service.Delete(ctx, companyID, branchID)

	// Assert
	require.Error(t, err)
	assert.True(t, apperror.IsNotFound(err))
	assert.Contains(t, err.Error(), "Branch not found")
	mockRepo.AssertExpectations(t)
}

func TestBranchService_Delete_WrongCompany(t *testing.T) {
	// Arrange
	service, mockRepo := setupBranchTest(t)
	ctx := context.Background()
	requestedCompanyID := int64(1)
	branchID := int64(10)

	// Branch belongs to company 2
	branchFromDifferentCompany := createTestBranch(branchID, int64(2), "BR001")

	mockRepo.EXPECT().
		GetByID(ctx, branchID).
		Return(branchFromDifferentCompany, nil).
		Once()

	// Act
	err := service.Delete(ctx, requestedCompanyID, branchID)

	// Assert
	require.Error(t, err)
	assert.True(t, apperror.IsNotFound(err))
	assert.Contains(t, err.Error(), "Branch not found")
	mockRepo.AssertExpectations(t)
}

func TestBranchService_Delete_GetByIDError(t *testing.T) {
	// Arrange
	service, mockRepo := setupBranchTest(t)
	ctx := context.Background()
	companyID := int64(1)
	branchID := int64(10)

	dbError := errors.New("database query error")
	mockRepo.EXPECT().
		GetByID(ctx, branchID).
		Return(nil, dbError).
		Once()

	// Act
	err := service.Delete(ctx, companyID, branchID)

	// Assert
	require.Error(t, err)
	assert.True(t, apperror.IsInternalError(err))
	mockRepo.AssertExpectations(t)
}

func TestBranchService_Delete_RepositoryError(t *testing.T) {
	// Arrange
	service, mockRepo := setupBranchTest(t)
	ctx := context.Background()
	companyID := int64(1)
	branchID := int64(10)

	existingBranch := createTestBranch(branchID, companyID, "BR001")

	mockRepo.EXPECT().
		GetByID(ctx, branchID).
		Return(existingBranch, nil).
		Once()

	dbError := errors.New("database delete error")
	mockRepo.EXPECT().
		Delete(ctx, branchID).
		Return(dbError).
		Once()

	// Act
	err := service.Delete(ctx, companyID, branchID)

	// Assert
	require.Error(t, err)
	assert.True(t, apperror.IsInternalError(err))
	mockRepo.AssertExpectations(t)
}

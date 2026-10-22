package mocks

import (
	"context"

	"github.com/irvanmhndra/nexpos-api/internal/dto"
	"github.com/stretchr/testify/mock"
)

// MockBranchService is a mock implementation of service.BranchServiceInterface
type MockBranchService struct {
	mock.Mock
}

func (m *MockBranchService) Create(ctx context.Context, companyID int64, req dto.CreateBranchRequest) (*dto.BranchResponse, error) {
	args := m.Called(ctx, companyID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.BranchResponse), args.Error(1)
}

func (m *MockBranchService) GetByID(ctx context.Context, companyID, id int64) (*dto.BranchResponse, error) {
	args := m.Called(ctx, companyID, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.BranchResponse), args.Error(1)
}

func (m *MockBranchService) List(ctx context.Context, companyID int64, req dto.ListBranchRequest) (*dto.BranchListResponse, error) {
	args := m.Called(ctx, companyID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.BranchListResponse), args.Error(1)
}

func (m *MockBranchService) Update(ctx context.Context, companyID, id int64, req dto.UpdateBranchRequest) (*dto.BranchResponse, error) {
	args := m.Called(ctx, companyID, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.BranchResponse), args.Error(1)
}

func (m *MockBranchService) Delete(ctx context.Context, companyID, id int64) error {
	args := m.Called(ctx, companyID, id)
	return args.Error(0)
}

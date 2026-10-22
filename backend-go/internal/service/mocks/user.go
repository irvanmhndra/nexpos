package mocks

import (
	"context"

	"github.com/irvanmhndra/nexpos-api/internal/dto"
	"github.com/irvanmhndra/nexpos-api/internal/service"
	"github.com/stretchr/testify/mock"
)

// MockUserService is a mock implementation of service.UserServiceInterface
type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) Create(ctx context.Context, companyID int64, req dto.CreateUserRequest) (*dto.UserResponse, error) {
	args := m.Called(ctx, companyID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.UserResponse), args.Error(1)
}

func (m *MockUserService) GetByID(ctx context.Context, companyID, id int64) (*dto.UserResponse, error) {
	args := m.Called(ctx, companyID, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.UserResponse), args.Error(1)
}

func (m *MockUserService) List(ctx context.Context, companyID int64, req dto.ListUsersRequest) (*service.UserListResult, error) {
	args := m.Called(ctx, companyID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.UserListResult), args.Error(1)
}

func (m *MockUserService) Update(ctx context.Context, companyID, id int64, req dto.UpdateUserRequest) (*dto.UserResponse, error) {
	args := m.Called(ctx, companyID, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.UserResponse), args.Error(1)
}

func (m *MockUserService) UpdateStatus(ctx context.Context, companyID, id int64, req dto.UpdateUserStatusRequest) error {
	args := m.Called(ctx, companyID, id, req)
	return args.Error(0)
}

func (m *MockUserService) Delete(ctx context.Context, companyID, id int64) error {
	args := m.Called(ctx, companyID, id)
	return args.Error(0)
}

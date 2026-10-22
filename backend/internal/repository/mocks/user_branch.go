package mocks

import (
	"context"

	"github.com/irvanmhndra/nexpos-api/internal/model"
	"github.com/stretchr/testify/mock"
)

type MockUserBranchRepository struct {
	mock.Mock
}

func NewMockUserBranchRepository(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockUserBranchRepository {
	m := &MockUserBranchRepository{}
	m.Test(t)
	t.Cleanup(func() { m.AssertExpectations(t) })
	return m
}

func (m *MockUserBranchRepository) EXPECT() *MockUserBranchRepositoryExpectation {
	return &MockUserBranchRepositoryExpectation{mock: m}
}

func (m *MockUserBranchRepository) Create(ctx context.Context, ub *model.UserBranch) error {
	ret := m.Called(ctx, ub)
	return ret.Error(0)
}

func (m *MockUserBranchRepository) Delete(ctx context.Context, userID, branchID int64) error {
	ret := m.Called(ctx, userID, branchID)
	return ret.Error(0)
}

func (m *MockUserBranchRepository) SetBranches(ctx context.Context, userID int64, branchIDs []int64, defaultBranchID int64) error {
	ret := m.Called(ctx, userID, branchIDs, defaultBranchID)
	return ret.Error(0)
}

func (m *MockUserBranchRepository) GetByUserID(ctx context.Context, userID int64) ([]*model.UserBranch, error) {
	ret := m.Called(ctx, userID)
	var r0 []*model.UserBranch
	if ret.Get(0) != nil {
		r0 = ret.Get(0).([]*model.UserBranch)
	}
	return r0, ret.Error(1)
}

func (m *MockUserBranchRepository) GetDefaultBranch(ctx context.Context, userID int64) (*model.Branch, error) {
	ret := m.Called(ctx, userID)
	var r0 *model.Branch
	if ret.Get(0) != nil {
		r0 = ret.Get(0).(*model.Branch)
	}
	return r0, ret.Error(1)
}

type MockUserBranchRepositoryExpectation struct {
	mock *MockUserBranchRepository
}

func (e *MockUserBranchRepositoryExpectation) Create(ctx context.Context, ub interface{}) *mock.Call {
	return e.mock.On("Create", ctx, ub)
}

func (e *MockUserBranchRepositoryExpectation) Delete(ctx context.Context, userID, branchID int64) *mock.Call {
	return e.mock.On("Delete", ctx, userID, branchID)
}

func (e *MockUserBranchRepositoryExpectation) SetBranches(ctx context.Context, userID int64, branchIDs []int64, defaultBranchID int64) *mock.Call {
	return e.mock.On("SetBranches", ctx, userID, branchIDs, defaultBranchID)
}

func (e *MockUserBranchRepositoryExpectation) GetByUserID(ctx context.Context, userID int64) *mock.Call {
	return e.mock.On("GetByUserID", ctx, userID)
}

func (e *MockUserBranchRepositoryExpectation) GetDefaultBranch(ctx context.Context, userID int64) *mock.Call {
	return e.mock.On("GetDefaultBranch", ctx, userID)
}

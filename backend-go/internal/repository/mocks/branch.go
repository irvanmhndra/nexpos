package mocks

import (
	"context"

	"github.com/irvanmhndra/nexpos-api/internal/model"
	"github.com/stretchr/testify/mock"
)

// MockBranchRepository is a mock implementation of repository.BranchRepository.
type MockBranchRepository struct {
	mock.Mock
}

func NewMockBranchRepository(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockBranchRepository {
	m := &MockBranchRepository{}
	m.Test(t)
	t.Cleanup(func() { m.AssertExpectations(t) })
	return m
}

func (m *MockBranchRepository) EXPECT() *MockBranchRepositoryExpectation {
	return &MockBranchRepositoryExpectation{mock: m}
}

func (m *MockBranchRepository) Create(ctx context.Context, branch *model.Branch) error {
	ret := m.Called(ctx, branch)
	return ret.Error(0)
}

func (m *MockBranchRepository) GetByID(ctx context.Context, id int64) (*model.Branch, error) {
	ret := m.Called(ctx, id)
	var r0 *model.Branch
	if ret.Get(0) != nil {
		r0 = ret.Get(0).(*model.Branch)
	}
	return r0, ret.Error(1)
}

func (m *MockBranchRepository) ListByCompanyID(ctx context.Context, companyID int64) ([]*model.Branch, error) {
	ret := m.Called(ctx, companyID)
	var r0 []*model.Branch
	if ret.Get(0) != nil {
		r0 = ret.Get(0).([]*model.Branch)
	}
	return r0, ret.Error(1)
}

func (m *MockBranchRepository) List(ctx context.Context, companyID int64, search string, isActive *bool, limit, offset int) ([]*model.Branch, int, error) {
	ret := m.Called(ctx, companyID, search, isActive, limit, offset)
	var r0 []*model.Branch
	if ret.Get(0) != nil {
		r0 = ret.Get(0).([]*model.Branch)
	}
	return r0, ret.Int(1), ret.Error(2)
}

func (m *MockBranchRepository) Update(ctx context.Context, branch *model.Branch) error {
	ret := m.Called(ctx, branch)
	return ret.Error(0)
}

func (m *MockBranchRepository) Delete(ctx context.Context, id int64) error {
	ret := m.Called(ctx, id)
	return ret.Error(0)
}

func (m *MockBranchRepository) CodeExists(ctx context.Context, companyID int64, code string, excludeID int64) (bool, error) {
	ret := m.Called(ctx, companyID, code, excludeID)
	return ret.Bool(0), ret.Error(1)
}

// MockBranchRepositoryExpectation is the expectation builder for MockBranchRepository.
type MockBranchRepositoryExpectation struct {
	mock *MockBranchRepository
}

func (e *MockBranchRepositoryExpectation) Create(ctx context.Context, branch interface{}) *mock.Call {
	return e.mock.On("Create", ctx, branch)
}

func (e *MockBranchRepositoryExpectation) GetByID(ctx context.Context, id int64) *mock.Call {
	return e.mock.On("GetByID", ctx, id)
}

func (e *MockBranchRepositoryExpectation) ListByCompanyID(ctx context.Context, companyID int64) *mock.Call {
	return e.mock.On("ListByCompanyID", ctx, companyID)
}

func (e *MockBranchRepositoryExpectation) List(ctx context.Context, companyID int64, search string, isActive *bool, limit, offset int) *mock.Call {
	return e.mock.On("List", ctx, companyID, search, isActive, limit, offset)
}

func (e *MockBranchRepositoryExpectation) Update(ctx context.Context, branch interface{}) *mock.Call {
	return e.mock.On("Update", ctx, branch)
}

func (e *MockBranchRepositoryExpectation) Delete(ctx context.Context, id int64) *mock.Call {
	return e.mock.On("Delete", ctx, id)
}

func (e *MockBranchRepositoryExpectation) CodeExists(ctx context.Context, companyID int64, code string, excludeID int64) *mock.Call {
	return e.mock.On("CodeExists", ctx, companyID, code, excludeID)
}

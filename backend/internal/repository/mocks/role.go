package mocks

import (
	"context"

	"github.com/irvanmhndra/nexpos-api/internal/model"
	"github.com/stretchr/testify/mock"
)

type MockRoleRepository struct {
	mock.Mock
}

func NewMockRoleRepository(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockRoleRepository {
	m := &MockRoleRepository{}
	m.Test(t)
	t.Cleanup(func() { m.AssertExpectations(t) })
	return m
}

func (m *MockRoleRepository) EXPECT() *MockRoleRepositoryExpectation {
	return &MockRoleRepositoryExpectation{mock: m}
}

func (m *MockRoleRepository) GetByID(ctx context.Context, id int64) (*model.Role, error) {
	ret := m.Called(ctx, id)
	var r0 *model.Role
	if ret.Get(0) != nil {
		r0 = ret.Get(0).(*model.Role)
	}
	return r0, ret.Error(1)
}

func (m *MockRoleRepository) GetByCode(ctx context.Context, code string, companyID *int64) (*model.Role, error) {
	ret := m.Called(ctx, code, companyID)
	var r0 *model.Role
	if ret.Get(0) != nil {
		r0 = ret.Get(0).(*model.Role)
	}
	return r0, ret.Error(1)
}

func (m *MockRoleRepository) ListByCompanyID(ctx context.Context, companyID int64) ([]*model.Role, error) {
	ret := m.Called(ctx, companyID)
	var r0 []*model.Role
	if ret.Get(0) != nil {
		r0 = ret.Get(0).([]*model.Role)
	}
	return r0, ret.Error(1)
}

func (m *MockRoleRepository) ListSystemRoles(ctx context.Context) ([]*model.Role, error) {
	ret := m.Called(ctx)
	var r0 []*model.Role
	if ret.Get(0) != nil {
		r0 = ret.Get(0).([]*model.Role)
	}
	return r0, ret.Error(1)
}

func (m *MockRoleRepository) Create(ctx context.Context, role *model.Role) error {
	ret := m.Called(ctx, role)
	return ret.Error(0)
}

func (m *MockRoleRepository) Update(ctx context.Context, role *model.Role) error {
	ret := m.Called(ctx, role)
	return ret.Error(0)
}

func (m *MockRoleRepository) Delete(ctx context.Context, id int64) error {
	ret := m.Called(ctx, id)
	return ret.Error(0)
}

func (m *MockRoleRepository) GetWithPermissions(ctx context.Context, id int64) (*model.Role, error) {
	ret := m.Called(ctx, id)
	var r0 *model.Role
	if ret.Get(0) != nil {
		r0 = ret.Get(0).(*model.Role)
	}
	return r0, ret.Error(1)
}

type MockRoleRepositoryExpectation struct {
	mock *MockRoleRepository
}

func (e *MockRoleRepositoryExpectation) GetByID(ctx context.Context, id int64) *mock.Call {
	return e.mock.On("GetByID", ctx, id)
}

func (e *MockRoleRepositoryExpectation) GetByCode(ctx context.Context, code string, companyID *int64) *mock.Call {
	return e.mock.On("GetByCode", ctx, code, companyID)
}

func (e *MockRoleRepositoryExpectation) ListByCompanyID(ctx context.Context, companyID int64) *mock.Call {
	return e.mock.On("ListByCompanyID", ctx, companyID)
}

func (e *MockRoleRepositoryExpectation) ListSystemRoles(ctx context.Context) *mock.Call {
	return e.mock.On("ListSystemRoles", ctx)
}

func (e *MockRoleRepositoryExpectation) Create(ctx context.Context, role interface{}) *mock.Call {
	return e.mock.On("Create", ctx, role)
}

func (e *MockRoleRepositoryExpectation) Update(ctx context.Context, role interface{}) *mock.Call {
	return e.mock.On("Update", ctx, role)
}

func (e *MockRoleRepositoryExpectation) Delete(ctx context.Context, id int64) *mock.Call {
	return e.mock.On("Delete", ctx, id)
}

func (e *MockRoleRepositoryExpectation) GetWithPermissions(ctx context.Context, id int64) *mock.Call {
	return e.mock.On("GetWithPermissions", ctx, id)
}

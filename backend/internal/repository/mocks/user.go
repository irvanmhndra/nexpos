package mocks

import (
	"context"

	"github.com/irvanmhndra/nexpos-api/internal/model"
	"github.com/stretchr/testify/mock"
)

type MockUserRepository struct {
	mock.Mock
}

func NewMockUserRepository(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockUserRepository {
	m := &MockUserRepository{}
	m.Test(t)
	t.Cleanup(func() { m.AssertExpectations(t) })
	return m
}

func (m *MockUserRepository) EXPECT() *MockUserRepositoryExpectation {
	return &MockUserRepositoryExpectation{mock: m}
}

func (m *MockUserRepository) Create(ctx context.Context, user *model.User) error {
	ret := m.Called(ctx, user)
	return ret.Error(0)
}

func (m *MockUserRepository) GetByID(ctx context.Context, companyID, id int64) (*model.User, error) {
	ret := m.Called(ctx, companyID, id)
	var r0 *model.User
	if ret.Get(0) != nil {
		r0 = ret.Get(0).(*model.User)
	}
	return r0, ret.Error(1)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	ret := m.Called(ctx, email)
	var r0 *model.User
	if ret.Get(0) != nil {
		r0 = ret.Get(0).(*model.User)
	}
	return r0, ret.Error(1)
}

func (m *MockUserRepository) List(ctx context.Context, companyID int64, limit, offset int) ([]*model.User, int, error) {
	ret := m.Called(ctx, companyID, limit, offset)
	var r0 []*model.User
	if ret.Get(0) != nil {
		r0 = ret.Get(0).([]*model.User)
	}
	return r0, ret.Int(1), ret.Error(2)
}

func (m *MockUserRepository) Update(ctx context.Context, user *model.User) error {
	ret := m.Called(ctx, user)
	return ret.Error(0)
}

func (m *MockUserRepository) UpdateStatus(ctx context.Context, companyID, id int64, status model.UserStatus) error {
	ret := m.Called(ctx, companyID, id, status)
	return ret.Error(0)
}

func (m *MockUserRepository) Delete(ctx context.Context, companyID, id int64) error {
	ret := m.Called(ctx, companyID, id)
	return ret.Error(0)
}

func (m *MockUserRepository) EmailExists(ctx context.Context, email string, excludeID int64) (bool, error) {
	ret := m.Called(ctx, email, excludeID)
	return ret.Bool(0), ret.Error(1)
}

type MockUserRepositoryExpectation struct {
	mock *MockUserRepository
}

func (e *MockUserRepositoryExpectation) Create(ctx context.Context, user interface{}) *mock.Call {
	return e.mock.On("Create", ctx, user)
}

func (e *MockUserRepositoryExpectation) GetByID(ctx context.Context, companyID, id int64) *mock.Call {
	return e.mock.On("GetByID", ctx, companyID, id)
}

func (e *MockUserRepositoryExpectation) GetByEmail(ctx context.Context, email string) *mock.Call {
	return e.mock.On("GetByEmail", ctx, email)
}

func (e *MockUserRepositoryExpectation) List(ctx context.Context, companyID int64, limit, offset int) *mock.Call {
	return e.mock.On("List", ctx, companyID, limit, offset)
}

func (e *MockUserRepositoryExpectation) Update(ctx context.Context, user interface{}) *mock.Call {
	return e.mock.On("Update", ctx, user)
}

func (e *MockUserRepositoryExpectation) UpdateStatus(ctx context.Context, companyID, id int64, status model.UserStatus) *mock.Call {
	return e.mock.On("UpdateStatus", ctx, companyID, id, status)
}

func (e *MockUserRepositoryExpectation) Delete(ctx context.Context, companyID, id int64) *mock.Call {
	return e.mock.On("Delete", ctx, companyID, id)
}

func (e *MockUserRepositoryExpectation) EmailExists(ctx context.Context, email string, excludeID int64) *mock.Call {
	return e.mock.On("EmailExists", ctx, email, excludeID)
}

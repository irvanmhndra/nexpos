package mocks

import (
	"context"

	"github.com/irvanmhndra/nexpos-api/internal/model"
	"github.com/stretchr/testify/mock"
)

type MockUserSessionRepository struct {
	mock.Mock
}

func NewMockUserSessionRepository(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockUserSessionRepository {
	m := &MockUserSessionRepository{}
	m.Test(t)
	t.Cleanup(func() { m.AssertExpectations(t) })
	return m
}

func (m *MockUserSessionRepository) EXPECT() *MockUserSessionRepositoryExpectation {
	return &MockUserSessionRepositoryExpectation{mock: m}
}

func (m *MockUserSessionRepository) Create(ctx context.Context, session *model.UserSession) error {
	ret := m.Called(ctx, session)
	return ret.Error(0)
}

func (m *MockUserSessionRepository) GetByAccessToken(ctx context.Context, accessToken string) (*model.UserSession, error) {
	ret := m.Called(ctx, accessToken)
	var r0 *model.UserSession
	if ret.Get(0) != nil {
		r0 = ret.Get(0).(*model.UserSession)
	}
	return r0, ret.Error(1)
}

func (m *MockUserSessionRepository) GetByRefreshToken(ctx context.Context, refreshToken string) (*model.UserSession, error) {
	ret := m.Called(ctx, refreshToken)
	var r0 *model.UserSession
	if ret.Get(0) != nil {
		r0 = ret.Get(0).(*model.UserSession)
	}
	return r0, ret.Error(1)
}

func (m *MockUserSessionRepository) UpdateLastUsed(ctx context.Context, id int64) error {
	ret := m.Called(ctx, id)
	return ret.Error(0)
}

func (m *MockUserSessionRepository) Revoke(ctx context.Context, id int64) error {
	ret := m.Called(ctx, id)
	return ret.Error(0)
}

func (m *MockUserSessionRepository) RevokeAllByUserID(ctx context.Context, userID int64) error {
	ret := m.Called(ctx, userID)
	return ret.Error(0)
}

type MockUserSessionRepositoryExpectation struct {
	mock *MockUserSessionRepository
}

func (e *MockUserSessionRepositoryExpectation) Create(ctx context.Context, session interface{}) *mock.Call {
	return e.mock.On("Create", ctx, session)
}

func (e *MockUserSessionRepositoryExpectation) GetByAccessToken(ctx context.Context, accessToken string) *mock.Call {
	return e.mock.On("GetByAccessToken", ctx, accessToken)
}

func (e *MockUserSessionRepositoryExpectation) GetByRefreshToken(ctx context.Context, refreshToken string) *mock.Call {
	return e.mock.On("GetByRefreshToken", ctx, refreshToken)
}

func (e *MockUserSessionRepositoryExpectation) UpdateLastUsed(ctx context.Context, id int64) *mock.Call {
	return e.mock.On("UpdateLastUsed", ctx, id)
}

func (e *MockUserSessionRepositoryExpectation) Revoke(ctx context.Context, id int64) *mock.Call {
	return e.mock.On("Revoke", ctx, id)
}

func (e *MockUserSessionRepositoryExpectation) RevokeAllByUserID(ctx context.Context, userID int64) *mock.Call {
	return e.mock.On("RevokeAllByUserID", ctx, userID)
}

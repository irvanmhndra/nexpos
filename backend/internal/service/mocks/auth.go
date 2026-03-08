package mocks

import (
	"context"

	"github.com/irvanmhndra/nexpos-api/internal/dto"
	"github.com/irvanmhndra/nexpos-api/internal/model"
	"github.com/stretchr/testify/mock"
)

// MockAuthService is a mock implementation of service.AuthServiceInterface
type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) Login(ctx context.Context, req dto.LoginRequest, ipAddress, userAgent string) (*dto.LoginResponse, error) {
	args := m.Called(ctx, req, ipAddress, userAgent)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.LoginResponse), args.Error(1)
}

func (m *MockAuthService) Register(ctx context.Context, req dto.RegisterRequest, ipAddress, userAgent string) (*dto.RegisterResponse, error) {
	args := m.Called(ctx, req, ipAddress, userAgent)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.RegisterResponse), args.Error(1)
}

func (m *MockAuthService) RefreshToken(ctx context.Context, refreshToken string) (*dto.RefreshTokenResponse, error) {
	args := m.Called(ctx, refreshToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.RefreshTokenResponse), args.Error(1)
}

func (m *MockAuthService) Logout(ctx context.Context, accessToken string) error {
	args := m.Called(ctx, accessToken)
	return args.Error(0)
}

func (m *MockAuthService) ValidateAccessToken(ctx context.Context, accessToken string) (*model.UserSession, error) {
	args := m.Called(ctx, accessToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.UserSession), args.Error(1)
}

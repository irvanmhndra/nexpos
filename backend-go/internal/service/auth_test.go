package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/irvanmhndra/nexpos-api/config"
	"github.com/irvanmhndra/nexpos-api/internal/dto"
	"github.com/irvanmhndra/nexpos-api/internal/model"
	repoMocks "github.com/irvanmhndra/nexpos-api/internal/repository/mocks"
	"github.com/irvanmhndra/nexpos-api/pkg/apperror"
	"github.com/irvanmhndra/nexpos-api/pkg/sessiontoken"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

type authTestSetup struct {
	svc            *AuthService
	userRepo       *repoMocks.MockUserRepository
	companyRepo    *repoMocks.MockCompanyRepository
	branchRepo     *repoMocks.MockBranchRepository
	sessionRepo    *repoMocks.MockUserSessionRepository
	roleRepo       *repoMocks.MockRoleRepository
	userBranchRepo *repoMocks.MockUserBranchRepository
}

func setupAuthTest(t *testing.T) *authTestSetup {
	t.Helper()
	s := &authTestSetup{
		userRepo:       repoMocks.NewMockUserRepository(t),
		companyRepo:    repoMocks.NewMockCompanyRepository(t),
		branchRepo:     repoMocks.NewMockBranchRepository(t),
		sessionRepo:    repoMocks.NewMockUserSessionRepository(t),
		roleRepo:       repoMocks.NewMockRoleRepository(t),
		userBranchRepo: repoMocks.NewMockUserBranchRepository(t),
	}
	s.svc = NewAuthService(
		s.userRepo, s.companyRepo, s.branchRepo,
		s.sessionRepo, s.roleRepo, s.userBranchRepo,
		config.JWTConfig{
			AccessTokenExpiry:  2 * time.Hour,
			RefreshTokenExpiry: 7 * 24 * time.Hour,
		},
	)
	return s
}

func hashPassword(password string) string {
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	return string(hash)
}

func createTestAuthUser(id, companyID int64, email string) *model.User {
	roleID := int64(1)
	return &model.User{
		ID:           id,
		CompanyID:    companyID,
		RoleID:       &roleID,
		Email:        email,
		PasswordHash: hashPassword("password123"),
		Name:         "Test User",
		Status:       model.UserStatusActive,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
}

func TestNewAuthService(t *testing.T) {
	s := setupAuthTest(t)
	assert.NotNil(t, s.svc)
}

// Login
func TestAuthService_Login_Success(t *testing.T) {
	s := setupAuthTest(t)
	ctx := context.Background()

	user := createTestAuthUser(1, 1, "test@example.com")
	company := &model.Company{ID: 1, Code: "TEST", Name: "Test Co"}
	branch := &model.Branch{ID: 1, Code: "HQ", Name: "HQ"}
	role := &model.Role{ID: 1, Code: "owner", Name: "Owner"}

	req := dto.LoginRequest{Email: "test@example.com", Password: "password123"}

	s.userRepo.EXPECT().GetByEmail(ctx, "test@example.com").Return(user, nil).Once()
	s.userBranchRepo.EXPECT().GetDefaultBranch(ctx, int64(1)).Return(branch, nil).Once()
	var stored *model.UserSession
	s.sessionRepo.EXPECT().Create(ctx, mock.MatchedBy(func(sess *model.UserSession) bool {
		stored = sess
		return true
	})).Return(nil).Once()
	s.companyRepo.EXPECT().GetByID(ctx, int64(1)).Return(company, nil).Once()
	s.roleRepo.EXPECT().GetByID(ctx, int64(1)).Return(role, nil).Once()

	result, err := s.svc.Login(ctx, req, "127.0.0.1", "test-agent")
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result.AccessToken)
	assert.NotEmpty(t, result.RefreshToken)
	// The database only ever sees digests of the tokens handed to the client
	assert.Equal(t, sessiontoken.Hash(result.AccessToken), stored.AccessTokenHash)
	assert.Equal(t, sessiontoken.Hash(result.RefreshToken), stored.RefreshTokenHash)
	assert.Equal(t, "test@example.com", result.User.Email)
	assert.Equal(t, "owner", result.User.Role)
}

func TestAuthService_Login_UserNotFound(t *testing.T) {
	s := setupAuthTest(t)
	ctx := context.Background()

	req := dto.LoginRequest{Email: "nobody@example.com", Password: "password123"}
	s.userRepo.EXPECT().GetByEmail(ctx, "nobody@example.com").Return(nil, nil).Once()

	result, err := s.svc.Login(ctx, req, "127.0.0.1", "test-agent")
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "invalid email or password")
}

func TestAuthService_Login_WrongPassword(t *testing.T) {
	s := setupAuthTest(t)
	ctx := context.Background()

	user := createTestAuthUser(1, 1, "test@example.com")
	req := dto.LoginRequest{Email: "test@example.com", Password: "wrongpassword"}
	s.userRepo.EXPECT().GetByEmail(ctx, "test@example.com").Return(user, nil).Once()

	result, err := s.svc.Login(ctx, req, "127.0.0.1", "test-agent")
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "invalid email or password")
}

func TestAuthService_Login_InactiveUser(t *testing.T) {
	s := setupAuthTest(t)
	ctx := context.Background()

	user := createTestAuthUser(1, 1, "test@example.com")
	user.Status = model.UserStatusInactive
	req := dto.LoginRequest{Email: "test@example.com", Password: "password123"}
	s.userRepo.EXPECT().GetByEmail(ctx, "test@example.com").Return(user, nil).Once()

	result, err := s.svc.Login(ctx, req, "127.0.0.1", "test-agent")
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "inactive")
}

func TestAuthService_Login_RepoError(t *testing.T) {
	s := setupAuthTest(t)
	ctx := context.Background()

	req := dto.LoginRequest{Email: "test@example.com", Password: "password123"}
	s.userRepo.EXPECT().GetByEmail(ctx, "test@example.com").Return(nil, errors.New("db error")).Once()

	result, err := s.svc.Login(ctx, req, "127.0.0.1", "test-agent")
	require.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, apperror.IsInternalError(err))
}

// Register
func TestAuthService_Register_Success(t *testing.T) {
	s := setupAuthTest(t)
	ctx := context.Background()

	ownerRole := &model.Role{ID: 1, Code: "owner", Name: "Owner"}
	req := dto.RegisterRequest{Name: "John", Email: "john@example.com", Password: "password123"}

	s.userRepo.EXPECT().GetByEmail(ctx, "john@example.com").Return(nil, nil).Once()
	s.roleRepo.EXPECT().GetByCode(ctx, "owner", (*int64)(nil)).Return(ownerRole, nil).Once()
	s.companyRepo.EXPECT().Create(ctx, mock.AnythingOfType("*model.Company")).Return(nil).
		Run(func(args mock.Arguments) {
			c := args.Get(1).(*model.Company)
			c.ID = 1
		}).Once()
	s.branchRepo.EXPECT().Create(ctx, mock.AnythingOfType("*model.Branch")).Return(nil).
		Run(func(args mock.Arguments) {
			b := args.Get(1).(*model.Branch)
			b.ID = 1
		}).Once()
	s.userRepo.EXPECT().Create(ctx, mock.AnythingOfType("*model.User")).Return(nil).
		Run(func(args mock.Arguments) {
			u := args.Get(1).(*model.User)
			u.ID = 1
		}).Once()
	s.userBranchRepo.EXPECT().Create(ctx, mock.AnythingOfType("*model.UserBranch")).Return(nil).Once()
	s.sessionRepo.EXPECT().Create(ctx, mock.AnythingOfType("*model.UserSession")).Return(nil).Once()

	result, err := s.svc.Register(ctx, req, "127.0.0.1", "test-agent")
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "john@example.com", result.User.Email)
	assert.NotEmpty(t, result.AccessToken)
}

func TestAuthService_Register_EmailExists(t *testing.T) {
	s := setupAuthTest(t)
	ctx := context.Background()

	existing := createTestAuthUser(1, 1, "existing@example.com")
	req := dto.RegisterRequest{Name: "John", Email: "existing@example.com", Password: "password123"}
	s.userRepo.EXPECT().GetByEmail(ctx, "existing@example.com").Return(existing, nil).Once()

	result, err := s.svc.Register(ctx, req, "127.0.0.1", "test-agent")
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "email already exists")
}

func TestAuthService_Register_OwnerRoleNotFound(t *testing.T) {
	s := setupAuthTest(t)
	ctx := context.Background()

	req := dto.RegisterRequest{Name: "John", Email: "john@example.com", Password: "password123"}
	s.userRepo.EXPECT().GetByEmail(ctx, "john@example.com").Return(nil, nil).Once()
	s.roleRepo.EXPECT().GetByCode(ctx, "owner", (*int64)(nil)).Return(nil, nil).Once()

	result, err := s.svc.Register(ctx, req, "127.0.0.1", "test-agent")
	require.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, apperror.IsInternalError(err))
}

// RefreshToken
func TestAuthService_RefreshToken_Success(t *testing.T) {
	s := setupAuthTest(t)
	ctx := context.Background()

	session := &model.UserSession{
		ID:                    1,
		UserID:                1,
		CompanyID:             1,
		RefreshTokenHash:      sessiontoken.Hash("valid_refresh"),
		RefreshTokenExpiresAt: time.Now().Add(24 * time.Hour), // not expired
	}

	var next *model.UserSession
	s.sessionRepo.EXPECT().GetByRefreshTokenHash(ctx, sessiontoken.Hash("valid_refresh")).Return(session, nil).Once()
	s.sessionRepo.EXPECT().Rotate(ctx, int64(1), mock.MatchedBy(func(n *model.UserSession) bool {
		next = n
		return true
	})).Return(true, nil).Once()

	result, err := s.svc.RefreshToken(ctx, "valid_refresh")
	require.NoError(t, err)
	assert.NotEmpty(t, result.AccessToken)
	assert.NotEmpty(t, result.RefreshToken)
	// Only digests of the new tokens are persisted
	assert.Equal(t, sessiontoken.Hash(result.AccessToken), next.AccessTokenHash)
	assert.Equal(t, sessiontoken.Hash(result.RefreshToken), next.RefreshTokenHash)
	assert.NotEqual(t, result.AccessToken, next.AccessTokenHash)
}

func TestAuthService_RefreshToken_AlreadyRotated(t *testing.T) {
	s := setupAuthTest(t)
	ctx := context.Background()

	session := &model.UserSession{
		ID:                    1,
		RefreshTokenExpiresAt: time.Now().Add(24 * time.Hour),
	}
	s.sessionRepo.EXPECT().GetByRefreshTokenHash(ctx, sessiontoken.Hash("raced")).Return(session, nil).Once()
	// A concurrent request exchanged the same refresh token first
	s.sessionRepo.EXPECT().Rotate(ctx, int64(1), mock.AnythingOfType("*model.UserSession")).Return(false, nil).Once()

	result, err := s.svc.RefreshToken(ctx, "raced")
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "invalid refresh token")
}

func TestAuthService_RefreshToken_InvalidToken(t *testing.T) {
	s := setupAuthTest(t)
	ctx := context.Background()

	s.sessionRepo.EXPECT().GetByRefreshTokenHash(ctx, sessiontoken.Hash("invalid")).Return(nil, nil).Once()

	result, err := s.svc.RefreshToken(ctx, "invalid")
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "invalid refresh token")
}

func TestAuthService_RefreshToken_Expired(t *testing.T) {
	s := setupAuthTest(t)
	ctx := context.Background()

	session := &model.UserSession{
		ID:                    1,
		RefreshTokenHash:      sessiontoken.Hash("expired"),
		RefreshTokenExpiresAt: time.Now().Add(-1 * time.Hour), // expired
	}
	s.sessionRepo.EXPECT().GetByRefreshTokenHash(ctx, sessiontoken.Hash("expired")).Return(session, nil).Once()

	result, err := s.svc.RefreshToken(ctx, "expired")
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "refresh token expired")
}

// Logout
func TestAuthService_Logout_Success(t *testing.T) {
	s := setupAuthTest(t)
	ctx := context.Background()

	session := &model.UserSession{ID: 1, AccessTokenHash: sessiontoken.Hash("valid_token")}
	s.sessionRepo.EXPECT().GetByAccessTokenHash(ctx, sessiontoken.Hash("valid_token")).Return(session, nil).Once()
	s.sessionRepo.EXPECT().Revoke(ctx, int64(1)).Return(nil).Once()

	err := s.svc.Logout(ctx, "valid_token")
	require.NoError(t, err)
}

func TestAuthService_Logout_AlreadyLoggedOut(t *testing.T) {
	s := setupAuthTest(t)
	ctx := context.Background()

	s.sessionRepo.EXPECT().GetByAccessTokenHash(ctx, sessiontoken.Hash("invalid_token")).Return(nil, nil).Once()

	err := s.svc.Logout(ctx, "invalid_token")
	require.NoError(t, err) // should not error
}

// ValidateAccessToken
func TestAuthService_ValidateAccessToken_Success(t *testing.T) {
	s := setupAuthTest(t)
	ctx := context.Background()

	session := &model.UserSession{
		ID:                   1,
		AccessTokenHash:      sessiontoken.Hash("valid"),
		AccessTokenExpiresAt: time.Now().Add(1 * time.Hour), // not expired
	}
	s.sessionRepo.EXPECT().GetByAccessTokenHash(ctx, sessiontoken.Hash("valid")).Return(session, nil).Once()
	s.sessionRepo.EXPECT().UpdateLastUsed(ctx, int64(1)).Return(nil).Once()

	result, err := s.svc.ValidateAccessToken(ctx, "valid")
	require.NoError(t, err)
	assert.NotNil(t, result)
}

func TestAuthService_ValidateAccessToken_Invalid(t *testing.T) {
	s := setupAuthTest(t)
	ctx := context.Background()

	s.sessionRepo.EXPECT().GetByAccessTokenHash(ctx, sessiontoken.Hash("invalid")).Return(nil, nil).Once()

	result, err := s.svc.ValidateAccessToken(ctx, "invalid")
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "invalid access token")
}

func TestAuthService_ValidateAccessToken_Expired(t *testing.T) {
	s := setupAuthTest(t)
	ctx := context.Background()

	session := &model.UserSession{
		ID:                   1,
		AccessTokenHash:      sessiontoken.Hash("expired"),
		AccessTokenExpiresAt: time.Now().Add(-1 * time.Hour), // expired
	}
	s.sessionRepo.EXPECT().GetByAccessTokenHash(ctx, sessiontoken.Hash("expired")).Return(session, nil).Once()

	result, err := s.svc.ValidateAccessToken(ctx, "expired")
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "access token expired")
}

package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/irvanmhndra/nexpos-api/config"
	"github.com/irvanmhndra/nexpos-api/internal/dto"
	"github.com/irvanmhndra/nexpos-api/internal/model"
	"github.com/irvanmhndra/nexpos-api/internal/repository"
	"github.com/irvanmhndra/nexpos-api/pkg/apperror"
	"github.com/irvanmhndra/nexpos-api/pkg/sessiontoken"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo       repository.UserRepository
	companyRepo    repository.CompanyRepository
	branchRepo     repository.BranchRepository
	sessionRepo    repository.UserSessionRepository
	roleRepo       repository.RoleRepository
	userBranchRepo repository.UserBranchRepository
	jwtConfig      config.JWTConfig
}

func NewAuthService(
	userRepo repository.UserRepository,
	companyRepo repository.CompanyRepository,
	branchRepo repository.BranchRepository,
	sessionRepo repository.UserSessionRepository,
	roleRepo repository.RoleRepository,
	userBranchRepo repository.UserBranchRepository,
	jwtConfig config.JWTConfig,
) *AuthService {
	return &AuthService{
		userRepo:       userRepo,
		companyRepo:    companyRepo,
		branchRepo:     branchRepo,
		sessionRepo:    sessionRepo,
		roleRepo:       roleRepo,
		userBranchRepo: userBranchRepo,
		jwtConfig:      jwtConfig,
	}
}

func (s *AuthService) Login(ctx context.Context, req dto.LoginRequest, ipAddress, userAgent string) (*dto.LoginResponse, error) {
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, apperror.InternalError(err)
	}
	if user == nil {
		return nil, apperror.InvalidCredentials()
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, apperror.InvalidCredentials()
	}

	if !user.IsActive() {
		return nil, apperror.UserInactive()
	}

	// Get default branch
	var defaultBranch *dto.BranchInfo
	var branchID *int64
	branch, branchErr := s.userBranchRepo.GetDefaultBranch(ctx, user.ID)
	if branchErr == nil && branch != nil {
		branchID = &branch.ID
		defaultBranch = &dto.BranchInfo{
			ID:   branch.ID,
			Code: branch.Code,
			Name: branch.Name,
		}
	}

	// Generate tokens
	accessToken := sessiontoken.New()
	refreshToken := sessiontoken.New()

	// Create session
	session := &model.UserSession{
		UserID:                user.ID,
		CompanyID:             user.CompanyID,
		BranchID:              branchID,
		AccessTokenHash:       sessiontoken.Hash(accessToken),
		AccessTokenExpiresAt:  time.Now().Add(s.jwtConfig.AccessTokenExpiry),
		RefreshTokenHash:      sessiontoken.Hash(refreshToken),
		RefreshTokenExpiresAt: time.Now().Add(s.jwtConfig.RefreshTokenExpiry),
		IPAddress:             &ipAddress,
		UserAgent:             &userAgent,
	}

	if err := s.sessionRepo.Create(ctx, session); err != nil {
		return nil, apperror.InternalError(err)
	}

	// Fetch company details
	company, err := s.companyRepo.GetByID(ctx, user.CompanyID)
	if err != nil {
		return nil, apperror.InternalError(err)
	}

	// Get role name
	roleName := "user"
	if user.RoleID != nil {
		role, err := s.roleRepo.GetByID(ctx, *user.RoleID)
		if err == nil && role != nil {
			roleName = role.Code
		}
	}

	return &dto.LoginResponse{
		User: &dto.AuthUserResponse{
			ID:    user.ID,
			Name:  user.Name,
			Email: user.Email,
			Role:  roleName,
			Company: &dto.CompanyInfo{
				ID:   company.ID,
				Code: company.Code,
				Name: company.Name,
			},
			DefaultBranch: defaultBranch,
		},
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int(s.jwtConfig.AccessTokenExpiry.Seconds()),
	}, nil
}

func (s *AuthService) Register(ctx context.Context, req dto.RegisterRequest, ipAddress, userAgent string) (*dto.RegisterResponse, error) {
	// Check if email already exists
	existing, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, apperror.InternalError(err)
	}
	if existing != nil {
		return nil, apperror.EmailAlreadyExists()
	}

	// Get owner role (first user of new company is always owner)
	ownerRole, err := s.roleRepo.GetByCode(ctx, "owner", nil)
	if err != nil {
		return nil, apperror.InternalError(err)
	}
	if ownerRole == nil {
		return nil, apperror.InternalError(fmt.Errorf("owner role not found"))
	}

	// Generate unique company code
	companyCode := generateCompanyCode(req.Name)

	// Create company with default template
	company := &model.Company{
		Code:     companyCode,
		Name:     fmt.Sprintf("%s's Business", req.Name),
		IsActive: true,
	}

	if err := s.companyRepo.Create(ctx, company); err != nil {
		return nil, apperror.InternalError(err)
	}

	// Create default branch
	branch := &model.Branch{
		CompanyID: company.ID,
		Code:      "HQ",
		Name:      "Headquarters",
		IsActive:  true,
	}

	if err := s.branchRepo.Create(ctx, branch); err != nil {
		return nil, apperror.InternalError(err)
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, apperror.InternalError(err)
	}

	// Create user with owner role
	user := &model.User{
		CompanyID:    company.ID,
		RoleID:       &ownerRole.ID,
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		Name:         req.Name,
		Status:       model.UserStatusActive,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, apperror.InternalError(err)
	}

	// Assign user to default branch
	userBranch := &model.UserBranch{
		UserID:    user.ID,
		BranchID:  branch.ID,
		IsDefault: true,
	}
	if err := s.userBranchRepo.Create(ctx, userBranch); err != nil {
		return nil, apperror.InternalError(err)
	}

	// Generate tokens
	accessToken := sessiontoken.New()
	refreshToken := sessiontoken.New()

	// Create session
	session := &model.UserSession{
		UserID:                user.ID,
		CompanyID:             company.ID,
		BranchID:              &branch.ID,
		AccessTokenHash:       sessiontoken.Hash(accessToken),
		AccessTokenExpiresAt:  time.Now().Add(s.jwtConfig.AccessTokenExpiry),
		RefreshTokenHash:      sessiontoken.Hash(refreshToken),
		RefreshTokenExpiresAt: time.Now().Add(s.jwtConfig.RefreshTokenExpiry),
		IPAddress:             &ipAddress,
		UserAgent:             &userAgent,
	}

	if err := s.sessionRepo.Create(ctx, session); err != nil {
		return nil, apperror.InternalError(err)
	}

	return &dto.RegisterResponse{
		User: &dto.AuthUserResponse{
			ID:    user.ID,
			Name:  user.Name,
			Email: user.Email,
			Role:  ownerRole.Code,
			Company: &dto.CompanyInfo{
				ID:   company.ID,
				Code: company.Code,
				Name: company.Name,
			},
			// The user was just assigned to this branch as their default
			// above; surface it so the register response matches login.
			DefaultBranch: &dto.BranchInfo{
				ID:   branch.ID,
				Code: branch.Code,
				Name: branch.Name,
			},
		},
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int(s.jwtConfig.AccessTokenExpiry.Seconds()),
	}, nil
}

func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*dto.RefreshTokenResponse, error) {
	session, err := s.sessionRepo.GetByRefreshTokenHash(ctx, sessiontoken.Hash(refreshToken))
	if err != nil {
		return nil, apperror.InternalError(err)
	}
	if session == nil {
		return nil, apperror.Unauthorized("invalid refresh token")
	}

	if session.IsRefreshTokenExpired() {
		return nil, apperror.Unauthorized("refresh token expired")
	}

	// Generate new tokens
	newAccessToken := sessiontoken.New()
	newRefreshToken := sessiontoken.New()

	// Replace the old session with a new one atomically
	newSession := &model.UserSession{
		UserID:                session.UserID,
		CompanyID:             session.CompanyID,
		BranchID:              session.BranchID,
		AccessTokenHash:       sessiontoken.Hash(newAccessToken),
		AccessTokenExpiresAt:  time.Now().Add(s.jwtConfig.AccessTokenExpiry),
		RefreshTokenHash:      sessiontoken.Hash(newRefreshToken),
		RefreshTokenExpiresAt: time.Now().Add(s.jwtConfig.RefreshTokenExpiry),
		IPAddress:             session.IPAddress,
		UserAgent:             session.UserAgent,
	}

	rotated, err := s.sessionRepo.Rotate(ctx, session.ID, newSession)
	if err != nil {
		return nil, apperror.InternalError(err)
	}
	if !rotated {
		// Another request exchanged this refresh token first.
		return nil, apperror.Unauthorized("invalid refresh token")
	}

	return &dto.RefreshTokenResponse{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
		ExpiresIn:    int(s.jwtConfig.AccessTokenExpiry.Seconds()),
	}, nil
}

func (s *AuthService) Logout(ctx context.Context, accessToken string) error {
	session, err := s.sessionRepo.GetByAccessTokenHash(ctx, sessiontoken.Hash(accessToken))
	if err != nil {
		return apperror.InternalError(err)
	}
	if session == nil {
		return nil // Already logged out
	}

	return s.sessionRepo.Revoke(ctx, session.ID)
}

func (s *AuthService) ValidateAccessToken(ctx context.Context, accessToken string) (*model.UserSession, error) {
	session, err := s.sessionRepo.GetByAccessTokenHash(ctx, sessiontoken.Hash(accessToken))
	if err != nil {
		return nil, apperror.InternalError(err)
	}
	if session == nil {
		return nil, apperror.Unauthorized("invalid access token")
	}

	if session.IsAccessTokenExpired() {
		return nil, apperror.Unauthorized("access token expired")
	}

	// Update last used
	_ = s.sessionRepo.UpdateLastUsed(ctx, session.ID)

	return session, nil
}

func generateCompanyCode(name string) string {
	// Generate a short code from name + random suffix
	code := strings.ToUpper(strings.ReplaceAll(name, " ", ""))
	if len(code) > 5 {
		code = code[:5]
	}

	bytes := make([]byte, 3)
	rand.Read(bytes)
	suffix := strings.ToUpper(hex.EncodeToString(bytes))[:4]

	return fmt.Sprintf("%s-%s", code, suffix)
}

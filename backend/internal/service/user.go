package service

import (
	"context"

	"github.com/irvanmhndra/pos-core-api/internal/dto"
	"github.com/irvanmhndra/pos-core-api/internal/model"
	"github.com/irvanmhndra/pos-core-api/internal/repository"
	"github.com/irvanmhndra/pos-core-api/pkg/apperror"
	"github.com/irvanmhndra/pos-core-api/pkg/httputil"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	userRepo repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

func (s *UserService) Create(ctx context.Context, companyID int64, req dto.CreateUserRequest) (*dto.UserResponse, error) {
	// Check if email already exists
	existing, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, apperror.InternalError(err)
	}
	if existing != nil {
		return nil, apperror.EmailAlreadyExists()
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, apperror.InternalError(err)
	}

	user := &model.User{
		CompanyID:    companyID,
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		Name:         req.Name,
		Status:       model.UserStatusActive,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, apperror.InternalError(err)
	}

	return dto.NewUserResponse(user), nil
}

func (s *UserService) GetByID(ctx context.Context, companyID, id int64) (*dto.UserResponse, error) {
	user, err := s.userRepo.GetByID(ctx, companyID, id)
	if err != nil {
		return nil, apperror.InternalError(err)
	}
	if user == nil {
		return nil, apperror.NotFound("user")
	}

	return dto.NewUserResponse(user), nil
}

type UserListResult struct {
	Users      []*dto.UserResponse
	Pagination *httputil.Pagination
}

func (s *UserService) List(ctx context.Context, companyID int64, req dto.ListUsersRequest) (*UserListResult, error) {
	req.SetDefaults()

	users, total, err := s.userRepo.List(ctx, companyID, req.Limit, req.Offset())
	if err != nil {
		return nil, apperror.InternalError(err)
	}

	totalPages := (total + req.Limit - 1) / req.Limit
	pagination := &httputil.Pagination{
		TotalRecords: total,
		TotalPages:   totalPages,
		CurrentPage:  req.Page,
		PerPage:      req.Limit,
		Count:        len(users),
	}

	if req.Page < totalPages {
		nextPage := req.Page + 1
		pagination.NextPage = &nextPage
	}
	if req.Page > 1 {
		prevPage := req.Page - 1
		pagination.PrevPage = &prevPage
	}

	return &UserListResult{
		Users:      dto.NewUserListResponse(users),
		Pagination: pagination,
	}, nil
}

func (s *UserService) Update(ctx context.Context, companyID, id int64, req dto.UpdateUserRequest) (*dto.UserResponse, error) {
	user, err := s.userRepo.GetByID(ctx, companyID, id)
	if err != nil {
		return nil, apperror.InternalError(err)
	}
	if user == nil {
		return nil, apperror.NotFound("user")
	}

	if req.Email != "" && req.Email != user.Email {
		exists, err := s.userRepo.EmailExists(ctx, req.Email, id)
		if err != nil {
			return nil, apperror.InternalError(err)
		}
		if exists {
			return nil, apperror.EmailAlreadyExists()
		}
		user.Email = req.Email
	}

	if req.Name != "" {
		user.Name = req.Name
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, apperror.InternalError(err)
	}

	return dto.NewUserResponse(user), nil
}

func (s *UserService) UpdateStatus(ctx context.Context, companyID, id int64, req dto.UpdateUserStatusRequest) error {
	user, err := s.userRepo.GetByID(ctx, companyID, id)
	if err != nil {
		return apperror.InternalError(err)
	}
	if user == nil {
		return apperror.NotFound("user")
	}

	status := model.UserStatus(req.Status)
	if err := s.userRepo.UpdateStatus(ctx, companyID, id, status); err != nil {
		return apperror.InternalError(err)
	}

	return nil
}

func (s *UserService) Delete(ctx context.Context, companyID, id int64) error {
	user, err := s.userRepo.GetByID(ctx, companyID, id)
	if err != nil {
		return apperror.InternalError(err)
	}
	if user == nil {
		return apperror.NotFound("user")
	}

	if err := s.userRepo.Delete(ctx, companyID, id); err != nil {
		return apperror.InternalError(err)
	}

	return nil
}

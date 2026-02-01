package dto

import (
	"time"

	"github.com/irvanmhndra/pos-core-api/internal/model"
)

// ============== Requests ==============

type CreateUserRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	Name     string `json:"name" validate:"required,min=2,max=100"`
}

type UpdateUserRequest struct {
	Email string `json:"email" validate:"omitempty,email"`
	Name  string `json:"name" validate:"omitempty,min=2,max=100"`
}

type UpdateUserStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=active inactive"`
}

type ListUsersRequest struct {
	Page  int `query:"page"`
	Limit int `query:"limit"`
}

func (r *ListUsersRequest) SetDefaults() {
	if r.Page < 1 {
		r.Page = 1
	}
	if r.Limit < 1 || r.Limit > 100 {
		r.Limit = 10
	}
}

func (r *ListUsersRequest) Offset() int {
	return (r.Page - 1) * r.Limit
}

// ============== Responses ==============

type UserResponse struct {
	ID        int64     `json:"id"`
	CompanyID int64     `json:"company_id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func NewUserResponse(u *model.User) *UserResponse {
	if u == nil {
		return nil
	}
	return &UserResponse{
		ID:        u.ID,
		CompanyID: u.CompanyID,
		Email:     u.Email,
		Name:      u.Name,
		Status:    string(u.Status),
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

func NewUserListResponse(users []*model.User) []*UserResponse {
	result := make([]*UserResponse, 0, len(users))
	for _, u := range users {
		result = append(result, NewUserResponse(u))
	}
	return result
}

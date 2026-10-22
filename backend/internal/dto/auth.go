package dto

// ============== Requests ==============

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type RegisterRequest struct {
	Name     string `json:"name" validate:"required,min=3,max=100"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// ============== Responses ==============

type CompanyInfo struct {
	ID   int64  `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

type BranchInfo struct {
	ID   int64  `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

type AuthUserResponse struct {
	ID            int64        `json:"id"`
	Name          string       `json:"name"`
	Email         string       `json:"email"`
	Role          string       `json:"role"`
	Company       *CompanyInfo `json:"company"`
	DefaultBranch *BranchInfo  `json:"default_branch,omitempty"`
}

type LoginResponse struct {
	User         *AuthUserResponse `json:"user"`
	AccessToken  string            `json:"access_token"`
	RefreshToken string            `json:"refresh_token"`
	ExpiresIn    int               `json:"expires_in"`
}

type RegisterResponse struct {
	User         *AuthUserResponse `json:"user"`
	AccessToken  string            `json:"access_token"`
	RefreshToken string            `json:"refresh_token"`
	ExpiresIn    int               `json:"expires_in"`
}

type RefreshTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

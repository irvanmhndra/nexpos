package apperror

import "net/http"

type AppError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	HTTPStatus int    `json:"-"`
	Details    any    `json:"details,omitempty"`
}

func (e *AppError) Error() string {
	return e.Message
}

// Common errors
func BadRequest(message string) *AppError {
	return &AppError{Code: "BAD_REQUEST", Message: message, HTTPStatus: http.StatusBadRequest}
}

func NotFound(resource string) *AppError {
	return &AppError{Code: "NOT_FOUND", Message: resource + " not found", HTTPStatus: http.StatusNotFound}
}

func Unauthorized(message string) *AppError {
	if message == "" {
		message = "unauthorized"
	}
	return &AppError{Code: "UNAUTHORIZED", Message: message, HTTPStatus: http.StatusUnauthorized}
}

func Forbidden(message string) *AppError {
	return &AppError{Code: "FORBIDDEN", Message: message, HTTPStatus: http.StatusForbidden}
}

func Conflict(message string) *AppError {
	return &AppError{Code: "CONFLICT", Message: message, HTTPStatus: http.StatusConflict}
}

func InternalError(err error) *AppError {
	// Include actual error message for debugging
	message := "internal server error"
	if err != nil {
		message = err.Error()
	}
	return &AppError{Code: "INTERNAL_ERROR", Message: message, HTTPStatus: http.StatusInternalServerError}
}

func ValidationError(message string, details any) *AppError {
	return &AppError{
		Code:       "VALIDATION_ERROR",
		Message:    message,
		HTTPStatus: http.StatusUnprocessableEntity,
		Details:    details,
	}
}

// Domain-specific errors
func EmailAlreadyExists() *AppError {
	return &AppError{
		Code:       "EMAIL_ALREADY_EXISTS",
		Message:    "email already exists",
		HTTPStatus: http.StatusConflict,
	}
}

func InvalidCredentials() *AppError {
	return &AppError{
		Code:       "INVALID_CREDENTIALS",
		Message:    "invalid email or password",
		HTTPStatus: http.StatusUnauthorized,
	}
}

func UserInactive() *AppError {
	return &AppError{
		Code:       "USER_INACTIVE",
		Message:    "user account is inactive",
		HTTPStatus: http.StatusForbidden,
	}
}

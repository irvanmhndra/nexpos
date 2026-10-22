package httputil

import (
	"encoding/json"
	"errors"
	"io"
	"strings"

	"github.com/irvanmhndra/nexpos-api/pkg/apperror"
	"github.com/labstack/echo/v5"
)

type Response struct {
	Success   bool   `json:"success"`
	Message   string `json:"message,omitempty"`
	Data      any    `json:"data,omitempty"`
	ErrorCode string `json:"error_code,omitempty"`
	Errors    any    `json:"errors,omitempty"`
	Meta      *Meta  `json:"meta,omitempty"`
}

type Meta struct {
	Pagination *Pagination `json:"pagination,omitempty"`
	Summary    any         `json:"summary,omitempty"`
}

type Pagination struct {
	TotalRecords int  `json:"total_records"`
	TotalPages   int  `json:"total_pages"`
	CurrentPage  int  `json:"current_page"`
	PerPage      int  `json:"per_page"`
	Count        int  `json:"count"`
	NextPage     *int `json:"next_page"`
	PrevPage     *int `json:"prev_page"`
}

func Success(c *echo.Context, status int, message string, data any) error {
	return (*c).JSON(status, Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func SuccessWithPagination(c *echo.Context, status int, message string, data any, pagination *Pagination, summary any) error {
	return (*c).JSON(status, Response{
		Success: true,
		Message: message,
		Data:    data,
		Meta:    &Meta{Pagination: pagination, Summary: summary},
	})
}

func Error(c *echo.Context, err error) error {
	if appErr, ok := err.(*apperror.AppError); ok {
		resp := Response{
			Success:   false,
			Message:   appErr.Message,
			ErrorCode: appErr.Code,
		}
		if appErr.Details != nil {
			resp.Errors = appErr.Details
		}
		return (*c).JSON(appErr.HTTPStatus, resp)
	}
	return (*c).JSON(500, Response{
		Success:   false,
		Message:   "internal server error",
		ErrorCode: "INTERNAL_ERROR",
	})
}

type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func ValidationError(c *echo.Context, errors []FieldError) error {
	return (*c).JSON(422, Response{
		Success:   false,
		Message:   "Validation failed",
		ErrorCode: "VALIDATION_ERROR",
		Errors:    errors,
	})
}

// BindError converts a bind/unmarshal error into a user-friendly AppError.
func BindError(err error) *apperror.AppError {
	var typeErr *json.UnmarshalTypeError
	if errors.As(err, &typeErr) {
		field := typeErr.Field
		expected := typeErr.Type.String()
		return apperror.ValidationError("Validation failed", []FieldError{
			{Field: field, Message: "must be " + friendlyType(expected)},
		})
	}

	var syntaxErr *json.SyntaxError
	if errors.As(err, &syntaxErr) {
		return apperror.BadRequest("malformed JSON body")
	}

	if errors.Is(err, io.EOF) {
		return apperror.BadRequest("request body is empty")
	}

	return apperror.BadRequest("invalid request body")
}

func friendlyType(goType string) string {
	switch {
	case strings.HasPrefix(goType, "int"), strings.HasPrefix(goType, "float"):
		return "a number"
	case goType == "string":
		return "a string"
	case goType == "bool":
		return "true or false"
	default:
		return goType
	}
}

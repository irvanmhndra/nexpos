package httputil

import (
	"time"

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
	ServerTime string      `json:"server_time"`
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

func newMeta() *Meta {
	return &Meta{
		ServerTime: time.Now().UTC().Format(time.RFC3339),
	}
}

func Success(c *echo.Context, status int, message string, data any) error {
	return (*c).JSON(status, Response{
		Success: true,
		Message: message,
		Data:    data,
		Meta:    newMeta(),
	})
}

func SuccessWithPagination(c *echo.Context, status int, message string, data any, pagination *Pagination, summary any) error {
	meta := newMeta()
	meta.Pagination = pagination
	meta.Summary = summary

	return (*c).JSON(status, Response{
		Success: true,
		Message: message,
		Data:    data,
		Meta:    meta,
	})
}

func Error(c *echo.Context, err error) error {
	if appErr, ok := err.(*apperror.AppError); ok {
		resp := Response{
			Success:   false,
			Message:   appErr.Message,
			ErrorCode: appErr.Code,
			Meta:      newMeta(),
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
		Meta:      newMeta(),
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
		Meta:      newMeta(),
	})
}

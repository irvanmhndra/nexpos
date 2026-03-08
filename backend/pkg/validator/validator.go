package validator

import (
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/irvanmhndra/nexpos-api/pkg/httputil"
)

type CustomValidator struct {
	validator *validator.Validate
}

func New() *CustomValidator {
	v := validator.New()

	// Use JSON tag names for field names in error messages
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	return &CustomValidator{validator: v}
}

func (cv *CustomValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}

func (cv *CustomValidator) ValidateAndFormat(i interface{}) []httputil.FieldError {
	if err := cv.validator.Struct(i); err != nil {
		return FormatValidationErrors(err)
	}
	return nil
}

func FormatValidationErrors(err error) []httputil.FieldError {
	var fieldErrors []httputil.FieldError

	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrors {
			fieldErrors = append(fieldErrors, httputil.FieldError{
				Field:   e.Field(),
				Message: formatMessage(e),
			})
		}
	}

	return fieldErrors
}

func formatMessage(e validator.FieldError) string {
	switch e.Tag() {
	case "required":
		return e.Field() + " is required"
	case "email":
		return "Invalid email format"
	case "min":
		return e.Field() + " must be at least " + e.Param() + " characters"
	case "max":
		return e.Field() + " must be at most " + e.Param() + " characters"
	case "oneof":
		return e.Field() + " must be one of: " + e.Param()
	default:
		return e.Field() + " is invalid"
	}
}

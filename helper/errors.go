package helper

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
)

type AppError struct {
	Status  int                 // HTTP status code
	Message string              // Human-readable message
	Errors  map[string][]string // Validation error details
	cause   error               // Root cause for logging
}

func (e *AppError) Error() string {
	if e.cause != nil {
		return fmt.Sprintf("%d %s: %v", e.Status, e.Message, e.cause)
	}
	return fmt.Sprintf("%d %s", e.Status, e.Message)
}

func (e *AppError) Unwrap() error {
	return e.cause
}

func BadRequest(message string) *AppError {
	return &AppError{
		Status:  fiber.StatusBadRequest,
		Message: message,
	}
}

func Unauthorized(message string) *AppError {
	return &AppError{
		Status:  fiber.StatusUnauthorized,
		Message: message,
	}
}

func Forbidden(message string) *AppError {
	return &AppError{
		Status:  fiber.StatusForbidden,
		Message: message,
	}
}

func NotFound(message string) *AppError {
	return &AppError{
		Status:  fiber.StatusNotFound,
		Message: message,
	}
}

func Conflict(message string) *AppError {
	return &AppError{
		Status:  fiber.StatusConflict,
		Message: message,
	}
}

func Validation(errors map[string][]string) *AppError {
	return &AppError{
		Status:  fiber.StatusUnprocessableEntity,
		Message: "validasi gagal",
		Errors:  errors,
	}
}

func TooManyRequests(message string) *AppError {
	return &AppError{
		Status:  fiber.StatusTooManyRequests,
		Message: message,
	}
}

func Internal(cause error) *AppError {
	return &AppError{
		Status:  fiber.StatusInternalServerError,
		Message: "terjadi kesalahan pada server",
		cause:   cause,
	}
}

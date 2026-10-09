package config

import (
	"errors"
	"log/slog"

	"github.com/gofiber/fiber/v2"

	"siakad-mini/app/model"
	"siakad-mini/helper"
	"siakad-mini/middleware"
	"siakad-mini/route"
)

// NewApp merakit aplikasi Fiber, memasang middleware global, mendaftarkan rute, dan fallback 404.
func NewApp(logger *slog.Logger, deps route.Dependencies) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName:      GetEnv("APP_NAME", "SIAKAD Mini API"),
		ErrorHandler: newErrorHandler(logger),
	})

	allowedOrigins := GetEnv("ALLOWED_ORIGINS", "*")
	middleware.Register(app, logger, allowedOrigins)
	route.Register(app, deps)

	// Fallback untuk endpoint yang tidak terdaftar (404)
	app.Use(func(c *fiber.Ctx) error {
		return helper.NotFound("endpoint tidak ditemukan")
	})

	return app
}

// newErrorHandler menangani semua error (AppError, FiberError, maupun unknown error)
// dan menormalkannya ke dalam envelope model.ErrorResponse seragam.
func newErrorHandler(logger *slog.Logger) fiber.ErrorHandler {
	isProd := GetEnv("APP_ENV", "development") == "production"

	return func(c *fiber.Ctx, err error) error {
		requestID, _ := c.Locals("requestid").(string)

		var appErr *helper.AppError
		var fiberErr *fiber.Error

		statusCode := fiber.StatusInternalServerError
		message := "terjadi kesalahan pada server"
		var fieldErrors map[string][]string

		switch {
		case errors.As(err, &appErr):
			statusCode = appErr.Status
			message = appErr.Message
			fieldErrors = appErr.Errors
		case errors.As(err, &fiberErr):
			statusCode = fiberErr.Code
			message = fiberErr.Message
		default:
			if !isProd {
				message = err.Error()
			}
		}

		if statusCode >= fiber.StatusInternalServerError {
			var causeStr string
			if appErr != nil && appErr.Unwrap() != nil {
				causeStr = appErr.Unwrap().Error()
			} else {
				causeStr = err.Error()
			}

			logger.Error("server_error",
				slog.String("request_id", requestID),
				slog.String("method", c.Method()),
				slog.String("path", c.Path()),
				slog.Int("status", statusCode),
				slog.String("error", causeStr),
			)
		} else {
			logger.Warn("client_error",
				slog.String("request_id", requestID),
				slog.String("method", c.Method()),
				slog.String("path", c.Path()),
				slog.Int("status", statusCode),
				slog.String("message", message),
			)
		}

		return c.Status(statusCode).JSON(model.ErrorResponse{
			Success: false,
			Message: message,
			Errors:  fieldErrors,
		})
	}
}

package middleware

import (
	"log/slog"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
)

// Register memasang semua middleware global ke aplikasi Fiber.
func Register(app *fiber.App, logger *slog.Logger, origins string) {
	// 1. Request ID — injeksi X-Request-ID ke setiap request
	app.Use(requestid.New())

	// 2. Recover — tangkap panic dan kembalikan 500
	app.Use(recover.New(recover.Config{
		EnableStackTrace: false,
	}))

	// 3. Helmet — set security headers
	app.Use(helmet.New())

	// 4. CORS
	app.Use(corsPolicy(origins))

	// 5. Request logger terstruktur
	app.Use(RequestLogger(logger))
}

// corsPolicy mengembalikan konfigurasi CORS dengan origins yang bisa dikonfigurasi.
func corsPolicy(origins string) fiber.Handler {
	return cors.New(cors.Config{
		AllowOrigins: origins,
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders: "Origin,Content-Type,Accept,Authorization,X-Request-ID",
		MaxAge:       86400,
	})
}

// RequestLogger mencatat setiap request secara terstruktur menggunakan slog.
func RequestLogger(logger *slog.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		duration := time.Since(start)

		// Coba ambil user dari locals (diisi oleh RequireAuth jika sudah login)
		userID := ""
		if u := c.Locals("authUser"); u != nil {
			userID = "authenticated"
		}

		logger.Info("http request",
			slog.String("method", c.Method()),
			slog.String("path", c.Path()),
			slog.Int("status", c.Response().StatusCode()),
			slog.Duration("duration", duration),
			slog.String("ip", c.IP()),
			slog.String("request_id", c.Locals("requestid").(string)),
			slog.String("user", userID),
		)
		return err
	}
}

// RequireJSON menolak request POST/PUT yang tidak menggunakan Content-Type: application/json.
func RequireJSON(c *fiber.Ctx) error {
	method := c.Method()
	if method == fiber.MethodPost || method == fiber.MethodPut {
		ct := c.Get(fiber.HeaderContentType)
		if !strings.Contains(ct, fiber.MIMEApplicationJSON) {
			return fiber.NewError(
				fiber.StatusUnsupportedMediaType,
				"Content-Type harus application/json",
			)
		}
	}
	return c.Next()
}

package middleware

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"

	"siakad-mini/app/model"
	"siakad-mini/helper"
)

// RequireAuth memvalidasi Bearer token dari header Authorization.
// Menyimpan AuthUser ke locals dengan key "authUser" jika valid;
// mengembalikan 401 jika token hilang, tidak valid, atau sudah kedaluwarsa.
func RequireAuth(jwt *helper.JWTManager) fiber.Handler {
	return func(c *fiber.Ctx) error {
		header := c.Get(fiber.HeaderAuthorization)
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			return helper.Unauthorized("token autentikasi diperlukan")
		}

		tokenStr := strings.TrimPrefix(header, "Bearer ")
		authUser, err := jwt.Parse(tokenStr)
		if err != nil {
			if errors.Is(err, helper.ErrExpiredToken) {
				return helper.Unauthorized("token sudah kedaluwarsa")
			}
			return helper.Unauthorized("token tidak valid")
		}

		c.Locals("authUser", authUser)
		return c.Next()
	}
}

// LoginRateLimiter membatasi percobaan login gagal: maks 5 per menit per IP.
// SkipSuccessfulRequests memastikan hanya request yang gagal (status >= 400) yang dihitung.
func LoginRateLimiter() fiber.Handler {
	return limiter.New(limiter.Config{
		Max:                    5,
		SkipSuccessfulRequests: true,
		LimitReached: func(c *fiber.Ctx) error {
			return helper.TooManyRequests("terlalu banyak percobaan login, coba lagi dalam 1 menit")
		},
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
		// Fiber limiter default: 1 menit window (dapat dikonfigurasi via Expiration jika perlu)
	})
}

// authUserFromLocals membaca AuthUser dari Fiber locals secara type-safe.
func authUserFromLocals(c *fiber.Ctx) (model.AuthUser, bool) {
	v := c.Locals("authUser")
	if v == nil {
		return model.AuthUser{}, false
	}
	u, ok := v.(model.AuthUser)
	return u, ok
}

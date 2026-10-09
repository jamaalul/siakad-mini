package middleware

import (
	"github.com/gofiber/fiber/v2"

	"siakad-mini/helper"
)

// RequireRole mengizinkan request hanya jika AuthUser memiliki salah satu role yang diberikan.
// Harus dipasang setelah RequireAuth karena bergantung pada locals "authUser".
func RequireRole(roles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authUser, ok := authUserFromLocals(c)
		if !ok {
			return helper.Unauthorized("belum terautentikasi")
		}
		if !helper.HasAnyRole(authUser.Role, roles...) {
			return helper.Forbidden("tidak memiliki izin untuk mengakses resource ini")
		}
		return c.Next()
	}
}

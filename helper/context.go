package helper

import (
	"github.com/gofiber/fiber/v2"

	"siakad-mini/app/model"
)

const LocalsAuthUser = "authUser"

// CurrentUser mengambil data AuthUser dari Fiber locals jika sudah diautentikasi.
func CurrentUser(c *fiber.Ctx) (model.AuthUser, bool) {
	user, ok := c.Locals(LocalsAuthUser).(model.AuthUser)
	return user, ok
}

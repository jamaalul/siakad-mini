package service

import (
	"regexp"
	"strings"

	"siakad-mini/app/model"
)

const minPasswordLength = 8

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// ValidateLogin memvalidasi payload LoginRequest tanpa dependensi HTTP atau database.
func ValidateLogin(req model.LoginRequest) map[string][]string {
	errs := make(map[string][]string)

	email := strings.TrimSpace(req.Email)
	if email == "" {
		errs["email"] = append(errs["email"], "wajib diisi")
	} else if !isValidEmail(email) {
		errs["email"] = append(errs["email"], "format email tidak valid")
	}

	if req.Password == "" {
		errs["password"] = append(errs["password"], "wajib diisi")
	} else if len(req.Password) < minPasswordLength {
		errs["password"] = append(errs["password"], "minimal 8 karakter")
	}

	if len(errs) == 0 {
		return nil
	}
	return errs
}

func isValidEmail(email string) bool {
	return emailRegex.MatchString(email)
}

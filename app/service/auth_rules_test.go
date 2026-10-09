package service

import (
	"testing"

	"siakad-mini/app/model"
)

func TestValidateLogin(t *testing.T) {
	t.Run("valid input", func(t *testing.T) {
		req := model.LoginRequest{
			Email:    "admin@siakad.test",
			Password: "Admin12345",
		}
		errs := ValidateLogin(req)
		if len(errs) != 0 {
			t.Errorf("seharusnya valid, tetapi dapat error: %v", errs)
		}
	})

	t.Run("empty fields", func(t *testing.T) {
		// Empty email and empty password
		errs := ValidateLogin(model.LoginRequest{
			Email:    "",
			Password: "",
		})
		if len(errs["email"]) == 0 || errs["email"][0] != "wajib diisi" {
			t.Errorf("error email seharusnya 'wajib diisi', dapat: %v", errs["email"])
		}
		if len(errs["password"]) == 0 || errs["password"][0] != "wajib diisi" {
			t.Errorf("error password seharusnya 'wajib diisi', dapat: %v", errs["password"])
		}

		// Whitespace only email
		errs = ValidateLogin(model.LoginRequest{
			Email:    "   ",
			Password: "ValidPassword123",
		})
		if len(errs["email"]) == 0 || errs["email"][0] != "wajib diisi" {
			t.Errorf("error email seharusnya 'wajib diisi', dapat: %v", errs["email"])
		}
	})

	t.Run("bad email", func(t *testing.T) {
		invalidEmails := []string{
			"admin",
			"admin@",
			"@siakad.test",
			"admin@siakad",
			"admin siakad.test",
		}
		for _, email := range invalidEmails {
			errs := ValidateLogin(model.LoginRequest{
				Email:    email,
				Password: "ValidPassword123",
			})
			if len(errs["email"]) == 0 || errs["email"][0] != "format email tidak valid" {
				t.Errorf("email %q seharusnya menghasilkan 'format email tidak valid', dapat: %v", email, errs["email"])
			}
		}
	})

	t.Run("short password", func(t *testing.T) {
		errs := ValidateLogin(model.LoginRequest{
			Email:    "admin@siakad.test",
			Password: "short",
		})
		if len(errs["password"]) == 0 || errs["password"][0] != "minimal 8 karakter" {
			t.Errorf("password pendek seharusnya menghasilkan 'minimal 8 karakter', dapat: %v", errs["password"])
		}
	})
}

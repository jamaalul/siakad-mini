package service

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"

	"siakad-mini/app/model"
	"siakad-mini/app/repository"
	"siakad-mini/helper"
)

type AuthService struct {
	users    repository.UserRepository
	students repository.StudentRepository
	jwt      *helper.JWTManager
}

func NewAuthService(
	users repository.UserRepository,
	students repository.StudentRepository,
	jwt *helper.JWTManager,
) *AuthService {
	return &AuthService{users: users, students: students, jwt: jwt}
}

// Login memvalidasi kredensial dan mengeluarkan access token JWT.
func (s *AuthService) Login(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	var req model.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}
	req.Email = strings.TrimSpace(req.Email)

	// Validasi struct menggunakan validator
	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	user, err := s.users.FindByEmail(ctx, req.Email)
	if err != nil {
		// Dummy compare agar waktu respons tidak mengungkap apakah email ada atau tidak
		helper.VerifyDummyPassword(req.Password)
		return helper.Unauthorized("email atau password salah")
	}

	if !helper.VerifyPassword(user.Password, req.Password) {
		return helper.Unauthorized("email atau password salah")
	}

	accessToken, err := s.jwt.GenerateAccess(user)
	if err != nil {
		return helper.Internal(err)
	}

	return helper.Success(c, fiber.StatusOK, "login berhasil", model.LoginResponse{
		AccessToken: accessToken,
		TokenType:   "Bearer",
		ExpiresIn:   s.jwt.ExpiresIn(),
		User: model.LoginUser{
			ID:    user.ID,
			Email: user.Email,
			Role:  user.Role,
		},
	})
}

// Me mengembalikan data pengguna yang sedang login.
// Untuk mahasiswa juga menyertakan profil student; 401 jika student sudah soft-deleted.
func (s *AuthService) Me(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	authUser, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Unauthorized("belum terautentikasi")
	}

	user, err := s.users.FindByID(ctx, authUser.UserID)
	if err != nil {
		return helper.Unauthorized("user tidak ditemukan")
	}

	resp := model.MeResponse{
		ID:        user.ID,
		Email:     user.Email,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
	}

	if user.Role == model.RoleMahasiswa {
		student, err := s.students.FindByUserID(ctx, user.ID)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				// Student soft-deleted atau tidak ditemukan -> 401
				return helper.Unauthorized("akun mahasiswa tidak aktif")
			}
			return helper.Internal(err)
		}
		resp.Student = &model.MeStudentInfo{
			NIM:      student.NIM,
			Nama:     student.Nama,
			Prodi:    student.Prodi,
			Angkatan: student.Angkatan,
		}
	}

	return helper.Success(c, fiber.StatusOK, "profil berhasil diambil", resp)
}

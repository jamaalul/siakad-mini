package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"siakad-mini/app/model"
	"siakad-mini/app/repository"
	"siakad-mini/helper"
)

const portfolioTimeout = 5 * time.Second

// PortfolioService menangani logika bisnis portofolio mahasiswa lintas database:
// verifikasi data mahasiswa di PostgreSQL dan penyimpanan profil portofolio di MongoDB.
type PortfolioService struct {
	portfolios repository.PortfolioRepository
	students   repository.StudentRepository
}

func NewPortfolioService(
	portfolios repository.PortfolioRepository,
	students repository.StudentRepository,
) *PortfolioService {
	return &PortfolioService{portfolios: portfolios, students: students}
}

func portfolioCtx(c *fiber.Ctx) (context.Context, context.CancelFunc) {
	return context.WithTimeout(c.UserContext(), portfolioTimeout)
}

func portfolioJSON(c *fiber.Ctx, status int, message string, data any) error {
	return c.Status(status).JSON(fiber.Map{"success": true, "message": message, "data": data})
}

func (s *PortfolioService) currentStudent(ctx context.Context, c *fiber.Ctx) (id int, name string, err error) {
	authUser, ok := helper.CurrentUser(c)
	if !ok {
		return 0, "", helper.Forbidden("belum terautentikasi")
	}
	student, err := s.students.FindByUserID(ctx, authUser.UserID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return 0, "", helper.Forbidden("akun ini tidak terhubung dengan data mahasiswa")
		}
		return 0, "", helper.Internal(err)
	}
	return student.ID, student.Nama, nil
}

// ---------- GET /students/:id/portfolio ----------
func (s *PortfolioService) Get(c *fiber.Ctx) error {
	ctx, cancel := portfolioCtx(c)
	defer cancel()

	studentID, _ := c.ParamsInt("id", 0)
	if studentID < 1 {
		return helper.BadRequest("id mahasiswa harus berupa angka positif")
	}

	// Periksa mahasiswa ada di PostgreSQL
	if _, err := s.students.FindByID(ctx, studentID); err != nil {
		return translatePortfolioError(err)
	}

	portfolio, err := s.portfolios.FindByStudent(ctx, studentID)
	if err != nil {
		return translatePortfolioError(err)
	}
	return portfolioJSON(c, fiber.StatusOK, "portofolio berhasil diambil", portfolio)
}

// ---------- PUT /students/:id/portfolio ----------
func (s *PortfolioService) Upsert(c *fiber.Ctx) error {
	ctx, cancel := portfolioCtx(c)
	defer cancel()

	studentID, _ := c.ParamsInt("id", 0)
	if studentID < 1 {
		return helper.BadRequest("id mahasiswa harus berupa angka positif")
	}

	// Periksa mahasiswa ada di PostgreSQL
	if _, err := s.students.FindByID(ctx, studentID); err != nil {
		return translatePortfolioError(err)
	}

	// Otorisasi: hanya pemilik atau admin yang boleh mengubah portofolio
	authUser, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Forbidden("belum terautentikasi")
	}
	isAdmin := string(authUser.Role) == string(model.RoleAdmin)
	if !isAdmin {
		myID, _, err := s.currentStudent(ctx, c)
		if err != nil {
			return err
		}
		if myID != studentID {
			return helper.Forbidden("hanya boleh mengubah portofolio sendiri")
		}
	}

	var req model.PutPortfolioRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}
	req.Headline = strings.TrimSpace(req.Headline)
	req.Bio = strings.TrimSpace(req.Bio)
	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	portfolio, err := s.portfolios.Upsert(ctx, studentID, req)
	if err != nil {
		return translatePortfolioError(err)
	}
	return portfolioJSON(c, fiber.StatusOK, "portofolio berhasil disimpan", portfolio)
}

// ---------- POST /students/:id/portfolio/projects ----------
func (s *PortfolioService) AddProject(c *fiber.Ctx) error {
	ctx, cancel := portfolioCtx(c)
	defer cancel()

	studentID, _ := c.ParamsInt("id", 0)
	if studentID < 1 {
		return helper.BadRequest("id mahasiswa harus berupa angka positif")
	}

	// Otorisasi: hanya pemilik atau admin
	authUser, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Forbidden("belum terautentikasi")
	}
	isAdmin := string(authUser.Role) == string(model.RoleAdmin)
	if !isAdmin {
		myID, _, err := s.currentStudent(ctx, c)
		if err != nil {
			return err
		}
		if myID != studentID {
			return helper.Forbidden("hanya boleh menambah proyek ke portofolio sendiri")
		}
	}

	var req model.AddProjectRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}
	req.Title = strings.TrimSpace(req.Title)
	req.URL = strings.TrimSpace(req.URL)
	req.Technologies = NormalizeTags(req.Technologies)
	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	portfolio, err := s.portfolios.AddProject(ctx, studentID, model.Project{
		Title: req.Title, Technologies: req.Technologies, URL: req.URL, Year: req.Year,
	})
	if err != nil {
		return translatePortfolioError(err)
	}
	return portfolioJSON(c, fiber.StatusCreated, "proyek berhasil ditambahkan", portfolio)
}

// ---------- DELETE /students/:id/portfolio/projects/:projectId ----------
func (s *PortfolioService) DeleteProject(c *fiber.Ctx) error {
	ctx, cancel := portfolioCtx(c)
	defer cancel()

	studentID, _ := c.ParamsInt("id", 0)
	if studentID < 1 {
		return helper.BadRequest("id mahasiswa harus berupa angka positif")
	}
	projectID := c.Params("projectId")

	// Otorisasi: hanya pemilik atau admin
	authUser, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Forbidden("belum terautentikasi")
	}
	isAdmin := string(authUser.Role) == string(model.RoleAdmin)
	if !isAdmin {
		myID, _, err := s.currentStudent(ctx, c)
		if err != nil {
			return err
		}
		if myID != studentID {
			return helper.Forbidden("hanya boleh menghapus proyek dari portofolio sendiri")
		}
	}

	if err := s.portfolios.DeleteProject(ctx, studentID, projectID); err != nil {
		return translatePortfolioError(err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// ---------- GET /portfolios/stats/technologies ----------
func (s *PortfolioService) TechStats(c *fiber.Ctx) error {
	ctx, cancel := portfolioCtx(c)
	defer cancel()

	stats, err := s.portfolios.TechStats(ctx)
	if err != nil {
		return helper.Internal(err)
	}
	return portfolioJSON(c, fiber.StatusOK, "statistik teknologi berhasil diambil", stats)
}

func translatePortfolioError(err error) error {

	var appErr *helper.AppError
	if errors.As(err, &appErr) {
		return err
	}
	var fe *fiber.Error
	if errors.As(err, &fe) {
		return err
	}
	switch {
	case errors.Is(err, repository.ErrInvalidID):
		return helper.BadRequest("id tidak valid")
	case errors.Is(err, repository.ErrNotFound):
		return helper.NotFound("data tidak ditemukan")
	case errors.Is(err, repository.ErrNotOwner):
		return helper.Forbidden("bukan pemilik data")
	case errors.Is(err, repository.ErrProjectLimit):
		return helper.Conflict("jumlah proyek sudah mencapai batas")
	case errors.Is(err, repository.ErrProjectNotFound):
		return helper.NotFound("proyek tidak ditemukan")
	default:
		return helper.Internal(err)
	}
}

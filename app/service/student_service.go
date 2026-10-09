package service

import (
	"errors"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"

	"siakad-mini/app/model"
	"siakad-mini/app/repository"
	"siakad-mini/helper"
)

type StudentService struct {
	students    repository.StudentRepository
	users       repository.UserRepository
	enrollments repository.EnrollmentRepository
}

func NewStudentService(
	students repository.StudentRepository,
	users repository.UserRepository,
	enrollments repository.EnrollmentRepository,
) *StudentService {
	return &StudentService{students: students, users: users, enrollments: enrollments}
}

// List mengembalikan daftar mahasiswa dengan filter, urutan, dan paginasi.
func (s *StudentService) List(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	q := helper.ParseStudentListQuery(c)

	sort, err := ValidateSort(q.Sort)
	if err != nil {
		return helper.BadRequest("parameter sort tidak valid: gunakan 'nama' atau '-ipk_terakhir'")
	}
	q.Sort = sort

	students, total, err := s.students.FindAll(ctx, q)
	if err != nil {
		return helper.Internal(err)
	}

	lastPage := CountLastPage(total, q.PerPage)
	meta := &model.Meta{
		CurrentPage: q.Page,
		PerPage:     q.PerPage,
		Total:       total,
		LastPage:    lastPage,
	}
	return helper.SuccessList(c, "daftar mahasiswa berhasil diambil", students, meta)
}

// Create membuat user + mahasiswa baru dalam satu transaksi; password = hash NIM.
func (s *StudentService) Create(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	var req model.CreateStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}
	req.NIM = strings.TrimSpace(req.NIM)
	req.Nama = strings.TrimSpace(req.Nama)
	req.Email = strings.TrimSpace(req.Email)

	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	hashed, err := helper.HashPassword(req.NIM)
	if err != nil {
		return helper.Internal(err)
	}

	student, err := s.students.CreateWithUser(ctx, req, hashed)
	if err != nil {
		if errors.Is(err, repository.ErrDuplicate) {
			return helper.Validation(map[string][]string{
				"nim":   {"NIM atau email sudah terdaftar"},
				"email": {"NIM atau email sudah terdaftar"},
			})
		}
		return helper.Internal(err)
	}

	return helper.Created(c, "mahasiswa berhasil dibuat", student,
		"/api/v1/students/"+strconv.Itoa(student.ID))
}

// Get mengembalikan detail mahasiswa beserta enrollments, total_sks, dan batas_sks.
func (s *StudentService) Get(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	id, ok := helper.ParamID(c)
	if !ok {
		return helper.BadRequest("id tidak valid")
	}

	authUser, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Unauthorized("belum terautentikasi")
	}

	student, err := s.students.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return helper.NotFound("mahasiswa tidak ditemukan")
		}
		return helper.Internal(err)
	}

	if !CanAccessStudent(authUser, student.UserID) {
		return helper.Forbidden("tidak memiliki akses ke data ini")
	}

	enrollments, err := s.enrollments.ListByStudent(ctx, student.ID)
	if err != nil {
		return helper.Internal(err)
	}
	if enrollments == nil {
		enrollments = []model.EnrollmentDetail{}
	}

	totalSKS := 0
	for _, e := range enrollments {
		totalSKS += e.SKS
	}
	batasSKS := MaxSKSByIPK(student.IPKTerakhir)

	detail := model.StudentDetail{
		Student:     student,
		Enrollments: enrollments,
		TotalSKS:    totalSKS,
		BatasSKS:    batasSKS,
	}

	return helper.Success(c, fiber.StatusOK, "detail mahasiswa berhasil diambil", detail)
}

// Update mengganti data mahasiswa (PUT); nim tidak dapat diubah.
func (s *StudentService) Update(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	id, ok := helper.ParamID(c)
	if !ok {
		return helper.BadRequest("id tidak valid")
	}

	var req model.UpdateStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}
	req.Nama = strings.TrimSpace(req.Nama)
	req.Prodi = strings.TrimSpace(req.Prodi)

	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	current, err := s.students.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return helper.NotFound("mahasiswa tidak ditemukan")
		}
		return helper.Internal(err)
	}

	updated := ApplyUpdate(current, req)

	result, err := s.students.Update(ctx, updated)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return helper.NotFound("mahasiswa tidak ditemukan")
		}
		if errors.Is(err, repository.ErrDuplicate) {
			return helper.Validation(map[string][]string{
				"nim": {"NIM sudah digunakan"},
			})
		}
		return helper.Internal(err)
	}

	return helper.Success(c, fiber.StatusOK, "mahasiswa berhasil diperbarui", result)
}

// Delete menghapus mahasiswa secara logis (soft delete); 204 No Content.
func (s *StudentService) Delete(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	id, ok := helper.ParamID(c)
	if !ok {
		return helper.BadRequest("id tidak valid")
	}

	if err := s.students.SoftDelete(ctx, id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return helper.NotFound("mahasiswa tidak ditemukan")
		}
		return helper.Internal(err)
	}

	return helper.NoContent(c)
}

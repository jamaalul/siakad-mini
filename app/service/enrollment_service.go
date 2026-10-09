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

type EnrollmentService struct {
	enrollments repository.EnrollmentRepository
	students    repository.StudentRepository
	courses     repository.CourseRepository
}

func NewEnrollmentService(
	enrollments repository.EnrollmentRepository,
	students repository.StudentRepository,
	courses repository.CourseRepository,
) *EnrollmentService {
	return &EnrollmentService{
		enrollments: enrollments,
		students:    students,
		courses:     courses,
	}
}

// Create mendaftarkan mahasiswa ke mata kuliah dalam satu transaksi terkunci.
// student_id selalu berasal dari JWT, tidak pernah dari request body.
func (s *EnrollmentService) Create(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	authUser, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Unauthorized("belum terautentikasi")
	}

	var req model.CreateEnrollmentRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}
	req.TahunAkademik = strings.TrimSpace(req.TahunAkademik)

	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	// Validasi format tahun akademik lebih ketat (second year = first + 1)
	if !ValidateAcademicYear(req.TahunAkademik) {
		return helper.Validation(map[string][]string{
			"tahun_akademik": {"format tidak valid: gunakan YYYY/YYYY+1-Ganjil atau Genap, contoh: 2026/2027-Ganjil"},
		})
	}

	// Resolusi student dari JWT user_id
	student, err := s.students.FindByUserID(ctx, authUser.UserID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return helper.Unauthorized("akun mahasiswa tidak aktif")
		}
		return helper.Internal(err)
	}

	var created model.Enrollment
	txErr := s.enrollments.WithinTx(ctx, func(tx repository.EnrollmentTx) error {
		// 1. Kunci baris student untuk mencegah race condition SKS
		_, err := tx.LockStudent(ctx, student.ID)
		if err != nil {
			return helper.Internal(err)
		}

		// 2. Kunci baris course untuk mencegah race condition kuota
		course, err := tx.LockCourse(ctx, req.CourseID)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return helper.Validation(map[string][]string{
					"course_id": {"mata kuliah tidak ditemukan"},
				})
			}
			return helper.Internal(err)
		}

		// 3. Periksa duplikat enrollment
		exists, err := tx.Exists(ctx, student.ID, req.CourseID, req.TahunAkademik)
		if err != nil {
			return helper.Internal(err)
		}
		if exists {
			return helper.Conflict("Anda sudah terdaftar di mata kuliah ini pada tahun akademik yang sama")
		}

		// 4. Periksa kuota
		terisi, err := tx.CountByCourse(ctx, req.CourseID)
		if err != nil {
			return helper.Internal(err)
		}
		if IsCourseFull(terisi, course.Kuota) {
			return helper.Validation(map[string][]string{
				"course_id": {"kuota mata kuliah penuh"},
			})
		}

		// 5. Periksa batas SKS
		taken, err := tx.SumSKS(ctx, student.ID, req.TahunAkademik)
		if err != nil {
			return helper.Internal(err)
		}
		max := MaxSKSByIPK(student.IPKTerakhir)
		if _, msg := CheckSKSLimit(taken, course.SKS, max); msg != "" {
			return helper.Validation(map[string][]string{
				"course_id": {msg},
			})
		}

		// 6. Buat enrollment
		e := model.Enrollment{
			StudentID:     student.ID,
			CourseID:      req.CourseID,
			TahunAkademik: req.TahunAkademik,
		}
		created, err = tx.Create(ctx, e)
		if err != nil {
			if errors.Is(err, repository.ErrDuplicate) {
				return helper.Conflict("Anda sudah terdaftar di mata kuliah ini pada tahun akademik yang sama")
			}
			return helper.Internal(err)
		}
		return nil
	})
	if txErr != nil {
		return txErr
	}

	return helper.Created(c, "pendaftaran mata kuliah berhasil", created,
		"/api/v1/enrollments/"+strconv.Itoa(created.ID))
}

// Delete membatalkan pendaftaran mata kuliah. Hanya pemilik enrollment yang bisa menghapus.
func (s *EnrollmentService) Delete(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	authUser, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Unauthorized("belum terautentikasi")
	}

	id, ok := helper.ParamID(c)
	if !ok {
		return helper.BadRequest("id tidak valid")
	}

	enrollment, err := s.enrollments.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return helper.NotFound("enrollment tidak ditemukan")
		}
		return helper.Internal(err)
	}

	// Resolusi student dari JWT user_id untuk verifikasi kepemilikan
	student, err := s.students.FindByUserID(ctx, authUser.UserID)
	if err != nil {
		return helper.Forbidden("tidak memiliki akses")
	}

	if enrollment.StudentID != student.ID {
		return helper.Forbidden("tidak memiliki akses ke enrollment ini")
	}

	if err := s.enrollments.Delete(ctx, id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return helper.NotFound("enrollment tidak ditemukan")
		}
		return helper.Internal(err)
	}

	return helper.NoContent(c)
}

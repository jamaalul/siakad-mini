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

const reviewTimeout = 5 * time.Second

// ReviewService memakai DUA database sekaligus: data kelas dan mahasiswa dibaca
// dari PostgreSQL, ulasan disimpan di MongoDB. Service-lah satu-satunya tempat
// yang mengenal keduanya.
type ReviewService struct {
	reviews  repository.ReviewRepository
	courses  repository.CourseRepository
	students repository.StudentRepository
}

func NewReviewService(
	reviews repository.ReviewRepository,
	courses repository.CourseRepository,
	students repository.StudentRepository,
) *ReviewService {
	return &ReviewService{reviews: reviews, courses: courses, students: students}
}

func reviewCtx(c *fiber.Ctx) (context.Context, context.CancelFunc) {
	return context.WithTimeout(c.UserContext(), reviewTimeout)
}

func reviewJSON(c *fiber.Ctx, status int, message string, data any) error {
	return c.Status(status).JSON(fiber.Map{"success": true, "message": message, "data": data})
}

// currentStudent mencari data mahasiswa milik akun yang sedang login.
// Tidak ada FOREIGN KEY ke MongoDB, jadi student_id yang disimpan di ulasan
// harus berasal dari sini, bukan dari input client.
func (s *ReviewService) currentStudent(ctx context.Context, c *fiber.Ctx) (id int, name string, err error) {
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

// ---------- POST /courses/:id/reviews ----------
func (s *ReviewService) Create(c *fiber.Ctx) error {
	ctx, cancel := reviewCtx(c)
	defer cancel()

	courseID, _ := c.ParamsInt("id", 0)
	if courseID < 1 {
		return helper.BadRequest("id kelas harus berupa angka positif")
	}

	var req model.CreateReviewRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}
	req.Comment = strings.TrimSpace(req.Comment)
	req.Tags = NormalizeTags(req.Tags)
	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	studentID, studentName, err := s.currentStudent(ctx, c)
	if err != nil {
		return err
	}

	// Tidak ada FOREIGN KEY yang menghubungkan MongoDB dengan PostgreSQL.
	// Pemeriksaan "kelasnya ada" dan "mahasiswa ini peserta kelasnya" harus
	// dilakukan sendiri oleh service, dengan bertanya ke PostgreSQL.
	if _, err := s.courses.FindByID(ctx, courseID); err != nil {
		return translateReviewError(err)
	}
	enrolled, err := s.courses.IsEnrolled(ctx, courseID, studentID)
	if err != nil {
		return helper.Internal(err)
	}
	if !enrolled {
		return helper.Forbidden("hanya peserta kelas yang boleh menulis ulasan")
	}

	created, err := s.reviews.Create(ctx, model.Review{
		CourseID: courseID, StudentID: studentID, StudentName: studentName,
		Rating: req.Rating, Comment: req.Comment, Tags: req.Tags,
	})
	if err != nil {
		return translateReviewError(err)
	}
	c.Set(fiber.HeaderLocation, "/api/v1/reviews/"+created.ID)
	return reviewJSON(c, fiber.StatusCreated, "ulasan berhasil disimpan", created)
}

// ---------- GET /courses/:id/reviews ----------
func (s *ReviewService) ListByCourse(c *fiber.Ctx) error {
	ctx, cancel := reviewCtx(c)
	defer cancel()

	courseID, _ := c.ParamsInt("id", 0)
	if courseID < 1 {
		return helper.BadRequest("id kelas harus berupa angka positif")
	}

	q := model.ReviewQuery{
		CourseID: courseID,
		Tag:      strings.ToLower(strings.TrimSpace(c.Query("tag"))),
		Before:   c.Query("before"),
		Limit:    c.QueryInt("limit", 10),
	}
	if q.Limit < 1 || q.Limit > 50 {
		q.Limit = 10
	}

	// Minta satu lebih banyak untuk mengetahui apakah masih ada halaman berikutnya.
	want := q.Limit
	q.Limit = want + 1

	reviews, err := s.reviews.FindByCourse(ctx, q)
	if err != nil {
		return translateReviewError(err)
	}
	hasMore := len(reviews) > want
	if hasMore {
		reviews = reviews[:want]
	}
	next := ""
	if hasMore {
		next = reviews[len(reviews)-1].ID
	}
	return c.JSON(fiber.Map{
		"success": true,
		"message": "daftar ulasan berhasil diambil",
		"data":    reviews,
		"meta":    fiber.Map{"limit": want, "has_more": hasMore, "next_before": next},
	})
}

// ---------- GET /courses/:id/reviews/summary ----------
func (s *ReviewService) GetSummary(c *fiber.Ctx) error {
	ctx, cancel := reviewCtx(c)
	defer cancel()

	courseID, _ := c.ParamsInt("id", 0)
	if courseID < 1 {
		return helper.BadRequest("id kelas harus berupa angka positif")
	}
	if _, err := s.courses.FindByID(ctx, courseID); err != nil {
		return translateReviewError(err)
	}
	summary, err := s.reviews.Summary(ctx, courseID)
	if err != nil {
		return helper.Internal(err)
	}
	return reviewJSON(c, fiber.StatusOK, "ringkasan ulasan berhasil diambil", summary)
}

// ---------- GET /reviews/:id ----------
func (s *ReviewService) Get(c *fiber.Ctx) error {
	ctx, cancel := reviewCtx(c)
	defer cancel()

	review, err := s.reviews.FindByID(ctx, c.Params("id"))
	if err != nil {
		return translateReviewError(err)
	}
	return reviewJSON(c, fiber.StatusOK, "ulasan ditemukan", review)
}

// ---------- PATCH /reviews/:id ----------
func (s *ReviewService) Update(c *fiber.Ctx) error {
	ctx, cancel := reviewCtx(c)
	defer cancel()

	var req model.UpdateReviewRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}
	if req.Rating == nil && req.Comment == nil && req.Tags == nil {
		return helper.BadRequest("tidak ada field yang diubah")
	}
	if req.Comment != nil {
		trimmed := strings.TrimSpace(*req.Comment)
		req.Comment = &trimmed
	}
	if req.Tags != nil {
		normalized := NormalizeTags(*req.Tags)
		req.Tags = &normalized
	}
	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	studentID, _, err := s.currentStudent(ctx, c)
	if err != nil {
		return err
	}
	updated, err := s.reviews.UpdateOwn(ctx, c.Params("id"), studentID, req)
	if err != nil {
		return translateReviewError(err)
	}
	return reviewJSON(c, fiber.StatusOK, "ulasan berhasil diubah", updated)
}

// ---------- DELETE /reviews/:id ----------
// Pemilik boleh menghapus ulasannya sendiri; admin boleh menghapus ulasan siapa
// pun (moderasi).
func (s *ReviewService) Delete(c *fiber.Ctx) error {
	ctx, cancel := reviewCtx(c)
	defer cancel()

	authUser, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Forbidden("belum terautentikasi")
	}
	isAdmin := string(authUser.Role) == string(model.RoleAdmin)

	studentID := 0
	if !isAdmin {
		var err error
		studentID, _, err = s.currentStudent(ctx, c)
		if err != nil {
			return err
		}
	}
	if err := s.reviews.Delete(ctx, c.Params("id"), studentID, isAdmin); err != nil {
		return translateReviewError(err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// ---------- POST /reviews/:id/replies ----------
func (s *ReviewService) AddReply(c *fiber.Ctx) error {
	ctx, cancel := reviewCtx(c)
	defer cancel()

	var req model.ReplyRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}
	req.Body = strings.TrimSpace(req.Body)
	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	studentID, studentName, err := s.currentStudent(ctx, c)
	if err != nil {
		return err
	}
	updated, err := s.reviews.AddReply(ctx, c.Params("id"), model.Reply{
		StudentID: studentID, StudentName: studentName, Body: req.Body,
	})
	if err != nil {
		return translateReviewError(err)
	}
	return reviewJSON(c, fiber.StatusCreated, "balasan berhasil ditambahkan", updated)
}

// ---------- POST /reviews/:id/helpful ----------
func (s *ReviewService) MarkHelpful(c *fiber.Ctx) error {
	ctx, cancel := reviewCtx(c)
	defer cancel()

	studentID, _, err := s.currentStudent(ctx, c)
	if err != nil {
		return err
	}
	updated, err := s.reviews.MarkHelpful(ctx, c.Params("id"), studentID)
	if err != nil {
		return translateReviewError(err)
	}
	return reviewJSON(c, fiber.StatusOK, "terima kasih atas penilaiannya", updated)
}

func translateReviewError(err error) error {
	// Error yang sudah berbentuk error HTTP dari helper diteruskan apa adanya.
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
		return helper.BadRequest("id ulasan tidak valid")
	case errors.Is(err, repository.ErrNotFound):
		return helper.NotFound("data tidak ditemukan")
	case errors.Is(err, repository.ErrDuplicate):
		return helper.Conflict("Anda sudah menulis ulasan untuk kelas ini, ubah ulasan yang ada")
	case errors.Is(err, repository.ErrNotOwner):
		return helper.Forbidden("hanya pemilik ulasan yang boleh melakukan ini")
	case errors.Is(err, repository.ErrOwnReview):
		return helper.Conflict("tidak dapat menilai ulasan sendiri")
	case errors.Is(err, repository.ErrAlreadyVoted):
		return helper.Conflict("Anda sudah menilai ulasan ini")
	case errors.Is(err, repository.ErrReplyLimit):
		return helper.Conflict("jumlah balasan sudah mencapai batas")
	default:
		return helper.Internal(err)
	}
}

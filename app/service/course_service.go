package service

import (
	"github.com/gofiber/fiber/v2"

	"siakad-mini/app/repository"
	"siakad-mini/helper"
)

type CourseService struct {
	courses repository.CourseRepository
}

func NewCourseService(courses repository.CourseRepository) *CourseService {
	return &CourseService{courses: courses}
}

// List mengembalikan daftar mata kuliah beserta kuota terisi dan sisa kuota.
func (s *CourseService) List(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	q := helper.ParseCourseListQuery(c)

	courses, err := s.courses.FindAllWithQuota(ctx, q)
	if err != nil {
		return helper.Internal(err)
	}

	return helper.Success(c, fiber.StatusOK, "daftar mata kuliah berhasil diambil", courses)
}

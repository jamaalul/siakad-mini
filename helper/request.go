package helper

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"siakad-mini/app/model"
)

// RequestContext memberi context dengan timeout 5 detik untuk setiap operasi database.
func RequestContext(c *fiber.Ctx) (context.Context, context.CancelFunc) {
	return context.WithTimeout(c.UserContext(), 5*time.Second)
}

// ParamID membaca parameter :id dari jalur URL dan memastikan nilainya angka positif.
func ParamID(c *fiber.Ctx) (int, bool) {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id < 1 {
		return 0, false
	}
	return id, true
}

// ParseStudentListQuery membaca query string untuk daftar mahasiswa dengan nilai bawaan dan batasan yang aman.
func ParseStudentListQuery(c *fiber.Ctx) model.StudentListQuery {
	page := c.QueryInt("page", 1)
	if page < 1 {
		page = 1
	}

	perPage := c.QueryInt("per_page", 10)
	if perPage < 1 {
		perPage = 10
	}
	if perPage > 50 {
		perPage = 50
	}

	return model.StudentListQuery{
		Page:     page,
		PerPage:  perPage,
		Prodi:    strings.TrimSpace(c.Query("prodi")),
		Angkatan: c.QueryInt("angkatan", 0),
		Search:   strings.TrimSpace(c.Query("search")),
		Sort:     strings.TrimSpace(c.Query("sort")),
	}
}

// ParseCourseListQuery membaca query string untuk daftar mata kuliah.
func ParseCourseListQuery(c *fiber.Ctx) model.CourseListQuery {
	q := model.CourseListQuery{
		Semester: c.QueryInt("semester", 0),
		Search:   strings.TrimSpace(c.Query("search")),
	}

	if raw := c.Query("available"); raw != "" {
		if v, err := strconv.ParseBool(raw); err == nil {
			q.Available = &v
		}
	}

	return q
}

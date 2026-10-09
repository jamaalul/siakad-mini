package route

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"siakad-mini/app/model"
	"siakad-mini/app/service"
	"siakad-mini/helper"
	"siakad-mini/middleware"
)

type Dependencies struct {
	Pool              *pgxpool.Pool
	JWT               *helper.JWTManager
	AuthService       *service.AuthService
	StudentService    *service.StudentService
	CourseService     *service.CourseService
	EnrollmentService *service.EnrollmentService
}

func Register(app *fiber.App, deps Dependencies) {
	api := app.Group("/api/v1")

	// -------------------------------------------------------------------------
	// Auth (/api/v1/auth)
	// -------------------------------------------------------------------------
	auth := api.Group("/auth")
	auth.Post("/login", middleware.RequireJSON, middleware.LoginRateLimiter(), deps.AuthService.Login)
	auth.Get("/me", middleware.RequireAuth(deps.JWT), deps.AuthService.Me)

	// -------------------------------------------------------------------------
	// Students (/api/v1/students)
	// -------------------------------------------------------------------------
	students := api.Group("/students", middleware.RequireAuth(deps.JWT))
	students.Get("/", middleware.RequireRole(string(model.RoleAdmin)), deps.StudentService.List)
	students.Post("/", middleware.RequireJSON, middleware.RequireRole(string(model.RoleAdmin)), deps.StudentService.Create)
	students.Get("/:id", deps.StudentService.Get)
	students.Put("/:id", middleware.RequireJSON, middleware.RequireRole(string(model.RoleAdmin)), deps.StudentService.Update)
	students.Delete("/:id", middleware.RequireRole(string(model.RoleAdmin)), deps.StudentService.Delete)

	// -------------------------------------------------------------------------
	// Courses (/api/v1/courses)
	// -------------------------------------------------------------------------
	courses := api.Group("/courses", middleware.RequireAuth(deps.JWT))
	courses.Get("/", deps.CourseService.List)

	// -------------------------------------------------------------------------
	// Enrollments (/api/v1/enrollments)
	// -------------------------------------------------------------------------
	enrollments := api.Group("/enrollments", middleware.RequireAuth(deps.JWT))
	enrollments.Post("/", middleware.RequireJSON, middleware.RequireRole(string(model.RoleMahasiswa)), deps.EnrollmentService.Create)
	enrollments.Delete("/:id", middleware.RequireRole(string(model.RoleMahasiswa)), deps.EnrollmentService.Delete)
}

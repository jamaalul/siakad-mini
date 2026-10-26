package route

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"

	"siakad-mini/app/model"
	"siakad-mini/app/service"
	"siakad-mini/helper"
	"siakad-mini/middleware"
)

type Dependencies struct {
	Pool              *pgxpool.Pool
	MongoClient       *mongo.Client
	JWT               *helper.JWTManager
	AuthService       *service.AuthService
	StudentService    *service.StudentService
	CourseService     *service.CourseService
	EnrollmentService *service.EnrollmentService
	ReviewService     *service.ReviewService
}

func Register(app *fiber.App, deps Dependencies) {
	// Endpoint Health Check dua database
	app.Get("/health", func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.Context(), 3*time.Second)
		defer cancel()

		pgStatus := "ok"
		if err := deps.Pool.Ping(ctx); err != nil {
			pgStatus = "down: " + err.Error()
		}

		mongoStatus := "ok"
		if deps.MongoClient == nil {
			mongoStatus = "down: client belum diinisialisasi"
		} else if err := deps.MongoClient.Ping(ctx, readpref.Primary()); err != nil {
			mongoStatus = "down: " + err.Error()
		}

		if pgStatus != "ok" || mongoStatus != "ok" {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"success": false,
				"message": "satu atau lebih layanan database tidak tersedia",
				"data": fiber.Map{
					"postgres": pgStatus,
					"mongodb":  mongoStatus,
				},
			})
		}

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"success": true,
			"message": "server, PostgreSQL, dan MongoDB berjalan normal",
			"data": fiber.Map{
				"postgres": "ok",
				"mongodb":  "ok",
			},
		})
	})

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
	courses.Get("/:id/reviews", deps.ReviewService.ListByCourse)
	courses.Get("/:id/reviews/summary", deps.ReviewService.GetSummary)
	courses.Post("/:id/reviews",
		middleware.RequireJSON,
		middleware.RequireRole(string(model.RoleMahasiswa)),
		deps.ReviewService.Create)

	// -------------------------------------------------------------------------
	// Reviews (/api/v1/reviews)
	// -------------------------------------------------------------------------
	reviews := api.Group("/reviews", middleware.RequireAuth(deps.JWT))
	reviews.Get("/:id", deps.ReviewService.Get)
	reviews.Patch("/:id", middleware.RequireJSON, middleware.RequireRole(string(model.RoleMahasiswa)), deps.ReviewService.Update)
	reviews.Delete("/:id", deps.ReviewService.Delete) // pemilik atau admin, diputuskan di service
	reviews.Post("/:id/replies", middleware.RequireJSON, middleware.RequireRole(string(model.RoleMahasiswa)), deps.ReviewService.AddReply)
	reviews.Post("/:id/helpful", middleware.RequireRole(string(model.RoleMahasiswa)), deps.ReviewService.MarkHelpful)

	// -------------------------------------------------------------------------
	// Enrollments (/api/v1/enrollments)
	// -------------------------------------------------------------------------
	enrollments := api.Group("/enrollments", middleware.RequireAuth(deps.JWT))
	enrollments.Post("/", middleware.RequireJSON, middleware.RequireRole(string(model.RoleMahasiswa)), deps.EnrollmentService.Create)
	enrollments.Delete("/:id", middleware.RequireRole(string(model.RoleMahasiswa)), deps.EnrollmentService.Delete)
}

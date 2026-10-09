package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"siakad-mini/app/repository"
	"siakad-mini/app/service"
	"siakad-mini/config"
	"siakad-mini/database"
	"siakad-mini/helper"
	"siakad-mini/route"
)

func main() {
	// 1. Muat environment (.env)
	config.LoadEnv()

	// 2. Inisialisasi structured logger (slog + lumberjack)
	logger := config.NewLogger()

	// 3. Buat database connection pool (pgxpool)
	pool, err := database.NewPool(context.Background())
	if err != nil {
		logger.Error("gagal terhubung ke database", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer pool.Close()

	// 4. Instansiasi repositories
	userRepo := repository.NewUserRepository(pool)
	studentRepo := repository.NewStudentRepository(pool)
	courseRepo := repository.NewCourseRepository(pool)
	enrollmentRepo := repository.NewEnrollmentRepository(pool)

	// 5. Inisialisasi JWT Manager
	jwtSecret := config.GetEnv("JWT_SECRET", "super-secret-key-change-in-production")
	jwtIssuer := config.GetEnv("JWT_ISSUER", "siakad-mini")
	jwtTTL, err := time.ParseDuration(config.GetEnv("JWT_ACCESS_TTL", "1h"))
	if err != nil {
		jwtTTL = time.Hour
	}
	jwtManager := helper.NewJWTManager(jwtSecret, jwtIssuer, jwtTTL)

	// 6 - 9. Instansiasi services
	authService := service.NewAuthService(userRepo, studentRepo, jwtManager)
	studentService := service.NewStudentService(studentRepo, userRepo, enrollmentRepo)
	courseService := service.NewCourseService(courseRepo)
	enrollmentService := service.NewEnrollmentService(enrollmentRepo, studentRepo, courseRepo)

	// 10. Pasang dependensi dan buat Fiber app
	deps := route.Dependencies{
		Pool:              pool,
		JWT:               jwtManager,
		AuthService:       authService,
		StudentService:    studentService,
		CourseService:     courseService,
		EnrollmentService: enrollmentService,
	}
	app := config.NewApp(logger, deps)

	// 11. Jalankan HTTP server di latar belakang
	port := config.GetEnv("APP_PORT", "3000")
	go func() {
		if err := app.Listen(":" + port); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server gagal berjalan", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()
	logger.Info("server berjalan", slog.String("port", port))

	// 12. Tangkap sinyal terminasi OS (SIGINT, SIGTERM)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// 13. Graceful shutdown dalam rentang waktu 10 detik
	logger.Info("sinyal berhenti diterima, menutup server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := app.ShutdownWithContext(ctx); err != nil {
		logger.Error("gagal mematikan server secara graceful", slog.String("error", err.Error()))
	}

	// 14. pool.Close() dipanggil via defer pool.Close()
	logger.Info("server dan koneksi database berhenti dengan rapi")
}

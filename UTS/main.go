package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"UTS/app/repository"
	"UTS/app/service"
	"UTS/config"
	"UTS/database"
	"UTS/helper"
	"UTS/route"
)

const minSecretLength = 32

func main() {
	// 1. Inisialisasi Environment & Logger
	config.LoadEnv()
	logger := config.NewLogger()

	jwtSecret := config.GetEnv("JWT_SECRET", "")
	if len(jwtSecret) < minSecretLength {
		logger.Error("JWT_SECRET tidak diisi atau kurang dari 32 karakter",
			slog.Int("minimal_karakter", minSecretLength))
		os.Exit(1)
	}

	// 2. Koneksi Database & Connection Pool
	pool, err := database.NewPool(context.Background())
	if err != nil {
		logger.Error("gagal terhubung ke database", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer pool.Close()

	// 3. Jalankan Migrasi & Seeder Database Otomatis
	logger.Info("menjalankan migrasi dan seeder database...")
	if err := database.RunMigrationsAndSeeds(context.Background(), pool, "migrations"); err != nil {
		logger.Error("gagal menjalankan migrasi/seeder", slog.String("error", err.Error()))
		os.Exit(1)
	}
	logger.Info("migrasi dan seeder database berhasil dijalankan")

	// 4. Inisialisasi JWT Manager
	accessTTLMinutes := config.GetEnvInt("JWT_ACCESS_TTL_MINUTES", 60)
	jwtManager := helper.NewJWTManager(
		jwtSecret,
		config.GetEnv("JWT_ISSUER", "siakad-mini"),
		time.Duration(accessTTLMinutes)*time.Minute,
	)

	// 5. Inisialisasi Repositories
	userRepo := repository.NewUserRepository(pool)
	studentRepo := repository.NewStudentRepository(pool)
	courseRepo := repository.NewCourseRepository(pool)
	enrollmentRepo := repository.NewEnrollmentRepository(pool)

	// 6. Inisialisasi Services
	authService := service.NewAuthService(userRepo, studentRepo, jwtManager)
	studentService := service.NewStudentService(pool, studentRepo, userRepo)
	courseService := service.NewCourseService(courseRepo)
	enrollmentService := service.NewEnrollmentService(pool, enrollmentRepo, courseRepo, studentRepo)

	// 7. Inisialisasi Fiber App & Routes
	app := config.NewApp(logger)
	route.Register(app, route.Dependencies{
		Pool:              pool,
		JWT:               jwtManager,
		AuthService:       authService,
		StudentService:    studentService,
		CourseService:     courseService,
		EnrollmentService: enrollmentService,
	})

	port := config.GetEnv("APP_PORT", "3000")

	go func() {
		if err := app.Listen(":" + port); err != nil {
			logger.Error("server berhenti", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()

	logger.Info("server SIAKAD Mini berhasil berjalan", slog.String("port", port))

	// 8. Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("menerima sinyal berhenti, mematikan server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := app.ShutdownWithContext(ctx); err != nil {
		logger.Error("gagal mematikan server secara bersih", slog.String("error", err.Error()))
	} else {
		logger.Info("server berhasil dimatikan secara bersih")
	}
}

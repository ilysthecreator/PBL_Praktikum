package route

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"UTS/app/service"
	"UTS/helper"
	"UTS/middleware"
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

	// Healthcheck endpoint
	api.Get("/health", func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.UserContext(), 2*time.Second)
		defer cancel()
		if err := deps.Pool.Ping(ctx); err != nil {
			return helper.Fail(c, fiber.StatusServiceUnavailable, "database tidak dapat dihubungi")
		}
		return helper.Success(c, fiber.StatusOK, "server dan database berjalan dengan baik", nil)
	})

	// --- Endpoint 1: Login ---
	auth := api.Group("/auth", middleware.RequireJSON)
	auth.Post("/login", middleware.LoginRateLimiter(), deps.AuthService.Login)

	// Endpoint dengan autentikasi Token
	protected := api.Group("", middleware.RequireAuth(deps.JWT))

	// --- Endpoint 2: Me ---
	protected.Get("/auth/me", deps.AuthService.Me)

	// --- Endpoint Mahasiswa (Students) ---
	students := protected.Group("/students", middleware.RequireJSON)

	// Endpoint 3: GET /api/v1/students (Admin)
	students.Get("/", middleware.RequireRole("admin"), deps.StudentService.List)

	// Endpoint 4: POST /api/v1/students (Admin)
	students.Post("/", middleware.RequireRole("admin"), deps.StudentService.Create)

	// Endpoint 5: GET /api/v1/students/{id} (Admin atau Mahasiswa data sendiri)
	students.Get("/:id", deps.StudentService.Get)

	// Endpoint 6: PUT /api/v1/students/{id} (Admin)
	students.Put("/:id", middleware.RequireRole("admin"), deps.StudentService.Update)

	// Endpoint 7: DELETE /api/v1/students/{id} (Admin)
	students.Delete("/:id", middleware.RequireRole("admin"), deps.StudentService.Delete)

	// --- Endpoint 8: Courses ---
	courses := protected.Group("/courses")
	// GET /api/v1/courses (Semua role)
	courses.Get("/", deps.CourseService.List)

	// --- Endpoint Enrollments (KRS) ---
	enrollments := protected.Group("/enrollments", middleware.RequireJSON)

	// Endpoint 9: POST /api/v1/enrollments (Mahasiswa)
	enrollments.Post("/", middleware.RequireRole("mahasiswa"), deps.EnrollmentService.Create)

	// Endpoint 10: DELETE /api/v1/enrollments/{id} (Mahasiswa milik sendiri)
	enrollments.Delete("/:id", middleware.RequireRole("mahasiswa"), deps.EnrollmentService.Delete)
}

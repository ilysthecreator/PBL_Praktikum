package route

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"api-students/app/service"
	"api-students/helper"
	"api-students/middleware"
)

type Dependencies struct {
	Pool           *pgxpool.Pool
	JWT            *helper.JWTManager
	Permissions    *helper.PermissionSet
	UserService    *service.UserService
	StudentService *service.StudentService
	NilaiService   *service.NilaiService
	AuthService    *service.AuthService
}

func Register(app *fiber.App, deps Dependencies) {
	api := app.Group("/api/v1")

	// --- publik ---
	api.Get("/health", healthCheck(deps.Pool))

	// --- autentikasi ---
	auth := api.Group("/auth", middleware.RequireJSON)
	auth.Post("/register", deps.AuthService.Register)
	auth.Post("/login", middleware.LoginRateLimiter(), deps.AuthService.Login)
	auth.Post("/refresh", deps.AuthService.Refresh)
	auth.Post("/logout", deps.AuthService.Logout)
	auth.Get("/me", middleware.RequireAuth(deps.JWT), deps.AuthService.Me)

	perms := deps.Permissions

	// --- user management (Langkah 6) ---
	// wajib login, hak akses diperiksa per endpoint
	users := api.Group("/users",
		middleware.RequireJSON,
		middleware.RequireAuth(deps.JWT),
	)

	// Hak dapat diputuskan tanpa melihat data -> middleware.
	users.Get("/",
		middleware.RequirePermission(perms, "user:list"),
		deps.UserService.List)
	users.Post("/",
		middleware.RequirePermission(perms, "user:update:any"),
		deps.UserService.Create)
	users.Delete("/:id",
		middleware.RequirePermission(perms, "user:delete"),
		deps.UserService.Delete)
	users.Patch("/:id/role",
		middleware.RequirePermission(perms, "role:assign"),
		deps.UserService.AssignRole)

	// Hak bergantung pada kepemilikan data -> diperiksa di service.
	users.Get("/:id", deps.UserService.Get)
	users.Put("/:id", deps.UserService.Replace)
	users.Patch("/:id", deps.UserService.Patch)

	// --- students ---
	students := api.Group("/students",
		middleware.RequireJSON,
		middleware.RequireAuth(deps.JWT),
	)
	students.Get("/", deps.StudentService.List)
	students.Get("/:id", deps.StudentService.Get)
	students.Post("/", deps.StudentService.Create)
	students.Put("/:id", deps.StudentService.Replace)
	students.Patch("/:id", deps.StudentService.Patch)
	students.Delete("/:id", deps.StudentService.Delete)

	students.Get("/:nim/nilai", deps.NilaiService.GetByNIM)

	// --- nilai ---
	nilai := api.Group("/nilai",
		middleware.RequireJSON,
		middleware.RequireAuth(deps.JWT),
	)
	nilai.Get("/:nim", deps.NilaiService.GetByNIM)
	nilai.Post("/", deps.NilaiService.Create)
}

// healthCheck melaporkan kondisi layanan beserta databasenya.
func healthCheck(pool *pgxpool.Pool) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.UserContext(), 2*time.Second)
		defer cancel()

		if err := pool.Ping(ctx); err != nil {
			return helper.Fail(c, fiber.StatusServiceUnavailable,
				"database tidak dapat dihubungi")
		}

		return helper.Success(c, fiber.StatusOK, "server dan database berjalan", nil)
	}
}

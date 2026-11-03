package route

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"latihan-fiber/app/service"
	"latihan-fiber/helper"
	"latihan-fiber/middleware"
)

type Dependencies struct {
	Pool           *pgxpool.Pool
	JWT            *helper.JWTManager
	Permissions    *helper.PermissionSet
	UserService    *service.UserService
	StudentService *service.StudentService
	NilaiService   *service.NilaiService
	AuthService    *service.AuthService
	CourseService      *service.CourseService
	SupervisorService  *service.SupervisorService
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

	// --- kelas dan KRS ---
	if deps.CourseService != nil {
		courses := api.Group("/courses",
			middleware.RequireJSON,
			middleware.RequireAuth(deps.JWT))

		// "/mine" didaftarkan SEBELUM "/:id". Fiber mencocokkan route sesuai
		// urutan pendaftaran; bila dibalik, "mine" akan dianggap sebagai :id.
		courses.Get("/", deps.CourseService.List)
		courses.Get("/mine", deps.CourseService.Mine)
		courses.Get("/:id", deps.CourseService.Get)
		courses.Post("/:id/enroll", deps.CourseService.Enroll)
		courses.Delete("/:id/enroll", deps.CourseService.Drop)
		courses.Patch("/:id", middleware.RequirePermission(perms, "course:manage"), deps.CourseService.UpdateQuota)

		enrollments := api.Group("/enrollments",
			middleware.RequireJSON,
			middleware.RequireAuth(deps.JWT))
		enrollments.Post("/swap", deps.CourseService.Swap)
	}

	// --- pembimbing tugas akhir (Tugas Mandiri D.2) ---
	if deps.SupervisorService != nil {
		supervisors := api.Group("/supervisors",
			middleware.RequireJSON,
			middleware.RequireAuth(deps.JWT))
		supervisors.Get("/", deps.SupervisorService.List)
		supervisors.Get("/:id", deps.SupervisorService.Get)
		supervisors.Post("/:id/assign", middleware.RequirePermission(perms, "supervision:manage"), deps.SupervisorService.Assign)
		supervisors.Patch("/:id", middleware.RequirePermission(perms, "supervision:manage"), deps.SupervisorService.UpdateQuota)

		supervisions := api.Group("/supervisions",
			middleware.RequireJSON,
			middleware.RequireAuth(deps.JWT))
		supervisions.Post("/switch", middleware.RequirePermission(perms, "supervision:manage"), deps.SupervisorService.Switch)
	}


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

	students := api.Group("/students",
		middleware.RequireJSON,
		middleware.RequireAuth(deps.JWT),
	)

	students.Get("/",
		middleware.RequirePermission(perms, "student:list"),
		deps.StudentService.List)
	students.Post("/",
		middleware.RequirePermission(perms, "student:create"),
		deps.StudentService.Create)
	students.Delete("/:id",
		middleware.RequirePermission(perms, "student:delete"),
		deps.StudentService.Delete)

	students.Get("/:id", deps.StudentService.Get)
	students.Put("/:id", deps.StudentService.Replace)
	students.Patch("/:id", deps.StudentService.Patch)

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
			return helper.ServiceUnavailable("database tidak dapat dihubungi")
		}

		return helper.Success(c, fiber.StatusOK, "server dan database berjalan", nil)
	}
}

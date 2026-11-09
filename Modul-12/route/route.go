package route

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"latihan-fiber/app/service"
	"latihan-fiber/helper"
	"latihan-fiber/middleware"
)

type Dependencies struct {
	Pool                *pgxpool.Pool
	Mongo               *mongo.Client
	JWT                 *helper.JWTManager
	Permissions         *helper.PermissionSet
	UserService         *service.UserService
	StudentService      *service.StudentService
	NilaiService        *service.NilaiService
	AuthService         *service.AuthService
	CourseService       *service.CourseService
	SupervisorService   *service.SupervisorService
	ReviewService       *service.ReviewService
	PortfolioService    *service.PortfolioService
	NotificationService *service.NotificationService
	EnableSwagger       bool
}

func Register(app *fiber.App, deps Dependencies) {
	if deps.EnableSwagger {
		// Swagger UI: http://localhost:3000/swagger/index.html
		app.Get("/swagger/*", swagger.HandlerDefault)
	}

	api := app.Group("/api/v1")

	// --- publik ---
	api.Get("/health", healthCheck(deps.Pool, deps.Mongo))

	// --- autentikasi ---
	auth := api.Group("/auth", middleware.RequireJSON)
	auth.Post("/register", deps.AuthService.Register)
	auth.Post("/login", middleware.LoginRateLimiter(), deps.AuthService.Login)
	auth.Post("/refresh", deps.AuthService.Refresh)
	auth.Post("/logout", deps.AuthService.Logout)
	auth.Get("/me", middleware.RequireAuth(deps.JWT), deps.AuthService.Me)

	perms := deps.Permissions

	// --- kelas dan KRS ---
	if deps.CourseService != nil || deps.ReviewService != nil {
		courses := api.Group("/courses",
			middleware.RequireJSON,
			middleware.RequireAuth(deps.JWT))

		if deps.CourseService != nil {
			// "/mine" didaftarkan SEBELUM "/:id". Fiber mencocokkan route sesuai
			// urutan pendaftaran; bila dibalik, "mine" akan dianggap sebagai :id.
			courses.Get("/", deps.CourseService.List)
			courses.Get("/mine", deps.CourseService.Mine)
			courses.Get("/:id", deps.CourseService.Get)
			courses.Post("/:id/enroll", deps.CourseService.Enroll)
			courses.Delete("/:id/enroll", deps.CourseService.Drop)
			courses.Patch("/:id", middleware.RequirePermission(perms, "course:manage"), deps.CourseService.UpdateQuota)
		}

		if deps.ReviewService != nil {
			courses.Get("/:id/reviews", deps.ReviewService.ListByCourse)
			courses.Get("/:id/reviews/summary", deps.ReviewService.Summary)
			courses.Post("/:id/reviews", deps.ReviewService.Create)
		}

		if deps.CourseService != nil {
			enrollments := api.Group("/enrollments",
				middleware.RequireJSON,
				middleware.RequireAuth(deps.JWT))
			enrollments.Post("/swap", deps.CourseService.Swap)
		}
	}

	// --- ulasan kelas (Modul 10 MongoDB) ---
	if deps.ReviewService != nil {
		reviews := api.Group("/reviews",
			middleware.RequireJSON,
			middleware.RequireAuth(deps.JWT))
		reviews.Get("/:id", deps.ReviewService.Get)
		reviews.Patch("/:id", deps.ReviewService.Update)
		reviews.Delete("/:id", deps.ReviewService.Delete)
		reviews.Post("/:id/replies", deps.ReviewService.Reply)
		reviews.Post("/:id/helpful", deps.ReviewService.MarkHelpful)
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

	// Portofolio Mahasiswa (Modul 10 Tugas Mandiri)
	if deps.PortfolioService != nil {
		students.Get("/:id/portfolio", deps.PortfolioService.Get)
		students.Put("/:id/portfolio", deps.PortfolioService.Upsert)
		students.Post("/:id/portfolio/projects", deps.PortfolioService.AddProject)
		students.Delete("/:id/portfolio/projects/:projectId", deps.PortfolioService.DeleteProject)

		portfolios := api.Group("/portfolios",
			middleware.RequireJSON,
			middleware.RequireAuth(deps.JWT),
		)
		portfolios.Get("/stats/technologies", deps.PortfolioService.GetTopTechnologies)
	}

	// --- nilai ---
	nilai := api.Group("/nilai",
		middleware.RequireJSON,
		middleware.RequireAuth(deps.JWT),
	)
	nilai.Get("/:nim", deps.NilaiService.GetByNIM)
	nilai.Post("/", deps.NilaiService.Create)

	// --- notifikasi (diisi oleh worker, dibaca lewat API) ---
	if deps.NotificationService != nil {
		notifications := api.Group("/notifications",
			middleware.RequireJSON,
			middleware.RequireAuth(deps.JWT))
		notifications.Get("/mine", deps.NotificationService.Mine)

		admin := api.Group("/admin",
			middleware.RequireJSON,
			middleware.RequireAuth(deps.JWT))
		admin.Get("/queue", middleware.RequirePermission(perms, "queue:monitor"), deps.NotificationService.Monitor)
	}
}

// healthCheck kini memeriksa DUA database. Aplikasi dianggap sehat hanya
// bila keduanya dapat dihubungi, dan pesan kegagalannya menyebut yang mana.
func healthCheck(pool *pgxpool.Pool, mongoClient *mongo.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.UserContext(), 2*time.Second)
		defer cancel()

		if pool != nil {
			if err := pool.Ping(ctx); err != nil {
				return helper.ServiceUnavailable("PostgreSQL tidak dapat dihubungi")
			}
		}

		if mongoClient != nil {
			if err := mongoClient.Ping(ctx, nil); err != nil {
				return helper.ServiceUnavailable("MongoDB tidak dapat dihubungi")
			}
		}

		return helper.Success(c, fiber.StatusOK, "server, PostgreSQL, dan MongoDB berjalan", nil)
	}
}

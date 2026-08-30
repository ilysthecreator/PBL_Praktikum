package main

import (
	"context"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"

	"modul3/api-students/app/repository"
	"modul3/api-students/config"
	"modul3/api-students/database"
)

// requireJSON memastikan request memiliki Content-Type application/json untuk POST/PUT/PATCH
func requireJSON(c *fiber.Ctx) error {
	if c.Method() == fiber.MethodPost || c.Method() == fiber.MethodPut || c.Method() == fiber.MethodPatch {
		if c.Get("Content-Type") != "application/json" {
			return fail(c, fiber.StatusBadRequest, "Content-Type wajib application/json")
		}
	}
	return c.Next()
}

func main() {
	config.LoadEnv()

	pool, err := database.NewPool(context.Background())
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer pool.Close()

	studentRepository := repository.NewStudentRepository(pool)
	studentHandler := NewStudentHandler(studentRepository)

	// 4. Aplikasi
	app := fiber.New()

	api := app.Group("/api/v1")

	api.Get("/health", func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.UserContext(), 2*time.Second)
		defer cancel()

		if err := pool.Ping(ctx); err != nil {
			return fail(c, fiber.StatusServiceUnavailable, "database tidak dapat dihubungi")
		}
		return ok(c, "server dan database berjalan", nil)
	})

	// Rute mahasiswa
	s := api.Group("/students", requireJSON)
	s.Get("/", studentHandler.List)
	s.Get("/:id", studentHandler.Get)
	s.Post("/", studentHandler.Create)
	s.Put("/:id", studentHandler.Replace)
	s.Patch("/:id", studentHandler.Patch)
	s.Delete("/:id", studentHandler.Delete)

	port := config.GetEnv("APP_PORT", "3000")
	log.Fatal(app.Listen(":" + port))
}
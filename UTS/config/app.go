package config

import (
	"errors"
	"log/slog"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"

	"UTS/app/model"
)

func NewApp(logger *slog.Logger) *fiber.App {
	isProd := GetEnv("APP_ENV", "development") == "production"

	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			var e *fiber.Error
			if errors.As(err, &e) {
				code = e.Code
			}

			// Cegah bocornya detail error internal di mode produksi
			msg := err.Error()
			if isProd && code == fiber.StatusInternalServerError {
				msg = "Terjadi kesalahan internal pada server"
			}

			logger.Error("HTTP Error",
				slog.Int("status", code),
				slog.String("path", c.Path()),
				slog.String("method", c.Method()),
				slog.String("error", err.Error()),
			)

			return c.Status(code).JSON(model.WebResponse{
				Success: false,
				Message: msg,
			})
		},
	})

	app.Use(recover.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowMethods: "GET, POST, PUT, DELETE, OPTIONS",
	}))

	return app
}

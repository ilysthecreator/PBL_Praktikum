package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"

	"UTS/helper"
)

func RequireJSON(c *fiber.Ctx) error {
	method := c.Method()
	if method == fiber.MethodPost || method == fiber.MethodPut || method == fiber.MethodPatch {
		ct := c.Get(fiber.HeaderContentType)
		if !strings.HasPrefix(strings.ToLower(ct), "application/json") {
			return helper.Fail(c, fiber.StatusUnsupportedMediaType, "Content-Type harus application/json")
		}
	}
	return c.Next()
}

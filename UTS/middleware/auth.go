package middleware

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"

	"UTS/helper"
)

func RequireAuth(jwtManager *helper.JWTManager) fiber.Handler {
	return func(c *fiber.Ctx) error {
		header := c.Get(fiber.HeaderAuthorization)
		if header == "" {
			c.Set("WWW-Authenticate", `Bearer realm="api"`)
			return helper.Fail(c, fiber.StatusUnauthorized, "header Authorization tidak ada atau tidak valid")
		}

		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.Set("WWW-Authenticate", `Bearer realm="api"`)
			return helper.Fail(c, fiber.StatusUnauthorized, "format token harus Bearer")
		}

		tokenStr := strings.TrimSpace(parts[1])
		if tokenStr == "" {
			c.Set("WWW-Authenticate", `Bearer realm="api"`)
			return helper.Fail(c, fiber.StatusUnauthorized, "token kosong")
		}

		claims, err := jwtManager.Parse(tokenStr)
		if err != nil {
			c.Set("WWW-Authenticate", `Bearer realm="api"`)
			if errors.Is(err, helper.ErrExpiredToken) {
				return helper.Fail(c, fiber.StatusUnauthorized, "access token kedaluwarsa")
			}
			return helper.Fail(c, fiber.StatusUnauthorized, "access token tidak valid")
		}

		c.Locals(helper.LocalsAuthUser, claims)
		return c.Next()
	}
}

func RequireRole(allowedRoles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		claims, err := helper.GetAuthUser(c)
		if err != nil {
			return helper.Fail(c, fiber.StatusUnauthorized, "unauthorized")
		}

		for _, role := range allowedRoles {
			if strings.EqualFold(claims.Role, role) {
				return c.Next()
			}
		}

		return helper.Fail(c, fiber.StatusForbidden, "akses ditolak: hak akses tidak memadai")
	}
}

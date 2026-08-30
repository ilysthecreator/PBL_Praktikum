package main

import (
	"context"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"

	"modul3/api-students/app/model"
)

// reqCtx memberi batas waktu untuk setiap operasi basis data (mencegah koneksi menggantung)
func reqCtx(c *fiber.Ctx) (context.Context, context.CancelFunc) {
	return context.WithTimeout(c.UserContext(), 5*time.Second)
}

func ok(c *fiber.Ctx, message string, data interface{}) error {
	return c.Status(fiber.StatusOK).JSON(model.WebResponse{
		Message: message,
		Data:    data,
	})
}

func okList(c *fiber.Ctx, message string, data interface{}, meta *model.Meta) error {
	return c.Status(fiber.StatusOK).JSON(model.WebResponse{
		Message: message,
		Data:    data,
		Meta:    meta,
	})
}

func created(c *fiber.Ctx, message string, data interface{}, location string) error {
	c.Set("Location", location)
	return c.Status(fiber.StatusCreated).JSON(model.WebResponse{
		Message: message,
		Data:    data,
	})
}

func noContent(c *fiber.Ctx) error {
	return c.SendStatus(fiber.StatusNoContent)
}

func fail(c *fiber.Ctx, status int, message string) error {
	return c.Status(status).JSON(model.WebResponse{
		Message: message,
	})
}

func failValidation(c *fiber.Ctx, errors interface{}) error {
	return c.Status(fiber.StatusBadRequest).JSON(model.WebResponse{
		Message: "validasi gagal",
		Errors:  errors,
	})
}

func parseListQuery(c *fiber.Ctx) model.ListQuery {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	if limit < 1 {
		limit = 10
	}

	sort := c.Query("sort", "id")
	order := c.Query("order", "asc")
	search := c.Query("search", "")
	grade := c.Query("grade", "")

	var isActive *bool
	if activeStr := c.Query("is_active"); activeStr != "" {
		parsed, err := strconv.ParseBool(activeStr)
		if err == nil {
			isActive = &parsed
		}
	}

	return model.ListQuery{
		Page:     page,
		Limit:    limit,
		Sort:     sort,
		Order:    order,
		Search:   search,
		IsActive: isActive,
		Grade:    grade,
	}
}

func paramID(c *fiber.Ctx) (int, bool) {
	id, err := c.ParamsInt("id")
	if err != nil || id <= 0 {
		return 0, false
	}
	return id, true
}
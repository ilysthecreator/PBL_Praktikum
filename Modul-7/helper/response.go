package helper

import (
	"github.com/gofiber/fiber/v2"

	"api-students/app/model"
)

// Success mengirim response status OK (atau kustom) dengan data.
func Success(c *fiber.Ctx, status int, message string, data any) error {
	return c.Status(status).JSON(model.WebResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// SuccessList mengirim response daftar data beserta metadata pagination offset.
func SuccessList(c *fiber.Ctx, message string, data any, meta *model.Meta) error {
	return c.Status(fiber.StatusOK).JSON(model.WebResponse{
		Success: true,
		Message: message,
		Data:    data,
		Meta:    meta,
	})
}

// SuccessCursor mengirim response daftar data dengan keyset cursor pagination.
func SuccessCursor(c *fiber.Ctx, message string, data any, meta *model.CursorMeta) error {
	return c.Status(fiber.StatusOK).JSON(model.CursorResponse{
		Success: true,
		Message: message,
		Data:    data,
		Meta:    meta,
	})
}

// Created mengirim 201 Created sekaligus memasang header Location.
func Created(c *fiber.Ctx, message string, data any, location string) error {
	c.Set("Location", location)
	return c.Status(fiber.StatusCreated).JSON(model.WebResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// NoContent mengirim 204 No Content tanpa response body.
func NoContent(c *fiber.Ctx) error {
	return c.SendStatus(fiber.StatusNoContent)
}

// CATATAN: helper.Fail dan helper.FailValidation DIHAPUS TOTAL sesuai Langkah 4 Modul 7
// Seluruh error handling dialihkan ke centralized ErrorHandler dengan mengembalikan AppError.

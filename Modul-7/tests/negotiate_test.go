package tests

import (
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"

	"api-students/app/model"
	"api-students/helper"
)

func TestWriteUsersCSV_NonEmpty(t *testing.T) {
	app := fiber.New()
	app.Get("/users/export", func(c *fiber.Ctx) error {
		users := []model.User{
			{
				ID:        1,
				Username:  "admin",
				Email:     "admin@unair.ac.id",
				Role:      "admin",
				IsActive:  true,
				CreatedAt: time.Now().UTC(),
			},
			{
				ID:        2,
				Username:  "fadhil",
				Email:     "fadhil@unair.ac.id",
				Role:      "user",
				IsActive:  true,
				CreatedAt: time.Now().UTC(),
			},
		}
		return helper.WriteUsersCSV(c, users)
	})

	req := httptest.NewRequest("GET", "/users/export", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test failed: %v", err)
	}

	if resp.StatusCode != 200 {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	buf := make([]byte, 1024)
	n, _ := resp.Body.Read(buf)
	body := string(buf[:n])

	if !strings.Contains(body, "id,username,email,role,is_active,created_at") {
		t.Errorf("header CSV tidak ditemukan: %s", body)
	}
	if !strings.Contains(body, "admin,admin@unair.ac.id") {
		t.Errorf("data user tidak ditemukan dalam CSV: %s", body)
	}
}

func TestNegotiateAcceptHeader(t *testing.T) {
	app := fiber.New()

	app.Get("/negotiate", func(c *fiber.Ctx) error {
		chosen, err := helper.Negotiate(c, helper.FormatJSON, helper.FormatCSV)
		if err != nil {
			return err
		}
		return c.SendString(chosen)
	})

	// 1. Accept: text/csv -> format CSV
	reqCSV := httptest.NewRequest("GET", "/negotiate", nil)
	reqCSV.Header.Set("Accept", "text/csv")
	respCSV, _ := app.Test(reqCSV)
	bufCSV := make([]byte, 64)
	n, _ := respCSV.Body.Read(bufCSV)
	if string(bufCSV[:n]) != helper.FormatCSV {
		t.Errorf("expected text/csv, got %s", string(bufCSV[:n]))
	}

	// 2. Accept: */* -> default JSON
	reqAny := httptest.NewRequest("GET", "/negotiate", nil)
	reqAny.Header.Set("Accept", "*/*")
	respAny, _ := app.Test(reqAny)
	bufAny := make([]byte, 64)
	nAny, _ := respAny.Body.Read(bufAny)
	if string(bufAny[:nAny]) != helper.FormatJSON {
		t.Errorf("expected default JSON for */*, got %s", string(bufAny[:nAny]))
	}
}

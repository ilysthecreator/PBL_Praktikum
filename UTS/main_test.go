package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"UTS/app/model"
	"UTS/app/repository"
	"UTS/app/service"
	"UTS/config"
	"UTS/database"
	"UTS/helper"
	"UTS/middleware"
	"UTS/route"
)

func setupTestApp(t *testing.T) (*fiber.App, *pgxpool.Pool, *helper.JWTManager) {
	config.LoadEnv()
	logger := config.NewLogger()

	jwtSecret := config.GetEnv("JWT_SECRET", "2441502a8d96bad551072bc97083cd7e16a90d09f056c43330b9ed9c4c1512f2")
	jwtManager := helper.NewJWTManager(jwtSecret, "siakad-mini-test", 60*time.Minute)

	pool, err := database.NewPool(context.Background())
	if err != nil {
		t.Fatalf("Gagal terhubung ke database test: %v", err)
	}

	// Migrasi & seeder
	if err := database.RunMigrationsAndSeeds(context.Background(), pool, "migrations"); err != nil {
		t.Fatalf("Gagal menjalankan migrasi/seeder: %v", err)
	}

	userRepo := repository.NewUserRepository(pool)
	studentRepo := repository.NewStudentRepository(pool)
	courseRepo := repository.NewCourseRepository(pool)
	enrollmentRepo := repository.NewEnrollmentRepository(pool)

	authService := service.NewAuthService(userRepo, studentRepo, jwtManager)
	studentService := service.NewStudentService(pool, studentRepo, userRepo)
	courseService := service.NewCourseService(courseRepo)
	enrollmentService := service.NewEnrollmentService(pool, enrollmentRepo, courseRepo, studentRepo)

	app := config.NewApp(logger)
	route.Register(app, route.Dependencies{
		Pool:              pool,
		JWT:               jwtManager,
		AuthService:       authService,
		StudentService:    studentService,
		CourseService:     courseService,
		EnrollmentService: enrollmentService,
	})

	return app, pool, jwtManager
}

func loginHelper(t *testing.T, app *fiber.App, email, password string) string {
	payload, _ := json.Marshal(map[string]string{
		"email":    email,
		"password": password,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("Gagal request login: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Login gagal dengan status %d", resp.StatusCode)
	}

	var res struct {
		Success bool                `json:"success"`
		Data    model.LoginResponse `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		t.Fatalf("Gagal decode response login: %v", err)
	}
	return res.Data.AccessToken
}

func Test10Endpoints(t *testing.T) {
	app, pool, _ := setupTestApp(t)
	defer pool.Close()

	// 1. Test Login Admin & Mahasiswa (Endpoint 1)
	adminToken := loginHelper(t, app, "admin@siakad.ac.id", "admin12345")
	if adminToken == "" {
		t.Fatal("Admin token kosong")
	}

	mhsToken := loginHelper(t, app, "rina.putri@siakad.ac.id", "187221000001")
	if mhsToken == "" {
		t.Fatal("Mahasiswa token kosong")
	}

	// Test Login Validation Error 422
	{
		invalidPayload, _ := json.Marshal(map[string]string{
			"email":    "notanemail",
			"password": "123",
		})
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(invalidPayload))
		req.Header.Set("Content-Type", "application/json")
		resp, _ := app.Test(req, -1)
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Errorf("Harusnya 422 untuk validasi login, didapat: %d", resp.StatusCode)
		}
	}

	// 2. Test GET /api/v1/auth/me (Endpoint 2)
	{
		req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
		req.Header.Set("Authorization", "Bearer "+mhsToken)
		resp, _ := app.Test(req, -1)
		if resp.StatusCode != http.StatusOK {
			t.Errorf("GET /auth/me gagal, status: %d", resp.StatusCode)
		}
		var meRes struct {
			Success bool                      `json:"success"`
			Data    model.UserProfileResponse `json:"data"`
		}
		json.NewDecoder(resp.Body).Decode(&meRes)
		if meRes.Data.Student == nil || meRes.Data.Student.NIM != "187221000001" {
			t.Errorf("Data student tidak lengkap di /auth/me")
		}
	}

	// 3. Test GET /api/v1/students (Endpoint 3)
	{
		// Harus 403 jika role mahasiswa
		reqMhs := httptest.NewRequest(http.MethodGet, "/api/v1/students", nil)
		reqMhs.Header.Set("Authorization", "Bearer "+mhsToken)
		respMhs, _ := app.Test(reqMhs, -1)
		if respMhs.StatusCode != http.StatusForbidden {
			t.Errorf("Harusnya 403 untuk mahasiswa akses GET /students, didapat: %d", respMhs.StatusCode)
		}

		// Harus 200 jika admin
		reqAdmin := httptest.NewRequest(http.MethodGet, "/api/v1/students?page=1&per_page=5&sort=nama", nil)
		reqAdmin.Header.Set("Authorization", "Bearer "+adminToken)
		respAdmin, _ := app.Test(reqAdmin, -1)
		if respAdmin.StatusCode != http.StatusOK {
			t.Errorf("Admin gagal akses GET /students, status: %d", respAdmin.StatusCode)
		}
	}

	// 4. Test POST /api/v1/students (Endpoint 4)
	var newStudentID int
	newNIM := "187221999999"
	{
		newMhsPayload, _ := json.Marshal(map[string]any{
			"nim":          newNIM,
			"nama":         "Siswa Baru",
			"email":        "siswa.baru@siakad.ac.id",
			"prodi":        "Teknik Komputer",
			"angkatan":     2024,
			"ipk_terakhir": 3.75,
		})
		req := httptest.NewRequest(http.MethodPost, "/api/v1/students", bytes.NewReader(newMhsPayload))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+adminToken)
		resp, _ := app.Test(req, -1)
		if resp.StatusCode != http.StatusCreated {
			t.Errorf("Gagal membuat mahasiswa baru, status: %d", resp.StatusCode)
		}

		var createdRes struct {
			Success bool                      `json:"success"`
			Data    model.StudentItemResponse `json:"data"`
		}
		json.NewDecoder(resp.Body).Decode(&createdRes)
		newStudentID = createdRes.Data.ID

		// Test Login dengan password baru = NIM
		newMhsToken := loginHelper(t, app, "siswa.baru@siakad.ac.id", newNIM)
		if newMhsToken == "" {
			t.Errorf("Mahasiswa baru gagal login dengan default password NIM")
		}
	}

	// 5. Test GET /api/v1/students/{id} (Endpoint 5)
	{
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/students/%d", newStudentID), nil)
		req.Header.Set("Authorization", "Bearer "+adminToken)
		resp, _ := app.Test(req, -1)
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Gagal GET /students/{id}, status: %d", resp.StatusCode)
		}

		var detailRes struct {
			Success bool                        `json:"success"`
			Data    model.StudentDetailResponse `json:"data"`
		}
		json.NewDecoder(resp.Body).Decode(&detailRes)
		if detailRes.Data.BatasSKS != 24 { // IPK 3.75 -> Batas SKS 24
			t.Errorf("Batas SKS salah, diharapkan 24 didapat: %d", detailRes.Data.BatasSKS)
		}
	}

	// 6. Test PUT /api/v1/students/{id} (Endpoint 6)
	{
		updatePayload, _ := json.Marshal(map[string]any{
			"nama":         "Siswa Baru Terupdate",
			"prodi":        "Teknik Komputer",
			"angkatan":     2024,
			"ipk_terakhir": 3.80,
		})
		req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/students/%d", newStudentID), bytes.NewReader(updatePayload))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+adminToken)
		resp, _ := app.Test(req, -1)
		if resp.StatusCode != http.StatusOK {
			t.Errorf("PUT /students/{id} gagal, status: %d", resp.StatusCode)
		}
	}

	// 7. Test GET /api/v1/courses (Endpoint 8)
	var firstCourseID int
	{
		req := httptest.NewRequest(http.MethodGet, "/api/v1/courses?available=true", nil)
		req.Header.Set("Authorization", "Bearer "+mhsToken)
		resp, _ := app.Test(req, -1)
		if resp.StatusCode != http.StatusOK {
			t.Errorf("GET /courses gagal, status: %d", resp.StatusCode)
		}
		var courseRes struct {
			Success bool           `json:"success"`
			Data    []model.Course `json:"data"`
		}
		json.NewDecoder(resp.Body).Decode(&courseRes)
		if len(courseRes.Data) == 0 {
			t.Errorf("Daftar mata kuliah kosong")
		} else {
			firstCourseID = courseRes.Data[0].ID
		}
	}

	// 8. Test POST /api/v1/enrollments (Endpoint 9) & Duplikasi 409
	var enrollmentID int
	{
		enrollPayload, _ := json.Marshal(map[string]any{
			"course_id":      firstCourseID,
			"tahun_akademik": "2026/2027-Ganjil",
		})
		req := httptest.NewRequest(http.MethodPost, "/api/v1/enrollments", bytes.NewReader(enrollPayload))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+mhsToken)
		resp, _ := app.Test(req, -1)
		if resp.StatusCode != http.StatusCreated {
			t.Errorf("POST /enrollments gagal, status: %d", resp.StatusCode)
		}

		var enrollRes struct {
			Success bool                     `json:"success"`
			Data    model.EnrollmentResponse `json:"data"`
		}
		json.NewDecoder(resp.Body).Decode(&enrollRes)
		enrollmentID = enrollRes.Data.ID

		// Coba ambil lagi mata kuliah yang sama di tahun akademik yang sama -> Harus 409 Conflict
		reqDup := httptest.NewRequest(http.MethodPost, "/api/v1/enrollments", bytes.NewReader(enrollPayload))
		reqDup.Header.Set("Content-Type", "application/json")
		reqDup.Header.Set("Authorization", "Bearer "+mhsToken)
		respDup, _ := app.Test(reqDup, -1)
		if respDup.StatusCode != http.StatusConflict {
			t.Errorf("Harusnya 409 saat ambil mata kuliah sama dua kali, didapat: %d", respDup.StatusCode)
		}
	}

	// 9. Test DELETE /api/v1/enrollments/{id} (Endpoint 10)
	{
		// Coba batalkan pakai akun admin -> 403 Forbidden
		reqAdmin := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/enrollments/%d", enrollmentID), nil)
		reqAdmin.Header.Set("Authorization", "Bearer "+adminToken)
		respAdmin, _ := app.Test(reqAdmin, -1)
		if respAdmin.StatusCode != http.StatusForbidden {
			t.Errorf("Admin tidak boleh batalkan enrollment mahasiswa (harus 403), didapat: %d", respAdmin.StatusCode)
		}

		// Batalkan dengan pemilik enrollment -> 204 No Content
		reqMhs := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/enrollments/%d", enrollmentID), nil)
		reqMhs.Header.Set("Authorization", "Bearer "+mhsToken)
		respMhs, _ := app.Test(reqMhs, -1)
		if respMhs.StatusCode != http.StatusNoContent {
			t.Errorf("Mahasiswa gagal batalkan enrollment, status: %d", respMhs.StatusCode)
		}
	}

	// 10. Test DELETE /api/v1/students/{id} (Endpoint 7 - Soft delete)
	{
		req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/students/%d", newStudentID), nil)
		req.Header.Set("Authorization", "Bearer "+adminToken)
		resp, _ := app.Test(req, -1)
		if resp.StatusCode != http.StatusNoContent {
			t.Errorf("Soft delete mahasiswa gagal, status: %d", resp.StatusCode)
		}

		// Verifikasi mahasiswa yang dihapus tidak bisa login lagi -> 401
		loginPayload, _ := json.Marshal(map[string]string{
			"email":    "siswa.baru@siakad.ac.id",
			"password": newNIM,
		})
		reqLogin := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(loginPayload))
		reqLogin.Header.Set("Content-Type", "application/json")
		respLogin, _ := app.Test(reqLogin, -1)
		if respLogin.StatusCode != http.StatusUnauthorized {
			t.Errorf("Mahasiswa yang di-soft-delete harusnya tidak bisa login (401), didapat: %d", respLogin.StatusCode)
		}
	}

	// 11. Test Rate Limiter 429
	{
		for i := 0; i < 6; i++ {
			middleware.GlobalLoginTracker.RecordFailure("192.168.1.100")
		}
		if !middleware.GlobalLoginTracker.IsBlocked("192.168.1.100") {
			t.Errorf("Harusnya IP terblokir setelah > 5 percobaan gagal")
		}
	}
}

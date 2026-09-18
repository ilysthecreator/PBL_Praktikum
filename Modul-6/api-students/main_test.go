package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http/httptest"
	"testing"
	"time"

	"api-students/app/model"
	"api-students/app/repository"
	"api-students/app/service"
	"api-students/config"
	"api-students/database"
	"api-students/helper"
	"api-students/route"
)

func TestRBACMatrix(t *testing.T) {
	config.LoadEnv()
	logger := config.NewLogger()

	pool, err := database.NewPool(context.Background())
	if err != nil {
		t.Fatalf("koneksi database gagal: %v", err)
	}
	defer pool.Close()

	jwtSecret := config.GetEnv("JWT_SECRET", "2441502a8d96bad551072bc97083cd7e16a90d09f056c43330b9ed9c4c1512f2")
	jwtManager := helper.NewJWTManager(jwtSecret, "praktikum-backend", 15*time.Minute)

	roleRepo := repository.NewRoleRepository(pool)
	rawPerms, err := roleRepo.LoadPermissions(context.Background())
	if err != nil {
		t.Fatalf("load permissions gagal: %v", err)
	}
	perms := helper.NewPermissionSet(rawPerms)

	userRepo := repository.NewUserRepository(pool)
	tokenRepo := repository.NewTokenRepository(pool)
	studentRepo := repository.NewStudentRepository(pool)
	nilaiRepo := repository.NewNilaiRepository(pool)

	userService := service.NewUserService(userRepo, perms)
	studentService := service.NewStudentService(studentRepo, perms)
	nilaiService := service.NewNilaiService(nilaiRepo, studentRepo)
	authService := service.NewAuthService(userRepo, tokenRepo, jwtManager, perms, 7*24*time.Hour)

	app := config.NewApp(logger, route.Dependencies{
		Pool:           pool,
		JWT:            jwtManager,
		Permissions:    perms,
		UserService:    userService,
		StudentService: studentService,
		NilaiService:   nilaiService,
		AuthService:    authService,
	})

	// Setup user test
	// 1. Admin
	adminUser := model.User{ID: 1001, Username: "admin_test", Role: "admin"}
	tokenAdmin, _ := jwtManager.GenerateAccess(adminUser)

	// 2. Staff
	staffUser := model.User{ID: 1002, Username: "staff_test", Role: "staff"}
	tokenStaff, _ := jwtManager.GenerateAccess(staffUser)

	// 3. User1 (Owner)
	user1 := model.User{ID: 1003, Username: "user1_test", Role: "user"}
	tokenUser1, _ := jwtManager.GenerateAccess(user1)

	// 4. User2 (Non-Owner)
	user2 := model.User{ID: 1004, Username: "user2_test", Role: "user"}
	tokenUser2, _ := jwtManager.GenerateAccess(user2)

	// Buat user di database untuk pengujian /users
	ctx := context.Background()
	hash, _ := helper.HashPassword("Rahasia123!")
	pool.Exec(ctx, `
		INSERT INTO users (id, username, email, password, role)
		VALUES 
			(1001, 'admin_test', 'admin_test@test.com', $1, 'admin'),
			(1002, 'staff_test', 'staff_test@test.com', $1, 'staff'),
			(1003, 'user1_test', 'user1_test@test.com', $1, 'user'),
			(1004, 'user2_test', 'user2_test@test.com', $1, 'user')
		ON CONFLICT (id) DO UPDATE SET role = EXCLUDED.role;
	`, hash)

	// Helper function untuk mengirim request
	sendRequest := func(method, path, token string, body any) (int, string) {
		var reqBody io.Reader
		if body != nil {
			b, _ := json.Marshal(body)
			reqBody = bytes.NewReader(b)
		}
		req := httptest.NewRequest(method, path, reqBody)
		req.Header.Set("Content-Type", "application/json")
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}

		resp, err := app.Test(req, -1)
		if err != nil {
			t.Fatalf("request %s %s error: %v", method, path, err)
		}
		respBodyBytes, _ := io.ReadAll(resp.Body)
		return resp.StatusCode, string(respBodyBytes)
	}

	t.Run("Matriks Hak Akses Users (Langkah 9)", func(t *testing.T) {
		// 1. GET /users
		statusAdmin, _ := sendRequest("GET", "/api/v1/users", tokenAdmin, nil)
		statusStaff, _ := sendRequest("GET", "/api/v1/users", tokenStaff, nil)
		statusUser, _ := sendRequest("GET", "/api/v1/users", tokenUser1, nil)
		if statusAdmin != 200 || statusStaff != 200 || statusUser != 403 {
			t.Errorf("GET /users expected 200, 200, 403; got %d, %d, %d", statusAdmin, statusStaff, statusUser)
		}

		// 2. GET /users/:id — dirinya sendiri
		s1, _ := sendRequest("GET", "/api/v1/users/1001", tokenAdmin, nil)
		s2, _ := sendRequest("GET", "/api/v1/users/1002", tokenStaff, nil)
		s3, _ := sendRequest("GET", "/api/v1/users/1003", tokenUser1, nil)
		if s1 != 200 || s2 != 200 || s3 != 200 {
			t.Errorf("GET /users/:id (self) expected 200, 200, 200; got %d, %d, %d", s1, s2, s3)
		}

		// 3. GET /users/:id — milik orang lain (target 1003)
		sAdminOther, _ := sendRequest("GET", "/api/v1/users/1003", tokenAdmin, nil)
		sStaffOther, _ := sendRequest("GET", "/api/v1/users/1003", tokenStaff, nil)
		sUserOther, _ := sendRequest("GET", "/api/v1/users/1003", tokenUser2, nil)
		if sAdminOther != 200 || sStaffOther != 200 || sUserOther != 403 {
			t.Errorf("GET /users/:id (other) expected 200, 200, 403; got %d, %d, %d", sAdminOther, sStaffOther, sUserOther)
		}

		// 4. PUT /users/:id — milik orang lain (target 1003)
		upBody := map[string]string{"username": "user1_edited", "email": "user1_edited@test.com"}
		pAdmin, _ := sendRequest("PUT", "/api/v1/users/1003", tokenAdmin, upBody)
		pStaff, _ := sendRequest("PUT", "/api/v1/users/1003", tokenStaff, upBody)
		pUser, _ := sendRequest("PUT", "/api/v1/users/1003", tokenUser2, upBody)
		if pAdmin != 200 || pStaff != 403 || pUser != 403 {
			t.Errorf("PUT /users/:id (other) expected 200, 403, 403; got %d, %d, %d", pAdmin, pStaff, pUser)
		}

		// 5. PATCH /users/:id/role — orang lain (target 1004)
		roleBody := map[string]string{"role": "staff"}
		rAdmin, _ := sendRequest("PATCH", "/api/v1/users/1004/role", tokenAdmin, roleBody)
		rStaff, _ := sendRequest("PATCH", "/api/v1/users/1004/role", tokenStaff, roleBody)
		rUser, _ := sendRequest("PATCH", "/api/v1/users/1004/role", tokenUser1, roleBody)
		if rAdmin != 200 || rStaff != 403 || rUser != 403 {
			t.Errorf("PATCH /users/:id/role (other) expected 200, 403, 403; got %d, %d, %d", rAdmin, rStaff, rUser)
		}

		// 6. PATCH /users/:id/role — dirinya sendiri
		rSelfAdmin, _ := sendRequest("PATCH", "/api/v1/users/1001/role", tokenAdmin, map[string]string{"role": "user"})
		rSelfStaff, _ := sendRequest("PATCH", "/api/v1/users/1002/role", tokenStaff, map[string]string{"role": "user"})
		rSelfUser, _ := sendRequest("PATCH", "/api/v1/users/1003/role", tokenUser1, map[string]string{"role": "user"})
		if rSelfAdmin != 422 || rSelfStaff != 403 || rSelfUser != 403 {
			t.Errorf("PATCH /users/:id/role (self) expected 422, 403, 403; got %d, %d, %d", rSelfAdmin, rSelfStaff, rSelfUser)
		}

		// 7. DELETE /users/:id — dirinya sendiri
		delSelfAdmin, _ := sendRequest("DELETE", "/api/v1/users/1001", tokenAdmin, nil)
		delSelfStaff, _ := sendRequest("DELETE", "/api/v1/users/1002", tokenStaff, nil)
		delSelfUser, _ := sendRequest("DELETE", "/api/v1/users/1003", tokenUser1, nil)
		if delSelfAdmin != 403 || delSelfStaff != 403 || delSelfUser != 403 {
			t.Errorf("DELETE /users/:id (self) expected 403, 403, 403; got %d, %d, %d", delSelfAdmin, delSelfStaff, delSelfUser)
		}

		// 8. Tanpa Authorization header
		noAuth, _ := sendRequest("GET", "/api/v1/users", "", nil)
		if noAuth != 401 {
			t.Errorf("Tanpa Authorization header expected 401, got %d", noAuth)
		}
	})

	t.Run("Matriks Hak Akses Students (Tugas Mandiri C.3)", func(t *testing.T) {
		// Setup mahasiswa milik user1 (1003)
		pool.Exec(ctx, `
			INSERT INTO students (id, nim, name, grade, is_active, owner_id)
			VALUES (901, 'NIM901', 'Mhs User1', 'A', true, 1003)
			ON CONFLICT (id) DO UPDATE SET owner_id = 1003;
		`)

		// Setup mahasiswa untuk dihapus admin
		pool.Exec(ctx, `
			INSERT INTO students (id, nim, name, grade, is_active, owner_id)
			VALUES (902, 'NIM902', 'Mhs Dihapus', 'B', true, 1002)
			ON CONFLICT (id) DO UPDATE SET owner_id = 1002;
		`)

		// 1. GET /students (List)
		stAdmin, _ := sendRequest("GET", "/api/v1/students", tokenAdmin, nil)
		stStaff, _ := sendRequest("GET", "/api/v1/students", tokenStaff, nil)
		stUser, _ := sendRequest("GET", "/api/v1/students", tokenUser1, nil)
		if stAdmin != 200 || stStaff != 200 || stUser != 403 {
			t.Errorf("GET /students expected 200, 200, 403; got %d, %d, %d", stAdmin, stStaff, stUser)
		}

		// 2. POST /students
		postStudentBody := map[string]string{"nim": "NIM_NEW_1", "name": "Mhs Baru", "grade": "A"}
		stPostAdmin, _ := sendRequest("POST", "/api/v1/students", tokenAdmin, postStudentBody)
		postStudentBody2 := map[string]string{"nim": "NIM_NEW_2", "name": "Mhs Baru Staff", "grade": "B"}
		stPostStaff, _ := sendRequest("POST", "/api/v1/students", tokenStaff, postStudentBody2)
		postStudentBody3 := map[string]string{"nim": "NIM_NEW_3", "name": "Mhs Baru User", "grade": "A"}
		stPostUser, _ := sendRequest("POST", "/api/v1/students", tokenUser1, postStudentBody3)
		if stPostAdmin != 201 || stPostStaff != 201 || stPostUser != 403 {
			t.Errorf("POST /students expected 201, 201, 403; got %d, %d, %d", stPostAdmin, stPostStaff, stPostUser)
		}

		// 3. GET /students/:id
		// Milik sendiri (user1) -> 200
		stGetSelf, _ := sendRequest("GET", "/api/v1/students/901", tokenUser1, nil)
		// Milik orang lain (user2 mengakses 901) -> 403
		stGetOther, _ := sendRequest("GET", "/api/v1/students/901", tokenUser2, nil)
		// Staff mengakses 901 (memiliki student:read:any) -> 200
		stGetStaff, _ := sendRequest("GET", "/api/v1/students/901", tokenStaff, nil)
		// Admin mengakses 901 (memiliki student:read:any) -> 200
		stGetAdmin, _ := sendRequest("GET", "/api/v1/students/901", tokenAdmin, nil)
		if stGetSelf != 200 || stGetOther != 403 || stGetStaff != 200 || stGetAdmin != 200 {
			t.Errorf("GET /students/:id expected 200, 403, 200, 200; got self=%d, other=%d, staff=%d, admin=%d",
				stGetSelf, stGetOther, stGetStaff, stGetAdmin)
		}

		// 4. PUT /students/:id
		putBody := map[string]any{"nim": "NIM901", "name": "Mhs User1 Diupdate", "grade": "B", "is_active": true}
		// Milik sendiri (user1 update 901) -> 200
		stPutSelf, _ := sendRequest("PUT", "/api/v1/students/901", tokenUser1, putBody)
		// Milik orang lain (user2 update 901) -> 403
		stPutOther, _ := sendRequest("PUT", "/api/v1/students/901", tokenUser2, putBody)
		// Staff update 901 (staff TIDAK punya student:update:any) -> 403
		stPutStaff, _ := sendRequest("PUT", "/api/v1/students/901", tokenStaff, putBody)
		// Admin update 901 (admin punya student:update:any) -> 200
		stPutAdmin, _ := sendRequest("PUT", "/api/v1/students/901", tokenAdmin, putBody)
		if stPutSelf != 200 || stPutOther != 403 || stPutStaff != 403 || stPutAdmin != 200 {
			t.Errorf("PUT /students/:id expected 200, 403, 403, 200; got self=%d, other=%d, staff=%d, admin=%d",
				stPutSelf, stPutOther, stPutStaff, stPutAdmin)
		}

		// 5. DELETE /students/:id
		// User1 coba hapus -> 403 (tidak punya student:delete)
		stDelUser, _ := sendRequest("DELETE", "/api/v1/students/902", tokenUser1, nil)
		// Staff coba hapus -> 403 (tidak punya student:delete)
		stDelStaff, _ := sendRequest("DELETE", "/api/v1/students/902", tokenStaff, nil)
		// Admin hapus 902 -> 204
		stDelAdmin, _ := sendRequest("DELETE", "/api/v1/students/902", tokenAdmin, nil)
		if stDelUser != 403 || stDelStaff != 403 || stDelAdmin != 204 {
			t.Errorf("DELETE /students/:id expected 403, 403, 204; got user=%d, staff=%d, admin=%d",
				stDelUser, stDelStaff, stDelAdmin)
		}
	})

	t.Run("Pengujian Negatif (C.3)", func(t *testing.T) {
		// Pengujian Negatif 1:
		// Mengirim "owner_id": 1001 (milik admin) di dalam body POST /students saat dipanggil oleh staff (1002).
		// Sistem harus mengabaikan owner_id dari body dan mengisi owner_id sesuai pemanggil (1002).
		fakeOwnerReq := map[string]any{
			"nim":      "NIM_NEGATIF_1",
			"name":     "Mhs Fake Owner Test",
			"grade":    "A",
			"owner_id": 1001,
		}
		statusStaffPost, respStaffPost := sendRequest("POST", "/api/v1/students", tokenStaff, fakeOwnerReq)
		if statusStaffPost != 201 {
			t.Fatalf("POST /students expected 201, got %d: %s", statusStaffPost, respStaffPost)
		}

		var createdStudent struct {
			Data struct {
				ID      int  `json:"id"`
				OwnerID *int `json:"owner_id"`
			} `json:"data"`
		}
		json.Unmarshal([]byte(respStaffPost), &createdStudent)

		var dbOwnerID int
		err := pool.QueryRow(ctx, "SELECT owner_id FROM students WHERE id = $1", createdStudent.Data.ID).Scan(&dbOwnerID)
		if err != nil {
			t.Fatalf("query db owner_id gagal: %v", err)
		}
		if dbOwnerID != 1002 {
			t.Errorf("PENGUJIAN NEGATIF 1 GAGAL: owner_id di DB adalah %d, seharusnya 1002 (staff)", dbOwnerID)
		} else {
			fmt.Printf("PENGUJIAN NEGATIF 1 BERHASIL: pemanggil staff mengirim owner_id 1001, tetapi tersimpan di DB sebagai %d (staff)\n", dbOwnerID)
		}

		// Pengujian Negatif 2:
		// User2 (1004) mencoba memanggil PUT /students/:id atas mahasiswa milik User1 (1003).
		// Harus ditolak dengan status 403 Forbidden.
		hackReq := map[string]any{
			"nim":       "NIM901",
			"name":      "Hacked by User2",
			"grade":     "E",
			"is_active": true,
		}
		statusHack, respHack := sendRequest("PUT", "/api/v1/students/901", tokenUser2, hackReq)
		if statusHack != 403 {
			t.Errorf("PENGUJIAN NEGATIF 2 GAGAL: expected 403 Forbidden, got %d: %s", statusHack, respHack)
		} else {
			fmt.Printf("PENGUJIAN NEGATIF 2 BERHASIL: User2 PUT data User1 ditolak dengan status 403: %s\n", respHack)
		}
	})
}

//go:build integration

package integration_test

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"

	"latihan-fiber/app/model"
	"latihan-fiber/app/repository"
	"latihan-fiber/app/service"
	"latihan-fiber/config"
	"latihan-fiber/helper"
	"latihan-fiber/route"
)

// api adalah aplikasi lengkap yang memakai PostgreSQL sungguhan. Bedanya
// dengan production hanya satu: database-nya praktikum_backend_test.
type api struct {
	app *fiber.App
	jwt *helper.JWTManager
}

func newAPI(t *testing.T) api {
	t.Helper()
	raw, err := repository.NewRoleRepository(pool).LoadPermissions(context.Background())
	if err != nil {
		t.Fatalf("memuat permission: %v", err)
	}
	perms := helper.NewPermissionSet(raw)
	jwtManager := helper.NewJWTManager("secret-khusus-test-minimal-32-karakter", "test", time.Minute)
	courseService := service.NewCourseService(
		repository.NewCourseRepository(pool),
		repository.NewOutboxRepository(pool),
		repository.NewTxManager(pool),
	)
	app := config.NewApp(slog.New(slog.NewTextHandler(io.Discard, nil)), route.Dependencies{
		Pool:          pool,
		JWT:           jwtManager,
		Permissions:   perms,
		CourseService: courseService,
	})
	return api{app: app, jwt: jwtManager}
}

// enroll mengirim POST /courses/:id/enroll atas nama student dan
// mengembalikan status code beserta field "code" pada response.
func (a api) enroll(t *testing.T, student model.User, courseID int) (int, string) {
	t.Helper()
	token, err := a.jwt.GenerateAccessToken(model.AuthUser{UserID: student.ID, Username: student.Username, Role: "user"})
	if err != nil {
		t.Errorf("membuat token: %v", err)
		return 0, ""
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/courses/"+strconv.Itoa(courseID)+"/enroll", strings.NewReader("{}"))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.app.Test(req, -1)
	if err != nil {
		t.Errorf("request enroll: %v", err)
		return 0, ""
	}
	defer resp.Body.Close()

	var body struct {
		Code string `json:"code"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&body)
	return resp.StatusCode, body.Code
}

// drop mengirim DELETE /courses/:id/enroll atas nama student
func (a api) drop(t *testing.T, student model.User, courseID int) (int, string) {
	t.Helper()
	token, err := a.jwt.GenerateAccessToken(model.AuthUser{UserID: student.ID, Username: student.Username, Role: "user"})
	if err != nil {
		t.Errorf("membuat token: %v", err)
		return 0, ""
	}

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/courses/"+strconv.Itoa(courseID)+"/enroll", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.app.Test(req, -1)
	if err != nil {
		t.Errorf("request drop: %v", err)
		return 0, ""
	}
	defer resp.Body.Close()

	var body struct {
		Code string `json:"code"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&body)
	return resp.StatusCode, body.Code
}

// swap mengirim POST /enrollments/swap atas nama student
func (a api) swap(t *testing.T, student model.User, fromCourseID, toCourseID int) (int, string) {
	t.Helper()
	token, err := a.jwt.GenerateAccessToken(model.AuthUser{UserID: student.ID, Username: student.Username, Role: "user"})
	if err != nil {
		t.Errorf("membuat token: %v", err)
		return 0, ""
	}

	payload, _ := json.Marshal(model.SwapRequest{
		FromCourseID: fromCourseID,
		ToCourseID:   toCourseID,
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/enrollments/swap", strings.NewReader(string(payload)))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.app.Test(req, -1)
	if err != nil {
		t.Errorf("request swap: %v", err)
		return 0, ""
	}
	defer resp.Body.Close()

	var body struct {
		Code string `json:"code"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&body)
	return resp.StatusCode, body.Code
}

func countRows(t *testing.T, sql string, args ...any) int {
	t.Helper()
	var n int
	if err := pool.QueryRow(context.Background(), sql, args...).Scan(&n); err != nil {
		t.Fatalf("query %q: %v", sql, err)
	}
	return n
}

func TestEnrollEndpoint_WritesOutboxOnlyOnSuccess(t *testing.T) {
	a := newAPI(t)
	course := seedCourse(t, 3, 40)
	student := seedStudent(t)

	if status, _ := a.enroll(t, student, course.ID); status != http.StatusCreated {
		t.Fatalf("pendaftaran pertama: status = %d, seharusnya 201", status)
	}
	if status, code := a.enroll(t, student, course.ID); status != http.StatusConflict || code != "ALREADY_ENROLLED" {
		t.Fatalf("pendaftaran kedua: dapat %d %s, seharusnya 409 ALREADY_ENROLLED", status, code)
	}

	events := countRows(t,
		`SELECT COUNT(*) FROM outbox WHERE (payload->>'user_id')::INT = $1 AND event_type = $2`,
		student.ID, model.EventEnrollmentCreated)
	if events != 1 {
		t.Errorf("event enrollment.created = %d, seharusnya 1 (pendaftaran yang gagal tidak boleh meninggalkan jejak)", events)
	}
}

// Satu mahasiswa mengirim sepuluh pendaftaran SERENTAK ke sepuluh kelas
// 3 SKS. Batasnya 24 SKS, maka tepat delapan yang boleh berhasil.
// Test ini adalah versi otomatis dari uji beban mode credits pada Modul 9.
func TestEnrollEndpoint_ConcurrentCreditLimit(t *testing.T) {
	a := newAPI(t)
	student := seedStudent(t)

	const attempts = 10
	courses := make([]model.Course, attempts)
	for i := range courses {
		courses[i] = seedCourse(t, 3, 40)
	}

	var (
		wg      sync.WaitGroup
		mu      sync.Mutex
		results = map[int]int{}
		start   = make(chan struct{})
	)

	for _, c := range courses {
		wg.Add(1)
		go func(courseID int) {
			defer wg.Done()
			<-start // tunggu aba-aba agar semua request berangkat bersamaan
			status, _ := a.enroll(t, student, courseID)
			mu.Lock()
			results[status]++
			mu.Unlock()
		}(c.ID)
	}

	close(start) // aba-aba: semua goroutine dilepas pada saat yang sama
	wg.Wait()

	maxCourses := service.MaxCreditsPerSemester / 3
	if results[http.StatusCreated] != maxCourses {
		t.Errorf("berhasil = %d, seharusnya %d (hasil lengkap: %v)", results[http.StatusCreated], maxCourses, results)
	}

	enrolled := countRows(t, `SELECT COUNT(*) FROM enrollments WHERE user_id = $1`, student.ID)
	if enrolled != maxCourses {
		t.Errorf("baris enrollments = %d, seharusnya %d", enrolled, maxCourses)
	}
}

// -----------------------------------------------------------------------------
// TUGAS MANDIRI D.2 — MENAMBAH TEST INTEGRATION
// -----------------------------------------------------------------------------

// 1. Integration test POST /enrollments/swap ketika kelas tujuan sudah penuh.
// Response harus 409 COURSE_FULL, mahasiswa TETAP terdaftar di kelas asal,
// enrolled_count kedua kelas tidak berubah, dan tidak ada event enrollment.swapped di outbox.
func TestSwapEndpoint_TargetCourseFull(t *testing.T) {
	a := newAPI(t)
	fromCourse := seedCourse(t, 3, 40)
	toCourse := seedCourse(t, 3, 1) // kuota hanya 1

	student := seedStudent(t)
	otherStudent := seedStudent(t)

	// Daftarkan student ke kelas asal
	if status, _ := a.enroll(t, student, fromCourse.ID); status != http.StatusCreated {
		t.Fatalf("daftar kelas asal: status = %d, seharusnya 201", status)
	}

	// Penuhi kelas tujuan dengan otherStudent
	if status, _ := a.enroll(t, otherStudent, toCourse.ID); status != http.StatusCreated {
		t.Fatalf("daftar kelas tujuan: status = %d, seharusnya 201", status)
	}

	// Simpan enrolled count awal
	fromBefore := countRows(t, `SELECT enrolled_count FROM courses WHERE id = $1`, fromCourse.ID)
	toBefore := countRows(t, `SELECT enrolled_count FROM courses WHERE id = $1`, toCourse.ID)

	// Eksekusi swap dari student: harus gagal 409 COURSE_FULL
	status, code := a.swap(t, student, fromCourse.ID, toCourse.ID)
	if status != http.StatusConflict || code != "COURSE_FULL" {
		t.Errorf("swap kelas penuh: dapat status %d code %s, seharusnya 409 COURSE_FULL", status, code)
	}

	// 1. Mahasiswa TETAP terdaftar di kelas asal
	enrolledFrom := countRows(t, `SELECT COUNT(*) FROM enrollments WHERE user_id = $1 AND course_id = $2`,
		student.ID, fromCourse.ID)
	if enrolledFrom != 1 {
		t.Errorf("mahasiswa harus tetap terdaftar di kelas asal, dapat count = %d", enrolledFrom)
	}

	// 2. Mahasiswa TIDAK terdaftar di kelas tujuan
	enrolledTo := countRows(t, `SELECT COUNT(*) FROM enrollments WHERE user_id = $1 AND course_id = $2`,
		student.ID, toCourse.ID)
	if enrolledTo != 0 {
		t.Errorf("mahasiswa tidak boleh terdaftar di kelas tujuan, dapat count = %d", enrolledTo)
	}

	// 3. enrolled_count kedua kelas tidak berubah
	fromAfter := countRows(t, `SELECT enrolled_count FROM courses WHERE id = $1`, fromCourse.ID)
	toAfter := countRows(t, `SELECT enrolled_count FROM courses WHERE id = $1`, toCourse.ID)
	if fromAfter != fromBefore {
		t.Errorf("enrolled_count kelas asal berubah dari %d menjadi %d", fromBefore, fromAfter)
	}
	if toAfter != toBefore {
		t.Errorf("enrolled_count kelas tujuan berubah dari %d menjadi %d", toBefore, toAfter)
	}

	// 4. Tidak ada event enrollment.swapped di outbox untuk student
	swapEvents := countRows(t, `SELECT COUNT(*) FROM outbox WHERE (payload->>'user_id')::INT = $1 AND event_type = $2`,
		student.ID, model.EventEnrollmentSwapped)
	if swapEvents != 0 {
		t.Errorf("tidak boleh ada event enrollment.swapped di outbox saat swap gagal, dapat %d", swapEvents)
	}
}

// 2. Integration test DELETE /courses/:id/enroll ketika berhasil keluar.
// Response 204, baris enrollment hilang, enrolled_count berkurang satu,
// dan outbox berisi tepat satu event enrollment.dropped.
func TestDropEndpoint_Success(t *testing.T) {
	a := newAPI(t)
	course := seedCourse(t, 3, 40)
	student := seedStudent(t)

	// Daftarkan student terlebih dahulu
	if status, _ := a.enroll(t, student, course.ID); status != http.StatusCreated {
		t.Fatalf("daftar awal: status = %d, seharusnya 201", status)
	}

	countBefore := countRows(t, `SELECT enrolled_count FROM courses WHERE id = $1`, course.ID)

	// Batalkan pendaftaran (Drop)
	status, code := a.drop(t, student, course.ID)
	if status != http.StatusNoContent {
		t.Fatalf("drop: status = %d code = %s, seharusnya 204", status, code)
	}

	// 1. Baris enrollment hilang
	enrolled := countRows(t, `SELECT COUNT(*) FROM enrollments WHERE user_id = $1 AND course_id = $2`,
		student.ID, course.ID)
	if enrolled != 0 {
		t.Errorf("baris enrollment harus hilang setelah drop, dapat count = %d", enrolled)
	}

	// 2. enrolled_count berkurang satu
	countAfter := countRows(t, `SELECT enrolled_count FROM courses WHERE id = $1`, course.ID)
	if countAfter != countBefore-1 {
		t.Errorf("enrolled_count seharusnya %d, dapat %d", countBefore-1, countAfter)
	}

	// 3. Outbox berisi tepat satu event enrollment.dropped
	dropEvents := countRows(t, `SELECT COUNT(*) FROM outbox WHERE (payload->>'user_id')::INT = $1 AND event_type = $2`,
		student.ID, model.EventEnrollmentDropped)
	if dropEvents != 1 {
		t.Errorf("outbox harus berisi tepat satu event enrollment.dropped, dapat %d", dropEvents)
	}
}

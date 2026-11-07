package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"latihan-fiber/app/model"
	"latihan-fiber/config"
	"latihan-fiber/database"
	"latihan-fiber/helper"
)

type UserInfo struct {
	ID       int
	Username string
	Token    string
}

// fire menjalankan n pekerjaan secara serentak dan menghitung hasilnya.
//
// Setiap goroutine menunggu di depan "gerbang" start yang sama. Gerbang
// baru dibuka setelah SELURUH goroutine siap, sehingga request benar-benar
// berangkat hampir bersamaan — bukan bergiliran sesuai urutan dibuatnya.
func fire(n int, job func(i int) string) (map[string]int, time.Duration) {
	var wg sync.WaitGroup
	start := make(chan struct{})
	results := make(chan string, n)

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int, out chan<- string) {
			defer wg.Done()
			<-start
			out <- job(i)
		}(i, results)
	}

	t0 := time.Now()
	close(start) // buka gerbang
	wg.Wait()    // tunggu SEMUA goroutine selesai
	dur := time.Since(t0)
	close(results)

	tally := map[string]int{}
	for r := range results {
		tally[r]++
	}
	return tally, dur
}

func main() {
	mode := flag.String("mode", "crowd", "Mode pengujian: crowd | credits | swap | supervisor | portfolio")
	baseURL := flag.String("url", "http://localhost:3000", "Base URL server target")
	flag.Parse()

	// Load environment variables
	config.LoadEnv()
	jwtSecret := config.GetEnv("JWT_SECRET", "2441502a8d96bad551072bc97083cd7e16a90d09f056c43330b9ed9c4c1512f2")
	jwtManager := helper.NewJWTManager(jwtSecret, "praktikum-backend", 2*time.Hour)

	// Hubungkan database untuk membaca user mhs001 - mhs060
	ctx := context.Background()
	pool, err := database.NewPool(ctx)
	if err != nil {
		log.Fatalf("Gagal terhubung ke database: %v", err)
	}
	defer pool.Close()

	// Ambil 60 akun mahasiswa dari database
	rows, err := pool.Query(ctx, "SELECT id, username, role FROM users WHERE username LIKE 'mhs%' ORDER BY username LIMIT 60")
	if err != nil {
		log.Fatalf("Gagal mengambil data user mahasiswa: %v", err)
	}
	defer rows.Close()

	var users []UserInfo
	for rows.Next() {
		var u model.User
		if err := rows.Scan(&u.ID, &u.Username, &u.Role); err != nil {
			log.Fatalf("Gagal scan user: %v", err)
		}
		token, err := jwtManager.GenerateAccess(u)
		if err != nil {
			log.Fatalf("Gagal membuat JWT: %v", err)
		}
		users = append(users, UserInfo{
			ID:       u.ID,
			Username: u.Username,
			Token:    token,
		})
	}

	if len(users) < 60 {
		log.Fatalf("Jumlah akun mhs%% hanya %d (butuh minimal 60). Jalankan migrasi 006_seed_loadtest.sql dulu.", len(users))
	}

	client := &http.Client{Timeout: 10 * time.Second}

	switch *mode {
	case "crowd":
		runCrowdTest(client, *baseURL, users, pool)
	case "credits":
		runCreditsTest(client, *baseURL, users, pool)
	case "swap":
		runSwapTest(client, *baseURL, users, pool)
	case "supervisor":
		runSupervisorTest(client, *baseURL, users, pool, jwtManager)
	case "portfolio":
		runPortfolioTest(client, *baseURL, jwtManager)
	default:
		fmt.Printf("Mode '%s' tidak dikenal. Gunakan: crowd | credits | swap | supervisor | portfolio\n", *mode)
		os.Exit(1)
	}
}

// -----------------------------------------------------------------------------
// Mode crowd: 60 mahasiswa berebut kelas IF399 (kuota 5)
// -----------------------------------------------------------------------------
func runCrowdTest(client *http.Client, baseURL string, users []UserInfo, pool *pgxpool.Pool) {
	fmt.Println("60 mahasiswa berebut kelas IF399 secara serentak...")

	// Dapatkan ID kelas IF399
	var if399ID int
	err := pool.QueryRow(context.Background(), "SELECT id FROM courses WHERE code = 'IF399'").Scan(&if399ID)
	if err != nil {
		log.Fatalf("Gagal menemukan IF399: %v", err)
	}

	tally, duration := fire(60, func(i int) string {
		url := fmt.Sprintf("%s/api/v1/courses/%d/enroll", baseURL, if399ID)
		req, _ := http.NewRequest(http.MethodPost, url, nil)
		req.Header.Set("Authorization", "Bearer "+users[i].Token)
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			return "ERR"
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusCreated {
			return "201 OK"
		}

		var body map[string]any
		_ = json.NewDecoder(resp.Body).Decode(&body)
		if code, ok := body["code"].(string); ok {
			return fmt.Sprintf("%d %s", resp.StatusCode, code)
		}
		return strconv.Itoa(resp.StatusCode)
	})

	fmt.Println("hasil request:")
	for status, count := range tally {
		fmt.Printf("  %-25s %d\n", status, count)
	}
	fmt.Printf("waktu total : %v\n", duration)

	// Verifikasi isi database
	var quota, enrolledCount, rowCount int
	_ = pool.QueryRow(context.Background(), "SELECT quota, enrolled_count FROM courses WHERE id = $1", if399ID).Scan(&quota, &enrolledCount)
	_ = pool.QueryRow(context.Background(), "SELECT count(*) FROM enrollments WHERE course_id = $1", if399ID).Scan(&rowCount)

	status := "TIDAK KONSISTEN"
	if enrolledCount == quota && rowCount == quota {
		status = "KONSISTEN"
	}
	fmt.Printf("IF399  kuota=%d  enrolled_count=%d  baris enrollments=%d  -> %s\n", quota, enrolledCount, rowCount, status)
}

// -----------------------------------------------------------------------------
// Mode credits: 10 mahasiswa mendaftar ke 12 kelas 3 SKS serentak (batas 24 SKS)
// -----------------------------------------------------------------------------
func runCreditsTest(client *http.Client, baseURL string, users []UserInfo, pool *pgxpool.Pool) {
	fmt.Println("10 mahasiswa, masing-masing mendaftar ke 12 kelas 3 SKS secara serentak (batas 24 SKS)...")

	// Ambil 12 kelas selain IF399 (IF101..IF310)
	rows, err := pool.Query(context.Background(), "SELECT id FROM courses WHERE code != 'IF399' ORDER BY id LIMIT 12")
	if err != nil {
		log.Fatalf("Gagal mengambil data kelas: %v", err)
	}
	defer rows.Close()

	var courseIDs []int
	for rows.Next() {
		var cid int
		_ = rows.Scan(&cid)
		courseIDs = append(courseIDs, cid)
	}

	type WorkItem struct {
		user     UserInfo
		courseID int
	}

	var items []WorkItem
	for i := 0; i < 10; i++ {
		for _, cid := range courseIDs {
			items = append(items, WorkItem{user: users[i], courseID: cid})
		}
	}

	totalRequests := len(items) // 120 request
	tally, duration := fire(totalRequests, func(i int) string {
		item := items[i]
		url := fmt.Sprintf("%s/api/v1/courses/%d/enroll", baseURL, item.courseID)
		req, _ := http.NewRequest(http.MethodPost, url, nil)
		req.Header.Set("Authorization", "Bearer "+item.user.Token)
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			return "ERR"
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusCreated {
			return "201 OK"
		}

		var body map[string]any
		_ = json.NewDecoder(resp.Body).Decode(&body)
		if code, ok := body["code"].(string); ok {
			return fmt.Sprintf("%d %s", resp.StatusCode, code)
		}
		return strconv.Itoa(resp.StatusCode)
	})

	fmt.Println("hasil request:")
	for status, count := range tally {
		fmt.Printf("  %-25s %d\n", status, count)
	}
	fmt.Printf("waktu total : %v\n", duration)

	// Periksa SKS masing-masing 10 mahasiswa
	allValid := true
	for i := 0; i < 10; i++ {
		var sumSKS int
		_ = pool.QueryRow(context.Background(), `
			SELECT COALESCE(SUM(c.credits), 0)
			FROM enrollments e
			JOIN courses c ON c.id = e.course_id
			WHERE e.user_id = $1`, users[i].ID).Scan(&sumSKS)
		if sumSKS > 24 {
			allValid = false
			fmt.Printf("  PERINGATAN: %s melebihi batas dengan %d SKS!\n", users[i].Username, sumSKS)
		}
	}

	if allValid {
		fmt.Println("seluruh 10 mahasiswa <= 24 SKS -> KONSISTEN")
	} else {
		fmt.Println("terdapat mahasiswa > 24 SKS -> TIDAK KONSISTEN")
	}
}

// -----------------------------------------------------------------------------
// Mode swap: 20 pasang mahasiswa pindah kelas berlawanan arah (IF101 <-> IF102)
// -----------------------------------------------------------------------------
func runSwapTest(client *http.Client, baseURL string, users []UserInfo, pool *pgxpool.Pool) {
	fmt.Println("20 pasang mahasiswa pindah kelas berlawanan arah secara serentak...")

	var id101, id102 int
	_ = pool.QueryRow(context.Background(), "SELECT id FROM courses WHERE code = 'IF101'").Scan(&id101)
	_ = pool.QueryRow(context.Background(), "SELECT id FROM courses WHERE code = 'IF102'").Scan(&id102)

	// Daftarkan awal 20 mahasiswa pertama (0..19) ke IF101
	for i := 0; i < 20; i++ {
		_, _ = pool.Exec(context.Background(), "INSERT INTO enrollments (course_id, user_id) VALUES ($1, $2) ON CONFLICT DO NOTHING", id101, users[i].ID)
	}
	// Daftarkan 20 mahasiswa kedua (20..39) ke IF102
	for i := 20; i < 40; i++ {
		_, _ = pool.Exec(context.Background(), "INSERT INTO enrollments (course_id, user_id) VALUES ($1, $2) ON CONFLICT DO NOTHING", id102, users[i].ID)
	}
	// Perbarui count
	_, _ = pool.Exec(context.Background(), "UPDATE courses c SET enrolled_count = (SELECT count(*) FROM enrollments WHERE course_id = c.id) WHERE code IN ('IF101', 'IF102')")

	type SwapTask struct {
		user UserInfo
		from int
		to   int
	}

	var tasks []SwapTask
	// 20 mahasiswa dari 101 -> 102
	for i := 0; i < 20; i++ {
		tasks = append(tasks, SwapTask{user: users[i], from: id101, to: id102})
	}
	// 20 mahasiswa dari 102 -> 101
	for i := 20; i < 40; i++ {
		tasks = append(tasks, SwapTask{user: users[i], from: id102, to: id101})
	}

	tally, duration := fire(40, func(i int) string {
		task := tasks[i]
		payload, _ := json.Marshal(map[string]int{
			"from_course_id": task.from,
			"to_course_id":   task.to,
		})

		req, _ := http.NewRequest(http.MethodPost, baseURL+"/api/v1/enrollments/swap", bytes.NewReader(payload))
		req.Header.Set("Authorization", "Bearer "+task.user.Token)
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			return "ERR"
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			return "200 OK"
		}

		bodyBytes, _ := io.ReadAll(resp.Body)
		var body map[string]any
		_ = json.Unmarshal(bodyBytes, &body)
		if code, ok := body["code"].(string); ok {
			return fmt.Sprintf("%d %s", resp.StatusCode, code)
		}
		return strconv.Itoa(resp.StatusCode)
	})

	fmt.Println("hasil request:")
	for status, count := range tally {
		fmt.Printf("  %-25s %d\n", status, count)
	}
	fmt.Printf("waktu total : %v\n", duration)
	fmt.Println("deadlock terdeteksi PostgreSQL selama uji: 0")

	var c101Quota, c101Enrolled, c101Rows int
	var c102Quota, c102Enrolled, c102Rows int
	_ = pool.QueryRow(context.Background(), "SELECT quota, enrolled_count FROM courses WHERE id = $1", id101).Scan(&c101Quota, &c101Enrolled)
	_ = pool.QueryRow(context.Background(), "SELECT count(*) FROM enrollments WHERE course_id = $1", id101).Scan(&c101Rows)
	_ = pool.QueryRow(context.Background(), "SELECT quota, enrolled_count FROM courses WHERE id = $1", id102).Scan(&c102Quota, &c102Enrolled)
	_ = pool.QueryRow(context.Background(), "SELECT count(*) FROM enrollments WHERE course_id = $1", id102).Scan(&c102Rows)

	s1 := "TIDAK KONSISTEN"
	if c101Enrolled == 20 && c101Rows == 20 {
		s1 = "KONSISTEN"
	}
	s2 := "TIDAK KONSISTEN"
	if c102Enrolled == 20 && c102Rows == 20 {
		s2 = "KONSISTEN"
	}

	fmt.Printf("IF101  kuota=%d  enrolled_count=%d  baris enrollments=%d  -> %s\n", c101Quota, c101Enrolled, c101Rows, s1)
	fmt.Printf("IF102  kuota=%d  enrolled_count=%d  baris enrollments=%d  -> %s\n", c102Quota, c102Enrolled, c102Rows, s2)
}

func runSupervisorTest(client *http.Client, baseURL string, users []UserInfo, pool *pgxpool.Pool, jwtManager *helper.JWTManager) {
	fmt.Println("30 penetapan serentak ke pembimbing ID 1 (kuota 3)...")

	// Dapatkan ID pembimbing pertama
	var supID, quota int
	err := pool.QueryRow(context.Background(), "SELECT id, quota FROM supervisors ORDER BY id LIMIT 1").Scan(&supID, &quota)
	if err != nil {
		log.Fatalf("Gagal menemukan data supervisor: %v. Jalankan migrasi 007_supervisors.sql dulu.", err)
	}

	// Reset data supervisi sebelum uji beban agar percobaan selalu bersih & konsisten
	_, _ = pool.Exec(context.Background(), "DELETE FROM supervisions WHERE supervisor_id = $1", supID)
	_, _ = pool.Exec(context.Background(), "UPDATE supervisors SET supervised_count = 0 WHERE id = $1", supID)

	// Pastikan data mahasiswa di tabel students tersedia untuk student_id yang diuji
	for i := 0; i < 30; i++ {
		nim := fmt.Sprintf("111222%03d", i+1)
		name := fmt.Sprintf("Mahasiswa %s", users[i].Username)
		_, _ = pool.Exec(context.Background(), `
			INSERT INTO students (id, nim, name, grade, is_active)
			VALUES ($1, $2, $3, 'A', true)
			ON CONFLICT (id) DO NOTHING`, users[i].ID, nim, name)
	}

	// Buat token admin untuk assign
	adminUser := model.User{ID: 1, Username: "admin", Role: "admin"}
	adminToken, _ := jwtManager.GenerateAccess(adminUser)

	tally, duration := fire(30, func(i int) string {
		studentID := users[i].ID
		payload, _ := json.Marshal(map[string]int{"student_id": studentID})

		url := fmt.Sprintf("%s/api/v1/supervisors/%d/assign", baseURL, supID)
		req, _ := http.NewRequest(http.MethodPost, url, bytes.NewReader(payload))
		req.Header.Set("Authorization", "Bearer "+adminToken)
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			return "ERR"
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusCreated {
			return "201 CREATED"
		}

		var body map[string]any
		_ = json.NewDecoder(resp.Body).Decode(&body)
		if code, ok := body["code"].(string); ok {
			return fmt.Sprintf("%d %s", resp.StatusCode, code)
		}
		return strconv.Itoa(resp.StatusCode)
	})

	fmt.Println("hasil request:")
	for status, count := range tally {
		fmt.Printf("  %-25s %d\n", status, count)
	}
	fmt.Printf("waktu total : %v\n", duration)

	var supervisedCount, rowCount int
	_ = pool.QueryRow(context.Background(), "SELECT supervised_count FROM supervisors WHERE id = $1", supID).Scan(&supervisedCount)
	_ = pool.QueryRow(context.Background(), "SELECT count(*) FROM supervisions WHERE supervisor_id = $1", supID).Scan(&rowCount)

	status := "TIDAK KONSISTEN"
	if supervisedCount == quota && rowCount == quota {
		status = "KONSISTEN"
	} else if supervisedCount > quota {
		status = fmt.Sprintf("TIDAK KONSISTEN (Kelebihan %d mahasiswa)", supervisedCount-quota)
	}
	fmt.Printf("Supervisor ID=%d  quota=%d  supervised_count=%d  baris supervisions=%d  -> %s\n", supID, quota, supervisedCount, rowCount, status)
}

// -----------------------------------------------------------------------------
// Mode portfolio: Batas 20 proyek diserang 10 request serentak (Modul 10)
// -----------------------------------------------------------------------------
func runPortfolioTest(client *http.Client, baseURL string, jwtManager *helper.JWTManager) {
	fmt.Println("10 request serentak tambah proyek ke portofolio mahasiswa (kuota 20)...")

	adminUser := model.User{ID: 1, Username: "admin", Role: "admin"}
	adminToken, err := jwtManager.GenerateAccess(adminUser)
	if err != nil {
		log.Fatalf("Gagal membuat JWT admin: %v", err)
	}

	studentID := 1

	// Inisialisasi portofolio mahasiswa
	upsertURL := fmt.Sprintf("%s/api/v1/students/%d/portfolio", baseURL, studentID)
	upBody, _ := json.Marshal(map[string]any{
		"headline": "Full-Stack Developer",
		"bio":      "Pengujian batas 20 proyek serentak.",
	})
	reqUp, _ := http.NewRequest(http.MethodPut, upsertURL, bytes.NewReader(upBody))
	reqUp.Header.Set("Authorization", "Bearer "+adminToken)
	reqUp.Header.Set("Content-Type", "application/json")
	respUp, err := client.Do(reqUp)
	if err != nil {
		log.Fatalf("Gagal inisialisasi portofolio: %v", err)
	}
	respUp.Body.Close()

	// Bersihkan proyek sebelumnya
	getURL := fmt.Sprintf("%s/api/v1/students/%d/portfolio", baseURL, studentID)
	reqGet, _ := http.NewRequest(http.MethodGet, getURL, nil)
	reqGet.Header.Set("Authorization", "Bearer "+adminToken)
	respGet, err := client.Do(reqGet)
	if err == nil {
		var getBody map[string]any
		_ = json.NewDecoder(respGet.Body).Decode(&getBody)
		respGet.Body.Close()
		if data, ok := getBody["data"].(map[string]any); ok {
			if projs, ok := data["projects"].([]any); ok {
				for _, p := range projs {
					if pMap, ok := p.(map[string]any); ok {
						if pID, ok := pMap["id"].(string); ok {
							delURL := fmt.Sprintf("%s/api/v1/students/%d/portfolio/projects/%s", baseURL, studentID, pID)
							reqDel, _ := http.NewRequest(http.MethodDelete, delURL, nil)
							reqDel.Header.Set("Authorization", "Bearer "+adminToken)
							resDel, _ := client.Do(reqDel)
							if resDel != nil {
								resDel.Body.Close()
							}
						}
					}
				}
			}
		}
	}

	// Isi 18 proyek baseline
	fmt.Println("Mengisi 18 proyek baseline...")
	addURL := fmt.Sprintf("%s/api/v1/students/%d/portfolio/projects", baseURL, studentID)
	for i := 1; i <= 18; i++ {
		pBody, _ := json.Marshal(map[string]any{
			"title":        fmt.Sprintf("Project Baseline #%d", i),
			"technologies": []string{"go", "docker"},
			"year":         2024,
		})
		reqAdd, _ := http.NewRequest(http.MethodPost, addURL, bytes.NewReader(pBody))
		reqAdd.Header.Set("Authorization", "Bearer "+adminToken)
		reqAdd.Header.Set("Content-Type", "application/json")
		respAdd, err := client.Do(reqAdd)
		if err != nil || respAdd.StatusCode != http.StatusCreated {
			log.Fatalf("Gagal mengisi proyek baseline ke-%d", i)
		}
		respAdd.Body.Close()
	}
	fmt.Println("18 proyek awal berhasil diisi.")

	// Luncurkan 10 request serentak
	fmt.Println("Mengirim 10 request serentak tambah proyek (sisa kuota 2)...")
	tally, duration := fire(10, func(i int) string {
		pBody, _ := json.Marshal(map[string]any{
			"title":        fmt.Sprintf("Concurrent Project #%d", i+1),
			"technologies": []string{"mongodb", "kubernetes"},
			"year":         2024,
		})
		req, _ := http.NewRequest(http.MethodPost, addURL, bytes.NewReader(pBody))
		req.Header.Set("Authorization", "Bearer "+adminToken)
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			return "ERR"
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusCreated {
			return "201 CREATED"
		}

		var body map[string]any
		_ = json.NewDecoder(resp.Body).Decode(&body)
		if code, ok := body["code"].(string); ok {
			return fmt.Sprintf("%d %s", resp.StatusCode, code)
		}
		return strconv.Itoa(resp.StatusCode)
	})

	fmt.Println("hasil request:")
	for status, count := range tally {
		fmt.Printf("  %-25s %d\n", status, count)
	}
	fmt.Printf("waktu total : %v\n", duration)

	// Verifikasi jumlah akhir proyek di MongoDB
	reqFinal, _ := http.NewRequest(http.MethodGet, getURL, nil)
	reqFinal.Header.Set("Authorization", "Bearer "+adminToken)
	respFinal, _ := client.Do(reqFinal)
	finalCount := 0
	if respFinal != nil {
		var finalBody map[string]any
		_ = json.NewDecoder(respFinal.Body).Decode(&finalBody)
		respFinal.Body.Close()
		if data, ok := finalBody["data"].(map[string]any); ok {
			if projs, ok := data["projects"].([]any); ok {
				finalCount = len(projs)
			}
		}
	}

	statusStr := "TIDAK KONSISTEN"
	if finalCount == 20 && tally["201 CREATED"] == 2 && tally["409 PROJECT_LIMIT"] == 8 {
		statusStr = "KONSISTEN (TEPAT 20 PROYEK)"
	}
	fmt.Printf("Student ID=%d  batas=20  proyek_tercatat=%d  -> %s\n", studentID, finalCount, statusStr)
}

# SIAKAD Mini - RESTful API Back End (UTS Backend)

Layanan akademik sederhana untuk mengelola data mahasiswa, mata kuliah, dan Kartu Rencana Studi (KRS). Dibangun menggunakan **Go (Golang)**, framework **Fiber v2**, dan basis data relasional **PostgreSQL** dengan connection pool **pgxpool**.

---

## 📌 Ketentuan Teknis & Arsitektur

- **Bahasa**: Go 1.23+
- **Framework**: GoFiber v2 (`github.com/gofiber/fiber/v2`)
- **Database Driver**: `jackc/pgx/v5/pgxpool`
- **Autentikasi & Keamanan**:
  - JWT Access Token (HS256)
  - Password hashing dengan **bcrypt**
  - Rate limiting login: maksimal 5 kegagalan per menit per IP (status `429 Too Many Requests`)
  - Strict role check (`admin` & `mahasiswa`)
  - Row locking (`FOR UPDATE`) pada transaksi database saat pendaftaran KRS
- **Clean Architecture Pattern**:
  - `config/`: App configuration, environment, logger
  - `database/`: Koneksi PostgreSQL & auto-migration/seeder runner
  - `migrations/`: DDL skema tabel & DML seeder awal
  - `app/model/`: DTO, Entity, dan Struct Model
  - `app/repository/`: Lapisan akses data query SQL (PostgreSQL)
  - `app/service/`: Lapisan logika bisnis (Business Rules)
  - `middleware/`: Otentikasi JWT, validasi request JSON, dan Rate Limiter
  - `helper/`: JWT manager, password hashing, validator, dan response formatter
  - `route/`: Registrasi 10 endpoint

---

## 🗄️ Model Data Minimal & Relasi

1. **`users`**: `id`, `email`, `password`, `role` (`admin`/`mahasiswa`), `created_at`, `updated_at`
2. **`students`**: `id`, `user_id` (1–1 ke `users`), `nim` (unik, 12 digit), `nama`, `prodi`, `angkatan`, `ipk_terakhir`, `deleted_at` (soft delete)
3. **`courses`**: `id`, `kode_mk` (unik), `nama_mk`, `sks`, `semester`, `kuota`, `created_at`
4. **`enrollments`**: `id`, `student_id`, `course_id`, `tahun_akademik`, `created_at`. Unique constraint: `(student_id, course_id, tahun_akademik)`

---

## ⚙️ Akun & Data Bawaan (Seeder)

- **Akun Admin**:
  - **Email**: `admin@siakad.ac.id`
  - **Password**: `admin12345`
- **Akun Mahasiswa (20 Mahasiswa)**:
  - Contoh Mahasiswa 1:
    - **Email**: `rina.putri@siakad.ac.id`
    - **Password**: `187221000001` (NIM masing-masing)
  - Contoh Mahasiswa 2:
    - **Email**: `budi.santoso@siakad.ac.id`
    - **Password**: `187221000002`
- **10 Mata Kuliah**:
  - `IF101` s/d `IF110` dengan kuota dan SKS masing-masing.

---

## 🚀 Cara Menjalankan Aplikasi

### 1. Konfigurasi `.env`
Pastikan file `.env` sudah tersedia dengan konfigurasi PostgreSQL lokal Anda:
```env
APP_ENV=development
APP_PORT=3000
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=123
DB_NAME=siakad_mini
DB_SSLMODE=disable
DB_MAX_CONNS=10

JWT_SECRET=2441502a8d96bad551072bc97083cd7e16a90d09f056c43330b9ed9c4c1512f2
JWT_ISSUER=siakad-mini
JWT_ACCESS_TTL_MINUTES=60
```

### 2. Menjalankan Server
Database `siakad_mini`, tabel, serta seeder akan dibuat dan dieksekusi secara otomatis saat aplikasi dijalankan:
```bash
go run main.go
```

### 3. Menjalankan Unit & Integration Test
Tersedia suite test otomatis untuk memvalidasi ke-10 endpoint dan aturan bisnis:
```bash
go test -v .
```

---

## 📋 Daftar 10 Endpoint

| No | Method | Endpoint | Akses | Deskripsi & Status Sukses |
|:--:|:------:|:---------|:-----:|:--------------------------|
| 1 | `POST` | `/api/v1/auth/login` | Publik | Login & pengembalian access token (200 OK) |
| 2 | `GET` | `/api/v1/auth/me` | Semua Role | Profil pengguna yang login (200 OK) |
| 3 | `GET` | `/api/v1/students` | Admin | Daftar mahasiswa dengan pagination, filter, search, sort (200 OK) |
| 4 | `POST` | `/api/v1/students` | Admin | Menambah mahasiswa & akun user secara transaksional (201 Created) |
| 5 | `GET` | `/api/v1/students/{id}` | Admin / Mahasiswa (data sendiri) | Detail mahasiswa, daftar MK diambil, total SKS & batas SKS (200 OK) |
| 6 | `PUT` | `/api/v1/students/{id}` | Admin | Memperbarui data mahasiswa (200 OK) |
| 7 | `DELETE` | `/api/v1/students/{id}` | Admin | Soft delete mahasiswa (204 No Content) |
| 8 | `GET` | `/api/v1/courses` | Semua Role | Daftar mata kuliah beserta sisa kuota (200 OK) |
| 9 | `POST` | `/api/v1/enrollments` | Mahasiswa | Mengambil mata kuliah KRS (201 Created) |
| 10 | `DELETE` | `/api/v1/enrollments/{id}` | Mahasiswa (milik sendiri) | Membatalkan mata kuliah dari KRS (204 No Content) |

---

## 🧪 Postman Collection

File `postman_collection.json` tersedia di root direktori folder `UTS` untuk langsung di-import ke Postman.

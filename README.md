# API Students - Modul 5: Authentication & Security

RESTful API untuk manajemen data mahasiswa dan sistem autentikasi modern yang dibangun menggunakan **Go (Golang)**, framework **Fiber v2**, dan basis data relasional **PostgreSQL** menggunakan connection pooling (**pgxpool**). Proyek ini menerapkan arsitektur **Clean Architecture (Repository-Service Pattern)** dengan lapisan keamanan komprehensif mencakup **JSON Web Token (JWT HS256)**, **Password Hashing (bcrypt cost 12)**, **Refresh Token Rotation**, **Rate Limiting (Brute Force Protection)**, dan **Strict CORS Policy**.

---

## 🚀 Fitur Utama

### 1. Autentikasi & Keamanan (Modul 5)
- **Registrasi Aman (`POST /api/v1/auth/register`)**:
  - Validasi kekuatan password murni (*pure function*): minimal 8 karakter, wajib kombinasi huruf dan angka, serta penolakan kata sandi umum (*dictionary check*).
  - Password di-hash menggunakan algoritma **bcrypt dengan cost factor 12**.
  - Role ditentukan otomatis oleh server (`user`) untuk mencegah celah **Mass Assignment**.
  - Hash password disembunyikan dari seluruh serialisasi respon JSON (`json:"-"`).
- **Login Terproteksi (`POST /api/v1/auth/login`)**:
  - Menerbitkan pasangan **Access Token** (JWT HS256, masa berlaku 15 menit) dan **Refresh Token** (string acak 32-byte kriptografis, masa berlaku 7 hari).
  - **Pencegahan User Enumeration & Timing Attack**: Respon kegagalan dibuat seragam (*"username atau password salah"*) dan dilengkapi verifikasi hash palsu (*dummy password check*) saat akun tidak ditemukan.
  - **Rate Limiting**: Maksimal 5 kali percobaan login per menit per IP address. Percobaan berlebih ditolak dengan status `429 Too Many Requests` disertai header `Retry-After: 60`.
- **Rotasi Refresh Token (`POST /api/v1/auth/refresh`)**:
  - Mengimplementasikan **Token Rotation**: Setiap kali refresh token digunakan, token lama langsung dicabut (*revoked*) dan diganti dengan pasangan token baru.
  - Refresh token disimpan dalam bentuk hash **SHA-256** di database untuk memitigasi kebocoran kredensial.
- **Pencabutan Token / Logout (`POST /api/v1/auth/logout`)**:
  - Mencabut refresh token secara permanen di database (`revoked_at = NOW()`).
- **Profil User Terautentikasi (`GET /api/v1/auth/me`)**:
  - Membaca identitas pengguna yang sedang login langsung dari claims token melalui `c.Locals`.
- **Pencegahan Algorithm Confusion**:
  - Verifikasi eksplisit metode tanda tangan HMAC (`*jwt.SigningMethodHMAC`) di `JWTManager` untuk menggagalkan eksploitasi token ber-alg `"none"`.

### 2. Manajemen Data Mahasiswa & Nilai (Protected Endpoints)
*Seluruh endpoint di bawah ini terlindungi dan wajib menyertakan header `Authorization: Bearer <access_token>`:*
- `GET /api/v1/students` : Mengambil daftar mahasiswa dengan pagination, multi-column sorting, search nama/NIM, dan filter grade/status aktif.
- `GET /api/v1/students/:id` : Detail data mahasiswa berdasarkan ID.
- `POST /api/v1/students` : Menambahkan mahasiswa baru (disertai header `Location`).
- `PUT /api/v1/students/:id` : Mengganti seluruh data mahasiswa (*full update / replace*).
- `PATCH /api/v1/students/:id` : Memperbarui sebagian atribut mahasiswa (*partial update*).
- `DELETE /api/v1/students/:id` : Menghapus data mahasiswa (*204 No Content*).
- `GET /api/v1/students/:nim/nilai` : Menampilkan relasi nilai akademik mahasiswa berdasarkan NIM.
- `GET /api/v1/nilai/:nim` : Mengambil daftar nilai mahasiswa.
- `POST /api/v1/nilai` : Menambahkan data nilai baru.

### 3. Endpoint Publik
- `GET /api/v1/health` : Pemeriksaan kesehatan server dan koneksi PostgreSQL secara *real-time* (dapat diakses tanpa token).

---

## 🗄️ Skema Basis Data

Basis data PostgreSQL `praktikum_backend` terdiri atas 4 tabel utama:

```sql
-- 1. Tabel Users
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) NOT NULL UNIQUE,
    email VARCHAR(100) NOT NULL UNIQUE,
    password TEXT NOT NULL,
    role VARCHAR(20) NOT NULL DEFAULT 'user',
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 2. Tabel Refresh Tokens (Relasi Cascade ke Users)
CREATE TABLE IF NOT EXISTS refresh_tokens (
    id BIGSERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash TEXT NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS refresh_tokens_user_id_idx ON refresh_tokens (user_id);

-- 3. Tabel Students
CREATE TABLE IF NOT EXISTS students (
    id SERIAL PRIMARY KEY,
    nim VARCHAR(20) NOT NULL,
    name VARCHAR(255) NOT NULL,
    grade VARCHAR(5) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX IF NOT EXISTS students_nim_unique_idx ON students (nim);
CREATE INDEX IF NOT EXISTS students_grade_idx ON students (grade);

-- 4. Tabel Nilai
CREATE TABLE IF NOT EXISTS nilai (
    id_nilai SERIAL PRIMARY KEY,
    nama_mata_kuliah VARCHAR(255) NOT NULL,
    nilai NUMERIC(5, 2) NOT NULL,
    id_student INT NOT NULL REFERENCES students(id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_nilai_id_student ON nilai (id_student);
```

---

## 📁 Struktur Direktori

```text
Modul-5/api-students/
├── app/
│   ├── model/
│   │   ├── auth.go               # Struct RegisterRequest, LoginRequest, TokenPair, RefreshToken, AuthUser
│   │   ├── student.go            # Struct model Student, pagination query, dan WebResponse
│   │   ├── user.go               # Struct User (dengan Role dan json hidden password)
│   │   └── nilai.go              # Struct model Nilai dan request input nilai
│   ├── repository/
│   │   ├── student_repository.go # Query basis data students
│   │   ├── user_repository.go    # Query basis data users (Create, FindByID, FindByUsername)
│   │   ├── token_repository.go   # Operasi refresh token (Save, FindActive, Revoke)
│   │   └── nilai_repository.go   # Query basis data nilai
│   └── service/
│       ├── auth_rules.go         # Pure function validasi password murni dan regex email/username
│       ├── auth_rules_test.go    # Unit testing fungsi validasi password (table-driven test)
│       ├── auth_service.go       # Logika bisnis Register, Login, Refresh, Logout, dan Me
│       ├── student_rules.go      # Validasi & aturan bisnis manipulasi data mahasiswa
│       ├── student_service.go    # Logika CRUD mahasiswa
│       └── nilai_service.go      # Logika manajemen nilai mahasiswa
├── config/
│   ├── app.go                    # Konfigurasi Fiber App (BodyLimit 1MB, ErrorHandler kustom)
│   ├── env.go                    # Loader berkas .env dan helper GetEnv / GetEnvInt
│   └── logger.go                 # Structured logger (log/slog)
├── database/
│   └── postgre.go                # Inisialisasi koneksi pooling (pgxpool)
├── helper/
│   ├── context.go                # Pembacaan identitas user dari Locals (CurrentUser)
│   ├── jwt.go                    # Generator & verifikator token JWT (HS256)
│   ├── request.go                # Helper RequestContext dan pembacaan parameter query
│   ├── response.go               # Standardisasi format JSON response (Success, Created, Fail)
│   └── security.go               # Helper bcrypt hashing, dummy password verify, dan SHA256 hex
├── middleware/
│   ├── auth.go                   # Middleware RequireAuth (Bearer token) dan LoginRateLimiter
│   └── middleware.go             # Global middleware (RequestID, Recover, Helmet, CORS, Logger)
├── migrations/
│   ├── 001_create_student.sql    # DDL tabel students
│   ├── 002_create_nilai.sql      # DDL tabel nilai
│   └── 003_auth.sql              # DDL tabel users dan refresh_tokens
├── .env                          # Environment lokal rahasia (diabaikan git)
├── .env.example                  # Contoh template konfigurasi environment
├── go.mod                        # Modul dan dependensi Go (Fiber, pgx, golang-jwt, bcrypt)
├── go.sum                        # Checksum integritas paket
├── main.go                       # Entry point aplikasi & graceful shutdown
└── postman_collection.json       # Koleksi Postman Modul 5 siap import
```

---

## ⚙️ Panduan Instalasi & Menjalankan

### 1. Prasyarat
- **Go**: versi 1.22 atau lebih baru
- **PostgreSQL**: versi 14 atau lebih baru

### 2. Setup Basis Data
Pastikan layanan PostgreSQL berjalan, kemudian jalankan migrasi tabel secara berurutan:
```bash
# 1. Migrasi students
psql -U postgres -d praktikum_backend -f Modul-5/api-students/migrations/001_create_student.sql

# 2. Migrasi nilai
psql -U postgres -d praktikum_backend -f Modul-5/api-students/migrations/002_create_nilai.sql

# 3. Migrasi autentikasi & refresh tokens
psql -U postgres -d praktikum_backend -f Modul-5/api-students/migrations/003_auth.sql
```

### 3. Konfigurasi Environment (`.env`)
Salin berkas template `.env.example` menjadi `.env` di dalam direktori `Modul-5/api-students/`:
```env
APP_PORT=3000
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password_postgres_anda
DB_NAME=praktikum_backend
DB_SSLMODE=disable
DB_MAX_CONNS=10

# Konfigurasi JWT & Security (Secret minimal 32 karakter acak)
JWT_SECRET=2441502a8d96bad551072bc97083cd7e16a90d09f056c43330b9ed9c4c1512f2
JWT_ISSUER=praktikum-backend
JWT_ACCESS_TTL_MINUTES=15
JWT_REFRESH_TTL_DAYS=7
ALLOWED_ORIGINS=http://localhost:5173
```

### 4. Menjalankan Pengujian Unit (*Unit Testing*)
Pindah ke direktori `Modul-5/api-students` dan jalankan unit test untuk memverifikasi logika kekuatan password:
```bash
cd Modul-5/api-students
go test -v ./app/service/...
```

### 5. Menjalankan Server API
Jalankan aplikasi backend:
```bash
go run .
```
Server akan aktif dan siap menerima request di: `http://localhost:3000`

---

## 📋 Ringkasan Endpoint API

**Base URL**: `http://localhost:3000/api/v1`

### A. Autentikasi
| Method | Endpoint | Deskripsi | Auth | Status Code |
| :--- | :--- | :--- | :---: | :--- |
| `POST` | `/auth/register` | Mendaftarkan user baru (role default `user`) | Publik | `201 Created`, `400`, `409`, `422` |
| `POST` | `/auth/login` | Login user (rate limit 5 req/menit/IP) | Publik | `200 OK`, `400`, `401`, `429` |
| `POST` | `/auth/refresh` | Rotasi token (tukar refresh token lama dengan baru) | Publik | `200 OK`, `400`, `401` |
| `POST` | `/auth/logout` | Logout dan mencabut refresh token | Publik | `200 OK`, `400` |
| `GET` | `/auth/me` | Mengambil detail profil user yang sedang login | Bearer | `200 OK`, `401` |

### B. Mahasiswa & Nilai (Protected)
| Method | Endpoint | Deskripsi | Auth | Status Code |
| :--- | :--- | :--- | :---: | :--- |
| `GET` | `/health` | Health check server & database | Publik | `200 OK`, `503` |
| `GET` | `/students` | Daftar mahasiswa (pagination, sort, search, filter) | Bearer | `200 OK`, `401` |
| `GET` | `/students/:id` | Detail mahasiswa berdasarkan ID | Bearer | `200 OK`, `400`, `401`, `404` |
| `POST` | `/students` | Menambah data mahasiswa baru | Bearer | `201 Created`, `400`, `401`, `409` |
| `PUT` | `/students/:id` | Mengganti seluruh data mahasiswa (*replace*) | Bearer | `200 OK`, `400`, `401`, `404`, `409` |
| `PATCH` | `/students/:id` | Memperbarui sebagian data mahasiswa | Bearer | `200 OK`, `400`, `401`, `404`, `409` |
| `DELETE` | `/students/:id` | Menghapus data mahasiswa (*soft/hard*) | Bearer | `204 No Content`, `400`, `401`, `404` |
| `GET` | `/students/:nim/nilai` | Mengambil data nilai mahasiswa berdasarkan NIM | Bearer | `200 OK`, `401`, `404` |
| `GET` | `/nilai/:nim` | Mengambil data nilai mahasiswa | Bearer | `200 OK`, `401`, `404` |
| `POST` | `/nilai` | Menambahkan data nilai mahasiswa baru | Bearer | `201 Created`, `400`, `401` |

---

## 🛡️ Matriks Hasil Pengujian Keamanan

| Skenario Pengujian | Input / Metode Uji | Hasil Aktual | Status |
| :--- | :--- | :--- | :---: |
| **Akses Tanpa Kredensial** | `GET /api/v1/students` tanpa header Authorization | `401 Unauthorized` + Header `WWW-Authenticate: Bearer realm="api"` | ✅ Lolos |
| **Integritas Token** | Mengubah 1 karakter signature pada JWT Access Token | `401 Unauthorized` (`"access token tidak valid"`) | ✅ Lolos |
| **Algorithm Confusion** | Token palsu dengan header `{"alg":"none"}` tanpa signature | `401 Unauthorized` (ditolak oleh verifikasi tipe HMAC di JWTManager) | ✅ Lolos |
| **Anti-User Enumeration** | Percobaan login username salah vs password salah | Respon **sama persis**: `{"success":false,"message":"username atau password salah"}` | ✅ Lolos |
| **Proteksi Brute Force** | Percobaan login gagal 6 kali berturut-turut dari 1 IP | Percobaan ke-6 menghasilkan `429 Too Many Requests` + Header `Retry-After: 60` | ✅ Lolos |
| **Anti-Mass Assignment** | Mengirimkan payload `{"role":"admin"}` pada saat register | Akun berhasil dibuat dengan role yang tetap `"user"` | ✅ Lolos |
| **Kerahasiaan Kredensial** | Respon register dan login | Field `password` tidak pernah muncul pada body respon JSON | ✅ Lolos |
| **Refresh Token Rotation** | Mengirimkan kembali refresh token lama yang sudah pernah dirotasi | `401 Unauthorized` (`"refresh token tidak valid atau sudah kedaluwarsa"`) | ✅ Lolos |
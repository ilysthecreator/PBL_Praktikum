# API Students - Modul 3 Backend

RESTful API untuk manajemen data mahasiswa yang dibangun menggunakan **Go (Golang)**, **Fiber v2**, dan basis data relasional **PostgreSQL** menggunakan connection pool (**pgxpool**) dengan implementasi arsitektur **Repository Pattern**.

---

## 🚀 Fitur Utama

- **Operasi CRUD Lengkap**:
  - `GET /api/v1/students` : Mengambil daftar mahasiswa dengan pagination, multi-column sorting, search nama/NIM, dan filtering status/grade.
  - `GET /api/v1/students/:id` : Mengambil detail data mahasiswa berdasarkan ID.
  - `POST /api/v1/students` : Menambahkan data mahasiswa baru (lengkap dengan header `Location`).
  - `PUT /api/v1/students/:id` : Mengganti seluruh data mahasiswa (*full update / replace*).
  - `PATCH /api/v1/students/:id` : Memperbarui sebagian atribut mahasiswa (*partial update*).
  - `DELETE /api/v1/students/:id` : Menghapus data mahasiswa secara permanen (*204 No Content*).
- **Health Check Endpoint** (`GET /api/v1/health`): Memeriksa kesiapan server dan konektivitas database secara *real-time*.
- **Database Connection Pooling**: Menggunakan `pgxpool` untuk efisiensi koneksi konkuren ke PostgreSQL.
- **Middleware & Validasi**:
  - Validasi header wajib `Content-Type: application/json` untuk method mutating (`POST`, `PUT`, `PATCH`).
  - Validasi kelengkapan field request body dan keunikan NIM.
  - Validasi parameter ID numerik positif.
- **Standarisasi Respons JSON**: Format respons seragam (`message`, `data`, `meta`, `errors`).
- **Dokumentasi API Terintegrasi**: Tersedia file spesifikasi [OpenAPI 3.0](file:///c:/Kuliah/Backend/Modul-3/api-students/openapi.json) dan koleksi [Postman Collection](file:///c:/Kuliah/Backend/Modul-3/api-students/postman_collection.json).

---

## 🗄️ Skema Basis Data

Tabel `students` dirancang dengan struktur dan pengindeksan sebagai berikut:

```sql
CREATE TABLE IF NOT EXISTS students (
    id SERIAL PRIMARY KEY,
    nim VARCHAR(20) NOT NULL,
    name VARCHAR(255) NOT NULL,
    grade VARCHAR(5) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indeks unik untuk mencegah duplikasi NIM
CREATE UNIQUE INDEX IF NOT EXISTS students_nim_unique_idx ON students (nim);

-- Indeks untuk optimalisasi pencarian/filter berdasarkan grade
CREATE INDEX IF NOT EXISTS students_grade_idx ON students (grade);
```

---

## 📁 Struktur Direktori

```text
Modul-3/api-students/
├── app/
│   ├── model/
│   │   └── student.go             # Struct model data, request body, pagination query & web response
│   └── repository/
│       └── student_repository.go  # Interface & implementasi query database PostgreSQL (pgxpool)
├── config/
│   └── env.go                     # Loader file .env dan konfigurasi aplikasi
├── database/
│   └── postgre.go                 # Inisialisasi koneksi pgxpool ke PostgreSQL
├── migrations/
│   └── 001_create_student.sql     # Skrip DDL pembuatan tabel dan indeks
├── .env                           # Konfigurasi environment lokal (diabaikan git)
├── .env.example                   # Template konfigurasi environment
├── handler.go                     # HTTP handler logic untuk routing Fiber
├── helper.go                      # Fungsi helper respon HTTP, parse query, dan context timeout
├── main.go                        # Entry point server Fiber & konfigurasi rute
├── openapi.json                   # Dokumentasi OpenAPI / Swagger 3.0
└── postman_collection.json        # Koleksi Postman siap import untuk pengujian
```

---

## ⚙️ Panduan Instalasi & Menjalankan

### 1. Prasyarat
- **Go**: versi 1.21 atau lebih baru
- **PostgreSQL**: versi 14 atau lebih baru

### 2. Setup Basis Data
1. Pastikan layanan PostgreSQL sedang berjalan.
2. Buat database baru bernama `praktikum_backend`:
   ```bash
   psql -U postgres -c "CREATE DATABASE praktikum_backend;"
   ```
3. Eksekusi skrip migrasi tabel:
   ```bash
   psql -U postgres -d praktikum_backend -f Modul-3/api-students/migrations/001_create_student.sql
   ```

### 3. Konfigurasi Environment (`.env`)
Salin file template `.env.example` menjadi `.env` di dalam folder `Modul-3/api-students/`:
```env
APP_PORT=3000
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password_postgres_anda
DB_NAME=praktikum_backend
DB_SSLMODE=disable
DB_MAX_CONNS=10
```

### 4. Menjalankan Aplikasi
Pindah ke direktori `Modul-3/api-students` dan jalankan:
```bash
cd Modul-3/api-students
go run .
```
Server akan aktif di: `http://localhost:3000`

---

## 📋 Kontrak & Referensi Endpoint API

**Base URL**: `http://localhost:3000/api/v1`

| Method | Endpoint | Deskripsi | Query / Param | Status Code |
| :--- | :--- | :--- | :--- | :--- |
| `GET` | `/health` | Pemeriksaan kesehatan server & database | - | `200 OK`, `503 Service Unavailable` |
| `GET` | `/students` | Daftar mahasiswa (pagination, sort, search, filter) | `page`, `limit`, `sort`, `order`, `search`, `grade`, `is_active` | `200 OK`, `500 Internal Server Error` |
| `GET` | `/students/:id` | Mengambil detail 1 mahasiswa berdasarkan ID | Path: `id` | `200 OK`, `400 Bad Request`, `404 Not Found` |
| `POST` | `/students` | Menambah data mahasiswa baru | Body: `nim`, `name`, `grade` | `201 Created`, `400 Bad Request`, `409 Conflict`, `500 Internal Server Error` |
| `PUT` | `/students/:id` | Mengganti seluruh data mahasiswa (*full update*) | Path: `id`, Body: `nim`, `name`, `grade`, `is_active` | `200 OK`, `400 Bad Request`, `404 Not Found`, `409 Conflict` |
| `PATCH` | `/students/:id` | Memperbarui sebagian atribut mahasiswa | Path: `id`, Body: opsional `nim`, `name`, `grade`, `is_active` | `200 OK`, `400 Bad Request`, `404 Not Found`, `409 Conflict` |
| `DELETE` | `/students/:id` | Menghapus data mahasiswa | Path: `id` | `204 No Content`, `400 Bad Request`, `404 Not Found` |

---

### Contoh Format Request & Response

#### 1. `GET /api/v1/students`
**Query Parameters:**
- `page`: Nomor halaman (default: `1`)
- `limit`: Jumlah item per halaman (default: `10`)
- `sort`: Field pengurutan (`id`, `nim`, `name`, `grade`, `created_at`) (default: `id`)
- `order`: Arah pengurutan (`asc` / `desc`) (default: `asc`)
- `search`: Pencarian nama atau NIM (contoh: `?search=ahmad`)
- `grade`: Filter nilai (contoh: `?grade=A`)
- `is_active`: Filter keaktifan (`true` / `false`)

**Response (`200 OK`):**
```json
{
  "message": "daftar mahasiswa berhasil diambil",
  "data": [
    {
      "id": 1,
      "nim": "220101001",
      "name": "Ahmad Dani",
      "grade": "A",
      "is_active": true,
      "created_at": "2026-08-30T10:00:00Z"
    }
  ],
  "meta": {
    "page": 1,
    "limit": 10,
    "total": 1,
    "total_pages": 1
  }
}
```

#### 2. `POST /api/v1/students`
**Request Body:**
```json
{
  "nim": "220101002",
  "name": "Siti Aminah",
  "grade": "A"
}
```
**Response (`201 Created`):**
*Header: `Location: /api/v1/students/2`*
```json
{
  "message": "mahasiswa berhasil dibuat",
  "data": {
    "id": 2,
    "nim": "220101002",
    "name": "Siti Aminah",
    "grade": "A",
    "is_active": true,
    "created_at": "2026-08-30T10:05:00Z"
  }
}
```

#### 3. `PUT /api/v1/students/:id`
**Request Body:**
```json
{
  "nim": "220101001",
  "name": "Ahmad Dani Pratama",
  "grade": "A+",
  "is_active": true
}
```
**Response (`200 OK`):**
```json
{
  "message": "mahasiswa berhasil diganti seluruhnya",
  "data": {
    "id": 1,
    "nim": "220101001",
    "name": "Ahmad Dani Pratama",
    "grade": "A+",
    "is_active": true,
    "created_at": "2026-08-30T10:00:00Z"
  }
}
```

#### 4. `PATCH /api/v1/students/:id`
**Request Body:**
```json
{
  "grade": "B+",
  "is_active": false
}
```
**Response (`200 OK`):**
```json
{
  "message": "mahasiswa berhasil diperbarui sebagian",
  "data": {
    "id": 1,
    "nim": "220101001",
    "name": "Ahmad Dani Pratama",
    "grade": "B+",
    "is_active": false,
    "created_at": "2026-08-30T10:00:00Z"
  }
}
```

---

## 🧪 Skenario Pengujian Khusus (Tugas Praktikum)

Untuk keperluan bukti tangkapan layar (*screenshot*) pengujian respon kode status tertentu:

### 1. Status `404 Not Found`
* **Situasi**: Memanggil ID mahasiswa yang tidak terdaftar di database.
* **Request**: `GET http://localhost:3000/api/v1/students/99999`
* **Hasil**:
  ```json
  {
    "message": "mahasiswa tidak ditemukan"
  }
  ```

### 2. Status `409 Conflict`
* **Situasi**: Menambah data mahasiswa baru atau mengubah NIM dengan nilai NIM yang sudah terdaftar sebelumnya.
* **Request**: `POST http://localhost:3000/api/v1/students`
* **Body**:
  ```json
  {
    "nim": "220101001",
    "name": "Mahasiswa Duplikat",
    "grade": "A"
  }
  ```
* **Hasil**:
  ```json
  {
    "message": "nim sudah dipakai"
  }
  ```

### 3. Status `503 Service Unavailable` / `500 Internal Server Error`
* **Situasi**: Layanan PostgreSQL dimatikan saat server API tetap berjalan.
* **Pengujian 1 (503 Service Unavailable)**:
  - Request: `GET http://localhost:3000/api/v1/health`
  - Hasil:
    ```json
    {
      "message": "database tidak dapat dihubungi"
    }
    ```
* **Pengujian 2 (500 Internal Server Error)**:
  - Request: `GET http://localhost:3000/api/v1/students`
  - Hasil:
    ```json
    {
      "message": "gagal mengambil data mahasiswa"
    }
    ```
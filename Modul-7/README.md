# Modul 7 — Advanced API Design

Proyek mandiri **Modul 7: Advanced API Design** untuk mata kuliah **Praktikum Pemrograman Backend Lanjut**, D4 Teknik Informatika, Fakultas Vokasi, Universitas Airlangga.

- **Nama**: Muhammad Fadhil Ilyas
- **NIM**: 434241099

---

## 🚀 Fitur Utama Modul 7

1. **Validasi Deklaratif (`go-playground/validator/v10`)**:
   - Menggantikan validasi imperatif panjang dengan aturan tag pada struct model (`required`, `min`, `max`, `email`, `alphanum`, `oneof`).
   - Objek `validator.Validate` diinisialisasi sekali secara global (*singleton with reflection cache*).
   - Pemetaan nama field JSON otomatis (`RegisterTagNameFunc`) agar pesan error sesuai kontrak JSON client.
   - Custom validator domain:
     - `nospace`: memastikan string tidak memuat whitespace.
     - `username`: format karakter username (`a-z`, `0-9`, `.`, `_`).
     - `strongpassword`: validasi kekuatan kata sandi murni (minimal 8 karakter, kombinasi huruf & angka, dictionary check).
     - `nim`: validasi NIM mahasiswa (9–18 digit angka murni).
   - Seluruh pesan error divalidasi dan diterjemahkan terpusat ke dalam Bahasa Indonesia yang komunikatif (`messageFor`).

2. **Perilaku Pointer & `omitnil` pada Endpoint PATCH**:
   - Menggunakan field pointer (`*string`, `*bool`) pada struct request partial update (`PatchUserRequest`, `PatchStudentRequest`).
   - Tag `omitnil` memastikan field yang tidak dikirim (`nil`) dilewati secara aman, sedangkan field yang dikirim bernilai string kosong `""` tetap divalidasi dan ditolak dengan status **422 Unprocessable Content**.

3. **Keyset Pagination (Cursor-based Pagination)**:
   - Menghilangkan inefisiensi `OFFSET` dan anomali data bergeser (*duplicate records / skipped records*).
   - Menggunakan penanda posisi (*cursor*) berbasis tuple leksikografis `(created_at, id)`. Kolom `id` bertindak sebagai tie-breaker unik jika timestamp identik.
   - Menggunakan format URL-safe Base64 encoding: `base64(nanoseconds|id)`.
   - Mengimplementasikan composite index `(created_at DESC, id DESC)` di PostgreSQL untuk eksekusi *Index Scan* instan tanpa *in-memory sort*.
   - Metadata respon minimalis dan efisien: `limit`, `next_cursor`, `has_more` tanpa membebani server dengan `COUNT(*)`.

4. **Content Negotiation (JSON & CSV)**:
   - Membedakan peran header `Content-Type` (format data yang dikirim client) dan header `Accept` (format balasan yang diinginkan client).
   - Mendukung format output `application/json` dan `text/csv` pada endpoint daftar (`/users` dan `/students`).
   - Request dengan header `Accept: text/csv` menghasilkan file CSV terunduh (`Content-Disposition: attachment; filename="*.csv"`).
   - Request dengan format yang tidak didukung (misal `Accept: application/xml`) ditolak dengan status **406 Not Acceptable**.
   - Header `Accept: */*` atau kosong dijawab dengan format bawaan (JSON) status 200 OK.
   - Penulisan CSV memanfaatkan `csv.Writer` dengan pemanggilan `Flush()` eksplisit.

5. **Centralized Error Handling (Error Handling Terpusat)**:
   - Seluruh service dan middleware hanya mengembalikan objek `AppError` tanpa membentuk response HTTP sendiri.
   - Satu `ErrorHandler` terpusat di Fiber mengonversi seluruh kegagalan menjadi format JSON standar:
     ```json
     {
       "success": false,
       "code": "ERROR_CODE",
       "message": "Pesan deskriptif untuk manusia",
       "fields": { "field_name": "detail kesalahan" },
       "request_id": "uuid-tracing"
     }
     ```
   - Enkapsulasi error internal: pesan asli database disimpan pada unexported field `cause` yang hanya dicatat di log server bersama request_id; detail teknis tidak pernah bocor ke client.
   - Pemisahan tingkat log berdasarkan penanggung jawab kegagalan:
     - **Status 4xx (Client Error)** dicatat dengan level `WARN` (`msg: "request_rejected"`).
     - **Status 5xx (Server Error)** dicatat dengan level `ERROR` (`msg: "request_failed"`).

---

## 📁 Struktur Direktori

```text
Modul-7/
├── go.mod
├── go.sum
├── main.go
├── main_test.go
├── config/
│   ├── app.go
│   ├── env.go
│   └── logger.go
├── database/
│   └── postgre.go
├── handlers/
│   └── handlers.go
├── middleware/
│   ├── auth.go
│   ├── authz.go
│   └── middleware.go
├── models/
│   └── models.go
├── app/
│   ├── model/
│   │   ├── auth.go
│   │   ├── nilai.go
│   │   ├── student.go
│   │   └── user.go
│   ├── repository/
│   │   ├── nilai_repository.go
│   │   ├── role_repository.go
│   │   ├── student_repository.go
│   │   ├── token_repository.go
│   │   └── user_repository.go
│   └── service/
│       ├── auth_service.go
│       ├── auth_rules_test.go
│       ├── authz_rules.go
│       ├── nilai_service.go
│       ├── student_authz_rules.go
│       ├── student_rules.go
│       ├── student_rules_test.go
│       ├── student_service.go
│       ├── user_rules.go
│       └── user_service.go
├── validators/
│   └── validator.go
├── pagination/
│   └── cursor.go
├── errors/
│   └── errors.go
├── helper/
│   ├── authz.go
│   ├── context.go
│   ├── cursor.go
│   ├── errors.go
│   ├── jwt.go
│   ├── negotiate.go
│   ├── request.go
│   ├── response.go
│   ├── security.go
│   └── validator.go
├── migrations/
│   ├── 001_create_student.sql
│   ├── 002_create_nilai.sql
│   ├── 003_auth.sql
│   ├── 003_rbac.sql
│   ├── 004_student_permissions.sql
│   ├── 004_cursor_index.sql
│   └── 005_student_cursor_index.sql
├── tests/
│   ├── cursor_test.go
│   ├── negotiate_test.go
│   └── validator_test.go
├── laporan/
│   ├── Tugas7_434241099.pdf
│   └── screenshots/
└── README.md
```

---

## 🛠️ Panduan Menjalankan & Pengujian

### 1. Prasyarat & Konfigurasi
Pastikan PostgreSQL telah aktif dan database `praktikum_backend` telah dikonfigurasi pada `.env`:
```env
APP_PORT=3000
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=123
DB_NAME=praktikum_backend
DB_SSLMODE=disable
JWT_SECRET=2441502a8d96bad551072bc97083cd7e16a90d09f056c43330b9ed9c4c1512f2
```

### 2. Migrasi Indeks Cursor
Jalankan migrasi komposit indeks keyset pagination:
```powershell
$env:PGPASSWORD="123"
psql -h localhost -U postgres -d praktikum_backend -f migrations/004_cursor_index.sql
psql -h localhost -U postgres -d praktikum_backend -f migrations/005_student_cursor_index.sql
```

### 3. Build & Jalankan Aplikasi
```powershell
# Kompilasi seluruh package
go build ./...

# Jalankan server
go run main.go
```

### 4. Eksekusi Seluruh Pengujian Otomatis
```powershell
go test -v ./...
```

---

## 📋 Berita Acara Perbaikan 9 Kesalahan Bagian B

| No | Berkas dan Baris | Jenis | Gejala | Akar Masalah | Perbaikan | Bukti |
|:--:|:---|:---:|:---|:---|:---|:---|
| 1 | `helper/errors.go` (Langkah 1) | Lolos compiler, salah perilaku | Respon validasi menghasilkan status HTTP 400 | Fungsi `Validation` menggunakan `fiber.StatusBadRequest` alih-alih 422 Unprocessable Content. | `Status: fiber.StatusUnprocessableEntity` | Request invalid register menghasilkan HTTP 422 (C.8 butir 2) |
| 2 | `config/app.go` (Langkah 2) | Ditolak compiler | `appErr.cause undefined` | Field `cause` bersifat unexported di package `helper` sehingga tidak dapat diakses langsung dari package `config`. | Ganti `appErr.cause.Error()` dengan `appErr.Error()` | `go build ./...` sukses tanpa error unexported field (C.8 butir 1) |
| 3 | `config/app.go` (Langkah 2) | Lolos compiler, salah perilaku | 4xx dicatat ERROR (`request_failed`), 5xx dicatat WARN (`request_rejected`) | Logika if terbalik (`appErr.Status < 500`), membalikkan level log kesalahan klien dan server. | Ubah kondisi menjadi `if appErr.Status >= fiber.StatusInternalServerError` | 404 dicatat WARN, 500 dicatat ERROR (C.7 & C.8 butir 8) |
| 4 | `middleware/middleware.go` (Langkah 3) | Ditolak compiler & salah perilaku | `status declared and not used`; status di log tercatat 200 | Atribut log menggunakan `c.Response().StatusCode()` yang belum terisi ErrorHandler alih-alih variabel `status`. | Gunakan `slog.Int("status", status)` pada `attrs` | Status pada access log sesuai respon klien (C.8 butir 9) |
| 5 | `app/service/user_service.go` (Langkah 4) | Lolos compiler, salah perilaku | Error query database membalas HTTP 200 OK dengan objek kosong | Blok default pada `translateError` mengembalikan `nil` alih-alih error internal. | Ubah `default:` menjadi `return helper.Internal(err)` | Kegagalan DB membalas 500 INTERNAL_ERROR (C.8 butir 6) |
| 6 | `helper/validator.go` (Langkah 5) | Lolos compiler, salah perilaku | Password kuat ditolak dengan string kosong `""`, password lemah lolos | Evaluasi custom validator `strongpassword` menggunakan `!= ""` padahal fungsi mengembalikan `""` saat valid. | Ubah evaluasi menjadi `== ""` | Password kuat diterima, password lemah ditolak (C.8 butir 3 & 4) |
| 7 | `app/model/user.go` (Langkah 6) | Ditolak compiler | `invalid operation: req.Username != nil` | Field `Username` pada `PatchUserRequest` bertipe `string` bukan pointer `*string`. | Ganti tipe field menjadi `Username *string` | PATCH `{"username":""}` ditolak 422, PATCH valid diterima 200 (C.8 butir 5) |
| 8 | `app/repository/user_repository.go` (Langkah 7) | Lolos compiler, salah perilaku | Halaman 1 terurut dari user terlama; pagination cursor rusak | Query SQL menggunakan `ORDER BY created_at ASC, id ASC` berlawanan dengan klausul keyset `<` dan indeks DESC. | Ubah menjadi `ORDER BY created_at DESC, id DESC` | Data terurut dari yang terbaru, halaman 2 tidak memuat duplikat (C.8 butir 10 & 11) |
| 9 | `helper/negotiate.go` (Langkah 8) | Lolos compiler, salah perilaku | Respon `Accept: text/csv` menghasilkan file CSV kosong (0 byte) | `csv.Writer` tidak memanggil `writer.Flush()` sebelum membaca isi `strings.Builder`. | Tambahkan `writer.Flush()` sebelum mengembalikan data buffer | Respon CSV berisi header dan baris data lengkap (C.8 butir 13) |

---

## 🔒 Matriks Respon Kode Error & Status HTTP

| Endpoint / Permintaan | Status HTTP | Kode Error (`code`) | Keterangan / Pemicu |
|:---|:---:|:---:|:---|
| `GET /api/v1/users/999999` | 404 | `NOT_FOUND` | User dengan ID yang diminta tidak ditemukan di database. |
| `GET /api/v1/users?cursor=bukanbase64!!` | 400 | `BAD_REQUEST` | Nilai cursor pada query string bukan base64 valid. |
| `GET /api/v1/tidakada` | 404 | `NOT_FOUND` | Path URL tidak terdaftar di router (*fallback handler*). |
| `GET /api/v1/users` (tanpa token) | 401 | `UNAUTHORIZED` | Header `Authorization: Bearer <token>` tidak disertakan atau token tidak valid. |
| `POST /api/v1/users` (`Content-Type: text/plain`) | 415 | `UNSUPPORTED_MEDIA_TYPE` | Body request tidak menggunakan MIME `application/json`. |
| `POST /api/v1/users` (JSON rusak) | 400 | `BAD_REQUEST` | Syntax payload JSON tidak valid atau malformed. |
| `GET /api/v1/users?is_active=mungkin` | 400 | `BAD_REQUEST` | Tipe data filter query string tidak sesuai (bukan boolean `true`/`false`). |
| `GET /api/v1/users` (`Accept: application/xml`) | 406 | `NOT_ACCEPTABLE` | Format balasan yang diminta client tidak tersedia di server. |
| `POST /api/v1/auth/register` (aturan tag dilanggar) | 422 | `VALIDATION_ERROR` | Nilai field pada body JSON melanggar aturan validasi deklaratif tag. |
| Database Down / Skema Rusak | 500 | `INTERNAL_ERROR` | Kesalahan internal server / koneksi database; detail teknis disembunyikan. |

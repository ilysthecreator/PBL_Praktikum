# Modul 10: API Polyglot dengan PostgreSQL dan MongoDB

Dokumentasi implementasi praktikum Pertemuan 10 — Pengembangan API dengan arsitektur multi-database (Polyglot Persistence) menggabungkan **PostgreSQL** (relational/ACID) dan **MongoDB** (document/NoSQL) menggunakan Go (Fiber) dan driver resmi `go.mongodb.org/mongo-driver/v2`.

---

## 1. Arsitektur Polyglot Persistence

Aplikasi ini membagi tanggung jawab penyimpanan data berdasarkan karakteristik dan pola aksesnya:

| Aspek | PostgreSQL | MongoDB |
|---|---|---|
| **Koleksi / Entitas** | `users`, `students`, `courses`, `enrollments`, `supervisors`, `roles` | `course_reviews`, `student_portfolios` |
| **Karakteristik** | Relasional, relasi kuat, integritas referensial (FK), transaksi ACID, row-level locking (`SELECT ... FOR UPDATE`). | Dokumen hierarkis (BSON), skema fleksibel, array tersarang (*embedded*), atomic operators (`$push`, `$pull`, `$addToSet`, `$inc`). |
| **Pola Akses** | Validasi keberadaan entitas, autentikasi/otorisasi, KRS dengan batasan kuota ketat. | Ulasan publik, ringkasan rating & tag terpopuler (*aggregation facet*), portofolio & proyek mahasiswa. |

Service bertindak sebagai satu-satunya layer yang menghubungkan kedua database (*polyglot boundary*). Sebelum menulis ke MongoDB, service melakukan verifikasi langsung ke PostgreSQL untuk memastikan entitas referensi benar-benar valid.

---

## 2. Struktur Proyek

```text
Modul-10/
├── app/
│   ├── model/
│   │   ├── review.go               # Model domain ulasan (bebas dari tipe driver MongoDB)
│   │   ├── portfolio.go            # Model domain portofolio & proyek mahasiswa
│   │   ├── course.go
│   │   ├── student.go
│   │   └── user.go
│   ├── repository/
│   │   ├── review_repository.go    # Operasi MongoDB ulasan (EnsureIndexes, Aggregation, Atomic)
│   │   ├── portfolio_repository.go # Operasi MongoDB portofolio (Upsert, Push limit, Pull, Unwind)
│   │   ├── course_repository.go    # PostgreSQL course & pengecekan IsEnrolled
│   │   └── student_repository.go
│   └── service/
│       ├── review_rules.go         # Normalisasi tag ulasan
│       ├── review_service.go       # Logika bisnis ulasan (validasi KRS di PostgreSQL)
│       └── portfolio_service.go    # Logika bisnis portofolio (validasi mhs di PostgreSQL)
├── config/
│   ├── app.go
│   └── logger.go
├── database/
│   ├── db.go                       # Pool koneksi PostgreSQL (pgxpool)
│   └── mongo.go                    # Client koneksi MongoDB (driver v2) dengan Ping saat startup
├── helper/
│   └── validator.go                # Custom validator penanganan min/max string dan slice
├── migrations/
│   └── 008_review_permissions.sql  # Permission review:moderate
├── route/
│   └── route.go                    # Routing API & Dual Health Check
├── latihan.js                      # Skrip mongosh panduan Langkah 2
├── test_runner.py                  # Skrip verifikasi & pengujian konkurensi otomatis
├── main.go                         # Inisialisasi pool, mongo, index, service, dan graceful shutdown
├── .env.example
└── README.md
```

---

## 3. Menjalankan Aplikasi

### A. Konfigurasi Environment (`.env`)
Salin file `.env.example` menjadi `.env` dan sesuaikan nilainya:
```env
APP_PORT=3000
DB_HOST=127.0.0.1
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=123
DB_NAME=praktikum_backend
MONGO_URI=mongodb://localhost:27017
MONGO_DATABASE=praktikum_backend
MONGO_MAX_POOL=50
JWT_SECRET=super-secret-key-that-is-at-least-32-chars-long
JWT_ISSUER=praktikum-backend
JWT_ACCESS_TTL_MINUTES=60
JWT_REFRESH_TTL_DAYS=7
```

### B. Migrasi Database PostgreSQL
Jalankan migrasi permission ulasan:
```bash
psql -U postgres -h 127.0.0.1 -d praktikum_backend -f migrations/008_review_permissions.sql
```

### C. Menjalankan Server
```bash
go run .
```
Saat pertama kali menyala, aplikasi akan memverifikasi koneksi kedua database serta secara otomatis mengeksekusi `EnsureIndexes` untuk koleksi MongoDB:
- `course_user_unique` (unique `{course_id: 1, user_id: 1}`)
- `course_newest` (`{course_id: 1, _id: -1}`)
- `course_tags` (`{course_id: 1, tags: 1}`)
- `student_unique` (unique `{student_id: 1}`)
- `project_technologies` (`{"projects.technologies": 1}`)

---

## 4. Fitur Ulasan Kelas (Bagian B)

### Rute API:
- `POST /api/v1/courses/:id/reviews` — Menulis ulasan (hanya peserta kelas terdaftar di PostgreSQL yang diizinkan).
- `GET /api/v1/courses/:id/reviews` — Melihat daftar ulasan kelas (cursor pagination dengan `limit` dan `before`, filter `tag`).
- `GET /api/v1/courses/:id/reviews/summary` — Ringkasan statistik bintang rata-rata, sebaran bintang 1-5, dan 5 tag terpopuler via `$facet`.
- `GET /api/v1/reviews/:id` — Detail ulasan.
- `PATCH /api/v1/reviews/:id` — Mengubah ulasan sendiri (`$set`).
- `DELETE /api/v1/reviews/:id` — Menghapus ulasan sendiri atau oleh pemegang izin `review:moderate`.
- `POST /api/v1/reviews/:id/replies` — Menambah balasan (dibatasi maksimal 50 balasan secara atomik dengan syarat `"replies.49": {"$exists": false}`).
- `POST /api/v1/reviews/:id/helpful` — Menandai bermanfaat secara atomik (`$addToSet` dan `$inc: {helpful_count: 1}`).

---

## 5. Fitur Portofolio Mahasiswa (Tugas Mandiri C)

### C.1 Rancangan Dokumen MongoDB (`student_portfolios`)
Setiap mahasiswa memiliki tepat satu dokumen portofolio. Skema dokumen lengkap:

```json
{
  "_id": { "$oid": "6aeee9cc28a7200b307edd6b" },
  "student_id": 1,
  "headline": "Full-Stack Developer & Cloud Enthusiast",
  "bio": "Mahasiswa tingkat akhir dengan ketertarikan mendalam pada arsitektur microservices dan basis data NoSQL.",
  "projects": [
    {
      "_id": { "$oid": "6aeee9d035a52c16c2a4d50c" },
      "title": "Sistem Informasi Akademik Terdistribusi",
      "technologies": ["go", "fiber", "postgresql", "mongodb", "docker"],
      "link": "https://github.com/mahasiswa/siakad-modul10",
      "year": 2024
    }
  ],
  "certificates": [
    {
      "_id": { "$oid": "6aeee9cc35a52c16c2a4d50a" },
      "name": "AWS Certified Solutions Architect",
      "issuer": "Amazon Web Services",
      "date": "2024-03"
    },
    {
      "_id": { "$oid": "6aeee9cc35a52c16c2a4d50b" },
      "name": "Certified Kubernetes Administrator",
      "issuer": "Cloud Native Computing Foundation",
      "date": "2024-08"
    }
  ],
  "created_at": { "$date": "2026-11-07T09:05:16.726Z" },
  "updated_at": { "$date": "2026-11-07T09:05:20.809Z" }
}
```

### C.2 Endpoints Portofolio:
1. `GET /api/v1/students/:id/portfolio` — Membaca portofolio. Menolak (404) jika mahasiswa belum ada di PostgreSQL.
2. `PUT /api/v1/students/:id/portfolio` — Membuat atau mengganti portofolio secara *upsert* (`SetUpsert(true)`).
3. `POST /api/v1/students/:id/portfolio/projects` — Menambah proyek ke dalam array `projects` menggunakan `$push`, dibatasi maksimal 20 proyek dengan filter `"projects.19": {"$exists": false}`.
4. `DELETE /api/v1/students/:id/portfolio/projects/:projectId` — Menghapus proyek dari array `projects` menggunakan `$pull`.
5. `GET /api/v1/portfolios/stats/technologies` — Aggregation pipeline menghitung 10 teknologi terbanyak dipakai:
   ```go
   mongo.Pipeline{
       {{Key: "$unwind", Value: "$projects"}},
       {{Key: "$unwind", Value: "$projects.technologies"}},
       {{Key: "$group", Value: bson.M{
           "_id": "$projects.technologies",
           "count": bson.M{"$sum": 1},
       }}},
       {{Key: "$sort", Value: bson.D{
           {Key: "count", Value: -1},
           {Key: "_id", Value: 1},
       }}},
       {{Key: "$limit", Value: 10}},
       {{Key: "$project", Value: bson.M{
           "_id": 0,
           "technology": "$_id",
           "count": 1,
       }}},
   }
   ```

---

## 6. Bukti Pengujian (C.3)

Semua pengujian telah diverifikasi melalui `test_runner.py` dan `curl`:

### 1. Dual Health Check (`GET /api/v1/health`)
```bash
curl -s http://localhost:3000/api/v1/health
```
**Response:**
```json
{
  "success": true,
  "message": "server, PostgreSQL, dan MongoDB berjalan"
}
```

### 2. Upsert Portofolio Mahasiswa (`PUT /api/v1/students/:id/portfolio`)
```bash
curl -s -X PUT http://localhost:3000/api/v1/students/1/portfolio \
  -H "Authorization: Bearer <TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
    "headline": "Full-Stack Developer & Cloud Enthusiast",
    "bio": "Mahasiswa tingkat akhir dengan ketertarikan mendalam pada arsitektur microservices dan basis data NoSQL.",
    "certificates": [
      {
        "name": "AWS Certified Solutions Architect",
        "issuer": "Amazon Web Services",
        "date": "2024-03"
      }
    ]
  }'
```
**Response:**
```json
{
  "success": true,
  "message": "portofolio berhasil disimpan",
  "data": {
    "id": "6aeee9cc28a7200b307edd6b",
    "student_id": 1,
    "headline": "Full-Stack Developer & Cloud Enthusiast",
    "bio": "Mahasiswa tingkat akhir dengan ketertarikan mendalam pada arsitektur microservices dan basis data NoSQL.",
    "projects": [],
    "certificates": [
      {
        "id": "6aeee9cc35a52c16c2a4d50a",
        "name": "AWS Certified Solutions Architect",
        "issuer": "Amazon Web Services",
        "date": "2024-03"
      }
    ],
    "created_at": "2026-11-07T09:05:16.726Z",
    "updated_at": "2026-11-07T09:05:16.726Z"
  }
}
```

### 3. Tambah Proyek Atomik (`POST /api/v1/students/:id/portfolio/projects`)
```bash
curl -s -X POST http://localhost:3000/api/v1/students/1/portfolio/projects \
  -H "Authorization: Bearer <TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Sistem Informasi Akademik Terdistribusi",
    "technologies": ["Go", "Fiber", "PostgreSQL", "MongoDB", "Docker"],
    "link": "https://github.com/mahasiswa/siakad-modul10",
    "year": 2024
  }'
```
**Response:**
```json
{
  "success": true,
  "message": "proyek berhasil ditambahkan",
  "data": {
    "id": "6aeee9cc28a7200b307edd6b",
    "student_id": 1,
    "headline": "Full-Stack Developer & Cloud Enthusiast",
    "bio": "Mahasiswa tingkat akhir dengan ketertarikan mendalam pada arsitektur microservices dan basis data NoSQL.",
    "projects": [
      {
        "id": "6aeee9d035a52c16c2a4d50c",
        "title": "Sistem Informasi Akademik Terdistribusi",
        "technologies": ["go", "fiber", "postgresql", "mongodb", "docker"],
        "link": "https://github.com/mahasiswa/siakad-modul10",
        "year": 2024
      }
    ],
    "created_at": "2026-11-07T09:05:16.726Z",
    "updated_at": "2026-11-07T09:05:20.809Z"
  }
}
```

### 4. Hapus Proyek Atomik (`DELETE /api/v1/students/:id/portfolio/projects/:projectId`)
```bash
curl -s -X DELETE http://localhost:3000/api/v1/students/1/portfolio/projects/6aeee9d035a52c16c2a4d50c \
  -H "Authorization: Bearer <TOKEN>"
```
**Status: `204 No Content`**

### 5. Statistik Top 10 Teknologi (`GET /api/v1/portfolios/stats/technologies`)
```bash
curl -s http://localhost:3000/api/v1/portfolios/stats/technologies \
  -H "Authorization: Bearer <TOKEN>"
```
**Response:**
```json
{
  "success": true,
  "message": "statistik teknologi berhasil diambil",
  "data": [
    { "technology": "go", "count": 36 },
    { "technology": "docker", "count": 18 },
    { "technology": "postgresql", "count": 18 },
    { "technology": "kubernetes", "count": 4 },
    { "technology": "mongodb", "count": 4 }
  ]
}
```

### 6. Bukti Uji Konkurensi Batas 20 Proyek
Dilakukan dengan mengisi 18 proyek awal pada portofolio mahasiswa, kemudian mengirim **10 permintaan serentak**:
- **Hasil:**
  - Status `201 Created` : **tepat 2 permintaan**
  - Status `409 Conflict` : **tepat 8 permintaan** (`PROJECT_LIMIT`: "jumlah proyek sudah mencapai batas maksimal 20")
  - Jumlah proyek akhir di database: **tepat 20 proyek**
- **Mekanisme Atomik:**
  Filter query MongoDB: `{"student_id": 34, "projects.19": {"$exists": false}}`. Operasi dilakukan langsung di dalam engine database; saat proyek ke-19 dan ke-20 berhasil masuk, `projects.19` menjadi ada, sehingga 8 request lainnya langsung gagal menemukan dokumen yang cocok dan ditolak secara aman tanpa race condition.

### 7. Kasus-Kasus Gagal yang Ditangani:
1. **Kasus 1 — Mahasiswa Tidak Terdaftar di PostgreSQL (`GET /api/v1/students/99999/portfolio`):**
   ```json
   {
     "success": false,
     "code": "NOT_FOUND",
     "message": "mahasiswa tidak ditemukan"
   }
   ```
2. **Kasus 2 — Validasi Struct Gagal / Field Wajib Kosong (`POST /api/v1/students/1/portfolio/projects` dengan title kosong):**
   ```json
   {
     "success": false,
     "code": "VALIDATION_ERROR",
     "message": "validasi gagal",
     "fields": {
       "title": "wajib diisi"
     }
   }
   ```
3. **Kasus 3 — Format ObjectID Tidak Valid (`DELETE /api/v1/students/1/portfolio/projects/bukan-hex-id`):**
   ```json
   {
     "success": false,
     "code": "BAD_REQUEST",
     "message": "format id tidak valid"
   }
   ```
4. **Kasus 4 — Percobaan NoSQL Injection (`PUT /api/v1/students/1/portfolio` dengan body `{"headline": {"$ne": null}}`):**
   ```json
   {
     "success": false,
     "code": "BAD_REQUEST",
     "message": "body harus berupa JSON yang valid"
   }
   ```
   *Penjelasan:* Pada Go, parser Fiber men-decode JSON langsung ke dalam struct yang bertipe data ketat (`Headline string`). Ketika penyerang menyisipkan operator NoSQL seperti `{"$ne": null}`, decoder menolak karena tipe objek JSON tidak dapat di-parse ke string Go. Selain itu, parameter query pada repository Go selalu dikonversi menjadi BSON eksplisit tanpa interpretasi query string dinamis, sehingga aplikasi kebal terhadap NoSQL Injection.

### 8. Bukti `explain("executionStats")` Penggunaan Index
Pengujian index melalui query MongoDB:
- **Index `student_unique` pada `student_portfolios`:**
  - Query: `db.student_portfolios.find({student_id: 1})`
  - Stage: `IXSCAN` pada `indexName: "student_unique"` dilanjutkan `FETCH`.
  - `nReturned: 1`, `totalDocsExamined: 1`, `executionTimeMillis: 0`. Database langsung melompat ke dokumen target tanpa memindai seluruh koleksi (*COLLSCAN*).
- **Index `project_technologies` pada `student_portfolios`:**
  - Query: `db.student_portfolios.find({"projects.technologies": "go"})`
  - Stage: `IXSCAN` pada multikey index `indexName: "project_technologies"`.
  - `totalDocsExamined: 1`.

---

## 7. Analisis Singkat (C.4)

### 10. Mengapa `projects` disarangkan (embedded) dan bukan disimpan di collection tersendiri? Sebutkan satu perubahan kebutuhan yang akan membuat keputusan itu harus dibalik.
**Jawaban:**  
Data proyek disarangkan (*embedded*) di dalam dokumen `student_portfolios` karena memiliki hubungan 1:N dengan kardinalitas terbatas (dibatasi maksimal 20 item per mahasiswa) dan pola aksesnya selalu dibaca serta diperbarui bersamaan dengan profil portofolio (*1-to-1 aggregate retrieval*). Pola *embedding* mengeliminasi *overhead* join antar-koleksi serta memungkinkan pembaruan atomik (`$push`, `$pull`) dalam satu perjalanan jaringan tanpa membutuhkan transaksi terdistribusi yang mahal. Keputusan ini harus dibalik (dipisahkan menjadi koleksi tersendiri) apabila muncul kebutuhan di mana proyek dapat dikerjakan secara kolaboratif lintas mahasiswa (relasi M:N antar mahasiswa dan proyek), proyek memiliki riwayat revisi dan lampiran berkas yang berpotensi menembus batas ukuran dokumen MongoDB 16 MB, atau proyek perlu diindeks dan dijelajahi secara mandiri sebagai entitas independen (misalnya fitur *explore/showcase* portofolio proyek di tingkat kampus).

### 11. Bila seorang mahasiswa dihapus dari PostgreSQL, apa yang terjadi pada portofolionya di MongoDB? Usulkan dua cara menanganinya beserta kelemahan masing-masing.
**Jawaban:**  
Karena PostgreSQL dan MongoDB merupakan dua database yang terpisah tanpa mekanisme *FOREIGN KEY CASCADE* lintas engine, penghapusan baris mahasiswa di PostgreSQL akan menyebabkan dokumen portofolionya di MongoDB tertinggal sebagai data yatim (*orphaned document*). Dua cara untuk menanganinya adalah:  
1. **Penghapusan Berantai di Application Layer (*Synchronous Cascading Delete*):** Pada handler atau service `DELETE /students/:id`, aplikasi terlebih dahulu menghapus data mahasiswa di PostgreSQL, lalu segera memanggil `portfolioRepo.DeleteByStudentID` di MongoDB. Kelemahannya adalah tidak adanya transaksi atomik terdistribusi (*Two-Phase Commit*); jika terjadi server *crash* atau jaringan ke MongoDB terputus setelah PostgreSQL melakukan *commit*, dokumen di MongoDB akan tetap tertinggal tanpa ada mekanisme *rollback* otomatis.  
2. **Asynchronous Event-Driven / Outbox Pattern (atau Periodic Garbage Collector):** Ketika mahasiswa dihapus di PostgreSQL, catat sebuah event pada tabel *outbox* atau kirim pesan ke message broker yang kemudian dikonsumsi oleh pekerja latar (*worker*) untuk menghapus dokumen terkait di MongoDB, atau jalankan cron job rekonsiliasi berkala. Kelemahannya adalah terjadi inkonsistensi sementara (*eventual consistency*) di mana portofolio masih dapat diakses selama beberapa waktu setelah mahasiswa dihapus, serta memerlukan kompleksitas infrastruktur tambahan (*message queue* atau *scheduler*).

### 12. Endpoint mana di aplikasi Anda yang tetap berjalan bila MongoDB mati, dan mana yang tidak? Apakah perilaku itu yang Anda inginkan?
**Jawaban:**  
Endpoint yang tetap berjalan saat MongoDB mati adalah seluruh endpoint yang proses bisnisnya hanya bergantung pada PostgreSQL, yaitu seluruh endpoint autentikasi (`/api/v1/auth/*`), manajemen pengguna (`/api/v1/users/*`), operasi CRUD mahasiswa dasar (`/api/v1/students` tanpa `/portfolio`), rekap nilai (`/api/v1/nilai/*`), kelas dan KRS (`/api/v1/courses`, `/api/v1/enrollments`), serta pembimbing skripsi (`/api/v1/supervisors`). Sebaliknya, endpoint yang akan gagal adalah endpoint dual health check (`/api/v1/health` akan menghasilkan 503 Service Unavailable karena MongoDB tidak merespons ping), seluruh fitur ulasan kelas (`/api/v1/courses/:id/reviews*` dan `/api/v1/reviews/*`), serta seluruh fitur portofolio mahasiswa (`/api/v1/students/:id/portfolio*` dan `/api/v1/portfolios/*`). Perilaku ini sangat tepat dan sesuai dengan prinsip *graceful degradation* dan *fault isolation*: gangguan pada fitur pelengkap yang tidak kritis (seperti ulasan kelas dan portofolio mahasiswa) tidak boleh melumpuhkan operasional akademik inti universitas seperti proses login, pencatatan KRS, dan verifikasi kelulusan.

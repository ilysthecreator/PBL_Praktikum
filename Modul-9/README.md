# Modul 9 — Transactions & Concurrency

Dokumentasi implementasi transaksi database, pengendalian konkurensi (*concurrency control*), pencegahan *race condition* dan *deadlock*, serta penyelesaian tugas mandiri penetapan pembimbing tugas akhir mahasiswa.

---

## Daftar Isi
1. [Struktur Proyek](#struktur-proyek)
2. [Berita Acara Perbaikan 9 Kesalahan (D.1)](#d1-berita-acara-perbaikan-kesembilan-kesalahan)
3. [Implementasi Fitur Utama (Bagian B)](#implementasi-fitur-utama-bagian-b)
4. [Tugas Mandiri: Pembimbing Tugas Akhir (D.2)](#d2-pembimbing-tugas-akhir-entity-students)
5. [Hasil dan Bukti Uji Beban (D.3)](#d3-bukti-uji-beban)
6. [Analisis Singkat (D.4)](#d4-analisis-singkat)
7. [Panduan Menjalankan Aplikasi](#panduan-menjalankan-aplikasi)

---

## Struktur Proyek

```
Modul-9/
├── .env.example
├── .gitignore
├── go.mod
├── go.sum
├── main.go
├── reset.sql
├── README.md
├── app/
│   ├── model/
│   │   ├── auth.go
│   │   ├── course.go
│   │   ├── nilai.go
│   │   ├── student.go
│   │   ├── supervisor.go
│   │   └── user.go
│   ├── repository/
│   │   ├── course_repository.go
│   │   ├── dbtx.go
│   │   ├── nilai_repository.go
│   │   ├── role_repository.go
│   │   ├── student_repository.go
│   │   ├── supervisor_repository.go
│   │   ├── token_repository.go
│   │   ├── tx.go
│   │   └── user_repository.go
│   └── service/
│       ├── auth_service.go
│       ├── course_rules.go
│       ├── course_service.go
│       ├── nilai_service.go
│       ├── student_service.go
│       ├── supervisor_service.go
│       └── user_service.go
├── config/
│   ├── app.go
│   ├── env.go
│   └── logger.go
├── database/
│   └── postgre.go
├── helper/
│   ├── authz.go
│   ├── context.go
│   ├── errors.go
│   ├── jwt.go
│   ├── request.go
│   ├── response.go
│   ├── security.go
│   └── validator.go
├── middleware/
│   ├── auth.go
│   ├── authz.go
│   └── middleware.go
├── migrations/
│   ├── 001_create_student.sql
│   ├── 002_create_nilai.sql
│   ├── 003_auth.sql
│   ├── 003_rbac.sql
│   ├── 004_cursor_index.sql
│   ├── 004_student_permissions.sql
│   ├── 005_courses.sql
│   ├── 006_seed_loadtest.sql
│   └── 007_supervisors.sql
└── tools/
    └── loadtest/
        └── main.go
```

---

## D.1 Berita Acara Perbaikan Kesembilan Kesalahan

Berikut adalah tabel dokumentasi 9 kesalahan yang sengaja disisipkan pada kode panduan Bagian B beserta analisis dan solusinya:

| No | Berkas dan Baris | Jenis | Gejala | Reprodusibilitas | Akar Masalah | Perbaikan dan Bukti |
|:--:|:-----------------|:------|:-------|:-----------------|:-------------|:-------------------|
| 1 | `app/repository/course_repository.go`<br>`WithTx` method declaration | Ditolak compiler | `cannot use coursePostgresRepository literal as CourseRepository in return: does not implement CourseRepository (wrong type for method WithTx)` | 100% (setiap kali compile) | Interface `CourseRepository` mendefinisikan return type `CourseRepository`, namun implementasi struct mengembalikan pointer konkret `*coursePostgresRepository`. Dalam Go, signature method interface harus identik. | Mengubah return type method menjadi `CourseRepository`. Lolos compile pada butir C.1-C.4. |
| 2 | `app/repository/tx.go`<br>`WithTx` & `runOnce` parameter | Ditolak compiler | `cannot use func literal (type func(repository.DBTX) error) as type func(pgx.Tx) error in argument to m.runOnce` | 100% (setiap kali compile) | Komentar di modul menyatakan fungsi menerima `DBTX` agar layer service tidak perlu mengimpor `pgx`, namun parameter fungsi pada kode ditulis `func(tx pgx.Tx) error`. | Mengubah parameter menjadi `fn func(tx DBTX) error`. Layer service kini murni terisolasi dari package `pgx`. |
| 3 | `app/service/course_service.go`<br>Blok import package | Ditolak compiler | `undefined: strconv` pada pemanggilan `strconv.Itoa(...)` di method `Enroll` dan `translateCourseError` | 100% (setiap kali compile) | Fungsi `strconv.Itoa` digunakan untuk menyusun header Location dan pesan error limit SKS, namun package `"strconv"` lupa diimpor di blok `import`. | Menambahkan `"strconv"` ke dalam deklarasi import `course_service.go`. Berhasil build tanpa pesan error compiler. |
| 4 | `app/repository/tx.go`<br>`runOnce` method | Lolos compiler tetapi salah perilaku | Koneksi database dari pool habis (*pool exhaustion*), request berikutnya hang hingga `context deadline exceeded` | Muncul saat terjadi error di dalam transaksi (misal: kuota penuh / validasi gagal) | Baris `defer tx.Rollback(ctx)` tidak ditulis di dalam `runOnce`. Bila transaksi gagal atau panic sebelum `tx.Commit()`, koneksi database dibiarkan terbuka dalam status `idle in transaction` selamanya. | Menambahkan `defer func() { _ = tx.Rollback(ctx) }()` segera setelah `m.pool.Begin(ctx)`. Lolos uji C.5 (175 kegagalan beruntun tanpa kebocoran koneksi). |
| 5 | `app/repository/course_repository.go`<br>`IncrementEnrolled` query SQL | Lolos compiler tetapi salah perilaku | Pelanggaran constraint `courses_enrolled_within_quota` (HTTP 500) saat mahasiswa ke-6 mendaftar kelas berkuota 5 | 100% pada mahasiswa ke-6 di kelas kuota 5 | Klausa WHERE memakai `enrolled_count <= quota`. Saat `enrolled_count = 5` dan `quota = 5`, kondisi bernilai benar sehingga di-increment menjadi 6, melanggar CHECK constraint database. | Mengubah kondisi menjadi `enrolled_count < quota`. Mahasiswa ke-6 kini menerima 409 `COURSE_FULL` dan database tetap konsisten (butir C.1 & C.4). |
| 6 | `app/repository/course_repository.go`<br>`LockStudent` query SQL | Lolos compiler tetapi salah perilaku | Mahasiswa dapat mengambil lebih dari 24 SKS saat mengirim banyak request enroll serentak | Sering terjadi pada uji serentak (mode credits gagal, mahasiswa bisa menembus 27-36 SKS) | Query `SELECT id FROM users WHERE id = $1` tidak menyertakan klausa `FOR UPDATE`. Akibatnya, request paralel membaca jumlah SKS lama yang belum ter-commit (*check-then-act race condition*). | Menambahkan `FOR UPDATE` pada query: `SELECT id FROM users WHERE id = $1 FOR UPDATE`. Seluruh 10 mahasiswa tertahan pada <= 24 SKS (butir C.2). |
| 7 | `app/service/course_service.go`<br>`Swap` method | Lolos compiler tetapi salah perilaku | Request swap menggantung (*deadlock*) dan memicu timeout koneksi | Sering terjadi saat banyak request swap berjalan bersamaan | Baris pengurangan peserta memanggil `s.courses.DecrementEnrolled` (menggunakan pool koneksi terpisah) alih-alih `repo.DecrementEnrolled` (di dalam transaksi). Query luar tertahan oleh lock transaksi sendiri (pelanggaran aturan emas ke-4). | Mengganti pemanggilan menjadi `repo.DecrementEnrolled(ctx, req.FromCourseID)`. Seluruh query berjalan di dalam satu koneksi transaksi (butir C.3). |
| 8 | `app/repository/course_repository.go`<br>`DeleteEnrollment` method | Lolos compiler tetapi salah perilaku | Menghapus pendaftaran yang tidak ada tetap berhasil (HTTP 204), dan `enrolled_count` berkurang secara tidak sah | 100% saat memanggil DELETE untuk kelas yang belum pernah diambil | PostgreSQL tidak menganggap penghapusan 0 baris sebagai error. Karena `tag.RowsAffected()` tidak diperiksa, method mengembalikan `nil` dan kode melanjutkan pengurangan kuota. | Memeriksa `tag.RowsAffected() == 0`, jika ya kembalikan `ErrNotEnrolled`. Menghasilkan 409 `NOT_ENROLLED` dan kuota tidak berkurang (butir C.4). |
| 9 | `app/repository/course_repository.go`<br>`UpdateQuota` query SQL | Lolos compiler tetapi salah perilaku | Nilai `version` tidak bertambah setelah kuota diubah; request admin lain dengan version usang tidak tertolak | 100% saat admin memperbarui kuota | Query UPDATE hanya meng-update kolom `quota` tanpa menambah `version = version + 1`. Akibatnya deteksi konflik versi pada *optimistic locking* gagal berfungsi. | Menambahkan `version = version + 1` pada klausa SET: `UPDATE courses SET quota = $1, version = version + 1 ...`. Version response naik dan update basi ditolak dengan 409 `VERSION_CONFLICT` (butir C.4). |

---

## Implementasi Fitur Utama (Bagian B)

### 1. UPDATE Bersyarat yang Atomik
Digunakan pada `IncrementEnrolled` untuk menghindari pola rentan *check-then-act*:
```sql
UPDATE courses SET enrolled_count = enrolled_count + 1
WHERE id = $1 AND enrolled_count < quota
RETURNING id, code, name, credits, quota, enrolled_count, version, created_at;
```
Pemeriksaan dan penulisan terjadi dalam satu operasi database tanpa celah *race condition*.

### 2. Pessimistic Locking (`SELECT ... FOR UPDATE`)
Digunakan pada:
- `LockStudent`: Mengunci baris `users` milik mahasiswa sehingga request pendaftaran serentak oleh mahasiswa yang sama diproses bergantian, menjaga integritas batas 24 SKS.
- `LockCourses`: Mengunci beberapa baris `courses` dengan klausa `ORDER BY id FOR UPDATE` untuk mencegah *deadlock* saat terjadi pertukaran kelas berlawanan arah (`IF101` <-> `IF102`).

### 3. Optimistic Locking (Kolom `version`)
Digunakan pada pengubahan kuota kelas (`PATCH /courses/:id`):
```sql
UPDATE courses SET quota = $1, version = version + 1
WHERE id = $2 AND version = $3 AND $1 >= enrolled_count
RETURNING id, code, name, credits, quota, enrolled_count, version, created_at;
```
Bila data sudah disentuh oleh pihak lain sebelum simpan, update mengenai 0 baris dan memicu error `409 VERSION_CONFLICT`.

---

## D.2 Pembimbing Tugas Akhir (Entity Students)

Penerapan pola transaksi dan konkurensi pada domain penetapan dosen pembimbing tugas akhir:

1. **Skema Database (`migrations/007_supervisors.sql`)**:
   - Tabel `supervisors`: kolom `id`, `name`, `quota`, `supervised_count`, `version`, `created_at`, serta CHECK constraint `supervised_count <= quota`.
   - Tabel `supervisions`: kolom `student_id` (dengan constraint `UNIQUE`), `supervisor_id`, `created_at`.
   - Constraint `UNIQUE (student_id)` di level database menjamin aturan bahwa satu mahasiswa hanya boleh memiliki satu pembimbing.

2. **Endpoint `POST /api/v1/supervisors/:id/assign`**:
   - Dijaga permission `supervision:manage` (role admin).
   - Menjalankan transaksi:
     1. Kunci baris student (`LockStudent`) untuk mencegah mahasiswa ditetapkan ke dua pembimbing secara paralel.
     2. UPDATE bersyarat atomik kuota pembimbing (`IncrementSupervised` dengan syarat `supervised_count < quota`).
     3. Insert ke `supervisions`. Jika mahasiswa sudah memiliki pembimbing, UNIQUE constraint memicu rollback otomatis.

3. **Endpoint `POST /api/v1/supervisions/switch`**:
   - Pindah pembimbing tugas akhir secara atomik dan bebas *deadlock*.
   - Mengunci kedua pembimbing dalam urutan ID menaik (`LockSupervisors`).
   - Menghapus pembimbing lama, mengurangi hitungan lama, menambah hitungan baru, dan mencatat supervisi baru. Jika pembimbing baru penuh, transaksi gagal dan mahasiswa tetap dengan pembimbing lama.

4. **Endpoint `PATCH /api/v1/supervisors/:id`**:
   - Mengubah kuota pembimbing dengan *optimistic locking* (`version = version + 1`).

---

## D.3 Bukti Uji Beban

### 1. Uji Beban KRS — Mode Crowd (60 mahasiswa berebut kelas IF399 kuota 5)
```text
60 mahasiswa berebut kelas IF399 secara serentak...
hasil request:
  201 OK                    5
  409 COURSE_FULL           55
waktu total : 64ms
IF399  kuota=5  enrolled_count=5  baris enrollments=5  -> KONSISTEN
```

### 2. Uji Beban KRS — Mode Credits (10 mahasiswa, masing-masing 12 kelas 3 SKS, batas 24 SKS)
```text
10 mahasiswa, masing-masing mendaftar ke 12 kelas 3 SKS secara serentak (batas 24 SKS)...
hasil request:
  201 OK                    80
  409 CREDIT_LIMIT_EXCEEDED 40
waktu total : 98ms
seluruh 10 mahasiswa <= 24 SKS -> KONSISTEN
```

### 3. Uji Beban KRS — Mode Swap (20 pasang mahasiswa swap IF101 <-> IF102)
```text
20 pasang mahasiswa pindah kelas berlawanan arah secara serentak...
hasil request:
  200 OK                    40
waktu total : 72ms
deadlock terdeteksi PostgreSQL selama uji: 0
IF101  kuota=40  enrolled_count=20  baris enrollments=20  -> KONSISTEN
IF102  kuota=40  enrolled_count=20  baris enrollments=20  -> KONSISTEN
```

### 4. Uji Beban Tugas Mandiri — Mode Supervisor (30 penetapan serentak ke pembimbing kuota 3)
```text
30 penetapan serentak ke seorang pembimbing berkuota 3...
hasil request:
  201 OK                    3
  409 SUPERVISOR_FULL       27
waktu total : 48ms
Supervisor #1  kuota=3  supervised_count=3  baris supervisions=3  -> KONSISTEN
```

### 5. Percobaan Pembanding (Negative Test)
Sebagai pembuktian keandalan uji beban, syarat kuota pada UPDATE dihilangkan sementara:
- **Hasil**: 30 mahasiswa mendaftar serentak semuanya mendapatkan status `201 OK` (kuota jebol hingga 30).
- **Kesimpulan**: Uji beban terbukti valid dan efektif mendeteksi *race condition* jika mekanisme proteksi atomik ditiadakan.

---

## D.4 Analisis Singkat

1. **Kunci Tambahan pada Pergantian Pembimbing**:
   Kuota pembimbing memang dapat dijaga dengan UPDATE bersyarat atomik, namun aturan *"satu mahasiswa hanya boleh memiliki satu pembimbing"* saat terjadi pergantian pembimbing tetap memerlukan kunci tambahan pada baris data mahasiswa (`LockStudent` dengan `SELECT id FROM students WHERE id = $1 FOR UPDATE` pada `supervisor_repository.go:73`). Tanpa penguncian baris mahasiswa ini, dua request pergantian pembimbing untuk mahasiswa yang sama dapat berjalan paralel: keduanya sama-sama berhasil menghapus data lama dan masing-masing mendaftarkan ke pembimbing baru yang berbeda secara bersamaan, sehingga mahasiswa memiliki status ganda sebelum salah satu transaksi selesai.

2. **Pessimistic Locking vs Optimistic Locking pada Form Ubah Kuota**:
   Jika endpoint pengubahan kuota menggunakan *pessimistic locking* (`FOR UPDATE`), baris dosen pembimbing di tabel database akan terkunci selama 30 menit ketika form dibuka oleh admin. Akibatnya, seluruh mahasiswa yang mencoba mendaftar ke dosen tersebut akan terhenti (*blocking*) menunggu koneksi database hingga batas waktu *timeout*, melumpuhkan layanan aplikasi. Sebaliknya, dengan *optimistic locking* (kolom `version`), tidak ada *lock* yang ditahan di database selama admin membuka form. Ketika admin menekan tombol Simpan setelah 30 menit, sistem hanya memverifikasi apakah `version` masih sama dengan saat form dibuka; jika data telah berubah di database, sistem dengan aman menolak penyimpanan dengan `409 VERSION_CONFLICT` tanpa pernah menghambat request mahasiswa lain.

3. **Pengiriman Email di Dalam WithTx dan Potensi Kegagalannya**:
   Karena `WithTx` dirancang untuk mengulang eksekusi (*retry*) fungsi jika terjadi konflik serialisasi (`40001`) atau *deadlock* (`40P01`), menaruh pengiriman email di dalam fungsi `fn` akan menyebabkan email terkirim berulang kali kepada dosen pembimbing setiap kali percobaan transaksi diulang. Selain itu, bila transaksi pada akhirnya dibatalkan (*ROLLBACK*), aksi pengiriman email ke server SMTP eksternal tidak ikut dibatalkan, menyebabkan notifikasi palsu terkirim. Oleh karena itu, pengiriman email **wajib dilakukan di luar transaksi**, tepatnya setelah `s.tx.WithTx(...)` berhasil mengembalikan nilai `nil` (transaksi telah ter-*commit* sempurna di database).

---

## Panduan Menjalankan Aplikasi

### 1. Migrasi Database
Jalankan migrasi tabel kelas, akun beban uji, dan pembimbing:
```bash
psql -U postgres -d praktikum_backend -f migrations/005_courses.sql
psql -U postgres -d praktikum_backend -f migrations/006_seed_loadtest.sql
psql -U postgres -d praktikum_backend -f migrations/007_supervisors.sql
```

### 2. Menjalankan Server
```bash
go run main.go
```

### 3. Reset Database Sebelum Uji Beban
```bash
psql -U postgres -d praktikum_backend -f reset.sql
```

### 4. Menjalankan Uji Beban
Di terminal terpisah:
```bash
# Uji 60 mahasiswa berebut kelas IF399 kuota 5
go run ./tools/loadtest -mode crowd

# Uji batas 24 SKS
go run ./tools/loadtest -mode credits

# Uji swap kelas berlawanan arah (bebas deadlock)
go run ./tools/loadtest -mode swap

# Uji 30 penetapan serentak pembimbing kuota 3
go run ./tools/loadtest -mode supervisor
```

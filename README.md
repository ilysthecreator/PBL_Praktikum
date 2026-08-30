# API Students (Modul 3 Backend)

Aplikasi RESTful API manajemen data mahasiswa yang dibangun menggunakan **Go**, **Fiber v2**, dan **PostgreSQL (pgxpool)** dengan penerapan arsitektur *Repository Pattern*.

## 📋 Skema Tabel Basis Data

Tabel `students` dirancang dengan struktur berikut:

* `id`: `SERIAL` (Primary Key)
* `nim`: `VARCHAR(20)` (`NOT NULL`, unik melalui `UNIQUE INDEX`)
* `name`: `VARCHAR(255)` (`NOT NULL`)
* `grade`: `VARCHAR(5)` (`NOT NULL`, diindeks untuk optimalisasi *filtering*)
* `is_active`: `BOOLEAN` (Default `TRUE`)
* `created_at`: `TIMESTAMPTZ` (Default `NOW()`)

## ⚙️ Cara Menyiapkan Basis Data dari Nol

1. Pastikan layanan PostgreSQL Anda aktif.
2. Buat database baru bernama `praktikum_backend` melalui terminal/psql:
   ```bash
   psql -U postgres -c "CREATE DATABASE praktikum_backend;"
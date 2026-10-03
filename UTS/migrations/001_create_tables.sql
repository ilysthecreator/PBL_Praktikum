-- Tabel users
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE,
    password TEXT NOT NULL,
    role VARCHAR(20) NOT NULL CHECK (role IN ('admin', 'mahasiswa')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Tabel students (1-1 ke users)
CREATE TABLE IF NOT EXISTS students (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    nim VARCHAR(12) NOT NULL UNIQUE,
    nama VARCHAR(255) NOT NULL,
    prodi VARCHAR(255) NOT NULL,
    angkatan INT NOT NULL,
    ipk_terakhir NUMERIC(3,2) NOT NULL DEFAULT 0.00,
    deleted_at TIMESTAMPTZ DEFAULT NULL
);

-- Tabel courses
CREATE TABLE IF NOT EXISTS courses (
    id SERIAL PRIMARY KEY,
    kode_mk VARCHAR(50) NOT NULL UNIQUE,
    nama_mk VARCHAR(255) NOT NULL,
    sks INT NOT NULL CHECK (sks > 0),
    semester INT NOT NULL CHECK (semester > 0),
    kuota INT NOT NULL CHECK (kuota >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Tabel enrollments (1-N dari students & courses, unique student+course+tahun_akademik)
CREATE TABLE IF NOT EXISTS enrollments (
    id SERIAL PRIMARY KEY,
    student_id INT NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    course_id INT NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    tahun_akademik VARCHAR(50) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uniq_student_course_ta UNIQUE (student_id, course_id, tahun_akademik)
);

CREATE INDEX IF NOT EXISTS idx_students_nim ON students(nim);
CREATE INDEX IF NOT EXISTS idx_students_deleted_at ON students(deleted_at);
CREATE INDEX IF NOT EXISTS idx_courses_kode_mk ON courses(kode_mk);
CREATE INDEX IF NOT EXISTS idx_enrollments_student ON enrollments(student_id);
CREATE INDEX IF NOT EXISTS idx_enrollments_course ON enrollments(course_id);

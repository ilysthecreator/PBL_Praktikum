CREATE TABLE IF NOT EXISTS nilai (
    id_nilai SERIAL PRIMARY KEY,
    nama_mata_kuliah VARCHAR(255) NOT NULL,
    nilai NUMERIC(5, 2) NOT NULL,
    id_student INT NOT NULL,
    CONSTRAINT fk_nilai_student FOREIGN KEY (id_student) REFERENCES students(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_nilai_id_student ON nilai (id_student);

INSERT INTO nilai (nama_mata_kuliah, nilai, id_student)
SELECT 'Pemrograman Backend', 88.50, id FROM students WHERE nim = '434241099'
ON CONFLICT DO NOTHING;

INSERT INTO nilai (nama_mata_kuliah, nilai, id_student)
SELECT 'Basis Data Lanjut', 92.00, id FROM students WHERE nim = '434241099'
ON CONFLICT DO NOTHING;

INSERT INTO nilai (nama_mata_kuliah, nilai, id_student)
SELECT 'Struktur Data & Algoritma', 85.00, id FROM students WHERE nim = '434241100'
ON CONFLICT DO NOTHING;

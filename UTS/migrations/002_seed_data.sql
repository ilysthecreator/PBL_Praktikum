-- Seeder SIAKAD Mini

-- 1. Akun Admin (Password: admin12345)
INSERT INTO users (email, password, role) VALUES ('admin@siakad.ac.id', '$2a$10$wvVSDifSJwhM2NSZzkZxVO57On8lZYbHw7Q8DYytRqH0HvUdECt3O', 'admin') ON CONFLICT (email) DO NOTHING;

-- 2. 20 Akun User Mahasiswa & Data Students (Password: NIM masing-masing)
DO $$
DECLARE
    v_user_id INT;
BEGIN

    IF NOT EXISTS (SELECT 1 FROM users WHERE email = 'rina.putri@siakad.ac.id') THEN
        INSERT INTO users (email, password, role) VALUES ('rina.putri@siakad.ac.id', '$2a$10$jvRlriEEY82Bt/.jKt9at..0kGuTLAZgLWbVOVi/0eJ0XzNalLzZq', 'mahasiswa') RETURNING id INTO v_user_id;
        INSERT INTO students (user_id, nim, nama, prodi, angkatan, ipk_terakhir) VALUES (v_user_id, '187221000001', 'Rina Putri', 'Sistem Informasi', 2022, 3.45);
    END IF;

    IF NOT EXISTS (SELECT 1 FROM users WHERE email = 'budi.santoso@siakad.ac.id') THEN
        INSERT INTO users (email, password, role) VALUES ('budi.santoso@siakad.ac.id', '$2a$10$KPR0bPEQ/bRGt9oc6MWClePixPHQObWWbGHBP4RUDy8nYcJptpNSq', 'mahasiswa') RETURNING id INTO v_user_id;
        INSERT INTO students (user_id, nim, nama, prodi, angkatan, ipk_terakhir) VALUES (v_user_id, '187221000002', 'Budi Santoso', 'Teknik Informatika', 2022, 3.80);
    END IF;

    IF NOT EXISTS (SELECT 1 FROM users WHERE email = 'siti.aminah@siakad.ac.id') THEN
        INSERT INTO users (email, password, role) VALUES ('siti.aminah@siakad.ac.id', '$2a$10$ojplbo2dLX7sV200pMe4eusi/KcUy6TckP9TyRnF5LHnWV/tH1h32', 'mahasiswa') RETURNING id INTO v_user_id;
        INSERT INTO students (user_id, nim, nama, prodi, angkatan, ipk_terakhir) VALUES (v_user_id, '187221000003', 'Siti Aminah', 'Sistem Informasi', 2023, 2.85);
    END IF;

    IF NOT EXISTS (SELECT 1 FROM users WHERE email = 'ahmad.fauzi@siakad.ac.id') THEN
        INSERT INTO users (email, password, role) VALUES ('ahmad.fauzi@siakad.ac.id', '$2a$10$2mqA5quTOSl9.iGYK84sjOaVOoT8PoAX2HHEI5sXxDmh.il2WP04u', 'mahasiswa') RETURNING id INTO v_user_id;
        INSERT INTO students (user_id, nim, nama, prodi, angkatan, ipk_terakhir) VALUES (v_user_id, '187221000004', 'Ahmad Fauzi', 'Teknik Informatika', 2023, 2.40);
    END IF;

    IF NOT EXISTS (SELECT 1 FROM users WHERE email = 'dewi.lestari@siakad.ac.id') THEN
        INSERT INTO users (email, password, role) VALUES ('dewi.lestari@siakad.ac.id', '$2a$10$tR2KfH62PwDFOBKAVMQwduzGIjl.5h9/SSj1lux.9hy24hj9nK0OO', 'mahasiswa') RETURNING id INTO v_user_id;
        INSERT INTO students (user_id, nim, nama, prodi, angkatan, ipk_terakhir) VALUES (v_user_id, '187221000005', 'Dewi Lestari', 'Sains Data', 2024, 3.60);
    END IF;

    IF NOT EXISTS (SELECT 1 FROM users WHERE email = 'eko.prasetyo@siakad.ac.id') THEN
        INSERT INTO users (email, password, role) VALUES ('eko.prasetyo@siakad.ac.id', '$2a$10$VBEBzsrpHPWCajhoxnR7kOuykeqXBHYzSuyBqUUJ3Y6Rae1RhvCLm', 'mahasiswa') RETURNING id INTO v_user_id;
        INSERT INTO students (user_id, nim, nama, prodi, angkatan, ipk_terakhir) VALUES (v_user_id, '187221000006', 'Eko Prasetyo', 'Teknik Informatika', 2022, 3.10);
    END IF;

    IF NOT EXISTS (SELECT 1 FROM users WHERE email = 'fajar.hidayat@siakad.ac.id') THEN
        INSERT INTO users (email, password, role) VALUES ('fajar.hidayat@siakad.ac.id', '$2a$10$oS/AgLoW/MyGl9qbMcOn1.rkcin/R/9YIIkUg4Lql97CIahC7oHGu', 'mahasiswa') RETURNING id INTO v_user_id;
        INSERT INTO students (user_id, nim, nama, prodi, angkatan, ipk_terakhir) VALUES (v_user_id, '187221000007', 'Fajar Hidayat', 'Sistem Informasi', 2023, 2.95);
    END IF;

    IF NOT EXISTS (SELECT 1 FROM users WHERE email = 'gita.gutawa@siakad.ac.id') THEN
        INSERT INTO users (email, password, role) VALUES ('gita.gutawa@siakad.ac.id', '$2a$10$3CriXYAuRCvBJLd7zpMdAeiUpZDoRV/0AUJ2hhQYXu/hSI17MxWsG', 'mahasiswa') RETURNING id INTO v_user_id;
        INSERT INTO students (user_id, nim, nama, prodi, angkatan, ipk_terakhir) VALUES (v_user_id, '187221000008', 'Gita Gutawa', 'Sains Data', 2024, 3.25);
    END IF;

    IF NOT EXISTS (SELECT 1 FROM users WHERE email = 'hadi.kurniawan@siakad.ac.id') THEN
        INSERT INTO users (email, password, role) VALUES ('hadi.kurniawan@siakad.ac.id', '$2a$10$90GsuhD8Sb2YWZLMEaWyW.aXBDB0ReIEuscz4N2B6IwdNEI10Gia.', 'mahasiswa') RETURNING id INTO v_user_id;
        INSERT INTO students (user_id, nim, nama, prodi, angkatan, ipk_terakhir) VALUES (v_user_id, '187221000009', 'Hadi Kurniawan', 'Teknik Informatika', 2021, 2.30);
    END IF;

    IF NOT EXISTS (SELECT 1 FROM users WHERE email = 'indah.permata@siakad.ac.id') THEN
        INSERT INTO users (email, password, role) VALUES ('indah.permata@siakad.ac.id', '$2a$10$yvyAe6p1IF.dPjVv08AaeOq5bBHozQlesMNI6iSdcAcrXi78zbaiS', 'mahasiswa') RETURNING id INTO v_user_id;
        INSERT INTO students (user_id, nim, nama, prodi, angkatan, ipk_terakhir) VALUES (v_user_id, '187221000010', 'Indah Permata', 'Sistem Informasi', 2022, 3.90);
    END IF;

    IF NOT EXISTS (SELECT 1 FROM users WHERE email = 'joko.widodo@siakad.ac.id') THEN
        INSERT INTO users (email, password, role) VALUES ('joko.widodo@siakad.ac.id', '$2a$10$JCI3aSVBGxQzj00fcGuD4OXaopUiWeYW1u6ySRJPuKLkMgCCCy1r6', 'mahasiswa') RETURNING id INTO v_user_id;
        INSERT INTO students (user_id, nim, nama, prodi, angkatan, ipk_terakhir) VALUES (v_user_id, '187221000011', 'Joko Widodo', 'Teknik Informatika', 2022, 2.75);
    END IF;

    IF NOT EXISTS (SELECT 1 FROM users WHERE email = 'kartika.sari@siakad.ac.id') THEN
        INSERT INTO users (email, password, role) VALUES ('kartika.sari@siakad.ac.id', '$2a$10$lTPeU/7uK7PsHKonlXuoN.ZSjZS6R7z.uQy.IO0QAUOCmgEyD77Fm', 'mahasiswa') RETURNING id INTO v_user_id;
        INSERT INTO students (user_id, nim, nama, prodi, angkatan, ipk_terakhir) VALUES (v_user_id, '187221000012', 'Kartika Sari', 'Sains Data', 2023, 3.55);
    END IF;

    IF NOT EXISTS (SELECT 1 FROM users WHERE email = 'lukman.hakim@siakad.ac.id') THEN
        INSERT INTO users (email, password, role) VALUES ('lukman.hakim@siakad.ac.id', '$2a$10$iMgQ6bTFFsuZih9M657uiehy/Rg7Sw3YgY5rkOyWIYYU.S2aWqahq', 'mahasiswa') RETURNING id INTO v_user_id;
        INSERT INTO students (user_id, nim, nama, prodi, angkatan, ipk_terakhir) VALUES (v_user_id, '187221000013', 'Lukman Hakim', 'Sistem Informasi', 2023, 3.15);
    END IF;

    IF NOT EXISTS (SELECT 1 FROM users WHERE email = 'mega.suryani@siakad.ac.id') THEN
        INSERT INTO users (email, password, role) VALUES ('mega.suryani@siakad.ac.id', '$2a$10$PT0tVcdRWKUSUaVLQd6AZ.sPyBIwFEiCUujH/.7ufvxgSgWOkGU.m', 'mahasiswa') RETURNING id INTO v_user_id;
        INSERT INTO students (user_id, nim, nama, prodi, angkatan, ipk_terakhir) VALUES (v_user_id, '187221000014', 'Mega Suryani', 'Teknik Informatika', 2024, 2.65);
    END IF;

    IF NOT EXISTS (SELECT 1 FROM users WHERE email = 'nanda.pratama@siakad.ac.id') THEN
        INSERT INTO users (email, password, role) VALUES ('nanda.pratama@siakad.ac.id', '$2a$10$1XqBBEYjjmm36Iq7fGWCbeVNjh0JVwVt4RpdFefqBi6h/7uTRQ.T6', 'mahasiswa') RETURNING id INTO v_user_id;
        INSERT INTO students (user_id, nim, nama, prodi, angkatan, ipk_terakhir) VALUES (v_user_id, '187221000015', 'Nanda Pratama', 'Sains Data', 2022, 3.70);
    END IF;

    IF NOT EXISTS (SELECT 1 FROM users WHERE email = 'oki.setiawan@siakad.ac.id') THEN
        INSERT INTO users (email, password, role) VALUES ('oki.setiawan@siakad.ac.id', '$2a$10$EF2gzWedl5OCbSeAOs0cTe0AM8SkqNMYyNKonSgSjUxcOZzciZoVG', 'mahasiswa') RETURNING id INTO v_user_id;
        INSERT INTO students (user_id, nim, nama, prodi, angkatan, ipk_terakhir) VALUES (v_user_id, '187221000016', 'Oki Setiawan', 'Teknik Informatika', 2021, 2.10);
    END IF;

    IF NOT EXISTS (SELECT 1 FROM users WHERE email = 'putri.rahayu@siakad.ac.id') THEN
        INSERT INTO users (email, password, role) VALUES ('putri.rahayu@siakad.ac.id', '$2a$10$heP1XsyJaEuPG0spDxi3xufaDFKe4eieiP7vsJX5FAH9QFQzxuHQy', 'mahasiswa') RETURNING id INTO v_user_id;
        INSERT INTO students (user_id, nim, nama, prodi, angkatan, ipk_terakhir) VALUES (v_user_id, '187221000017', 'Putri Rahayu', 'Sistem Informasi', 2023, 3.30);
    END IF;

    IF NOT EXISTS (SELECT 1 FROM users WHERE email = 'qori.maulana@siakad.ac.id') THEN
        INSERT INTO users (email, password, role) VALUES ('qori.maulana@siakad.ac.id', '$2a$10$fcSWcX36waJkuGEBP1bfbefV.3IaYrAbXSJA5aQ73FFBHRtf7YmJy', 'mahasiswa') RETURNING id INTO v_user_id;
        INSERT INTO students (user_id, nim, nama, prodi, angkatan, ipk_terakhir) VALUES (v_user_id, '187221000018', 'Qori Maulana', 'Teknik Informatika', 2024, 3.00);
    END IF;

    IF NOT EXISTS (SELECT 1 FROM users WHERE email = 'rian.saputra@siakad.ac.id') THEN
        INSERT INTO users (email, password, role) VALUES ('rian.saputra@siakad.ac.id', '$2a$10$9ePCmHjnDIhFdvxvl.Rw2eXpLiF016wStgXjz0R0cnolK/Dv5piO6', 'mahasiswa') RETURNING id INTO v_user_id;
        INSERT INTO students (user_id, nim, nama, prodi, angkatan, ipk_terakhir) VALUES (v_user_id, '187221000019', 'Rian Saputra', 'Sains Data', 2022, 2.90);
    END IF;

    IF NOT EXISTS (SELECT 1 FROM users WHERE email = 'tiara.andini@siakad.ac.id') THEN
        INSERT INTO users (email, password, role) VALUES ('tiara.andini@siakad.ac.id', '$2a$10$9eCUDtSX1bsoumI05cSS3ekjpSUvFEi.7wFcYhEGWz01SK44KTS9O', 'mahasiswa') RETURNING id INTO v_user_id;
        INSERT INTO students (user_id, nim, nama, prodi, angkatan, ipk_terakhir) VALUES (v_user_id, '187221000020', 'Tiara Andini', 'Sistem Informasi', 2023, 3.85);
    END IF;
END $$;

-- 3. 10 Data Mata Kuliah
INSERT INTO courses (kode_mk, nama_mk, sks, semester, kuota) VALUES ('IF101', 'Algoritma dan Pemrograman', 3, 1, 30) ON CONFLICT (kode_mk) DO NOTHING;
INSERT INTO courses (kode_mk, nama_mk, sks, semester, kuota) VALUES ('IF102', 'Struktur Data', 3, 2, 30) ON CONFLICT (kode_mk) DO NOTHING;
INSERT INTO courses (kode_mk, nama_mk, sks, semester, kuota) VALUES ('IF103', 'Basis Data', 3, 3, 30) ON CONFLICT (kode_mk) DO NOTHING;
INSERT INTO courses (kode_mk, nama_mk, sks, semester, kuota) VALUES ('IF104', 'Pemrograman Web', 3, 4, 25) ON CONFLICT (kode_mk) DO NOTHING;
INSERT INTO courses (kode_mk, nama_mk, sks, semester, kuota) VALUES ('IF105', 'Pemrograman Backend', 3, 5, 25) ON CONFLICT (kode_mk) DO NOTHING;
INSERT INTO courses (kode_mk, nama_mk, sks, semester, kuota) VALUES ('IF106', 'Rekayasa Perangkat Lunak', 3, 5, 30) ON CONFLICT (kode_mk) DO NOTHING;
INSERT INTO courses (kode_mk, nama_mk, sks, semester, kuota) VALUES ('IF107', 'Jaringan Komputer', 3, 4, 30) ON CONFLICT (kode_mk) DO NOTHING;
INSERT INTO courses (kode_mk, nama_mk, sks, semester, kuota) VALUES ('IF108', 'Kecerdasan Buatan', 3, 6, 20) ON CONFLICT (kode_mk) DO NOTHING;
INSERT INTO courses (kode_mk, nama_mk, sks, semester, kuota) VALUES ('IF109', 'Keamanan Informasi', 3, 6, 25) ON CONFLICT (kode_mk) DO NOTHING;
INSERT INTO courses (kode_mk, nama_mk, sks, semester, kuota) VALUES ('IF110', 'Skripsi', 6, 8, 50) ON CONFLICT (kode_mk) DO NOTHING;

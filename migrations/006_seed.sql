CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- Admin user
INSERT INTO users (email, password, role)
VALUES ('admin@siakad.test', crypt('Admin12345', gen_salt('bf', 12)), 'admin')
ON CONFLICT (email) DO NOTHING;

-- Mahasiswa users
INSERT INTO users (email, password, role) VALUES
('mahasiswa01@siakad.test', crypt('187221000001', gen_salt('bf', 12)), 'mahasiswa'),
('mahasiswa02@siakad.test', crypt('187221000002', gen_salt('bf', 12)), 'mahasiswa'),
('mahasiswa03@siakad.test', crypt('187221000003', gen_salt('bf', 12)), 'mahasiswa'),
('mahasiswa04@siakad.test', crypt('187221000004', gen_salt('bf', 12)), 'mahasiswa'),
('mahasiswa05@siakad.test', crypt('187221000005', gen_salt('bf', 12)), 'mahasiswa'),
('mahasiswa06@siakad.test', crypt('187221000006', gen_salt('bf', 12)), 'mahasiswa'),
('mahasiswa07@siakad.test', crypt('187221000007', gen_salt('bf', 12)), 'mahasiswa'),
('mahasiswa08@siakad.test', crypt('187221000008', gen_salt('bf', 12)), 'mahasiswa'),
('mahasiswa09@siakad.test', crypt('187221000009', gen_salt('bf', 12)), 'mahasiswa'),
('mahasiswa10@siakad.test', crypt('187221000010', gen_salt('bf', 12)), 'mahasiswa'),
('mahasiswa11@siakad.test', crypt('187221000011', gen_salt('bf', 12)), 'mahasiswa'),
('mahasiswa12@siakad.test', crypt('187221000012', gen_salt('bf', 12)), 'mahasiswa'),
('mahasiswa13@siakad.test', crypt('187221000013', gen_salt('bf', 12)), 'mahasiswa'),
('mahasiswa14@siakad.test', crypt('187221000014', gen_salt('bf', 12)), 'mahasiswa'),
('mahasiswa15@siakad.test', crypt('187221000015', gen_salt('bf', 12)), 'mahasiswa'),
('mahasiswa16@siakad.test', crypt('187221000016', gen_salt('bf', 12)), 'mahasiswa'),
('mahasiswa17@siakad.test', crypt('187221000017', gen_salt('bf', 12)), 'mahasiswa'),
('mahasiswa18@siakad.test', crypt('187221000018', gen_salt('bf', 12)), 'mahasiswa'),
('mahasiswa19@siakad.test', crypt('187221000019', gen_salt('bf', 12)), 'mahasiswa'),
('mahasiswa20@siakad.test', crypt('187221000020', gen_salt('bf', 12)), 'mahasiswa')
ON CONFLICT (email) DO NOTHING;

-- Mahasiswa student records
INSERT INTO students (user_id, nim, nama, prodi, angkatan, ipk_terakhir)
SELECT u.id, s.nim, s.nama, s.prodi, s.angkatan, s.ipk_terakhir
FROM (VALUES
    ('mahasiswa01@siakad.test', '187221000001', 'Ahmad Fauzi', 'Teknik Informatika', 2021, 3.85::numeric(3,2)),
    ('mahasiswa02@siakad.test', '187221000002', 'Budi Santoso', 'Teknik Informatika', 2021, 3.70::numeric(3,2)),
    ('mahasiswa03@siakad.test', '187221000003', 'Citra Dewi', 'Sistem Informasi', 2022, 3.55::numeric(3,2)),
    ('mahasiswa04@siakad.test', '187221000004', 'Dimas Pratama', 'Sistem Informasi', 2022, 3.40::numeric(3,2)),
    ('mahasiswa05@siakad.test', '187221000005', 'Eka Putri', 'Teknik Elektro', 2023, 3.25::numeric(3,2)),
    ('mahasiswa06@siakad.test', '187221000006', 'Fajar Hidayat', 'Teknik Elektro', 2023, 3.10::numeric(3,2)),
    ('mahasiswa07@siakad.test', '187221000007', 'Gita Lestari', 'Teknik Komputer', 2024, 3.00::numeric(3,2)),
    ('mahasiswa08@siakad.test', '187221000008', 'Hadi Nugroho', 'Teknik Informatika', 2021, 2.95::numeric(3,2)),
    ('mahasiswa09@siakad.test', '187221000009', 'Indah Permata', 'Teknik Informatika', 2022, 2.85::numeric(3,2)),
    ('mahasiswa10@siakad.test', '187221000010', 'Joko Widodo', 'Sistem Informasi', 2022, 2.75::numeric(3,2)),
    ('mahasiswa11@siakad.test', '187221000011', 'Kartika Sari', 'Sistem Informasi', 2023, 2.65::numeric(3,2)),
    ('mahasiswa12@siakad.test', '187221000012', 'Lukman Hakim', 'Teknik Elektro', 2023, 2.55::numeric(3,2)),
    ('mahasiswa13@siakad.test', '187221000013', 'Maya Anggraini', 'Teknik Komputer', 2024, 2.50::numeric(3,2)),
    ('mahasiswa14@siakad.test', '187221000014', 'Naufal Rizqi', 'Teknik Komputer', 2024, 2.45::numeric(3,2)),
    ('mahasiswa15@siakad.test', '187221000015', 'Olivia Rahma', 'Teknik Informatika', 2021, 2.35::numeric(3,2)),
    ('mahasiswa16@siakad.test', '187221000016', 'Panji Kusuma', 'Teknik Informatika', 2022, 2.20::numeric(3,2)),
    ('mahasiswa17@siakad.test', '187221000017', 'Qori Amalia', 'Sistem Informasi', 2022, 2.10::numeric(3,2)),
    ('mahasiswa18@siakad.test', '187221000018', 'Rian Saputra', 'Teknik Elektro', 2023, 1.95::numeric(3,2)),
    ('mahasiswa19@siakad.test', '187221000019', 'Siti Nurhaliza', 'Teknik Elektro', 2023, 1.80::numeric(3,2)),
    ('mahasiswa20@siakad.test', '187221000020', 'Taufik Hidayat', 'Teknik Komputer', 2024, 1.50::numeric(3,2))
) AS s(email, nim, nama, prodi, angkatan, ipk_terakhir)
JOIN users u ON u.email = s.email
ON CONFLICT (nim) DO NOTHING;

-- Courses
INSERT INTO courses (kode_mk, nama_mk, sks, semester, kuota) VALUES
('CS101', 'Dasar Pemrograman', 3, 1, 30),
('CS102', 'Struktur Data dan Algoritma', 4, 2, 25),
('CS201', 'Basis Data', 3, 3, 2),
('CS202', 'Pemrograman Berorientasi Objek', 3, 3, 1),
('CS301', 'Jaringan Komputer', 3, 4, 20),
('CS302', 'Rekayasa Perangkat Lunak', 3, 4, 25),
('CS401', 'Kecerdasan Buatan', 3, 5, 20),
('CS402', 'Keamanan Informasi', 3, 5, 1),
('CS501', 'Pemrograman Web Lanjut', 3, 6, 15),
('UNV101', 'Pendidikan Pancasila', 2, 1, 40)
ON CONFLICT (kode_mk) DO NOTHING;

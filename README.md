# SIAKAD Mini — RESTful API (UTS Praktikum Backend Engineering)

> Sistem Informasi Akademik Mini (**SIAKAD Mini**) untuk pengelolaan mahasiswa, mata kuliah, dan kartu rencana studi (KRS/Enrollment) berbasis **Clean Architecture**.

---

## 🛠️ Tech Stack & Architecture

- **Bahasa Pemrograman**: Go 1.26+
- **Web Framework**: [Fiber v2](https://github.com/gofiber/fiber)
- **Database**: PostgreSQL 14+ via [pgx/v5](https://github.com/jackc/pgx)
- **Arsitektur**: Clean Architecture (Onion / Hexagonal layers)
- **Keamanan**: JWT (HS256), bcrypt (cost 12), Helmet, Rate Limiter (Fiber Limiter), CORS, Timing-Safe Password Compare
- **Logging**: Structured Logging (`log/slog`) dengan rotasi berkas otomatis ([lumberjack.v2](https://github.com/natefinch/lumberjack))

```
+-------------------------------------------------------------+
|  FRAMEWORKS & DRIVERS (Outermost)                           |
|  config/ · database/ · middleware/ · route/ · helper/       |
|  Fiber v2 · pgx · lumberjack · godotenv · jwt               |
+-------------------------------------------------------------+
|  INTERFACE ADAPTERS                                         |
|  middleware/   -- Konversi HTTP <-> Domain                  |
|  route/        -- Mapping HTTP endpoint ke handler          |
|  helper/       -- Adapter JWT, errors, format respons       |
+-------------------------------------------------------------+
|  USE CASES (app/service/)                                   |
|  AuthService       -- login / me                            |
|  StudentService    -- CRUD + soft delete + ownership authz  |
|  CourseService     -- list mata kuliah dengan kuota         |
|  EnrollmentService -- KRS transaksional (SELECT FOR UPDATE) |
|  *_rules.go        -- Business rules murni (100% testable)  |
+-------------------------------------------------------------+
|  ENTITIES (app/model/) [Innermost - No External Deps]       |
|  User · Student · Course · Enrollment · DTOs                |
+-------------------------------------------------------------+
```

---

## 📂 Struktur Direktori

```
siakad-mini/
├── main.go                         # Bootstrap & graceful shutdown
├── go.mod / go.sum                 # Dependensi Go module
├── .env.example                    # Template konfigurasi environment
├── migrations/                     # Berkas SQL DDL & seeder
│   ├── 001_create_users.sql
│   ├── 002_create_students.sql
│   ├── 003_create_courses.sql
│   ├── 004_create_enrollments.sql  # UNIQUE (student_id, course_id, tahun_akademik)
│   ├── 005_indexes.sql             # Indeks filter, pencarian, dan foreign key
│   └── 006_seed.sql                # Seeder 1 admin, 20 mahasiswa, 10 mata kuliah
├── app/
│   ├── model/                      # Domain entities & DTO
│   ├── repository/                 # Data access interfaces & implementasi pgx
│   └── service/                    # Business logic & pure testable rules
├── config/                         # App bootstrap, env parser, structured logger
├── database/                       # PostgreSQL connection pool (pgxpool)
├── helper/                         # JWT, authz, response, validator, security
├── middleware/                     # Global middleware, auth guard, rate limiter, role guard
├── route/                          # Registrasi 10 endpoint API
└── logs/                           # Berkas log aplikasi (app.log terotasi)
```

---

## 🚀 Panduan Instalasi & Menjalankan

### 1. Prasyarat
- Go 1.26 atau lebih baru
- PostgreSQL 14 atau lebih baru

### 2. Kloning & Pengaturan Environment
Salin template berkas konfigurasi:
```bash
cp .env.example .env
```
Sesuaikan konfigurasi koneksi database di berkas `.env`:
```ini
APP_PORT=3000
APP_NAME=siakad-mini
APP_ENV=development
LOG_LEVEL=info
ALLOWED_ORIGINS=*

DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=
DB_NAME=siakad_mini
DB_SSLMODE=disable
DB_MAX_CONNS=10

JWT_SECRET=super-secret-key-change-in-production
JWT_ISSUER=siakad-mini
JWT_ACCESS_TTL=1h
```

### 3. Buat Database & Jalankan Migrasi
Buat database PostgreSQL:
```bash
createdb -U postgres siakad_mini
```

Jalankan semua berkas migrasi berurutan (termasuk seeder):
**Linux / macOS / Git Bash:**
```bash
for f in migrations/*.sql; do psql -U postgres -d siakad_mini -f "$f"; done
```
**Windows PowerShell:**
```powershell
Get-ChildItem migrations\*.sql | Sort-Object Name | ForEach-Object { psql -U postgres -d siakad_mini -f $_.FullName }
```

### 4. Jalankan Aplikasi
```bash
go run .
```
Aplikasi akan berjalan pada port `3000` (atau sesuai konfigurasi `APP_PORT`).

---

## 👥 Akun Bawaan (Seed Accounts)

Data awal telah dimasukkan via `006_seed.sql`:

| Role | Email | Password | Keterangan |
|---|---|---|---|
| **Admin** | `admin@siakad.test` | `Admin12345` | Akses penuh CRUD mahasiswa, lihat mata kuliah |
| **Mahasiswa** | `mahasiswa01@siakad.test` | `187221000001` | Password = NIM masing-masing (12 digit) |
| **Mahasiswa** | `mahasiswa02@siakad.test` | `187221000002` | Password = NIM masing-masing |
| ... | ... | ... | Tersedia total 20 mahasiswa (`mahasiswa01` s.d `mahasiswa20`) |

---

## 📜 Aturan Bisnis (Business Rules) & Asumsi

1. **Batas SKS Berdasarkan IPK Terakhir**:
   - `IPK >= 3.00`: Maksimal **24 SKS**
   - `2.50 <= IPK < 3.00`: Maksimal **21 SKS**
   - `IPK < 2.50`: Maksimal **18 SKS**
   *(Divalidasi secara atomik di dalam transaksi database menggunakan `SELECT ... FOR UPDATE`)*
2. **Pencegahan Duplikasi KRS**: Mahasiswa tidak dapat mengambil mata kuliah yang sama lebih dari sekali pada tahun akademik yang sama (`UNIQUE (student_id, course_id, tahun_akademik)` &rarr; `409 Conflict`).
3. **Pemeriksaan Kuota**: Mahasiswa tidak dapat mendaftar jika kuota mata kuliah sudah penuh (`terisi >= kuota` &rarr; `422 Unprocessable Entity`).
4. **Otorisasi Kepemilikan (Ownership Check)**:
   - Mahasiswa hanya dapat melihat profil/KRS miliknya sendiri (`GET /students/:id`).
   - Mahasiswa hanya dapat menghapus/membatalkan KRS miliknya sendiri (`DELETE /enrollments/:id`).
   - `student_id` pendaftaran selalu diambil dari JWT pengguna yang login, tidak pernah dari request body.
5. **Soft Delete Mahasiswa**: Penghapusan mahasiswa mengisi kolom `deleted_at`. Mahasiswa yang telah di-soft-delete tidak akan muncul dalam list, detailnya menghasilkan `404`, dan akunnya diblokir dari login (`401`).
6. **Rate Limiting Login**: Maksimal 5 kali kegagalan login per menit per IP (`429 Too Many Requests`). Percobaan berhasil tidak memotong kuota limiter.
7. **Uniform Response Envelope**: Seluruh respons API dibungkus format seragam:
   ```json
   {
     "success": true,
     "message": "pesan",
     "data": {},
     "meta": {}
   }
   ```
   Atau untuk error:
   ```json
   {
     "success": false,
     "message": "pesan error",
     "errors": {}
   }
   ```

---

## 📡 Daftar 10 Endpoint API & Contoh cURL

Base URL: `http://localhost:3000/api/v1`

### 1. POST `/auth/login` (Public)
Login pengguna dan memperoleh access token JWT.
```bash
curl -X POST http://localhost:3000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@siakad.test","password":"Admin12345"}'
```

### 2. GET `/auth/me` (Authenticated)
Melihat profil pengguna yang sedang login.
```bash
curl -X GET http://localhost:3000/api/v1/auth/me \
  -H "Authorization: Bearer <TOKEN>"
```

### 3. GET `/students` (Admin Only)
Melihat daftar mahasiswa dengan paginasi, filter prodi/angkatan, pencarian, dan pengurutan (`nama`, `-ipk_terakhir`).
```bash
curl -X GET "http://localhost:3000/api/v1/students?page=1&per_page=10&search=Ahmad&sort=-ipk_terakhir" \
  -H "Authorization: Bearer <ADMIN_TOKEN>"
```

### 4. POST `/students` (Admin Only)
Menambahkan mahasiswa baru sekaligus membuat akun user dalam 1 transaksi (password default = NIM).
```bash
curl -X POST http://localhost:3000/api/v1/students \
  -H "Authorization: Bearer <ADMIN_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
    "nim": "187221000099",
    "nama": "Farhan Kurniawan",
    "email": "farhan@siakad.test",
    "prodi": "Teknik Informatika",
    "angkatan": 2024,
    "ipk_terakhir": 3.75
  }'
```

### 5. GET `/students/:id` (Admin / Mahasiswa Pemilik)
Melihat detail mahasiswa, riwayat KRS, total SKS yang diambil, dan batas maksimal SKS.
```bash
curl -X GET http://localhost:3000/api/v1/students/1 \
  -H "Authorization: Bearer <TOKEN>"
```

### 6. PUT `/students/:id` (Admin Only)
Memperbarui data mahasiswa (NIM bersifat immutable).
```bash
curl -X PUT http://localhost:3000/api/v1/students/1 \
  -H "Authorization: Bearer <ADMIN_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
    "nama": "Ahmad Fauzi Mubarak",
    "prodi": "Teknik Informatika",
    "angkatan": 2021,
    "ipk_terakhir": 3.90
  }'
```

### 7. DELETE `/students/:id` (Admin Only)
Menghapus mahasiswa secara soft delete (menyetel `deleted_at`).
```bash
curl -X DELETE http://localhost:3000/api/v1/students/1 \
  -H "Authorization: Bearer <ADMIN_TOKEN>"
```

### 8. GET `/courses` (Authenticated)
Melihat daftar mata kuliah beserta kalkulasi kuota real-time (`terisi` dan `sisa_kuota`).
```bash
curl -X GET "http://localhost:3000/api/v1/courses?semester=1&available=true" \
  -H "Authorization: Bearer <TOKEN>"
```

### 9. POST `/enrollments` (Mahasiswa Only)
Mendaftar mata kuliah (KRS) dengan verifikasi kuota, tahun akademik, dan batas SKS.
```bash
curl -X POST http://localhost:3000/api/v1/enrollments \
  -H "Authorization: Bearer <MAHASISWA_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
    "course_id": 1,
    "tahun_akademik": "2026/2027-Ganjil"
  }'
```

### 10. DELETE `/enrollments/:id` (Mahasiswa Pemilik)
Membatalkan pendaftaran mata kuliah (kuota otomatis bertambah kembali).
```bash
curl -X DELETE http://localhost:3000/api/v1/enrollments/1 \
  -H "Authorization: Bearer <MAHASISWA_TOKEN>"
```

---

## 🧪 Pengujian (Testing)

Jalankan seluruh rangkaian unit test otomatis:
```bash
go test -v ./...
```
Unit test mencakup:
- Aturan validasi login (`app/service/auth_rules_test.go`)
- Aturan otorisasi akses mahasiswa & pagination (`app/service/student_rules_test.go`)
- Aturan batas SKS, format tahun akademik, dan kuota KRS (`app/service/enrollment_rules_test.go`)

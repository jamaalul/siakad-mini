# Project Plan — `siakad-mini` (UTS PBE)

> Go REST API for **SIAKAD Mini** (students, courses, and KRS/enrollment management), built with **Clean Architecture** principles.  
> Runtime: **Go 1.26** · Web Framework: **Fiber v2** · Database: **PostgreSQL** (via `pgx/v5`)  
> Scope: exactly **10 endpoints**, **2 roles** (`admin`, `mahasiswa`), token auth on every endpoint except login.

---

## Table of Contents

1. [Folder Structure](#folder-structure)
2. [Clean Architecture Layers](#clean-architecture-layers)
3. [Dependency Graph](#dependency-graph)
4. [Package-by-Package Reference](#package-by-package-reference)
5. [API Routes](#api-routes)
6. [Endpoint Specification & Error Matrix](#endpoint-specification--error-matrix)
7. [Database Schema, Migrations & Seeder](#database-schema-migrations--seeder)
8. [Business Rules](#business-rules)
9. [Authorization (Role-Based Access)](#authorization-role-based-access)
10. [Authentication Flow](#authentication-flow)
11. [Enrollment Transaction Flow](#enrollment-transaction-flow)
12. [External Dependencies](#external-dependencies)
13. [Environment Variables](#environment-variables)
14. [Startup Sequence](#startup-sequence)
15. [Implementation Roadmap (Incremental Commits)](#implementation-roadmap-incremental-commits)
16. [Testing Plan](#testing-plan)
17. [Requirements Checklist](#requirements-checklist)
18. [Assumptions & Open Decisions](#assumptions--open-decisions)

---

## Folder Structure

```
siakad-mini/
├── main.go                         # Entry point — wires everything together
├── go.mod                          # Module: siakad-mini (Go 1.26)
├── go.sum
├── .env / .env.example             # Environment config
├── siakad-mini.exe                 # Compiled binary
│
├── config/                         # [Infrastructure] App-wide config & bootstrap
│   ├── app.go                      #   Creates Fiber app, registers middleware + routes
│   ├── env.go                      #   Loads .env, GetEnv / GetEnvInt helpers
│   └── logger.go                   #   Structured slog logger with log-rotation (lumberjack)
│
├── database/                       # [Infrastructure] Database connection
│   └── postgres.go                 #   pgxpool.Pool factory with health-ping
│
├── migrations/                     # Raw SQL migration files (run manually)
│   ├── 001_create_users.sql
│   ├── 002_create_students.sql
│   ├── 003_create_courses.sql
│   ├── 004_create_enrollments.sql  #   UNIQUE (student_id, course_id, tahun_akademik)
│   ├── 005_indexes.sql             #   Filter / search / join indexes
│   └── 006_seed.sql                #   Seed data: 1 admin, 20 mahasiswa, 10 courses
│
├── app/                            # Core application (Domain + Use-case layers)
│   ├── model/                      # [Domain] Entities & DTOs (no framework deps)
│   │   ├── student.go              #   Student, request/response types, StudentListQuery
│   │   ├── course.go               #   Course, CourseWithQuota, CourseListQuery
│   │   ├── enrollment.go           #   Enrollment, CreateEnrollmentRequest
│   │   ├── auth.go                 #   LoginRequest, LoginResponse, AuthUser
│   │   └── user.go                 #   User, Role constants, ErrorResponse
│   │
│   ├── repository/                 # [Use-case boundary] Repository interfaces + Postgres impls
│   │   ├── student_repository.go   #   StudentRepository interface + studentPostgresRepository
│   │   ├── user_repository.go      #   UserRepository interface + userPostgresRepository
│   │   ├── course_repository.go    #   CourseRepository interface + coursePostgresRepository
│   │   └── enrollment_repository.go#   EnrollmentRepository interface + enrollmentPostgresRepository
│   │
│   └── service/                    # [Use-case] Business logic + HTTP handlers
│       ├── student_service.go      #   StudentService — List, Get, Create, Update, Delete
│       ├── student_rules.go        #   Pure funcs: ApplyUpdate, ValidateSort, CountLastPage
│       ├── student_authz_rules.go  #   CanAccessStudent — admin-or-owner check
│       ├── student_rules_test.go   #   Unit tests for student rules
│       ├── course_service.go       #   CourseService — List
│       ├── enrollment_service.go   #   EnrollmentService — Create, Delete
│       ├── enrollment_rules.go     #   Pure funcs: MaxSKSByIPK, CheckSKSLimit, ValidateAcademicYear
│       ├── enrollment_rules_test.go#   Unit tests for enrollment rules
│       ├── auth_service.go         #   AuthService — Login, Me
│       ├── auth_rules.go           #   Pure funcs: ValidateLogin
│       └── auth_rules_test.go      #   Unit tests for auth rules
│
├── middleware/                     # [Interface] HTTP middleware
│   ├── middleware.go               #   Register global middleware; RequestLogger, RequireJSON, CORS
│   ├── auth.go                     #   RequireAuth (JWT Bearer), LoginRateLimiter
│   └── authz.go                    #   RequireRole (admin / mahasiswa gate)
│
├── route/                          # [Interface] Route registration
│   └── route.go                    #   Dependencies struct, Register()
│
├── helper/                         # [Shared] Cross-cutting utilities (no business logic)
│   ├── jwt.go                      #   JWTManager — GenerateAccess, Parse (HS256)
│   ├── authz.go                    #   HasAnyRole
│   ├── security.go                 #   HashPassword, VerifyPassword, VerifyDummyPassword
│   ├── errors.go                   #   AppError type + factory funcs (BadRequest, NotFound, ...)
│   ├── response.go                 #   Success, SuccessList, Created, NoContent
│   ├── request.go                  #   RequestContext, ParamID, ParseStudentListQuery, ParseCourseListQuery
│   ├── context.go                  #   CurrentUser (reads AuthUser from Fiber locals)
│   └── validator.go                #   ValidateStruct + custom tags (nim, angkatan, tahunakademik)
│
└── logs/                           # Runtime log output (auto-created)
    └── app.log                     #   Rotated JSON logs (lumberjack: 10 MB, 5 backups, 14 days)
```

Same layout as the reference project. Differences are only the data-model files: `course` / `enrollment` replace the refresh-token, RBAC, cursor, and CSV files, and seeding lives in `006_seed.sql`.

---

## Clean Architecture Layers

```
+-------------------------------------------------------------+
|  FRAMEWORKS & DRIVERS  (outermost — replaceable)            |
|  config/ · database/ · middleware/ · route/ · helper/       |
|  Fiber v2 · pgx · lumberjack · godotenv · jwt               |
+-------------------------------------------------------------+
|  INTERFACE ADAPTERS                                         |
|  middleware/   — converts HTTP <-> domain                   |
|  route/        — wires HTTP verbs to service handlers       |
|  helper/       — shared adapters (JWT, errors, responses)   |
+-------------------------------------------------------------+
|  USE CASES  (app/service/)                                  |
|  AuthService       — login / me                             |
|  StudentService    — CRUD + soft delete + ownership authz   |
|  CourseService     — list with computed quota               |
|  EnrollmentService — transactional KRS add / cancel         |
|  *_rules.go        — pure, testable business rules          |
+-------------------------------------------------------------+
|  ENTITIES  (app/model/)                        [innermost]  |
|  User · Student · Course · Enrollment · AuthUser · DTOs     |
+-------------------------------------------------------------+
```

**Key principle**: Inner layers (`model`, `service`) have **zero** imports from outer layers.  
The `repository` package sits at the boundary — it defines interfaces (inner) whose implementations use `pgx` (outer). Operations that must be atomic (create student + user, enroll) are exposed as repository methods, so services never touch `pgx.Tx` directly.

---

## Dependency Graph

```
main.go
 +-> config          (LoadEnv, NewLogger, NewApp)
 +-> database        (NewPool)            --uses--> config
 +-> app/repository  (New*Repository)     --uses--> app/model, pgx
 +-> helper          (NewJWTManager)
 +-> app/service     (NewAuthService, NewStudentService, NewCourseService, NewEnrollmentService)
 |    +--uses--> app/repository (interfaces), app/model, helper
 +-> route           (Dependencies struct, Register)
 |    +--uses--> app/service, helper, middleware
 +-> config.NewApp
      +--uses--> middleware.Register, route.Register

middleware
 +--uses--> helper (AppError, CurrentUser, JWTManager, HasAnyRole)

helper
 +--uses--> app/model (only)
```

---

## Package-by-Package Reference

### `app/model` — Domain Entities & DTOs

| Type | Purpose |
|---|---|
| `User` | User entity (`id`, `email`, `password`, `role`, `created_at`); `Role` type (`admin` / `mahasiswa`) |
| `AuthUser` | Slim auth context stored in Fiber locals (`user_id`, `email`, `role`) |
| `LoginRequest` | POST /auth/login body — `email` (required, email format), `password` (required, min 8) |
| `LoginResponse` | `access_token`, `token_type`, `expires_in`, `user{id,email,role}` |
| `MeResponse` | User data + optional `student{nim,nama,prodi,angkatan}` when role = `mahasiswa` |
| `Student` | Core entity (`id`, `user_id`, `nim`, `nama`, `prodi`, `angkatan`, `ipk_terakhir`, `deleted_at`) |
| `CreateStudentRequest` | POST body — `nim` (12 digits), `nama`, `email`, `prodi`, `angkatan`, `ipk_terakhir` (optional) |
| `UpdateStudentRequest` | PUT body — `nama`, `prodi`, `angkatan`, `ipk_terakhir` (`nim` is immutable and not accepted) |
| `StudentListQuery` | `page`, `per_page`, `prodi`, `angkatan`, `search`, `sort` |
| `StudentDetail` | Student + `enrollments[]` + `total_sks` + `batas_sks` |
| `Course` | Core entity (`id`, `kode_mk`, `nama_mk`, `sks`, `semester`, `kuota`) |
| `CourseWithQuota` | Course + `terisi` + `sisa_kuota` (computed from `enrollments`) |
| `CourseListQuery` | `semester`, `search`, `available` |
| `Enrollment` | Entity (`id`, `student_id`, `course_id`, `tahun_akademik`, `created_at`) |
| `CreateEnrollmentRequest` | POST body — `course_id`, `tahun_akademik` (e.g. `2026/2027-Ganjil`) |
| `EnrollmentDetail` | Enrollment joined with course info (used in student detail) |
| `Meta` | `current_page`, `per_page`, `total`, `last_page` |
| `WebResponse` | Standard success envelope (`success`, `message`, `data`, `meta`) |
| `ErrorResponse` | Standard error envelope (`success`, `message`, `errors` map of field -> messages) |

---

### `app/repository` — Data Access

All repositories are defined as **interfaces** (enabling mocks in tests), with a single private `*postgresRepository` implementation each. Each is backed by `pgxpool.Pool`; methods that must be atomic open their own `pgx.Tx` internally.

| Interface | Methods |
|---|---|
| `UserRepository` | `FindByEmail` (excludes soft-deleted mahasiswa), `FindByID` |
| `StudentRepository` | `FindAll` (filters + pagination), `FindByID`, `FindByUserID`, `CreateWithUser` (users + students in one transaction), `Update`, `SoftDelete` |
| `CourseRepository` | `FindAllWithQuota` (JOIN/COUNT on enrollments), `FindByID` |
| `EnrollmentRepository` | `FindByID`, `Delete`, `ListByStudent`, `TotalSKSByStudent`, `WithinTx(ctx, fn func(EnrollmentTx) error) error` — begin / commit / rollback |
| `EnrollmentTx` (passed to `WithinTx`) | `LockStudent`, `LockCourse` (`SELECT ... FOR UPDATE`), `Exists`, `CountByCourse`, `SumSKS`, `Create` — all bound to one transaction |

**Sentinel errors** (defined in `student_repository.go`, shared across the package):

| Error | Meaning |
|---|---|
| `ErrNotFound` | `pgx.ErrNoRows` — resource absent |
| `ErrDuplicate` | PostgreSQL error code `23505` — unique violation (nim, email, or enrollment) |

**Query strategies**:
- `StudentRepository.FindAll` — LIMIT/OFFSET with dynamic `WHERE deleted_at IS NULL` + optional `prodi`, `angkatan`, `search` (`ILIKE` on `nim` OR `nama`); `ORDER BY` chosen from a **whitelist** (`nama` → `nama ASC`, `-ipk_terakhir` → `ipk_terakhir DESC`), never interpolated from user input; all values use bind parameters.
- `CourseRepository.FindAllWithQuota` — `LEFT JOIN enrollments` + `GROUP BY`, `terisi = COUNT(e.id)`, `sisa_kuota = kuota - COUNT(e.id)`; `available=true` adds `HAVING COUNT(e.id) < kuota`.

---

### `app/service` — Use Cases

#### `AuthService`

| Method | HTTP verb | Logic |
|---|---|---|
| `Login` | `POST /auth/login` | Validate -> find user (soft-deleted mahasiswa excluded) -> verify bcrypt (dummy compare when user not found) -> issue access JWT |
| `Me` | `GET /auth/me` | Load user; if role = `mahasiswa` also load the student profile (401 if the student was soft-deleted) |

#### `StudentService`

| Method | HTTP verb | Logic |
|---|---|---|
| `List` | `GET /students` | Parse query (defaults/limits), filters, search, whitelisted sort, meta |
| `Get` | `GET /students/:id` | 404 if missing/soft-deleted -> admin-or-owner check (403) -> attach enrollments, `total_sks`, `batas_sks` |
| `Create` | `POST /students` | Validate -> **one transaction**: insert `users` (role `mahasiswa`, password = `HashPassword(nim)`) + insert `students`; duplicate nim/email -> 422 |
| `Update` | `PUT /students/:id` | Validate -> `ApplyUpdate` onto existing record (`nim` untouched) |
| `Delete` | `DELETE /students/:id` | Soft delete (`deleted_at = now()`); student disappears from list and cannot log in |

#### `CourseService`

| Method | HTTP verb | Logic |
|---|---|---|
| `List` | `GET /courses` | Filters `semester`, `search`, `available`; returns `terisi` + `sisa_kuota` |

#### `EnrollmentService`

| Method | HTTP verb | Logic |
|---|---|---|
| `Create` | `POST /enrollments` | Single transaction with row locking: duplicate (409) -> quota (422) -> SKS limit (422); see [Enrollment Transaction Flow](#enrollment-transaction-flow) |
| `Delete` | `DELETE /enrollments/:id` | 404 if missing -> 403 if `enrollment.student_id` is not the caller's -> delete (quota is computed, so it frees automatically) |

#### Pure business-rule functions (no HTTP, fully testable)

| Function | File | Purpose |
|---|---|---|
| `ValidateLogin(req)` | `auth_rules.go` | Non-empty email/password, email format, min 8 chars |
| `CountLastPage(total, perPage)` | `student_rules.go` | Ceiling division for `meta.last_page` |
| `ValidateSort(sort)` | `student_rules.go` | Only `nama` or `-ipk_terakhir` accepted (empty = default) |
| `ApplyUpdate(current, req)` | `student_rules.go` | Merges allowed fields onto the existing student (never changes `nim`) |
| `CanAccessStudent(current, studentUserID)` | `student_authz_rules.go` | `true` if role is `admin` OR `current.UserID == studentUserID` |
| `MaxSKSByIPK(ipk)` | `enrollment_rules.go` | `>= 3.00` -> 24; `2.50–2.99` -> 21; `< 2.50` -> 18 |
| `CheckSKSLimit(taken, adding, max)` | `enrollment_rules.go` | Returns remaining SKS and an error message such as "Sisa SKS Anda X, mata kuliah membutuhkan Y" |
| `ValidateAcademicYear(s)` | `enrollment_rules.go` | Format `YYYY/YYYY+1-Ganjil\|Genap` |
| `IsCourseFull(terisi, kuota)` | `enrollment_rules.go` | `terisi >= kuota` |

---

### `middleware` — HTTP Guards

| Middleware | File | Description |
|---|---|---|
| `requestid.New()` | (fiber built-in) | Injects `X-Request-ID` / `requestid` local |
| `recover.New()` | (fiber built-in) | Panic recovery -> 500 response (no stack trace in production) |
| `helmet.New()` | (fiber built-in) | Sets security HTTP headers |
| `corsPolicy(origins)` | `middleware.go` | CORS with configurable allowed origins |
| `RequestLogger(logger)` | `middleware.go` | Structured request log (method, path, status, duration, IP, user) |
| `RequireJSON` | `middleware.go` | Rejects non-`application/json` on POST/PUT |
| `RequireAuth(jwt)` | `auth.go` | Validates `Authorization: Bearer <token>`; 401 if missing/invalid/expired; stores `AuthUser` in locals |
| `LoginRateLimiter()` | `auth.go` | Fiber `limiter`: max 5 **failed** attempts/min per IP on `/auth/login` (`SkipSuccessfulRequests`), returns 429 |
| `RequireRole(roles...)` | `authz.go` | Role gate — 403 if `AuthUser.Role` is not in the allowed list |

---

### `helper` — Shared Utilities

| File | Key Types / Functions |
|---|---|
| `jwt.go` | `JWTManager` — `GenerateAccess(user)` (HS256), `Parse(token)` -> `AuthUser`, exposes TTL for `expires_in` |
| `authz.go` | `HasAnyRole(role, allowed...)` |
| `security.go` | `HashPassword` (bcrypt cost 12), `VerifyPassword`, `VerifyDummyPassword` (timing-safe) |
| `errors.go` | `AppError{Status, Message, Fields, cause}` + factories: `BadRequest`, `Unauthorized`, `Forbidden`, `NotFound`, `Conflict` (409), `Validation` (422), `Internal`, `TooManyRequests` (429) |
| `response.go` | `Success`, `SuccessList` (data + meta), `Created`, `NoContent` |
| `request.go` | `RequestContext` (5s timeout), `ParamID`, `ParseStudentListQuery` (`page` default 1, `per_page` default 10, max 50), `ParseCourseListQuery` |
| `context.go` | `CurrentUser(c)` — reads `AuthUser` from Fiber locals key `"authUser"` |
| `validator.go` | `ValidateStruct` -> `map[string][]string` (Indonesian messages) + custom tags: `nim` (regex `^[0-9]{12}$`), `angkatan` (4 digits, `<=` current year), `tahunakademik` |

---

### `config` — Bootstrap

| File | Purpose |
|---|---|
| `env.go` | `LoadEnv()` (godotenv), `GetEnv(key, fallback)`, `GetEnvInt(key, fallback)` |
| `logger.go` | `NewLogger()` — `slog.JSONHandler` writing to stdout **+** `logs/app.log` via lumberjack (10 MB rotate, 5 backups, 14 days TTL) |
| `app.go` | `NewApp(logger, deps)` — creates `fiber.App`, installs a global `ErrorHandler` that normalises all errors to `ErrorResponse`; when `APP_ENV=production`, unexpected errors return a generic 500 message and the real cause is logged only |

---

### `database`

| File | Purpose |
|---|---|
| `postgres.go` | `NewPool(ctx)` — builds DSN from env, creates `pgxpool.Pool` (max 10 conns, min 2, 1h lifetime, 30m idle), pings on start |

---

### `route`

| File | Purpose |
|---|---|
| `route.go` | `Dependencies` struct (pool, JWT, services), `Register(app, deps)` |

---

## API Routes

All routes are under `/api/v1`. Every route except `POST /auth/login` is behind `RequireAuth`.

### Auth (`/api/v1/auth`)

| # | Method | Path | Middleware | Handler |
|---|---|---|---|---|
| 1 | `POST` | `/login` | RequireJSON, LoginRateLimiter | `AuthService.Login` |
| 2 | `GET` | `/me` | RequireAuth | `AuthService.Me` |

### Students (`/api/v1/students`)

| # | Method | Path | Middleware | Handler |
|---|---|---|---|---|
| 3 | `GET` | `/` | RequireAuth, RequireRole(`admin`) | `StudentService.List` |
| 4 | `POST` | `/` | RequireAuth, RequireJSON, RequireRole(`admin`) | `StudentService.Create` |
| 5 | `GET` | `/:id` | RequireAuth | `StudentService.Get` *(internal admin-or-owner check)* |
| 6 | `PUT` | `/:id` | RequireAuth, RequireJSON, RequireRole(`admin`) | `StudentService.Update` |
| 7 | `DELETE` | `/:id` | RequireAuth, RequireRole(`admin`) | `StudentService.Delete` |

### Courses (`/api/v1/courses`)

| # | Method | Path | Middleware | Handler |
|---|---|---|---|---|
| 8 | `GET` | `/` | RequireAuth | `CourseService.List` |

### Enrollments (`/api/v1/enrollments`)

| # | Method | Path | Middleware | Handler |
|---|---|---|---|---|
| 9 | `POST` | `/` | RequireAuth, RequireJSON, RequireRole(`mahasiswa`) | `EnrollmentService.Create` |
| 10 | `DELETE` | `/:id` | RequireAuth, RequireRole(`mahasiswa`) | `EnrollmentService.Delete` *(internal ownership check)* |

---

## Endpoint Specification & Error Matrix

| # | Endpoint | Success | Error codes | Notes |
|---|---|---|---|---|
| 1 | `POST /auth/login` | 200 | 401 wrong credentials · 422 validation · 429 > 5 failed/min | Returns `access_token`, `token_type`, `expires_in`, `user{id,email,role}` |
| 2 | `GET /auth/me` | 200 | 401 token missing/invalid/expired | Adds `student{nim,nama,prodi,angkatan}` for mahasiswa |
| 3 | `GET /students` | 200 | 401 · 403 | `data[]` + `meta{current_page,per_page,total,last_page}` |
| 4 | `POST /students` | 201 | 401 · 403 · 422 duplicate nim/email or invalid | `users` + `students` in one transaction; initial password = hashed `nim` |
| 5 | `GET /students/:id` | 200 | 401 · 403 other student's data · 404 missing/soft-deleted | Adds `enrollments[]`, `total_sks`, `batas_sks` |
| 6 | `PUT /students/:id` | 200 | 401 · 403 · 404 · 422 | `nim` cannot be changed |
| 7 | `DELETE /students/:id` | 204 | 401 · 403 · 404 | Soft delete; blocks login afterwards |
| 8 | `GET /courses` | 200 | 401 | Each item has `terisi` and `sisa_kuota` |
| 9 | `POST /enrollments` | 201 | 401 · 403 not mahasiswa · 409 already taken · 422 full quota / SKS exceeded / invalid `course_id` | SKS error message states remaining SKS |
| 10 | `DELETE /enrollments/:id` | 204 | 401 · 403 other student's enrollment · 404 | Quota frees automatically (computed) |

Every response uses the uniform envelope. Any unhandled error becomes **500** with a generic message in production. Status codes in use: `200, 201, 204, 401, 403, 404, 409, 422, 429, 500`.

---

## Database Schema, Migrations & Seeder

Migrations are plain SQL files applied in numeric order (`psql -f`).

| File | Creates / Alters |
|---|---|
| `001_create_users.sql` | `users` (id, email UNIQUE, password, role CHECK IN (`admin`,`mahasiswa`), created_at) |
| `002_create_students.sql` | `students` (id, user_id FK UNIQUE, nim UNIQUE, nama, prodi, angkatan, ipk_terakhir `NUMERIC(3,2)` CHECK 0–4, deleted_at nullable) |
| `003_create_courses.sql` | `courses` (id, kode_mk UNIQUE, nama_mk, sks CHECK > 0, semester, kuota CHECK >= 0) |
| `004_create_enrollments.sql` | `enrollments` (id, student_id FK, course_id FK, tahun_akademik, created_at) + `UNIQUE (student_id, course_id, tahun_akademik)` |
| `005_indexes.sql` | `students(prodi, angkatan) WHERE deleted_at IS NULL`, `courses(semester)`, `enrollments(course_id)`, `enrollments(student_id, tahun_akademik)` |
| `006_seed.sql` | Seed data: 1 admin, 20 mahasiswa (with `users` rows), 10 courses |

### ER Summary

```
users (id PK, email UNIQUE, password, role, created_at)
  |
  +--1:1-- students (id PK, user_id FK->users.id UNIQUE, nim UNIQUE, nama, prodi,
  |                  angkatan, ipk_terakhir, deleted_at nullable)
  |            |
  |            +--1:N-- enrollments (id PK, student_id FK, course_id FK,
  |                                   tahun_akademik, created_at,
  |                                   UNIQUE (student_id, course_id, tahun_akademik))
  |                          |
  |                          +--N:1-- courses (id PK, kode_mk UNIQUE, nama_mk, sks,
  |                                            semester, kuota)
```

### Seed Data (`migrations/006_seed.sql`)

| Data | Count | Details |
|---|---|---|
| Admin | 1 | `admin@siakad.test`, development password `Admin12345` (change outside development) |
| Mahasiswa | 20 | NIM `187221000001`–`187221000020` (12 digits), email `mahasiswa01@siakad.test` ..., password = hashed NIM, mixed `prodi` and `angkatan` (2021–2024), IPK spread across all three tiers (>= 3.00, 2.50–2.99, < 2.50) to exercise the SKS rule |
| Courses | 10 | SKS 2–4, various `semester`, a few with very small `kuota` (1–2) to test "quota full" |

Passwords are hashed inside SQL with pgcrypto (`CREATE EXTENSION IF NOT EXISTS pgcrypto;` and `crypt(plain, gen_salt('bf', 12))`), which yields bcrypt hashes that Go's `VerifyPassword` accepts. Inserts use `ON CONFLICT DO NOTHING`, so the file can be re-run safely, and no plaintext password is ever stored.

---

## Business Rules

| # | Rule | Enforced in | Error |
|---|---|---|---|
| 1 | SKS limit per semester by latest IPK: `>= 3.00` -> 24, `2.50–2.99` -> 21, `< 2.50` -> 18 | `MaxSKSByIPK` + `CheckSKSLimit`, evaluated inside the enrollment transaction | 422 (message includes remaining SKS) |
| 2 | Same course cannot be taken twice in the same academic year | Pre-check `Exists` **and** DB `UNIQUE (student_id, course_id, tahun_akademik)` as a safety net (`23505` -> 409) | 409 |
| 3 | A course with full quota cannot be taken | `IsCourseFull` after locking the course row | 422 |
| 4 | Students can only access/modify their own KRS | `student_id` is **always** derived from the JWT user, never from the request body; `CanAccessStudent` / ownership check on delete | 403 |

---

## Authorization (Role-Based Access)

Two fixed roles, enforced by `RequireRole` at the route level plus ownership checks inside services. No permission tables are needed.

| Endpoint | admin | mahasiswa |
|---|:---:|:---:|
| `POST /auth/login` | YES (public) | YES (public) |
| `GET /auth/me` | YES | YES |
| `GET /students` | YES | NO (403) |
| `POST /students` | YES | NO (403) |
| `GET /students/:id` | YES | **Own record only** (403 otherwise) |
| `PUT /students/:id` | YES | NO (403) |
| `DELETE /students/:id` | YES | NO (403) |
| `GET /courses` | YES | YES |
| `POST /enrollments` | NO (403) | YES |
| `DELETE /enrollments/:id` | NO (403) | **Own enrollment only** (403 otherwise) |

**Ownership check**: `students.user_id == current.UserID` for student detail; `enrollments.student_id == current student's id` for enrollment delete.

---

## Authentication Flow

```
[Client]                         [Server]
   |-- POST /auth/login --------> AuthService.Login
   |   { email, password }         * validate body (422 on failure)
   |                               * LoginRateLimiter: > 5 failed/min per IP -> 429
   |                               * find user by email (soft-deleted mahasiswa excluded)
   |                               * verify bcrypt hash (dummy compare if not found -> 401)
   |                               * issue access JWT (HS256, claims: sub, email, role, iss, exp)
   |<-- { access_token, token_type: "Bearer", expires_in, user } ----
   |
   |-- GET /api/v1/... ---------> middleware.RequireAuth
   |   Authorization: Bearer <AT> * parse & verify JWT signature + expiry (401 otherwise)
   |                               * store AuthUser in c.Locals("authUser")
   |                               * RequireRole(...) -> 403 if role not allowed
   |<-- 200 / 401 / 403 ----------
```

No refresh-token flow: the case study only requires an access token with `expires_in`.

---

## Enrollment Transaction Flow

`POST /enrollments` runs entirely inside `EnrollmentRepository.WithinTx`. Locks are always taken in the same order (**student -> course**) to avoid deadlocks.

```
1. Validate body: course_id required, tahun_akademik matches YYYY/YYYY+1-Ganjil|Genap   -> 422
2. Resolve current student from JWT user_id (must exist, not soft-deleted)               -> 401/403
BEGIN
3. SELECT ... FROM students WHERE id = $1 FOR UPDATE   -- serializes this student's enrollments (SKS race)
4. SELECT ... FROM courses  WHERE id = $2 FOR UPDATE   -- serializes quota checks per course; missing -> 422
5. Exists(student, course, tahun_akademik)?            -> 409 Conflict
6. COUNT(enrollments for course) >= kuota?             -> 422 "Kuota mata kuliah penuh"
7. taken = SUM(sks) of student's enrollments in tahun_akademik
   max   = MaxSKSByIPK(student.ipk_terakhir)
   taken + course.sks > max?                           -> 422 "Total SKS melebihi batas. Sisa SKS: X"
8. INSERT INTO enrollments (...)                       -- UNIQUE violation (23505) -> 409
COMMIT  -> 201 Created
```

Any returned error rolls the transaction back.

---

## External Dependencies

### Direct (declared in `go.mod`)

| Package | Version | Usage |
|---|---|---|
| `github.com/gofiber/fiber/v2` | v2.52.15 | HTTP framework + built-in middleware (cors, helmet, recover, requestid, limiter) |
| `github.com/golang-jwt/jwt/v5` | v5.3.1 | JWT generation & parsing (HS256) |
| `github.com/jackc/pgx/v5` | v5.10.0 | PostgreSQL driver + connection pool (`pgxpool`) + transactions |
| `github.com/go-playground/validator/v10` | latest v10 | Struct validation in `helper/validator.go` |
| `github.com/joho/godotenv` | v1.5.1 | `.env` file loading |
| `golang.org/x/crypto` | v0.57.0 | bcrypt password hashing |
| `gopkg.in/natefinch/lumberjack.v2` | v2.2.1 | Log file rotation |

### Notable Indirect

| Package | Pulled by |
|---|---|
| `github.com/google/uuid` | Fiber (request-id) |
| `github.com/valyala/fasthttp` | Fiber internals |
| `github.com/klauspost/compress` | Fiber compression |
| `github.com/andybalholm/brotli` | Fiber Brotli support |

Versions mirror the reference project; pin them with `go get` and commit `go.sum`.

---

## Environment Variables

| Variable | Default | Description |
|---|---|---|
| `APP_PORT` | `3000` | HTTP listen port |
| `APP_NAME` | `siakad-mini` | Fiber app name |
| `APP_ENV` | `development` | `development` / `production` (production hides internal error details) |
| `LOG_LEVEL` | `info` | `debug` / `info` / `warn` / `error` |
| `ALLOWED_ORIGINS` | `*` | CORS allowed origins |
| `DB_HOST` | `localhost` | PostgreSQL host |
| `DB_PORT` | `5432` | PostgreSQL port |
| `DB_USER` | `postgres` | DB username |
| `DB_PASSWORD` | *(empty)* | DB password |
| `DB_NAME` | `siakad_mini` | Database name |
| `DB_SSLMODE` | `disable` | `disable` / `require` / `verify-full` |
| `DB_MAX_CONNS` | `10` | Max connections in pool |
| `JWT_SECRET` | `changeme` | HMAC signing secret — **must be changed in production** |
| `JWT_ISSUER` | `siakad-mini` | JWT `iss` claim |
| `JWT_ACCESS_TTL` | `1h` | Access token TTL (`time.ParseDuration` format) |
| `LOGIN_RATE_MAX` | `5` | Max failed login attempts per window |
| `LOGIN_RATE_WINDOW` | `1m` | Rate-limit window |

---

## Startup Sequence

```
main()
  1. config.LoadEnv()                     -- load .env
  2. config.NewLogger()                   -- setup slog + lumberjack
  3. database.NewPool(ctx)                -- open pgxpool, ping DB
  4. repository.New*Repository(pool)      -- create user/student/course/enrollment repos
  5. helper.NewJWTManager(...)            -- build JWT manager
  6. service.NewAuthService(...)          -- wire auth use-cases
  7. service.NewStudentService(...)       -- wire student use-cases
  8. service.NewCourseService(...)        -- wire course use-cases
  9. service.NewEnrollmentService(...)    -- wire enrollment use-cases (uses EnrollmentRepository.WithinTx)
 10. config.NewApp(logger, deps)          -- create Fiber + middleware + routes
 11. app.Listen(":PORT")  [goroutine]     -- start HTTP server
 12. signal.Notify(SIGINT/SIGTERM)        -- wait for shutdown signal
 13. app.ShutdownWithContext(10s)         -- graceful shutdown
 14. pool.Close()                         -- release DB connections
```

One-time setup before step 1: `createdb siakad_mini` -> run `migrations/*.sql` in order (`006_seed.sql` inserts the seed data).

---

## Implementation Roadmap (Incremental Commits)

The case study requires a GitHub repo with a **gradual commit history**. Suggested order (one or more commits per step, conventional-commit style):

| Phase | Commit scope | Deliverable |
|---|---|---|
| 1 | `chore: init project` | `go mod init`, folder skeleton, `.gitignore`, `.env.example`, README stub |
| 2 | `feat(config): env, logger, db pool` | `config/`, `database/`, health of DB connection |
| 3 | `feat(db): migrations` | `001`–`004` table files + `005` indexes |
| 4 | `feat(helper): errors, response, security, validator` | AppError, envelope helpers, bcrypt, custom validator tags |
| 5 | `feat(db): seed data` | `006_seed.sql` — 1 admin, 20 mahasiswa, 10 courses |
| 6 | `feat(auth): login + jwt + me` | Endpoints 1–2, `RequireAuth`, `LoginRateLimiter` (429) |
| 7 | `feat(authz): role middleware` | `RequireRole`, 403 behaviour |
| 8 | `feat(student): list, create, detail, update, delete` | Endpoints 3–7, pagination/filter/search/sort, soft delete, create in a transaction |
| 9 | `feat(course): list with quota` | Endpoint 8 (`terisi`, `sisa_kuota`, `available`) |
| 10 | `feat(enrollment): add + cancel KRS` | Endpoints 9–10, `WithinTx`, row locking, SKS/quota/duplicate rules |
| 11 | `test: unit tests for rules` | `*_rules_test.go` |
| 12 | `docs: README + API examples` | Setup steps, curl/Postman examples, final polish |

---

## Testing Plan

**Unit tests (pure rules, `go test ./app/service/...`)**

| Area | Cases |
|---|---|
| `MaxSKSByIPK` | boundaries: 4.00, 3.00, 2.99, 2.50, 2.49, 0.00 |
| `CheckSKSLimit` | exactly at limit (allowed), one over (rejected), remaining-SKS message |
| `ValidateAcademicYear` | valid `2026/2027-Ganjil`, `2026/2027-Genap`; invalid `2026/2028-Ganjil`, `2026-Ganjil`, empty |
| `CanAccessStudent` | admin (any), owner, other mahasiswa |
| `ApplyUpdate` / `CountLastPage` / `ValidateSort` | nim unchanged, ceiling division, rejects unknown sort |
| `ValidateLogin` | bad email, short password, empty fields |

**Manual / integration checklist (Postman or curl)**

- Login OK (200) · wrong password (401) · invalid body (422) · 6th failed attempt in a minute (429)
- No token / bad token / expired token -> 401 on every protected endpoint
- mahasiswa calling admin endpoints -> 403 · admin calling `POST /enrollments` -> 403
- Create student with duplicate nim/email -> 422; success creates both `users` and `students` rows (and neither if one fails)
- Soft-deleted student: absent from list, 404 on detail, cannot log in
- Enroll twice in same course + same `tahun_akademik` -> 409
- Enroll in a full course -> 422 · exceed SKS cap for each IPK tier -> 422 with remaining SKS in message
- Concurrent enrollments (e.g. two requests racing for the last seat) -> only one succeeds
- Cancel own enrollment -> 204 and `sisa_kuota` increases · cancel someone else's -> 403 · unknown id -> 404
- Force an internal error with `APP_ENV=production` -> generic 500, no stack trace

---

## Requirements Checklist

| Requirement | Where it is covered |
|---|---|
| Go + Fiber (per practicum) + PostgreSQL | Header, [External Dependencies](#external-dependencies) |
| Exactly 10 endpoints as defined | [API Routes](#api-routes) |
| Migration + seeder (1 admin, 20 mahasiswa, 10 courses) | [Database Schema, Migrations & Seeder](#database-schema-migrations--seeder) (`006_seed.sql`) |
| Passwords stored as hashes | `helper/security.go` (bcrypt), seeder, `POST /students` |
| Token auth on all endpoints except login | `RequireAuth` on every route except `POST /auth/login` |
| Business rules 1–4 | [Business Rules](#business-rules) |
| Uniform JSON response + error format | `model.WebResponse` / `model.ErrorResponse`, global ErrorHandler |
| Status codes 200/201/204/401/403/404/409/422/429/500 | [Endpoint Specification & Error Matrix](#endpoint-specification--error-matrix) |
| No stack-trace leak in production | `APP_ENV=production` handling in `config/app.go` |
| Transactions (create student, enroll) with row locking | `CreateWithUser`, `EnrollmentRepository.WithinTx`, [Enrollment Transaction Flow](#enrollment-transaction-flow) |
| Rate limiting login (5 failures/min -> 429) | `LoginRateLimiter` |
| GitHub repo with gradual commits | [Implementation Roadmap](#implementation-roadmap-incremental-commits) |

---

## Assumptions & Open Decisions

The case study leaves a few points open. Defaults chosen for this plan (adjust if the lecturer says otherwise):

1. **`terisi` / quota scope** — counted across **all** `enrollments` of a course (the courses endpoint has no `tahun_akademik` parameter). If quota should be per academic year, add a `tahun_akademik` filter to the count and to `GET /courses`.
2. **`total_sks` in student detail** — sum of SKS across all of the student's enrollments; each listed enrollment carries its `tahun_akademik`. The **SKS cap check** on enrollment is always per `tahun_akademik`.
3. **`ipk_terakhir` is optional** — stored as `NOT NULL DEFAULT 0.00`, so a student without IPK falls in the lowest tier (18 SKS).
4. **`tahun_akademik` format** — `YYYY/YYYY+1-Ganjil` or `-Genap`; the second year must equal the first + 1.
5. **Rate limiter** counts only failed login responses (status >= 400) per IP, so successful logins do not consume the budget.
6. **Unique NIM/email stay unique after soft delete** — re-adding a deleted student's NIM returns 422 (duplicate).
7. **`PUT` ignores `nim`** if it is sent in the body (immutable); `email` is not updatable through this endpoint.
8. **Existing tokens of a soft-deleted student** remain cryptographically valid until expiry, but `GET /auth/me` returns 401 and student-scoped lookups return not found; login is blocked immediately.
9. **Error envelope** follows the case study (`success`, `message`, `errors`); the request id is exposed via the `X-Request-ID` header and logs rather than in the body.
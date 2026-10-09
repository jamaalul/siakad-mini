# Implementation Plan — `siakad-mini` (UTS PBE)

> Companion to `PROJECTPLAN.md`. The project is built **from the innermost Clean Architecture layer outward**, so every step only imports code that already exists. `go build ./...` must pass after every step.  
> **One step = one commit.**  
> **IDE agent:** reuse code from `../api-students` wherever possible — see [Reference Project](#reference-project-reuse-from-api-students).

---

## Table of Contents

1. [Reference Project: Reuse from `../api-students`](#reference-project-reuse-from-api-students)
2. [Build Order Overview](#build-order-overview)
3. [Rules for Every Step](#rules-for-every-step)
4. [Allowed Imports per Layer](#allowed-imports-per-layer)
5. [Steps](#steps)
   - [Phase A — Foundation](#phase-a--foundation)
   - [Phase B — Entities (innermost)](#phase-b--entities-innermost)
   - [Phase C — Shared Helpers & Pure Business Rules](#phase-c--shared-helpers--pure-business-rules)
   - [Phase D — Infrastructure & Repositories](#phase-d--infrastructure--repositories)
   - [Phase E — Use Cases (Services)](#phase-e--use-cases-services)
   - [Phase F — Interface Adapters (Middleware & Routes)](#phase-f--interface-adapters-middleware--routes)
   - [Phase G — Bootstrap (outermost)](#phase-g--bootstrap-outermost)
6. [End-to-End Smoke Test](#end-to-end-smoke-test)

---

## Reference Project: Reuse from `../api-students`

**Instruction for the IDE agent:** a finished sibling project, `api-students`, lives at `../api-students` (relative to this project's root). It uses the same stack and the same Clean Architecture layout. **Before writing any file in a step, look for its counterpart in `../api-students` and reuse it** instead of writing from scratch. Read the real code there first; do not guess from file names.

### How to reuse

1. Treat `../api-students` as **read-only**. Never modify it, and never copy `.env`, `logs/`, `.git/`, `*.exe`, or `go.sum`.
2. Copy the relevant file, then adapt it to this project: change the import path `api-students/...` to `siakad-mini/...`, and rename or remove anything that does not belong to this case study.
3. Use the **same dependency versions** as the reference `go.mod` (`go get <pkg>@<version>` instead of `@latest`).
4. If a counterpart is missing or differs from what this plan describes, follow **this plan** and mention the difference in your summary.
5. Do **not** carry over features this project does not have: refresh tokens, RBAC permission sets / role tables, cursor pagination, CSV export, `register` / `logout` / `refresh` / `PATCH` endpoints, the health-check endpoint, `username`, and the `strongpassword` / `nospace` validators.
6. Keep one step = one commit, even when most of the code is copied.

### Reuse map

| Step | Target file(s) here | Reference file(s) in `../api-students` | What to adapt |
|---:|---|---|---|
| 1 | `.gitignore`, `.env.example`, `go.mod` | `.env.example`, `go.mod` | New module name `siakad-mini`; env defaults per `PROJECTPLAN.md` (`DB_NAME=siakad_mini`, add `APP_ENV`, `LOGIN_RATE_*`, drop `JWT_REFRESH_TTL`); reuse dependency versions |
| 5 | `helper/errors.go` | `helper/errors.go` | Reuse almost as-is; ensure `Validation` = 422, `Conflict` = 409, `TooManyRequests` = 429; drop the `code` field if unused |
| 5 | `helper/response.go` | `helper/response.go` | Keep `Success`, `SuccessList`, `Created`, `NoContent`; drop `SuccessCursor`; meta = `current_page`, `per_page`, `total`, `last_page` |
| 5 | `helper/security.go` | `helper/security.go` | Keep `HashPassword`, `VerifyPassword`, `VerifyDummyPassword`; drop `RandomToken`, `SHA256Hex` |
| 5 | `helper/validator.go` | `helper/validator.go` | Keep the `ValidateStruct` core; `nim` becomes exactly 12 digits (`^[0-9]{12}$`); add `angkatan`, `tahunakademik`; drop `username`, `strongpassword`, `nospace`; error output must be `map[string][]string` |
| 6 | `helper/jwt.go` | `helper/jwt.go` | Claims use `email` instead of `username`; drop refresh-related code |
| 6 | `helper/authz.go` | `helper/authz.go` | Replace `PermissionSet` with `HasAnyRole(role, allowed...)` |
| 6 | `helper/context.go` | `helper/context.go` | Reuse as-is |
| 6 | `helper/request.go` | `helper/request.go` | Keep `RequestContext` and `ParamID`; replace `ParseListQuery` with `ParseStudentListQuery` / `ParseCourseListQuery`; drop cursor parsing |
| 7 | `app/service/auth_rules.go` + test | `app/service/auth_rules.go` + test | Keep only `ValidateLogin`; drop register and password-strength rules |
| 8 | `app/service/student_rules.go` + test | `app/service/student_rules.go` + test | `CountTotalPages` -> `CountLastPage`; `ApplyPatch` -> `ApplyUpdate`; drop `IsEmptyPatch`; add `ValidateSort` |
| 8 | `app/service/student_authz_rules.go` | `app/service/student_authz_rules.go` | "owner-or-permission" becomes "admin-or-owner" (`CanAccessStudent(current, studentUserID)`) |
| 10 | `config/env.go`, `config/logger.go` | `config/env.go`, `config/logger.go` | Reuse as-is |
| 10 | `database/postgres.go` | `database/postgres.go` | Reuse as-is (check default `DB_NAME`) |
| 11 | `app/repository/student_repository.go` | `app/repository/student_repository.go` | Keep sentinel errors `ErrNotFound` / `ErrDuplicate` and the dynamic `WHERE` / `ORDER BY` builder pattern; new columns and filters; add `deleted_at IS NULL`, `SoftDelete`, `CreateWithUser`; drop `FindAfterCursor` |
| 11 | `app/repository/user_repository.go` | `app/repository/user_repository.go` | Lookup by `email`; exclude soft-deleted mahasiswa; drop `Create` (moved into `CreateWithUser`) |
| 12 | `app/repository/course_repository.go` | *(no counterpart)* | Use `student_repository.go` as the style template |
| 13 | `app/repository/enrollment_repository.go` | *(no counterpart)* | Use `student_repository.go` as the style template; transaction code is new |
| 14 | `app/service/auth_service.go` | `app/service/auth_service.go` | Keep the handler pattern (parse -> validate -> repo -> response) and the timing-safe login; keep only `Login` and `Me`; no refresh tokens |
| 15 | `app/service/student_service.go` | `app/service/student_service.go` | Keep handler structure and error mapping; `Replace` -> `Update` (`PUT`); drop `Patch`, CSV and cursor logic; `Delete` becomes soft delete |
| 16–17 | `course_service.go`, `enrollment_service.go` | *(no counterpart)* | Follow the handler pattern from `student_service.go` |
| 18 | `middleware/middleware.go` | `middleware/middleware.go` | Reuse as-is (`Register`, `RequestLogger`, `RequireJSON`, CORS) |
| 19 | `middleware/auth.go` | `middleware/auth.go` | Keep `RequireAuth`; `LoginRateLimiter` must count **failed** attempts only (`SkipSuccessfulRequests`) |
| 19 | `middleware/authz.go` | `middleware/authz.go` | `RequirePermission` -> `RequireRole(roles...)` |
| 20 | `route/route.go` | `route/route.go` | Keep `Dependencies` and `Register`; new route table from `PROJECTPLAN.md`; drop the health-check handler |
| 21 | `config/app.go` | `config/app.go` | Keep the `fiber.App` setup and global `ErrorHandler`; error body must be `{success, message, errors}` with **no** `code`, `fields`, or `request_id`; production hides internal errors |
| 22 | `main.go` | `main.go` | Keep startup order and graceful shutdown; drop `LoadPermissions` / `PermissionSet` / token repository wiring; add `CourseService` and `EnrollmentService` |

Steps 2, 3, 4, 9 (migrations, seed, models, enrollment rules) have no direct counterpart; only borrow naming and style conventions from the reference.

---

## Build Order Overview

```
 inside                                                                        outside
   |                                                                              |
   v                                                                              v
 init -> migrations/seed -> model -> helper -> pure rules -> infra -> repository
      -> services (handlers) -> middleware -> route -> config/app -> main
```

| Step | Commit message | Layer |
|---:|---|---|
| 1 | `chore: init project` | Foundation |
| 2 | `feat(db): add schema migrations` | Foundation |
| 3 | `feat(db): add seed data` | Foundation |
| 4 | `feat(model): add entities and DTOs` | Entities |
| 5 | `feat(helper): add errors, response, security, validator` | Shared |
| 6 | `feat(helper): add jwt, authz, context, request` | Shared |
| 7 | `feat(service): add auth rules` | Use case (pure) |
| 8 | `feat(service): add student rules and authz rule` | Use case (pure) |
| 9 | `feat(service): add enrollment rules` | Use case (pure) |
| 10 | `feat(infra): add env, logger, postgres pool` | Infrastructure |
| 11 | `feat(repository): add user and student repositories` | Repository |
| 12 | `feat(repository): add course repository` | Repository |
| 13 | `feat(repository): add enrollment repository with transaction` | Repository |
| 14 | `feat(auth): add AuthService (login, me)` | Use case |
| 15 | `feat(student): add StudentService` | Use case |
| 16 | `feat(course): add CourseService` | Use case |
| 17 | `feat(enrollment): add EnrollmentService` | Use case |
| 18 | `feat(middleware): add global middleware` | Interface |
| 19 | `feat(middleware): add auth, rate limiter, role guard` | Interface |
| 20 | `feat(route): register all 10 endpoints` | Interface |
| 21 | `feat(config): add fiber app and error handler` | Bootstrap |
| 22 | `feat: wire main and graceful shutdown` | Bootstrap |
| 23 | `docs: add README and API examples` | Docs |

---

## Rules for Every Step

Before committing a step:

```bash
gofmt -w .
go vet ./...
go build ./...
go test ./...        # from step 7 onward
go mod tidy          # whenever a step introduces a new dependency
```

- Only import packages from earlier steps (see the table below).
- Do not commit `.env`, `logs/`, or compiled binaries.
- Commit message follows `type(scope): summary` (Conventional Commits).

---

## Allowed Imports per Layer

| Package | May import |
|---|---|
| `app/model` | standard library only |
| `helper` | `app/model`, Fiber, jwt, validator, bcrypt |
| `app/service/*_rules.go` | `app/model`, standard library (pure, no HTTP or DB) |
| `config` (env, logger), `database` | standard library, godotenv, lumberjack, pgx |
| `app/repository` | `app/model`, pgx |
| `app/service` (services) | `app/model`, `app/repository` (interfaces), `helper`, Fiber |
| `middleware` | `helper`, `app/model`, Fiber |
| `route` | `app/service`, `middleware`, `helper`, Fiber |
| `config/app.go` | `middleware`, `route`, `helper`, `app/model`, Fiber |
| `main.go` | everything |

---

## Steps

### Phase A — Foundation

#### Step 1 — `chore: init project`

**Files:** `go.mod`, `.gitignore`, `.env.example`, `main.go` (placeholder), folder skeleton

- `go mod init siakad-mini`
- Create folders: `config database migrations app/model app/repository app/service middleware route helper`
- `.gitignore`: `.env`, `logs/`, `*.exe`
- `.env.example`: all variables from the *Environment Variables* table in `PROJECTPLAN.md`
- Placeholder `main.go` that prints the app name

**Verify:** `go run .`

---

#### Step 2 — `feat(db): add schema migrations`

**Files:** `migrations/001_create_users.sql` … `005_indexes.sql`

- `001` users: `id`, `email` UNIQUE, `password`, `role` CHECK (`admin`,`mahasiswa`), `created_at`
- `002` students: `user_id` FK UNIQUE, `nim` UNIQUE, `nama`, `prodi`, `angkatan`, `ipk_terakhir NUMERIC(3,2) NOT NULL DEFAULT 0` CHECK 0–4, `deleted_at`
- `003` courses: `kode_mk` UNIQUE, `nama_mk`, `sks` CHECK > 0, `semester`, `kuota` CHECK >= 0
- `004` enrollments: FKs + `UNIQUE (student_id, course_id, tahun_akademik)`
- `005` indexes: partial index on students (`prodi`, `angkatan`) where `deleted_at IS NULL`, `courses(semester)`, `enrollments(course_id)`, `enrollments(student_id, tahun_akademik)`

**Verify:**
```bash
createdb siakad_mini
for f in migrations/00[1-5]_*.sql; do psql -d siakad_mini -f "$f"; done
psql -d siakad_mini -c "\dt"
```

---

#### Step 3 — `feat(db): add seed data`

**Files:** `migrations/006_seed.sql`

- `CREATE EXTENSION IF NOT EXISTS pgcrypto;`
- 1 admin (`admin@siakad.test`), 20 mahasiswa (NIM `187221000001`–`…020`, password = hashed NIM), 10 courses
- Hash with `crypt(plain, gen_salt('bf', 12))`; use `ON CONFLICT DO NOTHING`
- IPK spread over all three tiers; some courses with `kuota` 1–2

**Verify:**
```bash
psql -d siakad_mini -f migrations/006_seed.sql
psql -d siakad_mini -c "SELECT role, count(*) FROM users GROUP BY role;"   # admin 1, mahasiswa 20
psql -d siakad_mini -c "SELECT count(*) FROM courses;"                      # 10
```

---

### Phase B — Entities (innermost)

#### Step 4 — `feat(model): add entities and DTOs`

**Files:** `app/model/user.go`, `auth.go`, `student.go`, `course.go`, `enrollment.go`

- `user.go`: `User`, `Role` constants (`admin`, `mahasiswa`), `WebResponse`, `Meta`, `ErrorResponse`
- `auth.go`: `LoginRequest` (validate tags: `required,email` / `required,min=8`), `LoginResponse`, `AuthUser`, `MeResponse`
- `student.go`: `Student`, `CreateStudentRequest`, `UpdateStudentRequest` (no `nim`), `StudentListQuery`, `StudentDetail`
- `course.go`: `Course`, `CourseWithQuota` (`terisi`, `sisa_kuota`), `CourseListQuery`
- `enrollment.go`: `Enrollment`, `CreateEnrollmentRequest`, `EnrollmentDetail`
- Add `json` tags matching the case study (`nim`, `nama`, `ipk_terakhir`, `kode_mk`, ...) and `validate` tags

**Verify:** `go build ./...`

---

### Phase C — Shared Helpers & Pure Business Rules

#### Step 5 — `feat(helper): add errors, response, security, validator`

**Files:** `helper/errors.go`, `response.go`, `security.go`, `validator.go`  
**New deps:** `go get github.com/gofiber/fiber/v2 github.com/go-playground/validator/v10 golang.org/x/crypto`

- `errors.go`: `AppError{Status, Message, Fields, cause}` + `BadRequest`, `Unauthorized`, `Forbidden`, `NotFound`, `Conflict` (409), `Validation` (422), `TooManyRequests` (429), `Internal`
- `response.go`: `Success`, `SuccessList` (data + meta), `Created`, `NoContent` using the uniform envelope
- `security.go`: `HashPassword` (bcrypt cost 12), `VerifyPassword`, `VerifyDummyPassword`
- `validator.go`: `ValidateStruct` -> `map[string][]string` with Indonesian messages; custom tags `nim` (`^[0-9]{12}$`), `angkatan` (4 digits, <= current year), `tahunakademik` (shape `YYYY/YYYY-Ganjil|Genap`)

**Verify:** `go build ./...`

---

#### Step 6 — `feat(helper): add jwt, authz, context, request`

**Files:** `helper/jwt.go`, `authz.go`, `context.go`, `request.go`  
**New deps:** `go get github.com/golang-jwt/jwt/v5`

- `jwt.go`: `JWTManager` — `GenerateAccess(user)` (HS256, claims `sub`, `email`, `role`, `iss`, `exp`), `Parse(token) -> AuthUser`, TTL getter for `expires_in`
- `authz.go`: `HasAnyRole(role, allowed...)`
- `context.go`: `CurrentUser(c)` reads `AuthUser` from `c.Locals("authUser")`
- `request.go`: `RequestContext` (5s timeout), `ParamID`, `ParseStudentListQuery` (`page` default 1, `per_page` default 10, max 50), `ParseCourseListQuery` (`available=true`)

**Verify:** `go build ./...`

---

#### Step 7 — `feat(service): add auth rules`

**Files:** `app/service/auth_rules.go`, `auth_rules_test.go`

- `ValidateLogin(req)`: empty fields, email format, min 8 chars
- Tests: valid input, bad email, short password, empty fields

**Verify:** `go test ./app/service/...`

---

#### Step 8 — `feat(service): add student rules and authz rule`

**Files:** `app/service/student_rules.go`, `student_authz_rules.go`, `student_rules_test.go`

- `CountLastPage(total, perPage)` (ceiling division)
- `ValidateSort(sort)`: only `nama` or `-ipk_terakhir`, empty = default
- `ApplyUpdate(current, req)`: merge allowed fields, never touch `nim`
- `CanAccessStudent(current, studentUserID)`: admin OR owner
- Tests: page math (0 items, exact multiple, remainder), unknown sort rejected, `nim` unchanged, admin / owner / other mahasiswa

**Verify:** `go test ./app/service/...`

---

#### Step 9 — `feat(service): add enrollment rules`

**Files:** `app/service/enrollment_rules.go`, `enrollment_rules_test.go`

- `MaxSKSByIPK(ipk)`: `>= 3.00` -> 24, `2.50–2.99` -> 21, `< 2.50` -> 18
- `CheckSKSLimit(taken, adding, max)`: returns remaining SKS and a message such as "Sisa SKS Anda X, mata kuliah ini membutuhkan Y"
- `ValidateAcademicYear(s)`: format + second year = first + 1
- `IsCourseFull(terisi, kuota)`
- Tests: IPK boundaries (4.00, 3.00, 2.99, 2.50, 2.49, 0.00), exactly at limit vs one over, valid/invalid academic years

**Verify:** `go test ./app/service/...`

---

### Phase D — Infrastructure & Repositories

> `env`, `logger`, and `database` only depend on libraries, never on inner layers, so they can be built here to give the repositories a real database to run against. `config/app.go` (which depends on outer layers) waits until Step 21.

#### Step 10 — `feat(infra): add env, logger, postgres pool`

**Files:** `config/env.go`, `config/logger.go`, `database/postgres.go`  
**New deps:** `go get github.com/joho/godotenv gopkg.in/natefinch/lumberjack.v2 github.com/jackc/pgx/v5`

- `env.go`: `LoadEnv`, `GetEnv`, `GetEnvInt`
- `logger.go`: `slog.JSONHandler` -> stdout + `logs/app.log` (10 MB, 5 backups, 14 days)
- `postgres.go`: `NewPool(ctx)` — DSN from env, max 10 / min 2 conns, ping on start

**Verify:** temporary `main.go` that calls `LoadEnv`, `NewLogger`, `NewPool` and logs "connected" (do not commit the temporary change).

---

#### Step 11 — `feat(repository): add user and student repositories`

**Files:** `app/repository/user_repository.go`, `student_repository.go`

- Sentinel errors `ErrNotFound` (`pgx.ErrNoRows`), `ErrDuplicate` (code `23505`) in `student_repository.go`
- `UserRepository`: `FindByEmail` (LEFT JOIN students; exclude soft-deleted mahasiswa), `FindByID`
- `StudentRepository`: `FindAll` (filters `prodi`/`angkatan`, `search` ILIKE on `nim` OR `nama`, **whitelisted** ORDER BY, LIMIT/OFFSET, total count), `FindByID` and `FindByUserID` (both `deleted_at IS NULL`), `CreateWithUser` (users + students in one tx), `Update`, `SoftDelete`
- All values via bind parameters

**Verify:** `go build ./...`, then quick check with the temporary main: `FindByEmail("admin@siakad.test")` and `FindAll` return seed data.

---

#### Step 12 — `feat(repository): add course repository`

**Files:** `app/repository/course_repository.go`

- `FindAllWithQuota`: `LEFT JOIN enrollments` + `GROUP BY`; `terisi = COUNT(e.id)`, `sisa_kuota = kuota - COUNT(e.id)`; filters `semester`, `search` (`kode_mk` OR `nama_mk`), `available=true` -> `HAVING COUNT(e.id) < kuota`
- `FindByID`

**Verify:** `go build ./...`; temporary main lists 10 courses with quota.

---

#### Step 13 — `feat(repository): add enrollment repository with transaction`

**Files:** `app/repository/enrollment_repository.go`

- `EnrollmentRepository`: `FindByID`, `Delete`, `ListByStudent`, `TotalSKSByStudent`, `WithinTx(ctx, fn func(EnrollmentTx) error) error` (begin / commit / rollback)
- `EnrollmentTx`: `LockStudent` and `LockCourse` (`SELECT ... FOR UPDATE`), `Exists`, `CountByCourse`, `SumSKS(student, tahun)`, `Create` (maps `23505` -> `ErrDuplicate`)

**Verify:** `go build ./...`

---

### Phase E — Use Cases (Services)

> Each service is a constructor plus its Fiber handler methods. They can only be exercised over HTTP after Step 20, so for now verification is compile-time plus the rule tests.

#### Step 14 — `feat(auth): add AuthService (login, me)`

**Files:** `app/service/auth_service.go`

- `NewAuthService(userRepo, studentRepo, jwt)`
- `Login`: parse + validate -> `FindByEmail` -> `VerifyPassword` (dummy compare if not found) -> 401 on failure -> `LoginResponse{access_token, token_type: "Bearer", expires_in, user}`
- `Me`: current user; if `mahasiswa`, attach `student{nim,nama,prodi,angkatan}`; 401 if the student is soft-deleted

**Verify:** `go build ./... && go test ./...`

---

#### Step 15 — `feat(student): add StudentService`

**Files:** `app/service/student_service.go`

- `NewStudentService(studentRepo, userRepo, enrollmentRepo)`
- `List`: parse query -> `ValidateSort` -> repo -> `SuccessList` with `Meta` (`CountLastPage`)
- `Create`: validate -> `HashPassword(nim)` -> `CreateWithUser`; `ErrDuplicate` -> 422 on `nim` / `email`; respond 201
- `Get`: 404 if missing -> `CanAccessStudent` (403) -> attach enrollments, `total_sks`, `batas_sks` (`MaxSKSByIPK`)
- `Update`: validate -> `ApplyUpdate` -> repo (404 / 422)
- `Delete`: `SoftDelete` -> 204

**Verify:** `go build ./... && go test ./...`

---

#### Step 16 — `feat(course): add CourseService`

**Files:** `app/service/course_service.go`

- `NewCourseService(courseRepo)`
- `List`: parse `semester`, `search`, `available` -> `FindAllWithQuota` -> `Success`

**Verify:** `go build ./... && go test ./...`

---

#### Step 17 — `feat(enrollment): add EnrollmentService`

**Files:** `app/service/enrollment_service.go`

- `NewEnrollmentService(enrollmentRepo, studentRepo, courseRepo)`
- `Create`: validate body (`ValidateAcademicYear`) -> student from JWT user -> `WithinTx`: `LockStudent` -> `LockCourse` (missing -> 422) -> `Exists` (409) -> `IsCourseFull` (422) -> `CheckSKSLimit` (422 with remaining SKS) -> `Create` (`ErrDuplicate` -> 409) -> 201
- `Delete`: `FindByID` (404) -> `enrollment.student_id` must equal the caller's student id (403) -> `Delete` -> 204
- `student_id` always comes from the JWT, never from the request body

**Verify:** `go build ./... && go test ./...`

---

### Phase F — Interface Adapters (Middleware & Routes)

#### Step 18 — `feat(middleware): add global middleware`

**Files:** `middleware/middleware.go`

- `Register(app, logger, origins)`: `requestid`, `recover`, `helmet`, CORS (`corsPolicy`), `RequestLogger`
- `RequireJSON`: rejects non-`application/json` on POST/PUT

**Verify:** `go build ./...`

---

#### Step 19 — `feat(middleware): add auth, rate limiter, role guard`

**Files:** `middleware/auth.go`, `middleware/authz.go`

- `RequireAuth(jwt)`: Bearer token -> `Parse` -> store `AuthUser` in locals; 401 on missing/invalid/expired
- `LoginRateLimiter()`: Fiber `limiter`, max 5 per minute per IP, `SkipSuccessfulRequests`, 429 via `TooManyRequests`
- `RequireRole(roles...)`: 403 if the role is not allowed

**Verify:** `go build ./...`

---

#### Step 20 — `feat(route): register all 10 endpoints`

**Files:** `route/route.go`

- `Dependencies` struct (pool, JWT, 4 services)
- `Register(app, deps)` under `/api/v1`, with the middleware chains from the *API Routes* tables in `PROJECTPLAN.md`:
  - Public: `POST /auth/login` (RequireJSON, LoginRateLimiter)
  - Authenticated: `GET /auth/me`, `GET /courses`, `GET /students/:id`
  - Admin: `GET`/`POST /students`, `PUT`/`DELETE /students/:id`
  - Mahasiswa: `POST /enrollments`, `DELETE /enrollments/:id`

**Verify:** `go build ./...`

---

### Phase G — Bootstrap (outermost)

#### Step 21 — `feat(config): add fiber app and error handler`

**Files:** `config/app.go`

- `NewApp(logger, deps)`: create `fiber.App`, call `middleware.Register`, call `route.Register`
- Global `ErrorHandler`: `AppError` -> its status + `ErrorResponse{success:false, message, errors}`; unknown errors -> 500 with a generic message (`APP_ENV=production` hides the cause and never leaks a stack trace; the real error goes to the log)
- Unknown routes -> 404 in the same envelope

**Verify:** `go build ./...`

---

#### Step 22 — `feat: wire main and graceful shutdown`

**Files:** `main.go`

- Follow the *Startup Sequence* in `PROJECTPLAN.md`: `LoadEnv` -> `NewLogger` -> `NewPool` -> repositories -> `JWTManager` -> services -> `NewApp` -> `Listen` in a goroutine -> wait for SIGINT/SIGTERM -> `ShutdownWithContext(10s)` -> `pool.Close()`
- Remove the temporary verification code from earlier steps

**Verify:** `go run .` and run the [smoke test](#end-to-end-smoke-test) below.

---

#### Step 23 — `docs: add README and API examples`

**Files:** `README.md`

- Setup: install Go and PostgreSQL, create the database, run migrations (including `006_seed.sql`), copy `.env.example` to `.env`, `go run .`
- Seed accounts: admin credentials and the NIM-as-password rule for mahasiswa
- One `curl` example per endpoint
- Business rules and assumptions (copy from `PROJECTPLAN.md`)

---

## End-to-End Smoke Test

Run after Step 22 (replace `$TOKEN` with the token from login):

```bash
BASE=http://localhost:3000/api/v1

# 1. Login as admin
curl -s -X POST $BASE/auth/login -H "Content-Type: application/json" \
  -d '{"email":"admin@siakad.test","password":"Admin12345"}'

# 2. Me
curl -s $BASE/auth/me -H "Authorization: Bearer $TOKEN"

# 3. List students (pagination, filter, search, sort)
curl -s "$BASE/students?page=1&per_page=5&search=Rina&sort=-ipk_terakhir" -H "Authorization: Bearer $TOKEN"

# 4. Create student
curl -s -X POST $BASE/students -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
  -d '{"nim":"187221000099","nama":"Test","email":"test@siakad.test","prodi":"Sistem Informasi","angkatan":2024}'

# 8. Courses
curl -s "$BASE/courses?available=true" -H "Authorization: Bearer $TOKEN"
```

Then log in as a mahasiswa (`mahasiswa01@siakad.test`, password = their NIM) and check endpoints 5, 9, 10.

| Check | Expected |
|---|---|
| Wrong password | 401 |
| 6th failed login within a minute | 429 |
| No token on any protected endpoint | 401 |
| Mahasiswa calls `GET /students` | 403 |
| Mahasiswa opens another student's detail | 403 |
| Admin calls `POST /enrollments` | 403 |
| Duplicate NIM on create | 422 |
| Enroll the same course twice in the same `tahun_akademik` | 409 |
| Enroll in a full course | 422 |
| Exceed the SKS cap | 422, message includes remaining SKS |
| Cancel own enrollment | 204, `sisa_kuota` increases |
| Cancel someone else's enrollment | 403 |
| Delete student, then list / detail / login | gone from list, 404, login fails |
package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"siakad-mini/app/model"
)

var (
	ErrNotFound  = errors.New("data tidak ditemukan")
	ErrDuplicate = errors.New("data sudah ada")
)

// OrderByMap adalah whitelist urutan yang diizinkan — tidak pernah diinterpolasi langsung.
var orderByMap = map[string]string{
	"nama":          "nama ASC",
	"-ipk_terakhir": "ipk_terakhir DESC",
}

type StudentRepository interface {
	FindAll(ctx context.Context, q model.StudentListQuery) ([]model.Student, int, error)
	FindByID(ctx context.Context, id int) (model.Student, error)
	FindByUserID(ctx context.Context, userID int) (model.Student, error)
	CreateWithUser(ctx context.Context, req model.CreateStudentRequest, hashedPassword string) (model.Student, error)
	Update(ctx context.Context, s model.Student) (model.Student, error)
	SoftDelete(ctx context.Context, id int) error
}

type studentPostgresRepository struct {
	pool *pgxpool.Pool
}

func NewStudentRepository(pool *pgxpool.Pool) StudentRepository {
	return &studentPostgresRepository{pool: pool}
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}

// FindAll mengambil daftar mahasiswa dengan filter, pencarian ILIKE, urutan whitelisted, dan paginasi.
func (r *studentPostgresRepository) FindAll(
	ctx context.Context, q model.StudentListQuery,
) ([]model.Student, int, error) {
	args := []any{}
	where := " WHERE s.deleted_at IS NULL"

	if q.Prodi != "" {
		args = append(args, q.Prodi)
		where += fmt.Sprintf(" AND s.prodi = $%d", len(args))
	}
	if q.Angkatan > 0 {
		args = append(args, q.Angkatan)
		where += fmt.Sprintf(" AND s.angkatan = $%d", len(args))
	}
	if q.Search != "" {
		args = append(args, "%"+q.Search+"%")
		where += fmt.Sprintf(" AND (s.nim ILIKE $%d OR s.nama ILIKE $%d)", len(args), len(args))
	}

	var total int
	err := r.pool.QueryRow(ctx,
		"SELECT COUNT(*) FROM students s"+where, args...,
	).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("menghitung mahasiswa: %w", err)
	}

	orderBy, ok := orderByMap[q.Sort]
	if !ok {
		orderBy = "nama ASC"
	}

	args = append(args, q.PerPage, q.Offset())
	sqlText := fmt.Sprintf(
		`SELECT s.id, s.user_id, s.nim, s.nama, s.prodi, s.angkatan, s.ipk_terakhir, s.deleted_at
		 FROM students s%s
		 ORDER BY %s
		 LIMIT $%d OFFSET $%d`,
		where, orderBy, len(args)-1, len(args),
	)

	rows, err := r.pool.Query(ctx, sqlText, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("mengambil daftar mahasiswa: %w", err)
	}
	defer rows.Close()

	hasil := []model.Student{}
	for rows.Next() {
		var s model.Student
		if err := rows.Scan(
			&s.ID, &s.UserID, &s.NIM, &s.Nama, &s.Prodi,
			&s.Angkatan, &s.IPKTerakhir, &s.DeletedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("membaca baris mahasiswa: %w", err)
		}
		hasil = append(hasil, s)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("membaca hasil query: %w", err)
	}

	return hasil, total, nil
}

// FindByID mengambil mahasiswa berdasarkan ID (deleted_at IS NULL).
func (r *studentPostgresRepository) FindByID(
	ctx context.Context, id int,
) (model.Student, error) {
	var s model.Student
	err := r.pool.QueryRow(ctx,
		`SELECT id, user_id, nim, nama, prodi, angkatan, ipk_terakhir, deleted_at
		 FROM students
		 WHERE id = $1 AND deleted_at IS NULL`, id,
	).Scan(&s.ID, &s.UserID, &s.NIM, &s.Nama, &s.Prodi, &s.Angkatan, &s.IPKTerakhir, &s.DeletedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Student{}, ErrNotFound
		}
		return model.Student{}, fmt.Errorf("mengambil mahasiswa by id: %w", err)
	}
	return s, nil
}

// FindByUserID mengambil mahasiswa berdasarkan user_id (deleted_at IS NULL).
func (r *studentPostgresRepository) FindByUserID(
	ctx context.Context, userID int,
) (model.Student, error) {
	var s model.Student
	err := r.pool.QueryRow(ctx,
		`SELECT id, user_id, nim, nama, prodi, angkatan, ipk_terakhir, deleted_at
		 FROM students
		 WHERE user_id = $1 AND deleted_at IS NULL`, userID,
	).Scan(&s.ID, &s.UserID, &s.NIM, &s.Nama, &s.Prodi, &s.Angkatan, &s.IPKTerakhir, &s.DeletedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Student{}, ErrNotFound
		}
		return model.Student{}, fmt.Errorf("mengambil mahasiswa by user_id: %w", err)
	}
	return s, nil
}

// CreateWithUser menyisipkan user dan student dalam satu transaksi.
func (r *studentPostgresRepository) CreateWithUser(
	ctx context.Context, req model.CreateStudentRequest, hashedPassword string,
) (model.Student, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return model.Student{}, fmt.Errorf("memulai transaksi: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	var userID int
	err = tx.QueryRow(ctx,
		`INSERT INTO users (email, password, role)
		 VALUES ($1, $2, 'mahasiswa')
		 RETURNING id`,
		req.Email, hashedPassword,
	).Scan(&userID)
	if err != nil {
		if isUniqueViolation(err) {
			return model.Student{}, ErrDuplicate
		}
		return model.Student{}, fmt.Errorf("menyimpan user: %w", err)
	}

	ipk := 0.0
	if req.IPKTerakhir != nil {
		ipk = *req.IPKTerakhir
	}

	var s model.Student
	err = tx.QueryRow(ctx,
		`INSERT INTO students (user_id, nim, nama, prodi, angkatan, ipk_terakhir)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id, user_id, nim, nama, prodi, angkatan, ipk_terakhir, deleted_at`,
		userID, req.NIM, req.Nama, req.Prodi, req.Angkatan, ipk,
	).Scan(&s.ID, &s.UserID, &s.NIM, &s.Nama, &s.Prodi, &s.Angkatan, &s.IPKTerakhir, &s.DeletedAt)
	if err != nil {
		if isUniqueViolation(err) {
			return model.Student{}, ErrDuplicate
		}
		return model.Student{}, fmt.Errorf("menyimpan mahasiswa: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return model.Student{}, fmt.Errorf("commit transaksi: %w", err)
	}

	return s, nil
}

// Update memperbarui data mahasiswa.
func (r *studentPostgresRepository) Update(
	ctx context.Context, s model.Student,
) (model.Student, error) {
	err := r.pool.QueryRow(ctx,
		`UPDATE students
		 SET nama = $1, prodi = $2, angkatan = $3, ipk_terakhir = $4
		 WHERE id = $5 AND deleted_at IS NULL
		 RETURNING id, user_id, nim, nama, prodi, angkatan, ipk_terakhir, deleted_at`,
		s.Nama, s.Prodi, s.Angkatan, s.IPKTerakhir, s.ID,
	).Scan(&s.ID, &s.UserID, &s.NIM, &s.Nama, &s.Prodi, &s.Angkatan, &s.IPKTerakhir, &s.DeletedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Student{}, ErrNotFound
		}
		if isUniqueViolation(err) {
			return model.Student{}, ErrDuplicate
		}
		return model.Student{}, fmt.Errorf("memperbarui mahasiswa: %w", err)
	}
	return s, nil
}

// SoftDelete menandai mahasiswa sebagai terhapus secara logis dengan mengisi deleted_at.
func (r *studentPostgresRepository) SoftDelete(ctx context.Context, id int) error {
	now := time.Now()
	tag, err := r.pool.Exec(ctx,
		`UPDATE students SET deleted_at = $1 WHERE id = $2 AND deleted_at IS NULL`,
		now, id,
	)
	if err != nil {
		return fmt.Errorf("soft delete mahasiswa: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

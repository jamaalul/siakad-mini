package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"siakad-mini/app/model"
)

// EnrollmentTx adalah antarmuka untuk operasi di dalam satu transaksi database.
// Semua method terikat ke pgx.Tx yang sama sehingga bersifat atomik.
type EnrollmentTx interface {
	// LockStudent mengunci baris student untuk mencegah race condition SKS.
	LockStudent(ctx context.Context, studentID int) (model.Student, error)
	// LockCourse mengunci baris course untuk mencegah race condition kuota.
	LockCourse(ctx context.Context, courseID int) (model.Course, error)
	// Exists memeriksa apakah enrollment sudah ada (duplicate).
	Exists(ctx context.Context, studentID, courseID int, tahunAkademik string) (bool, error)
	// CountByCourse menghitung jumlah peserta di satu mata kuliah.
	CountByCourse(ctx context.Context, courseID int) (int, error)
	// SumSKS menjumlahkan SKS yang diambil mahasiswa dalam tahun akademik tertentu.
	SumSKS(ctx context.Context, studentID int, tahunAkademik string) (int, error)
	// Create menyisipkan enrollment baru; 23505 -> ErrDuplicate.
	Create(ctx context.Context, e model.Enrollment) (model.Enrollment, error)
}

// EnrollmentRepository menyediakan operasi enrollment di luar dan di dalam transaksi.
type EnrollmentRepository interface {
	FindByID(ctx context.Context, id int) (model.Enrollment, error)
	Delete(ctx context.Context, id int) error
	ListByStudent(ctx context.Context, studentID int) ([]model.EnrollmentDetail, error)
	TotalSKSByStudent(ctx context.Context, studentID int, tahunAkademik string) (int, error)
	// WithinTx membuka transaksi baru, meneruskan EnrollmentTx ke fn,
	// lalu commit jika fn kembali nil, atau rollback jika ada error.
	WithinTx(ctx context.Context, fn func(EnrollmentTx) error) error
}

// -------------------------------------------------------------------
// Implementasi
// -------------------------------------------------------------------

type enrollmentPostgresRepository struct {
	pool *pgxpool.Pool
}

func NewEnrollmentRepository(pool *pgxpool.Pool) EnrollmentRepository {
	return &enrollmentPostgresRepository{pool: pool}
}

func (r *enrollmentPostgresRepository) FindByID(
	ctx context.Context, id int,
) (model.Enrollment, error) {
	var e model.Enrollment
	err := r.pool.QueryRow(ctx,
		`SELECT id, student_id, course_id, tahun_akademik, created_at
		 FROM enrollments WHERE id = $1`, id,
	).Scan(&e.ID, &e.StudentID, &e.CourseID, &e.TahunAkademik, &e.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Enrollment{}, ErrNotFound
		}
		return model.Enrollment{}, fmt.Errorf("mengambil enrollment by id: %w", err)
	}
	return e, nil
}

func (r *enrollmentPostgresRepository) Delete(ctx context.Context, id int) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM enrollments WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("menghapus enrollment: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// ListByStudent mengambil semua enrollment milik mahasiswa beserta info mata kuliah.
func (r *enrollmentPostgresRepository) ListByStudent(
	ctx context.Context, studentID int,
) ([]model.EnrollmentDetail, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT e.id, e.student_id, e.course_id,
		        c.kode_mk, c.nama_mk, c.sks, c.semester,
		        e.tahun_akademik, e.created_at
		 FROM enrollments e
		 JOIN courses c ON c.id = e.course_id
		 WHERE e.student_id = $1
		 ORDER BY e.created_at DESC`, studentID,
	)
	if err != nil {
		return nil, fmt.Errorf("mengambil enrollment mahasiswa: %w", err)
	}
	defer rows.Close()

	hasil := []model.EnrollmentDetail{}
	for rows.Next() {
		var d model.EnrollmentDetail
		if err := rows.Scan(
			&d.ID, &d.StudentID, &d.CourseID,
			&d.KodeMK, &d.NamaMK, &d.SKS, &d.Semester,
			&d.TahunAkademik, &d.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("membaca baris enrollment: %w", err)
		}
		hasil = append(hasil, d)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("membaca hasil query: %w", err)
	}
	return hasil, nil
}

// TotalSKSByStudent menjumlahkan SKS yang diambil mahasiswa pada tahun akademik tertentu.
func (r *enrollmentPostgresRepository) TotalSKSByStudent(
	ctx context.Context, studentID int, tahunAkademik string,
) (int, error) {
	var total int
	err := r.pool.QueryRow(ctx,
		`SELECT COALESCE(SUM(c.sks), 0)
		 FROM enrollments e
		 JOIN courses c ON c.id = e.course_id
		 WHERE e.student_id = $1 AND e.tahun_akademik = $2`,
		studentID, tahunAkademik,
	).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("menghitung total SKS: %w", err)
	}
	return total, nil
}

// WithinTx membuka transaksi, meneruskan enrollmentTx ke fn, lalu commit atau rollback.
func (r *enrollmentPostgresRepository) WithinTx(
	ctx context.Context, fn func(EnrollmentTx) error,
) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("memulai transaksi enrollment: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	etx := &enrollmentTx{tx: tx}
	if err := fn(etx); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaksi enrollment: %w", err)
	}
	return nil
}

// -------------------------------------------------------------------
// EnrollmentTx implementation (bound to one pgx.Tx)
// -------------------------------------------------------------------

type enrollmentTx struct {
	tx pgx.Tx
}

func (t *enrollmentTx) LockStudent(
	ctx context.Context, studentID int,
) (model.Student, error) {
	var s model.Student
	err := t.tx.QueryRow(ctx,
		`SELECT id, user_id, nim, nama, prodi, angkatan, ipk_terakhir, deleted_at
		 FROM students
		 WHERE id = $1 AND deleted_at IS NULL
		 FOR UPDATE`, studentID,
	).Scan(&s.ID, &s.UserID, &s.NIM, &s.Nama, &s.Prodi, &s.Angkatan, &s.IPKTerakhir, &s.DeletedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Student{}, ErrNotFound
		}
		return model.Student{}, fmt.Errorf("lock student: %w", err)
	}
	return s, nil
}

func (t *enrollmentTx) LockCourse(
	ctx context.Context, courseID int,
) (model.Course, error) {
	var c model.Course
	err := t.tx.QueryRow(ctx,
		`SELECT id, kode_mk, nama_mk, sks, semester, kuota
		 FROM courses WHERE id = $1
		 FOR UPDATE`, courseID,
	).Scan(&c.ID, &c.KodeMK, &c.NamaMK, &c.SKS, &c.Semester, &c.Kuota)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Course{}, ErrNotFound
		}
		return model.Course{}, fmt.Errorf("lock course: %w", err)
	}
	return c, nil
}

func (t *enrollmentTx) Exists(
	ctx context.Context, studentID, courseID int, tahunAkademik string,
) (bool, error) {
	var count int
	err := t.tx.QueryRow(ctx,
		`SELECT COUNT(*) FROM enrollments
		 WHERE student_id = $1 AND course_id = $2 AND tahun_akademik = $3`,
		studentID, courseID, tahunAkademik,
	).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("memeriksa enrollment: %w", err)
	}
	return count > 0, nil
}

func (t *enrollmentTx) CountByCourse(
	ctx context.Context, courseID int,
) (int, error) {
	var count int
	err := t.tx.QueryRow(ctx,
		`SELECT COUNT(*) FROM enrollments WHERE course_id = $1`, courseID,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("menghitung peserta mata kuliah: %w", err)
	}
	return count, nil
}

func (t *enrollmentTx) SumSKS(
	ctx context.Context, studentID int, tahunAkademik string,
) (int, error) {
	var total int
	err := t.tx.QueryRow(ctx,
		`SELECT COALESCE(SUM(c.sks), 0)
		 FROM enrollments e
		 JOIN courses c ON c.id = e.course_id
		 WHERE e.student_id = $1 AND e.tahun_akademik = $2`,
		studentID, tahunAkademik,
	).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("menghitung SKS dalam transaksi: %w", err)
	}
	return total, nil
}

func (t *enrollmentTx) Create(
	ctx context.Context, e model.Enrollment,
) (model.Enrollment, error) {
	err := t.tx.QueryRow(ctx,
		`INSERT INTO enrollments (student_id, course_id, tahun_akademik)
		 VALUES ($1, $2, $3)
		 RETURNING id, created_at`,
		e.StudentID, e.CourseID, e.TahunAkademik,
	).Scan(&e.ID, &e.CreatedAt)
	if err != nil {
		if isUniqueViolation(err) {
			return model.Enrollment{}, ErrDuplicate
		}
		return model.Enrollment{}, fmt.Errorf("menyimpan enrollment: %w", err)
	}
	return e, nil
}

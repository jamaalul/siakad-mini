package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"siakad-mini/app/model"
)

type CourseRepository interface {
	FindAllWithQuota(ctx context.Context, q model.CourseListQuery) ([]model.CourseWithQuota, error)
	FindByID(ctx context.Context, id int) (model.Course, error)
}

type coursePostgresRepository struct {
	pool *pgxpool.Pool
}

func NewCourseRepository(pool *pgxpool.Pool) CourseRepository {
	return &coursePostgresRepository{pool: pool}
}

// FindAllWithQuota mengambil daftar mata kuliah beserta kuota terisi dan sisa kuota.
// Menggunakan LEFT JOIN ke enrollments + GROUP BY untuk menghitung terisi secara real-time.
func (r *coursePostgresRepository) FindAllWithQuota(
	ctx context.Context, q model.CourseListQuery,
) ([]model.CourseWithQuota, error) {
	args := []any{}
	where := " WHERE 1 = 1"
	having := ""

	if q.Semester > 0 {
		args = append(args, q.Semester)
		where += fmt.Sprintf(" AND c.semester = $%d", len(args))
	}
	if q.Search != "" {
		args = append(args, "%"+q.Search+"%")
		where += fmt.Sprintf(" AND (c.kode_mk ILIKE $%d OR c.nama_mk ILIKE $%d)", len(args), len(args))
	}
	if q.Available != nil && *q.Available {
		having = " HAVING COUNT(e.id) < c.kuota"
	}

	sqlText := fmt.Sprintf(
		`SELECT c.id, c.kode_mk, c.nama_mk, c.sks, c.semester, c.kuota,
		        COUNT(e.id) AS terisi,
		        c.kuota - COUNT(e.id) AS sisa_kuota
		 FROM courses c
		 LEFT JOIN enrollments e ON e.course_id = c.id
		 %s
		 GROUP BY c.id%s
		 ORDER BY c.semester, c.kode_mk`,
		where, having,
	)

	rows, err := r.pool.Query(ctx, sqlText, args...)
	if err != nil {
		return nil, fmt.Errorf("mengambil daftar mata kuliah: %w", err)
	}
	defer rows.Close()

	hasil := []model.CourseWithQuota{}
	for rows.Next() {
		var c model.CourseWithQuota
		if err := rows.Scan(
			&c.ID, &c.KodeMK, &c.NamaMK, &c.SKS, &c.Semester, &c.Kuota,
			&c.Terisi, &c.SisaKuota,
		); err != nil {
			return nil, fmt.Errorf("membaca baris mata kuliah: %w", err)
		}
		hasil = append(hasil, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("membaca hasil query: %w", err)
	}

	return hasil, nil
}

// FindByID mengambil mata kuliah berdasarkan ID.
func (r *coursePostgresRepository) FindByID(
	ctx context.Context, id int,
) (model.Course, error) {
	var c model.Course
	err := r.pool.QueryRow(ctx,
		`SELECT id, kode_mk, nama_mk, sks, semester, kuota
		 FROM courses WHERE id = $1`, id,
	).Scan(&c.ID, &c.KodeMK, &c.NamaMK, &c.SKS, &c.Semester, &c.Kuota)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Course{}, ErrNotFound
		}
		return model.Course{}, fmt.Errorf("mengambil mata kuliah by id: %w", err)
	}
	return c, nil
}

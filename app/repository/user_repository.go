package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"siakad-mini/app/model"
)

type UserRepository interface {
	FindByEmail(ctx context.Context, email string) (model.User, error)
	FindByID(ctx context.Context, id int) (model.User, error)
}

type userPostgresRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) UserRepository {
	return &userPostgresRepository{pool: pool}
}

// FindByEmail mencari user berdasarkan email.
// Untuk role mahasiswa, mengecualikan yang sudah soft-deleted (deleted_at IS NOT NULL).
func (r *userPostgresRepository) FindByEmail(
	ctx context.Context, email string,
) (model.User, error) {
	var u model.User
	err := r.pool.QueryRow(ctx,
		`SELECT u.id, u.email, u.password, u.role, u.created_at
		 FROM users u
		 LEFT JOIN students s ON s.user_id = u.id
		 WHERE LOWER(u.email) = LOWER($1)
		   AND (u.role != 'mahasiswa' OR s.deleted_at IS NULL)`,
		email,
	).Scan(&u.ID, &u.Email, &u.Password, &u.Role, &u.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.User{}, ErrNotFound
		}
		return model.User{}, fmt.Errorf("mengambil user by email: %w", err)
	}
	return u, nil
}

// FindByID mencari user berdasarkan ID.
func (r *userPostgresRepository) FindByID(
	ctx context.Context, id int,
) (model.User, error) {
	var u model.User
	err := r.pool.QueryRow(ctx,
		`SELECT id, email, password, role, created_at
		 FROM users WHERE id = $1`, id,
	).Scan(&u.ID, &u.Email, &u.Password, &u.Role, &u.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.User{}, ErrNotFound
		}
		return model.User{}, fmt.Errorf("mengambil user by id: %w", err)
	}
	return u, nil
}

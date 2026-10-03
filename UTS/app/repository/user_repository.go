package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"UTS/app/model"
)

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	query := `SELECT id, email, password, role, created_at, updated_at FROM users WHERE email = $1`
	var u model.User
	err := r.pool.QueryRow(ctx, query, email).Scan(&u.ID, &u.Email, &u.Password, &u.Role, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

func (r *UserRepository) FindByID(ctx context.Context, id int) (*model.User, error) {
	query := `SELECT id, email, password, role, created_at, updated_at FROM users WHERE id = $1`
	var u model.User
	err := r.pool.QueryRow(ctx, query, id).Scan(&u.ID, &u.Email, &u.Password, &u.Role, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

func (r *UserRepository) Create(ctx context.Context, tx pgx.Tx, email, passwordHash, role string) (int, error) {
	query := `INSERT INTO users (email, password, role) VALUES ($1, $2, $3) RETURNING id`
	var id int
	err := tx.QueryRow(ctx, query, email, passwordHash, role).Scan(&id)
	return id, err
}

package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"peminjaman-buku/app/model"
)

type UserRepository interface {
	Create(ctx context.Context, user model.User) (model.User, error)
	FindByID(ctx context.Context, id int) (model.User, error)
	FindByUsername(ctx context.Context, username string) (model.User, error)
}

const userColumns = "id, username, email, password, role, created_at"

type userPostgresRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) UserRepository {
	return &userPostgresRepository{pool: pool}
}

func scanUser(row pgx.Row) (model.User, error) {
	var user model.User
	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.Password, &user.Role, &user.CreatedAt)
	return user, err
}

func (r *userPostgresRepository) Create(ctx context.Context, user model.User) (model.User, error) {
	created, err := scanUser(r.pool.QueryRow(ctx,
		"INSERT INTO users (username, email, password, role) VALUES ($1, $2, $3, $4) RETURNING "+userColumns,
		user.Username, user.Email, user.Password, user.Role,
	))
	if err != nil {
		return model.User{}, translatePgError(err, "menyimpan user")
	}
	return created, nil
}

func (r *userPostgresRepository) FindByID(ctx context.Context, id int) (model.User, error) {
	user, err := scanUser(r.pool.QueryRow(ctx,
		"SELECT "+userColumns+" FROM users WHERE id = $1", id,
	))
	if err != nil {
		return model.User{}, translatePgError(err, "mengambil user")
	}
	return user, nil
}

func (r *userPostgresRepository) FindByUsername(ctx context.Context, username string) (model.User, error) {
	user, err := scanUser(r.pool.QueryRow(ctx,
		"SELECT "+userColumns+" FROM users WHERE LOWER(username) = LOWER($1)", username,
	))
	if err != nil {
		return model.User{}, translatePgError(err, "mengambil user")
	}
	return user, nil
}
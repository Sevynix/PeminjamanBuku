package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"peminjaman-buku/app/model"
)

type UserRepository interface {
	FindAll(ctx context.Context, q model.UserListQuery) ([]model.User, int, error)
	FindByID(ctx context.Context, id int) (model.User, error)
	FindByUsername(ctx context.Context, username string) (model.User, error)
	Create(ctx context.Context, user model.User) (model.User, error)
	Update(ctx context.Context, user model.User) (model.User, error)
	UpdateRole(ctx context.Context, id int, role string) (model.User, error)
	Delete(ctx context.Context, id int) error
}

const userColumns = "id, username, email, password, role, created_at"

var userSortColumns = map[string]string{
	"id":         "id",
	"username":   "username",
	"email":      "email",
	"created_at": "created_at",
}

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

func buildUserFilter(q model.UserListQuery) (string, []any) {
	conditions := []string{}
	args := []any{}

	if q.Search != "" {
		args = append(args, "%"+likeEscaper.Replace(q.Search)+"%")
		conditions = append(conditions,
			fmt.Sprintf("(username ILIKE $%d OR email ILIKE $%d)", len(args), len(args)))
	}

	if q.Role != "" {
		args = append(args, q.Role)
		conditions = append(conditions, fmt.Sprintf("role = $%d", len(args)))
	}

	if len(conditions) == 0 {
		return "", args
	}
	return " WHERE " + strings.Join(conditions, " AND "), args
}

func (r *userPostgresRepository) FindAll(
	ctx context.Context, q model.UserListQuery,
) ([]model.User, int, error) {
	where, args := buildUserFilter(q)

	var total int
	err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM users"+where, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("menghitung user: %w", err)
	}

	column, found := userSortColumns[q.Sort]
	if !found {
		column = "id"
	}
	direction := "ASC"
	if q.Order == "desc" {
		direction = "DESC"
	}
	orderBy := column + " " + direction
	if column != "id" {
		orderBy += ", id " + direction
	}

	query := fmt.Sprintf(
		"SELECT %s FROM users%s ORDER BY %s LIMIT $%d OFFSET $%d",
		userColumns, where, orderBy, len(args)+1, len(args)+2,
	)
	args = append(args, q.Limit, q.Offset())

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("mengambil daftar user: %w", err)
	}
	defer rows.Close()

	users := []model.User{}
	for rows.Next() {
		user, err := scanUser(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("membaca baris user: %w", err)
		}
		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("membaca hasil query user: %w", err)
	}

	return users, total, nil
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

func (r *userPostgresRepository) Update(ctx context.Context, user model.User) (model.User, error) {
	updated, err := scanUser(r.pool.QueryRow(ctx,
		"UPDATE users SET username = $1, email = $2 WHERE id = $3 RETURNING "+userColumns,
		user.Username, user.Email, user.ID,
	))
	if err != nil {
		return model.User{}, translatePgError(err, "memperbarui user")
	}
	return updated, nil
}

func (r *userPostgresRepository) UpdateRole(ctx context.Context, id int, role string) (model.User, error) {
	updated, err := scanUser(r.pool.QueryRow(ctx,
		"UPDATE users SET role = $1 WHERE id = $2 RETURNING "+userColumns,
		role, id,
	))
	if err != nil {
		return model.User{}, translatePgError(err, "mengubah role user")
	}
	return updated, nil
}

func (r *userPostgresRepository) Delete(ctx context.Context, id int) error {
	tag, err := r.pool.Exec(ctx, "DELETE FROM users WHERE id = $1", id)
	if err != nil {
		return translatePgError(err, "menghapus user")
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
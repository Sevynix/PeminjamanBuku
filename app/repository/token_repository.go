package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type TokenRepository interface {
	Save(ctx context.Context, userID int, tokenHash string, expiresAt time.Time) error
	Consume(ctx context.Context, tokenHash string) (int, error)
	Revoke(ctx context.Context, tokenHash string) error
}

type tokenPostgresRepository struct {
	pool *pgxpool.Pool
}

func NewTokenRepository(pool *pgxpool.Pool) TokenRepository {
	return &tokenPostgresRepository{pool: pool}
}

func (r *tokenPostgresRepository) Save(
	ctx context.Context, userID int, tokenHash string, expiresAt time.Time,
) error {
	_, err := r.pool.Exec(ctx,
		"INSERT INTO refresh_tokens (user_id, token_hash, expires_at) VALUES ($1, $2, $3)",
		userID, tokenHash, expiresAt,
	)
	if err != nil {
		return fmt.Errorf("menyimpan refresh token: %w", err)
	}
	return nil
}

func (r *tokenPostgresRepository) Consume(ctx context.Context, tokenHash string) (int, error) {
	var userID int
	err := r.pool.QueryRow(ctx,
		`UPDATE refresh_tokens SET revoked_at = NOW()
		 WHERE token_hash = $1 AND revoked_at IS NULL AND expires_at > NOW()
		 RETURNING user_id`, tokenHash,
	).Scan(&userID)
	if err != nil {
		return 0, translatePgError(err, "memakai refresh token")
	}
	return userID, nil
}

func (r *tokenPostgresRepository) Revoke(ctx context.Context, tokenHash string) error {
	_, err := r.pool.Exec(ctx,
		"UPDATE refresh_tokens SET revoked_at = NOW() WHERE token_hash = $1 AND revoked_at IS NULL",
		tokenHash,
	)
	if err != nil {
		return fmt.Errorf("mencabut refresh token: %w", err)
	}
	return nil
}s
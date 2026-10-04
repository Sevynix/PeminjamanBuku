package repository

import (
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrNotFound        = errors.New("data tidak ditemukan")
	ErrDuplicate       = errors.New("data sudah ada")
	ErrConflict        = errors.New("data masih direferensikan")
	ErrCheckViolation  = errors.New("data melanggar batasan database")
	ErrOutOfStock      = errors.New("stok buku habis")
	ErrLimitReached    = errors.New("batas pinjaman aktif tercapai")
	ErrAlreadyReturned = errors.New("pinjaman sudah dikembalikan")
)

func translatePgError(err error, action string) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			return ErrDuplicate
		case "23503":
			return ErrConflict
		case "23514":
			return ErrCheckViolation
		}
	}

	return fmt.Errorf("%s: %w", action, err)
}
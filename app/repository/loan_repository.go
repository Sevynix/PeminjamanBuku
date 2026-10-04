package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"peminjaman-buku/app/model"
)

type LoanRepository interface {
	Borrow(ctx context.Context, userID, bookID int, dueAt time.Time, maxActive int) (model.Loan, error)
	Return(ctx context.Context, id int) (model.Loan, error)
	FindByID(ctx context.Context, id int) (model.Loan, error)
	FindPage(ctx context.Context, q model.LoanPageQuery) ([]model.Loan, error)
}

const loanSelect = `SELECT l.id, l.user_id, u.username, l.book_id, b.title, l.borrowed_at, l.due_at, l.returned_at
	FROM loans l
	JOIN users u ON u.id = l.user_id
	JOIN books b ON b.id = l.book_id`

var loanStatusConditions = map[string]string{
	"active":   "l.returned_at IS NULL",
	"returned": "l.returned_at IS NOT NULL",
	"overdue":  "l.returned_at IS NULL AND l.due_at < NOW()",
}

type rowQuerier interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type loanPostgresRepository struct {
	pool *pgxpool.Pool
}

func NewLoanRepository(pool *pgxpool.Pool) LoanRepository {
	return &loanPostgresRepository{pool: pool}
}

func scanLoan(row pgx.Row) (model.Loan, error) {
	var loan model.Loan
	err := row.Scan(
		&loan.ID, &loan.UserID, &loan.Username, &loan.BookID, &loan.BookTitle,
		&loan.BorrowedAt, &loan.DueAt, &loan.ReturnedAt,
	)
	return loan, err
}

func findLoan(ctx context.Context, q rowQuerier, id int) (model.Loan, error) {
	return scanLoan(q.QueryRow(ctx, loanSelect+" WHERE l.id = $1", id))
}

func (r *loanPostgresRepository) Borrow(
	ctx context.Context, userID, bookID int, dueAt time.Time, maxActive int,
) (model.Loan, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return model.Loan{}, fmt.Errorf("memulai transaksi peminjaman: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx, "SELECT 1 FROM users WHERE id = $1 FOR UPDATE", userID); err != nil {
		return model.Loan{}, fmt.Errorf("mengunci user: %w", err)
	}

	var active int
	err = tx.QueryRow(ctx,
		"SELECT COUNT(*) FROM loans WHERE user_id = $1 AND returned_at IS NULL", userID,
	).Scan(&active)
	if err != nil {
		return model.Loan{}, fmt.Errorf("menghitung pinjaman aktif: %w", err)
	}
	if active >= maxActive {
		return model.Loan{}, ErrLimitReached
	}

	tag, err := tx.Exec(ctx, "UPDATE books SET stock = stock - 1 WHERE id = $1 AND stock > 0", bookID)
	if err != nil {
		return model.Loan{}, translatePgError(err, "mengurangi stok buku")
	}
	if tag.RowsAffected() == 0 {
		var exists bool
		err := tx.QueryRow(ctx, "SELECT EXISTS (SELECT 1 FROM books WHERE id = $1)", bookID).Scan(&exists)
		if err != nil {
			return model.Loan{}, fmt.Errorf("memeriksa buku: %w", err)
		}
		if !exists {
			return model.Loan{}, ErrNotFound
		}
		return model.Loan{}, ErrOutOfStock
	}

	var loanID int
	err = tx.QueryRow(ctx,
		"INSERT INTO loans (user_id, book_id, due_at) VALUES ($1, $2, $3) RETURNING id",
		userID, bookID, dueAt,
	).Scan(&loanID)
	if err != nil {
		return model.Loan{}, translatePgError(err, "menyimpan pinjaman")
	}

	loan, err := findLoan(ctx, tx, loanID)
	if err != nil {
		return model.Loan{}, translatePgError(err, "mengambil pinjaman")
	}

	if err := tx.Commit(ctx); err != nil {
		return model.Loan{}, fmt.Errorf("menyimpan transaksi peminjaman: %w", err)
	}
	return loan, nil
}

func (r *loanPostgresRepository) Return(ctx context.Context, id int) (model.Loan, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return model.Loan{}, fmt.Errorf("memulai transaksi pengembalian: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var bookID int
	var returnedAt *time.Time
	err = tx.QueryRow(ctx,
		"SELECT book_id, returned_at FROM loans WHERE id = $1 FOR UPDATE", id,
	).Scan(&bookID, &returnedAt)
	if err != nil {
		return model.Loan{}, translatePgError(err, "mengambil pinjaman")
	}
	if returnedAt != nil {
		return model.Loan{}, ErrAlreadyReturned
	}

	if _, err := tx.Exec(ctx, "UPDATE loans SET returned_at = NOW() WHERE id = $1", id); err != nil {
		return model.Loan{}, translatePgError(err, "menandai pinjaman kembali")
	}
	if _, err := tx.Exec(ctx, "UPDATE books SET stock = stock + 1 WHERE id = $1", bookID); err != nil {
		return model.Loan{}, translatePgError(err, "menambah stok buku")
	}

	loan, err := findLoan(ctx, tx, id)
	if err != nil {
		return model.Loan{}, translatePgError(err, "mengambil pinjaman")
	}

	if err := tx.Commit(ctx); err != nil {
		return model.Loan{}, fmt.Errorf("menyimpan transaksi pengembalian: %w", err)
	}
	return loan, nil
}

func (r *loanPostgresRepository) FindByID(ctx context.Context, id int) (model.Loan, error) {
	loan, err := findLoan(ctx, r.pool, id)
	if err != nil {
		return model.Loan{}, translatePgError(err, "mengambil pinjaman")
	}
	return loan, nil
}

func (r *loanPostgresRepository) FindPage(ctx context.Context, q model.LoanPageQuery) ([]model.Loan, error) {
	conditions := []string{}
	args := []any{}

	if q.UserID > 0 {
		args = append(args, q.UserID)
		conditions = append(conditions, fmt.Sprintf("l.user_id = $%d", len(args)))
	}

	if condition, found := loanStatusConditions[q.Status]; found {
		conditions = append(conditions, condition)
	}

	if q.After != nil {
		args = append(args, q.After.BorrowedAt, q.After.ID)
		conditions = append(conditions,
			fmt.Sprintf("(l.borrowed_at, l.id) < ($%d, $%d)", len(args)-1, len(args)))
	}

	where := ""
	if len(conditions) > 0 {
		where = " WHERE " + strings.Join(conditions, " AND ")
	}

	args = append(args, q.Limit+1)
	query := fmt.Sprintf(
		"%s%s ORDER BY l.borrowed_at DESC, l.id DESC LIMIT $%d",
		loanSelect, where, len(args),
	)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("mengambil daftar pinjaman: %w", err)
	}
	defer rows.Close()

	loans := []model.Loan{}
	for rows.Next() {
		loan, err := scanLoan(rows)
		if err != nil {
			return nil, fmt.Errorf("membaca baris pinjaman: %w", err)
		}
		loans = append(loans, loan)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("membaca hasil query pinjaman: %w", err)
	}

	return loans, nil
}
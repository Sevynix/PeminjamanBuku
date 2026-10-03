package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"

	"peminjaman-buku/app/model"
)

type BookRepository interface {
	FindAll(ctx context.Context, q model.BookListQuery) ([]model.Book, int, error)
	FindByID(ctx context.Context, id int) (model.Book, error)
	Create(ctx context.Context, book model.Book) (model.Book, error)
	Update(ctx context.Context, book model.Book) (model.Book, error)
	Delete(ctx context.Context, id int) error
}

var bookSortColumns = map[string]string{
	"id":         "id",
	"title":      "title",
	"stock":      "stock",
	"created_at": "created_at",
}

var likeEscaper = strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)

type bookPostgresRepository struct {
	pool *pgxpool.Pool
}

func NewBookRepository(pool *pgxpool.Pool) BookRepository {
	return &bookPostgresRepository{pool: pool}
}

func buildBookFilter(q model.BookListQuery) (string, []any) {
	conditions := []string{}
	args := []any{}

	if q.Search != "" {
		args = append(args, "%"+likeEscaper.Replace(q.Search)+"%")
		conditions = append(conditions, fmt.Sprintf("title ILIKE $%d", len(args)))
	}

	if q.Available != nil {
		if *q.Available {
			conditions = append(conditions, "stock > 0")
		} else {
			conditions = append(conditions, "stock = 0")
		}
	}

	if len(conditions) == 0 {
		return "", args
	}
	return " WHERE " + strings.Join(conditions, " AND "), args
}

func (r *bookPostgresRepository) FindAll(
	ctx context.Context, q model.BookListQuery,
) ([]model.Book, int, error) {
	where, args := buildBookFilter(q)

	var total int
	err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM books"+where, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("menghitung buku: %w", err)
	}

	column, found := bookSortColumns[q.Sort]
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
		"SELECT id, title, stock, created_at FROM books%s ORDER BY %s LIMIT $%d OFFSET $%d",
		where, orderBy, len(args)+1, len(args)+2,
	)
	args = append(args, q.Limit, q.Offset())

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("mengambil daftar buku: %w", err)
	}
	defer rows.Close()

	books := []model.Book{}
	for rows.Next() {
		var book model.Book
		if err := rows.Scan(&book.ID, &book.Title, &book.Stock, &book.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("membaca baris buku: %w", err)
		}
		books = append(books, book)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("membaca hasil query buku: %w", err)
	}

	return books, total, nil
}

func (r *bookPostgresRepository) FindByID(ctx context.Context, id int) (model.Book, error) {
	var book model.Book
	err := r.pool.QueryRow(ctx,
		"SELECT id, title, stock, created_at FROM books WHERE id = $1", id,
	).Scan(&book.ID, &book.Title, &book.Stock, &book.CreatedAt)
	if err != nil {
		return model.Book{}, translatePgError(err, "mengambil buku")
	}
	return book, nil
}

func (r *bookPostgresRepository) Create(ctx context.Context, book model.Book) (model.Book, error) {
	err := r.pool.QueryRow(ctx,
		"INSERT INTO books (title, stock) VALUES ($1, $2) RETURNING id, title, stock, created_at",
		book.Title, book.Stock,
	).Scan(&book.ID, &book.Title, &book.Stock, &book.CreatedAt)
	if err != nil {
		return model.Book{}, translatePgError(err, "menyimpan buku")
	}
	return book, nil
}

func (r *bookPostgresRepository) Update(ctx context.Context, book model.Book) (model.Book, error) {
	err := r.pool.QueryRow(ctx,
		"UPDATE books SET title = $1, stock = $2 WHERE id = $3 RETURNING id, title, stock, created_at",
		book.Title, book.Stock, book.ID,
	).Scan(&book.ID, &book.Title, &book.Stock, &book.CreatedAt)
	if err != nil {
		return model.Book{}, translatePgError(err, "memperbarui buku")
	}
	return book, nil
}

func (r *bookPostgresRepository) Delete(ctx context.Context, id int) error {
	tag, err := r.pool.Exec(ctx, "DELETE FROM books WHERE id = $1", id)
	if err != nil {
		return translatePgError(err, "menghapus buku")
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
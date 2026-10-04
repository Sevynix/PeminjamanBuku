package model

import "time"

type Loan struct {
	ID         int        `json:"id"`
	UserID     int        `json:"user_id"`
	Username   string     `json:"username"`
	BookID     int        `json:"book_id"`
	BookTitle  string     `json:"book_title"`
	BorrowedAt time.Time  `json:"borrowed_at"`
	DueAt      time.Time  `json:"due_at"`
	ReturnedAt *time.Time `json:"returned_at"`
	Status     string     `json:"status"`
}

type BorrowRequest struct {
	BookID int `json:"book_id" validate:"required,min=1,max=2147483647"`
}

type Cursor struct {
	BorrowedAt time.Time
	ID         int
}

type LoanPageQuery struct {
	Limit  int
	After  *Cursor
	Status string
	UserID int
}
package model

import "time"

type Book struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Stock     int       `json:"stock"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateBookRequest struct {
	Title string `json:"title" validate:"required,max=200"`
	Stock *int   `json:"stock" validate:"required,min=0,max=100000"`
}

type ReplaceBookRequest struct {
	Title string `json:"title" validate:"required,max=200"`
	Stock *int   `json:"stock" validate:"required,min=0,max=100000"`
}

type PatchBookRequest struct {
	Title *string `json:"title,omitempty" validate:"omitnil,min=1,max=200"`
	Stock *int    `json:"stock,omitempty" validate:"omitnil,min=0,max=100000"`
}

type BookListQuery struct {
	ListQuery
	Available *bool
}
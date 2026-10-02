package repository

import "errors"

var (
	ErrNotFound       = errors.New("data tidak ditemukan")
	ErrDuplicate      = errors.New("data sudah ada")
	ErrConflict       = errors.New("data masih direferensikan")
	ErrCheckViolation = errors.New("data melanggar batasan database")
)
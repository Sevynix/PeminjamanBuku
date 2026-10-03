package service

import (
	"errors"

	"peminjaman-buku/app/repository"
	"peminjaman-buku/helper"
)

type errorMessages struct {
	NotFound  string
	Duplicate string
	Conflict  string
	Check     string
}

func translateError(err error, messages errorMessages, internalMessage string) error {
	switch {
	case errors.Is(err, repository.ErrNotFound):
		return helper.NotFound(messages.NotFound)
	case errors.Is(err, repository.ErrDuplicate):
		return helper.Conflict(messages.Duplicate).WithCause(err)
	case errors.Is(err, repository.ErrConflict):
		return helper.Conflict(messages.Conflict).WithCause(err)
	case errors.Is(err, repository.ErrCheckViolation):
		return helper.Validation(map[string]string{"data": messages.Check}).WithCause(err)
	default:
		return helper.Internal(internalMessage, err)
	}
}
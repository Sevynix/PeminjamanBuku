package helper

import (
	"errors"

	"github.com/gofiber/fiber/v2"
)

const (
	CodeValidation       = "VALIDATION_ERROR"
	CodeBadRequest       = "BAD_REQUEST"
	CodeUnauthorized     = "UNAUTHORIZED"
	CodeForbidden        = "FORBIDDEN"
	CodeNotFound         = "NOT_FOUND"
	CodeMethodNotAllowed = "METHOD_NOT_ALLOWED"
	CodeConflict         = "CONFLICT"
	CodeUnsupportedMedia = "UNSUPPORTED_MEDIA_TYPE"
	CodeNotAcceptable    = "NOT_ACCEPTABLE"
	CodeTooManyRequests  = "TOO_MANY_REQUESTS"
	CodeInternal         = "INTERNAL_ERROR"
)

type AppError struct {
	Status  int
	Code    string
	Message string
	Fields  map[string]string
	Headers map[string]string
	cause   error
}

func (e *AppError) Error() string {
	if e.cause != nil {
		return e.Message + ": " + e.cause.Error()
	}
	return e.Message
}

func (e *AppError) Unwrap() error { return e.cause }

func (e *AppError) WithCause(err error) *AppError {
	e.cause = err
	return e
}

func (e *AppError) WithHeader(key, value string) *AppError {
	if e.Headers == nil {
		e.Headers = map[string]string{}
	}
	e.Headers[key] = value
	return e
}

func newError(status int, code, message string) *AppError {
	return &AppError{Status: status, Code: code, Message: message}
}

func BadRequest(message string) *AppError {
	return newError(fiber.StatusBadRequest, CodeBadRequest, message)
}

func Unauthorized(message string) *AppError {
	return newError(fiber.StatusUnauthorized, CodeUnauthorized, message)
}

func Forbidden(message string) *AppError {
	return newError(fiber.StatusForbidden, CodeForbidden, message)
}

func NotFound(message string) *AppError {
	return newError(fiber.StatusNotFound, CodeNotFound, message)
}

func Conflict(message string) *AppError {
	return newError(fiber.StatusConflict, CodeConflict, message)
}

func UnsupportedMediaType(message string) *AppError {
	return newError(fiber.StatusUnsupportedMediaType, CodeUnsupportedMedia, message)
}

func NotAcceptable(message string) *AppError {
	return newError(fiber.StatusNotAcceptable, CodeNotAcceptable, message)
}

func TooManyRequests(message string) *AppError {
	return newError(fiber.StatusTooManyRequests, CodeTooManyRequests, message)
}

func Validation(fields map[string]string) *AppError {
	e := newError(fiber.StatusUnprocessableEntity, CodeValidation, "validasi gagal")
	e.Fields = fields
	return e
}

func Internal(message string, cause error) *AppError {
	return newError(fiber.StatusInternalServerError, CodeInternal, message).WithCause(cause)
}

func AsAppError(err error) *AppError {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr
	}

	var fiberErr *fiber.Error
	if errors.As(err, &fiberErr) {
		switch fiberErr.Code {
		case fiber.StatusNotFound:
			return NotFound("endpoint tidak ditemukan")
		case fiber.StatusMethodNotAllowed:
			return newError(fiber.StatusMethodNotAllowed, CodeMethodNotAllowed,
				"metode tidak didukung pada endpoint ini")
		}
		if fiberErr.Code < fiber.StatusInternalServerError {
			return newError(fiberErr.Code, CodeBadRequest, fiberErr.Message)
		}
	}

	return Internal("terjadi kesalahan pada server", err)
}
package service

import (
	"strings"

	"peminjaman-buku/app/model"
)

func NormalizeCreate(req model.CreateBookRequest) model.CreateBookRequest {
	req.Title = strings.TrimSpace(req.Title)
	return req
}

func NormalizeReplace(req model.ReplaceBookRequest) model.ReplaceBookRequest {
	req.Title = strings.TrimSpace(req.Title)
	return req
}

func NormalizePatch(req model.PatchBookRequest) model.PatchBookRequest {
	if req.Title != nil {
		trimmed := strings.TrimSpace(*req.Title)
		req.Title = &trimmed
	}
	return req
}

func IsEmptyPatch(req model.PatchBookRequest) bool {
	return req.Title == nil && req.Stock == nil
}

func ApplyPatch(current model.Book, req model.PatchBookRequest) model.Book {
	if req.Title != nil {
		current.Title = *req.Title
	}
	if req.Stock != nil {
		current.Stock = *req.Stock
	}
	return current
}

func CountTotalPages(total, limit int) int {
	if limit <= 0 {
		return 0
	}
	return (total + limit - 1) / limit
}
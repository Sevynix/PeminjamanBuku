package service

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"peminjaman-buku/app/model"
	"peminjaman-buku/app/repository"
	"peminjaman-buku/helper"
)

var loanMessages = errorMessages{
	NotFound:  "pinjaman tidak ditemukan",
	Duplicate: "buku ini masih Anda pinjam",
	Conflict:  "permintaan peminjaman tidak dapat diproses",
	Check:     "data pinjaman melanggar batasan",
}

type LoanService struct {
	repo  repository.LoanRepository
	perms *helper.PermissionSet
}

func NewLoanService(repo repository.LoanRepository, perms *helper.PermissionSet) *LoanService {
	return &LoanService{repo: repo, perms: perms}
}

func (s *LoanService) Borrow(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	current, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Unauthorized("belum terautentikasi")
	}

	var req model.BorrowRequest
	if err := helper.BindJSON(c, &req); err != nil {
		return err
	}
	if err := helper.Validate(&req); err != nil {
		return err
	}

	now := time.Now()
	loan, err := s.repo.Borrow(ctx, current.UserID, req.BookID, CalcDueDate(now), MaxActiveLoans)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrNotFound):
			return helper.NotFound("buku tidak ditemukan")
		case errors.Is(err, repository.ErrOutOfStock):
			return helper.Conflict("stok buku habis")
		case errors.Is(err, repository.ErrLimitReached):
			return helper.Validation(map[string]string{
				"book_id": fmt.Sprintf("batas %d pinjaman aktif sudah tercapai", MaxActiveLoans),
			})
		}
		return translateError(err, loanMessages, "gagal memproses peminjaman")
	}

	return helper.Created(c, "peminjaman berhasil", WithStatus(loan, now), "/api/v1/loans/"+strconv.Itoa(loan.ID))
}

func (s *LoanService) ListMine(c *fiber.Ctx) error {
	current, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Unauthorized("belum terautentikasi")
	}
	return s.listLoans(c, current.UserID)
}

func (s *LoanService) List(c *fiber.Ctx) error {
	return s.listLoans(c, 0)
}

func (s *LoanService) listLoans(c *fiber.Ctx, userID int) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	format, err := helper.Negotiate(c, helper.FormatJSON, helper.FormatCSV)
	if err != nil {
		return err
	}

	limit, after, err := helper.ParseCursorParams(c)
	if err != nil {
		return err
	}

	status := strings.ToLower(strings.TrimSpace(c.Query("status")))
	if status != "" && !IsValidLoanStatus(status) {
		return helper.BadRequest("status harus salah satu dari: active, returned, overdue")
	}

	loans, err := s.repo.FindPage(ctx, model.LoanPageQuery{
		Limit:  limit,
		After:  after,
		Status: status,
		UserID: userID,
	})
	if err != nil {
		return translateError(err, loanMessages, "gagal mengambil daftar pinjaman")
	}

	hasMore := len(loans) > limit
	if hasMore {
		loans = loans[:limit]
	}

	now := time.Now()
	for i := range loans {
		loans[i] = WithStatus(loans[i], now)
	}

	nextCursor := ""
	if hasMore && len(loans) > 0 {
		last := loans[len(loans)-1]
		nextCursor = helper.EncodeCursor(last.BorrowedAt, last.ID)
	}

	if format == helper.FormatCSV {
		if nextCursor != "" {
			c.Set("X-Next-Cursor", nextCursor)
		}
		return helper.WriteLoansCSV(c, loans)
	}

	return helper.SuccessCursor(c, "daftar pinjaman berhasil diambil", loans, model.CursorMeta{
		Limit:      limit,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	})
}

func (s *LoanService) Get(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	current, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Unauthorized("belum terautentikasi")
	}

	id, err := helper.ParamID(c)
	if err != nil {
		return err
	}

	loan, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return translateError(err, loanMessages, "gagal mengambil data pinjaman")
	}

	if !CanAccessLoan(current, loan.UserID, s.perms, "loan:read:any") {
		return helper.Forbidden("tidak berhak mengakses pinjaman milik user lain")
	}

	return helper.Success(c, fiber.StatusOK, "pinjaman ditemukan", WithStatus(loan, time.Now()))
}

func (s *LoanService) Return(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	id, err := helper.ParamID(c)
	if err != nil {
		return err
	}

	loan, err := s.repo.Return(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrAlreadyReturned) {
			return helper.Conflict("pinjaman sudah dikembalikan")
		}
		return translateError(err, loanMessages, "gagal memproses pengembalian")
	}

	return helper.Success(c, fiber.StatusOK, "buku berhasil dikembalikan", WithStatus(loan, time.Now()))
}
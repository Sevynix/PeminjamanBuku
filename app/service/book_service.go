package service

import (
	"strconv"

	"github.com/gofiber/fiber/v2"

	"peminjaman-buku/app/model"
	"peminjaman-buku/app/repository"
	"peminjaman-buku/helper"
)

var bookMessages = errorMessages{
	NotFound:  "buku tidak ditemukan",
	Duplicate: "judul buku sudah ada",
	Conflict:  "buku tidak dapat dihapus karena masih memiliki riwayat peminjaman",
	Check:     "data buku melanggar batasan",
}

var bookSortFields = map[string]bool{
	"id":         true,
	"title":      true,
	"stock":      true,
	"created_at": true,
}

type BookService struct {
	repo repository.BookRepository
}

func NewBookService(repo repository.BookRepository) *BookService {
	return &BookService{repo: repo}
}

func (s *BookService) List(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	available, err := helper.QueryBool(c, "available")
	if err != nil {
		return err
	}

	q := model.BookListQuery{
		ListQuery: helper.ParseListQuery(c, bookSortFields),
		Available: available,
	}

	books, total, err := s.repo.FindAll(ctx, q)
	if err != nil {
		return translateError(err, bookMessages, "gagal mengambil daftar buku")
	}

	return helper.SuccessList(c, "daftar buku berhasil diambil", books, model.Meta{
		Page:       q.Page,
		Limit:      q.Limit,
		Total:      total,
		TotalPages: CountTotalPages(total, q.Limit),
	})
}

func (s *BookService) Get(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	id, err := helper.ParamID(c)
	if err != nil {
		return err
	}

	book, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return translateError(err, bookMessages, "gagal mengambil data buku")
	}

	return helper.Success(c, fiber.StatusOK, "buku ditemukan", book)
}

func (s *BookService) Create(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	var req model.CreateBookRequest
	if err := helper.BindJSON(c, &req); err != nil {
		return err
	}
	req = NormalizeCreate(req)
	if err := helper.Validate(&req); err != nil {
		return err
	}

	book, err := s.repo.Create(ctx, model.Book{Title: req.Title, Stock: *req.Stock})
	if err != nil {
		return translateError(err, bookMessages, "gagal menyimpan buku")
	}

	return helper.Created(c, "buku berhasil dibuat", book, "/api/v1/books/"+strconv.Itoa(book.ID))
}

func (s *BookService) Replace(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	id, err := helper.ParamID(c)
	if err != nil {
		return err
	}

	var req model.ReplaceBookRequest
	if err := helper.BindJSON(c, &req); err != nil {
		return err
	}
	req = NormalizeReplace(req)
	if err := helper.Validate(&req); err != nil {
		return err
	}

	book, err := s.repo.Update(ctx, model.Book{ID: id, Title: req.Title, Stock: *req.Stock})
	if err != nil {
		return translateError(err, bookMessages, "gagal memperbarui buku")
	}

	return helper.Success(c, fiber.StatusOK, "buku berhasil diganti seluruhnya", book)
}

func (s *BookService) Patch(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	id, err := helper.ParamID(c)
	if err != nil {
		return err
	}

	var req model.PatchBookRequest
	if err := helper.BindJSON(c, &req); err != nil {
		return err
	}
	if IsEmptyPatch(req) {
		return helper.BadRequest("tidak ada field yang diubah")
	}
	req = NormalizePatch(req)
	if err := helper.Validate(&req); err != nil {
		return err
	}

	current, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return translateError(err, bookMessages, "gagal mengambil data buku")
	}

	book, err := s.repo.Update(ctx, ApplyPatch(current, req))
	if err != nil {
		return translateError(err, bookMessages, "gagal memperbarui buku")
	}

	return helper.Success(c, fiber.StatusOK, "buku berhasil diperbarui sebagian", book)
}

func (s *BookService) Delete(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	id, err := helper.ParamID(c)
	if err != nil {
		return err
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return translateError(err, bookMessages, "gagal menghapus buku")
	}

	return helper.NoContent(c)
}
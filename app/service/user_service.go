package service

import (
	"strings"

	"github.com/gofiber/fiber/v2"

	"peminjaman-buku/app/model"
	"peminjaman-buku/app/repository"
	"peminjaman-buku/helper"
)

var userSortFields = map[string]bool{
	"id":         true,
	"username":   true,
	"email":      true,
	"created_at": true,
}

type UserService struct {
	repo  repository.UserRepository
	perms *helper.PermissionSet
}

func NewUserService(repo repository.UserRepository, perms *helper.PermissionSet) *UserService {
	return &UserService{repo: repo, perms: perms}
}

func (s *UserService) List(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	q := model.UserListQuery{
		ListQuery: helper.ParseListQuery(c, userSortFields),
		Role:      strings.TrimSpace(c.Query("role")),
	}

	users, total, err := s.repo.FindAll(ctx, q)
	if err != nil {
		return translateError(err, userMessages, "gagal mengambil daftar user")
	}

	return helper.SuccessList(c, "daftar user berhasil diambil", users, model.Meta{
		Page:       q.Page,
		Limit:      q.Limit,
		Total:      total,
		TotalPages: CountTotalPages(total, q.Limit),
	})
}

func (s *UserService) Get(c *fiber.Ctx) error {
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

	if !CanAccessUser(current, id, s.perms, "user:read:any") {
		return helper.Forbidden("tidak berhak mengakses data user lain")
	}

	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return translateError(err, userMessages, "gagal mengambil data user")
	}

	return helper.Success(c, fiber.StatusOK, "user ditemukan", user)
}

func (s *UserService) Patch(c *fiber.Ctx) error {
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

	if !CanAccessUser(current, id, s.perms, "user:update:any") {
		return helper.Forbidden("tidak berhak mengubah data user lain")
	}

	var req model.PatchUserRequest
	if err := helper.BindJSON(c, &req); err != nil {
		return err
	}
	if IsEmptyUserPatch(req) {
		return helper.BadRequest("tidak ada field yang diubah")
	}
	req = NormalizeUserPatch(req)
	if err := helper.Validate(&req); err != nil {
		return err
	}

	existing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return translateError(err, userMessages, "gagal mengambil data user")
	}

	user, err := s.repo.Update(ctx, ApplyUserPatch(existing, req))
	if err != nil {
		return translateError(err, userMessages, "gagal memperbarui user")
	}

	return helper.Success(c, fiber.StatusOK, "user berhasil diperbarui", user)
}

func (s *UserService) Delete(c *fiber.Ctx) error {
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

	if current.UserID == id {
		return helper.Forbidden("tidak boleh menghapus akun sendiri")
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return translateError(err, userMessages, "gagal menghapus user")
	}

	return helper.NoContent(c)
}

func (s *UserService) AssignRole(c *fiber.Ctx) error {
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

	var req model.AssignRoleRequest
	if err := helper.BindJSON(c, &req); err != nil {
		return err
	}
	req.Role = strings.TrimSpace(req.Role)
	if err := helper.Validate(&req); err != nil {
		return err
	}
	if fields := ValidateAssignRole(current, id, req.Role, s.perms); len(fields) > 0 {
		return helper.Validation(fields)
	}

	user, err := s.repo.UpdateRole(ctx, id, req.Role)
	if err != nil {
		return translateError(err, userMessages, "gagal mengubah role user")
	}

	return helper.Success(c, fiber.StatusOK, "role user berhasil diubah", user)
}
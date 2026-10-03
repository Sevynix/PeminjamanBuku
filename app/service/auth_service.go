package service

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"peminjaman-buku/app/model"
	"peminjaman-buku/app/repository"
	"peminjaman-buku/helper"
)

const (
	refreshTokenBytes  = 32
	roleUser           = "user"
	invalidCredentials = "username atau password salah"
)

var userMessages = errorMessages{
	NotFound:  "user tidak ditemukan",
	Duplicate: "username atau email sudah terdaftar",
	Conflict:  "user tidak dapat dihapus karena masih memiliki data terkait",
	Check:     "data user melanggar batasan",
}

type AuthService struct {
	users      repository.UserRepository
	tokens     repository.TokenRepository
	jwt        *helper.JWTManager
	refreshTTL time.Duration
}

func NewAuthService(
	users repository.UserRepository,
	tokens repository.TokenRepository,
	jwtManager *helper.JWTManager,
	refreshTTL time.Duration,
) *AuthService {
	return &AuthService{users: users, tokens: tokens, jwt: jwtManager, refreshTTL: refreshTTL}
}

func (s *AuthService) Register(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	var req model.RegisterRequest
	if err := helper.BindJSON(c, &req); err != nil {
		return err
	}
	req.Username = strings.TrimSpace(req.Username)
	req.Email = strings.TrimSpace(req.Email)
	if err := helper.Validate(&req); err != nil {
		return err
	}

	hashed, err := helper.HashPassword(req.Password)
	if err != nil {
		return helper.Internal("gagal memproses password", err)
	}

	created, err := s.users.Create(ctx, model.User{
		Username: req.Username,
		Email:    req.Email,
		Password: hashed,
		Role:     roleUser,
	})
	if err != nil {
		return translateError(err, userMessages, "gagal mendaftarkan user")
	}

	return helper.Created(c, "pendaftaran berhasil", created, "/api/v1/users/"+strconv.Itoa(created.ID))
}

func (s *AuthService) Login(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	var req model.LoginRequest
	if err := helper.BindJSON(c, &req); err != nil {
		return err
	}
	req.Username = strings.TrimSpace(req.Username)
	if err := helper.Validate(&req); err != nil {
		return err
	}

	user, err := s.users.FindByUsername(ctx, req.Username)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			helper.VerifyDummyPassword(req.Password)
			return helper.Unauthorized(invalidCredentials)
		}
		return helper.Internal("gagal memproses login", err)
	}

	if !helper.VerifyPassword(user.Password, req.Password) {
		return helper.Unauthorized(invalidCredentials)
	}

	pair, err := s.issueTokenPair(ctx, user)
	if err != nil {
		return helper.Internal("gagal membuat token", err)
	}

	return helper.Success(c, fiber.StatusOK, "login berhasil", pair)
}

func (s *AuthService) Refresh(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	var req model.RefreshRequest
	if err := helper.BindJSON(c, &req); err != nil {
		return err
	}
	req.RefreshToken = strings.TrimSpace(req.RefreshToken)
	if err := helper.Validate(&req); err != nil {
		return err
	}

	userID, err := s.tokens.Consume(ctx, helper.SHA256Hex(req.RefreshToken))
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return helper.Unauthorized("refresh token tidak valid atau sudah kedaluwarsa")
		}
		return helper.Internal("gagal memperbarui token", err)
	}

	user, err := s.users.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return helper.Unauthorized("akun tidak dapat dipakai")
		}
		return helper.Internal("gagal memperbarui token", err)
	}

	pair, err := s.issueTokenPair(ctx, user)
	if err != nil {
		return helper.Internal("gagal membuat token", err)
	}

	return helper.Success(c, fiber.StatusOK, "token berhasil diperbarui", pair)
}

func (s *AuthService) Logout(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	var req model.RefreshRequest
	if err := helper.BindJSON(c, &req); err != nil {
		return err
	}

	if token := strings.TrimSpace(req.RefreshToken); token != "" {
		_ = s.tokens.Revoke(ctx, helper.SHA256Hex(token))
	}

	return helper.Success(c, fiber.StatusOK, "logout berhasil", nil)
}

func (s *AuthService) Me(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	current, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Unauthorized("belum terautentikasi")
	}

	user, err := s.users.FindByID(ctx, current.UserID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return helper.Unauthorized("user tidak ditemukan")
		}
		return helper.Internal("gagal mengambil profil", err)
	}

	return helper.Success(c, fiber.StatusOK, "profil berhasil diambil", user)
}

func (s *AuthService) issueTokenPair(ctx context.Context, user model.User) (model.TokenPair, error) {
	accessToken, err := s.jwt.GenerateAccess(user)
	if err != nil {
		return model.TokenPair{}, err
	}

	refreshToken, err := helper.RandomToken(refreshTokenBytes)
	if err != nil {
		return model.TokenPair{}, err
	}

	err = s.tokens.Save(ctx, user.ID, helper.SHA256Hex(refreshToken), time.Now().Add(s.refreshTTL))
	if err != nil {
		return model.TokenPair{}, err
	}

	return model.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int(s.jwt.AccessTTL().Seconds()),
	}, nil
}
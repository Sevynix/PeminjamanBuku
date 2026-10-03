package middleware

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"

	"peminjaman-buku/helper"
)

func unauthorized(message string) *helper.AppError {
	return helper.Unauthorized(message).WithHeader(fiber.HeaderWWWAuthenticate, `Bearer realm="api"`)
}

func RequireAuth(jwtManager *helper.JWTManager) fiber.Handler {
	return func(c *fiber.Ctx) error {
		token, ok := bearerToken(c)
		if !ok {
			return unauthorized("header Authorization tidak ada atau salah bentuk")
		}

		authUser, err := jwtManager.Parse(token)
		if err != nil {
			if errors.Is(err, helper.ErrExpiredToken) {
				return unauthorized("access token kedaluwarsa")
			}
			return unauthorized("access token tidak valid")
		}

		c.Locals(helper.LocalsAuthUser, authUser)
		return c.Next()
	}
}

func bearerToken(c *fiber.Ctx) (string, bool) {
	parts := strings.SplitN(c.Get(fiber.HeaderAuthorization), " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", false
	}

	token := strings.TrimSpace(parts[1])
	return token, token != ""
}
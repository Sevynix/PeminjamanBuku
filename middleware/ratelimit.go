package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"

	"peminjaman-buku/helper"
)

func LoginRateLimiter() fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        5,
		Expiration: time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return helper.TooManyRequests("terlalu banyak percobaan login, coba lagi dalam satu menit").
				WithHeader(fiber.HeaderRetryAfter, "60")
		},
	})
}
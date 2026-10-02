package middleware

import (
	"log/slog"
	"mime"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"

	"peminjaman-buku/helper"
)

func Register(app *fiber.App, logger *slog.Logger) {
	app.Use(requestid.New())
	app.Use(RequestLogger(logger))
	app.Use(recover.New())
	app.Use(helmet.New())
	app.Use(cors.New())
}

func RequestLogger(logger *slog.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		err := c.Next()
		if err != nil {
			if handlerErr := c.App().ErrorHandler(c, err); handlerErr != nil {
				_ = c.SendStatus(fiber.StatusInternalServerError)
			}
		}

		logger.Info("http_request",
			slog.String("request_id", helper.RequestID(c)),
			slog.String("method", c.Method()),
			slog.String("path", c.Path()),
			slog.Int("status", c.Response().StatusCode()),
			slog.Duration("duration", time.Since(start)),
			slog.String("ip", c.IP()),
		)

		return nil
	}
}

var methodsWithBody = map[string]bool{
	fiber.MethodPost:  true,
	fiber.MethodPut:   true,
	fiber.MethodPatch: true,
}

func RequireJSON(c *fiber.Ctx) error {
	if methodsWithBody[c.Method()] && len(c.Body()) > 0 {
		mediaType, _, err := mime.ParseMediaType(c.Get(fiber.HeaderContentType))
		if err != nil || mediaType != fiber.MIMEApplicationJSON {
			return helper.UnsupportedMediaType("Content-Type harus application/json")
		}
	}
	return c.Next()
}
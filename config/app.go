package config

import (
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v2"

	"peminjaman-buku/helper"
	"peminjaman-buku/middleware"
	"peminjaman-buku/route"
)

func NewApp(logger *slog.Logger, deps route.Dependencies) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName:      GetEnv("APP_NAME", "Peminjaman Buku API"),
		ErrorHandler: newErrorHandler(logger),
		BodyLimit:    1 * 1024 * 1024,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	})

	middleware.Register(app, logger)
	route.Register(app, deps)

	app.Use(func(c *fiber.Ctx) error {
		return helper.NotFound("endpoint tidak ditemukan")
	})

	return app
}

func newErrorHandler(logger *slog.Logger) fiber.ErrorHandler {
	return func(c *fiber.Ctx, err error) error {
		appErr := helper.AsAppError(err)

		attrs := []any{
			slog.String("request_id", helper.RequestID(c)),
			slog.String("method", c.Method()),
			slog.String("path", c.Path()),
			slog.String("code", appErr.Code),
			slog.Int("status", appErr.Status),
			slog.String("error", appErr.Error()),
		}
		if appErr.Status >= fiber.StatusInternalServerError {
			logger.Error("request_failed", attrs...)
		} else {
			logger.Warn("request_rejected", attrs...)
		}

		return helper.WriteError(c, appErr)
	}
}
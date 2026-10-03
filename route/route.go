package route

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"peminjaman-buku/app/service"
	"peminjaman-buku/helper"
	"peminjaman-buku/middleware"
)

type Dependencies struct {
	Pool        *pgxpool.Pool
	JWT         *helper.JWTManager
	AuthService *service.AuthService
	BookService *service.BookService
}

func Register(app *fiber.App, deps Dependencies) {
	api := app.Group("/api/v1")

	api.Get("/health", healthCheck(deps.Pool))

	auth := api.Group("/auth", middleware.RequireJSON)
	auth.Post("/register", deps.AuthService.Register)
	auth.Post("/login", middleware.LoginRateLimiter(), deps.AuthService.Login)
	auth.Post("/refresh", deps.AuthService.Refresh)
	auth.Post("/logout", deps.AuthService.Logout)
	auth.Get("/me", middleware.RequireAuth(deps.JWT), deps.AuthService.Me)

	books := api.Group("/books", middleware.RequireAuth(deps.JWT), middleware.RequireJSON)
	books.Get("/", deps.BookService.List)
	books.Post("/", deps.BookService.Create)
	books.Get("/:id", deps.BookService.Get)
	books.Put("/:id", deps.BookService.Replace)
	books.Patch("/:id", deps.BookService.Patch)
	books.Delete("/:id", deps.BookService.Delete)
}

func healthCheck(pool *pgxpool.Pool) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.UserContext(), 2*time.Second)
		defer cancel()

		if err := pool.Ping(ctx); err != nil {
			return helper.ServiceUnavailable("database tidak dapat dihubungi").WithCause(err)
		}
		return helper.Success(c, fiber.StatusOK, "server dan database berjalan", nil)
	}
}
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
	Permissions *helper.PermissionSet
	AuthService *service.AuthService
	BookService *service.BookService
	UserService *service.UserService
}

func Register(app *fiber.App, deps Dependencies) {
	api := app.Group("/api/v1")

	api.Get("/health", healthCheck(deps.Pool))

	registerAuth(api, deps)
	registerBooks(api, deps)
	registerUsers(api, deps)
}

func registerAuth(api fiber.Router, deps Dependencies) {
	auth := api.Group("/auth", middleware.RequireJSON)
	auth.Post("/register", deps.AuthService.Register)
	auth.Post("/login", middleware.LoginRateLimiter(), deps.AuthService.Login)
	auth.Post("/refresh", deps.AuthService.Refresh)
	auth.Post("/logout", deps.AuthService.Logout)
	auth.Get("/me", middleware.RequireAuth(deps.JWT), deps.AuthService.Me)
}

func registerBooks(api fiber.Router, deps Dependencies) {
	books := api.Group("/books", middleware.RequireAuth(deps.JWT), middleware.RequireJSON)
	perms := deps.Permissions

	books.Get("/", deps.BookService.List)
	books.Get("/:id", deps.BookService.Get)

	books.Post("/", middleware.RequirePermission(perms, "book:create"), deps.BookService.Create)
	books.Put("/:id", middleware.RequirePermission(perms, "book:update"), deps.BookService.Replace)
	books.Patch("/:id", middleware.RequirePermission(perms, "book:update"), deps.BookService.Patch)
	books.Delete("/:id", middleware.RequirePermission(perms, "book:delete"), deps.BookService.Delete)
}

func registerUsers(api fiber.Router, deps Dependencies) {
	users := api.Group("/users", middleware.RequireAuth(deps.JWT), middleware.RequireJSON)

	registerPermissionGuarded(users, deps)
	registerOwnershipChecked(users, deps)
}

func registerPermissionGuarded(users fiber.Router, deps Dependencies) {
	perms := deps.Permissions

	users.Get("/", middleware.RequirePermission(perms, "user:list"), deps.UserService.List)
	users.Delete("/:id", middleware.RequirePermission(perms, "user:delete"), deps.UserService.Delete)
	users.Patch("/:id/role", middleware.RequirePermission(perms, "role:assign"), deps.UserService.AssignRole)
}

func registerOwnershipChecked(users fiber.Router, deps Dependencies) {
	users.Get("/:id", deps.UserService.Get)
	users.Patch("/:id", deps.UserService.Patch)
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
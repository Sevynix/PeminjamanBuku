package helper

import (
	"context"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"peminjaman-buku/app/model"
)

const (
	DefaultLimit = 10
	MaxLimit     = 100
	MaxPage      = 100000
	maxSearchLen = 100
)

func RequestContext(c *fiber.Ctx) (context.Context, context.CancelFunc) {
	return context.WithTimeout(c.UserContext(), 5*time.Second)
}

func RequestID(c *fiber.Ctx) string {
	id, _ := c.Locals("requestid").(string)
	return id
}

func ParamID(c *fiber.Ctx) (int, error) {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id < 1 || id > math.MaxInt32 {
		return 0, BadRequest("id harus berupa angka positif yang valid")
	}
	return id, nil
}

func ParseListQuery(c *fiber.Ctx, allowedSort map[string]bool) model.ListQuery {
	q := model.ListQuery{
		Page:   c.QueryInt("page", 1),
		Limit:  c.QueryInt("limit", DefaultLimit),
		Search: strings.TrimSpace(c.Query("search")),
		Sort:   c.Query("sort", "id"),
		Order:  strings.ToLower(c.Query("order", "asc")),
	}

	if q.Page < 1 {
		q.Page = 1
	}
	if q.Page > MaxPage {
		q.Page = MaxPage
	}
	if q.Limit < 1 {
		q.Limit = DefaultLimit
	}
	if q.Limit > MaxLimit {
		q.Limit = MaxLimit
	}
	if runes := []rune(q.Search); len(runes) > maxSearchLen {
		q.Search = string(runes[:maxSearchLen])
	}
	if !allowedSort[q.Sort] {
		q.Sort = "id"
	}
	if q.Order != "desc" {
		q.Order = "asc"
	}
	return q
}

func QueryBool(c *fiber.Ctx, key string) (*bool, error) {
	raw := strings.TrimSpace(c.Query(key))
	if raw == "" {
		return nil, nil
	}
	value, err := strconv.ParseBool(raw)
	if err != nil {
		return nil, BadRequest(key + " harus berupa true atau false")
	}
	return &value, nil
}

func QueryPositiveInt(c *fiber.Ctx, key string) (*int, error) {
	raw := strings.TrimSpace(c.Query(key))
	if raw == "" {
		return nil, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 1 || value > math.MaxInt32 {
		return nil, BadRequest(key + " harus berupa angka positif yang valid")
	}
	return &value, nil
}

func BindJSON(c *fiber.Ctx, dst any) error {
	if err := c.BodyParser(dst); err != nil {
		return BadRequest("body harus berupa JSON yang valid").WithCause(err)
	}
	return nil
}
package helper

import (
	"encoding/base64"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"peminjaman-buku/app/model"
)

func EncodeCursor(borrowedAt time.Time, id int) string {
	raw := strconv.FormatInt(borrowedAt.UTC().UnixNano(), 10) + "|" + strconv.Itoa(id)
	return base64.RawURLEncoding.EncodeToString([]byte(raw))
}

func DecodeCursor(encoded string) (model.Cursor, error) {
	invalid := BadRequest("cursor tidak valid")

	decoded, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return model.Cursor{}, invalid
	}

	parts := strings.Split(string(decoded), "|")
	if len(parts) != 2 {
		return model.Cursor{}, invalid
	}

	nanos, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return model.Cursor{}, invalid
	}

	id, err := strconv.Atoi(parts[1])
	if err != nil || id < 1 || id > math.MaxInt32 {
		return model.Cursor{}, invalid
	}

	return model.Cursor{BorrowedAt: time.Unix(0, nanos).UTC(), ID: id}, nil
}

func ParseCursorParams(c *fiber.Ctx) (int, *model.Cursor, error) {
	limit := c.QueryInt("limit", DefaultLimit)
	if limit < 1 {
		limit = DefaultLimit
	}
	if limit > MaxLimit {
		limit = MaxLimit
	}

	raw := strings.TrimSpace(c.Query("cursor"))
	if raw == "" {
		return limit, nil, nil
	}

	cursor, err := DecodeCursor(raw)
	if err != nil {
		return 0, nil, err
	}
	return limit, &cursor, nil
}
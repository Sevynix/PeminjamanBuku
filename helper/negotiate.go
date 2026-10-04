package helper

import (
	"strings"

	"github.com/gofiber/fiber/v2"
)

const (
	FormatJSON = fiber.MIMEApplicationJSON
	FormatCSV  = "text/csv"
)

func Negotiate(c *fiber.Ctx, offered ...string) (string, error) {
	if strings.TrimSpace(c.Get(fiber.HeaderAccept)) == "" {
		return offered[0], nil
	}

	chosen := c.Accepts(offered...)
	if chosen == "" {
		return "", NotAcceptable("format yang diminta tidak tersedia, pilih salah satu dari: " + strings.Join(offered, ", "))
	}
	return chosen, nil
}
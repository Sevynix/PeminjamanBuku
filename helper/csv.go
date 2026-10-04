package helper

import (
	"encoding/csv"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"peminjaman-buku/app/model"
)

func safeCell(value string) string {
	if value != "" && strings.ContainsRune("=+-@\t\r", rune(value[0])) {
		return "'" + value
	}
	return value
}

func formatTime(t time.Time) string {
	return t.UTC().Format(time.RFC3339)
}

func WriteLoansCSV(c *fiber.Ctx, loans []model.Loan) error {
	var buffer strings.Builder
	writer := csv.NewWriter(&buffer)

	header := []string{"id", "username", "book_title", "borrowed_at", "due_at", "returned_at", "status"}
	if err := writer.Write(header); err != nil {
		return Internal("gagal membuat CSV", err)
	}

	for _, loan := range loans {
		returnedAt := ""
		if loan.ReturnedAt != nil {
			returnedAt = formatTime(*loan.ReturnedAt)
		}

		row := []string{
			strconv.Itoa(loan.ID),
			safeCell(loan.Username),
			safeCell(loan.BookTitle),
			formatTime(loan.BorrowedAt),
			formatTime(loan.DueAt),
			returnedAt,
			loan.Status,
		}
		if err := writer.Write(row); err != nil {
			return Internal("gagal membuat CSV", err)
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return Internal("gagal membuat CSV", err)
	}

	c.Set(fiber.HeaderContentType, FormatCSV+"; charset=utf-8")
	c.Set(fiber.HeaderContentDisposition, `attachment; filename="loans.csv"`)
	return c.SendString(buffer.String())
}
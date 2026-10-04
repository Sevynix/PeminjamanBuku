package service

import (
	"time"

	"peminjaman-buku/app/model"
	"peminjaman-buku/helper"
)

const (
	MaxActiveLoans   = 3
	LoanDurationDays = 7

	StatusActive   = "active"
	StatusReturned = "returned"
	StatusOverdue  = "overdue"
)

func CalcDueDate(borrowedAt time.Time) time.Time {
	return borrowedAt.AddDate(0, 0, LoanDurationDays)
}

func LoanStatus(returnedAt *time.Time, dueAt, now time.Time) string {
	if returnedAt != nil {
		return StatusReturned
	}
	if now.After(dueAt) {
		return StatusOverdue
	}
	return StatusActive
}

func WithStatus(loan model.Loan, now time.Time) model.Loan {
	loan.Status = LoanStatus(loan.ReturnedAt, loan.DueAt, now)
	return loan
}

func IsValidLoanStatus(status string) bool {
	return status == StatusActive || status == StatusReturned || status == StatusOverdue
}

func CanAccessLoan(
	current model.AuthUser,
	ownerID int,
	perms *helper.PermissionSet,
	anyPermission string,
) bool {
	if current.UserID == ownerID {
		return true
	}
	return perms.Can(current.Role, anyPermission)
}
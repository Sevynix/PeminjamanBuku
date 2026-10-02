package helper

import (
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

const (
	MinPasswordLength = 8
	MaxPasswordBytes = 72
)

var weakPasswords = map[string]bool{
	"password1": true, "password123": true, "qwerty123": true,
	"admin123": true, "abc12345": true, "welcome1": true,
}

func CheckPasswordStrength(password string) string {
	if utf8.RuneCountInString(password) < MinPasswordLength {
		return "minimal " + strconv.Itoa(MinPasswordLength) + " karakter"
	}
	if len(password) > MaxPasswordBytes {
		return "terlalu panjang, maksimal " + strconv.Itoa(MaxPasswordBytes) + " byte"
	}

	var hasLetter, hasDigit bool
	for _, r := range password {
		switch {
		case unicode.IsLetter(r):
			hasLetter = true
		case unicode.IsDigit(r):
			hasDigit = true
		}
	}
	if !hasLetter || !hasDigit {
		return "harus memuat huruf dan angka"
	}

	if weakPasswords[strings.ToLower(password)] {
		return "password terlalu umum"
	}
	return ""
}

func IsValidUsername(username string) bool {
	if username == "" {
		return false
	}
	for _, r := range username {
		isLetter := (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')
		isDigit := r >= '0' && r <= '9'
		if !isLetter && !isDigit && r != '.' && r != '_' {
			return false
		}
	}
	return true
}

var isbnCleaner = strings.NewReplacer("-", "", " ", "")

func NormalizeISBN(isbn string) string {
	return isbnCleaner.Replace(strings.TrimSpace(isbn))
}

func IsValidISBN(isbn string) bool {
	normalized := NormalizeISBN(isbn)
	if len(normalized) != 13 {
		return false
	}
	for _, r := range normalized {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}
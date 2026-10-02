package helper

import (
	"errors"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate = newValidator()

func newValidator() *validator.Validate {
	v := validator.New()

	v.RegisterTagNameFunc(func(field reflect.StructField) string {
		name := strings.SplitN(field.Tag.Get("json"), ",", 2)[0]
		if name == "" || name == "-" {
			return field.Name
		}
		return name
	})

	mustRegister(v, "username", func(fl validator.FieldLevel) bool {
		return IsValidUsername(fl.Field().String())
	})
	mustRegister(v, "isbn", func(fl validator.FieldLevel) bool {
		return IsValidISBN(fl.Field().String())
	})
	mustRegister(v, "strongpassword", func(fl validator.FieldLevel) bool {
		return CheckPasswordStrength(fl.Field().String()) == ""
	})

	return v
}

func mustRegister(v *validator.Validate, tag string, fn validator.Func) {
	if err := v.RegisterValidation(tag, fn); err != nil {
		panic("gagal mendaftarkan aturan validasi " + tag + ": " + err.Error())
	}
}

func Validate(s any) error {
	err := validate.Struct(s)
	if err == nil {
		return nil
	}

	var invalid *validator.InvalidValidationError
	if errors.As(err, &invalid) {
		return Internal("terjadi kesalahan pada server", err)
	}

	var fieldErrors validator.ValidationErrors
	if !errors.As(err, &fieldErrors) {
		return Internal("terjadi kesalahan pada server", err)
	}

	fields := make(map[string]string, len(fieldErrors))
	for _, fe := range fieldErrors {
		if _, exists := fields[fe.Field()]; !exists {
			fields[fe.Field()] = messageFor(fe)
		}
	}
	return Validation(fields)
}

func messageFor(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "wajib diisi"
	case "email":
		return "format email tidak valid"
	case "min":
		if fe.Kind() == reflect.String {
			return "minimal " + fe.Param() + " karakter"
		}
		return "nilai minimal " + fe.Param()
	case "max":
		if fe.Kind() == reflect.String {
			return "maksimal " + fe.Param() + " karakter"
		}
		return "nilai maksimal " + fe.Param()
	case "oneof":
		return "harus salah satu dari: " + strings.ReplaceAll(fe.Param(), " ", ", ")
	case "username":
		return "hanya boleh huruf, angka, titik, dan garis bawah"
	case "isbn":
		return "format ISBN tidak valid (harus 13 digit)"
	case "strongpassword":
		if value, ok := fe.Value().(string); ok {
			if msg := CheckPasswordStrength(value); msg != "" {
				return msg
			}
		}
		return "password tidak memenuhi syarat"
	default:
		return "tidak memenuhi aturan " + fe.Tag()
	}
}
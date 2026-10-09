package validators

import (
	"errors"
	"reflect"
	"strings"
	"unicode"

	"github.com/go-playground/validator/v10"
)

var validate = newValidator()

var commonPasswords = map[string]bool{
	"password":    true,
	"password123": true,
	"admin123":    true,
	"rahasia123":  true,
	"12345678":    true,
	"qwerty123":   true,
}

func passwordStrength(password string) string {
	if len(password) < 8 {
		return "minimal 8 karakter"
	}
	if commonPasswords[strings.ToLower(password)] {
		return "password terlalu umum"
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
	return ""
}

func newValidator() *validator.Validate {
	v := validator.New()

	// Pemetaan nama field ke JSON tag
	v.RegisterTagNameFunc(func(field reflect.StructField) string {
		name := strings.SplitN(field.Tag.Get("json"), ",", 2)[0]
		if name == "" || name == "-" {
			return field.Name
		}
		return name
	})

	_ = v.RegisterValidation("nospace", func(fl validator.FieldLevel) bool {
		return !strings.ContainsAny(fl.Field().String(), " \t\n\r")
	})

	_ = v.RegisterValidation("username", func(fl validator.FieldLevel) bool {
		val := fl.Field().String()
		if val == "" {
			return false
		}
		for _, r := range val {
			if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '.' && r != '_' {
				return false
			}
		}
		return true
	})

	// Perbaikan Bug 6: == "" menandakan password kuat (valid)
	_ = v.RegisterValidation("strongpassword", func(fl validator.FieldLevel) bool {
		return passwordStrength(fl.Field().String()) == ""
	})

	// Custom validator untuk Entity Students (D.2 butir 2)
	_ = v.RegisterValidation("nim", func(fl validator.FieldLevel) bool {
		val := fl.Field().String()
		if len(val) < 9 || len(val) > 18 {
			return false
		}
		for _, r := range val {
			if !unicode.IsDigit(r) {
				return false
			}
		}
		return true
	})

	return v
}

func ValidateStruct(s any) map[string]string {
	err := validate.Struct(s)
	if err == nil {
		return nil
	}

	var invalid *validator.InvalidValidationError
	if errors.As(err, &invalid) {
		return map[string]string{"_": "objek yang divalidasi tidak sah"}
	}

	var fieldErrors validator.ValidationErrors
	if !errors.As(err, &fieldErrors) {
		return map[string]string{"_": "validasi gagal"}
	}

	result := make(map[string]string, len(fieldErrors))
	for _, fe := range fieldErrors {
		if _, exists := result[fe.Field()]; !exists {
			result[fe.Field()] = messageFor(fe)
		}
	}
	return result
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
	case "alphanum":
		return "hanya boleh berisi huruf dan angka"
	case "nospace":
		return "tidak boleh mengandung spasi"
	case "username":
		return "hanya boleh huruf, angka, titik, dan garis bawah"
	case "strongpassword":
		if value, ok := fe.Value().(string); ok {
			msg := passwordStrength(value)
			if msg != "" {
				return msg
			}
		}
		return "password tidak memenuhi syarat"
	case "oneof":
		return "harus salah satu dari: " + strings.ReplaceAll(fe.Param(), " ", ", ")
	case "nim":
		return "format NIM tidak valid (harus 9-18 digit angka)"
	default:
		return "tidak memenuhi aturan " + fe.Tag()
	}
}

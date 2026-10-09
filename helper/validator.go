package helper

import (
	"errors"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
)

var (
	validate           = newValidator()
	nimRegex           = regexp.MustCompile(`^[0-9]{12}$`)
	tahunAkademikRegex = regexp.MustCompile(`^[0-9]{4}/[0-9]{4}-(Ganjil|Genap)$`)
)

func newValidator() *validator.Validate {
	v := validator.New()

	v.RegisterTagNameFunc(func(field reflect.StructField) string {
		name := strings.SplitN(field.Tag.Get("json"), ",", 2)[0]
		if name == "" || name == "-" {
			return field.Name
		}
		return name
	})

	_ = v.RegisterValidation("nim", func(fl validator.FieldLevel) bool {
		return nimRegex.MatchString(fl.Field().String())
	})

	_ = v.RegisterValidation("angkatan", func(fl validator.FieldLevel) bool {
		var year int
		switch fl.Field().Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			year = int(fl.Field().Int())
		case reflect.String:
			var err error
			year, err = strconv.Atoi(fl.Field().String())
			if err != nil {
				return false
			}
		default:
			return false
		}
		return year >= 1000 && year <= 9999 && year <= time.Now().Year()
	})

	_ = v.RegisterValidation("tahunakademik", func(fl validator.FieldLevel) bool {
		return tahunAkademikRegex.MatchString(fl.Field().String())
	})

	return v
}

func ValidateStruct(s any) map[string][]string {
	err := validate.Struct(s)
	if err == nil {
		return nil
	}

	var invalid *validator.InvalidValidationError
	if errors.As(err, &invalid) {
		return map[string][]string{"_": {"objek yang divalidasi tidak sah"}}
	}

	var fieldErrors validator.ValidationErrors
	if !errors.As(err, &fieldErrors) {
		return map[string][]string{"_": {"validasi gagal"}}
	}

	result := make(map[string][]string, len(fieldErrors))
	for _, fe := range fieldErrors {
		fieldName := fe.Field()
		result[fieldName] = append(result[fieldName], messageFor(fe))
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
	case "gt":
		return "harus lebih besar dari " + fe.Param()
	case "gte":
		return "minimal " + fe.Param()
	case "lt":
		return "harus kurang dari " + fe.Param()
	case "lte":
		return "maksimal " + fe.Param()
	case "alphanum":
		return "hanya boleh berisi huruf dan angka"
	case "nim":
		return "format NIM tidak valid (harus 12 digit angka)"
	case "angkatan":
		return "angkatan tidak valid (harus 4 digit angka dan tidak melebihi tahun saat ini)"
	case "tahunakademik":
		return "format tahun akademik tidak valid (contoh: 2026/2027-Ganjil)"
	default:
		return "tidak memenuhi aturan " + fe.Tag()
	}
}

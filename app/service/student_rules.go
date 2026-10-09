package service

import (
	"errors"
	"strings"

	"siakad-mini/app/model"
)

var ErrInvalidSort = errors.New("parameter sort tidak valid: hanya diperbolehkan 'nama' atau '-ipk_terakhir'")

const DefaultSort = "nama"

// CountLastPage menghitung jumlah halaman terakhir menggunakan pembagian pembulatan ke atas (ceiling division).
func CountLastPage(total, perPage int) int {
	if perPage <= 0 || total <= 0 {
		return 0
	}
	return (total + perPage - 1) / perPage
}

// ValidateSort memastikan urutan sort hanya 'nama' atau '-ipk_terakhir'. Jika kosong, mengembalikan nilai default.
func ValidateSort(sort string) (string, error) {
	trimmed := strings.TrimSpace(sort)
	if trimmed == "" {
		return DefaultSort, nil
	}
	switch trimmed {
	case "nama", "-ipk_terakhir":
		return trimmed, nil
	default:
		return "", ErrInvalidSort
	}
}

// ApplyUpdate menerapkan perubahan dari UpdateStudentRequest ke entity Student saat ini.
// Kolom NIM tidak pernah diubah (immutable).
func ApplyUpdate(current model.Student, req model.UpdateStudentRequest) model.Student {
	current.Nama = strings.TrimSpace(req.Nama)
	current.Prodi = strings.TrimSpace(req.Prodi)
	current.Angkatan = req.Angkatan
	if req.IPKTerakhir != nil {
		current.IPKTerakhir = *req.IPKTerakhir
	}
	return current
}

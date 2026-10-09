package service

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var academicYearRegex = regexp.MustCompile(`^([0-9]{4})/([0-9]{4})-(Ganjil|Genap)$`)

// MaxSKSByIPK menentukan batas maksimal SKS berdasarkan IPK terakhir mahasiswa:
// - IPK >= 3.00 -> 24 SKS
// - IPK 2.50–2.99 -> 21 SKS
// - IPK < 2.50 -> 18 SKS
func MaxSKSByIPK(ipk float64) int {
	switch {
	case ipk >= 3.00:
		return 24
	case ipk >= 2.50:
		return 21
	default:
		return 18
	}
}

// CheckSKSLimit memeriksa apakah penambahan SKS mata kuliah masih dalam batas maksimum.
// Mengembalikan sisa SKS yang tersedia (max - taken) dan pesan error jika melebihi batas.
// Jika masih dalam batas (termasuk tepat di batas limit), pesan error bernilai kosong ("").
func CheckSKSLimit(taken, adding, max int) (int, string) {
	remaining := max - taken
	if remaining < 0 {
		remaining = 0
	}
	if taken+adding > max {
		return remaining, fmt.Sprintf("Total SKS melebihi batas. Sisa SKS Anda %d, mata kuliah ini membutuhkan %d", remaining, adding)
	}
	return remaining, ""
}

// ValidateAcademicYear memvalidasi format tahun akademik:
// Bentuk harus YYYY/YYYY+1-(Ganjil|Genap) dan tahun kedua harus persis tahun pertama + 1.
func ValidateAcademicYear(s string) bool {
	matches := academicYearRegex.FindStringSubmatch(strings.TrimSpace(s))
	if len(matches) != 4 {
		return false
	}
	y1, err1 := strconv.Atoi(matches[1])
	y2, err2 := strconv.Atoi(matches[2])
	if err1 != nil || err2 != nil {
		return false
	}
	return y2 == y1+1
}

// IsCourseFull memeriksa apakah kuota kelas mata kuliah sudah penuh (terisi >= kuota).
func IsCourseFull(terisi, kuota int) bool {
	return terisi >= kuota
}

package service

import (
	"errors"
	"testing"

	"siakad-mini/app/model"
)

func TestCountLastPage(t *testing.T) {
	cases := []struct {
		name    string
		total   int
		perPage int
		want    int
	}{
		{"0 items", 0, 10, 0},
		{"exact multiple (10 items, 10 per page)", 10, 10, 1},
		{"exact multiple (20 items, 10 per page)", 20, 10, 2},
		{"remainder (1 item, 10 per page)", 1, 10, 1},
		{"remainder (11 items, 10 per page)", 11, 10, 2},
		{"remainder (25 items, 10 per page)", 25, 10, 3},
		{"invalid perPage 0", 15, 0, 0},
		{"negative perPage", 15, -1, 0},
		{"negative total", -5, 10, 0},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := CountLastPage(tc.total, tc.perPage)
			if got != tc.want {
				t.Errorf("total=%d perPage=%d: got %d, want %d", tc.total, tc.perPage, got, tc.want)
			}
		})
	}
}

func TestValidateSort(t *testing.T) {
	t.Run("empty sort returns default", func(t *testing.T) {
		got, err := ValidateSort("")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != "nama" {
			t.Errorf("expected 'nama', got %q", got)
		}

		got, err = ValidateSort("   ")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != "nama" {
			t.Errorf("expected 'nama', got %q", got)
		}
	})

	t.Run("valid sort options", func(t *testing.T) {
		for _, s := range []string{"nama", "-ipk_terakhir"} {
			got, err := ValidateSort(s)
			if err != nil {
				t.Errorf("sort %q should be valid, got err: %v", s, err)
			}
			if got != s {
				t.Errorf("expected %q, got %q", s, got)
			}
		}
	})

	t.Run("unknown sort rejected", func(t *testing.T) {
		invalidSorts := []string{"ipk_terakhir", "nim", "id", "-nama", "random"}
		for _, s := range invalidSorts {
			_, err := ValidateSort(s)
			if err == nil {
				t.Errorf("sort %q should be rejected, but was accepted", s)
			}
			if !errors.Is(err, ErrInvalidSort) {
				t.Errorf("sort %q: expected ErrInvalidSort, got %v", s, err)
			}
		}
	})
}

func TestApplyUpdate(t *testing.T) {
	initial := model.Student{
		ID:          1,
		UserID:      10,
		NIM:         "187221000001",
		Nama:        "Budi Santoso",
		Prodi:       "Teknik Informatika",
		Angkatan:    2021,
		IPKTerakhir: 3.50,
	}

	newIPK := 3.75
	req := model.UpdateStudentRequest{
		Nama:        "Budi Santoso Updated",
		Prodi:       "Sistem Informasi",
		Angkatan:    2022,
		IPKTerakhir: &newIPK,
	}

	updated := ApplyUpdate(initial, req)

	// Updated fields
	if updated.Nama != "Budi Santoso Updated" {
		t.Errorf("Nama not updated: got %q", updated.Nama)
	}
	if updated.Prodi != "Sistem Informasi" {
		t.Errorf("Prodi not updated: got %q", updated.Prodi)
	}
	if updated.Angkatan != 2022 {
		t.Errorf("Angkatan not updated: got %d", updated.Angkatan)
	}
	if updated.IPKTerakhir != 3.75 {
		t.Errorf("IPKTerakhir not updated: got %f", updated.IPKTerakhir)
	}

	// Immutable fields MUST remain untouched
	if updated.NIM != "187221000001" {
		t.Errorf("NIM must NOT change: got %q, want %q", updated.NIM, "187221000001")
	}
	if updated.ID != 1 {
		t.Errorf("ID must NOT change: got %d, want %d", updated.ID, 1)
	}
	if updated.UserID != 10 {
		t.Errorf("UserID must NOT change: got %d, want %d", updated.UserID, 10)
	}
}

func TestCanAccessStudent(t *testing.T) {
	adminUser := model.AuthUser{
		UserID: 1,
		Email:  "admin@siakad.test",
		Role:   "admin",
	}

	studentOwner := model.AuthUser{
		UserID: 2,
		Email:  "mahasiswa01@siakad.test",
		Role:   "mahasiswa",
	}

	otherStudent := model.AuthUser{
		UserID: 3,
		Email:  "mahasiswa02@siakad.test",
		Role:   "mahasiswa",
	}

	targetStudentUserID := 2

	t.Run("admin can access any student", func(t *testing.T) {
		if !CanAccessStudent(adminUser, targetStudentUserID) {
			t.Errorf("admin should be able to access any student")
		}
	})

	t.Run("owner can access own student record", func(t *testing.T) {
		if !CanAccessStudent(studentOwner, targetStudentUserID) {
			t.Errorf("student owner should be able to access own record")
		}
	})

	t.Run("other student cannot access student record", func(t *testing.T) {
		if CanAccessStudent(otherStudent, targetStudentUserID) {
			t.Errorf("other student should NOT be able to access different student's record")
		}
	})
}

package service

import (
	"strings"
	"testing"
)

func TestMaxSKSByIPK(t *testing.T) {
	cases := []struct {
		name string
		ipk  float64
		want int
	}{
		{"boundary 4.00", 4.00, 24},
		{"above 3.00 (3.50)", 3.50, 24},
		{"boundary 3.00", 3.00, 24},
		{"boundary 2.99", 2.99, 21},
		{"between 2.50 and 2.99 (2.75)", 2.75, 21},
		{"boundary 2.50", 2.50, 21},
		{"boundary 2.49", 2.49, 18},
		{"below 2.50 (1.80)", 1.80, 18},
		{"boundary 0.00", 0.00, 18},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := MaxSKSByIPK(tc.ipk)
			if got != tc.want {
				t.Errorf("IPK %.2f: got %d, want %d", tc.ipk, got, tc.want)
			}
		})
	}
}

func TestCheckSKSLimit(t *testing.T) {
	max := 24

	t.Run("within limit", func(t *testing.T) {
		remaining, msg := CheckSKSLimit(10, 3, max)
		if msg != "" {
			t.Errorf("expected no error message, got %q", msg)
		}
		if remaining != 14 {
			t.Errorf("expected remaining 14, got %d", remaining)
		}
	})

	t.Run("exactly at limit", func(t *testing.T) {
		remaining, msg := CheckSKSLimit(21, 3, max)
		if msg != "" {
			t.Errorf("expected no error message when exactly at limit, got %q", msg)
		}
		if remaining != 3 {
			t.Errorf("expected remaining 3, got %d", remaining)
		}
	})

	t.Run("one over limit", func(t *testing.T) {
		remaining, msg := CheckSKSLimit(21, 4, max)
		if msg == "" {
			t.Errorf("expected error message when exceeding limit, got empty")
		}
		if remaining != 3 {
			t.Errorf("expected remaining 3, got %d", remaining)
		}
		if !strings.Contains(msg, "Sisa SKS Anda 3") || !strings.Contains(msg, "membutuhkan 4") {
			t.Errorf("error message should specify remaining 3 and needed 4, got: %q", msg)
		}
	})

	t.Run("already full", func(t *testing.T) {
		remaining, msg := CheckSKSLimit(24, 2, max)
		if msg == "" {
			t.Errorf("expected error message when already at cap, got empty")
		}
		if remaining != 0 {
			t.Errorf("expected remaining 0, got %d", remaining)
		}
	})
}

func TestValidateAcademicYear(t *testing.T) {
	validYears := []string{
		"2026/2027-Ganjil",
		"2026/2027-Genap",
		"2024/2025-Ganjil",
		" 2024/2025-Genap ",
	}
	for _, y := range validYears {
		t.Run("valid: "+y, func(t *testing.T) {
			if !ValidateAcademicYear(y) {
				t.Errorf("academic year %q should be valid", y)
			}
		})
	}

	invalidYears := []string{
		"2026/2028-Ganjil", // gap of 2 years
		"2027/2026-Ganjil", // reversed
		"2026/2026-Ganjil", // same year
		"2026-Ganjil",      // missing second year
		"2026/2027",        // missing semester
		"2026/2027-Pendek", // invalid semester name
		"",                 // empty
		"   ",              // whitespace
		"abcd/efgh-Ganjil", // letters instead of years
	}
	for _, y := range invalidYears {
		t.Run("invalid: "+y, func(t *testing.T) {
			if ValidateAcademicYear(y) {
				t.Errorf("academic year %q should be invalid", y)
			}
		})
	}
}

func TestIsCourseFull(t *testing.T) {
	cases := []struct {
		name   string
		terisi int
		kuota  int
		want   bool
	}{
		{"seats remaining", 15, 20, false},
		{"exactly full", 20, 20, true},
		{"over capacity", 21, 20, true},
		{"zero quota full", 0, 0, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := IsCourseFull(tc.terisi, tc.kuota)
			if got != tc.want {
				t.Errorf("terisi=%d kuota=%d: got %v, want %v", tc.terisi, tc.kuota, got, tc.want)
			}
		})
	}
}

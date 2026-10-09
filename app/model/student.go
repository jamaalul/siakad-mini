package model

import "time"

type Student struct {
	ID          int        `json:"id"`
	UserID      int        `json:"user_id"`
	NIM         string     `json:"nim"`
	Nama        string     `json:"nama"`
	Prodi       string     `json:"prodi"`
	Angkatan    int        `json:"angkatan"`
	IPKTerakhir float64    `json:"ipk_terakhir"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
}

type CreateStudentRequest struct {
	NIM         string   `json:"nim" validate:"required,nim"`
	Nama        string   `json:"nama" validate:"required,min=3,max=100"`
	Email       string   `json:"email" validate:"required,email"`
	Prodi       string   `json:"prodi" validate:"required"`
	Angkatan    int      `json:"angkatan" validate:"required,angkatan"`
	IPKTerakhir *float64 `json:"ipk_terakhir,omitempty" validate:"omitempty,min=0,max=4"`
}

type UpdateStudentRequest struct {
	Nama        string   `json:"nama" validate:"required,min=3,max=100"`
	Prodi       string   `json:"prodi" validate:"required"`
	Angkatan    int      `json:"angkatan" validate:"required,angkatan"`
	IPKTerakhir *float64 `json:"ipk_terakhir,omitempty" validate:"omitempty,min=0,max=4"`
}

type StudentListQuery struct {
	Page     int    `json:"page"`
	PerPage  int    `json:"per_page"`
	Prodi    string `json:"prodi"`
	Angkatan int    `json:"angkatan"`
	Search   string `json:"search"`
	Sort     string `json:"sort"`
}

func (q StudentListQuery) Offset() int {
	if q.Page < 1 {
		return 0
	}
	return (q.Page - 1) * q.PerPage
}

type StudentDetail struct {
	Student
	Enrollments []EnrollmentDetail `json:"enrollments"`
	TotalSKS    int                `json:"total_sks"`
	BatasSKS    int                `json:"batas_sks"`
}

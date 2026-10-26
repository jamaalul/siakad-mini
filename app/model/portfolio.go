package model

import "time"

// Portfolio merepresentasikan portofolio mahasiswa. ID disimpan sebagai string (heksadesimal)
// agar package model tetap tidak bergantung pada driver database.
type Portfolio struct {
	ID           string        `json:"id"`
	StudentID    int           `json:"student_id"`
	Headline     string        `json:"headline"`
	Bio          string        `json:"bio"`
	Projects     []Project     `json:"projects"`
	Certificates []Certificate `json:"certificates"`
	CreatedAt    time.Time     `json:"created_at"`
	UpdatedAt    time.Time     `json:"updated_at"`
}

// Project disimpan secara embedded di dalam array Projects pada dokumen Portfolio.
type Project struct {
	ID           string   `json:"id"`
	Title        string   `json:"title"`
	Technologies []string `json:"technologies"`
	URL          string   `json:"url"`
	Year         int      `json:"year"`
}

// Certificate merepresentasikan sertifikat yang tersimpan embedded dalam dokumen Portfolio.
type Certificate struct {
	ID       string    `json:"id"`
	Name     string    `json:"name"`
	Issuer   string    `json:"issuer"`
	IssuedAt time.Time `json:"issued_at"`
}

// PutPortfolioRequest digunakan untuk membuat atau memperbarui profil dasar portofolio.
type PutPortfolioRequest struct {
	Headline string `json:"headline" validate:"required,min=5,max=150"`
	Bio      string `json:"bio"      validate:"required,min=10,max=1000"`
}

// AddProjectRequest digunakan untuk menambahkan item proyek baru ke array Projects.
type AddProjectRequest struct {
	Title        string   `json:"title"        validate:"required,min=3,max=100"`
	Technologies []string `json:"technologies" validate:"required,min=1,max=10,dive,min=2,max=30"`
	URL          string   `json:"url"          validate:"omitempty,url"`
	Year         int      `json:"year"         validate:"required,min=2000,max=2100"`
}

// TechCount merepresentasikan agregasi penggunaan teknologi dari seluruh proyek portofolio.
type TechCount struct {
	Technology string `json:"technology"`
	Total      int    `json:"total"`
}

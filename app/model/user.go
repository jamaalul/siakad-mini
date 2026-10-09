package model

import "time"

type Role string

const (
	RoleAdmin     Role = "admin"
	RoleMahasiswa Role = "mahasiswa"
)

type User struct {
	ID        int       `json:"id"`
	Email     string    `json:"email"`
	Password  string    `json:"-"`
	Role      Role      `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

type WebResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
	Meta    any    `json:"meta,omitempty"`
}

type Meta struct {
	CurrentPage int `json:"current_page"`
	PerPage     int `json:"per_page"`
	Total       int `json:"total"`
	LastPage    int `json:"last_page"`
}

type ErrorResponse struct {
	Success bool                `json:"success"`
	Message string              `json:"message"`
	Errors  map[string][]string `json:"errors,omitempty"`
}

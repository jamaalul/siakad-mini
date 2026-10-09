package model

import "time"

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type LoginUser struct {
	ID    int    `json:"id"`
	Email string `json:"email"`
	Role  Role   `json:"role"`
}

type LoginResponse struct {
	AccessToken string    `json:"access_token"`
	TokenType   string    `json:"token_type"`
	ExpiresIn   int       `json:"expires_in"`
	User        LoginUser `json:"user"`
}

type AuthUser struct {
	UserID int    `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
}

type MeStudentInfo struct {
	NIM      string `json:"nim"`
	Nama     string `json:"nama"`
	Prodi    string `json:"prodi"`
	Angkatan int    `json:"angkatan"`
}

type MeResponse struct {
	ID        int            `json:"id"`
	Email     string         `json:"email"`
	Role      Role           `json:"role"`
	CreatedAt time.Time      `json:"created_at"`
	Student   *MeStudentInfo `json:"student,omitempty"`
}

package service

import (
	"siakad-mini/app/model"
)

// CanAccessStudent memeriksa hak akses ke data detail mahasiswa.
// Diperbolehkan jika pengguna adalah admin ATAU pemilik data tersebut (current.UserID == studentUserID).
func CanAccessStudent(current model.AuthUser, studentUserID int) bool {
	if current.Role == string(model.RoleAdmin) {
		return true
	}
	return current.UserID == studentUserID
}

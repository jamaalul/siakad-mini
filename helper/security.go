package helper

import (
	"golang.org/x/crypto/bcrypt"
)

const bcryptCost = 12

var dummyHash = []byte("$2a$12$abcdefghijklmnopqrstuuLKa3Bt1TCmU/6zvhZ8x4nq1yBiuGvS")

func HashPassword(plain string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(plain), bcryptCost)
	if err != nil {
		return "", err
	}
	return string(hashed), nil
}

func VerifyPassword(hash, plain string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)) == nil
}

func VerifyDummyPassword(plain string) {
	_ = bcrypt.CompareHashAndPassword(dummyHash, []byte(plain))
}

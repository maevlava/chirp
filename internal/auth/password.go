package auth

import (
	"errors"
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}
func CheckPassword(hash, password string) error {
	userPassword := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if userPassword != nil {
		return errors.New("password does not match")
	}
	return nil
}

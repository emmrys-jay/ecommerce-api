package util

import "golang.org/x/crypto/bcrypt"

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hash), nil
}

func VerifyPassword(password, hashpassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashpassword), []byte(password))

	return err == nil
}

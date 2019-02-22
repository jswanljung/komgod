package backend

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// These should really return an error and a string...

func sessionToken() string {
	c := 300
	cbytes := make([]byte, c)
	_, err := rand.Read(cbytes)
	if err != nil {
		fmt.Println("error:", err)
		return ""
	}
	str := base64.StdEncoding.EncodeToString(cbytes)
	return str
}

func passwordHash(password string) string {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		fmt.Println("error:", err)
		return ""
	}
	return string(hash)
}

// But make this one boolean...

func confirmPassword(hash, password string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil && err == bcrypt.ErrMismatchedHashAndPassword {
		return LoginFailedError
	}
	return err
}

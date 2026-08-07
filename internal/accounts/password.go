package accounts

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"golang.org/x/crypto/argon2"
)

var usernamePattern = regexp.MustCompile(`^[A-Za-z0-9_]{3,24}$`)

func ValidateUsername(username string) error {
	if !usernamePattern.MatchString(username) {
		return errors.New("username must contain 3-24 latin letters, digits or underscores")
	}
	return nil
}

func ValidatePassword(password string) error {
	if len(password) < 6 || len(password) > 128 {
		return errors.New("password must contain 6-128 characters")
	}
	return nil
}

func HashPassword(password string) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	hash := argon2.IDKey([]byte(password), salt, 2, 64*1024, 1, 32)
	return fmt.Sprintf("$argon2id$v=19$m=65536,t=2,p=1$%s$%s",
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	), nil
}

func CheckPassword(password, encoded string) bool {
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 || parts[1] != "argon2id" || parts[2] != "v=19" || parts[3] != "m=65536,t=2,p=1" {
		return false
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false
	}
	want, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil || len(want) != 32 {
		return false
	}
	got := argon2.IDKey([]byte(password), salt, 2, 64*1024, 1, 32)
	return subtle.ConstantTimeCompare(got, want) == 1
}

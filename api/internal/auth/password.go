package auth

import (
	"golang.org/x/crypto/bcrypt"
)

// HashPassword hashes a plaintext password using bcrypt. In a real implementation,
// this function remains the same, but the caller would persist the resulting hash
// via a repository backed by a database (e.g., Postgres).
func HashPassword(plain string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// CheckPassword compares a bcrypt password hash with a candidate plaintext password.
func CheckPassword(passwordHash, candidate string) error {
	return bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(candidate))
}

package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type SessionUser struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	FullName string `json:"full_name"`
	Role     string `json:"role"`
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func CheckPassword(password, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

// GenerateSessionToken genereert een cryptografisch veilige random hex string van 64 tekens
func GenerateSessionToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", errors.New("kan geen random bytes genereren voor sessie")
	}
	return hex.EncodeToString(bytes), nil
}

// SessionDuration geeft de standaard geldigheidsduur van een sessie (30 dagen)
func SessionDuration() time.Duration {
	return 30 * 24 * time.Hour
}
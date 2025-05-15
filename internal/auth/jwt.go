package auth

import (
	"github.com/dgrijalva/jwt-go"
	"time"
)

var secretKey = []byte("secret_key")

// Claims структура для JWT
type Claims struct {
	UserID int `json:"user_id"`
	jwt.StandardClaims
}

// Генерация JWT
func GenerateToken(userID int, expiresAt time.Time) (string, error) {
	claims := Claims{
		UserID: userID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expiresAt.Unix(),
			IssuedAt:  time.Now().Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secretKey)
}

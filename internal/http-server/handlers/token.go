package handlers

import (
	"encoding/json"
	"github.com/dgrijalva/jwt-go"
	"log"
	"net/http"
)

type TokenRequest struct {
	Token string `json:"token"`
}

func ValidateTokenHandler(w http.ResponseWriter, r *http.Request) {
	var req TokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	tokenStr := req.Token
	if tokenStr == "" {
		http.Error(w, "Token is required", http.StatusBadRequest)
		return
	}

	// Подставь свой ключ
	secretKey := []byte("secret_key")

	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		// Проверка метода подписи
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return secretKey, nil
	})
	log.Println(token, err, req)
	if err != nil || !token.Valid {
		http.Error(w, "Token is bad", http.StatusUnauthorized)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Token is valid"))

}

package handlers

import (
	"encoding/json"
	"github.com/dgrijalva/jwt-go"
	"kursach/internal/storage/postgres"
	"net/http"
)

type TokenRequest struct {
	Token string `json:"token"`
}

func ValidateTokenHandler(userStorage postgres.UserStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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

		secretKey := []byte("secret_key")

		// Парсим токен и проверяем подпись
		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return secretKey, nil
		})
		if err != nil || !token.Valid {
			http.Error(w, "Token is invalid", http.StatusUnauthorized)
			return
		}

		// Извлекаем user_id из payload
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok || !token.Valid {
			http.Error(w, "Invalid token claims", http.StatusUnauthorized)
			return
		}

		userIDFloat, ok := claims["user_id"].(float64)
		if !ok {
			http.Error(w, "Invalid user ID in token", http.StatusUnauthorized)
			return
		}
		userID := int(userIDFloat)

		// Получаем информацию о пользователе
		userInfo, err := userStorage.GetUserInfo(r.Context(), userID)
		if err != nil {
			http.Error(w, "Failed to fetch user info", http.StatusInternalServerError)
			return
		}

		// Формируем ответ
		resp := LoginResponse{
			Token: tokenStr,
			User:  userInfo,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}

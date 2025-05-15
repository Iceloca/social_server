package handlers

import (
	"encoding/json"
	"golang.org/x/crypto/bcrypt"
	"kursach/internal/auth"
	"net/http"
	"strconv"
	"time"
)

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string                 `json:"token"`
	User  map[string]interface{} `json:"user"`
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	// Получение user_id и хеш-пароля
	userID, hashedPassword, err := h.UserStorage.GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Проверка пароля
	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(req.Password)); err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Генерация токена
	expiresAt := time.Now().Add(24 * time.Hour)
	token, err := auth.GenerateToken(userID, expiresAt)
	if err != nil {
		http.Error(w, "Token generation failed", http.StatusInternalServerError)
		return
	}

	// Сохраняем токен
	if err := h.UserStorage.SaveUserToken(r.Context(), userID, token, expiresAt); err != nil {
		http.Error(w, "Failed to save token", http.StatusInternalServerError)
		return
	}

	// Получаем user_info
	userInfo, err := h.UserStorage.GetUserInfo(r.Context(), userID)
	if err != nil {
		http.Error(w, "Failed to fetch user info", http.StatusInternalServerError)
		return
	}

	// Возвращаем ответ
	resp := LoginResponse{
		Token: token,
		User:  userInfo,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *UserHandler) GetUserInfoHandler(w http.ResponseWriter, r *http.Request) {
	// Получаем userId из query-параметров
	userIDStr := r.URL.Query().Get("userId")
	if userIDStr == "" {
		http.Error(w, "Missing userId parameter", http.StatusBadRequest)
		return
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "Invalid userId parameter", http.StatusBadRequest)
		return
	}

	userInfo, err := h.UserStorage.GetUserInfo(r.Context(), userID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(userInfo)
}

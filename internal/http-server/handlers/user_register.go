package handlers

import (
	"encoding/json"
	"golang.org/x/crypto/bcrypt"
	_ "golang.org/x/crypto/bcrypt"
	"kursach/internal/auth"
	"kursach/internal/storage/postgres"
	"net/http"
	"strings"
	"time"
)

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	UserName string `json:"user_name"`
	UserTag  string `json:"user_tag"`
}

type UserHandler struct {
	UserStorage *postgres.UserStorage
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Валидация
	req.UserTag = strings.TrimSpace(req.UserTag)
	if req.UserTag == "" || req.Email == "" || req.Password == "" || req.UserName == "" {
		http.Error(w, "Missing fields", http.StatusBadRequest)
		return
	}

	// Проверка на уникальность
	taken, err := h.UserStorage.IsUserTagTaken(r.Context(), req.UserTag)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}
	if taken {
		http.Error(w, "User tag already taken", http.StatusConflict)
		return
	}
	emailTaken, err := h.UserStorage.IsEmailTaken(r.Context(), req.Email)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}
	if emailTaken {
		http.Error(w, "Email already registered", http.StatusConflict)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}

	// Вызов процедуры для создания пользователя
	err = h.UserStorage.CreateFullUser(r.Context(), req.Email, string(hashedPassword), req.UserName, req.UserTag)
	if err != nil {
		http.Error(w, "Could not create user", http.StatusInternalServerError)
		return
	}

	// Получаем ID пользователя, который был только что создан
	// Мы предполагаем, что ID можно извлечь с помощью метода GetUserID, например, по user_tag
	userInfo, err := h.UserStorage.GetUserInfoByTag(r.Context(), req.UserTag)
	if err != nil {
		http.Error(w, "Could not retrieve user info", http.StatusInternalServerError)
		return
	}

	// Генерация JWT токена
	token, err := auth.GenerateToken(userInfo["user_id"].(int), time.Now().Add(24*time.Hour))
	if err != nil {
		http.Error(w, "Could not generate token", http.StatusInternalServerError)
		return
	}

	// Сохраняем токен в базу данных
	err = h.UserStorage.SaveUserToken(r.Context(), userInfo["user_id"].(int), token, time.Now().Add(24*time.Hour))
	if err != nil {
		http.Error(w, "Could not save token", http.StatusInternalServerError)
		return
	}

	// Отправляем ответ с информацией о пользователе и токеном
	response := map[string]interface{}{
		"user_info": userInfo,
		"token":     token,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Could not send response", http.StatusInternalServerError)
	}
}

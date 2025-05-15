package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"kursach/internal/storage/postgres"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

type UpdateUserHandler struct {
	UserStorage *postgres.UserStorage
}

func (h *UpdateUserHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	if err := r.ParseMultipartForm(10 << 20); err != nil { // 10 MB max
		http.Error(w, "Could not parse form", http.StatusBadRequest)
		return
	}
	userIDStr := r.FormValue("user_id")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "Invalid userId", http.StatusBadRequest)
		return
	}

	log.Println("user_id =", r.FormValue("user_id"))

	updates := make(map[string]interface{})

	// Обработка текстовых полей
	for key, values := range r.MultipartForm.Value {
		if len(values) > 0 {
			updates[key] = values[0]
		}
	}

	// Обработка файлов
	for key, files := range r.MultipartForm.File {
		file, err := files[0].Open()
		if err != nil {
			http.Error(w, "Error opening file", http.StatusInternalServerError)
			return
		}
		defer file.Close()

		// Создание директории если нет
		uploadDir := "./uploads/avatars"
		if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
			http.Error(w, "Error creating upload directory", http.StatusInternalServerError)
			return
		}

		// Сохраняем файл
		filename := fmt.Sprintf("user_%d_%s", userID, files[0].Filename)
		path := filepath.Join(uploadDir, filename)

		dst, err := os.Create(path)
		if err != nil {
			http.Error(w, "Error saving file", http.StatusInternalServerError)
			return
		}
		defer dst.Close()
		if _, err := io.Copy(dst, file); err != nil {
			http.Error(w, "Error writing file", http.StatusInternalServerError)
			return
		}

		// Сохраняем URL (например: /static/avatars/user_123_file.png)
		avatarURL := fmt.Sprintf("localhost:8082/static/avatars/%s", filename)
		updates[key] = avatarURL
	}

	// Обновляем пользователя
	err = h.UserStorage.UpdateUser(r.Context(), userID, updates)
	if err != nil {
		http.Error(w, "Could not update user", http.StatusInternalServerError)
		return
	}

	userInfo, err := h.UserStorage.GetUserInfo(r.Context(), userID)
	if err != nil {
		http.Error(w, "Could not fetch user info", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(userInfo)
}

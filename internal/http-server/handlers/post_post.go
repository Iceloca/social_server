package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/google/uuid"
	"kursach/internal/storage/postgres"
)

type PostHandler struct {
	UserStorage *postgres.UserStorage
	PostStorage *postgres.PostStorage
}

func (h *PostHandler) AddPost(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(20 << 20); err != nil { // 20 MB max
		http.Error(w, "Could not parse form", http.StatusBadRequest)
		return
	}

	userIDStr := r.FormValue("userId")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "Invalid userId", http.StatusBadRequest)
		return
	}

	// Проверяем, есть ли пользователь
	_, err = h.UserStorage.GetUserInfo(r.Context(), userID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	title := r.FormValue("title")
	description := r.FormValue("description")

	var imageURL string
	// Обработка файла (если есть)
	fileHeaders := r.MultipartForm.File["file"]
	if len(fileHeaders) > 0 {
		file, err := fileHeaders[0].Open()
		if err != nil {
			http.Error(w, "Error opening file", http.StatusInternalServerError)
			return
		}
		defer file.Close()

		uploadDir := "./uploads/posts"
		if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
			http.Error(w, "Error creating upload directory", http.StatusInternalServerError)
			return
		}

		filename := fmt.Sprintf("post_%s_%s", uuid.NewString(), filepath.Base(fileHeaders[0].Filename))
		filepath := filepath.Join(uploadDir, filename)

		dst, err := os.Create(filepath)
		if err != nil {
			http.Error(w, "Error saving file", http.StatusInternalServerError)
			return
		}
		defer dst.Close()

		if _, err := io.Copy(dst, file); err != nil {
			http.Error(w, "Error writing file", http.StatusInternalServerError)
			return
		}

		imageURL = "localhost:8082/static/posts/" + filename
	}

	// Создаем пост
	newPost := postgres.Post{
		AuthorID:    userID,
		Title:       title,
		Description: description,
		ImageURL:    imageURL,
		CreatedAt:   time.Now(),
	}

	tags := r.MultipartForm.Value["tags"] // например: ["tag1", "tag2"]

	// Сохраняем пост (без тегов)
	err = h.PostStorage.CreatePost(r.Context(), &newPost)
	if err != nil {
		http.Error(w, "Could not create post", http.StatusInternalServerError)
		return
	}

	// Сохраняем теги и связи
	for _, tagName := range tags {
		tagID, err := h.PostStorage.GetOrCreateTag(r.Context(), tagName)
		if err != nil {
			http.Error(w, "Error processing tags", http.StatusInternalServerError)
			return
		}

		err = h.PostStorage.AddPostTag(r.Context(), newPost.ID, tagID)
		if err != nil {
			http.Error(w, "Error linking post and tag", http.StatusInternalServerError)
			return
		}
	}

	// Отдаем ответ с новым постом
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(newPost)
}

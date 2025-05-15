package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
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
type PostsResult struct {
	Posts   []postgres.PostResponse `json:"posts"`
	HasMore bool                    `json:"hasMore"`
}
type GetPostsRequest struct {
	StartIndex int  `json:"start_index"`
	Amount     int  `json:"amount"`
	UserID     *int `json:"user_id,omitempty"` // опционально
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
func (h *PostHandler) GetPostsHandler(w http.ResponseWriter, r *http.Request) {
	// Чтение query-параметров
	query := r.URL.Query()

	startIndexStr := query.Get("startIndex")
	amountStr := query.Get("amount")
	userIDStr := query.Get("userId") // может быть пустым

	// Парсим startIndex
	startIndex, err := strconv.Atoi(startIndexStr)
	if err != nil {
		http.Error(w, "Invalid startIndex", http.StatusBadRequest)
		return
	}

	// Парсим amount
	amount, err := strconv.Atoi(amountStr)
	if err != nil {
		http.Error(w, "Invalid amount", http.StatusBadRequest)
		return
	}

	// userId — опционален
	var userID *int
	if userIDStr != "" {
		uid, err := strconv.Atoi(userIDStr)
		if err != nil {
			http.Error(w, "Invalid userId", http.StatusBadRequest)
			return
		}
		userID = &uid
	}

	// Получаем посты
	posts, hasMore, err := h.PostStorage.GetPosts(r.Context(), userID, startIndex, amount)
	log.Println(err)
	if err != nil {
		http.Error(w, "Failed to get posts", http.StatusInternalServerError)
		return
	}

	// Формируем ответ
	resp := PostsResult{
		Posts:   posts,
		HasMore: hasMore,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *PostHandler) DeletePostHandler(w http.ResponseWriter, r *http.Request) {
	postIDStr := r.URL.Query().Get("post_id")
	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		http.Error(w, "Invalid post_id", http.StatusBadRequest)
		return
	}

	if err := h.PostStorage.DeletePost(r.Context(), postID); err != nil {
		http.Error(w, "Failed to delete post", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
func (h *PostHandler) GetFavoritePostsHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	startIndexStr := query.Get("startIndex")
	amountStr := query.Get("amount")
	userIDStr := query.Get("userId")

	startIndex, err := strconv.Atoi(startIndexStr)
	if err != nil {
		http.Error(w, "Invalid startIndex", http.StatusBadRequest)
		return
	}

	amount, err := strconv.Atoi(amountStr)
	if err != nil {
		http.Error(w, "Invalid amount", http.StatusBadRequest)
		return
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "Invalid userId", http.StatusBadRequest)
		return
	}

	posts, hasMore, err := h.PostStorage.GetFavoritePosts(r.Context(), userID, startIndex, amount)
	if err != nil {
		http.Error(w, "Failed to get favorite posts", http.StatusInternalServerError)
		return
	}

	resp := PostsResult{
		Posts:   posts,
		HasMore: hasMore,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
)

type FavoriteRequest struct {
	UserID int `json:"user_id"`
	PostID int `json:"post_id"`
}

func (h *PostHandler) AddToFavoritesHandler(w http.ResponseWriter, r *http.Request) {
	var req FavoriteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := h.PostStorage.AddToFavorites(r.Context(), req.UserID, req.PostID); err != nil {
		http.Error(w, "Failed to add to favorites", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *PostHandler) RemoveFromFavoritesHandler(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.URL.Query().Get("user_id")
	postIDStr := r.URL.Query().Get("post_id")

	userID, err1 := strconv.Atoi(userIDStr)
	postID, err2 := strconv.Atoi(postIDStr)
	if err1 != nil || err2 != nil {
		http.Error(w, "Invalid user_id or post_id", http.StatusBadRequest)
		return
	}

	if err := h.PostStorage.RemoveFromFavorites(r.Context(), userID, postID); err != nil {
		http.Error(w, "Failed to remove from favorites", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

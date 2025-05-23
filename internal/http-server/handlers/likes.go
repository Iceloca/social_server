package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
)

type LikeRequest struct {
	UserID int `json:"user_id"`
	PostID int `json:"post_id"`
}

func (h *PostHandler) AddLikeHandler(w http.ResponseWriter, r *http.Request) {
	var req LikeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := h.PostStorage.AddLike(r.Context(), req.UserID, req.PostID); err != nil {
		http.Error(w, "Could not add like", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *PostHandler) RemoveLikeHandler(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.URL.Query().Get("user_id")
	postIDStr := r.URL.Query().Get("post_id")

	userID, err1 := strconv.Atoi(userIDStr)
	postID, err2 := strconv.Atoi(postIDStr)
	if err1 != nil || err2 != nil {
		http.Error(w, "Invalid user_id or post_id", http.StatusBadRequest)
		return
	}

	if err := h.PostStorage.RemoveLike(r.Context(), userID, postID); err != nil {
		http.Error(w, "Could not remove like", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

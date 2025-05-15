package handlers

import (
	"encoding/json"
	"net/http"
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
	var req LikeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := h.PostStorage.RemoveLike(r.Context(), req.UserID, req.PostID); err != nil {
		http.Error(w, "Could not remove like", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

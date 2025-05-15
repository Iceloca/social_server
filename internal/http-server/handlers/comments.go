package handlers

import (
	"encoding/json"
	"kursach/internal/storage/postgres"
	"log"
	"net/http"
)

type CommentRequest struct {
	AuthorID int    `json:"author_id"`
	PostID   int    `json:"post_id"`
	Comment  string `json:"comment"`
}

func (h *PostHandler) AddCommentHandler(w http.ResponseWriter, r *http.Request) {
	var req CommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	comment := &postgres.Comment{
		AuthorID: req.AuthorID,
		PostID:   req.PostID,
		Text:     req.Comment,
	}

	if err := h.PostStorage.AddComment(r.Context(), comment); err != nil {
		http.Error(w, "Could not add comment", http.StatusInternalServerError)
		log.Println(err)
		log.Println(comment)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(comment)
}

type DeleteCommentRequest struct {
	CommentID int `json:"comment_id"`
	UserID    int `json:"user_id"`
}

func (h *PostHandler) DeleteCommentHandler(w http.ResponseWriter, r *http.Request) {
	var req DeleteCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := h.PostStorage.DeleteComment(r.Context(), req.CommentID, req.UserID); err != nil {
		http.Error(w, "Could not delete comment", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

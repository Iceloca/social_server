package handlers

import (
	"encoding/json"
	"net/http"
)

type CreateReportRequest struct {
	ReporterID  int    `json:"reporter_id"`
	Description string `json:"description"`
	PostID      *int   `json:"post_id,omitempty"`
	CommentID   *int   `json:"comment_id,omitempty"`
}

func (h *PostHandler) CreateReportHandler(w http.ResponseWriter, r *http.Request) {
	var req CreateReportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.PostID == nil && req.CommentID == nil {
		http.Error(w, "Either post_id or comment_id must be provided", http.StatusBadRequest)
		return
	}
	if req.PostID != nil && req.CommentID != nil {
		http.Error(w, "Only one of post_id or comment_id should be provided", http.StatusBadRequest)
		return
	}

	var targetID, reportTypeID int
	if req.PostID != nil {
		targetID = *req.PostID
		reportTypeID = 1 // post
	} else {
		targetID = *req.CommentID
		reportTypeID = 2 // comment
	}

	err := h.PostStorage.CreateReport(r.Context(), req.ReporterID, targetID, reportTypeID, req.Description)
	if err != nil {
		http.Error(w, "Failed to create report", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

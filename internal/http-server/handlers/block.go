package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
)

func (h *PostHandler) AddBlockHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		BlockerID int `json:"blocker_id"`
		BlockedID int `json:"blocked_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	err := h.PostStorage.AddUserBlock(r.Context(), req.BlockerID, req.BlockedID)
	if err != nil {
		http.Error(w, "Could not add block", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (h *PostHandler) CheckBlockHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	blockerIDs, ok1 := query["blocker_id"]
	blockedIDs, ok2 := query["blocked_id"]

	if !ok1 || !ok2 || len(blockerIDs) == 0 || len(blockedIDs) == 0 {
		http.Error(w, "Missing query parameters", http.StatusBadRequest)
		return
	}

	blockerID, err := strconv.Atoi(blockerIDs[0])
	if err != nil {
		http.Error(w, "Invalid blocker_id", http.StatusBadRequest)
		return
	}

	blockedID, err := strconv.Atoi(blockedIDs[0])
	if err != nil {
		http.Error(w, "Invalid blocked_id", http.StatusBadRequest)
		return
	}

	blocked, err := h.PostStorage.IsUserBlocked(r.Context(), blockerID, blockedID)
	if err != nil {
		http.Error(w, "Error checking block", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]bool{"blocked": blocked})
}

// В handlers
func (h *PostHandler) RemoveBlockHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	blockerIDs, ok1 := query["blocker_id"]
	blockedIDs, ok2 := query["blocked_id"]

	if !ok1 || !ok2 || len(blockerIDs) == 0 || len(blockedIDs) == 0 {
		http.Error(w, "Missing query parameters", http.StatusBadRequest)
		return
	}

	blockerID, err := strconv.Atoi(blockerIDs[0])
	if err != nil {
		http.Error(w, "Invalid blocker_id", http.StatusBadRequest)
		return
	}

	blockedID, err := strconv.Atoi(blockedIDs[0])
	if err != nil {
		http.Error(w, "Invalid blocked_id", http.StatusBadRequest)
		return
	}

	err = h.PostStorage.RemoveBlock(r.Context(), blockerID, blockedID)
	if err != nil {
		http.Error(w, "Error removing block", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

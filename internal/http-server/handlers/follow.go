package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
)

type FollowRequest struct {
	FollowerID  int `json:"follower_id"`
	FollowingID int `json:"following_id"`
}

func (h *PostHandler) AddFollowHandler(w http.ResponseWriter, r *http.Request) {
	var req FollowRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.FollowerID == 0 || req.FollowingID == 0 {
		http.Error(w, "follower_id and following_id are required", http.StatusBadRequest)
		return
	}

	if err := h.PostStorage.AddFollow(r.Context(), req.FollowerID, req.FollowingID); err != nil {
		http.Error(w, "Failed to add follow", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *PostHandler) RemoveFollowHandler(w http.ResponseWriter, r *http.Request) {
	followerID, err1 := strconv.Atoi(r.URL.Query().Get("follower_id"))
	followingID, err2 := strconv.Atoi(r.URL.Query().Get("following_id"))

	if err1 != nil || err2 != nil {
		http.Error(w, "Invalid follower_id or following_id", http.StatusBadRequest)
		return
	}

	if err := h.PostStorage.RemoveFollow(r.Context(), followerID, followingID); err != nil {
		http.Error(w, "Failed to remove follow", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *PostHandler) GetFollowingsHandler(w http.ResponseWriter, r *http.Request) {
	followerID, err := strconv.Atoi(r.URL.Query().Get("follower_id"))
	if err != nil {
		http.Error(w, "Invalid follower_id", http.StatusBadRequest)
		return
	}

	followings, err := h.PostStorage.GetUserFollowings(r.Context(), followerID)
	if err != nil {
		http.Error(w, "Failed to get followings", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(followings)
}

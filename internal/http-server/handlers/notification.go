package handlers

import (
	"encoding/json"
	"kursach/internal/storage/postgres"
	"net/http"
	"strconv"
)

type NotificationHandler struct {
	NotificationStorage *postgres.NotificationStorage
}

func (h *NotificationHandler) GetUnreadNotificationsHandler(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.URL.Query().Get("userId")
	if userIDStr == "" {
		http.Error(w, "Missing userId", http.StatusBadRequest)
		return
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "Invalid userId", http.StatusBadRequest)
		return
	}

	notifications, err := h.NotificationStorage.GetUnreadNotifications(r.Context(), userID)
	if err != nil {
		http.Error(w, "Failed to fetch notifications", http.StatusInternalServerError)
		return
	}
	err = h.NotificationStorage.MarkAllAsRead(r.Context(), userID)
	if err != nil {
		http.Error(w, "Failed to mark as read", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(notifications)
}

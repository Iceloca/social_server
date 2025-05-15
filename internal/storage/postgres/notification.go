package postgres

import (
	"context"
	"database/sql"
	"log"
	"time"
)

type NotificationStorage struct {
	db *sql.DB
}

type Notification struct {
	ID        int       `json:"notification_id"`
	UserID    int       `json:"user_id"`
	TypeID    int       `json:"type_id"`
	EntityID  int       `json:"entity_id"`
	IsRead    bool      `json:"is_read"`
	CreatedAt time.Time `json:"created_at"`
}

func NewNotificationStorage(db *sql.DB) *NotificationStorage {
	return &NotificationStorage{db: db}
}

func (s *NotificationStorage) GetUnreadNotifications(ctx context.Context, userID int) ([]Notification, error) {
	const query = `SELECT notification_id, user_id, type_id, entity_id, is_read, created_at FROM get_unread_notifications($1)`

	rows, err := s.db.QueryContext(ctx, query, userID)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer rows.Close()

	var notifications []Notification
	for rows.Next() {
		var n Notification
		err := rows.Scan(&n.ID, &n.UserID, &n.TypeID, &n.EntityID, &n.IsRead, &n.CreatedAt)
		if err != nil {
			return nil, err
		}
		log.Println(err)
		notifications = append(notifications, n)
	}

	return notifications, nil
}
func (s *NotificationStorage) MarkAllAsRead(ctx context.Context, userID int) error {
	const query = `CALL mark_all_notifications_as_read($1)`
	_, err := s.db.ExecContext(ctx, query, userID)
	return err
}

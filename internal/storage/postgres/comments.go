package postgres

import (
	"context"
	"errors"
	"time"
)

type Comment struct {
	ID        int       `json:"comment_id"`
	AuthorID  int       `json:"author_id"`
	PostID    int       `json:"post_id"`
	Text      string    `json:"comment"`
	CreatedAt time.Time `json:"created_at"`
}

func (s *PostStorage) AddComment(ctx context.Context, comment *Comment) error {
	const query = `
		INSERT INTO comments (author_id, post_id, comment)
		VALUES ($1, $2, $3)
		RETURNING comment_id, created_at
	`
	return s.db.QueryRowContext(ctx, query,
		comment.AuthorID,
		comment.PostID,
		comment.Text,
	).Scan(&comment.ID, &comment.CreatedAt)
}

func (s *PostStorage) DeleteComment(ctx context.Context, commentID int, userID int) error {
	const query = `
		DELETE FROM comments
		WHERE comment_id = $1 AND author_id = $2
	`
	res, err := s.db.ExecContext(ctx, query, commentID, userID)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return errors.New("no comment deleted")
	}
	return nil
}

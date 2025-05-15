package postgres

import (
	"context"
	"errors"
	"time"
)

type Comment struct {
	ID            int       `json:"comment_id"`
	AuthorID      int       `json:"author_id"`
	AuthorUserTag string    `json:"author_user_tag,omitempty"`
	PostID        int       `json:"post_id"`
	Text          string    `json:"comment"`
	CreatedAt     time.Time `json:"created_at"`
}

func (s *PostStorage) AddComment(ctx context.Context, comment *Comment) error {
	const query = `CALL comment_post($1, $2, $3)`
	_ = s.db.QueryRowContext(ctx, query, comment.AuthorID, comment.PostID, comment.Text)

	// Если нужно вернуть ID и CreatedAt, можно возвращать их из процедуры через OUT-параметры или
	// сразу после CALL делать SELECT:
	const getQuery = `
		SELECT comment_id, created_at
		FROM comments
		WHERE author_id = $1 AND post_id = $2
		ORDER BY created_at DESC
		LIMIT 1
	`
	return s.db.QueryRowContext(ctx, getQuery, comment.AuthorID, comment.PostID).
		Scan(&comment.ID, &comment.CreatedAt)
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

func (s *PostStorage) GetCommentsByPostID(ctx context.Context, postID int) ([]CommentBrief, error) {
	const query = `
		SELECT comment_id, post_id, comment, comment_created_at, author_id, author_user_tag
		FROM view_comment_with_author_tag
		WHERE post_id = $1
		ORDER BY comment_created_at ASC
	`

	rows, err := s.db.QueryContext(ctx, query, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []CommentBrief
	for rows.Next() {
		var c CommentBrief
		if err := rows.Scan(&c.CommentID, &c.PostID, &c.Comment, &c.CreatedAt, &c.AuthorID, &c.AuthorTag); err != nil {
			return nil, err
		}
		comments = append(comments, c)
	}

	return comments, nil
}

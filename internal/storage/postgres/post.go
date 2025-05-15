package postgres

import (
	"context"
	"database/sql"
	"time"
)

type PostStorage struct {
	db *sql.DB
}

type Post struct {
	ID          string    `json:"id"`
	AuthorID    int       `json:"author_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	ImageURL    string    `json:"image_url"`
	CreatedAt   time.Time `json:"created_at"`
}

func NewPostStorage(db *sql.DB) *PostStorage {
	return &PostStorage{db: db}
}

func (s *PostStorage) CreatePost(ctx context.Context, post *Post) error {
	const query = `
		INSERT INTO posts (author_id, title, description, image_url, created_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING post_id
	`
	return s.db.QueryRowContext(ctx, query,
		post.AuthorID,
		post.Title,
		post.Description,
		post.ImageURL,
		post.CreatedAt,
	).Scan(&post.ID)
}
func (s *PostStorage) GetOrCreateTag(ctx context.Context, name string) (int, error) {
	var tagID int
	err := s.db.QueryRowContext(ctx, "SELECT tag_id FROM tags WHERE name=$1", name).Scan(&tagID)
	if err == sql.ErrNoRows {
		err = s.db.QueryRowContext(ctx, "INSERT INTO tags (name) VALUES ($1) RETURNING tag_id", name).Scan(&tagID)
	}
	return tagID, err
}

func (s *PostStorage) AddPostTag(ctx context.Context, postID string, tagID int) error {
	_, err := s.db.ExecContext(ctx, "INSERT INTO post_tags (post_id, tag_id) VALUES ($1, $2)", postID, tagID)
	return err
}

func (s *PostStorage) AddLike(ctx context.Context, userID, postID int) error {
	const query = `
		INSERT INTO likes (user_id, post_id)
		VALUES ($1, $2)
		ON CONFLICT DO NOTHING
	`
	_, err := s.db.ExecContext(ctx, query, userID, postID)
	return err
}

func (s *PostStorage) RemoveLike(ctx context.Context, userID, postID int) error {
	const query = `
		DELETE FROM likes
		WHERE user_id = $1 AND post_id = $2
	`
	_, err := s.db.ExecContext(ctx, query, userID, postID)
	return err
}

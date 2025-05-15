package postgres

import (
	"context"
	"database/sql"
	"log"
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

type PostResponse struct {
	PostID      int            `json:"post_id"`
	Title       string         `json:"title"`
	Description string         `json:"description"`
	ImageURL    string         `json:"image_url"`
	CreatedAt   time.Time      `json:"created_at"`
	AuthorID    int            `json:"author_id"`
	AuthorName  string         `json:"author_name"`
	LikeCount   int            `json:"like_count"`
	Comments    []CommentBrief `json:"comments"`
	Tags        []TagBrief     `json:"tags"`
}

type CommentBrief struct {
	CommentID int       `json:"comment_id"`
	PostID    int       `json:"post_id"`
	Comment   string    `json:"comment"`
	CreatedAt time.Time `json:"created_at"`
	AuthorID  int       `json:"author_id"`
	AuthorTag string    `json:"author_tag"`
}

type TagBrief struct {
	TagID int    `json:"tag_id"`
	Name  string `json:"name"`
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

func (s *PostStorage) GetTagsByPostID(ctx context.Context, postID int) ([]TagBrief, error) {
	const query = `
		SELECT t.tag_id, t.name
		FROM post_tags pt
		JOIN tags t ON t.tag_id = pt.tag_id
		WHERE pt.post_id = $1
	`

	rows, err := s.db.QueryContext(ctx, query, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []TagBrief
	for rows.Next() {
		var t TagBrief
		if err := rows.Scan(&t.TagID, &t.Name); err != nil {
			return nil, err
		}
		tags = append(tags, t)
	}

	return tags, nil
}

func (s *PostStorage) GetPosts(ctx context.Context, userID *int, startIndex, amount int) ([]PostResponse, bool, error) {
	var (
		posts []PostResponse
		rows  *sql.Rows
		err   error
	)
	log.Printf("Fetching posts: startIndex=%d, amount=%d, userID=%d", startIndex, amount, userID)

	baseQuery := `
		SELECT post_id, title, description, image_url, post_created_at, author_id, author_user_name, like_count
		FROM view_post_summary
	`
	args := []interface{}{}

	if userID != nil {
		baseQuery += ` WHERE author_id = $1 ORDER BY post_created_at DESC LIMIT $2 OFFSET $3`
		args = append(args, *userID, amount, startIndex)
	} else {
		baseQuery += ` ORDER BY post_created_at DESC LIMIT $1 OFFSET $2`
		args = append(args, amount, startIndex)
	}

	rows, err = s.db.QueryContext(ctx, baseQuery, args...)
	if err != nil {
		return nil, false, err
	}
	defer rows.Close()

	for rows.Next() {
		var post PostResponse
		if err := rows.Scan(
			&post.PostID, &post.Title, &post.Description, &post.ImageURL,
			&post.CreatedAt, &post.AuthorID, &post.AuthorName, &post.LikeCount,
		); err != nil {
			return nil, false, err
		}

		// Комментарии
		comments, err := s.GetCommentsByPostID(ctx, post.PostID)
		if err != nil {
			return nil, false, err
		}
		post.Comments = comments

		// Теги
		tags, err := s.GetTagsByPostID(ctx, post.PostID)
		if err != nil {
			return nil, false, err
		}
		post.Tags = tags

		posts = append(posts, post)
	}

	// Проверка, есть ли еще посты
	var totalCount int
	countQuery := "SELECT COUNT(*) FROM view_post_summary"
	if userID != nil {
		countQuery += " WHERE author_id = $1"
		err = s.db.QueryRowContext(ctx, countQuery, *userID).Scan(&totalCount)
	} else {
		err = s.db.QueryRowContext(ctx, countQuery).Scan(&totalCount)
	}
	if err != nil {
		return nil, false, err
	}

	hasMore := startIndex+amount < totalCount

	return posts, hasMore, nil
}

func (s *PostStorage) DeletePost(ctx context.Context, postID int) error {
	const query = `DELETE FROM posts WHERE post_id = $1`
	_, err := s.db.ExecContext(ctx, query, postID)
	return err
}

func (s *PostStorage) GetFavoritePosts(ctx context.Context, userID, startIndex, amount int) ([]PostResponse, bool, error) {
	const baseQuery = `
		SELECT post_id, title, description, image_url, post_created_at, author_id, author_user_name, like_count
		FROM view_favorite_post_summary
		WHERE favorited_by_user_id = $1
		ORDER BY post_created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := s.db.QueryContext(ctx, baseQuery, userID, amount, startIndex)
	if err != nil {
		return nil, false, err
	}
	defer rows.Close()

	var posts []PostResponse
	for rows.Next() {
		var post PostResponse
		if err := rows.Scan(
			&post.PostID, &post.Title, &post.Description, &post.ImageURL,
			&post.CreatedAt, &post.AuthorID, &post.AuthorName, &post.LikeCount,
		); err != nil {
			return nil, false, err
		}

		comments, err := s.GetCommentsByPostID(ctx, post.PostID)
		if err != nil {
			return nil, false, err
		}
		post.Comments = comments

		tags, err := s.GetTagsByPostID(ctx, post.PostID)
		if err != nil {
			return nil, false, err
		}
		post.Tags = tags

		posts = append(posts, post)
	}

	var totalCount int
	err = s.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM view_favorite_post_summary WHERE favorited_by_user_id = $1`, userID,
	).Scan(&totalCount)
	if err != nil {
		return nil, false, err
	}

	hasMore := startIndex+amount < totalCount

	return posts, hasMore, nil
}

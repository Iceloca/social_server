package postgres

import (
	"context"
	"log"
)

func (s *PostStorage) AddToFavorites(ctx context.Context, userID, postID int) error {
	const query = `
		INSERT INTO favorite_posts (user_id, post_id)
		VALUES ($1, $2)
		ON CONFLICT DO NOTHING
	`
	_, err := s.db.ExecContext(ctx, query, userID, postID)
	log.Println(err)
	return err
}

func (s *PostStorage) RemoveFromFavorites(ctx context.Context, userID, postID int) error {
	const query = `
		DELETE FROM favorite_posts
		WHERE user_id = $1 AND post_id = $2
	`
	_, err := s.db.ExecContext(ctx, query, userID, postID)
	return err
}

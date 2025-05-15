package postgres

import "context"

func (s *PostStorage) AddLike(ctx context.Context, userID, postID int) error {
	const query = `CALL like_post($1, $2)`
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

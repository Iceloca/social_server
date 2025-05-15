package postgres

import "context"

func (s *PostStorage) AddUserBlock(ctx context.Context, blockerID, blockedID int) error {
	const query = `
        INSERT INTO user_blocks (blocker_id, blocked_id)
        VALUES ($1, $2)
        ON CONFLICT DO NOTHING
    `
	_, err := s.db.ExecContext(ctx, query, blockerID, blockedID)
	return err
}

func (s *PostStorage) IsUserBlocked(ctx context.Context, blockerID, blockedID int) (bool, error) {
	const query = `
        SELECT EXISTS (
            SELECT 1 FROM user_blocks WHERE blocker_id = $1 AND blocked_id = $2
        )
    `
	var exists bool
	err := s.db.QueryRowContext(ctx, query, blockerID, blockedID).Scan(&exists)
	return exists, err
}
func (s *PostStorage) RemoveBlock(ctx context.Context, blockerID, blockedID int) error {
	const query = `
		DELETE FROM user_blocks
		WHERE blocker_id = $1 AND blocked_id = $2
	`
	_, err := s.db.ExecContext(ctx, query, blockerID, blockedID)
	return err
}

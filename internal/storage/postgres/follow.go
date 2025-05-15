package postgres

import "context"

func (s *PostStorage) AddFollow(ctx context.Context, followerID, followingID int) error {
	const query = `
		CALL follow($1, $2)
	`
	_, err := s.db.ExecContext(ctx, query, followerID, followingID)
	return err
}
func (s *PostStorage) RemoveFollow(ctx context.Context, followerID, followingID int) error {
	const query = `
		DELETE FROM follows
		WHERE follower_id = $1 AND following_id = $2
	`
	_, err := s.db.ExecContext(ctx, query, followerID, followingID)
	return err
}
func (s *PostStorage) GetUserFollowings(ctx context.Context, followerID int) ([]int, error) {
	const query = `
		SELECT following_id
		FROM follows
		WHERE follower_id = $1
	`
	rows, err := s.db.QueryContext(ctx, query, followerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var followings []int
	for rows.Next() {
		var followingID int
		if err := rows.Scan(&followingID); err != nil {
			return nil, err
		}
		followings = append(followings, followingID)
	}
	return followings, nil
}

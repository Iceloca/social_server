package postgres

import "context"

func (s *PostStorage) CreateReport(ctx context.Context, reporterID, targetID, reportTypeID int, description string) error {
	const query = `
		INSERT INTO reports (reporter_id, target_id, report_type_id, description)
		VALUES ($1, $2, $3, $4)
	`
	_, err := s.db.ExecContext(ctx, query, reporterID, targetID, reportTypeID, description)
	return err
}

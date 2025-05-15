package postgres

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"kursach/internal/config"
)

type Storage struct {
	db *sql.DB
}

func (s *Storage) DB() *sql.DB {
	return s.db
}

func New(cfg config.PostgresCfg) (*Storage, error) {
	const op = "storage.postgres.new"

	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// Проверка соединения
	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("%s: unable to ping db: %w", op, err)
	}

	return &Storage{db: db}, nil
}

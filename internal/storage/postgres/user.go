package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type UserStorage struct {
	db *sql.DB
}

func NewUserStorage(db *sql.DB) *UserStorage {
	return &UserStorage{db: db}
}

func (s *UserStorage) IsUserTagTaken(ctx context.Context, tag string) (bool, error) {
	const query = `SELECT 1 FROM user_info WHERE user_tag = $1 LIMIT 1`
	var dummy int
	err := s.db.QueryRowContext(ctx, query, tag).Scan(&dummy)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("checking tag: %w", err)
	}
	return true, nil
}

func (s *UserStorage) CreateFullUser(ctx context.Context, email, password, name, tag string) error {
	const query = `CALL create_full_user($1, $2, $3, $4)`
	_, err := s.db.ExecContext(ctx, query, email, password, name, tag)
	if err != nil {
		return fmt.Errorf("creating user: %w", err)
	}
	return nil
}

// Запись токена в таблицу user_tokens
func (s *UserStorage) SaveUserToken(ctx context.Context, userID int, token string, expiresAt time.Time) error {
	const query = `INSERT INTO user_tokens (token, user_id, expires_at) VALUES ($1, $2, $3)`
	_, err := s.db.ExecContext(ctx, query, token, userID, expiresAt)
	if err != nil {
		return fmt.Errorf("saving token: %w", err)
	}
	return nil
}

// Получение информации о пользователе по user_id
func (s *UserStorage) GetUserInfo(ctx context.Context, userID int) (map[string]interface{}, error) {
	const query = `SELECT user_name, user_tag, theme, language, avatar_url, description FROM user_info WHERE user_id = $1`
	row := s.db.QueryRowContext(ctx, query, userID)

	userInfo := make(map[string]interface{})

	// Declare variables to hold the column data
	var userName, theme, userTag, language string
	var avatarURL, description sql.NullString
	// Scan the row into variables
	err := row.Scan(&userName, &userTag, &theme, &language, &avatarURL, &description)
	if err != nil {
		if err == sql.ErrNoRows {
			// Логируем отсутствие записи
			fmt.Printf("No user found with user_tag: %s\n", userTag)
			return nil, fmt.Errorf("user not found: %w", err)
		}
		// Логируем другие ошибки
		fmt.Printf("Error while scanning user data for user_tag %s: %v\n", userTag, err)
		return nil, fmt.Errorf("fetching user info by tag: %w", err)
	}
	if avatarURL.Valid {
		userInfo["avatar_url"] = avatarURL.String
	} else {
		userInfo["avatar_url"] = ""
	}

	if description.Valid {
		userInfo["description"] = description.String
	} else {
		userInfo["description"] = ""
	}
	// Заполняем map с полученными данными
	userInfo["user_id"] = userID
	userInfo["user_name"] = userName
	userInfo["user_tag"] = userTag
	userInfo["theme"] = theme
	userInfo["language"] = language

	// Выводим логи, чтобы удостовериться, что данные получены
	fmt.Printf("Retrieved user info: %+v\n", userInfo)

	return userInfo, nil
}

func (s *UserStorage) GetUserInfoByTag(ctx context.Context, userTag string) (map[string]interface{}, error) {
	const query = `SELECT user_id, user_name, user_tag, theme, language, avatar_url, description FROM user_info WHERE user_tag = $1`
	row := s.db.QueryRowContext(ctx, query, userTag)

	userInfo := make(map[string]interface{})

	// Declare variables to hold the column data
	var userID int
	var userName, theme, language string
	var avatarURL, description sql.NullString
	// Scan the row into variables
	err := row.Scan(&userID, &userName, &userTag, &theme, &language, &avatarURL, &description)
	if err != nil {
		if err == sql.ErrNoRows {
			// Логируем отсутствие записи
			fmt.Printf("No user found with user_tag: %s\n", userTag)
			return nil, fmt.Errorf("user not found: %w", err)
		}
		// Логируем другие ошибки
		fmt.Printf("Error while scanning user data for user_tag %s: %v\n", userTag, err)
		return nil, fmt.Errorf("fetching user info by tag: %w", err)
	}
	if avatarURL.Valid {
		userInfo["avatar_url"] = avatarURL.String
	} else {
		userInfo["avatar_url"] = ""
	}

	if description.Valid {
		userInfo["description"] = description.String
	} else {
		userInfo["description"] = ""
	}
	// Заполняем map с полученными данными
	userInfo["user_id"] = userID
	userInfo["user_name"] = userName
	userInfo["user_tag"] = userTag
	userInfo["theme"] = theme
	userInfo["language"] = language

	// Выводим логи, чтобы удостовериться, что данные получены
	fmt.Printf("Retrieved user info: %+v\n", userInfo)

	return userInfo, nil
}

func (s *UserStorage) GetUserByEmail(ctx context.Context, email string) (int, string, error) {
	const query = "SELECT user_id, password FROM users WHERE email = $1"
	var userID int
	var hashedPassword string
	err := s.db.QueryRowContext(ctx, query, email).Scan(&userID, &hashedPassword)
	if err == sql.ErrNoRows {
		return 0, "", fmt.Errorf("user not found")
	}
	if err != nil {
		return 0, "", fmt.Errorf("query error: %w", err)
	}
	return userID, hashedPassword, nil
}

func (s *UserStorage) IsEmailTaken(ctx context.Context, email string) (bool, error) {
	const query = `SELECT 1 FROM users WHERE email = $1 LIMIT 1`
	var dummy int
	err := s.db.QueryRowContext(ctx, query, email).Scan(&dummy)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("checking email: %w", err)
	}
	return true, nil
}

func (s *UserStorage) UpdateUser(ctx context.Context, userID int, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return nil
	}

	query := "UPDATE user_info SET "
	args := []interface{}{}
	i := 1

	for key, value := range updates {
		if i > 1 {
			query += ", "
		}
		query += fmt.Sprintf("%s = $%d", key, i)
		args = append(args, value)
		i++
	}

	query += fmt.Sprintf(" WHERE user_id = $%d", i)
	args = append(args, userID)

	_, err := s.db.ExecContext(ctx, query, args...)
	return err
}

package postgres

import (
	"context"
	"database/sql"
	"errors"

	_ "github.com/lib/pq"
)

var ErrNotFound = errors.New("not found")

type Store struct {
	db *sql.DB
}

func (s *Store) DB() *sql.DB {
	return s.db
}

func NewStore(databaseURL string) (*Store, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return &Store{db: db}, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

type User struct {
	ID           int64  `json:"id"`
	Email        string `json:"email"`
	PasswordHash string `json:"-"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

type UserSettings struct {
	UserID            int64  `json:"user_id"`
	Timezone          string `json:"timezone"`
	AvgWindowDays     int    `json:"avg_window_days"`
	DefaultIncludeAvg bool   `json:"default_include_avg"`
}

func (s *Store) CreateUser(ctx context.Context, email, passwordHash string) (*User, error) {
	query := `
		INSERT INTO users (email, password_hash, created_at, updated_at)
		VALUES ($1, $2, NOW(), NOW())
		RETURNING id, email, password_hash, created_at, updated_at
	`
	user := &User{}
	err := s.db.QueryRowContext(ctx, query, email, passwordHash).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *Store) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	query := `
		SELECT id, email, password_hash, created_at, updated_at
		FROM users
		WHERE email = $1
	`
	user := &User{}
	err := s.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *Store) GetUserByID(ctx context.Context, id int64) (*User, error) {
	query := `
		SELECT id, email, password_hash, created_at, updated_at
		FROM users
		WHERE id = $1
	`
	user := &User{}
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *Store) CreateUserSettings(ctx context.Context, userID int64) (*UserSettings, error) {
	query := `
		INSERT INTO user_settings (user_id, timezone, avg_window_days, default_include_avg, created_at, updated_at)
		VALUES ($1, 'UTC', 30, false, NOW(), NOW())
		ON CONFLICT (user_id) DO UPDATE SET 
			timezone = EXCLUDED.timezone,
			avg_window_days = EXCLUDED.avg_window_days,
			default_include_avg = EXCLUDED.default_include_avg
		RETURNING user_id, timezone, avg_window_days, default_include_avg
	`
	settings := &UserSettings{}
	err := s.db.QueryRowContext(ctx, query, userID).Scan(
		&settings.UserID, &settings.Timezone, &settings.AvgWindowDays, &settings.DefaultIncludeAvg,
	)
	if err != nil {
		return nil, err
	}
	return settings, nil
}

func (s *Store) GetUserSettings(ctx context.Context, userID int64) (*UserSettings, error) {
	query := `
		SELECT user_id, timezone, avg_window_days, default_include_avg
		FROM user_settings
		WHERE user_id = $1
	`
	settings := &UserSettings{}
	err := s.db.QueryRowContext(ctx, query, userID).Scan(
		&settings.UserID, &settings.Timezone, &settings.AvgWindowDays, &settings.DefaultIncludeAvg,
	)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return settings, nil
}

func (s *Store) UpdateUserSettings(ctx context.Context, userID int64, timezone string, avgWindowDays int, defaultIncludeAvg bool) (*UserSettings, error) {
	query := `
		INSERT INTO user_settings (user_id, timezone, avg_window_days, default_include_avg)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id) DO UPDATE SET 
			timezone = EXCLUDED.timezone,
			avg_window_days = EXCLUDED.avg_window_days,
			default_include_avg = EXCLUDED.default_include_avg
		RETURNING user_id, timezone, avg_window_days, default_include_avg
	`
	settings := &UserSettings{}
	err := s.db.QueryRowContext(ctx, query, userID, timezone, avgWindowDays, defaultIncludeAvg).Scan(
		&settings.UserID, &settings.Timezone, &settings.AvgWindowDays, &settings.DefaultIncludeAvg,
	)
	if err != nil {
		return nil, err
	}
	return settings, nil
}

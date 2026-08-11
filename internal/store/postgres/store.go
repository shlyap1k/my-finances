package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

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

type IncomeRule struct {
	ID          int64  `json:"id"`
	UserID      int64  `json:"user_id"`
	Name        string `json:"name"`
	AmountMinor int64  `json:"amount_minor"`
	MonthlyDay  int    `json:"monthly_day"`
	StartDate   string `json:"start_date"`
	EndDate     *string `json:"end_date,omitempty"`
	Active      bool   `json:"active"`
	Note        *string `json:"note,omitempty"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

func (s *Store) CreateIncomeRule(ctx context.Context, userID int64, name string, amountMinor int64, monthlyDay int, startDate string, endDate *string, active bool, note *string) (*IncomeRule, error) {
	query := `
		INSERT INTO income_rules (user_id, name, amount_minor, monthly_day, start_date, end_date, active, note, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())
		RETURNING id, user_id, name, amount_minor, monthly_day, start_date, end_date, active, note, created_at, updated_at
	`
	rule := &IncomeRule{}
	err := s.db.QueryRowContext(ctx, query, userID, name, amountMinor, monthlyDay, startDate, endDate, active, note).Scan(
		&rule.ID, &rule.UserID, &rule.Name, &rule.AmountMinor, &rule.MonthlyDay,
		&rule.StartDate, &rule.EndDate, &rule.Active, &rule.Note, &rule.CreatedAt, &rule.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return rule, nil
}

func (s *Store) GetIncomeRules(ctx context.Context, userID int64, active *bool) ([]*IncomeRule, error) {
	query := `
		SELECT id, user_id, name, amount_minor, monthly_day, start_date, end_date, active, note, created_at, updated_at
		FROM income_rules
		WHERE user_id = $1
	`
	args := []interface{}{userID}
	argIndex := 2
	
	if active != nil {
		query += " AND active = $" + fmt.Sprint(argIndex)
		args = append(args, *active)
		argIndex++
	}
	
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var rules []*IncomeRule
	for rows.Next() {
		rule := &IncomeRule{}
		err := rows.Scan(
			&rule.ID, &rule.UserID, &rule.Name, &rule.AmountMinor, &rule.MonthlyDay,
			&rule.StartDate, &rule.EndDate, &rule.Active, &rule.Note, &rule.CreatedAt, &rule.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		rules = append(rules, rule)
	}
	
	if err := rows.Err(); err != nil {
		return nil, err
	}
	
	return rules, nil
}

func (s *Store) GetIncomeRuleByID(ctx context.Context, userID int64, id int64) (*IncomeRule, error) {
	query := `
		SELECT id, user_id, name, amount_minor, monthly_day, start_date, end_date, active, note, created_at, updated_at
		FROM income_rules
		WHERE id = $1 AND user_id = $2
	`
	rule := &IncomeRule{}
	err := s.db.QueryRowContext(ctx, query, id, userID).Scan(
		&rule.ID, &rule.UserID, &rule.Name, &rule.AmountMinor, &rule.MonthlyDay,
		&rule.StartDate, &rule.EndDate, &rule.Active, &rule.Note, &rule.CreatedAt, &rule.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return rule, nil
}

func (s *Store) UpdateIncomeRule(ctx context.Context, userID int64, id int64, name string, amountMinor int64, monthlyDay int, startDate string, endDate *string, active bool, note *string) (*IncomeRule, error) {
	query := `
		UPDATE income_rules
		SET name = $3, amount_minor = $4, monthly_day = $5, start_date = $6, end_date = $7, active = $8, note = $9, updated_at = NOW()
		WHERE id = $1 AND user_id = $2
		RETURNING id, user_id, name, amount_minor, monthly_day, start_date, end_date, active, note, created_at, updated_at
	`
	rule := &IncomeRule{}
	err := s.db.QueryRowContext(ctx, query, id, userID, name, amountMinor, monthlyDay, startDate, endDate, active, note).Scan(
		&rule.ID, &rule.UserID, &rule.Name, &rule.AmountMinor, &rule.MonthlyDay,
		&rule.StartDate, &rule.EndDate, &rule.Active, &rule.Note, &rule.CreatedAt, &rule.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return rule, nil
}

func (s *Store) DeleteIncomeRule(ctx context.Context, userID int64, id int64) error {
	query := `
		DELETE FROM income_rules
		WHERE id = $1 AND user_id = $2
	`
	result, err := s.db.ExecContext(ctx, query, id, userID)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

type ExpenseRule struct {
	ID          int64   `json:"id"`
	UserID      int64   `json:"user_id"`
	Name        string  `json:"name"`
	AmountMinor int64   `json:"amount_minor"`
	MonthlyDay  int     `json:"monthly_day"`
	StartDate   string  `json:"start_date"`
	EndDate     *string `json:"end_date,omitempty"`
	Active      bool    `json:"active"`
	Note        *string `json:"note,omitempty"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

func (s *Store) CreateExpenseRule(ctx context.Context, userID int64, name string, amountMinor int64, monthlyDay int, startDate string, endDate *string, active bool, note *string) (*ExpenseRule, error) {
	query := `
		INSERT INTO expense_rules (user_id, name, amount_minor, monthly_day, start_date, end_date, active, note, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())
		RETURNING id, user_id, name, amount_minor, monthly_day, start_date, end_date, active, note, created_at, updated_at
	`
	rule := &ExpenseRule{}
	err := s.db.QueryRowContext(ctx, query, userID, name, amountMinor, monthlyDay, startDate, endDate, active, note).Scan(
		&rule.ID, &rule.UserID, &rule.Name, &rule.AmountMinor, &rule.MonthlyDay,
		&rule.StartDate, &rule.EndDate, &rule.Active, &rule.Note, &rule.CreatedAt, &rule.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return rule, nil
}

func (s *Store) GetExpenseRules(ctx context.Context, userID int64, active *bool) ([]*ExpenseRule, error) {
	query := `
		SELECT id, user_id, name, amount_minor, monthly_day, start_date, end_date, active, note, created_at, updated_at
		FROM expense_rules
		WHERE user_id = $1
	`
	args := []interface{}{userID}
	argIndex := 2
	
	if active != nil {
		query += " AND active = $" + fmt.Sprint(argIndex)
		args = append(args, *active)
		argIndex++
	}
	
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var rules []*ExpenseRule
	for rows.Next() {
		rule := &ExpenseRule{}
		err := rows.Scan(
			&rule.ID, &rule.UserID, &rule.Name, &rule.AmountMinor, &rule.MonthlyDay,
			&rule.StartDate, &rule.EndDate, &rule.Active, &rule.Note, &rule.CreatedAt, &rule.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		rules = append(rules, rule)
	}
	
	if err := rows.Err(); err != nil {
		return nil, err
	}
	
	return rules, nil
}

func (s *Store) GetExpenseRuleByID(ctx context.Context, userID int64, id int64) (*ExpenseRule, error) {
	query := `
		SELECT id, user_id, name, amount_minor, monthly_day, start_date, end_date, active, note, created_at, updated_at
		FROM expense_rules
		WHERE id = $1 AND user_id = $2
	`
	rule := &ExpenseRule{}
	err := s.db.QueryRowContext(ctx, query, id, userID).Scan(
		&rule.ID, &rule.UserID, &rule.Name, &rule.AmountMinor, &rule.MonthlyDay,
		&rule.StartDate, &rule.EndDate, &rule.Active, &rule.Note, &rule.CreatedAt, &rule.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return rule, nil
}

func (s *Store) UpdateExpenseRule(ctx context.Context, userID int64, id int64, name string, amountMinor int64, monthlyDay int, startDate string, endDate *string, active bool, note *string) (*ExpenseRule, error) {
	query := `
		UPDATE expense_rules
		SET name = $3, amount_minor = $4, monthly_day = $5, start_date = $6, end_date = $7, active = $8, note = $9, updated_at = NOW()
		WHERE id = $1 AND user_id = $2
		RETURNING id, user_id, name, amount_minor, monthly_day, start_date, end_date, active, note, created_at, updated_at
	`
	rule := &ExpenseRule{}
	err := s.db.QueryRowContext(ctx, query, id, userID, name, amountMinor, monthlyDay, startDate, endDate, active, note).Scan(
		&rule.ID, &rule.UserID, &rule.Name, &rule.AmountMinor, &rule.MonthlyDay,
		&rule.StartDate, &rule.EndDate, &rule.Active, &rule.Note, &rule.CreatedAt, &rule.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return rule, nil
}

func (s *Store) DeleteExpenseRule(ctx context.Context, userID int64, id int64) error {
	query := `
		DELETE FROM expense_rules
		WHERE id = $1 AND user_id = $2
	`
	result, err := s.db.ExecContext(ctx, query, id, userID)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

type SavingsAccount struct {
	ID        int64  `json:"id"`
	UserID    int64  `json:"user_id"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func (s *Store) CreateSavingsAccount(ctx context.Context, userID int64, name string) (*SavingsAccount, error) {
	query := `
		INSERT INTO savings_accounts (user_id, name, created_at, updated_at)
		VALUES ($1, $2, NOW(), NOW())
		RETURNING id, user_id, name, created_at, updated_at
	`
	account := &SavingsAccount{}
	err := s.db.QueryRowContext(ctx, query, userID, name).Scan(
		&account.ID, &account.UserID, &account.Name, &account.CreatedAt, &account.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return account, nil
}

func (s *Store) GetSavingsAccounts(ctx context.Context, userID int64) ([]*SavingsAccount, error) {
	query := `
		SELECT id, user_id, name, created_at, updated_at
		FROM savings_accounts
		WHERE user_id = $1
	`
	rows, err := s.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []*SavingsAccount
	for rows.Next() {
		account := &SavingsAccount{}
		err := rows.Scan(
			&account.ID, &account.UserID, &account.Name, &account.CreatedAt, &account.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, account)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return accounts, nil
}

func (s *Store) GetSavingsAccountByID(ctx context.Context, userID int64, id int64) (*SavingsAccount, error) {
	query := `
		SELECT id, user_id, name, created_at, updated_at
		FROM savings_accounts
		WHERE id = $1 AND user_id = $2
	`
	account := &SavingsAccount{}
	err := s.db.QueryRowContext(ctx, query, id, userID).Scan(
		&account.ID, &account.UserID, &account.Name, &account.CreatedAt, &account.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return account, nil
}

func (s *Store) UpdateSavingsAccount(ctx context.Context, userID int64, id int64, name string) (*SavingsAccount, error) {
	query := `
		UPDATE savings_accounts
		SET name = $3, updated_at = NOW()
		WHERE id = $1 AND user_id = $2
		RETURNING id, user_id, name, created_at, updated_at
	`
	account := &SavingsAccount{}
	err := s.db.QueryRowContext(ctx, query, id, userID, name).Scan(
		&account.ID, &account.UserID, &account.Name, &account.CreatedAt, &account.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return account, nil
}

func (s *Store) DeleteSavingsAccount(ctx context.Context, userID int64, id int64) error {
	query := `
		DELETE FROM savings_accounts
		WHERE id = $1 AND user_id = $2
	`
	result, err := s.db.ExecContext(ctx, query, id, userID)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

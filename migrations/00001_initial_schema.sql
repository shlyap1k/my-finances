-- +goose Up
-- +goose StatementBegin

-- Users table
CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- User settings table
CREATE TABLE IF NOT EXISTS user_settings (
    user_id BIGINT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    timezone TEXT NOT NULL DEFAULT 'UTC',
    avg_window_days INTEGER NOT NULL DEFAULT 30 CHECK (avg_window_days >= 1 AND avg_window_days <= 365),
    default_include_avg BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

-- Income rules table (regular incomes)
CREATE TABLE IF NOT EXISTS income_rules (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    amount_minor BIGINT NOT NULL CHECK (amount_minor > 0),
    monthly_day INTEGER NOT NULL CHECK (monthly_day >= 1 AND monthly_day <= 31),
    start_date DATE NOT NULL,
    end_date DATE,
    active BOOLEAN NOT NULL DEFAULT TRUE,
    note TEXT,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

-- Expense rules table (mandatory regular expenses)
CREATE TABLE IF NOT EXISTS expense_rules (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    amount_minor BIGINT NOT NULL CHECK (amount_minor > 0),
    monthly_day INTEGER NOT NULL CHECK (monthly_day >= 1 AND monthly_day <= 31),
    start_date DATE NOT NULL,
    end_date DATE,
    active BOOLEAN NOT NULL DEFAULT TRUE,
    note TEXT,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

-- Savings accounts table
CREATE TABLE IF NOT EXISTS savings_accounts (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

-- Savings rules table
CREATE TABLE IF NOT EXISTS savings_rules (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    income_rule_id BIGINT REFERENCES income_rules(id) ON DELETE CASCADE,
    savings_account_id BIGINT NOT NULL REFERENCES savings_accounts(id) ON DELETE CASCADE,
    percent_bp INTEGER, -- percentage in basis points (e.g., 100 = 1%)
    fixed_amount_minor BIGINT CHECK (fixed_amount_minor >= 0),
    priority INTEGER NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

-- Transactions table
CREATE TABLE IF NOT EXISTS transactions (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type TEXT NOT NULL CHECK (type IN ('income', 'expense', 'savings_deposit', 'savings_withdraw')),
    status TEXT NOT NULL CHECK (status IN ('planned', 'actual')),
    origin TEXT NOT NULL CHECK (origin IN ('manual', 'regular_income', 'savings_rule', 'obligation_confirmation')),
    amount_minor BIGINT NOT NULL,
    occurred_on DATE NOT NULL,
    source_rule_occurred_on DATE,
    exclude_from_average BOOLEAN NOT NULL DEFAULT FALSE,
    savings_account_id BIGINT REFERENCES savings_accounts(id) ON DELETE SET NULL,
    note TEXT,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_income_rules_user_id ON income_rules(user_id);
CREATE INDEX IF NOT EXISTS idx_expense_rules_user_id ON expense_rules(user_id);
CREATE INDEX IF NOT EXISTS idx_savings_accounts_user_id ON savings_accounts(user_id);
CREATE INDEX IF NOT EXISTS idx_savings_rules_user_id ON savings_rules(user_id);
CREATE INDEX IF NOT EXISTS idx_savings_rules_income_rule_id ON savings_rules(income_rule_id);
CREATE INDEX IF NOT EXISTS idx_transactions_user_id ON transactions(user_id);
CREATE INDEX IF NOT EXISTS idx_transactions_occurred_on ON transactions(occurred_on);
CREATE INDEX IF NOT EXISTS idx_transactions_type ON transactions(type);
CREATE INDEX IF NOT EXISTS idx_transactions_status ON transactions(status);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_transactions_status;
DROP INDEX IF EXISTS idx_transactions_type;
DROP INDEX IF EXISTS idx_transactions_occurred_on;
DROP INDEX IF EXISTS idx_transactions_user_id;
DROP INDEX IF EXISTS idx_savings_rules_income_rule_id;
DROP INDEX IF EXISTS idx_savings_rules_user_id;
DROP INDEX IF EXISTS idx_savings_accounts_user_id;
DROP INDEX IF EXISTS idx_expense_rules_user_id;
DROP INDEX IF EXISTS idx_income_rules_user_id;

DROP TABLE IF EXISTS transactions;
DROP TABLE IF EXISTS savings_rules;
DROP TABLE IF EXISTS savings_accounts;
DROP TABLE IF EXISTS expense_rules;
DROP TABLE IF EXISTS income_rules;
DROP TABLE IF EXISTS user_settings;
DROP TABLE IF EXISTS users;

-- +goose StatementEnd

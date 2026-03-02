package database

// migrations holds all schema migration DDL strings, ordered by version.
var migrations = []string{
	migrationV1,
}

const migrationV1 = `
CREATE TABLE user_profiles (
	id         TEXT PRIMARY KEY,
	name       TEXT NOT NULL,
	created_at TEXT NOT NULL,
	updated_at TEXT NOT NULL
);

CREATE TABLE brokerage_accounts (
	id            TEXT PRIMARY KEY,
	profile_id    TEXT NOT NULL REFERENCES user_profiles(id) ON DELETE CASCADE,
	broker_name   TEXT NOT NULL,
	buy_fee_pct   REAL NOT NULL DEFAULT 0,
	sell_fee_pct  REAL NOT NULL DEFAULT 0,
	is_manual_fee INTEGER NOT NULL DEFAULT 0,
	created_at    TEXT NOT NULL,
	updated_at    TEXT NOT NULL
);

CREATE TABLE portfolios (
	id                  TEXT PRIMARY KEY,
	brokerage_acct_id   TEXT NOT NULL REFERENCES brokerage_accounts(id) ON DELETE CASCADE,
	name                TEXT NOT NULL,
	mode                TEXT NOT NULL CHECK(mode IN ('VALUE', 'DIVIDEND')),
	risk_profile        TEXT NOT NULL CHECK(risk_profile IN ('CONSERVATIVE', 'MODERATE', 'AGGRESSIVE')),
	capital             REAL NOT NULL DEFAULT 0,
	monthly_addition    REAL NOT NULL DEFAULT 0,
	max_stocks          INTEGER NOT NULL DEFAULT 0,
	universe            TEXT NOT NULL DEFAULT '[]',
	created_at          TEXT NOT NULL,
	updated_at          TEXT NOT NULL
);

CREATE TABLE holdings (
	id            TEXT PRIMARY KEY,
	portfolio_id  TEXT NOT NULL REFERENCES portfolios(id) ON DELETE CASCADE,
	ticker        TEXT NOT NULL,
	avg_buy_price REAL NOT NULL DEFAULT 0,
	lots          INTEGER NOT NULL DEFAULT 0,
	created_at    TEXT NOT NULL,
	updated_at    TEXT NOT NULL,
	UNIQUE(portfolio_id, ticker)
);

CREATE TABLE buy_transactions (
	id         TEXT PRIMARY KEY,
	holding_id TEXT NOT NULL REFERENCES holdings(id) ON DELETE CASCADE,
	date       TEXT NOT NULL,
	price      REAL NOT NULL,
	lots       INTEGER NOT NULL,
	fee        REAL NOT NULL DEFAULT 0,
	created_at TEXT NOT NULL
);

CREATE TABLE stock_data (
	id             TEXT PRIMARY KEY,
	ticker         TEXT NOT NULL,
	price          REAL NOT NULL DEFAULT 0,
	high_52_week   REAL NOT NULL DEFAULT 0,
	low_52_week    REAL NOT NULL DEFAULT 0,
	eps            REAL NOT NULL DEFAULT 0,
	bvps           REAL NOT NULL DEFAULT 0,
	roe            REAL NOT NULL DEFAULT 0,
	der            REAL NOT NULL DEFAULT 0,
	pbv            REAL NOT NULL DEFAULT 0,
	per            REAL NOT NULL DEFAULT 0,
	dividend_yield REAL NOT NULL DEFAULT 0,
	payout_ratio   REAL NOT NULL DEFAULT 0,
	fetched_at     TEXT NOT NULL,
	source         TEXT NOT NULL,
	UNIQUE(ticker, source)
);
`

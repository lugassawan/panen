package database

// migrations holds all schema migration DDL strings, ordered by version.
var migrations = []string{
	migrationV1,
	migrationV2,
	migrationV3,
	migrationV4,
	migrationV5,
	migrationV6,
	migrationV7,
	migrationV8,
	migrationV9,
	migrationV10,
	migrationV11,
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

const migrationV2 = `
ALTER TABLE brokerage_accounts ADD COLUMN sell_tax_pct REAL NOT NULL DEFAULT 0;
ALTER TABLE brokerage_accounts ADD COLUMN broker_code TEXT NOT NULL DEFAULT '';
`

const migrationV3 = `
CREATE TABLE watchlists (
	id         TEXT PRIMARY KEY,
	profile_id TEXT NOT NULL REFERENCES user_profiles(id) ON DELETE CASCADE,
	name       TEXT NOT NULL,
	created_at TEXT NOT NULL,
	updated_at TEXT NOT NULL,
	UNIQUE(profile_id, name)
);
CREATE TABLE watchlist_items (
	id           TEXT PRIMARY KEY,
	watchlist_id TEXT NOT NULL REFERENCES watchlists(id) ON DELETE CASCADE,
	ticker       TEXT NOT NULL,
	created_at   TEXT NOT NULL,
	UNIQUE(watchlist_id, ticker)
);
`

const migrationV4 = `
CREATE TABLE app_settings (
	key   TEXT PRIMARY KEY,
	value TEXT NOT NULL
);
INSERT INTO app_settings (key, value) VALUES ('auto_refresh_enabled', '1');
INSERT INTO app_settings (key, value) VALUES ('refresh_interval_minutes', '720');
INSERT INTO app_settings (key, value) VALUES ('last_refreshed_at', '');
`

const migrationV5 = `
CREATE TABLE checklist_results (
	id            TEXT PRIMARY KEY,
	portfolio_id  TEXT NOT NULL REFERENCES portfolios(id) ON DELETE CASCADE,
	ticker        TEXT NOT NULL,
	action        TEXT NOT NULL CHECK(action IN ('BUY','AVERAGE_DOWN','AVERAGE_UP','SELL_EXIT','SELL_STOP','HOLD')),
	manual_checks TEXT NOT NULL DEFAULT '{}',
	created_at    TEXT NOT NULL,
	updated_at    TEXT NOT NULL,
	UNIQUE(portfolio_id, ticker, action)
);
`

const migrationV6 = `
CREATE TABLE payday_events (
	id            TEXT PRIMARY KEY,
	month         TEXT NOT NULL,
	portfolio_id  TEXT NOT NULL REFERENCES portfolios(id) ON DELETE CASCADE,
	expected      REAL NOT NULL DEFAULT 0,
	actual        REAL NOT NULL DEFAULT 0,
	status        TEXT NOT NULL CHECK(status IN ('SCHEDULED','PENDING','CONFIRMED','DEFERRED','SKIPPED')),
	defer_until   TEXT,
	confirmed_at  TEXT,
	created_at    TEXT NOT NULL,
	updated_at    TEXT NOT NULL,
	UNIQUE(month, portfolio_id)
);
CREATE TABLE cash_flows (
	id            TEXT PRIMARY KEY,
	portfolio_id  TEXT NOT NULL REFERENCES portfolios(id) ON DELETE CASCADE,
	type          TEXT NOT NULL CHECK(type IN ('INITIAL','MONTHLY','DIVIDEND','SALE')),
	amount        REAL NOT NULL DEFAULT 0,
	date          TEXT NOT NULL,
	note          TEXT NOT NULL DEFAULT '',
	created_at    TEXT NOT NULL
);
INSERT INTO app_settings (key, value) VALUES ('payday_day', '0');
`

const migrationV7 = `
CREATE TABLE crash_capital (
	id           TEXT PRIMARY KEY,
	portfolio_id TEXT NOT NULL REFERENCES portfolios(id) ON DELETE CASCADE,
	amount       REAL NOT NULL DEFAULT 0,
	deployed     REAL NOT NULL DEFAULT 0,
	created_at   TEXT NOT NULL,
	updated_at   TEXT NOT NULL,
	UNIQUE(portfolio_id)
);
INSERT INTO app_settings (key, value) VALUES ('crash_deploy_pct_normal', '30');
INSERT INTO app_settings (key, value) VALUES ('crash_deploy_pct_crash', '40');
INSERT INTO app_settings (key, value) VALUES ('crash_deploy_pct_extreme', '30');
`

const migrationV8 = `
CREATE TABLE holding_peaks (
	id         TEXT PRIMARY KEY,
	holding_id TEXT NOT NULL UNIQUE REFERENCES holdings(id) ON DELETE CASCADE,
	peak_price REAL NOT NULL DEFAULT 0,
	updated_at TEXT NOT NULL
);
`

const migrationV9 = `
CREATE TABLE price_history (
	id     TEXT PRIMARY KEY,
	ticker TEXT NOT NULL,
	date   TEXT NOT NULL,
	open   REAL NOT NULL DEFAULT 0,
	high   REAL NOT NULL DEFAULT 0,
	low    REAL NOT NULL DEFAULT 0,
	close  REAL NOT NULL DEFAULT 0,
	volume INTEGER NOT NULL DEFAULT 0,
	source TEXT NOT NULL,
	UNIQUE(ticker, date, source)
);
CREATE INDEX idx_price_history_ticker_date ON price_history(ticker, date);
`

const migrationV10 = `
CREATE TABLE dividend_history (
	id      TEXT PRIMARY KEY,
	ticker  TEXT NOT NULL,
	ex_date TEXT NOT NULL,
	amount  REAL NOT NULL DEFAULT 0,
	source  TEXT NOT NULL,
	UNIQUE(ticker, ex_date, source)
);
CREATE INDEX idx_dividend_history_ticker_date ON dividend_history(ticker, ex_date);
`

const migrationV11 = `
CREATE TABLE financial_snapshots (
	id             TEXT PRIMARY KEY,
	ticker         TEXT NOT NULL,
	price          REAL NOT NULL DEFAULT 0,
	eps            REAL NOT NULL DEFAULT 0,
	bvps           REAL NOT NULL DEFAULT 0,
	roe            REAL NOT NULL DEFAULT 0,
	der            REAL NOT NULL DEFAULT 0,
	pbv            REAL NOT NULL DEFAULT 0,
	per            REAL NOT NULL DEFAULT 0,
	dividend_yield REAL NOT NULL DEFAULT 0,
	payout_ratio   REAL NOT NULL DEFAULT 0,
	source         TEXT NOT NULL,
	fetched_at     TEXT NOT NULL
);
CREATE INDEX idx_financial_snapshots_ticker ON financial_snapshots(ticker, fetched_at);

CREATE TABLE fundamental_alerts (
	id          TEXT PRIMARY KEY,
	ticker      TEXT NOT NULL,
	metric      TEXT NOT NULL,
	severity    TEXT NOT NULL CHECK(severity IN ('MINOR','WARNING','CRITICAL')),
	old_value   REAL NOT NULL DEFAULT 0,
	new_value   REAL NOT NULL DEFAULT 0,
	change_pct  REAL NOT NULL DEFAULT 0,
	status      TEXT NOT NULL CHECK(status IN ('ACTIVE','ACKNOWLEDGED','RESOLVED')) DEFAULT 'ACTIVE',
	detected_at TEXT NOT NULL,
	resolved_at TEXT
);
CREATE INDEX idx_fundamental_alerts_ticker ON fundamental_alerts(ticker, status);
CREATE INDEX idx_fundamental_alerts_status ON fundamental_alerts(status, detected_at);
`

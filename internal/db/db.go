package db

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	*sql.DB
}

func New(path string) (*DB, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}
	return &DB{db}, nil
}

func (db *DB) Migrate() error {
	schema := `
CREATE TABLE IF NOT EXISTS api_keys (
    key TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    initial_balance DECIMAL NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS config (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS orders (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    api_key TEXT NOT NULL,
    symbol TEXT NOT NULL,
    side TEXT NOT NULL CHECK (side IN ('BUY', 'SELL')),
    type TEXT NOT NULL CHECK (type = 'LIMIT'),
    price DECIMAL NOT NULL,
    quantity DECIMAL NOT NULL,
    executed_qty DECIMAL DEFAULT 0,
    leverage INTEGER DEFAULT 1,
    status TEXT DEFAULT 'NEW' CHECK (status IN ('NEW', 'PARTIALLY_FILLED', 'FILLED', 'CANCELLED')),
    client_order_id TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (api_key) REFERENCES api_keys(key)
);

CREATE INDEX IF NOT EXISTS idx_orders_api_key ON orders(api_key);
CREATE INDEX IF NOT EXISTS idx_orders_symbol_status ON orders(symbol, status);

CREATE TABLE IF NOT EXISTS trades (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    order_id INTEGER NOT NULL,
    api_key TEXT NOT NULL,
    symbol TEXT NOT NULL,
    side TEXT NOT NULL,
    price DECIMAL NOT NULL,
    quantity DECIMAL NOT NULL,
    quote_qty DECIMAL NOT NULL,
    fee DECIMAL NOT NULL,
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (order_id) REFERENCES orders(id),
    FOREIGN KEY (api_key) REFERENCES api_keys(key)
);

CREATE INDEX IF NOT EXISTS idx_trades_api_key ON trades(api_key);
CREATE INDEX IF NOT EXISTS idx_trades_symbol ON trades(symbol);

CREATE TABLE IF NOT EXISTS positions (
    api_key TEXT NOT NULL,
    symbol TEXT NOT NULL,
    side TEXT NOT NULL CHECK (side IN ('LONG', 'SHORT')),
    entry_price DECIMAL NOT NULL,
    size DECIMAL NOT NULL,
    leverage INTEGER NOT NULL,
    margin DECIMAL NOT NULL,
    unrealized_pnl DECIMAL DEFAULT 0,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (api_key, symbol),
    FOREIGN KEY (api_key) REFERENCES api_keys(key)
);

CREATE TABLE IF NOT EXISTS balances (
    api_key TEXT PRIMARY KEY,
    available DECIMAL NOT NULL,
    frozen DECIMAL DEFAULT 0,
    total_pnl DECIMAL DEFAULT 0,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (api_key) REFERENCES api_keys(key)
);

CREATE TABLE IF NOT EXISTS pnl_snapshots (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    api_key TEXT NOT NULL,
    total_pnl DECIMAL NOT NULL,
    available DECIMAL NOT NULL,
    frozen DECIMAL NOT NULL,
    snapshot_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (api_key) REFERENCES api_keys(key)
);

CREATE INDEX IF NOT EXISTS idx_pnl_snapshots_api_key ON pnl_snapshots(api_key);
CREATE INDEX IF NOT EXISTS idx_pnl_snapshots_time ON pnl_snapshots(snapshot_at);
`
	_, err := db.Exec(schema)
	return err
}

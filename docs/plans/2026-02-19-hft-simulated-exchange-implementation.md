# 模拟交易所系统实施计划

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 构建一个基于币安真实 trade 流的永续合约模拟交易所，支持 CCXT 兼容 API 和策略排行榜

**Architecture:** 单体 Go 服务，包含数据收集器、撮合引擎、API Server 和 Web Dashboard，SQLite 存储，WebSocket 实时推送

**Tech Stack:** Go 1.21+, Gin, gorilla/websocket, SQLite3, Vue.js (简易前端)

---

## Task 1: 项目初始化

**Files:**
- Create: `go.mod`
- Create: `main.go`
- Create: `.gitignore`

**Step 1: 初始化 Go 模块**

```bash
cd /home/wurong/workspaces/ShadowHFT
go mod init hft-sim
```

**Step 2: 创建 main.go 骨架**

```go
package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	log.Println("Starting HFT Simulated Exchange...")

	// Graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	log.Println("Shutting down...")
}
```

**Step 3: 创建 .gitignore**

```
*.db
*.db-journal
.env
bin/
dist/
```

**Step 4: Commit**

```bash
git add go.mod main.go .gitignore
git commit -m "chore: initialize project"
```

---

## Task 2: 数据库模型和迁移

**Files:**
- Create: `internal/db/db.go`
- Create: `internal/db/schema.go`
- Test: `internal/db/db_test.go`

**Step 1: 安装依赖**

```bash
go get github.com/mattn/go-sqlite3
go get github.com/stretchr/testify
```

**Step 2: 创建数据库连接和迁移**

```go
// internal/db/db.go
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
`
	_, err := db.Exec(schema)
	return err
}
```

**Step 3: 创建测试**

```go
// internal/db/db_test.go
package db

import (
	"os"
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDB_Migrate(t *testing.T) {
	dbPath := "/tmp/test_hft.db"
	os.Remove(dbPath)
	defer os.Remove(dbPath)

	db, err := New(dbPath)
	require.NoError(t, err)
	defer db.Close()

	err = db.Migrate()
	assert.NoError(t, err)

	// Verify tables exist
	tables := []string{"api_keys", "config", "orders", "trades", "positions", "balances"}
	for _, table := range tables {
		var name string
		err := db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&name)
		assert.NoError(t, err, "table %s should exist", table)
		assert.Equal(t, table, name)
	}
}
```

**Step 4: 运行测试**

```bash
go test ./internal/db -v
```

Expected: PASS

**Step 5: Commit**

```bash
git add internal/db/
git commit -m "feat: add database models and migration"
```

---

## Task 3: 配置管理

**Files:**
- Create: `internal/config/config.go`
- Modify: `internal/db/db.go` (添加配置读写方法)

**Step 1: 创建配置结构体**

```go
// internal/config/config.go
package config

import (
	"encoding/json"
	"hft-sim/internal/db"
)

type Config struct {
	db *db.DB
}

func New(db *db.DB) *Config {
	return &Config{db: db}
}

func (c *Config) Get(key string) (string, error) {
	var value string
	err := c.db.QueryRow("SELECT value FROM config WHERE key = ?", key).Scan(&value)
	return value, err
}

func (c *Config) Set(key, value string) error {
	_, err := c.db.Exec("INSERT OR REPLACE INTO config (key, value) VALUES (?, ?)", key, value)
	return err
}

func (c *Config) GetStringSlice(key string) ([]string, error) {
	value, err := c.Get(key)
	if err != nil {
		return nil, err
	}
	var result []string
	err = json.Unmarshal([]byte(value), &result)
	return result, err
}

func (c *Config) SetStringSlice(key string, value []string) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return c.Set(key, string(data))
}

// 默认配置
func (c *Config) InitDefaults() error {
	defaults := map[string]string{
		"supported_symbols":       `["BTCUSDT","ETHUSDT"]`,
		"max_leverage":            "125",
		"default_leverage":        "10",
		"maintenance_margin_rate": "0.005",
		"trade_fee_maker":         "0.0002",
		"trade_fee_taker":         "0.0005",
		"binance_ws_url":          "wss://stream.binance.com:9443/ws",
		"max_orders_per_api_key":  "100",
		"order_expire_hours":      "168",
	}

	for key, value := range defaults {
		_, err := c.Get(key)
		if err != nil {
			if err := c.Set(key, value); err != nil {
				return err
			}
		}
	}
	return nil
}
```

**Step 2: Commit**

```bash
git add internal/config/
git commit -m "feat: add config management"
```

---

## Task 4: 币安 WebSocket 数据收集器

**Files:**
- Create: `internal/collector/collector.go`
- Test: `internal/collector/collector_test.go`

**Step 1: 创建 Trade 数据结构**

```go
// internal/collector/collector.go
package collector

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Trade 币安 aggTrade 数据
type Trade struct {
	EventType  string `json:"e"`
	EventTime  int64  `json:"E"`
	Symbol     string `json:"s"`
	TradeID    int64  `json:"a"`
	Price      string `json:"p"`
	Quantity   string `json:"q"`
	FirstTrade int64  `json:"f"`
	LastTrade  int64  `json:"l"`
	TradeTime  int64  `json:"T"`
	IsBuyerMM  bool   `json:"m"`
}

type Collector struct {
	wsURL     string
	symbols   []string
	conn      *websocket.Conn
	trades    chan Trade
	stop      chan struct{}
	mu        sync.RWMutex
	handlers  []func(Trade)
}

func New(wsURL string, symbols []string) *Collector {
	return &Collector{
		wsURL:   wsURL,
		symbols: symbols,
		trades:  make(chan Trade, 1000),
		stop:    make(chan struct{}),
		handlers: make([]func(Trade), 0),
	}
}

func (c *Collector) AddHandler(handler func(Trade)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.handlers = append(c.handlers, handler)
}

func (c *Collector) connect() error {
	// 构建订阅 URL
	streams := ""
	for i, symbol := range c.symbols {
		if i > 0 {
			streams += "/"
		}
		streams += fmt.Sprintf("%s@aggTrade", symbol)
	}

	url := fmt.Sprintf("%s/%s", c.wsURL, streams)
	log.Printf("Connecting to Binance: %s", url)

	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return err
	}
	c.conn = conn
	return nil
}

func (c *Collector) Start() error {
	if err := c.connect(); err != nil {
		return err
	}

	go c.readLoop()
	go c.dispatchLoop()
	return nil
}

func (c *Collector) readLoop() {
	defer close(c.trades)

	for {
		select {
		case <-c.stop:
			return
		default:
			_, message, err := c.conn.ReadMessage()
			if err != nil {
				log.Printf("WebSocket read error: %v", err)
				time.Sleep(time.Second)
				c.reconnect()
				continue
			}

			var trade Trade
			if err := json.Unmarshal(message, &trade); err != nil {
				log.Printf("Unmarshal error: %v", err)
				continue
			}

			if trade.EventType == "aggTrade" {
				select {
				case c.trades <- trade:
				default:
					log.Println("Trade channel full, dropping trade")
				}
			}
		}
	}
}

func (c *Collector) dispatchLoop() {
	for trade := range c.trades {
		c.mu.RLock()
		handlers := c.handlers
		c.mu.RUnlock()

		for _, handler := range handlers {
			go handler(trade)
		}
	}
}

func (c *Collector) reconnect() {
	c.conn.Close()
	for {
		if err := c.connect(); err == nil {
			log.Println("Reconnected to Binance")
			return
		}
		log.Println("Reconnect failed, retrying...")
		time.Sleep(time.Second * 5)
	}
}

func (c *Collector) Stop() {
	close(c.stop)
	if c.conn != nil {
		c.conn.Close()
	}
}
```

**Step 2: 安装依赖**

```bash
go get github.com/gorilla/websocket
```

**Step 3: Commit**

```bash
git add internal/collector/
git commit -m "feat: add binance websocket collector"
```

---

## Task 5: 订单和持仓管理

**Files:**
- Create: `internal/models/order.go`
- Create: `internal/models/position.go`
- Create: `internal/models/balance.go`
- Create: `internal/store/order_store.go`
- Create: `internal/store/position_store.go`

**Step 1: 订单模型**

```go
// internal/models/order.go
package models

import "time"

type OrderStatus string

const (
	OrderStatusNew             OrderStatus = "NEW"
	OrderStatusPartiallyFilled OrderStatus = "PARTIALLY_FILLED"
	OrderStatusFilled          OrderStatus = "FILLED"
	OrderStatusCancelled       OrderStatus = "CANCELLED"
)

type Side string

const (
	SideBuy  Side = "BUY"
	SideSell Side = "SELL"
)

type Order struct {
	ID            int64       `json:"orderId"`
	APIKey        string      `json:"-"`
	Symbol        string      `json:"symbol"`
	Side          Side        `json:"side"`
	Type          string      `json:"type"` // LIMIT only
	Price         float64     `json:"price,string"`
	Quantity      float64     `json:"origQty,string"`
	ExecutedQty   float64     `json:"executedQty,string"`
	Leverage      int         `json:"leverage"`
	Status        OrderStatus `json:"status"`
	ClientOrderID string      `json:"clientOrderId"`
	CreatedAt     time.Time   `json:"time"`
	UpdatedAt     time.Time   `json:"updateTime"`
}

type Trade struct {
	ID        int64     `json:"id"`
	OrderID   int64     `json:"orderId"`
	APIKey    string    `json:"-"`
	Symbol    string    `json:"symbol"`
	Side      Side      `json:"side"`
	Price     float64   `json:"price,string"`
	Quantity  float64   `json:"qty,string"`
	QuoteQty  float64   `json:"quoteQty,string"`
	Fee       float64   `json:"fee,string"`
	Timestamp time.Time `json:"time"`
}
```

**Step 2: 持仓模型**

```go
// internal/models/position.go
package models

import "time"

type PositionSide string

const (
	PositionSideLong  PositionSide = "LONG"
	PositionSideShort PositionSide = "SHORT"
)

type Position struct {
	APIKey         string       `json:"-"`
	Symbol         string       `json:"symbol"`
	Side           PositionSide `json:"side"`
	EntryPrice     float64      `json:"entryPrice,string"`
	Size           float64      `json:"size,string"`
	Leverage       int          `json:"leverage"`
	Margin         float64      `json:"margin,string"`
	UnrealizedPNL  float64      `json:"unrealizedPnl,string"`
	UpdatedAt      time.Time    `json:"updatedAt"`
}

// Balance 账户余额
type Balance struct {
	APIKey     string  `json:"-"`
	Available  float64 `json:"available,string"`
	Frozen     float64 `json:"frozen,string"`
	TotalPNL   float64 `json:"totalPnl,string"`
}
```

**Step 3: 订单存储**

```go
// internal/store/order_store.go
package store

import (
	"database/sql"
	"hft-sim/internal/models"
)

type OrderStore struct {
	db *sql.DB
}

func NewOrderStore(db *sql.DB) *OrderStore {
	return &OrderStore{db: db}
}

func (s *OrderStore) Create(order *models.Order) error {
	query := `
		INSERT INTO orders (api_key, symbol, side, type, price, quantity, leverage, client_order_id)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`
	result, err := s.db.Exec(query, order.APIKey, order.Symbol, order.Side, order.Type,
		order.Price, order.Quantity, order.Leverage, order.ClientOrderID)
	if err != nil {
		return err
	}
	order.ID, _ = result.LastInsertId()
	return nil
}

func (s *OrderStore) GetByAPIKey(apiKey string) ([]models.Order, error) {
	query := `SELECT id, api_key, symbol, side, type, price, quantity, executed_qty,
	          leverage, status, client_order_id, created_at, updated_at
	          FROM orders WHERE api_key = ? ORDER BY created_at DESC`

	rows, err := s.db.Query(query, apiKey)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		var o models.Order
		err := rows.Scan(&o.ID, &o.APIKey, &o.Symbol, &o.Side, &o.Type, &o.Price,
			&o.Quantity, &o.ExecutedQty, &o.Leverage, &o.Status, &o.ClientOrderID,
			&o.CreatedAt, &o.UpdatedAt)
		if err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}
	return orders, nil
}

func (s *OrderStore) GetOpenBySymbol(symbol string) ([]models.Order, error) {
	query := `SELECT id, api_key, symbol, side, type, price, quantity, executed_qty,
	          leverage, status, client_order_id, created_at, updated_at
	          FROM orders WHERE symbol = ? AND status IN ('NEW', 'PARTIALLY_FILLED')`

	rows, err := s.db.Query(query, symbol)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		var o models.Order
		err := rows.Scan(&o.ID, &o.APIKey, &o.Symbol, &o.Side, &o.Type, &o.Price,
			&o.Quantity, &o.ExecutedQty, &o.Leverage, &o.Status, &o.ClientOrderID,
			&o.CreatedAt, &o.UpdatedAt)
		if err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}
	return orders, nil
}

func (s *OrderStore) UpdateStatus(id int64, status models.OrderStatus, executedQty float64) error {
	_, err := s.db.Exec(`UPDATE orders SET status = ?, executed_qty = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		status, executedQty, id)
	return err
}

func (s *OrderStore) Cancel(id int64) error {
	_, err := s.db.Exec(`UPDATE orders SET status = 'CANCELLED', updated_at = CURRENT_TIMESTAMP WHERE id = ? AND status IN ('NEW', 'PARTIALLY_FILLED')`, id)
	return err
}

func (s *OrderStore) CreateTrade(trade *models.Trade) error {
	query := `
		INSERT INTO trades (order_id, api_key, symbol, side, price, quantity, quote_qty, fee)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := s.db.Exec(query, trade.OrderID, trade.APIKey, trade.Symbol, trade.Side,
		trade.Price, trade.Quantity, trade.QuoteQty, trade.Fee)
	return err
}
```

**Step 4: 持仓存储**

```go
// internal/store/position_store.go
package store

import (
	"database/sql"
	"hft-sim/internal/models"
)

type PositionStore struct {
	db *sql.DB
}

func NewPositionStore(db *sql.DB) *PositionStore {
	return &PositionStore{db: db}
}

func (s *PositionStore) Get(apiKey, symbol string) (*models.Position, error) {
	query := `SELECT api_key, symbol, side, entry_price, size, leverage, margin, unrealized_pnl, updated_at
	          FROM positions WHERE api_key = ? AND symbol = ?`

	var p models.Position
	err := s.db.QueryRow(query, apiKey, symbol).Scan(
		&p.APIKey, &p.Symbol, &p.Side, &p.EntryPrice, &p.Size,
		&p.Leverage, &p.Margin, &p.UnrealizedPNL, &p.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (s *PositionStore) GetByAPIKey(apiKey string) ([]models.Position, error) {
	query := `SELECT api_key, symbol, side, entry_price, size, leverage, margin, unrealized_pnl, updated_at
	          FROM positions WHERE api_key = ?`

	rows, err := s.db.Query(query, apiKey)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var positions []models.Position
	for rows.Next() {
		var p models.Position
		err := rows.Scan(&p.APIKey, &p.Symbol, &p.Side, &p.EntryPrice, &p.Size,
			&p.Leverage, &p.Margin, &p.UnrealizedPNL, &p.UpdatedAt)
		if err != nil {
			return nil, err
		}
		positions = append(positions, p)
	}
	return positions, nil
}

func (s *PositionStore) Save(position *models.Position) error {
	query := `
		INSERT INTO positions (api_key, symbol, side, entry_price, size, leverage, margin, unrealized_pnl)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(api_key, symbol) DO UPDATE SET
			side = excluded.side,
			entry_price = excluded.entry_price,
			size = excluded.size,
			leverage = excluded.leverage,
			margin = excluded.margin,
			unrealized_pnl = excluded.unrealized_pnl,
			updated_at = CURRENT_TIMESTAMP
	`
	_, err := s.db.Exec(query, position.APIKey, position.Symbol, position.Side,
		position.EntryPrice, position.Size, position.Leverage, position.Margin, position.UnrealizedPNL)
	return err
}

func (s *PositionStore) Delete(apiKey, symbol string) error {
	_, err := s.db.Exec(`DELETE FROM positions WHERE api_key = ? AND symbol = ?`, apiKey, symbol)
	return err
}
```

**Step 5: Commit**

```bash
git add internal/models/ internal/store/
git commit -m "feat: add order and position models with stores"
```

---

## Task 6: 撮合引擎

**Files:**
- Create: `internal/matching/engine.go`
- Test: `internal/matching/engine_test.go`

**Step 1: 撮合引擎实现**

```go
// internal/matching/engine.go
package matching

import (
	"database/sql"
	"log"
	"strconv"
	"time"

	"hft-sim/internal/collector"
	"hft-sim/internal/models"
	"hft-sim/internal/store"
)

type Engine struct {
	orderStore    *store.OrderStore
	positionStore *store.PositionStore
	balanceStore  *store.BalanceStore
	configStore   *store.ConfigStore
}

func NewEngine(db *sql.DB) *Engine {
	return &Engine{
		orderStore:    store.NewOrderStore(db),
		positionStore: store.NewPositionStore(db),
		balanceStore:  store.NewBalanceStore(db),
		configStore:   store.NewConfigStore(db),
	}
}

func (e *Engine) OnTrade(trade collector.Trade) {
	price, _ := strconv.ParseFloat(trade.Price, 64)

	// 获取该 symbol 的所有未成交订单
	orders, err := e.orderStore.GetOpenBySymbol(trade.Symbol)
	if err != nil {
		log.Printf("Error getting open orders: %v", err)
		return
	}

	for _, order := range orders {
		if e.shouldMatch(order, price) {
			e.matchOrder(&order, price)
		}
	}
}

// shouldMatch 判断订单是否应该成交
// 买单：当 trade price < limit price 时成交（price-through）
// 卖单：当 trade price > limit price 时成交
func (e *Engine) shouldMatch(order models.Order, tradePrice float64) bool {
	if order.Status == models.OrderStatusFilled || order.Status == models.OrderStatusCancelled {
		return false
	}

	remainingQty := order.Quantity - order.ExecutedQty
	if remainingQty <= 0 {
		return false
	}

	switch order.Side {
	case models.SideBuy:
		// 买单：价格击穿（低于限价）
		return tradePrice < order.Price
	case models.SideSell:
		// 卖单：价格击穿（高于限价）
		return tradePrice > order.Price
	}
	return false
}

func (e *Engine) matchOrder(order *models.Order, price float64) {
	remainingQty := order.Quantity - order.ExecutedQty

	// 获取手续费率
	makerFee := 0.0002 // 从配置读取

	// 创建成交记录
	quoteQty := remainingQty * price
	fee := quoteQty * makerFee

	trade := &models.Trade{
		OrderID:   order.ID,
		APIKey:    order.APIKey,
		Symbol:    order.Symbol,
		Side:      order.Side,
		Price:     price,
		Quantity:  remainingQty,
		QuoteQty:  quoteQty,
		Fee:       fee,
		Timestamp: time.Now(),
	}

	if err := e.orderStore.CreateTrade(trade); err != nil {
		log.Printf("Error creating trade: %v", err)
		return
	}

	// 更新订单状态
	newExecutedQty := order.Quantity
	if err := e.orderStore.UpdateStatus(order.ID, models.OrderStatusFilled, newExecutedQty); err != nil {
		log.Printf("Error updating order: %v", err)
		return
	}

	// 更新持仓
	if err := e.updatePosition(order, remainingQty, price); err != nil {
		log.Printf("Error updating position: %v", err)
		return
	}

	// 更新余额
	if err := e.updateBalance(order, remainingQty, price, fee); err != nil {
		log.Printf("Error updating balance: %v", err)
		return
	}

	log.Printf("Order %d matched: %s %s @ %f", order.ID, order.Side, order.Symbol, price)
}

func (e *Engine) updatePosition(order *models.Order, qty, price float64) error {
	// 简化版：开新仓或平仓逻辑
	position, err := e.positionStore.Get(order.APIKey, order.Symbol)
	if err != nil {
		return err
	}

	// 如果没有持仓或反向持仓，开新仓
	if position == nil || (position.Side == models.PositionSideLong && order.Side == models.SideSell) ||
		(position.Side == models.PositionSideShort && order.Side == models.SideBuy) {

		newSide := models.PositionSideLong
		if order.Side == models.SideSell {
			newSide = models.PositionSideShort
		}

		newPosition := &models.Position{
			APIKey:     order.APIKey,
			Symbol:     order.Symbol,
			Side:       newSide,
			EntryPrice: price,
			Size:       qty,
			Leverage:   order.Leverage,
			Margin:     qty * price / float64(order.Leverage),
		}
		return e.positionStore.Save(newPosition)
	}

	// 同向加仓或平仓逻辑（简化）
	// TODO: 完整实现
	return nil
}

func (e *Engine) updateBalance(order *models.Order, qty, price, fee float64) error {
	balance, err := e.balanceStore.Get(order.APIKey)
	if err != nil {
		return err
	}

	// 扣除手续费
	balance.Available -= fee
	balance.TotalPNL -= fee

	return e.balanceStore.Update(balance)
}
```

**Step 2: Commit**

```bash
git add internal/matching/
git commit -m "feat: add matching engine with price-through logic"
```

---

## Task 7: CCXT 兼容 API

**Files:**
- Create: `internal/api/server.go`
- Create: `internal/api/handlers.go`
- Create: `internal/api/middleware.go`

**Step 1: 安装 Gin**

```bash
go get github.com/gin-gonic/gin
go get github.com/gin-contrib/cors
```

**Step 2: API 服务器**

```go
// internal/api/server.go
package api

import (
	"database/sql"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"hft-sim/internal/store"
)

type Server struct {
	router       *gin.Engine
	orderStore   *store.OrderStore
	balanceStore *store.BalanceStore
	positionStore *store.PositionStore
}

func NewServer(db *sql.DB) *Server {
	s := &Server{
		router:       gin.Default(),
		orderStore:   store.NewOrderStore(db),
		balanceStore: store.NewBalanceStore(db),
		positionStore: store.NewPositionStore(db),
	}

	s.setupRoutes()
	return s
}

func (s *Server) setupRoutes() {
	s.router.Use(cors.Default())

	// API v3 (CCXT 兼容)
	api := s.router.Group("/api/v3")
	{
		api.Use(s.authMiddleware())

		api.GET("/account", s.getAccount)
		api.POST("/order", s.createOrder)
		api.DELETE("/order", s.cancelOrder)
		api.GET("/order", s.getOrder)
		api.GET("/openOrders", s.getOpenOrders)
		api.GET("/myTrades", s.getMyTrades)
	}

	// Public endpoints
	s.router.GET("/api/v3/exchangeInfo", s.getExchangeInfo)
	s.router.GET("/api/v3/depth", s.getDepth)
}

func (s *Server) Run(addr string) error {
	return s.router.Run(addr)
}
```

**Step 3: 认证中间件**

```go
// internal/api/middleware.go
package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (s *Server) authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-MBX-APIKEY")
		if apiKey == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"code": -2015, "msg": "Invalid API-key, IP, or permissions for action"})
			c.Abort()
			return
		}

		// 验证 API Key 是否存在
		// TODO: 从数据库验证

		c.Set("apiKey", apiKey)
		c.Next()
	}
}
```

**Step 4: 订单处理函数**

```go
// internal/api/handlers.go
package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"hft-sim/internal/models"
)

func (s *Server) getAccount(c *gin.Context) {
	apiKey := c.GetString("apiKey")

	balance, err := s.balanceStore.Get(apiKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1000, "msg": err.Error()})
		return
	}

	positions, _ := s.positionStore.GetByAPIKey(apiKey)

	c.JSON(http.StatusOK, gin.H{
		"makerCommission":  2,
		"takerCommission":  5,
		"buyerCommission":  0,
		"sellerCommission": 0,
		"canTrade":         true,
		"canWithdraw":      false,
		"canDeposit":       false,
		"updateTime":       time.Now().UnixMilli(),
		"accountType":      "FUTURES",
		"balances": []gin.H{
			{
				"asset":  "USDT",
				"free":   balance.Available,
				"locked": balance.Frozen,
			},
		},
		"positions": positions,
	})
}

type CreateOrderRequest struct {
	Symbol        string `json:"symbol" binding:"required"`
	Side          string `json:"side" binding:"required"`
	Type          string `json:"type" binding:"required"`
	TimeInForce   string `json:"timeInForce"`
	Quantity      string `json:"quantity" binding:"required"`
	Price         string `json:"price" binding:"required"`
	Leverage      int    `json:"leverage"`
	ClientOrderID string `json:"newClientOrderId"`
}

func (s *Server) createOrder(c *gin.Context) {
	apiKey := c.GetString("apiKey")

	var req CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1100, "msg": "Illegal characters found in parameter"})
		return
	}

	// 只支持限价单
	if req.Type != "LIMIT" {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1106, "msg": "只支持 LIMIT 订单类型"})
		return
	}

	quantity, _ := strconv.ParseFloat(req.Quantity, 64)
	price, _ := strconv.ParseFloat(req.Price, 64)
	leverage := req.Leverage
	if leverage == 0 {
		leverage = 10 // 默认杠杆
	}

	order := &models.Order{
		APIKey:        apiKey,
		Symbol:        req.Symbol,
		Side:          models.Side(req.Side),
		Type:          req.Type,
		Price:         price,
		Quantity:      quantity,
		Leverage:      leverage,
		ClientOrderID: req.ClientOrderID,
		Status:        models.OrderStatusNew,
	}

	if err := s.orderStore.Create(order); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1000, "msg": err.Error()})
		return
	}

	c.JSON(http.StatusOK, order)
}

func (s *Server) cancelOrder(c *gin.Context) {
	orderID, _ := strconv.ParseInt(c.Query("orderId"), 10, 64)

	if err := s.orderStore.Cancel(orderID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1000, "msg": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"orderId": orderID,
		"status":  "CANCELLED",
	})
}

func (s *Server) getOrder(c *gin.Context) {
	// TODO: 实现查询单个订单
}

func (s *Server) getOpenOrders(c *gin.Context) {
	apiKey := c.GetString("apiKey")

	orders, err := s.orderStore.GetByAPIKey(apiKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1000, "msg": err.Error()})
		return
	}

	c.JSON(http.StatusOK, orders)
}

func (s *Server) getMyTrades(c *gin.Context) {
	// TODO: 实现查询成交历史
}

func (s *Server) getExchangeInfo(c *gin.Context) {
	// TODO: 返回支持的币种信息
	c.JSON(http.StatusOK, gin.H{
		"symbols": []gin.H{
			{
				"symbol":     "BTCUSDT",
				"status":     "TRADING",
				"baseAsset":  "BTC",
				"quoteAsset": "USDT",
			},
		},
	})
}

func (s *Server) getDepth(c *gin.Context) {
	// TODO: 返回模拟 orderbook
	c.JSON(http.StatusOK, gin.H{
		"lastUpdateId": time.Now().UnixMilli(),
		"bids":         [][]string{},
		"asks":         [][]string{},
	})
}
```

**Step 5: 补充 BalanceStore**

```go
// internal/store/balance_store.go
package store

import (
	"database/sql"
	"hft-sim/internal/models"
)

type BalanceStore struct {
	db *sql.DB
}

func NewBalanceStore(db *sql.DB) *BalanceStore {
	return &BalanceStore{db: db}
}

func (s *BalanceStore) Get(apiKey string) (*models.Balance, error) {
	query := `SELECT api_key, available, frozen, total_pnl FROM balances WHERE api_key = ?`

	var b models.Balance
	err := s.db.QueryRow(query, apiKey).Scan(&b.APIKey, &b.Available, &b.Frozen, &b.TotalPNL)
	if err == sql.ErrNoRows {
		return &models.Balance{APIKey: apiKey, Available: 0}, nil
	}
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func (s *BalanceStore) Update(balance *models.Balance) error {
	query := `
		INSERT INTO balances (api_key, available, frozen, total_pnl)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(api_key) DO UPDATE SET
			available = excluded.available,
			frozen = excluded.frozen,
			total_pnl = excluded.total_pnl,
			updated_at = CURRENT_TIMESTAMP
	`
	_, err := s.db.Exec(query, balance.APIKey, balance.Available, balance.Frozen, balance.TotalPNL)
	return err
}
```

**Step 6: Commit**

```bash
git add internal/api/ internal/store/balance_store.go
git commit -m "feat: add CCXT-compatible REST API"
```

---

## Task 8: WebSocket 实时推送

**Files:**
- Create: `internal/api/websocket.go`

**Step 1: WebSocket 处理**

```go
// internal/api/websocket.go
package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (s *Server) handleWebSocket(c *gin.Context) {
	apiKey := c.Query("apiKey")
	if apiKey == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"code": -2015, "msg": "Invalid API-key"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	// 订阅处理
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			break
		}

		var req struct {
			Method string   `json:"method"`
			Params []string `json:"params"`
			ID     int      `json:"id"`
		}

		if err := json.Unmarshal(message, &req); err != nil {
			continue
		}

		// 处理订阅请求
		if req.Method == "SUBSCRIBE" {
			// 发送确认
			conn.WriteJSON(gin.H{"result": nil, "id": req.ID})
		}
	}
}
```

**Step 2: Commit**

```bash
git add internal/api/websocket.go
git commit -m "feat: add websocket support"
```

---

## Task 9: 邀请码管理

**Files:**
- Create: `internal/admin/handler.go`
- Create: `cmd/admin/main.go`

**Step 1: 管理命令行工具**

```go
// cmd/admin/main.go
package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/google/uuid"
	"hft-sim/internal/db"
)

func main() {
	var (
		action    = flag.String("action", "", "create|list|delete")
		name      = flag.String("name", "", "Strategy name")
		desc      = flag.String("desc", "", "Strategy description")
		balance   = flag.Float64("balance", 10000, "Initial balance")
		apiKey    = flag.String("key", "", "API Key (for delete)")
		dbPath    = flag.String("db", "hft.db", "Database path")
	)
	flag.Parse()

	database, err := db.New(*dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()

	switch *action {
	case "create":
		createKey(database, *name, *desc, *balance)
	case "list":
		listKeys(database)
	case "delete":
		deleteKey(database, *apiKey)
	default:
		fmt.Println("Usage: admin -action=create -name=\"MyStrategy\" -balance=10000")
	}
}

func createKey(database *db.DB, name, desc string, balance float64) {
	key := uuid.New().String()
	_, err := database.Exec(
		"INSERT INTO api_keys (key, name, description, initial_balance) VALUES (?, ?, ?, ?)",
		key, name, desc, balance)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Created API Key: %s\n", key)
}

func listKeys(database *db.DB) {
	rows, err := database.Query("SELECT key, name, description, initial_balance, created_at FROM api_keys")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	fmt.Printf("%-36s %-20s %-15s %s\n", "API Key", "Name", "Balance", "Created")
	for rows.Next() {
		var key, name, desc string
		var balance float64
		var created string
		rows.Scan(&key, &name, &desc, &balance, &created)
		fmt.Printf("%-36s %-20s %-15.2f %s\n", key, name, balance, created)
	}
}

func deleteKey(database *db.DB, key string) {
	_, err := database.Exec("DELETE FROM api_keys WHERE key = ?", key)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Deleted")
}
```

**Step 2: 安装依赖**

```bash
go get github.com/google/uuid
```

**Step 3: Commit**

```bash
git add cmd/admin/
git commit -m "feat: add admin CLI for API key management"
```

---

## Task 10: 主程序整合

**Files:**
- Modify: `main.go`

**Step 1: 更新主程序**

```go
package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"hft-sim/internal/api"
	"hft-sim/internal/collector"
	"hft-sim/internal/config"
	"hft-sim/internal/db"
	"hft-sim/internal/matching"
)

func main() {
	log.Println("Starting HFT Simulated Exchange...")

	// 初始化数据库
	database, err := db.New("hft.db")
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()

	if err := database.Migrate(); err != nil {
		log.Fatal(err)
	}

	// 初始化配置
	cfg := config.New(database)
	if err := cfg.InitDefaults(); err != nil {
		log.Fatal(err)
	}

	// 获取配置
	symbols, _ := cfg.GetStringSlice("supported_symbols")
	wsURL, _ := cfg.Get("binance_ws_url")

	// 启动撮合引擎
	engine := matching.NewEngine(database.DB)

	// 启动数据收集器
	collector := collector.New(wsURL, symbols)
	collector.AddHandler(engine.OnTrade)

	if err := collector.Start(); err != nil {
		log.Fatal(err)
	}

	// 启动 API 服务器
	server := api.NewServer(database.DB)
	go func() {
		if err := server.Run(":8080"); err != nil {
			log.Fatal(err)
		}
	}()

	log.Println("Server running on :8080")

	// Graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	log.Println("Shutting down...")
	collector.Stop()
}
```

**Step 2: Commit**

```bash
git add main.go
git commit -m "feat: wire up all components in main"
```

---

## Task 11: 前端 Dashboard

**Files:**
- Create: `web/index.html`
- Create: `web/app.js`
- Modify: `internal/api/server.go` (添加静态文件服务)

**Step 1: 创建前端页面**

```html
<!-- web/index.html -->
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>HFT 策略排行榜</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { font-family: -apple-system, sans-serif; background: #0d1117; color: #c9d1d9; }
        .header { padding: 20px; border-bottom: 1px solid #30363d; }
        .header h1 { color: #58a6ff; }
        .container { max-width: 1200px; margin: 0 auto; padding: 20px; }
        table { width: 100%; border-collapse: collapse; }
        th, td { padding: 12px; text-align: left; border-bottom: 1px solid #30363d; }
        th { color: #8b949e; }
        .rank { font-weight: bold; color: #58a6ff; }
        .profit { color: #3fb950; }
        .loss { color: #f85149; }
        .strategy-name { font-weight: 500; }
        .strategy-desc { color: #8b949e; font-size: 12px; }
    </style>
</head>
<body>
    <div class="header">
        <div class="container">
            <h1>HFT 策略竞技场</h1>
            <p>基于真实币安数据的模拟交易</p>
        </div>
    </div>
    <div class="container">
        <table id="leaderboard">
            <thead>
                <tr>
                    <th>排名</th>
                    <th>策略</th>
                    <th>累计收益</th>
                    <th>交易次数</th>
                    <th>胜率</th>
                </tr>
            </thead>
            <tbody></tbody>
        </table>
    </div>
    <script src="app.js"></script>
</body>
</html>
```

**Step 2: 前端 JS**

```javascript
// web/app.js
async function loadLeaderboard() {
    // TODO: 实现从后端获取数据
    const mockData = [
        { rank: 1, name: 'MakerBot v2', desc: '联系: maker@example.com', profit: 25.5, trades: 150, winRate: 65 },
        { rank: 2, name: 'Grid Hunter', desc: '高频网格策略', profit: 18.2, trades: 320, winRate: 58 },
    ];

    const tbody = document.querySelector('#leaderboard tbody');
    tbody.innerHTML = mockData.map(s => `
        <tr>
            <td class="rank">#${s.rank}</td>
            <td>
                <div class="strategy-name">${s.name}</div>
                <div class="strategy-desc">${s.desc}</div>
            </td>
            <td class="${s.profit >= 0 ? 'profit' : 'loss'}">${s.profit >= 0 ? '+' : ''}${s.profit}%</td>
            <td>${s.trades}</td>
            <td>${s.winRate}%</td>
        </tr>
    `).join('');
}

loadLeaderboard();
setInterval(loadLeaderboard, 5000);
```

**Step 3: 添加静态文件服务**

```go
// 在 internal/api/server.go 的 setupRoutes 中添加
func (s *Server) setupRoutes() {
	// ... existing routes ...

	// Static files
	s.router.Static("/", "./web")
}
```

**Step 4: Commit**

```bash
git add web/ internal/api/server.go
git commit -m "feat: add web dashboard"
```

---

## Task 12: 测试与验证

**Step 1: 构建项目**

```bash
cd /home/wurong/workspaces/ShadowHFT
go build -o bin/hft-sim .
go build -o bin/admin ./cmd/admin
```

**Step 2: 创建 API Key**

```bash
./bin/admin -action=create -name="TestStrategy" -desc="测试策略" -balance=10000
```

**Step 3: 启动服务**

```bash
./bin/hft-sim
```

**Step 4: 测试 API**

```bash
# 查询账户
curl -H "X-MBX-APIKEY: <your-api-key>" http://localhost:8080/api/v3/account

# 创建订单
curl -X POST http://localhost:8080/api/v3/order \
  -H "X-MBX-APIKEY: <your-api-key>" \
  -H "Content-Type: application/json" \
  -d '{
    "symbol": "BTCUSDT",
    "side": "BUY",
    "type": "LIMIT",
    "quantity": "0.01",
    "price": "40000",
    "leverage": 10
  }'
```

**Step 5: 最终提交**

```bash
git add .
git commit -m "feat: complete hft simulated exchange implementation"
```

---

## 后续优化项

1. **完善撮合逻辑**：处理部分成交、反向持仓平仓
2. **强平机制**：监控保证金率，触发强平
3. **统计数据**：实现收益率、最大回撤、夏普比率计算
4. **WebSocket 推送**：实时推送订单状态和持仓变化
5. **前端完善**：连接真实 API，实时更新排行榜
6. **测试覆盖**：单元测试和集成测试

package collector

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
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
	wsURL          string
	symbols        []string
	conn           *websocket.Conn
	trades         chan Trade
	stop           chan struct{}
	mu             sync.RWMutex
	handlers       []func(Trade)
	latestTrades   map[string]Trade // symbol -> latest trade
}

func New(wsURL string, symbols []string) *Collector {
	return &Collector{
		wsURL:        wsURL,
		symbols:      symbols,
		trades:       make(chan Trade, 1000),
		stop:         make(chan struct{}),
		handlers:     make([]func(Trade), 0),
		latestTrades: make(map[string]Trade),
	}
}

// GetLatestTrades returns the latest trade for each symbol
func (c *Collector) GetLatestTrades() map[string]Trade {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Return a copy
	result := make(map[string]Trade, len(c.latestTrades))
	for k, v := range c.latestTrades {
		result[k] = v
	}
	return result
}

func (c *Collector) AddHandler(handler func(Trade)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.handlers = append(c.handlers, handler)
}

func (c *Collector) connect() error {
	// 构建订阅 URL (symbol 必须是小写)
	streams := ""
	for i, symbol := range c.symbols {
		if i > 0 {
			streams += "/"
		}
		streams += fmt.Sprintf("%s@aggTrade", strings.ToLower(symbol))
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
		// Save latest trade for each symbol
		c.mu.Lock()
		c.latestTrades[trade.Symbol] = trade
		c.mu.Unlock()

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

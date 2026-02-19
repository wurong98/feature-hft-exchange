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

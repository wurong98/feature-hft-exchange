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

package api

import (
	"fmt"
	"math"
	"math/rand"
	"sync"
	"time"
)

// OrderbookLevel 订单簿档位
type OrderbookLevel struct {
	Price    string `json:"price"`
	Quantity string `json:"quantity"`
}

// Orderbook 订单簿快照
type OrderbookSnapshot struct {
	Symbol       string           `json:"symbol"`
	LastUpdateID int64            `json:"lastUpdateId"`
	Bids         []OrderbookLevel `json:"bids"`
	Asks         []OrderbookLevel `json:"asks"`
}

// Orderbook 模拟订单簿
type Orderbook struct {
	mu     sync.RWMutex
	prices map[string]float64 // symbol -> last price
}

func NewOrderbook() *Orderbook {
	return &Orderbook{
		prices: map[string]float64{
			"BTCUSDT": 67000,
			"ETHUSDT": 3500,
		},
	}
}

// UpdatePrice 更新最新价格
func (ob *Orderbook) UpdatePrice(symbol string, price float64) {
	ob.mu.Lock()
	defer ob.mu.Unlock()
	ob.prices[symbol] = price
}

// GetSnapshot 获取订单簿快照
func (ob *Orderbook) GetSnapshot(symbol string) *OrderbookSnapshot {
	ob.mu.RLock()
	defer ob.mu.RUnlock()

	basePrice, ok := ob.prices[symbol]
	if !ok {
		basePrice = 67000 // 默认价格
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	// 生成买盘 (bids) - 价格低于当前价
	bids := make([]OrderbookLevel, 10)
	for i := 0; i < 10; i++ {
		priceOffset := float64(i+1) * basePrice * 0.0001 * (0.5 + r.Float64())
		price := basePrice - priceOffset
		quantity := 0.1 + r.Float64()*2
		bids[i] = OrderbookLevel{
			Price:    formatPrice(price),
			Quantity: formatQuantity(quantity),
		}
	}

	// 生成卖盘 (asks) - 价格高于当前价
	asks := make([]OrderbookLevel, 10)
	for i := 0; i < 10; i++ {
		priceOffset := float64(i+1) * basePrice * 0.0001 * (0.5 + r.Float64())
		price := basePrice + priceOffset
		quantity := 0.1 + r.Float64()*2
		asks[i] = OrderbookLevel{
			Price:    formatPrice(price),
			Quantity: formatQuantity(quantity),
		}
	}

	return &OrderbookSnapshot{
		Symbol:       symbol,
		LastUpdateID: time.Now().UnixMilli(),
		Bids:         bids,
		Asks:         asks,
	}
}

func formatPrice(price float64) string {
	if price >= 1000 {
		return formatWithPrecision(price, 2)
	}
	return formatWithPrecision(price, 4)
}

func formatQuantity(qty float64) string {
	return formatWithPrecision(qty, 4)
}

func formatWithPrecision(val float64, prec int) string {
	p := math.Pow(10, float64(prec))
	return fmt.Sprintf("%.*f", prec, math.Round(val*p)/p)
}

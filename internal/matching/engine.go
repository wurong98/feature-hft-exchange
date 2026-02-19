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

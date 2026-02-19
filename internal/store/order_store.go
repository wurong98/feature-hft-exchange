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

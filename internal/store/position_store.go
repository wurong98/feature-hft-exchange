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

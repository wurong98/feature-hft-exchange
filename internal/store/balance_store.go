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

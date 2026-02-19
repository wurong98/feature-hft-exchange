package store

import (
	"database/sql"
)

// LeaderboardEntry 排行榜条目
type LeaderboardEntry struct {
	APIKey         string  `json:"apiKey"`
	Name           string  `json:"name"`
	Description    string  `json:"description"`
	InitialBalance float64 `json:"initialBalance,string"`
	Available      float64 `json:"available,string"`
	TotalPNL       float64 `json:"totalPnl,string"`
	TradeCount     int     `json:"tradeCount"`
	WinCount       int     `json:"winCount"`
	ROI            float64 `json:"roi,string"`
}

type LeaderboardStore struct {
	db *sql.DB
}

func NewLeaderboardStore(db *sql.DB) *LeaderboardStore {
	return &LeaderboardStore{db: db}
}

func (s *LeaderboardStore) GetLeaderboard() ([]LeaderboardEntry, error) {
	query := `
		SELECT
			a.key,
			a.name,
			a.description,
			a.initial_balance,
			COALESCE(b.available, 0) as available,
			COALESCE(b.total_pnl, 0) as total_pnl,
			COUNT(t.id) as trade_count
		FROM api_keys a
		LEFT JOIN balances b ON a.key = b.api_key
		LEFT JOIN trades t ON a.key = t.api_key
		GROUP BY a.key
		ORDER BY total_pnl DESC
	`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []LeaderboardEntry
	for rows.Next() {
		var e LeaderboardEntry
		var tradeCount int
		err := rows.Scan(
			&e.APIKey, &e.Name, &e.Description,
			&e.InitialBalance, &e.Available, &e.TotalPNL,
			&tradeCount)
		if err != nil {
			return nil, err
		}
		e.TradeCount = tradeCount
		// 计算收益率
		if e.InitialBalance > 0 {
			e.ROI = (e.TotalPNL / e.InitialBalance) * 100
		}
		entries = append(entries, e)
	}
	return entries, nil
}

// StrategyStats 策略统计数据
type StrategyStats struct {
	APIKey         string  `json:"apiKey"`
	Name           string  `json:"name"`
	InitialBalance float64 `json:"initialBalance,string"`
	Available      float64 `json:"available,string"`
	Frozen         float64 `json:"frozen,string"`
	TotalPNL       float64 `json:"totalPnl,string"`
	TradeCount     int     `json:"tradeCount"`
	WinCount       int     `json:"winCount"`
	LossCount      int     `json:"lossCount"`
	ROI            float64 `json:"roi,string"`
	MaxDrawdown    float64 `json:"maxDrawdown,string"`
}

func (s *LeaderboardStore) GetStrategyStats(apiKey string) (*StrategyStats, error) {
	query := `
		SELECT
			a.key,
			a.name,
			a.initial_balance,
			COALESCE(b.available, 0) as available,
			COALESCE(b.frozen, 0) as frozen,
			COALESCE(b.total_pnl, 0) as total_pnl,
			COUNT(t.id) as trade_count
		FROM api_keys a
		LEFT JOIN balances b ON a.key = b.api_key
		LEFT JOIN trades t ON a.key = t.api_key
		WHERE a.key = ?
		GROUP BY a.key
	`

	var stats StrategyStats
	var tradeCount int
	err := s.db.QueryRow(query, apiKey).Scan(
		&stats.APIKey, &stats.Name, &stats.InitialBalance,
		&stats.Available, &stats.Frozen, &stats.TotalPNL,
		&tradeCount)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	stats.TradeCount = tradeCount
	if stats.InitialBalance > 0 {
		stats.ROI = (stats.TotalPNL / stats.InitialBalance) * 100
	}

	return &stats, nil
}

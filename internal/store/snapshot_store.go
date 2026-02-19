package store

import (
	"database/sql"
	"time"
)

// PNLSnapshot 收益快照
type PNLSnapshot struct {
	ID         int64     `json:"id"`
	APIKey     string    `json:"apiKey"`
	TotalPNL   float64   `json:"totalPnl,string"`
	Available  float64   `json:"available,string"`
	Frozen     float64   `json:"frozen,string"`
	SnapshotAt time.Time `json:"snapshotAt"`
}

type SnapshotStore struct {
	db *sql.DB
}

func NewSnapshotStore(db *sql.DB) *SnapshotStore {
	return &SnapshotStore{db: db}
}

// CreateSnapshot 创建收益快照
func (s *SnapshotStore) CreateSnapshot(apiKey string, totalPNL, available, frozen float64) error {
	query := `
		INSERT INTO pnl_snapshots (api_key, total_pnl, available, frozen)
		VALUES (?, ?, ?, ?)
	`
	_, err := s.db.Exec(query, apiKey, totalPNL, available, frozen)
	return err
}

// GetSnapshotsByAPIKey 获取策略的收益快照历史
func (s *SnapshotStore) GetSnapshotsByAPIKey(apiKey string, limit int) ([]PNLSnapshot, error) {
	query := `
		SELECT id, api_key, total_pnl, available, frozen, snapshot_at
		FROM pnl_snapshots
		WHERE api_key = ?
		ORDER BY snapshot_at DESC
		LIMIT ?
	`

	rows, err := s.db.Query(query, apiKey, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var snapshots []PNLSnapshot
	for rows.Next() {
		var snap PNLSnapshot
		err := rows.Scan(
			&snap.ID, &snap.APIKey, &snap.TotalPNL, &snap.Available,
			&snap.Frozen, &snap.SnapshotAt)
		if err != nil {
			return nil, err
		}
		snapshots = append(snapshots, snap)
	}
	return snapshots, nil
}

// GetLatestSnapshot 获取最新的收益快照
func (s *SnapshotStore) GetLatestSnapshot(apiKey string) (*PNLSnapshot, error) {
	query := `
		SELECT id, api_key, total_pnl, available, frozen, snapshot_at
		FROM pnl_snapshots
		WHERE api_key = ?
		ORDER BY snapshot_at DESC
		LIMIT 1
	`

	var snap PNLSnapshot
	err := s.db.QueryRow(query, apiKey).Scan(
		&snap.ID, &snap.APIKey, &snap.TotalPNL, &snap.Available,
		&snap.Frozen, &snap.SnapshotAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &snap, nil
}

// GetSnapshotByTimeRange 获取时间范围内的收益快照
func (s *SnapshotStore) GetSnapshotByTimeRange(apiKey string, start, end time.Time) ([]PNLSnapshot, error) {
	query := `
		SELECT id, api_key, total_pnl, available, frozen, snapshot_at
		FROM pnl_snapshots
		WHERE api_key = ? AND snapshot_at BETWEEN ? AND ?
		ORDER BY snapshot_at ASC
	`

	rows, err := s.db.Query(query, apiKey, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var snapshots []PNLSnapshot
	for rows.Next() {
		var snap PNLSnapshot
		err := rows.Scan(
			&snap.ID, &snap.APIKey, &snap.TotalPNL, &snap.Available,
			&snap.Frozen, &snap.SnapshotAt)
		if err != nil {
			return nil, err
		}
		snapshots = append(snapshots, snap)
	}
	return snapshots, nil
}

// DeleteOldSnapshots 删除旧的快照（保留最近N天）
func (s *SnapshotStore) DeleteOldSnapshots(days int) error {
	query := `
		DELETE FROM pnl_snapshots
		WHERE snapshot_at < datetime('now', ?)
	`
	_, err := s.db.Exec(query, "-"+string(rune(days*24))+" hours")
	return err
}

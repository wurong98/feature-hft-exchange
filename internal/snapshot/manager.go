package snapshot

import (
	"database/sql"
	"log"
	"time"

	"hft-sim/internal/store"
)

type Manager struct {
	db            *sql.DB
	snapshotStore *store.SnapshotStore
	balanceStore  *store.BalanceStore
	stop          chan struct{}
}

func NewManager(db *sql.DB) *Manager {
	return &Manager{
		db:            db,
	snapshotStore: store.NewSnapshotStore(db),
		balanceStore:  store.NewBalanceStore(db),
		stop:          make(chan struct{}),
	}
}

// Start 启动定时快照任务
func (m *Manager) Start() {
	log.Println("Starting PNL snapshot manager...")

	// 立即执行一次
	m.takeSnapshot()

	// 每10分钟执行一次
	ticker := time.NewTicker(10 * time.Minute)
	go func() {
		for {
			select {
			case <-ticker.C:
				m.takeSnapshot()
			case <-m.stop:
				ticker.Stop()
				return
			}
		}
	}()
}

// Stop 停止定时任务
func (m *Manager) Stop() {
	close(m.stop)
}

// takeSnapshot 为所有策略创建收益快照
func (m *Manager) takeSnapshot() {
	// 获取所有 API Keys
	rows, err := m.db.Query("SELECT key FROM api_keys")
	if err != nil {
		log.Printf("Error getting api keys for snapshot: %v", err)
		return
	}
	defer rows.Close()

	var keys []string
	for rows.Next() {
		var key string
		if err := rows.Scan(&key); err != nil {
			continue
		}
		keys = append(keys, key)
	}

	// 为每个策略创建快照
	for _, key := range keys {
		balance, err := m.balanceStore.Get(key)
		if err != nil {
			log.Printf("Error getting balance for %s: %v", key, err)
			continue
		}

		err = m.snapshotStore.CreateSnapshot(key, balance.TotalPNL, balance.Available, balance.Frozen)
		if err != nil {
			log.Printf("Error creating snapshot for %s: %v", key, err)
			continue
		}
	}

	log.Printf("PNL snapshot taken for %d strategies", len(keys))
}

// GetSnapshots 获取策略的收益快照
func (m *Manager) GetSnapshots(apiKey string, limit int) ([]store.PNLSnapshot, error) {
	return m.snapshotStore.GetSnapshotsByAPIKey(apiKey, limit)
}

// GetSnapshotsByTimeRange 获取时间范围内的收益快照
func (m *Manager) GetSnapshotsByTimeRange(apiKey string, start, end time.Time) ([]store.PNLSnapshot, error) {
	return m.snapshotStore.GetSnapshotByTimeRange(apiKey, start, end)
}

// TakeSnapshotNow 立即为指定策略创建快照
func (m *Manager) TakeSnapshotNow(apiKey string) error {
	balance, err := m.balanceStore.Get(apiKey)
	if err != nil {
		return err
	}

	return m.snapshotStore.CreateSnapshot(apiKey, balance.TotalPNL, balance.Available, balance.Frozen)
}

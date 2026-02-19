package db

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDB_Migrate(t *testing.T) {
	dbPath := "/tmp/test_hft.db"
	os.Remove(dbPath)
	defer os.Remove(dbPath)

	db, err := New(dbPath)
	require.NoError(t, err)
	defer db.Close()

	err = db.Migrate()
	assert.NoError(t, err)

	// Verify tables exist
	tables := []string{"api_keys", "config", "orders", "trades", "positions", "balances"}
	for _, table := range tables {
		var name string
		err := db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&name)
		assert.NoError(t, err, "table %s should exist", table)
		assert.Equal(t, table, name)
	}
}

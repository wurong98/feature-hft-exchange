package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/google/uuid"
	"hft-sim/internal/db"
)

func main() {
	var (
		action  = flag.String("action", "", "create|list|delete")
		name    = flag.String("name", "", "Strategy name")
		desc    = flag.String("desc", "", "Strategy description")
		balance = flag.Float64("balance", 10000, "Initial balance")
		apiKey  = flag.String("key", "", "API Key (for delete)")
		dbPath  = flag.String("db", "hft.db", "Database path")
	)
	flag.Parse()

	database, err := db.New(*dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()

	switch *action {
	case "create":
		createKey(database, *name, *desc, *balance)
	case "list":
		listKeys(database)
	case "delete":
		deleteKey(database, *apiKey)
	default:
		fmt.Println("Usage: admin -action=create -name=\"MyStrategy\" -balance=10000")
	}
}

func createKey(database *db.DB, name, desc string, balance float64) {
	key := uuid.New().String()

	// 创建 API Key
	_, err := database.Exec(
		"INSERT INTO api_keys (key, name, description, initial_balance) VALUES (?, ?, ?, ?)",
		key, name, desc, balance)
	if err != nil {
		log.Fatal(err)
	}

	// 初始化余额记录
	_, err = database.Exec(
		"INSERT INTO balances (api_key, available, frozen, total_pnl) VALUES (?, ?, 0, 0)",
		key, balance)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Created API Key: %s\n", key)
	fmt.Printf("Initial Balance: %.2f USDT\n", balance)
}

func listKeys(database *db.DB) {
	rows, err := database.Query("SELECT key, name, description, initial_balance, created_at FROM api_keys")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	fmt.Printf("%-36s %-20s %-15s %s\n", "API Key", "Name", "Balance", "Created")
	for rows.Next() {
		var key, name, desc string
		var balance float64
		var created string
		rows.Scan(&key, &name, &desc, &balance, &created)
		fmt.Printf("%-36s %-20s %-15.2f %s\n", key, name, balance, created)
	}
}

func deleteKey(database *db.DB, key string) {
	_, err := database.Exec("DELETE FROM api_keys WHERE key = ?", key)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Deleted")
}

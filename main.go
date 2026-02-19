package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"hft-sim/internal/api"
	"hft-sim/internal/collector"
	"hft-sim/internal/config"
	"hft-sim/internal/db"
	"hft-sim/internal/matching"
	"hft-sim/internal/snapshot"
)

func main() {
	log.Println("Starting HFT Simulated Exchange...")

	// 初始化数据库
	database, err := db.New("hft.db")
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()

	if err := database.Migrate(); err != nil {
		log.Fatal(err)
	}

	// 初始化配置
	cfg := config.New(database)
	if err := cfg.InitDefaults(); err != nil {
		log.Fatal(err)
	}

	// 获取配置
	symbols, _ := cfg.GetStringSlice("supported_symbols")
	wsURL, _ := cfg.Get("binance_ws_url")

	// 启动撮合引擎
	engine := matching.NewEngine(database.DB)

	// 启动数据收集器
	coll := collector.New(wsURL, symbols)
	coll.AddHandler(engine.OnTrade)

	if err := coll.Start(); err != nil {
		log.Fatal(err)
	}

	// 启动 API 服务器
	server := api.NewServer(database.DB)
	go func() {
		if err := server.Run(":8080"); err != nil {
			log.Fatal(err)
		}
	}()

	// 启动收益快照管理器
	snapshotMgr := snapshot.NewManager(database.DB)
	snapshotMgr.Start()
	defer snapshotMgr.Stop()

	log.Println("Server running on :8080")

	// Graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	log.Println("Shutting down...")
	coll.Stop()
}

package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"hft-sim/internal/models"
)

func (s *Server) getAccount(c *gin.Context) {
	apiKey := c.GetString("apiKey")

	balance, err := s.balanceStore.Get(apiKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1000, "msg": err.Error()})
		return
	}

	positions, _ := s.positionStore.GetByAPIKey(apiKey)

	c.JSON(http.StatusOK, gin.H{
		"makerCommission":  2,
		"takerCommission":  5,
		"buyerCommission":  0,
		"sellerCommission": 0,
		"canTrade":         true,
		"canWithdraw":      false,
		"canDeposit":       false,
		"updateTime":       time.Now().UnixMilli(),
		"accountType":      "FUTURES",
		"balances": []gin.H{
			{
				"asset":  "USDT",
				"free":   balance.Available,
				"locked": balance.Frozen,
			},
		},
		"positions": positions,
	})
}

type CreateOrderRequest struct {
	Symbol        string `json:"symbol" binding:"required"`
	Side          string `json:"side" binding:"required"`
	Type          string `json:"type" binding:"required"`
	TimeInForce   string `json:"timeInForce"`
	Quantity      string `json:"quantity" binding:"required"`
	Price         string `json:"price" binding:"required"`
	Leverage      int    `json:"leverage"`
	ClientOrderID string `json:"newClientOrderId"`
}

func (s *Server) createOrder(c *gin.Context) {
	apiKey := c.GetString("apiKey")

	var req CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1100, "msg": "Illegal characters found in parameter"})
		return
	}

	// 只支持限价单
	if req.Type != "LIMIT" {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1106, "msg": "只支持 LIMIT 订单类型"})
		return
	}

	quantity, _ := strconv.ParseFloat(req.Quantity, 64)
	price, _ := strconv.ParseFloat(req.Price, 64)
	leverage := req.Leverage
	if leverage == 0 {
		leverage = 10 // 默认杠杆
	}

	order := &models.Order{
		APIKey:        apiKey,
		Symbol:        req.Symbol,
		Side:          models.Side(req.Side),
		Type:          req.Type,
		Price:         price,
		Quantity:      quantity,
		Leverage:      leverage,
		ClientOrderID: req.ClientOrderID,
		Status:        models.OrderStatusNew,
	}

	if err := s.orderStore.Create(order); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1000, "msg": err.Error()})
		return
	}

	c.JSON(http.StatusOK, order)
}

func (s *Server) cancelOrder(c *gin.Context) {
	orderID, _ := strconv.ParseInt(c.Query("orderId"), 10, 64)

	if err := s.orderStore.Cancel(orderID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1000, "msg": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"orderId": orderID,
		"status":  "CANCELLED",
	})
}

func (s *Server) getOrder(c *gin.Context) {
	// TODO: 实现查询单个订单
	c.JSON(http.StatusNotImplemented, gin.H{"code": -1000, "msg": "Not implemented"})
}

func (s *Server) getOpenOrders(c *gin.Context) {
	apiKey := c.GetString("apiKey")

	orders, err := s.orderStore.GetByAPIKey(apiKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1000, "msg": err.Error()})
		return
	}

	c.JSON(http.StatusOK, orders)
}

func (s *Server) getMyTrades(c *gin.Context) {
	// TODO: 实现查询成交历史
	c.JSON(http.StatusOK, []gin.H{})
}

func (s *Server) getExchangeInfo(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"symbols": []gin.H{
			{
				"symbol":     "BTCUSDT",
				"status":     "TRADING",
				"baseAsset":  "BTC",
				"quoteAsset": "USDT",
			},
			{
				"symbol":     "ETHUSDT",
				"status":     "TRADING",
				"baseAsset":  "ETH",
				"quoteAsset": "USDT",
			},
		},
	})
}

func (s *Server) getDepth(c *gin.Context) {
	// TODO: 返回模拟 orderbook
	c.JSON(http.StatusOK, gin.H{
		"lastUpdateId": time.Now().UnixMilli(),
		"bids":         [][]string{},
		"asks":         [][]string{},
	})
}

// ========== Dashboard API ==========

func (s *Server) getLeaderboard(c *gin.Context) {
	entries, err := s.leaderboardStore.GetLeaderboard()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, entries)
}

func (s *Server) getStrategyDetail(c *gin.Context) {
	apiKey := c.Param("apiKey")
	stats, err := s.leaderboardStore.GetStrategyStats(apiKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if stats == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Strategy not found"})
		return
	}
	c.JSON(http.StatusOK, stats)
}

func (s *Server) getStrategyTrades(c *gin.Context) {
	apiKey := c.Param("apiKey")
	trades, err := s.orderStore.GetTradesByAPIKey(apiKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, trades)
}

func (s *Server) getStrategyPositions(c *gin.Context) {
	apiKey := c.Param("apiKey")
	positions, err := s.positionStore.GetByAPIKey(apiKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, positions)
}

// getStrategyOrders 获取策略的当前委托（未成交订单）
func (s *Server) getStrategyOrders(c *gin.Context) {
	apiKey := c.Param("apiKey")
	orders, err := s.orderStore.GetByAPIKey(apiKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 只返回未成交的订单
	var openOrders []models.Order
	for _, order := range orders {
		if order.Status == models.OrderStatusNew || order.Status == models.OrderStatusPartiallyFilled {
			openOrders = append(openOrders, order)
		}
	}

	c.JSON(http.StatusOK, openOrders)
}

// getStrategySnapshots 获取策略的收益快照
func (s *Server) getStrategySnapshots(c *gin.Context) {
	apiKey := c.Param("apiKey")
	limit := 100 // 默认返回最近100条
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	snapshots, err := s.snapshotStore.GetSnapshotsByAPIKey(apiKey, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, snapshots)
}

// getOrderbook 获取订单簿深度
func (s *Server) getOrderbook(c *gin.Context) {
	symbol := c.Param("symbol")
	if symbol == "" {
		symbol = "BTCUSDT"
	}

	snapshot := s.orderbook.GetSnapshot(symbol)
	c.JSON(http.StatusOK, snapshot)
}

package api

import (
	"database/sql"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"hft-sim/internal/store"
)

type Server struct {
	router        *gin.Engine
	orderStore    *store.OrderStore
	balanceStore  *store.BalanceStore
	positionStore *store.PositionStore
}

func NewServer(db *sql.DB) *Server {
	s := &Server{
		router:        gin.Default(),
		orderStore:    store.NewOrderStore(db),
		balanceStore:  store.NewBalanceStore(db),
		positionStore: store.NewPositionStore(db),
	}

	s.setupRoutes()
	return s
}

func (s *Server) setupRoutes() {
	s.router.Use(cors.Default())

	// API v3 (CCXT 兼容)
	api := s.router.Group("/api/v3")
	{
		api.Use(s.authMiddleware())

		api.GET("/account", s.getAccount)
		api.POST("/order", s.createOrder)
		api.DELETE("/order", s.cancelOrder)
		api.GET("/order", s.getOrder)
		api.GET("/openOrders", s.getOpenOrders)
		api.GET("/myTrades", s.getMyTrades)
	}

	// Public endpoints
	s.router.GET("/api/v3/exchangeInfo", s.getExchangeInfo)
	s.router.GET("/api/v3/depth", s.getDepth)

	// WebSocket
	s.router.GET("/ws", s.handleWebSocket)

	// Static files (Dashboard)
	s.router.Static("/dashboard", "./web")

	// Redirect root to dashboard
	s.router.GET("/", func(c *gin.Context) {
		c.Redirect(302, "/dashboard")
	})
}

func (s *Server) Run(addr string) error {
	return s.router.Run(addr)
}

package api

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (s *Server) handleWebSocket(c *gin.Context) {
	apiKey := c.Query("apiKey")
	if apiKey == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"code": -2015, "msg": "Invalid API-key"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	// 订阅处理
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			break
		}

		var req struct {
			Method string   `json:"method"`
			Params []string `json:"params"`
			ID     int      `json:"id"`
		}

		if err := json.Unmarshal(message, &req); err != nil {
			continue
		}

		// 处理订阅请求
		if req.Method == "SUBSCRIBE" {
			// 发送确认
			conn.WriteJSON(gin.H{"result": nil, "id": req.ID})
		}
	}
}

package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (s *Server) authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-MBX-APIKEY")
		if apiKey == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"code": -2015, "msg": "Invalid API-key, IP, or permissions for action"})
			c.Abort()
			return
		}

		// 验证 API Key 是否存在
		// TODO: 从数据库验证

		c.Set("apiKey", apiKey)
		c.Next()
	}
}

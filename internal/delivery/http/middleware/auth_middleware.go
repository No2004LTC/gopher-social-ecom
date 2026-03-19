package middleware

import (
	"net/http"
	"strings"

	"github.com/No2004LTC/gopher-social-ecom/pkg/auth"
	"github.com/gin-gonic/gin"
)

func AuthMiddleware(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := ""

		// 1. Thử lấy từ Header Authorization: Bearer <token>
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) == 2 && parts[0] == "Bearer" {
				tokenString = parts[1]
			}
		}

		// 2. Nếu Header không có, thử lấy từ Query Param ?token=<token>
		// Cách này cực kỳ quan trọng cho kết nối WebSocket ban đầu
		if tokenString == "" {
			tokenString = c.Query("token")
		}

		if tokenString == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Thiếu token"})
			return
		}

		// 3. Validate Token
		userID, err := auth.ValidateToken(tokenString, secret)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token lỏ hoặc hết hạn"})
			return
		}

		// Lưu userID vào context để các tầng sau (Handler/Hub) sử dụng
		c.Set("user_id", userID)
		c.Next()
	}
}

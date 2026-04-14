package middleware

import (
	"net/http"

	"github.com/No2004LTC/gopher-social-ecom/pkg/auth"
	"github.com/gin-gonic/gin"
)

func AuthMiddleware(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")

		// 2. Nếu Header không có (trường hợp WebSocket), lấy từ Query URL
		if tokenString == "" {
			tokenString = c.Query("token")
		} else {
			// Nếu có Header thì cắt bỏ chữ "Bearer "
			if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
				tokenString = tokenString[7:]
			}
		}

		// Nếu Header không có, thử lấy từ Query Param ?token=<token>
		if tokenString == "" {
			c.AbortWithStatusJSON(401, gin.H{"error": "Unauthorized"})
			return
		}

		// Nếu vẫn không có token, trả về lỗi
		if tokenString == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Thiếu token"})
			return
		}

		// kiểm tra token
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

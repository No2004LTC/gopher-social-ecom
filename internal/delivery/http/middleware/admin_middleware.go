package middleware

import (
	"net/http"

	"github.com/No2004LTC/gopher-social-ecom/internal/domain"
	"github.com/gin-gonic/gin"
)

// AdminMiddleware nhận vào Repo Admin để check email trực tiếp từ DB
func AdminMiddleware(adminRepo domain.AdminRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Lấy user_id do AuthMiddleware của cậu gài vào trước đó
		userIDVal, exists := c.Get("user_id")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Vui lòng đăng nhập trước"})
			return
		}

		// 2. Ép kiểu dữ liệu (Cậu check xem ValidateToken trả về int64 hay string nhé, ở đây tớ để int64 theo chuẩn DB)
		userID, ok := userIDVal.(int64)
		if !ok {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Lỗi hệ thống khi xác thực quyền"})
			return
		}

		// 3. Dùng Repo gọi xuống DB tìm user theo ID độc lập
		user, err := adminRepo.GetByID(c.Request.Context(), userID)
		if err != nil || user == nil {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Tài khoản không tồn tại hoặc đã bị xóa"})
			return
		}

		// 4. Kiểm tra xem có đúng Email tối cao của cậu không
		if user.Email != "lethanhcong20052004@gmail.com" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Quyền truy cập bị từ chối: Bạn không phải Admin hệ thống"})
			return
		}

		// Gợi ý: Lưu luôn cục user xịn này vào context để các Handler phía sau (như hàm BanUser)
		// lấy ra dùng luôn, đỡ mất công chọc vào DB truy vấn lại một lần nữa.
		c.Set("currentUser", user)

		c.Next()
	}
}

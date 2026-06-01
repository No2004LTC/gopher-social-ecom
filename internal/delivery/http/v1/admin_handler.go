package v1

import (
	"github.com/No2004LTC/gopher-social-ecom/internal/domain"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

type AdminHandler struct {
	adminUsecase domain.AdminUsecase
}

func NewAdminHandler(au domain.AdminUsecase) *AdminHandler {
	return &AdminHandler{adminUsecase: au}
}

// 1. GET /api/v1/admin/stats -> Lấy số liệu 3 ô Card
func (h *AdminHandler) GetDashboardStats(c *gin.Context) {
	stats, err := h.adminUsecase.GetDashboardStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, stats)
}

// GET /api/v1/admin/growth -> Biểu đồ tăng trưởng 7 ngày
func (h *AdminHandler) GetGrowthStats(c *gin.Context) {
	daysStr := c.DefaultQuery("days", "7")
	days, err := strconv.Atoi(daysStr)
	if err != nil {
		days = 7
	}

	data, err := h.adminUsecase.GetGrowthStats(c.Request.Context(), days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}

// 2. GET /api/v1/admin/users -> Lấy danh sách thành viên + Tìm kiếm
func (h *AdminHandler) GetAllUsers(c *gin.Context) {
	keyword := c.Query("keyword") // Lấy từ query param: ?keyword=lethanh

	users, err := h.adminUsecase.GetAllUsers(c.Request.Context(), keyword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, users)
}

// 3. PUT /api/v1/admin/users/:id/ban -> Hạ lệnh khóa tài khoản
func (h *AdminHandler) BanUser(c *gin.Context) {
	// Lấy ID user cần BAN từ URL biến động (:id)
	targetIDStr := c.Param("id")
	targetID, err := strconv.ParseInt(targetIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID người dùng không hợp lệ"})
		return
	}

	// Lấy email của chính admin đang thực hiện lệnh (để tránh tự BAN bản thân)
	val, _ := c.Get("currentUser")
	admin := val.(*domain.User)

	err = h.adminUsecase.BanUser(c.Request.Context(), targetID, admin.Email)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "đã khóa tài khoản thành công"})
}

// PUT /api/v1/admin/users/:id/unban -> Gỡ khóa tài khoản
func (h *AdminHandler) UnbanUser(c *gin.Context) {
	targetIDStr := c.Param("id")
	targetID, err := strconv.ParseInt(targetIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID người dùng không hợp lệ"})
		return
	}

	err = h.adminUsecase.UnbanUser(c.Request.Context(), targetID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "đã gỡ khóa tài khoản thành công"})
}

// 4. GET /api/v1/admin/posts -> Dòng thời gian rà soát nội dung bài viết
func (h *AdminHandler) GetModerationFeed(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "20")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 20
	}

	posts, err := h.adminUsecase.GetModerationFeed(c.Request.Context(), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, posts)
}

// 5. DELETE /api/v1/admin/posts/:id -> Admin xóa bài viết bất kỳ
func (h *AdminHandler) DeletePost(c *gin.Context) {
	postIDStr := c.Param("id")
	postID, err := strconv.ParseInt(postIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID bài viết không hợp lệ"})
		return
	}

	err = h.adminUsecase.AdminDeletePost(c.Request.Context(), postID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "đã xóa bài viết thành công"})
}

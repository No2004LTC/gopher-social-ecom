package v1

import (
	"net/http"
	"strconv"

	"github.com/No2004LTC/gopher-social-ecom/internal/delivery/http/response"
	"github.com/No2004LTC/gopher-social-ecom/internal/domain"
	"github.com/No2004LTC/gopher-social-ecom/internal/dto"
	"github.com/gin-gonic/gin"
)

type NotificationHandler struct {
	notiUC domain.NotificationUsecase
}

func NewNotificationHandler(notiUC domain.NotificationUsecase) *NotificationHandler {
	return &NotificationHandler{notiUC: notiUC}
}

func (h *NotificationHandler) GetNotifications(c *gin.Context) {
	// Lấy ID từ token
	uid, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Không tìm thấy thông tin xác thực")
		return
	}
	userID := uid.(int64)

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	// Gọi Usecase
	notifications, err := h.notiUC.GetUserNotifications(c.Request.Context(), userID, limit, offset)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Lỗi khi tải thông báo")
		return
	}

	// Trả về cho Frontend
	response.Success(c, "Lấy thông báo thành công", notifications)
}

func (h *PostHandler) GetDiscoveryFeed(c *gin.Context) {
	// 1. Lấy thông tin từ Context và Query
	userID := c.MustGet("user_id").(int64)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	// 2. Gọi Usecase mới (Hàm GetPosts vạn năng)
	// targetUserID = 0 để lấy bài viết của tất cả mọi người (Discovery/Global Feed)
	posts, err := h.postUC.GetPosts(c.Request.Context(), userID, 0, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 3. Mapping sang DTO PostResponse để trả về cho FE (Giúp hiện tim đỏ, bookmark vàng)
	response := make([]dto.PostResponse, 0)
	for _, p := range posts {
		// Khởi tạo Author mặc định để tránh nil pointer
		authorData := dto.ActorCompact{
			ID:        0,
			Username:  "Unknown",
			AvatarURL: "",
		}

		if p.User != nil {
			authorData.ID = p.User.ID
			authorData.Username = p.User.Username
			authorData.AvatarURL = p.User.AvatarURL
		}

		response = append(response, dto.PostResponse{
			ID:            p.ID,
			Content:       p.Content,
			ImageURL:      p.ImageURL,
			LikesCount:    p.LikesCount,
			CommentsCount: p.CommentsCount,
			IsLiked:       p.IsLiked,
			IsSaved:       p.IsSaved,
			CreatedAt:     p.CreatedAt,
			Author:        authorData,
		})
	}

	// 4. Trả về đúng format mà NewsFeed.tsx đang mong đợi
	c.JSON(http.StatusOK, gin.H{
		"data": response,
	})
}

func (h *NotificationHandler) MarkAsRead(c *gin.Context) {
	notiID, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	err := h.notiUC.MarkAsRead(c.Request.Context(), notiID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Đã đọc thông báo"})
}

// [GET] /api/v1/notifications/unread-count -> Lấy số lượng thông báo chưa đọc
func (h *NotificationHandler) GetUnreadCount(c *gin.Context) {
	uid, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Không tìm thấy thông tin xác thực")
		return
	}
	userID := uid.(int64)

	count, err := h.notiUC.GetUnreadCount(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Lỗi khi đếm thông báo: "+err.Error())
		return
	}

	response.Success(c, "Lấy số lượng thành công", gin.H{
		"unread_count": count,
	})
}

// [PUT] /api/v1/notifications/read-all
func (h *NotificationHandler) MarkAllAsRead(c *gin.Context) {
	uid, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Không tìm thấy thông tin xác thực")
		return
	}
	userID := uid.(int64)

	err := h.notiUC.MarkAllAsRead(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Lỗi cập nhật thông báo")
		return
	}

	response.Success(c, "Đã đánh dấu tất cả", nil)
}

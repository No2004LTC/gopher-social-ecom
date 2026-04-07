package v1

import (
	"net/http"
	"strconv"

	"github.com/No2004LTC/gopher-social-ecom/internal/domain"
	"github.com/gin-gonic/gin"
)

type NotificationHandler struct {
	notiUC domain.NotificationUsecase
}

func NewNotificationHandler(notiUC domain.NotificationUsecase) *NotificationHandler {
	return &NotificationHandler{notiUC: notiUC}
}

func (h *NotificationHandler) GetNotifications(c *gin.Context) {
	userID := c.MustGet("user_id").(int64)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))

	notifications, err := h.notiUC.GetNotifications(c.Request.Context(), userID, page)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, notifications)
}

func (h *PostHandler) GetDiscoveryFeed(c *gin.Context) {
	userID := c.MustGet("user_id").(int64)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))

	posts, err := h.postUC.GetDiscoveryFeed(c.Request.Context(), userID, page)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, posts)
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

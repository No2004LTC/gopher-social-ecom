package v1

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/No2004LTC/gopher-social-ecom/internal/delivery/ws"
	"github.com/No2004LTC/gopher-social-ecom/internal/domain"
	"github.com/gin-gonic/gin"
)

type WSHandler struct {
	hub    *ws.Hub
	chatUC domain.ChatUsecase
}

func NewWSHandler(hub *ws.Hub, chatUC domain.ChatUsecase) *WSHandler {
	return &WSHandler{
		hub:    hub,
		chatUC: chatUC,
	}
}

// ServeWS
func (h *WSHandler) ServeWS(c *gin.Context) {
	userID := c.MustGet("user_id").(int64)
	ws.ServeWS(h.hub, c.Writer, c.Request, userID)
}

// SendMessage
func (h *WSHandler) SendMessage(c *gin.Context) {
	userID := c.MustGet("user_id").(int64)

	var req struct {
		ToUserID int64  `json:"to_user_id" binding:"required"`
		Content  string `json:"content" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	msg := &domain.Message{
		FromUserID: userID,
		ToUserID:   req.ToUserID,
		Content:    req.Content,
		CreatedAt:  time.Now(),
		IsRead:     false,
	}

	err := h.chatUC.SaveMessage(c.Request.Context(), msg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	payload, _ := json.Marshal(map[string]interface{}{
		"type": "NEW_MESSAGE",
		"data": msg,
	})
	h.hub.SendToUser(req.ToUserID, payload)

	c.JSON(http.StatusOK, msg)
}

// GetHistory
func (h *WSHandler) GetHistory(c *gin.Context) {
	userID := c.MustGet("user_id").(int64)
	toUserID, _ := strconv.ParseInt(c.Param("to_user_id"), 10, 64)

	messages, err := h.chatUC.GetChatHistory(c.Request.Context(), userID, toUserID, 50)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, messages)
}

// GetConversations
func (h *WSHandler) GetConversations(c *gin.Context) {
	userID := c.MustGet("user_id").(int64)

	categorized, err := h.chatUC.GetCategorizedConversations(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, categorized)
}

// GetUnreadCount
func (h *WSHandler) GetUnreadCount(c *gin.Context) {
	userID := c.MustGet("user_id").(int64)

	count, err := h.chatUC.GetUnreadCount(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"unread_count": count})
}

// MarkAsRead
func (h *WSHandler) MarkAsRead(c *gin.Context) {
	userID := c.MustGet("user_id").(int64)
	partnerID, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	err := h.chatUC.MarkMessagesAsRead(c.Request.Context(), userID, partnerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "success"})
}

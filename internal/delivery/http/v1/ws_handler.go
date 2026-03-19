package v1

import (
	"net/http"
	"strconv" // Cần để parse ID từ URL

	"github.com/No2004LTC/gopher-social-ecom/internal/delivery/ws"
	"github.com/No2004LTC/gopher-social-ecom/internal/domain" // Thêm domain
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type WSHandler struct {
	Hub    *ws.Hub
	ChatUC domain.ChatUsecase // Cần cái này để lấy lịch sử chat từ DB
}

// Cập nhật hàm NewWSHandler để nhận thêm chatUC
func NewWSHandler(hub *ws.Hub, chatUC domain.ChatUsecase) *WSHandler {
	return &WSHandler{
		Hub:    hub,
		ChatUC: chatUC,
	}
}

func (h *WSHandler) ServeWS(c *gin.Context) {
	userID := c.MustGet("user_id").(int64)

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	client := &ws.Client{
		Hub:    h.Hub,
		Conn:   conn,
		Send:   make(chan []byte, 256),
		UserID: userID,
	}

	h.Hub.Register <- client

	go client.WritePump()
	go client.ReadPump()
}

// API lấy lịch sử chat
func (h *WSHandler) GetHistory(c *gin.Context) {
	userID := c.MustGet("user_id").(int64)

	// Lấy to_user_id từ URL: /api/v1/chats/:to_user_id
	toUserID, err := strconv.ParseInt(c.Param("to_user_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID người nhận không hợp lệ"})
		return
	}

	// Gọi Usecase lấy 50 tin nhắn gần nhất
	messages, err := h.ChatUC.GetChatHistory(c.Request.Context(), userID, toUserID, 50)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, messages)
}

package v1

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/No2004LTC/gopher-social-ecom/internal/delivery/http/response"
	"github.com/No2004LTC/gopher-social-ecom/internal/delivery/ws"
	"github.com/No2004LTC/gopher-social-ecom/internal/domain"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Bỏ qua CORS
	},
}

type WSHandler struct {
	Hub    *ws.Hub
	ChatUC domain.ChatUsecase
}

func NewWSHandler(hub *ws.Hub, chatUC domain.ChatUsecase) *WSHandler {
	return &WSHandler{
		Hub:    hub,
		ChatUC: chatUC,
	}
}

// [GET] /api/v1/ws -> Đầu cầu mở ống WebSocket
func (h *WSHandler) ServeWS(c *gin.Context) {
	log.Println("🚀 ServeWS đang được gọi...") // Thêm dòng này

	userID := c.MustGet("user_id").(int64)
	log.Printf("👤 User %d đang cố gắng nâng cấp lên WS", userID)

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("❌ Lỗi Upgrade WS: %v", err)
		return
	}

	client := &ws.Client{
		Hub:    h.Hub,
		Conn:   conn,
		Send:   make(chan []byte, 256),
		UserID: userID,
	}

	log.Printf("📡 Gửi tín hiệu Register cho User %d vào Hub", userID)
	h.Hub.Register <- client

	go client.WritePump()
	go client.ReadPump()
}

// [POST] /api/v1/chats -> GỬI TIN NHẮN (HYBRID)
func (h *WSHandler) SendMessage(c *gin.Context) {
	fromUserID := c.MustGet("user_id").(int64)

	var req struct {
		ToUserID int64  `json:"to_user_id" binding:"required"`
		Content  string `json:"content" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dữ liệu không hợp lệ"})
		return
	}

	msg := &domain.Message{
		FromUserID: fromUserID,
		ToUserID:   req.ToUserID,
		Content:    req.Content,
	}

	// 1. LƯU XUỐNG DATABASE
	if err := h.ChatUC.SaveMessage(c.Request.Context(), msg); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể gửi tin nhắn"})
		return
	}

	// 2. BẮN TIN NHẮN QUA WEBSOCKET CHO NGƯỜI NHẬN (REAL-TIME)
	// Đóng gói data có gắn thêm type để FE dễ xử lý
	wsPayload := map[string]interface{}{
		"type": "NEW_MESSAGE",
		"data": msg,
	}
	payloadBytes, _ := json.Marshal(wsPayload)

	// Gọi hàm gửi đích danh của Hub (Tớ sẽ hướng dẫn thêm hàm này vào Hub ở dưới)
	h.Hub.SendToUser(req.ToUserID, payloadBytes)

	// 3. Trả kết quả HTTP 200 cho người gửi
	c.JSON(http.StatusOK, gin.H{"message": "Đã gửi", "data": msg})
}

// [GET] /api/v1/chats/:to_user_id -> LẤY LỊCH SỬ CHAT
func (h *WSHandler) GetHistory(c *gin.Context) {
	userID := c.MustGet("user_id").(int64)

	toUserID, err := strconv.ParseInt(c.Param("to_user_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID người nhận không hợp lệ"})
		return
	}

	messages, err := h.ChatUC.GetChatHistory(c.Request.Context(), userID, toUserID, 50)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": messages})
}

// [GET] /api/v1/chats/unread-count -> Đếm tổng số tin nhắn chưa đọc
func (h *WSHandler) GetUnreadCount(c *gin.Context) {
	// 1. Lấy userID từ Token
	uid, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Không tìm thấy thông tin xác thực")
		return
	}
	userID := uid.(int64) // Ép kiểu tùy thuộc vào middleware của cậu

	// 2. Gọi Usecase
	count, err := h.ChatUC.GetUnreadCount(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Lỗi khi đếm tin nhắn chưa đọc: "+err.Error())
		return
	}

	// 3. Trả về Frontend
	response.Success(c, "Lấy số lượng thành công", gin.H{
		"unread_count": count,
	})
}

// [GET] /api/v1/chats/conversations -> Lấy danh sách hội thoại (Đã chia Bạn Bè / Người Lạ)
func (h *WSHandler) GetConversations(c *gin.Context) {
	// 1. Lấy userID
	uid, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Không tìm thấy thông tin xác thực")
		return
	}
	userID := uid.(int64)

	// 2. Gọi Usecase phân loại
	categorizedConvos, err := h.ChatUC.GetCategorizedConversations(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Lỗi khi tải danh sách hội thoại: "+err.Error())
		return
	}

	// 3. Trả về Frontend. Lúc này data sẽ có dạng { "friends": [...], "strangers": [...] }
	response.Success(c, "Lấy danh sách hội thoại thành công", categorizedConvos)
}

// [PUT] /api/v1/chats/:id/read -> Đánh dấu đã đọc tin nhắn của một người
func (h *WSHandler) MarkAsRead(c *gin.Context) {
	uid, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Không tìm thấy thông tin xác thực")
		return
	}
	myUserID := uid.(int64)

	partnerIDStr := c.Param("id")
	partnerID, err := strconv.ParseInt(partnerIDStr, 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "ID người dùng không hợp lệ")
		return
	}

	err = h.ChatUC.MarkMessagesAsRead(c.Request.Context(), myUserID, partnerID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Lỗi cập nhật trạng thái tin nhắn")
		return
	}

	response.Success(c, "Đã xem", nil)
}

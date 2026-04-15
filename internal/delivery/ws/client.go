package ws

import (
	"context"
	"log"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

type Client struct {
	Hub    *Hub
	Conn   *websocket.Conn
	Send   chan []byte
	UserID int64
}

func (c *Client) ReadPump() {
	defer func() {
		// 1. Gửi tín hiệu hủy đăng ký cho Hub
		c.Hub.Unregister <- c

		// 2. 👉 REDIS: XỬ LÝ TRẠNG THÁI OFFLINE (GIẢM COUNT)
		// Lưu ý: Cậu cần đảm bảo struct Hub đã được gắn thêm trường Redis (*redis.Client)
		if c.Hub.Redis != nil {
			ctx := context.Background()
			userIDStr := strconv.FormatInt(c.UserID, 10)

			// Giảm số lượng connection (tab) đi 1
			newCount, err := c.Hub.Redis.HIncrBy(ctx, "system:online_users", userIDStr, -1).Result()
			if err == nil && newCount <= 0 {
				// Nếu count <= 0 nghĩa là user đã đóng tab cuối cùng -> Xóa hẳn khỏi map
				c.Hub.Redis.HDel(ctx, "system:online_users", userIDStr)
			}
		}

		// 3. Đóng ống nối
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(maxMessageSize)
	// Đặt thời gian timeout, nếu sau pongWait mà không thấy trình duyệt phản hồi -> Rớt mạng
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error { c.Conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	for {
		// Chỉ đọc để giữ kết nối, bỏ qua payload vì Frontend không gửi chat qua ống này nữa
		_, _, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Lỗi kết nối WS: %v", err)
			}
			break // Thoát vòng lặp -> Chạy defer để ngắt kết nối và trừ count Redis
		}
	}
}

func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)
			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			// Gửi gói tin Ping định kỳ để giữ kết nối không bị Proxy/Nginx ngắt
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

package ws

import (
	"context"
	"encoding/json"
	"log"
	"strconv"
	"sync"

	"github.com/No2004LTC/gopher-social-ecom/internal/domain"
	"github.com/redis/go-redis/v9"
)

type Hub struct {
	Clients    sync.Map
	Broadcast  chan []byte
	Register   chan *Client
	Unregister chan *Client
	Redis      *redis.Client
}

func NewHub(rdb *redis.Client) *Hub {
	return &Hub{
		Broadcast:  make(chan []byte, 256),
		Register:   make(chan *Client, 100),
		Unregister: make(chan *Client, 100),
		Redis:      rdb,
	}
}

func (h *Hub) SendToUser(userID int64, payload []byte) {
	ctx := context.Background()
	envelope := map[string]interface{}{
		"to_user_id": userID,
		"payload":    string(payload),
	}
	data, _ := json.Marshal(envelope)
	h.Redis.Publish(ctx, "system:ws_messages", data)
}

func (h *Hub) BroadcastNotification(noti domain.Notification) {
	ctx := context.Background()
	payload, _ := json.Marshal(noti)
	h.Redis.Publish(ctx, "system:notifications", payload)
}

func (h *Hub) Run() {
	go func() {
		ctx := context.Background()
		// 🚀 THÊM "system:status_change" VÀO DANH SÁCH ĐĂNG KÝ
		pubsub := h.Redis.Subscribe(ctx, "system:notifications", "system:ws_messages", "system:status_change")
		defer pubsub.Close()

		for msg := range pubsub.Channel() {
			switch msg.Channel {
			case "system:notifications":
				var noti domain.Notification
				if err := json.Unmarshal([]byte(msg.Payload), &noti); err == nil {
					h.sendLocal(noti.UserID, "NOTIFICATION", noti)
				} else {
					log.Printf("❌ Lỗi parse Notification từ Redis: %v", err)
				}

			case "system:ws_messages":
				var env struct {
					ToUserID int64  `json:"to_user_id"`
					Payload  string `json:"payload"`
				}
				if err := json.Unmarshal([]byte(msg.Payload), &env); err == nil {
					h.sendRawLocal(env.ToUserID, []byte(env.Payload))
				}

			// 🚀 HỨNG SỰ KIỆN ONLINE/OFFLINE TỪ REDIS VÀ BẮN XUỐNG TẤT CẢ CLIENT
			case "system:status_change":
				h.Broadcast <- []byte(msg.Payload)
			}
		}
	}()

	for {
		select {
		case client := <-h.Register:
			h.Clients.Store(client.UserID, client)
			h.updateOnlineStatus(client.UserID, true)

		case client := <-h.Unregister:
			// 🚀 BẢO VỆ CHỐNG XÓA NHẦM TAB MỚI
			if val, ok := h.Clients.Load(client.UserID); ok {
				currentClient := val.(*Client)
				// CHỈ xóa khỏi map nếu client đang đòi ngắt ĐÚNG LÀ client đang lưu
				if currentClient == client {
					h.Clients.Delete(client.UserID)
				}
			}
			close(client.Send)
			h.updateOnlineStatus(client.UserID, false)

		case message := <-h.Broadcast:
			h.Clients.Range(func(_, value interface{}) bool {
				client := value.(*Client)
				select {
				case client.Send <- message:
				default:
					log.Printf("⚠️ Hàng đợi của User %d đang đầy, drop tin nhắn broadcast", client.UserID)
				}
				return true
			})
		}
	}
}

func (h *Hub) sendLocal(userID int64, eventType string, data interface{}) {
	if val, ok := h.Clients.Load(userID); ok {
		client := val.(*Client)
		msg, _ := json.Marshal(map[string]interface{}{
			"type": eventType,
			"data": data,
		})

		select {
		case client.Send <- msg:
		default:
			log.Printf("⚠️ Không thể gửi %s cho User %d do kẹt kênh", eventType, userID)
		}
	}
}

func (h *Hub) sendRawLocal(userID int64, payload []byte) {
	if val, ok := h.Clients.Load(userID); ok {
		client := val.(*Client)

		select {
		case client.Send <- payload:
		default:
			log.Printf("⚠️ Không thể gửi Chat cho User %d do kẹt kênh", userID)
		}
	}
}

func (h *Hub) updateOnlineStatus(userID int64, isOnline bool) {
	if h.Redis == nil {
		return
	}
	ctx := context.Background()
	userIDStr := strconv.FormatInt(userID, 10)

	status := "offline"
	if isOnline {
		h.Redis.HIncrBy(ctx, "system:online_users", userIDStr, 1)
		status = "online"
	} else {
		val, _ := h.Redis.HIncrBy(ctx, "system:online_users", userIDStr, -1).Result()
		if val <= 0 {
			h.Redis.HDel(ctx, "system:online_users", userIDStr)
			status = "offline"
		} else {
			if _, exists := h.Clients.Load(userID); !exists {
				h.Redis.HDel(ctx, "system:online_users", userIDStr)
				status = "offline"
			} else {
				status = "online"
			}
		}
	}

	statusNotify, _ := json.Marshal(map[string]interface{}{
		"type": "USER_STATUS_CHANGE",
		"data": map[string]interface{}{"user_id": userID, "status": status},
	})

	// 🚀 BỎ h.Broadcast CŨ ĐI, BẮN THẲNG LÊN REDIS ĐỂ LAN TRUYỀN TOÀN HỆ THỐNG
	h.Redis.Publish(ctx, "system:status_change", statusNotify)
}

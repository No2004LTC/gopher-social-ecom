package ws

import (
	"context"
	"encoding/json"
	"log"
	"strconv"
	"sync"

	"github.com/No2004LTC/gopher-social-ecom/internal/domain"
	"github.com/redis/go-redis/v9" // 👉 NHỚ IMPORT REDIS
)

type Hub struct {
	// Dùng sync.Map cực an toàn
	Clients       sync.Map
	Broadcast     chan []byte
	Notifications chan domain.Notification
	Register      chan *Client
	Unregister    chan *Client

	// 👉 ĐÃ THÊM: Con trỏ Redis để các Client có thể dùng chung
	Redis *redis.Client
}

// 👉 CẬP NHẬT CONSTRUCTOR: Bắt buộc truyền Redis vào khi tạo Hub
func NewHub(rdb *redis.Client) *Hub {
	return &Hub{
		Broadcast:     make(chan []byte),
		Register:      make(chan *Client),
		Unregister:    make(chan *Client),
		Notifications: make(chan domain.Notification, 100),
		Redis:         rdb, // Gắn Redis vào Hub
	}
}

// Hàm gửi tin nhắn trực tiếp đến 1 User cụ thể
func (h *Hub) SendToUser(userID int64, message []byte) {
	if val, ok := h.Clients.Load(userID); ok {
		client := val.(*Client)
		select {
		case client.Send <- message:
		default:
			h.Clients.Delete(userID)
			close(client.Send)
		}
	}
}

func (h *Hub) BroadcastNotification(noti domain.Notification) {
	select {
	case h.Notifications <- noti:
	default:
		log.Printf("Cảnh báo: Hàng chờ thông báo đã đầy, bỏ qua thông báo cho User %d", noti.UserID)
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.Clients.Store(client.UserID, client)
			if h.Redis != nil {
				ctx := context.Background()
				h.Redis.HIncrBy(ctx, "system:online_users", strconv.FormatInt(client.UserID, 10), 1)

				// 👉 PHÁT TÍN HIỆU: Báo cho tất cả mọi người là có thay đổi trạng thái online
				statusNotify, _ := json.Marshal(map[string]interface{}{
					"type": "USER_STATUS_CHANGE",
					"data": map[string]interface{}{
						"user_id": client.UserID,
						"status":  "online",
					},
				})
				h.Broadcast <- statusNotify // Gửi cho tất cả client đang kết nối
			}

		case client := <-h.Unregister:
			if _, ok := h.Clients.Load(client.UserID); ok {
				h.Clients.Delete(client.UserID)
				close(client.Send)

				if h.Redis != nil {
					ctx := context.Background()
					userIDStr := strconv.FormatInt(client.UserID, 10)
					newCount, _ := h.Redis.HIncrBy(ctx, "system:online_users", userIDStr, -1).Result()
					if newCount <= 0 {
						h.Redis.HDel(ctx, "system:online_users", userIDStr)

						// 👉 PHÁT TÍN HIỆU: Báo cho mọi người là User này đã offline
						statusNotify, _ := json.Marshal(map[string]interface{}{
							"type": "USER_STATUS_CHANGE",
							"data": map[string]interface{}{
								"user_id": client.UserID,
								"status":  "offline",
							},
						})
						h.Broadcast <- statusNotify
					}
				}
			}

		case message := <-h.Broadcast:
			h.Clients.Range(func(key, value interface{}) bool {
				client := value.(*Client)
				client.Send <- message
				return true
			})

		case noti := <-h.Notifications:
			if val, ok := h.Clients.Load(noti.UserID); ok {
				client := val.(*Client)
				payload, _ := json.Marshal(map[string]interface{}{
					"type": "NOTIFICATION",
					"data": noti,
				})
				select {
				case client.Send <- payload:
				default:
					log.Printf("Không thể gửi thông báo cho User %d", noti.UserID)
				}
			}
		}
	}
}

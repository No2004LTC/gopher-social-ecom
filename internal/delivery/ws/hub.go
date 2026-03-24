package ws

import (
	"context"
	"encoding/json"
	"log"
	"sync"

	"github.com/No2004LTC/gopher-social-ecom/internal/domain"
)

type Hub struct {
	Clients        sync.Map
	Broadcast      chan []byte
	PrivateMessage chan domain.Message
	Notifications  chan domain.Notification
	Register       chan *Client // Thêm cái này
	Unregister     chan *Client // Và cái này
	ChatUC         domain.ChatUsecase
}

func NewHub(chatUC domain.ChatUsecase) *Hub {
	return &Hub{
		Broadcast:      make(chan []byte),
		PrivateMessage: make(chan domain.Message),
		Register:       make(chan *Client), // Khởi tạo
		Unregister:     make(chan *Client), // Khởi tạo
		Notifications:  make(chan domain.Notification, 100),
		ChatUC:         chatUC,
	}
}

func (h *Hub) BroadcastNotification(noti domain.Notification) {
	// Dùng select với default để nếu channel đầy thì cũng không làm treo App
	select {
	case h.Notifications <- noti:
		// Gửi thành công vào hàng chờ
	default:
		log.Printf("Cảnh báo: Hàng chờ thông báo đã đầy, bỏ qua thông báo cho User %d", noti.UserID)
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			// Lưu client vào bản đồ dựa trên UserID
			h.Clients.Store(client.UserID, client)
			log.Printf("User %d đã kết nối", client.UserID)

		case client := <-h.Unregister:
			// Khi user ngắt kết nối, xóa họ khỏi Map và đóng channel Send
			if _, ok := h.Clients.Load(client.UserID); ok {
				h.Clients.Delete(client.UserID)
				close(client.Send)
				log.Printf("User %d đã thoát", client.UserID)
			}

		case msg := <-h.PrivateMessage:
			// --- BƯỚC 1: LƯU VÀO DATABASE (Chạy ngầm) ---
			go func(m domain.Message) {
				err := h.ChatUC.SaveMessage(context.Background(), &m)
				if err != nil {
					log.Printf("Lỗi khi lưu tin nhắn: %v", err)
				}
			}(msg)

			// --- BƯỚC 2: ĐẨY TIN NHẮN REAL-TIME ---
			// Gửi cho người nhận
			if val, ok := h.Clients.Load(msg.ToUserID); ok {
				targetClient := val.(*Client)
				payload, _ := json.Marshal(msg)
				targetClient.Send <- payload
			}

			// Gửi cho người gửi (để đồng bộ giao diện)
			if val, ok := h.Clients.Load(msg.FromUserID); ok {
				senderClient := val.(*Client)
				payload, _ := json.Marshal(msg)
				senderClient.Send <- payload
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
				// Gói lại thành một chuẩn JSON để Frontend dễ phân biệt với tin nhắn chat
				payload, _ := json.Marshal(map[string]interface{}{
					"type": "NOTIFICATION",
					"data": noti,
				})
				select {
				case client.Send <- payload:
				default:
					log.Printf("Không thể gửi thông báo cho User %d, hàng chờ của client đã đầy", noti.UserID)
				}
			}
		}
	}
}

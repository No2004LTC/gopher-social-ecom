package ws

import (
	"sync"

	"github.com/gorilla/websocket"
)

// WsManager quản lý kết nối an toàn với concurrency
type WsManager struct {
	// Dùng sync.Map cực kỳ quan trọng để tránh crash khi nhiều user truy cập cùng lúc
	clients sync.Map
}

func NewWsManager() *WsManager {
	return &WsManager{}
}

// Thêm User vào trạm khi họ online
func (m *WsManager) AddClient(userID int64, conn *websocket.Conn) {
	m.clients.Store(userID, conn)
}

// Rút ống khi User thoát app/mất mạng
func (m *WsManager) RemoveClient(userID int64) {
	m.clients.Delete(userID)
}

// Bắn thông báo chuẩn JSON xuống React
func (m *WsManager) SendToUser(userID int64, eventType string, payload interface{}) error {
	conn, ok := m.clients.Load(userID)
	if !ok {
		return nil // User đang offline, bỏ qua không lỗi lầm gì
	}

	wsConn := conn.(*websocket.Conn)

	// Đóng gói data có Type để React dễ if/else
	message := map[string]interface{}{
		"type": eventType,
		"data": payload,
	}

	// Gửi JSON xuống ống
	return wsConn.WriteJSON(message)
}

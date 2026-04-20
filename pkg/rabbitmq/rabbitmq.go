package rabbitmq

import (
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Hàm này trả về Connection và Channel để gửi/nhận tin nhắn
func ConnectRabbitMQ(url string) (*amqp.Connection, *amqp.Channel) {
	conn, err := amqp.Dial(url)
	if err != nil {
		log.Fatalf("❌ Không thể kết nối RabbitMQ: %v", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("❌ Không thể mở Channel: %v", err)
	}

	// Khai báo một cái hàng đợi (Queue) tên là "user_interactions"
	// Nếu chưa có nó sẽ tự tạo, nếu có rồi nó dùng lại.
	_, err = ch.QueueDeclare(
		"user_interactions", // Tên Queue
		true,                // Durable (Lưu xuống ổ cứng, khởi động lại ko mất)
		false,               // Delete when unused
		false,               // Exclusive
		false,               // No-wait
		nil,                 // Arguments
	)
	if err != nil {
		log.Fatalf("❌ Lỗi khai báo Queue: %v", err)
	}

	return conn, ch
}

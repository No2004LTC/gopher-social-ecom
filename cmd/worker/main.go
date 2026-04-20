package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/No2004LTC/gopher-social-ecom/config"
	"github.com/No2004LTC/gopher-social-ecom/pkg/rabbitmq"
)

func main() {
	fmt.Println("🚀 Bắt đầu khởi động Background Worker...")

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("❌ Lỗi load config: %v", err)
	}

	// 1. Kết nối RabbitMQ
	conn, ch := rabbitmq.ConnectRabbitMQ(cfg.RabbitMQUrl)
	defer conn.Close()
	defer ch.Close()

	// 2. Đăng ký nhận tin nhắn từ Queue "user_interactions"
	msgs, err := ch.Consume(
		"user_interactions", // Tên Queue
		"",                  // Consumer Name (để trống tự random)
		true,                // Auto-Ack (Tự động báo là đã xử lý xong)
		false,               // Exclusive
		false,               // No-local
		false,               // No-wait
		nil,                 // Args
	)
	if err != nil {
		log.Fatalf("❌ Lỗi đăng ký Consumer: %v", err)
	}

	fmt.Println("🎧 Worker đang chầu chực nghe ngóng ở Queue 'user_interactions'...")

	// 3. Vòng lặp chờ tin nhắn tới
	go func() {
		for d := range msgs {
			// KHI CÓ TIN NHẮN TỚI, NÓ SẼ CHẠY VÀO ĐÂY
			fmt.Printf("🔥 [WORKER ĐÃ BẮT ĐƯỢC]: %s\n", d.Body)
			// Tương lai cậu sẽ gọi hàm Update Database ở đây
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	fmt.Println("\n🛑 Đang dọn dẹp và tắt Worker an toàn...")
}

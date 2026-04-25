package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/No2004LTC/gopher-social-ecom/config"
	"github.com/No2004LTC/gopher-social-ecom/internal/delivery/ws"
	"github.com/No2004LTC/gopher-social-ecom/internal/domain"
	"github.com/No2004LTC/gopher-social-ecom/internal/repository/postgres"
	"github.com/No2004LTC/gopher-social-ecom/internal/usecase"
	"github.com/No2004LTC/gopher-social-ecom/pkg/db"
	"github.com/redis/go-redis/v9"
	segmentio "github.com/segmentio/kafka-go"
)

func main() {
	fmt.Println("🚀 Khởi động Kafka Interaction Worker...")

	// CẤU HÌNH
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("❌ Lỗi load config: %v", err)
	}

	database, err := db.ConnectDB(cfg)
	if err != nil {
		log.Fatalf("❌ Lỗi kết nối Database: %v", err)
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	hub := ws.NewHub(redisClient)

	notiRepo := postgres.NewNotificationRepository(database)
	notiUC := usecase.NewNotificationUsecase(notiRepo, hub)
	interactionRepo := postgres.NewInteractionRepository(database)

	reader := segmentio.NewReader(segmentio.ReaderConfig{
		Brokers:  []string{cfg.KafkaBroker},
		Topic:    "user_interactions",
		GroupID:  "interaction-worker-group",
		MinBytes: 1,
		MaxBytes: 1e6,
	})

	fmt.Println("🎧 Worker đang chầu chực nghe ngóng Kafka...")

	go func() {
		for {
			ctx := context.Background()
			m, err := reader.FetchMessage(ctx)
			if err != nil {
				log.Printf("❌ Lỗi khi fetch tin nhắn Kafka: %v", err)
				continue
			}

			var event domain.InteractionEvent
			if err := json.Unmarshal(m.Value, &event); err != nil {
				log.Printf("❌ Lỗi parse JSON: %v", err)
				reader.CommitMessages(ctx, m)
				continue
			}

			fmt.Printf("\n🔥 [KAFKA] Đang xử lý: User %d -> %s -> Post %d\n", event.UserID, event.Action, event.PostID)

			dbErr := error(nil)
			if event.Action == "LIKE" {
				dbErr = interactionRepo.LikePost(ctx, event.UserID, event.PostID)
			} else if event.Action == "UNLIKE" {
				dbErr = interactionRepo.UnlikePost(ctx, event.UserID, event.PostID)
			}

			if dbErr != nil {
				log.Printf("❌ Lỗi Database: %v", dbErr)
				reader.CommitMessages(ctx, m)
				continue
			}

			if event.Action == "LIKE" {
				ownerID := interactionRepo.GetPostOwner(ctx, event.PostID)

				if event.UserID != ownerID {
					noti := &domain.Notification{
						UserID:   ownerID,
						ActorID:  event.UserID,
						Type:     "LIKE",
						EntityID: event.PostID,
						Message:  "đã thích bài viết của bạn.",
					}

					if err := notiUC.SendNotification(ctx, noti); err != nil {
						log.Printf("⚠️ Lỗi bắn thông báo: %v", err)
					} else {
						fmt.Printf("🔔 Đã đẩy thông báo tới User %d\n", ownerID)
					}
				}
			}

			if err := reader.CommitMessages(ctx, m); err != nil {
				log.Printf("❌ Lỗi Commit Kafka (Tin nhắn có thể bị lặp): %v", err)
			} else {
				fmt.Println("✅ Hoàn tất và Commit thành công.")
			}
		}
	}()

	// GRACEFUL SHUTDOWN
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	fmt.Println("\n🛑 Đang dọn dẹp và tắt Worker...")
	reader.Close()
}

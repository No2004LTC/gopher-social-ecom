package main

import (
	"log"

	"github.com/No2004LTC/gopher-social-ecom/config"
	"github.com/No2004LTC/gopher-social-ecom/pkg/utils"
)

func main() {
	log.Println("--- Starting Gopher-Social-Ecom App ---")

	// 1. Load cấu hình từ file .env
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("❌ Không thể load config: %v", err)
	}
	log.Println("✅ Cấu hình hệ thống: OK")

	// 2. Kết nối tới Database (Postgres)
	db, err := utils.ConnectDB(cfg)
	if err != nil {
		log.Fatalf("❌ Kết nối Database thất bại: %v", err)
	}
	log.Println("✅ Kết nối Database: THÀNH CÔNG")

	// Kiểm tra xem bảng Users có tồn tại chưa (Nếu bạn đã chạy Task 4 - Migration)
	if db.Migrator().HasTable("users") {
		log.Println("✅ Bảng 'users' đã sẵn sàng trong Database.")
	}

	// Sau này: Khởi tạo Router và chạy Server ở đây...
	log.Printf("🚀 Server sẽ lắng nghe tại cổng: %s", cfg.AppPort)
}

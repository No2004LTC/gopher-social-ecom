package main

import (
	"log"
	"os"
	"strconv"

	"github.com/No2004LTC/gopher-social-ecom/config"
	"github.com/No2004LTC/gopher-social-ecom/pkg/db"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	// 1. Load Config
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Không thể load config: %v", err)
	}

	// 2. Kết nối DB qua GORM
	db, err := db.ConnectDB(cfg)
	if err != nil {
		log.Fatalf("Kết nối DB thất bại: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Không thể lấy sql.DB: %v", err)
	}
	defer sqlDB.Close()

	// 3. Khởi tạo driver migration
	driver, err := postgres.WithInstance(sqlDB, &postgres.Config{})
	if err != nil {
		log.Fatalf("Khởi tạo driver migration thất bại: %v", err)
	}

	// 4. Trỏ tới thư mục chứa file SQL
	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations/sql",
		"postgres", driver)
	if err != nil {
		log.Fatalf("Khởi tạo instance migrate thất bại: %v", err)
	}

	// 5. Đọc tham số từ dòng lệnh
	args := os.Args
	command := "up"
	if len(args) > 1 {
		command = args[1]
	}

	switch command {
	case "up":
		log.Println("🚀 Đang chạy migration UP...")
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("❌ Lỗi UP: %v", err)
		}
	case "down":
		log.Println("⏪ Đang chạy migration DOWN (1 step)...")
		if err := m.Steps(-1); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("❌ Lỗi DOWN: %v", err)
		}
	case "drop":
		log.Println("🔥 Đang xóa sạch bảng và lịch sử migration (DROP)...")
		if err := m.Drop(); err != nil {
			log.Fatalf("❌ Lỗi DROP: %v", err)
		}
	case "force":
		if len(args) < 3 {
			log.Fatalf("❌ Lệnh force yêu cầu version (VD: go run main.go force 202403092230)")
		}
		version, err := strconv.Atoi(args[2])
		if err != nil {
			log.Fatalf("❌ Version phải là một con số: %v", err)
		}
		log.Printf("🛠️ Đang ép buộc (Force) về version: %d...\n", version)
		if err := m.Force(version); err != nil {
			log.Fatalf("❌ Lỗi FORCE: %v", err)
		}
	default:
		log.Println("✨ Không có thay đổi nào mới để cập nhật.")
	}

	log.Println("✅ Thao tác hoàn tất!")
}

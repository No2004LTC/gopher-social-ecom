package main

import (
	"github.com/No2004LTC/gopher-social-ecom/config"
	"github.com/No2004LTC/gopher-social-ecom/pkg/utils"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"log"
)

func main() {
	// 1. Load Config (Dùng chung bộ config với API)
	cfg, _ := config.LoadConfig()

	// 2. Kết nối DB
	db, _ := utils.ConnectDB(cfg)
	sqlDB, _ := db.DB()

	// 3. Khởi tạo driver migration
	driver, _ := postgres.WithInstance(sqlDB, &postgres.Config{})

	// 4. Trỏ tới thư mục chứa file SQL
	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations/sql",
		"postgres", driver)

	if err != nil {
		log.Fatal("Khởi tạo migration thất bại:", err)
	}

	// 5. Thực thi Up
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatal("Lỗi khi chạy migration:", err)
	}

	log.Println("✅ Migration hoàn tất thành công!")
}

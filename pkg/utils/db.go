package utils

import (
	"fmt"

	"github.com/No2004LTC/gopher-social-ecom/config" // Thay đổi username của bạn
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// ConnectDB nhận struct Config và trả về một đối tượng kết nối Database (*gorm.DB)
func ConnectDB(cfg *config.Config) (*gorm.DB, error) {
	// DSN (Data Source Name): Chuỗi định danh kết nối chuẩn của Postgres
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		cfg.DBHost, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBPort)

	// Mở kết nối sử dụng GORM
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return db, nil
}

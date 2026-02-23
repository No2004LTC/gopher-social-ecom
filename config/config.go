package config

import (
	"github.com/spf13/viper"
)

// Config chứa toàn bộ thông số cấu hình của ứng dụng
// mapstructure giúp Viper ánh xạ tên key từ .env vào đúng trường trong struct
type Config struct {
	DBHost     string `mapstructure:"DB_HOST"` // Struct tags để Viper biết cách ánh xạ từ file .env vào struct
	DBPort     string `mapstructure:"DB_PORT"`
	DBUser     string `mapstructure:"DB_USER"`
	DBPassword string `mapstructure:"DB_PASSWORD"`
	DBName     string `mapstructure:"DB_NAME"`
	AppPort    string `mapstructure:"APP_PORT"`
}

// LoadConfig sẽ tìm file .env và nạp giá trị vào struct Config
func LoadConfig() (config Config, err error) {
	viper.AddConfigPath(".")    // Tìm file cấu hình ở thư mục hiện tại (root)
	viper.SetConfigFile(".env") // Tên file cụ thể là .env
	viper.AutomaticEnv()        // Cho phép ghi đè bằng biến môi trường hệ thống

	err = viper.ReadInConfig() // Bắt đầu đọc file
	if err != nil {
		return
	}

	// Unmarshal chuyển đổi dữ liệu từ file vào struct
	err = viper.Unmarshal(&config)
	return
}

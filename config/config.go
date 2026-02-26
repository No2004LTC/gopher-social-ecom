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
	JWTSecret  string `mapstructure:"JWT_SECRET"`
	JWTExpiry  string `mapstructure:"JWT_EXPIRY"`

	// MinIO settings
	MinioEndpoint  string `mapstructure:"MINIO_ENDPOINT"`
	MinioAccessKey string `mapstructure:"MINIO_ACCESS_KEY"`
	MinioSecretKey string `mapstructure:"MINIO_SECRET_KEY"`
	MinioUseSSL    bool   `mapstructure:"MINIO_USE_SSL"`
	MinioBucket    string `mapstructure:"MINIO_BUCKET_NAME"`
}

// LoadConfig sẽ tìm file .env và nạp giá trị vào struct Config
func LoadConfig() (*Config, error) {
	config := &Config{}
	viper.AddConfigPath(".")    // Tìm file cấu hình ở thư mục hiện tại (root)
	viper.SetConfigFile(".env") // Tên file cụ thể là .env
	viper.AutomaticEnv()        // Cho phép ghi đè bằng biến môi trường hệ thống

	err := viper.ReadInConfig() // Bắt đầu đọc file
	if err != nil {
		return nil, err
	}

	// Unmarshal chuyển đổi dữ liệu từ file vào struct
	err = viper.Unmarshal(config)
	if err != nil {
		return nil, err
	}
	return config, nil
}

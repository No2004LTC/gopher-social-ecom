package config

import (
	"github.com/spf13/viper"
)

// Config chứa toàn bộ thông số cấu hình của ứng dụng
// mapstructure giúp Viper ánh xạ tên key từ .env vào đúng trường trong struct
type Config struct {
	DBHost     string `mapstructure:"DB_HOST"`
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

	// Thêm cấu hình SMTP Gửi Mail
	SMTPHost     string `mapstructure:"SMTP_HOST"`
	SMTPPort     string `mapstructure:"SMTP_PORT"`
	SMTPUser     string `mapstructure:"SMTP_USER"`
	SMTPPassword string `mapstructure:"SMTP_PASSWORD"`
	SenderEmail  string `mapstructure:"SENDER_EMAIL"`

	//Hứng deploy
	Env string `mapstructure:"ENV"`

	// KafkaBroker
	KafkaBroker string `mapstructure:"KAFKA_BROKER"`
}

// LoadConfig sẽ tìm file .env và nạp giá trị vào struct Config
func LoadConfig() (*Config, error) {
	config := &Config{}
	viper.AddConfigPath(".")
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err != nil {
		return nil, err
	}

	err = viper.Unmarshal(config)
	if err != nil {
		return nil, err
	}
	return config, nil
}

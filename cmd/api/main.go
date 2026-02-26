package main

import (
	"log"

	"github.com/No2004LTC/gopher-social-ecom/config"
	"github.com/No2004LTC/gopher-social-ecom/internal/delivery/http/middleware"
	"github.com/No2004LTC/gopher-social-ecom/internal/delivery/http/v1"
	"github.com/No2004LTC/gopher-social-ecom/internal/repository/postgres"
	"github.com/No2004LTC/gopher-social-ecom/internal/usecase"
	"github.com/No2004LTC/gopher-social-ecom/pkg/utils"
	"github.com/gin-gonic/gin"
)

func main() {
	// 1. Load Config
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("Không thể load config:", err)
	}

	// 2. Kết nối Database
	db, err := utils.ConnectDB(cfg)
	if err != nil {
		log.Fatal("Kết nối DB thất bại:", err)
	}

	// 3. Khởi tạo các tầng (Dependency Injection)
	userRepo := postgres.NewUserRepository(db)
	authUsecase := usecase.NewAuthUsecase(userRepo, cfg)

	// 3.5 Khởi tạo MinIO client
	s3Client, err := utils.NewS3Client(cfg.MinioEndpoint, cfg.MinioAccessKey, cfg.MinioSecretKey, cfg.MinioBucket, cfg.MinioUseSSL)
	if err != nil {
		log.Fatal("Không thể khởi tạo S3 client:", err)
	}

	authHandler := v1.NewAuthHandler(authUsecase, s3Client)

	// 4. Khởi tạo Gin Router
	r := gin.Default()

	// 5. Định nghĩa Routes (Sử dụng Grouping cho Versioning)
	api := r.Group("/api")
	{
		v1Group := api.Group("/v1")
		{
			auth := v1Group.Group("/auth")
			{
				auth.POST("/register", authHandler.Register)
				auth.POST("/login", authHandler.Login)

			}

			protected := v1Group.Group("/users")
			protected.Use(middleware.AuthMiddleware(cfg.JWTSecret)) // <-- "Anh bảo vệ" ở đây
			{
				protected.POST("/avatar", authHandler.UploadAvatar) // API upload ảnh
			}
		}
	}

	// 6. Chạy Server
	log.Printf("Server đang chạy tại cổng: %s", cfg.AppPort)
	if err := r.Run(":" + cfg.AppPort); err != nil {
		log.Fatal("Lỗi khi chạy server:", err)
	}
}

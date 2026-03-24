package main

import (
	"log"

	"github.com/No2004LTC/gopher-social-ecom/config"
	"github.com/No2004LTC/gopher-social-ecom/internal/delivery/http/middleware"
	"github.com/No2004LTC/gopher-social-ecom/internal/delivery/http/v1"
	"github.com/No2004LTC/gopher-social-ecom/internal/delivery/ws" // <-- THÊM DÒNG NÀY
	"github.com/No2004LTC/gopher-social-ecom/internal/repository/postgres"
	"github.com/No2004LTC/gopher-social-ecom/internal/usecase"
	"github.com/No2004LTC/gopher-social-ecom/pkg/db"
	"github.com/No2004LTC/gopher-social-ecom/pkg/storage"
	"github.com/gin-gonic/gin"
)

func main() {
	// 1. Load Config & 2. Connect DB (Giữ nguyên)
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}
	db, err := db.ConnectDB(cfg)
	if err != nil {
		log.Fatal(err)
	}

	// 3. Khởi tạo các tầng (Dependency Injection)
	// --- PHẦN WEBSOCKET KHỞI TẠO Ở ĐÂY ---
	chatRepo := postgres.NewChatRepository(db)
	chatUC := usecase.NewChatUsecase(chatRepo)
	hub := ws.NewHub(chatUC)
	go hub.Run() // Chạy Hub trong goroutine riêng để nó luôn lắng nghe
	wsHandler := v1.NewWSHandler(hub, chatUC)
	// -------------------------------------

	userRepo := postgres.NewUserRepository(db)
	authUsecase := usecase.NewAuthUsecase(userRepo, cfg)

	s3Client, err := storage.NewS3Client(cfg.MinioEndpoint, cfg.MinioAccessKey, cfg.MinioSecretKey, cfg.MinioBucket, cfg.MinioUseSSL)
	if err != nil {
		log.Fatal(err)
	}

	authHandler := v1.NewAuthHandler(authUsecase, s3Client)

	// 4. Khởi tạo Gin Router
	r := gin.Default()

	// 5. Định nghĩa Routes
	api := r.Group("/api")
	{
		v1Group := api.Group("/v1")
		{
			// --- ROUTE WEBSOCKET ĐẶT Ở ĐÂY ---
			// Nó nằm trong v1Group để được hưởng prefix /api/v1/ws
			v1Group.GET("/ws", middleware.AuthMiddleware(cfg.JWTSecret), wsHandler.ServeWS)
			v1Group.GET("/chats/:to_user_id", middleware.AuthMiddleware(cfg.JWTSecret), wsHandler.GetHistory)
			// --------------------------------

			auth := v1Group.Group("/auth")
			{
				auth.POST("/register", authHandler.Register)
				auth.POST("/login", authHandler.Login)
			}

			// ... (Các group users, posts giữ nguyên như code của cậu) ...
			protected := v1Group.Group("/users")
			protected.Use(middleware.AuthMiddleware(cfg.JWTSecret))
			{
				protected.GET("/me", authHandler.GetMe)
				protected.POST("/avatar", authHandler.UploadAvatar)
			}

			postRepo := postgres.NewPostRepository(db)
			postUC := usecase.NewPostUsecase(postRepo, s3Client)
			postHandler := v1.NewPostHandler(postUC)

			interRepo := postgres.NewInteractionRepository(db)
			interUC := usecase.NewInteractionUsecase(interRepo, hub)
			interHandler := v1.NewInteractionHandler(interUC)

			postGroup := v1Group.Group("/posts")
			postGroup.Use(middleware.AuthMiddleware(cfg.JWTSecret))
			{
				postGroup.POST("", postHandler.Create)
				postGroup.GET("", postHandler.GetNewsfeed)
				postGroup.POST("/:id/like", interHandler.ToggleLike)
				postGroup.POST("/:id/comments", interHandler.AddComment)
				postGroup.GET("/:id/comments", interHandler.GetComments)
			}
		}
	}

	// 6. Chạy Server
	log.Printf("Server đang chạy tại cổng: %s", cfg.AppPort)
	r.Run(":" + cfg.AppPort)
}

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
	// 1. Load Config & 2. Connect DB
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}
	db, err := db.ConnectDB(cfg)
	if err != nil {
		log.Fatal(err)
	}

	// --- 3. KHỞI TẠO CÁC TẦNG (Dependency Injection) ---

	// A. WebSocket & Chat (Cần có Hub trước)
	chatRepo := postgres.NewChatRepository(db)
	chatUC := usecase.NewChatUsecase(chatRepo)
	hub := ws.NewHub(chatUC)
	go hub.Run()
	wsHandler := v1.NewWSHandler(hub, chatUC)

	// B. HỆ THỐNG THÔNG BÁO (NOTIFICATION) - Cực kỳ quan trọng
	notiRepo := postgres.NewNotificationRepository(db)
	notiUC := usecase.NewNotificationUsecase(notiRepo, hub) // Truyền hub vào đây để bắn realtime
	notiHandler := v1.NewNotificationHandler(notiUC)

	// C. Các Repo & Usecase khác (Bây giờ truyền notiUC thay vì truyền hub lẻ tẻ)
	followRepo := postgres.NewFollowRepository(db)
	// Thay vì truyền hub, ta truyền notiUC để Follow xong thì lưu thông báo vào DB luôn
	followUC := usecase.NewFollowUsecase(followRepo, notiUC)
	followHandler := v1.NewFollowHandler(followUC)

	userRepo := postgres.NewUserRepository(db)
	authUsecase := usecase.NewAuthUsecase(userRepo, cfg)

	s3Client, err := storage.NewS3Client(cfg.MinioEndpoint, cfg.MinioAccessKey, cfg.MinioSecretKey, cfg.MinioBucket, cfg.MinioUseSSL)
	if err != nil {
		log.Fatal(err)
	}
	authHandler := v1.NewAuthHandler(authUsecase, s3Client)

	postRepo := postgres.NewPostRepository(db)
	// PostUsecase giờ nhận thêm notiUC để báo "Có bài mới" cho follower
	postUC := usecase.NewPostUsecase(postRepo, followRepo, s3Client, notiUC)
	postHandler := v1.NewPostHandler(postUC)

	interRepo := postgres.NewInteractionRepository(db)
	// InteractionUsecase nhận notiUC để báo "Có người Like/Comment"
	interUC := usecase.NewInteractionUsecase(interRepo, notiUC)
	interHandler := v1.NewInteractionHandler(interUC)

	// --- 4. KHỞI TẠO ROUTER ---
	r := gin.Default()

	api := r.Group("/api")
	{
		v1Group := api.Group("/v1")
		{
			// WebSocket Routes
			v1Group.GET("/ws", middleware.AuthMiddleware(cfg.JWTSecret), wsHandler.ServeWS)
			v1Group.GET("/chats/:to_user_id", middleware.AuthMiddleware(cfg.JWTSecret), wsHandler.GetHistory)

			// Auth Routes
			auth := v1Group.Group("/auth")
			{
				auth.POST("/register", authHandler.Register)
				auth.POST("/login", authHandler.Login)
			}

			// User & Follow Routes
			users := v1Group.Group("/users")
			users.Use(middleware.AuthMiddleware(cfg.JWTSecret))
			{
				users.GET("/me", authHandler.GetMe)
				users.GET("/search", authHandler.SearchUsers)
				users.PATCH("/profile", authHandler.UpdateProfile)
				users.POST("/avatar", authHandler.UploadAvatar)

				users.POST("/:id/follow", followHandler.Follow)
				users.POST("/:id/unfollow", followHandler.Unfollow)
			}

			// Post & Interaction Routes
			posts := v1Group.Group("/posts")
			posts.Use(middleware.AuthMiddleware(cfg.JWTSecret))
			{
				posts.POST("", postHandler.Create)
				posts.GET("", postHandler.GetNewsfeed)
				posts.GET("/discovery", postHandler.GetDiscoveryFeed) // API Discovery mình vừa làm
				posts.POST("/:id/like", interHandler.ToggleLike)
				posts.POST("/:id/comments", interHandler.AddComment)
				posts.GET("/:id/comments", interHandler.GetComments)
			}

			// Notification Routes (Mới thêm)
			notifications := v1Group.Group("/notifications")
			notifications.Use(middleware.AuthMiddleware(cfg.JWTSecret))
			{
				notifications.GET("", notiHandler.GetNotifications)
				notifications.PATCH("/:id/read", notiHandler.MarkAsRead)
			}
		}
	}

	log.Printf("Server đang chạy tại cổng: %s", cfg.AppPort)
	r.Run(":" + cfg.AppPort)
}

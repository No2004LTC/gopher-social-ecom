package main

import (
	"log"
	"time"

	"github.com/No2004LTC/gopher-social-ecom/config"
	"github.com/No2004LTC/gopher-social-ecom/internal/delivery/http/middleware"
	"github.com/No2004LTC/gopher-social-ecom/internal/delivery/http/v1"
	"github.com/No2004LTC/gopher-social-ecom/internal/delivery/ws"
	"github.com/No2004LTC/gopher-social-ecom/internal/repository/postgres"
	"github.com/No2004LTC/gopher-social-ecom/internal/usecase"
	"github.com/No2004LTC/gopher-social-ecom/pkg/db"
	"github.com/No2004LTC/gopher-social-ecom/pkg/mail"
	"github.com/No2004LTC/gopher-social-ecom/pkg/storage"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func main() {
	// 1. Load Config & Connect DB
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	database, err := db.ConnectDB(cfg)
	if err != nil {
		log.Fatal(err)
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	// 2. --- KHỞI TẠO REPOSITORIES ---
	authRepo := postgres.NewAuthRepository(database)
	userRepo := postgres.NewUserRepository(database)
	followRepo := postgres.NewFollowRepository(database)
	postRepo := postgres.NewPostRepository(database)
	chatRepo := postgres.NewChatRepository(database)
	notiRepo := postgres.NewNotificationRepository(database)
	interRepo := postgres.NewInteractionRepository(database)
	bookmarkRepo := postgres.NewBookmarkRepository(database)

	// 3. --- KHỞI TẠO INFRASTRUCTURE ---
	emailSender := mail.NewGmailSender(cfg)
	s3Client, err := storage.NewS3Client(cfg.MinioEndpoint, cfg.MinioAccessKey, cfg.MinioSecretKey, cfg.MinioBucket, cfg.MinioUseSSL)
	if err != nil {
		log.Fatal(err)
	}

	// Hub cho WebSocket
	hub := ws.NewHub(redisClient)
	go hub.Run()

	// 4. --- KHỞI TẠO USECASES (WIRING) ---
	// Tách Auth và User rõ ràng
	authUC := usecase.NewAuthUsecase(authRepo, cfg, redisClient, emailSender)
	// UserUsecase nhận thêm FollowRepo và PostRepo để đếm số lượng Profile
	userUC := usecase.NewUserUsecase(userRepo, followRepo, postRepo, redisClient)

	chatUC := usecase.NewChatUsecase(chatRepo, userRepo)
	notiUC := usecase.NewNotificationUsecase(notiRepo, hub)
	followUC := usecase.NewFollowUsecase(followRepo, notiUC)
	postUC := usecase.NewPostUsecase(postRepo, s3Client, notiUC)
	interUC := usecase.NewInteractionUsecase(interRepo, notiUC)
	bookmarkUC := usecase.NewBookmarkUseCase(bookmarkRepo)

	// 5. --- KHỞI TẠO HANDLERS ---
	authHandler := v1.NewAuthHandler(authUC)
	userHandler := v1.NewUserHandler(userUC, s3Client)
	wsHandler := v1.NewWSHandler(hub, chatUC)
	notiHandler := v1.NewNotificationHandler(notiUC)
	followHandler := v1.NewFollowHandler(followUC)
	postHandler := v1.NewPostHandler(postUC)
	interHandler := v1.NewInteractionHandler(interUC)
	bookmarkHandler := v1.NewBookmarkHandler(bookmarkUC)

	// 6. --- ROUTER SETUP ---
	r := gin.Default()

	// CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowMethods:     []string{"POST", "GET", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	v1Group := r.Group("/api/v1")
	{
		// WebSocket (Realtime)
		v1Group.GET("/ws", middleware.AuthMiddleware(cfg.JWTSecret), wsHandler.ServeWS)

		// Auth Group (Public)
		auth := v1Group.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/send-otp", authHandler.SendPasswordOTP)
			auth.POST("/reset-password", authHandler.ResetPassword)
		}

		// User Group (Private - Need Token)
		users := v1Group.Group("/users")
		users.Use(middleware.AuthMiddleware(cfg.JWTSecret))
		{
			users.GET("/me", userHandler.GetMe)
			users.GET("/search", userHandler.SearchUsers)
			users.PATCH("/profile", userHandler.UpdateProfile)
			users.POST("/avatar", userHandler.UploadAvatar)
			users.POST("/cover", userHandler.UploadCover)
			users.GET("/profile/:username", userHandler.GetUserProfile) // 🚀 ĐÃ FIX: Gọi sang userHandler
			users.GET("/id/:id", userHandler.GetUserByID)
			users.GET("/following", userHandler.GetFollowing)
			users.GET("/followers", userHandler.GetFollowers)
			users.GET("/suggestions", userHandler.GetSuggestions)
			users.GET("/online-contacts", userHandler.GetOnlineFriends)

			// Follow Actions
			users.POST("/:id/follow", followHandler.Follow)
			users.POST("/:id/unfollow", followHandler.Unfollow)
		}

		// Chat Group
		chats := v1Group.Group("/chats")
		chats.Use(middleware.AuthMiddleware(cfg.JWTSecret))
		{
			chats.GET("/unread-count", wsHandler.GetUnreadCount)
			chats.GET("/conversations", wsHandler.GetConversations)
			chats.POST("", wsHandler.SendMessage)
			chats.GET("/:to_user_id", wsHandler.GetHistory)
			chats.PUT("/:id/read", wsHandler.MarkAsRead)
		}

		// Post Group
		posts := v1Group.Group("/posts")
		posts.Use(middleware.AuthMiddleware(cfg.JWTSecret))
		{
			posts.GET("/feed", postHandler.GetGlobalFeed)
			posts.GET("/user/:user_id", postHandler.GetUserPosts)
			posts.POST("", postHandler.Create)
			posts.PUT("/:id", postHandler.UpdatePost)
			posts.DELETE("/:id", postHandler.DeletePost)
			posts.POST("/:id/like", interHandler.ToggleLike)
			posts.POST("/:id/save", bookmarkHandler.ToggleSave)
			posts.GET("/:id/comments", interHandler.GetComments)
			posts.POST("/:id/comments", interHandler.AddComment)
		}

		// Notifications
		notifications := v1Group.Group("/notifications")
		notifications.Use(middleware.AuthMiddleware(cfg.JWTSecret))
		{
			notifications.GET("/unread-count", notiHandler.GetUnreadCount)
			notifications.GET("", notiHandler.GetNotifications)
			notifications.PUT("/read-all", notiHandler.MarkAllAsRead)
			notifications.PUT("/:id/read", notiHandler.MarkAsRead)
		}

		// Bookmarks
		v1Group.GET("/bookmarks", middleware.AuthMiddleware(cfg.JWTSecret), bookmarkHandler.GetSavedFeed)
	}

	log.Printf("🚀 ConnectVN Backend đang chạy tại cổng: %s", cfg.AppPort)
	r.Run(":" + cfg.AppPort)
}

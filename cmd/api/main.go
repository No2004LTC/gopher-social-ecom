package main

import (
	"log"
	"time"

	"github.com/No2004LTC/gopher-social-ecom/config"
	"github.com/No2004LTC/gopher-social-ecom/internal/delivery/http/middleware"
	v1 "github.com/No2004LTC/gopher-social-ecom/internal/delivery/http/v1"
	"github.com/No2004LTC/gopher-social-ecom/internal/delivery/ws"
	kafkaRepo "github.com/No2004LTC/gopher-social-ecom/internal/repository/kafka"
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
	// ==========================================
	// 1. KHỞI TẠO HẠ TẦNG (INFRASTRUCTURE)
	// ==========================================
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	database, err := db.ConnectDB(cfg)
	if err != nil {
		log.Fatal(err)
	}

	redisClient := redis.NewClient(&redis.Options{Addr: "localhost:6379"})

	emailSender := mail.NewGmailSender(cfg)
	s3Client, err := storage.NewS3Client(cfg.MinioEndpoint, cfg.MinioAccessKey, cfg.MinioSecretKey, cfg.MinioBucket, cfg.MinioUseSSL)
	if err != nil {
		log.Fatal(err)
	}

	hub := ws.NewHub(redisClient)
	go hub.Run()

	// ==========================================
	// 2. WIRING THEO TỪNG FEATURE (MODULE)
	// ==========================================

	// --- MODULE: NOTIFICATION  ---
	notiRepo := postgres.NewNotificationRepository(database)
	notiUC := usecase.NewNotificationUsecase(notiRepo, hub)
	notiHandler := v1.NewNotificationHandler(notiUC)

	// --- MODULE: AUTH ---
	authRepo := postgres.NewAuthRepository(database)
	authUC := usecase.NewAuthUsecase(authRepo, cfg, redisClient, emailSender)
	authHandler := v1.NewAuthHandler(authUC)

	// --- MODULE: FOLLOW ---
	followRepo := postgres.NewFollowRepository(database)
	followUC := usecase.NewFollowUsecase(followRepo, notiUC)
	followHandler := v1.NewFollowHandler(followUC)

	// --- MODULE: POST ---
	postRepo := postgres.NewPostRepository(database)
	postUC := usecase.NewPostUsecase(postRepo, s3Client, notiUC)
	postHandler := v1.NewPostHandler(postUC)

	// --- MODULE: USER ---
	userRepo := postgres.NewUserRepository(database)
	userUC := usecase.NewUserUsecase(userRepo, followRepo, postRepo, redisClient)
	userHandler := v1.NewUserHandler(userUC, s3Client)

	// --- MODULE: CHAT ---
	chatRepo := postgres.NewChatRepository(database)
	chatUC := usecase.NewChatUsecase(chatRepo, userRepo)
	wsHandler := v1.NewWSHandler(hub, chatUC)

	// --- MODULE: INTERACTION (MQ, Like, Comment) ---
	interRepo := postgres.NewInteractionRepository(database)
	interactionMQRepo := kafkaRepo.NewInteractionMQ(cfg.KafkaBroker)
	interUC := usecase.NewInteractionUsecase(interRepo, notiUC, interactionMQRepo)
	interHandler := v1.NewInteractionHandler(interUC)

	// --- MODULE: BOOKMARK ---
	bookmarkRepo := postgres.NewBookmarkRepository(database)
	bookmarkUC := usecase.NewBookmarkUseCase(bookmarkRepo)
	bookmarkHandler := v1.NewBookmarkHandler(bookmarkUC)

	// ==========================================
	// 3. THIẾT LẬP ROUTER & BẬT SERVER
	// ==========================================
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowMethods:     []string{"POST", "GET", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	v1Group := r.Group("/api/v1")
	{
		// WebSocket
		v1Group.GET("/ws", middleware.AuthMiddleware(cfg.JWTSecret), wsHandler.ServeWS)

		// Auth Group
		auth := v1Group.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/send-otp", authHandler.SendPasswordOTP)
			auth.POST("/reset-password", authHandler.ResetPassword)
		}

		// User Group
		users := v1Group.Group("/users")
		users.Use(middleware.AuthMiddleware(cfg.JWTSecret))
		{
			users.GET("/me", userHandler.GetMe)
			users.GET("/search", userHandler.SearchUsers)
			users.PATCH("/profile", userHandler.UpdateProfile)
			users.POST("/avatar", userHandler.UploadAvatar)
			users.POST("/cover", userHandler.UploadCover)
			users.GET("/profile/:username", userHandler.GetUserProfile)
			users.GET("/id/:id", userHandler.GetUserByID)
			users.GET("/following", userHandler.GetFollowing)
			users.GET("/followers", userHandler.GetFollowers)
			users.GET("/suggestions", userHandler.GetSuggestions)
			users.GET("/online-contacts", userHandler.GetOnlineFriends)

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

		// Post & Interaction Group
		posts := v1Group.Group("/posts")
		posts.Use(middleware.AuthMiddleware(cfg.JWTSecret))
		{
			posts.GET("/feed", postHandler.GetGlobalFeed)
			posts.GET("/user/:user_id", postHandler.GetUserPosts)
			posts.POST("", postHandler.Create)
			posts.PUT("/:id", postHandler.UpdatePost)
			posts.DELETE("/:id", postHandler.DeletePost)

			posts.POST("/:id/like", interHandler.ToggleLike)
			posts.GET("/:id/comments", interHandler.GetComments)
			posts.POST("/:id/comments", interHandler.AddComment)

			posts.POST("/:id/save", bookmarkHandler.ToggleSave)
		}

		// Notifications Group
		notifications := v1Group.Group("/notifications")
		notifications.Use(middleware.AuthMiddleware(cfg.JWTSecret))
		{
			notifications.GET("/unread-count", notiHandler.GetUnreadCount)
			notifications.GET("", notiHandler.GetNotifications)
			notifications.PUT("/read-all", notiHandler.MarkAllAsRead)
			notifications.PUT("/:id/read", notiHandler.MarkAsRead)
		}

		// Bookmarks Group
		v1Group.GET("/bookmarks", middleware.AuthMiddleware(cfg.JWTSecret), bookmarkHandler.GetSavedFeed)
	}

	log.Printf("🚀 ConnectVN Backend đang chạy tại cổng: %s", cfg.AppPort)
	r.Run(":" + cfg.AppPort)
}

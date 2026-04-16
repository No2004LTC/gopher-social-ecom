package main

import (
	"context"
	"log"

	"github.com/No2004LTC/gopher-social-ecom/config"
	"github.com/No2004LTC/gopher-social-ecom/internal/delivery/http/middleware"
	"github.com/No2004LTC/gopher-social-ecom/internal/delivery/http/v1"
	"github.com/No2004LTC/gopher-social-ecom/internal/delivery/ws" // <-- THÊM DÒNG NÀY
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
	// 1. Load Config & 2. Connect DB
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}
	db, err := db.ConnectDB(cfg)
	if err != nil {
		log.Fatal(err)
	}
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379", // Địa chỉ mặc định của Redis (Sửa lại nếu cậu dùng port khác)
		Password: "",               // Mật khẩu (để trống nếu cài mặc định)
		DB:       0,                // Dùng Database số 0 mặc định
	})
	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("🔴 Lỗi kết nối Redis: %v.", err)
	}
	log.Println("🟢 Kết nối Redis thành công!")

	// --- 3. KHỞI TẠO CÁC TẦNG (Dependency Injection) ---

	// A. Hệ thống WebSocket & Chat
	chatRepo := postgres.NewChatRepository(db)
	chatUC := usecase.NewChatUsecase(chatRepo)

	// Hub nhận redisClient để ghi nhận trạng thái Online/Offline vào Redis
	hub := ws.NewHub(redisClient)
	go hub.Run()
	wsHandler := v1.NewWSHandler(hub, chatUC)

	// B. Hệ thống Thông báo
	notiRepo := postgres.NewNotificationRepository(db)
	notiUC := usecase.NewNotificationUsecase(notiRepo, hub)
	notiHandler := v1.NewNotificationHandler(notiUC)

	// C. Hệ thống User & Auth (QUAN TRỌNG NHẤT Ở ĐÂY)
	userRepo := postgres.NewUserRepository(db)
	emailSender := mail.NewGmailSender(cfg)
	// 👉 CẬP NHẬT: Đảm bảo NewAuthUsecase nhận ĐỦ 3 tham số: repo, config, và redisClient
	// Nếu NewAuthUsecase của cậu chưa nhận redisClient, hãy vào file đó thêm vào struct nhé!
	authUsecase := usecase.NewAuthUsecase(userRepo, cfg, redisClient, emailSender)

	s3Client, err := storage.NewS3Client(cfg.MinioEndpoint, cfg.MinioAccessKey, cfg.MinioSecretKey, cfg.MinioBucket, cfg.MinioUseSSL)
	if err != nil {
		log.Fatal(err)
	}
	// AuthHandler quản lý GetOnlineFriends -> Nó gọi authUsecase đã có Redis
	authHandler := v1.NewAuthHandler(authUsecase, s3Client)

	// D. Các hệ thống khác
	followRepo := postgres.NewFollowRepository(db)
	followUC := usecase.NewFollowUsecase(followRepo, notiUC)
	followHandler := v1.NewFollowHandler(followUC)

	postRepo := postgres.NewPostRepository(db)
	postUC := usecase.NewPostUsecase(postRepo, s3Client, notiUC)
	postHandler := v1.NewPostHandler(postUC)

	interRepo := postgres.NewInteractionRepository(db)
	interUC := usecase.NewInteractionUsecase(interRepo, notiUC)
	interHandler := v1.NewInteractionHandler(interUC)

	bookmarkRepo := postgres.NewBookmarkRepository(db)
	bookmarkUC := usecase.NewBookmarkUseCase(bookmarkRepo)
	bookmarkHandler := v1.NewBookmarkHandler(bookmarkUC)

	// --- 4. KHỞI TẠO ROUTER ---
	r := gin.Default()

	// CẤU HÌNH CORS ĐÃ ĐƯỢC FIX LỖI "Failed to fetch"
	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{"http://localhost:5173"}, // URL của React
		// 👉 ĐÃ BỔ SUNG "PATCH" VÀ "DELETE" VÀO DÒNG DƯỚI ĐÂY:
		AllowMethods:     []string{"POST", "GET", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	api := r.Group("/api")
	{
		v1Group := api.Group("/v1")
		{
			// WebSocket Routes
			v1Group.GET("/ws", middleware.AuthMiddleware(cfg.JWTSecret), wsHandler.ServeWS)
			chatGroup := v1Group.Group("/chats")
			chatGroup.Use(middleware.AuthMiddleware(cfg.JWTSecret)) // 👉 Áp dụng Auth 1 lần cho cả Group!
			{
				chatGroup.POST("", wsHandler.SendMessage)           // Tuyến: POST /api/v1/chats
				chatGroup.GET("/:to_user_id", wsHandler.GetHistory) // Tuyến: GET /api/v1/chats/:to_user_id
			}
			// Auth Routes
			auth := v1Group.Group("/auth")
			{
				auth.POST("/register", authHandler.Register)
				auth.POST("/login", authHandler.Login)
				auth.POST("/send-otp", authHandler.SendPasswordOTP)
				auth.POST("/reset-password", authHandler.ResetPassword)
			}

			// User & Follow Routes
			users := v1Group.Group("/users")
			users.Use(middleware.AuthMiddleware(cfg.JWTSecret))
			{
				users.GET("/me", authHandler.GetMe)
				users.GET("/search", authHandler.SearchUsers)
				users.PATCH("/profile", authHandler.UpdateProfile)
				users.POST("/avatar", authHandler.UploadAvatar)
				users.POST("/cover", authHandler.UploadCover)

				users.GET("/following", authHandler.GetFollowing)
				users.GET("/followers", authHandler.GetFollowers)
				users.GET("/profile/:username", authHandler.GetUserProfile)

				users.GET("/suggestions", authHandler.GetSuggestions)
				users.GET("/online-contacts", authHandler.GetOnlineFriends)

				users.POST("/:id/follow", followHandler.Follow)
				users.POST("/:id/unfollow", followHandler.Unfollow)
			}

			// Post & Interaction Routes
			posts := v1Group.Group("/posts")
			posts.Use(middleware.AuthMiddleware(cfg.JWTSecret))
			{
				// --- 1. LẤY DỮ LIỆU (READ) ---
				// Trang chủ: Hiện tất cả bài viết của mọi người
				posts.GET("/feed", postHandler.GetGlobalFeed)

				// Trang cá nhân: Hiện bài viết của 1 User cụ thể (của mình hoặc người khác)
				posts.GET("/user/:user_id", postHandler.GetUserPosts)

				// --- 2. THAO TÁC BÀI VIẾT (C.U.D) ---
				posts.POST("", postHandler.Create)
				posts.PUT("/:id", postHandler.UpdatePost)
				posts.DELETE("/:id", postHandler.DeletePost)

				// --- 3. TÀI NGUYÊN CON (Comments, Likes, Bookmarks) ---

				// Likes & Saves
				posts.POST("/:id/like", interHandler.ToggleLike)
				posts.POST("/:id/save", bookmarkHandler.ToggleSave)

				// Comments (Chuẩn REST)
				posts.GET("/:id/comments", interHandler.GetComments) // Lấy list cmt của post
				posts.POST("/:id/comments", interHandler.AddComment) // Thêm cmt vào post

				// Sửa/Xóa cmt: Dùng comment_id để định danh chính xác
				posts.PUT("/:id/comments/:comment_id", interHandler.UpdateComment)
				posts.DELETE("/:id/comments/:comment_id", interHandler.DeleteComment)
			}

			bookmarks := v1Group.Group("/bookmarks")
			bookmarks.Use(middleware.AuthMiddleware(cfg.JWTSecret))
			{
				// 👉 API XEM DANH SÁCH ĐÃ LƯU
				bookmarks.GET("", bookmarkHandler.GetSavedFeed)
			}

			// Notification Routes (Mới thêm)
			notifications := v1Group.Group("/notifications")
			notifications.Use(middleware.AuthMiddleware(cfg.JWTSecret))
			{
				notifications.GET("", notiHandler.GetNotifications)
				notifications.PUT("/:id/read", notiHandler.MarkAsRead)
			}
		}
	}

	log.Printf("Server đang chạy tại cổng: %s", cfg.AppPort)
	r.Run(":" + cfg.AppPort)
}

package v1

import (
	"fmt"
	"log"
	"net/http"

	"github.com/No2004LTC/gopher-social-ecom/internal/domain"
	"github.com/No2004LTC/gopher-social-ecom/pkg/storage"
	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v7"
)

// Goi den inteface UserUsecase de thuc hien cac logic lien quan den authentication
type AuthHandler struct {
	authUsecase domain.UserUsecase
	// s3Client là wrapper cho MinIO client để upload file
	s3Client *storage.S3Client
}

// ham khoi tao
func NewAuthHandler(u domain.UserUsecase, s3 *storage.S3Client) *AuthHandler {
	return &AuthHandler{authUsecase: u, s3Client: s3}
}

// / registerRequest: Đây chính là DTO (Data Transfer Object).
// Trong Go, ta định nghĩa Struct này ngay tại tầng Handler để hứng dữ liệu từ Client.
type registerRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// Xu ly request dang ky tu client, goi den usecase de thuc hien logic va tra ve response
func (h *AuthHandler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.authUsecase.Register(c.Request.Context(), req.Username, req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Đăng ký thành công"})
}

// Struct cho yêu cầu đăng nhập
type loginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// Phương thức Login
func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dữ liệu không hợp lệ"})
		return
	}

	// Gọi Usecase để thực hiện đăng nhập
	token, err := h.authUsecase.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// Trả về token cho người dùng
	c.JSON(http.StatusOK, gin.H{
		"access_token": token,
		"token_type":   "Bearer",
	})
}

func (h *AuthHandler) UploadAvatar(c *gin.Context) {
	// 1. Lấy userID từ Middleware (Chứng minh đã login)
	uid, exists := c.Get("user_id")
	log.Printf("[UploadAvatar] user_id from context: %v (type: %T), exists: %v\n", uid, uid, exists)

	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var userID int64
	switch v := uid.(type) {
	case int64:
		userID = v
	case uint32:
		userID = int64(v)
		log.Printf("[UploadAvatar] Converted uint32 to int64: %d\n", userID)
	case uint64:
		userID = int64(v)
		log.Printf("[UploadAvatar] Converted uint64 to int64: %d\n", userID)
	case float64:
		userID = int64(v)
		log.Printf("[UploadAvatar] Converted float64 to int64: %d\n", userID)
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID type"})
		return
	}

	log.Printf("[UploadAvatar] Final userID: %d\n", userID)

	// 2. Nhận file từ form-data (key là "avatar")
	file, err := c.FormFile("avatar")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File not found"})
		return
	}

	// 3. Mở file và đẩy lên MinIO
	f, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Cannot open file: " + err.Error()})
		return
	}
	defer func() { _ = f.Close() }()

	objectName := "avatars/" + file.Filename
	_, err = h.s3Client.PutObject(c.Request.Context(), objectName, f, file.Size, minio.PutObjectOptions{
		ContentType: file.Header.Get("Content-Type"),
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Upload failed: " + err.Error()})
		return
	}

	// 4. Tạo URL công khai
	url := fmt.Sprintf("http://%s/%s/%s", h.s3Client.Endpoint(), h.s3Client.Bucket(), objectName)

	// 5. Lưu vào DB thông qua usecase
	log.Printf("[UploadAvatar] Calling UpdateAvatar with userID: %d, url: %s\n", userID, url)
	if err := h.authUsecase.UpdateAvatar(c.Request.Context(), userID, url); err != nil {
		log.Printf("[UploadAvatar] UpdateAvatar error: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database update failed: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Avatar updated successfully",
		"url":     url,
		"user_id": userID,
	})
}

// [GET] /api/v1/users/me -> Xem profile
func (h *AuthHandler) GetMe(c *gin.Context) {
	// Lấy userID từ Middleware (phải khớp kiểu dữ liệu int64)
	uid, _ := c.Get("user_id")
	userID := uid.(int64)

	user, err := h.authUsecase.GetProfile(c.Request.Context(), userID)
	if err != nil {
		c.JSON(404, gin.H{"error": "Không tìm thấy user"})
		return
	}
	// Xóa password trước khi trả về
	user.PasswordHash = ""
	c.JSON(200, user)
}

// [PATCH] /api/v1/users/profile -> Cập nhật tên
func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	var input struct {
		Username string `json:"username" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": "Dữ liệu không hợp lệ"})
		return
	}

	uid, _ := c.Get("user_id")
	userID := uid.(int64)

	if err := h.authUsecase.UpdateProfile(c.Request.Context(), userID, input.Username); err != nil {
		c.JSON(500, gin.H{"error": "Cập nhật thất bại"})
		return
	}
	c.JSON(200, gin.H{"message": "Cập nhật thành công"})
}

func (h *AuthHandler) SearchUsers(c *gin.Context) {
	query := c.Query("q")
	currentUserID := c.MustGet("user_id").(int64)

	// Ở đây ta tạm thời fix cứng limit=10, offset=0
	// Sau này bạn có thể lấy từ c.Query("limit") nếu muốn
	users, err := h.authUsecase.SearchUsers(c.Request.Context(), currentUserID, query, 10, 0)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, users)
}

package v1

import (
	"fmt"
	"net/http"

	"github.com/No2004LTC/gopher-social-ecom/internal/domain"
	"github.com/No2004LTC/gopher-social-ecom/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v7"
)

// Goi den inteface UserUsecase de thuc hien cac logic lien quan den authentication
type AuthHandler struct {
	authUsecase domain.UserUsecase
	// s3Client là wrapper cho MinIO client để upload file
	s3Client *utils.S3Client
}

// ham khoi tao
func NewAuthHandler(u domain.UserUsecase, s3 *utils.S3Client) *AuthHandler {
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
	userID, _ := c.Get("user_id")

	// 2. Nhận file từ form-data (key là "avatar")
	file, err := c.FormFile("avatar")
	if err != nil {
		c.JSON(400, gin.H{"error": "Không tìm thấy file"})
		return
	}

	// 3. Mở file và đẩy lên MinIO
	f, err := file.Open()
	if err != nil {
		c.JSON(500, gin.H{"error": "Không thể mở file: " + err.Error()})
		return
	}
	defer func() { _ = f.Close() }()

	objectName := "avatars/" + file.Filename
	_, err = h.s3Client.PutObject(c.Request.Context(), objectName, f, file.Size, minio.PutObjectOptions{
		ContentType: file.Header.Get("Content-Type"),
	})

	if err != nil {
		c.JSON(500, gin.H{"error": "Lỗi upload: " + err.Error()})
		return
	}

	// 4. Trả về URL (Vì đã set public nên link này sẽ xem được luôn)
	url := fmt.Sprintf("http://%s/%s/%s", h.s3Client.Endpoint(), h.s3Client.Bucket(), objectName)
	c.JSON(200, gin.H{
		"message": "Upload thành công",
		"url":     url,
		"user_id": userID,
	})
}

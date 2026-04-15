package v1

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/No2004LTC/gopher-social-ecom/internal/delivery/http/response"
	"github.com/No2004LTC/gopher-social-ecom/internal/domain"
	"github.com/No2004LTC/gopher-social-ecom/internal/dto" // Đã thêm import DTO
	"github.com/No2004LTC/gopher-social-ecom/pkg/storage"
	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v7"
)

// AuthHandler xử lý các request liên quan đến User và Authentication
type AuthHandler struct {
	authUsecase domain.UserUsecase
	s3Client    *storage.S3Client
}

// NewAuthHandler khởi tạo AuthHandler
func NewAuthHandler(u domain.UserUsecase, s3 *storage.S3Client) *AuthHandler {
	return &AuthHandler{authUsecase: u, s3Client: s3}
}

// [POST] /api/v1/auth/register -> Đăng ký
func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest // Đã dùng DTO
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Dữ liệu không hợp lệ: "+err.Error())
		return
	}

	err := h.authUsecase.Register(c.Request.Context(), req.Username, req.Email, req.Password)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.SuccessMessage(c, "Đăng ký thành công")
}

// [POST] /api/v1/auth/login -> Đăng nhập
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Dữ liệu email không hợp lệ")
		return
	}

	// Gọi Usecase với req.Email
	token, user, err := h.authUsecase.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, err.Error())
		return
	}

	res := dto.LoginResponse{
		AccessToken: token,
		TokenType:   "Bearer",
		User: dto.AuthUserResponse{
			ID:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			AvatarURL: user.AvatarURL,
		},
	}

	c.JSON(http.StatusOK, res)
}

// [POST] /api/v1/users/avatar -> Upload Avatar
func (h *AuthHandler) UploadAvatar(c *gin.Context) {
	uid, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	var userID int64
	switch v := uid.(type) {
	case int64:
		userID = v
	case uint32:
		userID = int64(v)
	case uint64:
		userID = int64(v)
	case float64:
		userID = int64(v)
	default:
		response.Error(c, http.StatusInternalServerError, "Invalid user ID type")
		return
	}

	file, err := c.FormFile("avatar")
	if err != nil {
		response.Error(c, http.StatusBadRequest, "File not found")
		return
	}

	f, err := file.Open()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Cannot open file: "+err.Error())
		return
	}
	defer func() { _ = f.Close() }()

	objectName := "avatars/" + file.Filename
	_, err = h.s3Client.PutObject(c.Request.Context(), objectName, f, file.Size, minio.PutObjectOptions{
		ContentType: file.Header.Get("Content-Type"),
	})

	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Upload failed: "+err.Error())
		return
	}

	url := fmt.Sprintf("http://%s/%s/%s", h.s3Client.Endpoint(), h.s3Client.Bucket(), objectName)

	if err := h.authUsecase.UpdateAvatar(c.Request.Context(), userID, url); err != nil {
		response.Error(c, http.StatusInternalServerError, "Database update failed: "+err.Error())
		return
	}

	// Bọc kết quả vào Success Helper
	response.Success(c, "Avatar updated successfully", gin.H{
		"url":     url,
		"user_id": userID,
	})
}

// [GET] /api/v1/users/me -> Xem profile
func (h *AuthHandler) GetMe(c *gin.Context) {
	uid, _ := c.Get("user_id")
	userID := uid.(int64)

	user, err := h.authUsecase.GetProfile(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, http.StatusNotFound, "Không tìm thấy user")
		return
	}

	user.PasswordHash = ""
	response.Success(c, "Lấy thông tin thành công", user)
}

// [PATCH] /api/v1/users/profile -> Cập nhật tên
func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	var input dto.UpdateProfileRequest // Đã dùng DTO
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, http.StatusBadRequest, "Dữ liệu không hợp lệ")
		return
	}

	uid, _ := c.Get("user_id")
	userID := uid.(int64)

	if err := h.authUsecase.UpdateProfile(c.Request.Context(), userID, input.Username); err != nil {
		response.Error(c, http.StatusInternalServerError, "Cập nhật thất bại")
		return
	}

	response.SuccessMessage(c, "Cập nhật thành công")
}

// [GET] /api/v1/users/search -> Tìm kiếm người dùng
func (h *AuthHandler) SearchUsers(c *gin.Context) {
	query := c.Query("q")

	if query == "" {
		response.Success(c, "Danh sách rỗng", []interface{}{})
		return
	}

	currentUserID := c.MustGet("user_id").(int64)

	users, err := h.authUsecase.SearchUsers(c.Request.Context(), currentUserID, query, 10, 0)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, "Tìm kiếm thành công", users)
}

// [GET] /api/v1/users/following -> Lấy danh sách đang theo dõi
func (h *AuthHandler) GetFollowing(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Không tìm thấy thông tin xác thực")
		return
	}
	currentUserID := userID.(int64)

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	users, err := h.authUsecase.GetFollowing(c.Request.Context(), currentUserID, limit, offset)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Lỗi khi tải danh sách Đang theo dõi")
		return
	}

	response.Success(c, "Lấy danh sách thành công", users)
}

// [GET] /api/v1/users/followers -> Lấy danh sách người theo dõi mình
func (h *AuthHandler) GetFollowers(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Không tìm thấy thông tin xác thực")
		return
	}
	currentUserID := userID.(int64)

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	users, err := h.authUsecase.GetFollowers(c.Request.Context(), currentUserID, limit, offset)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Lỗi khi tải danh sách Người theo dõi")
		return
	}

	response.Success(c, "Lấy danh sách thành công", users)
}

func (h *AuthHandler) GetSuggestions(c *gin.Context) {
	// A. Lấy userID từ Token (Giả định AuthMiddleware của cậu set biến này là "user_id")
	// Chú ý: Cậu check lại AuthMiddleware của cậu xem key lưu ID là "user_id" hay gì nhé
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// B. Gọi Usecase (Truyền c.Request.Context() để đồng bộ timeout/cancel)
	suggestions, err := h.authUsecase.GetFriendSuggestions(c.Request.Context(), userID.(int64))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể lấy danh sách gợi ý"})
		return
	}

	// C. Map dữ liệu từ Domain Model sang Response DTO
	var res []dto.SuggestedUserResponse
	for _, s := range suggestions {
		res = append(res, dto.SuggestedUserResponse{
			ID:                 s.ID,
			Username:           s.Username,
			AvatarURL:          s.AvatarURL,
			MutualFriendsCount: s.MutualFriendsCount,
		})
	}

	// Xử lý mảng rỗng: Để React nhận được [] thay vì null khi không có gợi ý nào
	if res == nil {
		res = []dto.SuggestedUserResponse{}
	}

	// D. Trả về cho Frontend
	c.JSON(http.StatusOK, gin.H{
		"message": "Thành công",
		"data":    res,
	})
}

// GetOnlineFriends lấy danh sách bạn bè đang trực tuyến (quét qua Redis)
func (h *AuthHandler) GetOnlineFriends(c *gin.Context) {
	// 1. Lấy ID của người dùng từ Token (được AuthMiddleware ném vào Context)
	// Lưu ý: Cậu kiểm tra lại xem middleware của cậu dùng key là "user_id" hay "userID" nhé
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Không tìm thấy thông tin xác thực"})
		return
	}

	// 2. Gọi xuống tầng Usecase
	// Chú ý: Ở bài trước chúng ta đặt tên hàm trong Usecase là GetOnlineContacts
	onlineContacts, err := h.authUsecase.GetOnlineContacts(c.Request.Context(), userID.(int64))
	if err != nil {
		// Log lỗi ra console để dễ debug nếu cần
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể lấy danh sách bạn bè online"})
		return
	}

	// 3. Trả về cho Frontend
	c.JSON(http.StatusOK, gin.H{
		"message": "Lấy danh sách thành công",
		"data":    onlineContacts, // Cục này chính là mảng []dto.UserCompact đã gắn is_online = true
	})
}

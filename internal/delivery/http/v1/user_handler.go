package v1

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/No2004LTC/gopher-social-ecom/internal/delivery/http/response"
	"github.com/No2004LTC/gopher-social-ecom/internal/domain"
	"github.com/No2004LTC/gopher-social-ecom/internal/dto"
	"github.com/No2004LTC/gopher-social-ecom/pkg/storage"
	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v7"
)

type UserHandler struct {
	userUsecase domain.UserUsecase
	s3Client    *storage.S3Client
}

func NewUserHandler(u domain.UserUsecase, s3 *storage.S3Client) *UserHandler {
	return &UserHandler{userUsecase: u, s3Client: s3}
}

// GetMe
func (h *UserHandler) GetMe(c *gin.Context) {
	uid, _ := c.Get("user_id")
	userID := uid.(int64)

	user, err := h.userUsecase.GetProfile(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, http.StatusNotFound, "Không tìm thấy user")
		return
	}
	user.PasswordHash = ""
	response.Success(c, "Lấy thông tin thành công", user)
}

// SearchUsers
func (h *UserHandler) SearchUsers(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		response.Success(c, "Danh sách rỗng", []interface{}{})
		return
	}
	currentUserID := c.MustGet("user_id").(int64)
	users, err := h.userUsecase.SearchUsers(c.Request.Context(), currentUserID, query, 10, 0)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, "Tìm kiếm thành công", users)
}

// UpdateProfile
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID := c.MustGet("user_id").(int64)
	var input dto.UpdateProfileInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, http.StatusBadRequest, "Dữ liệu không hợp lệ")
		return
	}
	if err := h.userUsecase.UpdateProfile(c.Request.Context(), userID, input); err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, "Cập nhật thành công", nil)
}

// 4. Upload Avatar
func (h *UserHandler) UploadAvatar(c *gin.Context) {
	userID := c.MustGet("user_id").(int64)
	file, err := c.FormFile("avatar")
	if err != nil {
		response.Error(c, http.StatusBadRequest, "File not found")
		return
	}
	f, _ := file.Open()
	defer f.Close()

	objectName := "avatars/" + file.Filename
	h.s3Client.PutObject(c.Request.Context(), objectName, f, file.Size, minio.PutObjectOptions{
		ContentType: file.Header.Get("Content-Type"),
	})

	url := fmt.Sprintf("http://%s/%s/%s", h.s3Client.Endpoint(), h.s3Client.Bucket(), objectName)
	h.userUsecase.UpdateAvatar(c.Request.Context(), userID, url)
	response.Success(c, "Avatar updated", gin.H{"url": url})
}

// 5. Upload Cover
func (h *UserHandler) UploadCover(c *gin.Context) {
	userID := c.MustGet("user_id").(int64)
	file, err := c.FormFile("cover")
	if err != nil {
		response.Error(c, http.StatusBadRequest, "File not found")
		return
	}
	f, _ := file.Open()
	defer f.Close()

	objectName := "covers/" + file.Filename
	h.s3Client.PutObject(c.Request.Context(), objectName, f, file.Size, minio.PutObjectOptions{
		ContentType: file.Header.Get("Content-Type"),
	})

	url := fmt.Sprintf("http://%s/%s/%s", h.s3Client.Endpoint(), h.s3Client.Bucket(), objectName)
	h.userUsecase.UpdateCover(c.Request.Context(), userID, url)
	response.Success(c, "Cover updated", gin.H{"url": url})
}

// GetUserProfile
func (h *UserHandler) GetUserProfile(c *gin.Context) {
	uid, _ := c.Get("user_id")
	currentUserID := uid.(int64)
	targetUsername := c.Param("username")

	userProfile, err := h.userUsecase.GetUserProfileByUsername(c.Request.Context(), currentUserID, targetUsername)
	if err != nil {
		response.Error(c, http.StatusNotFound, "Không tìm thấy người dùng")
		return
	}
	response.Success(c, "Thành công", userProfile)
}

// Get by ID
func (h *UserHandler) GetUserByID(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.ParseInt(idStr, 10, 64)
	user, err := h.userUsecase.GetProfile(c.Request.Context(), id)
	if err != nil {
		response.Error(c, http.StatusNotFound, "User not found")
		return
	}
	user.PasswordHash = ""
	response.Success(c, "Thành công", user)
}

// GetFollowing
func (h *UserHandler) GetFollowing(c *gin.Context) {
	userID := c.MustGet("user_id").(int64)
	users, err := h.userUsecase.GetFollowing(c.Request.Context(), userID, 20, 0)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, "Thành công", users)
}

// GetFollowers
func (h *UserHandler) GetFollowers(c *gin.Context) {
	userID := c.MustGet("user_id").(int64)
	users, err := h.userUsecase.GetFollowers(c.Request.Context(), userID, 20, 0)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, "Thành công", users)
}

// Suggestions
func (h *UserHandler) GetSuggestions(c *gin.Context) {
	userID := c.MustGet("user_id").(int64)
	suggestions, err := h.userUsecase.GetFriendSuggestions(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, "Thành công", suggestions)
}

// Online Contacts
func (h *UserHandler) GetOnlineFriends(c *gin.Context) {
	userID := c.MustGet("user_id").(int64)
	onlineContacts, err := h.userUsecase.GetOnlineContacts(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, "Thành công", onlineContacts)
}

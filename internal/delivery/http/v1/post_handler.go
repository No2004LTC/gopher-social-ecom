package v1

import (
	"net/http"
	"strconv"

	"github.com/No2004LTC/gopher-social-ecom/internal/domain"
	"github.com/No2004LTC/gopher-social-ecom/internal/dto"
	"github.com/gin-gonic/gin"
)

type PostHandler struct {
	postUC domain.PostUsecase
}

func NewPostHandler(postUC domain.PostUsecase) *PostHandler {
	return &PostHandler{postUC: postUC}
}

// 1. TẠO BÀI VIẾT
func (h *PostHandler) Create(c *gin.Context) {
	userID := c.MustGet("user_id").(int64)
	content := c.PostForm("content")
	file, _ := c.FormFile("image")

	post := &domain.Post{
		UserID:  userID,
		Content: content,
	}

	if err := h.postUC.CreatePost(c.Request.Context(), post, file); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, post)
}

// 2. SỬA BÀI VIẾT
func (h *PostHandler) UpdatePost(c *gin.Context) {
	postID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	userID := c.MustGet("user_id").(int64)

	var req struct {
		Content string `json:"content" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dữ liệu không hợp lệ"})
		return
	}

	if err := h.postUC.UpdatePost(c.Request.Context(), postID, userID, req.Content); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Cập nhật thành công"})
}

// 3. XÓA BÀI VIẾT
func (h *PostHandler) DeletePost(c *gin.Context) {
	postID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	userID := c.MustGet("user_id").(int64)

	if err := h.postUC.DeletePost(c.Request.Context(), postID, userID); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Đã xóa bài viết"})
}

// 4. LẤY BẢNG TIN CHUNG (GLOBAL FEED)
func (h *PostHandler) GetGlobalFeed(c *gin.Context) {
	currentUserID := c.MustGet("user_id").(int64)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	posts, err := h.postUC.GetPosts(c.Request.Context(), currentUserID, 0, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.sendPostResponse(c, posts)
}

// 5. LẤY BÀI VIẾT CỦA USER CỤ THỂ (PROFILE)
func (h *PostHandler) GetUserPosts(c *gin.Context) {
	currentUserID := c.MustGet("user_id").(int64)
	targetUserID, _ := strconv.ParseInt(c.Param("user_id"), 10, 64)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	posts, err := h.postUC.GetPosts(c.Request.Context(), currentUserID, targetUserID, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.sendPostResponse(c, posts)
}

// Helper function để mapping từ Domain sang DTO cho đỡ viết lặp code
func (h *PostHandler) sendPostResponse(c *gin.Context, posts []domain.Post) {
	response := make([]dto.PostResponse, 0)
	for _, p := range posts {
		authorData := dto.ActorCompact{
			ID:        0,
			Username:  "Unknown",
			AvatarURL: "",
		}
		if p.User != nil {
			authorData.ID = p.User.ID
			authorData.Username = p.User.Username
			authorData.AvatarURL = p.User.AvatarURL
		}

		response = append(response, dto.PostResponse{
			ID:            p.ID,
			Content:       p.Content,
			ImageURL:      p.ImageURL,
			LikesCount:    p.LikesCount,
			CommentsCount: p.CommentsCount,
			IsLiked:       p.IsLiked,
			IsSaved:       p.IsSaved,
			CreatedAt:     p.CreatedAt,
			Author:        authorData,
		})
	}
	c.JSON(http.StatusOK, gin.H{"data": response})
}

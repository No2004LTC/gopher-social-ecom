package v1

import (
	"net/http"
	"strconv"

	"github.com/No2004LTC/gopher-social-ecom/internal/domain"
	"github.com/gin-gonic/gin"
)

type PostHandler struct {
	postUC domain.PostUsecase
}

func NewPostHandler(postUC domain.PostUsecase) *PostHandler {
	return &PostHandler{postUC: postUC}
}

func (h *PostHandler) Create(c *gin.Context) {
	// Lấy userID từ Middleware (đã check token lỏ hay không ở cửa trước)
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

func (h *PostHandler) GetNewsfeed(c *gin.Context) {
	// 1. Lấy userID từ token (Middleware đã đặt vào context)
	userID := c.MustGet("user_id").(int64)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	// 2. Truyền userID vào Usecase
	posts, err := h.postUC.GetFeed(c.Request.Context(), page, limit, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, posts)
}

package v1

import (
	"net/http"
	"strconv"

	"github.com/No2004LTC/gopher-social-ecom/internal/domain"
	"github.com/gin-gonic/gin"
)

type BookmarkHandler struct {
	bookmarkUC domain.BookmarkUseCase
}

func NewBookmarkHandler(buc domain.BookmarkUseCase) *BookmarkHandler {
	return &BookmarkHandler{bookmarkUC: buc}
}

// ToggleSave
func (h *BookmarkHandler) ToggleSave(c *gin.Context) {
	// Lấy post_id từ URL (/posts/:id/save)
	postID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID bài viết không hợp lệ"})
		return
	}

	userID := c.GetInt64("user_id")

	isSaved, err := h.bookmarkUC.ToggleSavePost(c.Request.Context(), userID, postID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi hệ thống khi lưu bài viết"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Thành công",
		"is_saved": isSaved,
	})
}

// GetSavedFeed
func (h *BookmarkHandler) GetSavedFeed(c *gin.Context) {
	userID := c.GetInt64("user_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))

	posts, err := h.bookmarkUC.GetSavedPosts(c.Request.Context(), userID, page)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể lấy danh sách bài đã lưu"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": posts})
}

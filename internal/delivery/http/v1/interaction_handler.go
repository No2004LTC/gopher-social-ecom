package v1

import (
	"net/http"
	"strconv"

	"github.com/No2004LTC/gopher-social-ecom/internal/domain"
	"github.com/gin-gonic/gin"
)

type InteractionHandler struct {
	interUC domain.InteractionUsecase
}

func NewInteractionHandler(interUC domain.InteractionUsecase) *InteractionHandler {
	return &InteractionHandler{interUC: interUC}
}

func (h *InteractionHandler) ToggleLike(c *gin.Context) {
	userID := c.MustGet("user_id").(int64)
	postID, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	isLiked, err := h.interUC.ToggleLike(c.Request.Context(), userID, postID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	msg := "Liked"
	if !isLiked {
		msg = "Unliked"
	}
	c.JSON(http.StatusOK, gin.H{"message": msg, "liked": isLiked})
}

func (h *InteractionHandler) AddComment(c *gin.Context) {
	userID := c.MustGet("user_id").(int64)
	postID, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var input struct {
		Content string `json:"content" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	comment, err := h.interUC.CommentPost(c.Request.Context(), userID, postID, input.Content)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, comment)
}

func (h *InteractionHandler) UpdateComment(c *gin.Context) {
	// 1. SỬA "id" THÀNH "comment_id"
	commentID, err := strconv.ParseInt(c.Param("comment_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID bình luận không hợp lệ"})
		return
	}

	// 2. Dùng MustGet cho đồng bộ với hàm Xóa
	userID := c.MustGet("user_id").(int64)

	// 3. Parse JSON Body
	var req struct {
		Content string `json:"content" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Vui lòng nhập nội dung bình luận"})
		return
	}

	// 4. Gọi Usecase
	err = h.interUC.UpdateComment(c.Request.Context(), commentID, userID, req.Content)
	if err != nil {
		if err.Error() == "nội dung bình luận không được để trống" {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	// 5. Thành công
	c.JSON(http.StatusOK, gin.H{"message": "Đã cập nhật bình luận"})
}

// HANDLER: XÓA BÌNH LUẬN
func (h *InteractionHandler) DeleteComment(c *gin.Context) {
	// 👉 1. PHẢI LÀ "comment_id" ĐỂ KHỚP VỚI ROUTER
	commentID, err := strconv.ParseInt(c.Param("comment_id"), 10, 64)
	if err != nil {
		// Nếu dùng response.Error của cậu
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID bình luận không hợp lệ"})
		return
	}

	userID := c.MustGet("user_id").(int64)

	// 2. Truyền đúng commentID và userID xuống Usecase
	err = h.interUC.DeleteComment(c.Request.Context(), commentID, userID)
	if err != nil {
		// 👉 Lỗi 403 văng ra từ đây nếu DB báo RowsAffected == 0
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Xóa bình luận thành công"})
}

func (h *InteractionHandler) GetComments(c *gin.Context) {
	postID, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	comments, err := h.interUC.GetPostComments(c.Request.Context(), postID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, comments)
}

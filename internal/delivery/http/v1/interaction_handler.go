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

func (h *InteractionHandler) GetComments(c *gin.Context) {
	postID, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	comments, err := h.interUC.GetPostComments(c.Request.Context(), postID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, comments)
}

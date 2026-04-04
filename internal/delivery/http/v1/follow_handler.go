package v1

import (
	"net/http"
	"strconv"

	"github.com/No2004LTC/gopher-social-ecom/internal/domain"
	"github.com/gin-gonic/gin"
)

type FollowHandler struct {
	followUC domain.FollowUsecase
}

func NewFollowHandler(followUC domain.FollowUsecase) *FollowHandler {
	return &FollowHandler{followUC: followUC}
}

func (h *FollowHandler) Follow(c *gin.Context) {
	followerID := c.MustGet("user_id").(int64)
	followingID, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	err := h.followUC.FollowUser(c.Request.Context(), followerID, followingID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Followed successfully"})
}

func (h *FollowHandler) Unfollow(c *gin.Context) {
	followerID := c.MustGet("user_id").(int64)
	followingID, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	err := h.followUC.UnfollowUser(c.Request.Context(), followerID, followingID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Unfollowed successfully"})
}

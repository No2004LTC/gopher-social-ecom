package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type SuccessResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func Success(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, SuccessResponse{
		Message: message,
		Data:    data,
	})
}

func SuccessMessage(c *gin.Context, message string) {
	c.JSON(http.StatusOK, SuccessResponse{
		Message: message,
	})
}

func Error(c *gin.Context, statusCode int, errMessage string) {
	c.JSON(statusCode, ErrorResponse{
		Error: errMessage,
	})
}

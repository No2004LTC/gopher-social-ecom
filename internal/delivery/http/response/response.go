package response // Hoặc package v1 tùy thư mục cậu đặt

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// SuccessResponse là cấu trúc chuẩn cho mọi request thành công
type SuccessResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"` // omitempty: tự ẩn đi nếu data là nil
}

// ErrorResponse là cấu trúc chuẩn cho request thất bại
type ErrorResponse struct {
	Error string `json:"error"`
}

// Success trả về 200 OK kèm dữ liệu
func Success(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, SuccessResponse{
		Message: message,
		Data:    data,
	})
}

// SuccessMessage trả về 200 OK chỉ có lời nhắn (ví dụ: follow thành công)
func SuccessMessage(c *gin.Context, message string) {
	c.JSON(http.StatusOK, SuccessResponse{
		Message: message,
	})
}

// Error trả về mã lỗi HTTP tùy chỉnh
func Error(c *gin.Context, statusCode int, errMessage string) {
	c.JSON(statusCode, ErrorResponse{
		Error: errMessage,
	})
}

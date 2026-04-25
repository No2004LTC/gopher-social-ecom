package v1

import (
	"net/http"

	"github.com/No2004LTC/gopher-social-ecom/internal/delivery/http/response"
	"github.com/No2004LTC/gopher-social-ecom/internal/domain"
	"github.com/No2004LTC/gopher-social-ecom/internal/dto"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authUsecase domain.AuthUsecase
}

func NewAuthHandler(u domain.AuthUsecase) *AuthHandler {
	return &AuthHandler{authUsecase: u}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Dữ liệu không hợp lệ")
		return
	}
	if err := h.authUsecase.Register(c.Request.Context(), req.Username, req.Email, req.Password); err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.SuccessMessage(c, "Đăng ký thành công")
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Dữ liệu email không hợp lệ")
		return
	}
	token, user, err := h.authUsecase.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, err.Error())
		return
	}
	c.JSON(http.StatusOK, dto.LoginResponse{
		AccessToken: token,
		TokenType:   "Bearer",
		User: dto.AuthUserResponse{
			ID: user.ID, Username: user.Username, Email: user.Email, AvatarURL: user.AvatarURL,
		},
	})
}

func (h *AuthHandler) SendPasswordOTP(c *gin.Context) {
	var req dto.SendOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Email không hợp lệ")
		return
	}
	if err := h.authUsecase.SendPasswordOTP(c.Request.Context(), req.Email); err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.SuccessMessage(c, "Mã OTP đã được gửi")
}

func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req dto.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Dữ liệu không hợp lệ")
		return
	}
	if err := h.authUsecase.ChangePasswordWithOTP(c.Request.Context(), req.Email, req.OTP, req.NewPassword); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.SuccessMessage(c, "Đổi mật khẩu thành công")
}

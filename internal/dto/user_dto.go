package dto

type UserCompact struct {
	ID           int64  `json:"id"`
	Username     string `json:"username"`
	AvatarURL    string `json:"avatar_url"`
	IsFollowing  bool   `json:"is_following"`   // Mình có đang follow họ không?
	IsFollowedBy bool   `json:"is_followed_by"` // Họ có đang follow mình không?
}

type LoginRequest struct {
	Identifier string `json:"identifier" binding:"required"`
	Password   string `json:"password" binding:"required"`
}

// RegisterRequest (Nếu cậu muốn dọn dẹp nốt cả luồng đăng ký)
type RegisterRequest struct {
	Username string `json:"username" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

type UpdateProfileRequest struct {
	Username string `json:"username" binding:"required"`
}

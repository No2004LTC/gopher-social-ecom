package dto

// --- INPUT (REQUEST) ---

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type UpdateProfileRequest struct {
	Username string `json:"username" binding:"required"`
}

// --- OUTPUT (RESPONSE) ---

// AuthUserResponse: Thông tin cơ bản của user đi kèm sau khi login thành công
type AuthUserResponse struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
}

// LoginResponse: Cục JSON trả về cho React chứa Token và User
type LoginResponse struct {
	AccessToken string           `json:"access_token"`
	TokenType   string           `json:"token_type"`
	User        AuthUserResponse `json:"user"`
}

// UserCompact: Dùng cho danh sách tìm kiếm, gợi ý kết bạn, hoặc danh sách follow
type UserCompact struct {
	ID           int64  `json:"id"`
	Username     string `json:"username"`
	AvatarURL    string `json:"avatar_url"`
	IsFollowing  bool   `json:"is_following"`
	IsFollowedBy bool   `json:"is_followed_by"`
	IsOnline     bool   `json:"is_online"`
}

type SuggestedUserResponse struct {
	ID                 int64  `json:"id"`
	Username           string `json:"username"`
	AvatarURL          string `json:"avatar_url"`
	MutualFriendsCount int    `json:"mutual_friends_count"`
}

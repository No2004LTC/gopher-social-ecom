package dto

// LoginRequest: Hứng dữ liệu từ form Login của React
type AuthUserResponse struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
}

// Cục JSON hoàn chỉnh trả về cho React
type LoginResponse struct {
	AccessToken string           `json:"access_token"`
	TokenType   string           `json:"token_type"`
	User        AuthUserResponse `json:"user"` // 👉 Bơm thêm cục user vào đây
}

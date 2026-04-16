package dto

// --- USER REQUESTS ---

type UpdateProfileInput struct {
	// Sử dụng con trỏ để hỗ trợ PATCH (chỉ cập nhật những trường gửi lên)
	Username  *string `json:"username" binding:"omitempty,min=3"`
	Bio       *string `json:"bio"`
	AvatarURL *string `json:"avatar_url"`
}

// --- USER RESPONSES ---

// UserCompact: Dùng cho các danh sách (Search, Followers, Suggestions)
type UserCompact struct {
	ID           int64  `json:"id"`
	Username     string `json:"username"`
	AvatarURL    string `json:"avatar_url"`
	IsFollowing  bool   `json:"is_following"`
	IsFollowedBy bool   `json:"is_followed_by"`
	IsOnline     bool   `json:"is_online"`
}

// SuggestedUserResponse: Dành riêng cho mục gợi ý kết bạn
type SuggestedUserResponse struct {
	ID                 int64  `json:"id"`
	Username           string `json:"username"`
	AvatarURL          string `json:"avatar_url"`
	MutualFriendsCount int    `json:"mutual_friends_count"`
}

// UserProfileResponse: "Trùm cuối" của trang Profile cá nhân
type UserProfileResponse struct {
	ID             int64  `json:"id"`
	Username       string `json:"username"`
	Bio            string `json:"bio"`
	AvatarURL      string `json:"avatar_url"`
	CoverURL       string `json:"cover_url"`
	FollowersCount int    `json:"followers_count"`
	FollowingCount int    `json:"following_count"`
	PostsCount     int    `json:"posts_count"`
	IsFollowing    bool   `json:"is_following"`
}

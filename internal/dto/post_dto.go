package dto

import "time"

// PostResponse là cục JSON hoàn chỉnh trả về cho React (Newsfeed)
type PostResponse struct {
	ID            int64        `json:"id"`
	Content       string       `json:"content"`
	ImageURL      string       `json:"image_url"`
	LikesCount    int64        `json:"likes_count"`
	CommentsCount int64        `json:"comments_count"`
	IsLiked       bool         `json:"is_liked"` // <-- FE CẦN CÁI NÀY ĐỂ HIỆN TIM ĐỎ
	CreatedAt     time.Time    `json:"created_at"`
	Author        ActorCompact `json:"author"` // Tái sử dụng ActorCompact từ phần Notification
	IsSaved       bool         `json:"is_saved"`
}

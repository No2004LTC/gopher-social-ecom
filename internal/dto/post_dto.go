package dto

import "time"

// PostResponse
type PostResponse struct {
	ID            int64        `json:"id"`
	Content       string       `json:"content"`
	ImageURL      string       `json:"image_url"`
	LikesCount    int64        `json:"likes_count"`
	CommentsCount int64        `json:"comments_count"`
	IsLiked       bool         `json:"is_liked"`
	CreatedAt     time.Time    `json:"created_at"`
	Author        ActorCompact `json:"author"`
	IsSaved       bool         `json:"is_saved"`
}

package dto

import "time"

type ActorCompact struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	AvatarURL string `json:"avatar_url"`
}

type NotificationResponse struct {
	ID        int64        `json:"id"`
	Type      string       `json:"type"`
	Message   string       `json:"message"`
	IsRead    bool         `json:"is_read"`
	CreatedAt time.Time    `json:"created_at"`
	Actor     ActorCompact `json:"actor"`
}

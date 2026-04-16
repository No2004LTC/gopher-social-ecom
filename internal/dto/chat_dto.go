package dto

import "time"

type Conversation struct {
	PartnerID        int64     `json:"partner_id"`
	PartnerUsername  string    `json:"partner_username"`
	PartnerAvatarURL string    `json:"partner_avatar_url"`
	LastMessage      string    `json:"last_message"`
	LastMessageAt    time.Time `json:"last_message_at"`
	UnreadCount      int       `json:"unread_count"`
	IsOnline         bool      `json:"is_online"`
}

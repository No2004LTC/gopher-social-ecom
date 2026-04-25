package usecase

import (
	"context"
	"sort"
	"time"

	"github.com/No2004LTC/gopher-social-ecom/internal/domain"
	"github.com/No2004LTC/gopher-social-ecom/internal/dto"
)

type chatUsecase struct {
	chatRepo domain.ChatRepository
	userRepo domain.UserRepository
}

func NewChatUsecase(chatRepo domain.ChatRepository, userRepo domain.UserRepository) domain.ChatUsecase {
	return &chatUsecase{
		chatRepo: chatRepo,
		userRepo: userRepo,
	}
}

// SaveMessage
func (uc *chatUsecase) SaveMessage(ctx context.Context, msg *domain.Message) error {
	return uc.chatRepo.SaveMessage(ctx, msg)
}

// GetChatHistory
func (uc *chatUsecase) GetChatHistory(ctx context.Context, user1, user2 int64, limit int) ([]domain.Message, error) {
	return uc.chatRepo.GetHistory(ctx, user1, user2, limit)
}

// GetUnreadCount
func (uc *chatUsecase) GetUnreadCount(ctx context.Context, userID int64) (int, error) {
	return uc.chatRepo.GetUnreadCount(ctx, userID)
}

// GetCategorizedConversations
func (uc *chatUsecase) GetCategorizedConversations(ctx context.Context, userID int64) (map[string][]dto.Conversation, error) {
	allConvos, err := uc.chatRepo.GetConversations(ctx, userID)
	if err != nil {
		return nil, err
	}

	followingList, _ := uc.userRepo.GetFollowing(ctx, userID, 1000, 0)

	friendMap := make(map[int64]dto.Conversation)
	var strangers []dto.Conversation

	for _, f := range followingList {
		friendMap[f.ID] = dto.Conversation{
			PartnerID:        f.ID,
			PartnerUsername:  f.Username,
			PartnerAvatarURL: f.AvatarURL,
			LastMessage:      "Hãy gửi lời chào...",
			LastMessageAt:    time.Time{},
			UnreadCount:      0,
			IsOnline:         false,
		}
	}

	for _, convo := range allConvos {
		if _, isFriend := friendMap[convo.PartnerID]; isFriend {
			friendMap[convo.PartnerID] = convo
		} else {
			strangers = append(strangers, convo)
		}
	}

	var friends []dto.Conversation
	for _, v := range friendMap {
		friends = append(friends, v)
	}

	sort.Slice(friends, func(i, j int) bool {
		return friends[i].LastMessageAt.After(friends[j].LastMessageAt)
	})

	sort.Slice(strangers, func(i, j int) bool {
		return strangers[i].LastMessageAt.After(strangers[j].LastMessageAt)
	})

	return map[string][]dto.Conversation{
		"friends":   friends,
		"strangers": strangers,
	}, nil
}

// MarkMessagesAsRead
func (uc *chatUsecase) MarkMessagesAsRead(ctx context.Context, myUserID, partnerID int64) error {
	return uc.chatRepo.MarkMessagesAsRead(ctx, myUserID, partnerID)
}

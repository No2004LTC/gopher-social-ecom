package usecase

import (
	"context"
	"sort"
	"time"

	"github.com/No2004LTC/gopher-social-ecom/internal/domain"
	"github.com/No2004LTC/gopher-social-ecom/internal/dto"
)

// 1. Cập nhật struct: Nạp thêm userRepo
type chatUsecase struct {
	chatRepo domain.ChatRepository // Tớ đổi tên repo -> chatRepo cho rõ nghĩa
	userRepo domain.UserRepository // 👉 THÊM VŨ KHÍ MỚI
}

// 2. Cập nhật Constructor: Nhận thêm userRepo
func NewChatUsecase(chatRepo domain.ChatRepository, userRepo domain.UserRepository) domain.ChatUsecase {
	return &chatUsecase{
		chatRepo: chatRepo,
		userRepo: userRepo,
	}
}

func (uc *chatUsecase) SaveMessage(ctx context.Context, msg *domain.Message) error {
	return uc.chatRepo.SaveMessage(ctx, msg)
}

func (uc *chatUsecase) GetChatHistory(ctx context.Context, user1, user2 int64, limit int) ([]domain.Message, error) {
	return uc.chatRepo.GetHistory(ctx, user1, user2, limit)
}

// 👉 Đừng quên hàm đếm số lượng chưa đọc nhé
func (uc *chatUsecase) GetUnreadCount(ctx context.Context, userID int64) (int, error) {
	return uc.chatRepo.GetUnreadCount(ctx, userID)
}

func (uc *chatUsecase) GetCategorizedConversations(ctx context.Context, userID int64) (map[string][]dto.Conversation, error) {
	// 1. Lấy toàn bộ những người đã có lịch sử nhắn tin từ Repo
	allConvos, err := uc.chatRepo.GetConversations(ctx, userID)
	if err != nil {
		return nil, err
	}

	// 2. Lấy danh sách những người mình đang FOLLOW (Bạn bè)
	followingList, _ := uc.userRepo.GetFollowing(ctx, userID, 1000, 0)

	// Tạo Map để quản lý cho nhanh và tránh trùng lặp
	friendMap := make(map[int64]dto.Conversation)
	strangers := []dto.Conversation{}

	// BƯỚC A: Đưa tất cả bạn bè vào Map trước (mặc định tin nhắn trống)
	for _, f := range followingList {
		friendMap[f.ID] = dto.Conversation{
			PartnerID:        f.ID,
			PartnerUsername:  f.Username,
			PartnerAvatarURL: f.AvatarURL,
			LastMessage:      "",          // Chưa có tin nhắn
			LastMessageAt:    time.Time{}, // Thời gian mặc định (cũ nhất)
			UnreadCount:      0,
			IsOnline:         false, // Sau này ghép thêm logic online sau
		}
	}

	// BƯỚC B: Duyệt qua danh sách tin nhắn từ Repo để "cập nhật" vào Map hoặc đẩy vào Strangers
	for _, convo := range allConvos {
		if _, isFriend := friendMap[convo.PartnerID]; isFriend {
			// Nếu là bạn bè: Cập nhật thông tin tin nhắn mới nhất vào Map
			friendMap[convo.PartnerID] = convo
		} else {
			// Nếu không phải bạn bè: Cho vào mảng người lạ
			strangers = append(strangers, convo)
		}
	}

	// BƯỚC C: Chuyển Map bạn bè thành Slice (Mảng) để sắp xếp
	friends := []dto.Conversation{}
	for _, v := range friendMap {
		friends = append(friends, v)
	}

	// BƯỚC D: SẮP XẾP (Quan trọng nhất)
	// Ai nhắn tin mới nhất (LastMessageAt lớn nhất) sẽ đứng đầu.
	// Ai chưa nhắn tin (LastMessageAt rỗng) sẽ đứng cuối.
	sort.Slice(friends, func(i, j int) bool {
		return friends[i].LastMessageAt.After(friends[j].LastMessageAt)
	})

	// Tương tự sắp xếp cho mảng người lạ (nếu có)
	sort.Slice(strangers, func(i, j int) bool {
		return strangers[i].LastMessageAt.After(strangers[j].LastMessageAt)
	})

	return map[string][]dto.Conversation{
		"friends":   friends,
		"strangers": strangers,
	}, nil
}

func (uc *chatUsecase) MarkMessagesAsRead(ctx context.Context, myUserID, partnerID int64) error {
	return uc.chatRepo.MarkMessagesAsRead(ctx, myUserID, partnerID)
}

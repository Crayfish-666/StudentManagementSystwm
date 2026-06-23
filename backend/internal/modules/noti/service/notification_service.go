package service

import (
	"student-system/internal/modules/noti/repository"
)

// NotificationService 通知业务服务层。
type NotificationService struct {
	repo *repository.NotificationRepository
}

// NewNotificationService 创建通知服务。
func NewNotificationService(repo *repository.NotificationRepository) *NotificationService {
	return &NotificationService{repo: repo}
}

// NotificationItem 通知列表项视图。
type NotificationItem struct {
	ID        int64  `json:"id"`
	Title     string `json:"title"`
	Level     string `json:"level"`
	CreatedAt string `json:"created_at"`
	IsRead    int    `json:"is_read"`
}

// ListMine 查询当前用户的通知列表，返回 items, total, unread_count。
func (s *NotificationService) ListMine(userID int64, isRead *int, page, pageSize int) (map[string]interface{}, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	notifications, total, err := s.repo.ListByRecipient(userID, isRead, page, pageSize)
	if err != nil {
		return nil, err
	}

	items := make([]NotificationItem, 0, len(notifications))
	for _, n := range notifications {
		level := "info"
		switch n.Priority {
		case "urgent":
			level = "urgent"
		case "high":
			level = "warning"
		case "low":
			level = "low"
		}
		items = append(items, NotificationItem{
			ID:        n.ID,
			Title:     n.Title,
			Level:     level,
			CreatedAt: n.CreatedAt.Format("2006-01-02T15:04:05+08:00"),
			IsRead:    n.IsRead,
		})
	}

	unreadCount, _ := s.repo.CountUnread(userID)

	result := map[string]interface{}{
		"items":       items,
		"total":       total,
		"unread_count": unreadCount,
	}
	return result, nil
}

// GetUnreadCount 获取未读通知数。
func (s *NotificationService) GetUnreadCount(userID int64) (int64, error) {
	return s.repo.CountUnread(userID)
}

// MarkRead 标记已读。
func (s *NotificationService) MarkRead(id int64, userID int64) error {
	return s.repo.MarkRead(id, userID)
}

// MarkAllRead 全部已读。
func (s *NotificationService) MarkAllRead(userID int64) error {
	return s.repo.MarkAllRead(userID)
}

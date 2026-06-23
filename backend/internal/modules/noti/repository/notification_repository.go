package repository

import (
	"time"

	"gorm.io/gorm"

	"student-system/internal/models"
)

// NotificationRepository 通知数据访问层。
type NotificationRepository struct {
	db *gorm.DB
}

// NewNotificationRepository 创建通知仓储。
func NewNotificationRepository(db *gorm.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

// Create 创建通知。
func (r *NotificationRepository) Create(n *models.Notification) error {
	return r.db.Create(n).Error
}

// GetByID 按 ID 查询通知。
func (r *NotificationRepository) GetByID(id int64) (*models.Notification, error) {
	var n models.Notification
	if err := r.db.Where("id = ? AND is_deleted = 0", id).First(&n).Error; err != nil {
		return nil, err
	}
	return &n, nil
}

// ListByRecipient 按接收人分页查询通知，按 created_at DESC 排序。
func (r *NotificationRepository) ListByRecipient(userID int64, isRead *int, page, pageSize int) ([]models.Notification, int64, error) {
	query := r.db.Where("recipient_user_id = ? AND is_deleted = 0", userID)

	if isRead != nil {
		query = query.Where("is_read = ?", *isRead)
	}

	var total int64
	if err := query.Model(&models.Notification{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var notifications []models.Notification
	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&notifications).Error; err != nil {
		return nil, 0, err
	}

	return notifications, total, nil
}

// CountUnread 统计未读通知数。
func (r *NotificationRepository) CountUnread(userID int64) (int64, error) {
	var count int64
	if err := r.db.Model(&models.Notification{}).
		Where("recipient_user_id = ? AND is_read = 0 AND is_deleted = 0", userID).
		Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// MarkRead 标记已读，设置 read_at。
func (r *NotificationRepository) MarkRead(id int64, userID int64) error {
	now := time.Now()
	return r.db.Model(&models.Notification{}).
		Where("id = ? AND recipient_user_id = ? AND is_deleted = 0", id, userID).
		Updates(map[string]interface{}{
			"is_read": 1,
			"read_at": now,
		}).Error
}

// MarkAllRead 全部已读。
func (r *NotificationRepository) MarkAllRead(userID int64) error {
	now := time.Now()
	return r.db.Model(&models.Notification{}).
		Where("recipient_user_id = ? AND is_read = 0 AND is_deleted = 0", userID).
		Updates(map[string]interface{}{
			"is_read": 1,
			"read_at": now,
		}).Error
}

// UpdateSendStatus 更新发送状态。
func (r *NotificationRepository) UpdateSendStatus(id int64, status string, lastError string) error {
	updates := map[string]interface{}{
		"send_status": status,
		"last_error":  lastError,
	}
	if status == "sent" {
		updates["sent_at"] = time.Now()
	}
	return r.db.Model(&models.Notification{}).
		Where("id = ?", id).
		Updates(updates).Error
}

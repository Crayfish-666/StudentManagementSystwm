package api

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"student-system/internal/modules/noti/service"
	"student-system/pkg/response"
)

// NotificationHandler 通知接口处理器。
type NotificationHandler struct {
	svc *service.NotificationService
}

// NewNotificationHandler 创建通知处理器。
func NewNotificationHandler(svc *service.NotificationService) *NotificationHandler {
	return &NotificationHandler{svc: svc}
}

// ListMine 查询当前用户通知列表。GET /api/v1/notifications/mine
func (h *NotificationHandler) ListMine(c *gin.Context) {
	uid, _ := c.Get("uid")
	userID, _ := uid.(int64)

	// 解析 is_read 参数（0/1），与 unread_only=true 互斥
	var isRead *int
	if v := c.Query("is_read"); v != "" {
		ir, err := strconv.Atoi(v)
		if err == nil && (ir == 0 || ir == 1) {
			isRead = &ir
		}
	} else if v := c.Query("unread_only"); v == "true" || v == "1" {
		ir := 0
		isRead = &ir
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	result, err := h.svc.ListMine(userID, isRead, page, pageSize)
	if err != nil {
		response.Fail(c, 1500, "查询通知列表失败")
		return
	}
	response.OK(c, result)
}

// UnreadCount 获取未读通知数。GET /api/v1/notifications/unread-count
func (h *NotificationHandler) UnreadCount(c *gin.Context) {
	uid, _ := c.Get("uid")
	userID, _ := uid.(int64)

	count, err := h.svc.GetUnreadCount(userID)
	if err != nil {
		response.Fail(c, 1500, "查询未读数失败")
		return
	}
	response.OK(c, gin.H{"unread_count": count})
}

// MarkRead 标记已读。POST /api/v1/notifications/:id/read
func (h *NotificationHandler) MarkRead(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 40002, "无效的通知 ID")
		return
	}

	uid, _ := c.Get("uid")
	userID, _ := uid.(int64)

	if err := h.svc.MarkRead(id, userID); err != nil {
		response.Fail(c, 1500, "标记已读失败")
		return
	}
	response.OK(c, gin.H{"id": id})
}

// MarkAllRead 全部已读。POST /api/v1/notifications/read-all
func (h *NotificationHandler) MarkAllRead(c *gin.Context) {
	uid, _ := c.Get("uid")
	userID, _ := uid.(int64)

	if err := h.svc.MarkAllRead(userID); err != nil {
		response.Fail(c, 1500, "全部已读失败")
		return
	}
	response.OK(c, gin.H{"message": "ok"})
}

// RegisterRoutes 注册通知相关路由。
func (h *NotificationHandler) RegisterRoutes(rg *gin.RouterGroup) {
	noti := rg.Group("/notifications")
	{
		noti.GET("/mine", h.ListMine)
		noti.GET("/unread-count", h.UnreadCount)
		noti.POST("/:id/read", h.MarkRead)
		noti.POST("/read-all", h.MarkAllRead)
	}
}

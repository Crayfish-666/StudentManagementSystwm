package api

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"student-system/internal/modules/ty/service"
	"student-system/pkg/response"
)

// DevelopmentMeetingHandler 发展大会接口处理器。
type DevelopmentMeetingHandler struct {
	svc *service.DevelopmentMeetingService
}

// NewDevelopmentMeetingHandler 创建发展大会处理器。
func NewDevelopmentMeetingHandler(svc *service.DevelopmentMeetingService) *DevelopmentMeetingHandler {
	return &DevelopmentMeetingHandler{svc: svc}
}

// Create 创建发展大会记录。POST /api/v1/ty/development-meetings
func (h *DevelopmentMeetingHandler) Create(c *gin.Context) {
	userID, name, role, ip, ua := actorFromCtx(c)

	var req service.CreateDevelopmentMeetingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 40001, "参数不完整: "+err.Error())
		return
	}

	meeting, err := h.svc.Create(userID, &req, name, role, ip, ua)
	if err != nil {
		msg := err.Error()
		code := 1500
		switch msg {
		case "发展对象不存在":
			code = 1404
		case "发展对象尚未完成审批流程，错误码:2621":
			code = 2621
		case "政审结论包含不合格，终止发展，错误码:2610":
			code = 2610
		case "政审基本合格，需延长培养期3个月，错误码:2611":
			code = 2611
		case "发展对象尚未完成政审，错误码:2621":
			code = 2621
		case "实到人数不足应到人数的2/3，错误码:2620",
			"赞成票数不满足要求，须超过实到人数的一半，错误码:2620":
			code = 2620
		case "无效的决策值，必须是 pass/reject", "会议时间格式错误":
			code = 40001
		}
		response.Fail(c, code, msg)
		return
	}
	response.OK(c, meeting)
}

// Get 获取发展大会详情。GET /api/v1/ty/development-meetings/:id
func (h *DevelopmentMeetingHandler) Get(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, 40002, "无效的发展大会 ID")
		return
	}

	meeting, err := h.svc.GetByID(id)
	if err != nil {
		response.Fail(c, 1404, "发展大会不存在")
		return
	}
	response.OK(c, meeting)
}

// List 列表查询发展大会。GET /api/v1/ty/development-meetings
func (h *DevelopmentMeetingHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	result, err := h.svc.List(page, pageSize)
	if err != nil {
		response.Fail(c, 1500, "查询发展大会列表失败")
		return
	}
	response.OK(c, result)
}

// ListByDevelopmentID 按发展对象ID查询。GET /api/v1/ty/development-meetings?development_id=xxx
func (h *DevelopmentMeetingHandler) ListByDevelopmentID(c *gin.Context) {
	var developmentID int64
	if v := c.Query("development_id"); v != "" {
		developmentID, _ = strconv.ParseInt(v, 10, 64)
	}
	if developmentID == 0 {
		response.Fail(c, 40002, "缺少 development_id 参数")
		return
	}

	meetings, err := h.svc.ListByDevelopmentID(developmentID)
	if err != nil {
		response.Fail(c, 1500, "查询发展大会失败")
		return
	}
	response.OK(c, gin.H{"items": meetings})
}

// RegisterRoutes 注册发展大会相关路由。
func (h *DevelopmentMeetingHandler) RegisterRoutes(rg *gin.RouterGroup, _ gin.HandlerFunc) {
	ty := rg.Group("/ty")
	{
		ty.GET("/development-meetings", h.List)
		ty.GET("/development-meetings/:id", h.Get)

		// 创建发展大会（含票数校验和联动操作）
		ty.POST("/development-meetings", h.Create)
	}
}
